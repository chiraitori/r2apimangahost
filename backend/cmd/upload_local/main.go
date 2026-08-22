package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"
	"unicode"

	"r2apimangahost/backend/internal/config"
	"r2apimangahost/backend/internal/db"
	"r2apimangahost/backend/internal/models"
	"r2apimangahost/backend/internal/storage"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

func naturalLess(a, b string) bool {
	chunksA := splitChunks(a)
	chunksB := splitChunks(b)

	for i := 0; i < len(chunksA) && i < len(chunksB); i++ {
		numA, errA := strconv.Atoi(chunksA[i])
		numB, errB := strconv.Atoi(chunksB[i])

		if errA == nil && errB == nil {
			if numA != numB {
				return numA < numB
			}
		} else {
			if strings.ToLower(chunksA[i]) != strings.ToLower(chunksB[i]) {
				return strings.ToLower(chunksA[i]) < strings.ToLower(chunksB[i])
			}
		}
	}
	return len(chunksA) < len(chunksB)
}

func splitChunks(s string) []string {
	var chunks []string
	var current strings.Builder
	isPrevDigit := false

	for i, r := range s {
		isDigit := unicode.IsDigit(r)
		if i > 0 && isDigit != isPrevDigit {
			chunks = append(chunks, current.String())
			current.Reset()
		}
		current.WriteRune(r)
		isPrevDigit = isDigit
	}
	if current.Len() > 0 {
		chunks = append(chunks, current.String())
	}
	return chunks
}

