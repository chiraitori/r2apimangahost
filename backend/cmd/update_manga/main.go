package main

import (
	"context"
	"io"
	"log"
	"net/http"
	"strings"
	"time"

	"r2apimangahost/backend/internal/config"
	"r2apimangahost/backend/internal/db"
	"r2apimangahost/backend/internal/models"
	"r2apimangahost/backend/internal/storage"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func main() {
	cfg := config.LoadConfig()
	log.Printf("Connecting to MongoDB: %s ...", cfg.DBName)
	mongoDB, err := db.ConnectMongoDB(cfg)
	if err != nil {
		log.Fatalf("❌ MongoDB error: %v", err)
	}
	defer mongoDB.Client.Disconnect(context.Background())

	r2Store, err := storage.NewR2Storage(cfg)
	if err != nil {
		log.Fatalf("❌ R2 error: %v", err)
	}

	ctx := context.Background()

	// 1. Download official high-res cover from MangaDex
	coverSourceURL := "https://uploads.mangadex.org/covers/02fd3073-8d24-4c3f-b23e-d9ecd6268a35/c0759278-2902-4868-a303-8ce7cc616f57.png"
	log.Printf("Downloading cover from %s ...", coverSourceURL)

	req, _ := http.NewRequestWithContext(ctx, "GET", coverSourceURL, nil)
	req.Header.Set("User-Agent", "MangaHost/1.0")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Fatalf("❌ Failed to download cover: %v", err)
	}
	defer resp.Body.Close()

	coverBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatalf("❌ Failed to read cover bytes: %v", err)
	}

	// 2. Upload cover to Cloudflare R2
	r2CoverKey := "mangas/momiji-to-kouyou/cover.png"
	coverR2URL, err := r2Store.UploadFile(ctx, r2CoverKey, strings.NewReader(string(coverBytes)), "image/png")
	if err != nil {
		log.Fatalf("❌ Failed to upload cover to R2: %v", err)
	}
	log.Printf("✅ Uploaded cover to R2: %s", coverR2URL)

	// 3. Find existing manga (e.g. trans-comics or create/update momiji-to-kouyou)
	newSlug := "momiji-to-kouyou"
	now := time.Now()

	var existingManga models.Manga
	err = mongoDB.Mangas.FindOne(ctx, bson.M{
		"$or": []bson.M{
			{"slug": "trans-comics"},
			{"slug": newSlug},
			{"title": bson.M{"$regex": "Momiji|Trans", "$options": "i"}},
		},
	}).Decode(&existingManga)

	description := `Konno Ao là một nữ sinh trung học xinh đẹp được mệnh danh là "Hoàng Tử". Touyama Kouyou - cậu bạn từng trêu chọc cô hồi tiểu học liên tục tỏ tình với cô.

"Tớ ghét con trai, tớ chỉ thích con gái thôi."

Đó là lý do Ao đưa ra khi từ chối cậu. Nhưng sau đó, người xuất hiện trước mặt cô lại là... "Momiji" - một cô gái siêu cấp dễ thương do chính Kouyou biến thành.

Ôm trong mình cảm giác có lỗi trong quá khứ và tình yêu hiện tại, Momiji tuyên bố: "Tớ sẽ chứng minh tớ có thể trở thành cô bạn gái tuyệt vời nhất của cậu."

Và thế là, câu chuyện tình cảm hài hước học đường đầy dở khóc dở cười chính thức bắt đầu.`

	altTitles := []string{"紅葉と紅葉", "Momiji & Kouyou", "Momiji and Kouyou"}
	genres := []string{"Romance", "Comedy", "Crossdressing", "Drama", "School Life", "Web Comic"}

	if err == nil {
		// Update existing manga
		update := bson.M{
			"title":       "Momiji to Kouyou",
			"slug":        newSlug,
			"altTitles":   altTitles,
			"coverUrl":    coverR2URL,
			"description": description,
			"author":      "Obiya Midori",
			"artist":      "Furatsu",
			"status":      models.StatusOngoing,
			"genres":      genres,
			"updatedAt":   now,
		}
		_, err = mongoDB.Mangas.UpdateOne(ctx, bson.M{"_id": existingManga.ID}, bson.M{"$set": update})
		if err != nil {
			log.Fatalf("❌ Failed to update manga: %v", err)
		}
		log.Printf("✅ Successfully updated Manga to 'Momiji to Kouyou' (ID: %s, Slug: %s)", existingManga.ID.Hex(), newSlug)

		// Update chapter mangaSlug reference
		_, _ = mongoDB.Chapters.UpdateMany(ctx, bson.M{"mangaId": existingManga.ID}, bson.M{"$set": bson.M{"mangaSlug": newSlug}})
	} else {
		// Create new manga
		newManga := models.Manga{
			ID:          primitive.NewObjectID(),
			Title:       "Momiji to Kouyou",
			AltTitles:   altTitles,
			Slug:        newSlug,
			CoverURL:    coverR2URL,
			Description: description,
			Author:      "Obiya Midori",
			Artist:      "Furatsu",
			Status:      models.StatusOngoing,
			Genres:      genres,
			Views:       0,
			Rating:      5.0,
			CreatedAt:   now,
			UpdatedAt:   now,
		}
		_, err = mongoDB.Mangas.InsertOne(ctx, newManga)
		if err != nil {
			log.Fatalf("❌ Failed to insert manga: %v", err)
		}
		log.Printf("✅ Successfully created Manga 'Momiji to Kouyou' (ID: %s)", newManga.ID.Hex())
	}

	log.Println("==================================================")
	log.Printf("🎉 Manga updated: Momiji to Kouyou (もみじと紅葉)")
	log.Printf("🖼️ Cover URL on R2: %s", coverR2URL)
	log.Printf("🌐 View on Web: http://localhost:3000/manga/%s", newSlug)
	log.Println("==================================================")
}
