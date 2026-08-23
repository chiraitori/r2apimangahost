package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"r2apimangahost/backend/internal/db"
	"r2apimangahost/backend/internal/models"
	"r2apimangahost/backend/internal/storage"

	"github.com/go-chi/chi/v5"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type MangaHandler struct {
	db      *db.Database
	storage *storage.R2Storage
}

func NewMangaHandler(db *db.Database, storage *storage.R2Storage) *MangaHandler {
	return &MangaHandler{
		db:      db,
		storage: storage,
	}
}

func (h *MangaHandler) GetMangas(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	q := r.URL.Query().Get("q")
	genre := r.URL.Query().Get("genre")
	status := r.URL.Query().Get("status")
	sortBy := r.URL.Query().Get("sort")
	orderStr := r.URL.Query().Get("order")
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))

	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 24
	}

	filter := bson.M{}

	if q != "" {
		quoted := regexp.QuoteMeta(q)
		filter["$or"] = []bson.M{
			{"title": bson.M{"$regex": quoted, "$options": "i"}},
			{"altTitles": bson.M{"$regex": quoted, "$options": "i"}},
			{"author": bson.M{"$regex": quoted, "$options": "i"}},
		}
	}

	if genre != "" {
		filter["genres"] = bson.M{"$in": []string{genre}}
	}

	if status != "" {
		filter["status"] = status
	}

	sortOrder := -1
	if orderStr == "asc" {
		sortOrder = 1
	}

	sortDoc := bson.D{{Key: "updatedAt", Value: sortOrder}}
	switch sortBy {
	case "title":
		sortDoc = bson.D{{Key: "title", Value: sortOrder}}
	case "views":
		sortDoc = bson.D{{Key: "views", Value: sortOrder}}
	case "rating":
		sortDoc = bson.D{{Key: "rating", Value: sortOrder}}
	case "createdAt":
		sortDoc = bson.D{{Key: "createdAt", Value: sortOrder}}
	}

	total, err := h.db.Mangas.CountDocuments(ctx, filter)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to count mangas: "+err.Error())
		return
	}

	findOptions := options.Find().
		SetSort(sortDoc).
		SetSkip(int64((page - 1) * limit)).
		SetLimit(int64(limit))

	cursor, err := h.db.Mangas.Find(ctx, filter, findOptions)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to fetch mangas: "+err.Error())
		return
	}
	defer cursor.Close(ctx)

	var mangas []models.Manga
	if err := cursor.All(ctx, &mangas); err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to decode mangas: "+err.Error())
		return
	}

	if mangas == nil {
		mangas = []models.Manga{}
	}

	// Populate latest chapter info for each manga
	for i := range mangas {
		var lastChapter models.Chapter
		err := h.db.Chapters.FindOne(ctx,
			bson.M{"mangaId": mangas[i].ID},
			options.FindOne().SetSort(bson.D{{Key: "chapterNumber", Value: -1}}),
		).Decode(&lastChapter)
		if err == nil {
			mangas[i].LastChapterNumber = &lastChapter.ChapterNumber
		}

		chapterCount, _ := h.db.Chapters.CountDocuments(ctx, bson.M{"mangaId": mangas[i].ID})
		mangas[i].ChapterCount = int(chapterCount)
	}

	totalPages := int((total + int64(limit) - 1) / int64(limit))

	writeSuccess(w, models.MangaListResponse{
		Data: mangas,
		Pagination: models.PaginationResponse{
			Page:       page,
			Limit:      limit,
			Total:      total,
			TotalPages: totalPages,
		},
	})
}

func (h *MangaHandler) GetManga(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	idOrSlug := chi.URLParam(r, "id")

	var filter bson.M
	if objID, err := primitive.ObjectIDFromHex(idOrSlug); err == nil {
		filter = bson.M{"_id": objID}
	} else {
		filter = bson.M{"slug": idOrSlug}
	}

	var manga models.Manga
	err := h.db.Mangas.FindOne(ctx, filter).Decode(&manga)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			writeError(w, http.StatusNotFound, "Manga not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "Database error: "+err.Error())
		return
	}

	// Increment view count asynchronously
	go func(id primitive.ObjectID) {
		_, _ = h.db.Mangas.UpdateOne(context.Background(), bson.M{"_id": id}, bson.M{"$inc": bson.M{"views": 1}})
	}(manga.ID)

	// Fetch all chapters sorted by chapterNumber desc
	chapterCursor, err := h.db.Chapters.Find(ctx,
		bson.M{"mangaId": manga.ID},
		options.Find().
			SetSort(bson.D{{Key: "chapterNumber", Value: -1}}).
			SetProjection(bson.M{"pages": 0}),
	)
	var chapters []models.ChapterSummary
	if err == nil {
		_ = chapterCursor.All(ctx, &chapters)
		chapterCursor.Close(ctx)
	}
	if chapters == nil {
		chapters = []models.ChapterSummary{}
	}

	manga.ChapterCount = len(chapters)
	if len(chapters) > 0 {
		manga.LastChapterNumber = &chapters[0].ChapterNumber
	}

	writeSuccess(w, models.MangaDetailResponse{
		Manga:    manga,
		Chapters: chapters,
	})
}

