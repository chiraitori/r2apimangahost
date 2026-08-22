package handlers

import (
	"net/http"
	"sort"

	"r2apimangahost/backend/internal/db"
	"r2apimangahost/backend/internal/models"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type GenreHandler struct {
	db *db.Database
}

func NewGenreHandler(db *db.Database) *GenreHandler {
	return &GenreHandler{db: db}
}

var defaultGenres = []string{
	"Action", "Adventure", "Comedy", "Drama", "Fantasy", "Harem",
	"Historical", "Horror", "Isekai", "Martial Arts", "Mecha",
	"Mystery", "Psychological", "Romance", "School Life", "Sci-Fi",
	"Seinen", "Shoujo", "Shounen", "Slice of Life", "Sports", "Supernatural",
	"Tragedy", "Webtoon", "Manhwa", "Manhua",
}

func (h *GenreHandler) GetGenres(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	genresFromDB, err := h.db.Mangas.Distinct(ctx, "genres", bson.M{})
	genreMap := make(map[string]bool)

	for _, g := range defaultGenres {
		genreMap[g] = true
	}

	if err == nil {
		for _, item := range genresFromDB {
			if str, ok := item.(string); ok && str != "" {
				genreMap[str] = true
			}
		}
	}

	var allGenres []string
	for g := range genreMap {
		allGenres = append(allGenres, g)
	}
	sort.Strings(allGenres)

	writeSuccess(w, allGenres)
}

type StatsResponse struct {
	TotalMangas   int64 `json:"totalMangas"`
	TotalChapters int64 `json:"totalChapters"`
	TotalViews    int64 `json:"totalViews"`
}

func (h *GenreHandler) GetStats(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	totalMangas, _ := h.db.Mangas.CountDocuments(ctx, bson.M{})
	totalChapters, _ := h.db.Chapters.CountDocuments(ctx, bson.M{})

	// Aggregate total views
	pipeline := []bson.M{
		{"$group": bson.M{"_id": nil, "totalViews": bson.M{"$sum": "$views"}}},
	}
	cursor, err := h.db.Mangas.Aggregate(ctx, pipeline)
	var totalViews int64 = 0
	if err == nil {
		var results []bson.M
		if err := cursor.All(ctx, &results); err == nil && len(results) > 0 {
			if views, ok := results[0]["totalViews"].(int64); ok {
				totalViews = views
			} else if viewsInt, ok := results[0]["totalViews"].(int32); ok {
				totalViews = int64(viewsInt)
			}
		}
		cursor.Close(ctx)
	}

	writeSuccess(w, StatsResponse{
		TotalMangas:   totalMangas,
		TotalChapters: totalChapters,
		TotalViews:    totalViews,
	})
}

type HomeResponse struct {
	Featured []models.Manga `json:"featured"`
	Latest   []models.Manga `json:"latest"`
	Popular  []models.Manga `json:"popular"`
}

func (h *GenreHandler) GetHome(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Featured: Top rated
	featCursor, _ := h.db.Mangas.Find(ctx, bson.M{}, options.Find().SetSort(bson.D{{Key: "rating", Value: -1}}).SetLimit(5))
	var featured []models.Manga
	if featCursor != nil {
		_ = featCursor.All(ctx, &featured)
		featCursor.Close(ctx)
	}

	// Latest updated
	latestCursor, _ := h.db.Mangas.Find(ctx, bson.M{}, options.Find().SetSort(bson.D{{Key: "updatedAt", Value: -1}}).SetLimit(12))
	var latest []models.Manga
	if latestCursor != nil {
		_ = latestCursor.All(ctx, &latest)
		latestCursor.Close(ctx)
	}

	// Populate latest chapter info
	for i := range latest {
		var lastChapter models.Chapter
		err := h.db.Chapters.FindOne(ctx,
			bson.M{"mangaId": latest[i].ID},
			options.FindOne().SetSort(bson.D{{Key: "chapterNumber", Value: -1}}),
		).Decode(&lastChapter)
		if err == nil {
			latest[i].LastChapterNumber = &lastChapter.ChapterNumber
		}
		cnt, _ := h.db.Chapters.CountDocuments(ctx, bson.M{"mangaId": latest[i].ID})
		latest[i].ChapterCount = int(cnt)
	}

	// Popular: Most views
	popCursor, _ := h.db.Mangas.Find(ctx, bson.M{}, options.Find().SetSort(bson.D{{Key: "views", Value: -1}}).SetLimit(10))
	var popular []models.Manga
	if popCursor != nil {
		_ = popCursor.All(ctx, &popular)
		popCursor.Close(ctx)
	}

	if featured == nil {
		featured = []models.Manga{}
	}
	if latest == nil {
		latest = []models.Manga{}
	}
	if popular == nil {
		popular = []models.Manga{}
	}

	writeSuccess(w, HomeResponse{
		Featured: featured,
		Latest:   latest,
		Popular:  popular,
	})
}
