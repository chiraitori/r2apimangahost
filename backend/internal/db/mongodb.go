package db

import (
	"context"
	"log"
	"time"

	"r2apimangahost/backend/internal/config"
	"r2apimangahost/backend/internal/models"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"golang.org/x/crypto/bcrypt"
)

type Database struct {
	Client   *mongo.Client
	DB       *mongo.Database
	Mangas   *mongo.Collection
	Chapters *mongo.Collection
	Users    *mongo.Collection
}

func ConnectMongoDB(cfg *config.Config) (*Database, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	clientOptions := options.Client().ApplyURI(cfg.MongoURI)
	// Optimize pool size for low memory usage
	clientOptions.SetMaxPoolSize(20)
	clientOptions.SetMinPoolSize(2)
	clientOptions.SetMaxConnIdleTime(5 * time.Minute)

	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		return nil, err
	}

	// Ping database
	if err := client.Ping(ctx, nil); err != nil {
		return nil, err
	}

	database := client.Database(cfg.DBName)
	dbInstance := &Database{
		Client:   client,
		DB:       database,
		Mangas:   database.Collection("mangas"),
		Chapters: database.Collection("chapters"),
		Users:    database.Collection("users"),
	}

	// Create Indexes
	if err := dbInstance.ensureIndexes(ctx); err != nil {
		log.Printf("[DB] Warning: Error creating indexes: %v", err)
	}

	// Seed Admin User
	if err := dbInstance.seedAdmin(ctx, cfg); err != nil {
		log.Printf("[DB] Warning: Error seeding admin: %v", err)
	}

	log.Printf("[DB] Successfully connected to MongoDB database: %s", cfg.DBName)
	return dbInstance, nil
}

func (d *Database) ensureIndexes(ctx context.Context) error {
	// Manga indexes: slug (unique), genres, updatedAt
	_, err := d.Mangas.Indexes().CreateMany(ctx, []mongo.IndexModel{
		{
			Keys:    bson.D{{Key: "slug", Value: 1}},
			Options: options.Index().SetUnique(true),
		},
		{
			Keys: bson.D{{Key: "title", Value: "text"}, {Key: "altTitles", Value: "text"}, {Key: "author", Value: "text"}},
		},
		{
			Keys: bson.D{{Key: "genres", Value: 1}},
		},
		{
			Keys: bson.D{{Key: "updatedAt", Value: -1}},
		},
	})
	if err != nil {
		return err
	}

	// Chapter indexes: mangaId + chapterNumber
	_, err = d.Chapters.Indexes().CreateMany(ctx, []mongo.IndexModel{
		{
			Keys: bson.D{{Key: "mangaId", Value: 1}, {Key: "chapterNumber", Value: -1}},
		},
		{
			Keys: bson.D{{Key: "mangaSlug", Value: 1}},
		},
		{
			Keys: bson.D{{Key: "createdAt", Value: -1}},
		},
	})
	if err != nil {
		return err
	}

	// User index: username (unique)
	_, err = d.Users.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys:    bson.D{{Key: "username", Value: 1}},
		Options: options.Index().SetUnique(true),
	})
	return err
}

func (d *Database) seedAdmin(ctx context.Context, cfg *config.Config) error {
	if cfg.AdminUsername == "" || cfg.AdminPassword == "" {
		return nil
	}

	var existingUser models.User
	err := d.Users.FindOne(ctx, bson.M{"username": cfg.AdminUsername}).Decode(&existingUser)
	if err == mongo.ErrNoDocuments {
		hashed, err := bcrypt.GenerateFromPassword([]byte(cfg.AdminPassword), bcrypt.DefaultCost)
		if err != nil {
			return err
		}

		admin := models.User{
			ID:           primitive.NewObjectID(),
			Username:     cfg.AdminUsername,
			PasswordHash: string(hashed),
			Role:         "admin",
			CreatedAt:    time.Now(),
		}

		_, err = d.Users.InsertOne(ctx, admin)
		if err != nil {
			return err
		}
		log.Printf("[DB] Initialized default admin account: %s", cfg.AdminUsername)
	}
	return nil
}
