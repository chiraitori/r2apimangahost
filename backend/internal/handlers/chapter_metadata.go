package handlers

import (
	"context"

	"r2apimangahost/backend/internal/models"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type chapterMetadata struct {
	MangaID           primitive.ObjectID `bson:"_id"`
	Count             int                `bson:"count"`
	LastChapterNumber float64            `bson:"lastChapterNumber"`
}

// populateChapterMetadata replaces two MongoDB queries per manga with one
// aggregation for the complete result set.
func populateChapterMetadata(ctx context.Context, chapters *mongo.Collection, mangas []models.Manga) {
	if len(mangas) == 0 {
		return
	}

	ids := make([]primitive.ObjectID, 0, len(mangas))
	for _, manga := range mangas {
		ids = append(ids, manga.ID)
	}

	pipeline := mongo.Pipeline{
		{{Key: "$match", Value: bson.M{"mangaId": bson.M{"$in": ids}}}},
		{{Key: "$group", Value: bson.M{
			"_id":               "$mangaId",
			"count":             bson.M{"$sum": 1},
			"lastChapterNumber": bson.M{"$max": "$chapterNumber"},
		}}},
	}

	cursor, err := chapters.Aggregate(ctx, pipeline)
	if err != nil {
		return
	}
	defer cursor.Close(ctx)

	var results []chapterMetadata
	if err := cursor.All(ctx, &results); err != nil {
		return
	}

	metadataByManga := make(map[primitive.ObjectID]chapterMetadata, len(results))
	for _, result := range results {
		metadataByManga[result.MangaID] = result
	}

	for i := range mangas {
		if metadata, ok := metadataByManga[mangas[i].ID]; ok {
			lastChapter := metadata.LastChapterNumber
			mangas[i].LastChapterNumber = &lastChapter
			mangas[i].ChapterCount = metadata.Count
		}
	}
}