func (h *MangaHandler) CreateManga(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var req models.CreateMangaRequest
	var coverFileBytes []byte
	var coverFileName string
	var coverContentType string

	isMultipart := strings.HasPrefix(r.Header.Get("Content-Type"), "multipart/form-data")
	if isMultipart {
		limitRequestBody(w, r, maxCoverRequestBytes)
	} else {
		limitRequestBody(w, r, maxJSONBodyBytes)
	}

	if isMultipart {
		if err := r.ParseMultipartForm(8 << 20); err != nil {
			writeParseError(w, err, "Failed to parse form")
			return
		}

		req.Title = r.FormValue("title")
		req.Slug = r.FormValue("slug")
		req.CoverURL = r.FormValue("coverUrl")
		req.Description = r.FormValue("description")
		req.Author = r.FormValue("author")
		req.Artist = r.FormValue("artist")
		req.Status = r.FormValue("status")

		if altTitlesStr := r.FormValue("altTitles"); altTitlesStr != "" {
			req.AltTitles = splitCommaOrLines(altTitlesStr)
		}
		if genresStr := r.FormValue("genres"); genresStr != "" {
			req.Genres = splitCommaOrLines(genresStr)
		}

		file, handler, err := r.FormFile("cover")
		if err == nil {
			defer file.Close()
			if !isAllowedImageFilename(handler.Filename) {
				writeError(w, http.StatusBadRequest, "Unsupported cover image format")
				return
			}
			if handler.Size <= 0 || handler.Size > maxCoverImageBytes {
				writeError(w, http.StatusRequestEntityTooLarge, "Cover image must be smaller than 16 MiB")
				return
			}
			coverFileName = handler.Filename
			coverContentType = handler.Header.Get("Content-Type")
			coverFileBytes, err = io.ReadAll(file)
			if err != nil {
				writeError(w, http.StatusBadRequest, "Failed to read cover image")
				return
			}
		}
	} else {
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeParseError(w, err, "Invalid JSON payload")
			return
		}
	}

	if strings.TrimSpace(req.Title) == "" {
		writeError(w, http.StatusBadRequest, "Title is required")
		return
	}

	slug := req.Slug
	if strings.TrimSpace(slug) == "" {
		slug = slugify(req.Title)
	}

	// Check if slug exists
	count, _ := h.db.Mangas.CountDocuments(ctx, bson.M{"slug": slug})
	if count > 0 {
		slug = fmt.Sprintf("%s-%d", slug, time.Now().Unix())
	}

	coverURL := req.CoverURL
	if len(coverFileBytes) > 0 {
		ext := strings.ToLower(filepath.Ext(coverFileName))
		if ext == "" {
			ext = ".jpg"
		}
		key := fmt.Sprintf("mangas/%s/cover%s", slug, ext)
		uploadedURL, err := h.storage.UploadFile(ctx, key, bytes.NewReader(coverFileBytes), coverContentType)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "Failed to upload cover to R2: "+err.Error())
			return
		}
		coverURL = uploadedURL
	}

	status := models.MangaStatus(req.Status)
	if status == "" {
		status = models.StatusOngoing
	}

	now := time.Now()
	newManga := models.Manga{
		ID:          primitive.NewObjectID(),
		Title:       req.Title,
		AltTitles:   req.AltTitles,
		Slug:        slug,
		CoverURL:    coverURL,
		Description: req.Description,
		Author:      req.Author,
		Artist:      req.Artist,
		Status:      status,
		Genres:      req.Genres,
		Views:       0,
		Rating:      5.0,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	if newManga.AltTitles == nil {
		newManga.AltTitles = []string{}
	}
	if newManga.Genres == nil {
		newManga.Genres = []string{}
	}

	_, err := h.db.Mangas.InsertOne(ctx, newManga)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to save manga: "+err.Error())
		return
	}

	writeSuccess(w, newManga, "Manga created successfully")
}

