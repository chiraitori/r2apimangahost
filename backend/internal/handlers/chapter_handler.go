package handlers

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"path/filepath"
	"sort"
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

type ChapterHandler struct {
	db      *db.Database
	storage *storage.R2Storage
}

func NewChapterHandler(db *db.Database, storage *storage.R2Storage) *ChapterHandler {
	return &ChapterHandler{
		db:      db,
		storage: storage,
	}
}

type ChapterDetailResponse struct {
	Chapter     models.Chapter `json:"chapter"`
	Manga       models.Manga   `json:"manga"`
	PrevChapter *ChapterNav    `json:"prevChapter,omitempty"`
	NextChapter *ChapterNav    `json:"nextChapter,omitempty"`
}

type ChapterNav struct {
	ID            string  `json:"id"`
	ChapterNumber float64 `json:"chapterNumber"`
	Title         string  `json:"title"`
}

func (h *ChapterHandler) GetChapter(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	chapterIDStr := chi.URLParam(r, "id")

	var filter bson.M
	if objID, err := primitive.ObjectIDFromHex(chapterIDStr); err == nil {
		filter = bson.M{"_id": objID}
	} else {
		writeError(w, http.StatusBadRequest, "Invalid chapter ID")
		return
	}

	var chapter models.Chapter
	if err := h.db.Chapters.FindOne(ctx, filter).Decode(&chapter); err != nil {
		if err == mongo.ErrNoDocuments {
			writeError(w, http.StatusNotFound, "Chapter not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "Database error: "+err.Error())
		return
	}

	var manga models.Manga
	_ = h.db.Mangas.FindOne(ctx, bson.M{"_id": chapter.MangaID}).Decode(&manga)

	// Find Prev and Next chapters for reader navigation
	var prevChap, nextChap models.Chapter
	// Prev chapter (smaller chapter number)
	_ = h.db.Chapters.FindOne(ctx,
		bson.M{
			"mangaId":       chapter.MangaID,
			"chapterNumber": bson.M{"$lt": chapter.ChapterNumber},
		},
		options.FindOne().SetSort(bson.D{{Key: "chapterNumber", Value: -1}}),
	).Decode(&prevChap)

	// Next chapter (greater chapter number)
	_ = h.db.Chapters.FindOne(ctx,
		bson.M{
			"mangaId":       chapter.MangaID,
			"chapterNumber": bson.M{"$gt": chapter.ChapterNumber},
		},
		options.FindOne().SetSort(bson.D{{Key: "chapterNumber", Value: 1}}),
	).Decode(&nextChap)

	resp := ChapterDetailResponse{
		Chapter: chapter,
		Manga:   manga,
	}

	if !prevChap.ID.IsZero() {
		resp.PrevChapter = &ChapterNav{
			ID:            prevChap.ID.Hex(),
			ChapterNumber: prevChap.ChapterNumber,
			Title:         prevChap.Title,
		}
	}
	if !nextChap.ID.IsZero() {
		resp.NextChapter = &ChapterNav{
			ID:            nextChap.ID.Hex(),
			ChapterNumber: nextChap.ChapterNumber,
			Title:         nextChap.Title,
		}
	}

	writeSuccess(w, resp)
}

func (h *ChapterHandler) UploadChapter(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	mangaIDOrSlug := chi.URLParam(r, "mangaId")
	limitRequestBody(w, r, maxChapterRequestBytes)

	// Keep only a small portion in memory; MaxBytesReader above enforces the hard total limit.
	if err := r.ParseMultipartForm(16 << 20); err != nil {
		writeParseError(w, err, "Failed to parse form")
		return
	}

	var mangaFilter bson.M
	if objID, err := primitive.ObjectIDFromHex(mangaIDOrSlug); err == nil {
		mangaFilter = bson.M{"_id": objID}
	} else {
		mangaFilter = bson.M{"slug": mangaIDOrSlug}
	}

	var manga models.Manga
	if err := h.db.Mangas.FindOne(ctx, mangaFilter).Decode(&manga); err != nil {
		writeError(w, http.StatusNotFound, "Manga not found")
		return
	}

	chapterNumStr := r.FormValue("chapterNumber")
	if chapterNumStr == "" {
		writeError(w, http.StatusBadRequest, "chapterNumber is required")
		return
	}

	chapterNumber, err := strconv.ParseFloat(chapterNumStr, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "Invalid chapterNumber: must be a number")
		return
	}

	title := r.FormValue("title")
	language := r.FormValue("language")
	if language == "" {
		language = "vi"
	}

	var volume *int
	if volStr := r.FormValue("volume"); volStr != "" {
		if v, err := strconv.Atoi(volStr); err == nil {
			volume = &v
		}
	}

	chapterKeyFolder := fmt.Sprintf("chapters/%s/c%s", manga.Slug, strings.ReplaceAll(strconv.FormatFloat(chapterNumber, 'f', -1, 64), ".", "_"))

	var pageURLs []string

	// Case 1: Uploaded via .zip / .cbz archive file
	archiveFile, archiveHeader, err := r.FormFile("archive")
	if err == nil {
		defer archiveFile.Close()
		if archiveHeader.Size <= 0 || archiveHeader.Size > maxArchiveBytes {
			writeError(w, http.StatusRequestEntityTooLarge, "Archive must be smaller than 200 MiB")
			return
		}
		ext := strings.ToLower(filepath.Ext(archiveHeader.Filename))
		if ext != ".zip" && ext != ".cbz" {
			writeError(w, http.StatusBadRequest, "Archive must be a .zip or .cbz file")
			return
		}

		zipBytes, err := io.ReadAll(archiveFile)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "Failed to read archive file: "+err.Error())
			return
		}

		urls, err := h.storage.ExtractAndUploadZip(ctx, zipBytes, chapterKeyFolder)
		if err != nil {
			writeError(w, http.StatusBadRequest, "Failed to extract and upload zip: "+err.Error())
			return
		}
		pageURLs = urls
	} else if files := r.MultipartForm.File["pages"]; len(files) > 0 {
		// Case 2: Uploaded via multiple individual image files
		sort.Slice(files, func(i, j int) bool {
			return naturalLess(files[i].Filename, files[j].Filename)
		})

		for idx, fh := range files {
			if !isAllowedImageFilename(fh.Filename) {
				writeError(w, http.StatusBadRequest, fmt.Sprintf("Page %d has an unsupported image format", idx+1))
				return
			}
			if fh.Size <= 0 || fh.Size > maxPageImageBytes {
				writeError(w, http.StatusRequestEntityTooLarge, fmt.Sprintf("Page %d must be smaller than 50 MiB", idx+1))
				return
			}
			f, err := fh.Open()
			if err != nil {
				continue
			}

			ext := strings.ToLower(filepath.Ext(fh.Filename))
			if ext == "" {
				ext = ".jpg"
			}

			key := fmt.Sprintf("%s/%03d%s", chapterKeyFolder, idx+1, ext)
			url, err := h.storage.UploadFile(ctx, key, f, fh.Header.Get("Content-Type"))
			f.Close()
			if err != nil {
				writeError(w, http.StatusInternalServerError, fmt.Sprintf("Failed to upload page %d to R2: %v", idx+1, err))
				return
			}
			pageURLs = append(pageURLs, url)
		}
	} else if manualUrlsStr := r.FormValue("pageUrls"); manualUrlsStr != "" {
		// Case 3: List of manual image URLs
		pageURLs = splitCommaOrLines(manualUrlsStr)
	}

	if len(pageURLs) == 0 {
		writeError(w, http.StatusBadRequest, "No pages provided. Please upload an archive (.zip/.cbz) or image files (pages)")
		return
	}

	now := time.Now()

	// Check if this chapter number already exists for this manga
	var existingChapter models.Chapter
	err = h.db.Chapters.FindOne(ctx, bson.M{
		"mangaId":       manga.ID,
		"chapterNumber": chapterNumber,
	}).Decode(&existingChapter)

	var savedChapter models.Chapter
	if err == nil {
		// Update existing chapter
		update := bson.M{
			"title":     title,
			"volume":    volume,
			"language":  language,
			"pages":     pageURLs,
			"pageCount": len(pageURLs),
			"updatedAt": now,
		}
		_, err = h.db.Chapters.UpdateOne(ctx, bson.M{"_id": existingChapter.ID}, bson.M{"$set": update})
		if err != nil {
			writeError(w, http.StatusInternalServerError, "Failed to update chapter: "+err.Error())
			return
		}
		savedChapter = existingChapter
		savedChapter.Title = title
		savedChapter.Volume = volume
		savedChapter.Pages = pageURLs
		savedChapter.PageCount = len(pageURLs)
		savedChapter.UpdatedAt = now
	} else {
		// Insert new chapter
		newChapter := models.Chapter{
			ID:            primitive.NewObjectID(),
			MangaID:       manga.ID,
			MangaSlug:     manga.Slug,
			ChapterNumber: chapterNumber,
			Title:         title,
			Volume:        volume,
			Language:      language,
			Pages:         pageURLs,
			PageCount:     len(pageURLs),
			CreatedAt:     now,
			UpdatedAt:     now,
		}
		_, err = h.db.Chapters.InsertOne(ctx, newChapter)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "Failed to insert chapter: "+err.Error())
			return
		}
		savedChapter = newChapter
	}

	// Bump manga updatedAt timestamp
	_, _ = h.db.Mangas.UpdateOne(ctx, bson.M{"_id": manga.ID}, bson.M{"$set": bson.M{"updatedAt": now}})

	writeSuccess(w, savedChapter, fmt.Sprintf("Chapter %.1f uploaded successfully with %d pages", chapterNumber, len(pageURLs)))
}

func (h *ChapterHandler) DeleteChapter(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	chapterIDStr := chi.URLParam(r, "id")

	objID, err := primitive.ObjectIDFromHex(chapterIDStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, "Invalid chapter ID")
		return
	}

	var chapter models.Chapter
	if err := h.db.Chapters.FindOne(ctx, bson.M{"_id": objID}).Decode(&chapter); err != nil {
		writeError(w, http.StatusNotFound, "Chapter not found")
		return
	}

	_, _ = h.db.Chapters.DeleteOne(ctx, bson.M{"_id": objID})

	// Delete R2 objects under chapters/<mangaSlug>/c<num> asynchronously
	chapterKeyFolder := fmt.Sprintf("chapters/%s/c%s", chapter.MangaSlug, strings.ReplaceAll(strconv.FormatFloat(chapter.ChapterNumber, 'f', -1, 64), ".", "_"))
	go func(folder string) {
		_ = h.storage.DeletePrefix(context.Background(), folder)
	}(chapterKeyFolder)

	writeSuccess(w, map[string]string{"id": chapterIDStr}, "Chapter deleted successfully")
}