func main() {
	dirPath := flag.String("dir", "", "Path to folder containing chapter images")
	mangaTitle := flag.String("title", "Trans Comics", "Manga Title")
	mangaSlug := flag.String("slug", "trans-comics", "Manga Slug")
	chapterNum := flag.Float64("chapter", 3, "Chapter Number")
	chapterTitle := flag.String("chapter-title", "", "Chapter Title")
	flag.Parse()

	if *dirPath == "" {
		log.Fatal("❌ Please specify folder path using -dir flag")
	}

	cfg := config.LoadConfig()
	log.Printf("Connecting to MongoDB: %s ...", cfg.DBName)
	mongoDB, err := db.ConnectMongoDB(cfg)
	if err != nil {
		log.Fatalf("❌ MongoDB connection error: %v", err)
	}
	defer mongoDB.Client.Disconnect(context.Background())

	log.Println("Initializing Cloudflare R2 Storage client...")
	r2Store, err := storage.NewR2Storage(cfg)
	if err != nil {
		log.Fatalf("❌ R2 initialization error: %v", err)
	}

	ctx := context.Background()

	// 1. Find or Create Manga
	var manga models.Manga
	err = mongoDB.Mangas.FindOne(ctx, bson.M{"slug": *mangaSlug}).Decode(&manga)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			log.Printf("Manga '%s' not found, creating new Manga entry...", *mangaTitle)
			now := time.Now()
			manga = models.Manga{
				ID:          primitive.NewObjectID(),
				Title:       *mangaTitle,
				AltTitles:   []string{},
				Slug:        *mangaSlug,
				CoverURL:    "",
				Description: "Bộ truyện manga lưu trữ trên Cloudflare R2",
				Author:      "Unknown",
				Artist:      "Unknown",
				Status:      models.StatusOngoing,
				Genres:      []string{"Action", "Fantasy"},
				Views:       0,
				Rating:      5.0,
				CreatedAt:   now,
				UpdatedAt:   now,
			}
			_, err = mongoDB.Mangas.InsertOne(ctx, manga)
			if err != nil {
				log.Fatalf("❌ Failed to create manga in DB: %v", err)
			}
			log.Printf("✅ Created manga '%s' (ID: %s, Slug: %s)", manga.Title, manga.ID.Hex(), manga.Slug)
		} else {
			log.Fatalf("❌ MongoDB query error: %v", err)
		}
	} else {
		log.Printf("✅ Found existing manga '%s' (ID: %s, Slug: %s)", manga.Title, manga.ID.Hex(), manga.Slug)
	}

	// 2. Read and Sort Images from Directory
	entries, err := os.ReadDir(*dirPath)
	if err != nil {
		log.Fatalf("❌ Failed to read directory %s: %v", *dirPath, err)
	}

	var imageFiles []string
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		ext := strings.ToLower(filepath.Ext(entry.Name()))
		if ext == ".png" || ext == ".jpg" || ext == ".jpeg" || ext == ".webp" || ext == ".avif" {
			imageFiles = append(imageFiles, entry.Name())
		}
	}

	if len(imageFiles) == 0 {
		log.Fatalf("❌ No image files found in %s", *dirPath)
	}

	sort.Slice(imageFiles, func(i, j int) bool {
		return naturalLess(imageFiles[i], imageFiles[j])
	})

	log.Printf("📂 Found %d image files in folder. Starting upload to Cloudflare R2...", len(imageFiles))

	// 3. Upload Each File to R2
	var pageURLs []string
	chapterKeyPrefix := fmt.Sprintf("chapters/%s/c%s", manga.Slug, strings.ReplaceAll(strconv.FormatFloat(*chapterNum, 'f', -1, 64), ".", "_"))

	for idx, fileName := range imageFiles {
		fullPath := filepath.Join(*dirPath, fileName)
		fileBytes, err := os.ReadFile(fullPath)
		if err != nil {
			log.Fatalf("❌ Failed to read %s: %v", fileName, err)
		}

		ext := strings.ToLower(filepath.Ext(fileName))
		r2Key := fmt.Sprintf("%s/%03d%s", chapterKeyPrefix, idx+1, ext)

		contentType := "image/png"
		if ext == ".jpg" || ext == ".jpeg" {
			contentType = "image/jpeg"
		} else if ext == ".webp" {
			contentType = "image/webp"
		}

		url, err := r2Store.UploadFile(ctx, r2Key, strings.NewReader(string(fileBytes)), contentType)
		if err != nil {
			log.Fatalf("❌ Failed to upload %s to R2 (%s): %v", fileName, r2Key, err)
		}

		pageURLs = append(pageURLs, url)
		fmt.Printf("   [%d/%d] Uploaded %s -> %s\n", idx+1, len(imageFiles), fileName, url)
	}

	// 4. Update Cover if manga has no cover
	if manga.CoverURL == "" && len(pageURLs) > 0 {
		log.Println("Setting page 1 as manga cover...")
		_, _ = mongoDB.Mangas.UpdateOne(ctx, bson.M{"_id": manga.ID}, bson.M{"$set": bson.M{"coverUrl": pageURLs[0]}})
	}

	// 5. Insert / Update Chapter in MongoDB
	now := time.Now()
	var existingChap models.Chapter
	err = mongoDB.Chapters.FindOne(ctx, bson.M{"mangaId": manga.ID, "chapterNumber": *chapterNum}).Decode(&existingChap)

	if err == nil {
		update := bson.M{
			"title":     *chapterTitle,
			"pages":     pageURLs,
			"pageCount": len(pageURLs),
			"updatedAt": now,
		}
		_, err = mongoDB.Chapters.UpdateOne(ctx, bson.M{"_id": existingChap.ID}, bson.M{"$set": update})
		if err != nil {
			log.Fatalf("❌ Failed to update chapter in DB: %v", err)
		}
		log.Printf("✅ Updated existing Chapter %.1f in MongoDB (ID: %s)", *chapterNum, existingChap.ID.Hex())
	} else {
		newChap := models.Chapter{
			ID:            primitive.NewObjectID(),
			MangaID:       manga.ID,
			MangaSlug:     manga.Slug,
			ChapterNumber: *chapterNum,
			Title:         *chapterTitle,
			Language:      "vi",
			Pages:         pageURLs,
			PageCount:     len(pageURLs),
			CreatedAt:     now,
			UpdatedAt:     now,
		}
		_, err = mongoDB.Chapters.InsertOne(ctx, newChap)
		if err != nil {
			log.Fatalf("❌ Failed to insert chapter into DB: %v", err)
		}
		log.Printf("✅ Inserted Chapter %.1f into MongoDB (ID: %s)", *chapterNum, newChap.ID.Hex())
	}

	// 6. Update Manga updatedAt
	_, _ = mongoDB.Mangas.UpdateOne(ctx, bson.M{"_id": manga.ID}, bson.M{"$set": bson.M{"updatedAt": now}})

	log.Println("==================================================")
	log.Printf("🎉 SUCCESS! Chapter %.1f with %d pages uploaded to Cloudflare R2!", *chapterNum, len(pageURLs))
	log.Printf("📖 Read URL on web: http://localhost:3000/manga/%s", manga.Slug)
	log.Println("==================================================")
}