func (h *MangaHandler) UpdateManga(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	idOrSlug := chi.URLParam(r, "id")

	var filter bson.M
	if objID, err := primitive.ObjectIDFromHex(idOrSlug); err == nil {
		filter = bson.M{"_id": objID}
	} else {
		filter = bson.M{"slug": idOrSlug}
	}

	var existing models.Manga
	if err := h.db.Mangas.FindOne(ctx, filter).Decode(&existing); err != nil {
		writeError(w, http.StatusNotFound, "Manga not found")
		return
	}

	updateDoc := bson.M{
		"updatedAt": time.Now(),
	}

	isMultipart := strings.HasPrefix(r.Header.Get("Content-Type"), "multipart/form-data")
	if isMultipart {
		limitRequestBody(w, r, maxCoverRequestBytes)
	} else {
		limitRequestBody(w, r, maxJSONBodyBytes)
	}
	if isMultipart {
		if err := r.ParseMultipartForm(8 << 20); err != nil {
			writeParseError(w, err, "Failed to parse form")
			return
		}

		if t := r.FormValue("title"); t != "" {
			updateDoc["title"] = t
		}
		if d := r.FormValue("description"); d != "" {
			updateDoc["description"] = d
		}
		if a := r.FormValue("author"); a != "" {
			updateDoc["author"] = a
		}
		if a := r.FormValue("artist"); a != "" {
			updateDoc["artist"] = a
		}
		if s := r.FormValue("status"); s != "" {
			updateDoc["status"] = models.MangaStatus(s)
		}
		if altStr := r.FormValue("altTitles"); altStr != "" {
			updateDoc["altTitles"] = splitCommaOrLines(altStr)
		}
		if genStr := r.FormValue("genres"); genStr != "" {
			updateDoc["genres"] = splitCommaOrLines(genStr)
		}
		if coverURL := r.FormValue("coverUrl"); coverURL != "" {
			updateDoc["coverUrl"] = coverURL
		}

		file, handler, err := r.FormFile("cover")
		if err == nil {
			defer file.Close()
			if !isAllowedImageFilename(handler.Filename) {
				writeError(w, http.StatusBadRequest, "Unsupported cover image format")
				return
			}
			if handler.Size <= 0 || handler.Size > maxCoverImageBytes {
				writeError(w, http.StatusRequestEntityTooLarge, "Cover image must be smaller than 16 MiB")
				return
			}
			ext := strings.ToLower(filepath.Ext(handler.Filename))
			if ext == "" {
				ext = ".jpg"
			}
			key := fmt.Sprintf("mangas/%s/cover-%d%s", existing.Slug, time.Now().Unix(), ext)
			uploadedURL, err := h.storage.UploadFile(ctx, key, file, handler.Header.Get("Content-Type"))
			if err != nil {
				writeError(w, http.StatusInternalServerError, "Failed to upload cover: "+err.Error())
				return
			}
			updateDoc["coverUrl"] = uploadedURL
		}
	} else {
		var req models.UpdateMangaRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeParseError(w, err, "Invalid JSON payload")
			return
		}

		if req.Title != nil {
			updateDoc["title"] = *req.Title
		}
		if req.Description != nil {
			updateDoc["description"] = *req.Description
		}
		if req.Author != nil {
			updateDoc["author"] = *req.Author
		}
		if req.Artist != nil {
			updateDoc["artist"] = *req.Artist
		}
		if req.Status != nil {
			updateDoc["status"] = models.MangaStatus(*req.Status)
		}
		if req.CoverURL != nil {
			updateDoc["coverUrl"] = *req.CoverURL
		}
		if req.AltTitles != nil {
			updateDoc["altTitles"] = *req.AltTitles
		}
		if req.Genres != nil {
			updateDoc["genres"] = *req.Genres
		}
	}

	_, err := h.db.Mangas.UpdateOne(ctx, filter, bson.M{"$set": updateDoc})
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to update manga: "+err.Error())
		return
	}

	var updated models.Manga
	_ = h.db.Mangas.FindOne(ctx, filter).Decode(&updated)
	writeSuccess(w, updated, "Manga updated successfully")
}

func (h *MangaHandler) DeleteManga(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	idOrSlug := chi.URLParam(r, "id")

	var filter bson.M
	if objID, err := primitive.ObjectIDFromHex(idOrSlug); err == nil {
		filter = bson.M{"_id": objID}
	} else {
		filter = bson.M{"slug": idOrSlug}
	}

	var manga models.Manga
	if err := h.db.Mangas.FindOne(ctx, filter).Decode(&manga); err != nil {
		writeError(w, http.StatusNotFound, "Manga not found")
		return
	}

	// Delete from MongoDB
	_, _ = h.db.Mangas.DeleteOne(ctx, bson.M{"_id": manga.ID})
	_, _ = h.db.Chapters.DeleteMany(ctx, bson.M{"mangaId": manga.ID})

	// Delete R2 objects under mangas/<slug> and chapters/<slug> asynchronously
	go func(slug string) {
		_ = h.storage.DeletePrefix(context.Background(), fmt.Sprintf("mangas/%s", slug))
		_ = h.storage.DeletePrefix(context.Background(), fmt.Sprintf("chapters/%s", slug))
	}(manga.Slug)

	writeSuccess(w, map[string]string{"id": manga.ID.Hex(), "slug": manga.Slug}, "Manga and associated chapters deleted")
}

func splitCommaOrLines(s string) []string {
	var result []string
	items := strings.FieldsFunc(s, func(r rune) bool {
		return r == ',' || r == '\n' || r == ';'
	})
	for _, item := range items {
		trimmed := strings.TrimSpace(item)
		if trimmed != "" {
			result = append(result, trimmed)
		}
	}
	return result
}
