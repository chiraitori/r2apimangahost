package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type MangaStatus string

const (
	StatusOngoing   MangaStatus = "Ongoing"
	StatusCompleted MangaStatus = "Completed"
	StatusHiatus    MangaStatus = "Hiatus"
	StatusCancelled MangaStatus = "Cancelled"
)

type Manga struct {
	ID          primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Title       string             `bson:"title" json:"title"`
	AltTitles   []string           `bson:"altTitles" json:"altTitles"`
	Slug        string             `bson:"slug" json:"slug"`
	CoverURL    string             `bson:"coverUrl" json:"coverUrl"`
	Description string             `bson:"description" json:"description"`
	Author      string             `bson:"author" json:"author"`
	Artist      string             `bson:"artist" json:"artist"`
	Status      MangaStatus        `bson:"status" json:"status"`
	Genres      []string           `bson:"genres" json:"genres"`
	Views       int64              `bson:"views" json:"views"`
	Rating      float64            `bson:"rating" json:"rating"`
	CreatedAt   time.Time          `bson:"createdAt" json:"createdAt"`
	UpdatedAt   time.Time          `bson:"updatedAt" json:"updatedAt"`

	// Virtual / Aggregated fields
	LastChapterNumber *float64 `bson:"-" json:"lastChapterNumber,omitempty"`
	ChapterCount      int      `bson:"-" json:"chapterCount,omitempty"`
}

type Chapter struct {
	ID            primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	MangaID       primitive.ObjectID `bson:"mangaId" json:"mangaId"`
	MangaSlug     string             `bson:"mangaSlug" json:"mangaSlug"`
	ChapterNumber float64            `bson:"chapterNumber" json:"chapterNumber"`
	Title         string             `bson:"title" json:"title"`
	Volume        *int               `bson:"volume,omitempty" json:"volume,omitempty"`
	Language      string             `bson:"language" json:"language"` // default "vi" or "en"
	Pages         []string           `bson:"pages" json:"pages"`       // Full URLs or relative paths
	PageCount     int                `bson:"pageCount" json:"pageCount"`
	CreatedAt     time.Time          `bson:"createdAt" json:"createdAt"`
	UpdatedAt     time.Time          `bson:"updatedAt" json:"updatedAt"`
}

// ChapterSummary is used in manga listings so page URLs are only fetched
// when a reader explicitly opens a chapter.
type ChapterSummary struct {
	ID            primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	MangaID       primitive.ObjectID `bson:"mangaId" json:"mangaId"`
	MangaSlug     string             `bson:"mangaSlug" json:"mangaSlug"`
	ChapterNumber float64            `bson:"chapterNumber" json:"chapterNumber"`
	Title         string             `bson:"title" json:"title"`
	Volume        *int               `bson:"volume,omitempty" json:"volume,omitempty"`
	Language      string             `bson:"language" json:"language"`
	PageCount     int                `bson:"pageCount" json:"pageCount"`
	CreatedAt     time.Time          `bson:"createdAt" json:"createdAt"`
	UpdatedAt     time.Time          `bson:"updatedAt" json:"updatedAt"`
}

type User struct {
	ID           primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Username     string             `bson:"username" json:"username"`
	PasswordHash string             `bson:"passwordHash" json:"-"`
	Role         string             `bson:"role" json:"role"` // "admin"
	CreatedAt    time.Time          `bson:"createdAt" json:"createdAt"`
}

// Request & Response helper structs
type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type LoginResponse struct {
	Token    string `json:"token"`
	Username string `json:"username"`
	Role     string `json:"role"`
}

type CreateMangaRequest struct {
	Title       string   `json:"title"`
	AltTitles   []string `json:"altTitles"`
	Slug        string   `json:"slug"`
	CoverURL    string   `json:"coverUrl"`
	Description string   `json:"description"`
	Author      string   `json:"author"`
	Artist      string   `json:"artist"`
	Status      string   `json:"status"`
	Genres      []string `json:"genres"`
}

type UpdateMangaRequest struct {
	Title       *string   `json:"title,omitempty"`
	AltTitles   *[]string `json:"altTitles,omitempty"`
	Slug        *string   `json:"slug,omitempty"`
	CoverURL    *string   `json:"coverUrl,omitempty"`
	Description *string   `json:"description,omitempty"`
	Author      *string   `json:"author,omitempty"`
	Artist      *string   `json:"artist,omitempty"`
	Status      *string   `json:"status,omitempty"`
	Genres      *[]string `json:"genres,omitempty"`
}

type MangaDetailResponse struct {
	Manga    Manga            `json:"manga"`
	Chapters []ChapterSummary `json:"chapters"`
}

type PaginationResponse struct {
	Page       int   `json:"page"`
	Limit      int   `json:"limit"`
	Total      int64 `json:"total"`
	TotalPages int   `json:"totalPages"`
}

type MangaListResponse struct {
	Data       []Manga            `json:"data"`
	Pagination PaginationResponse `json:"pagination"`
}

type APIResponse struct {
	Success bool        `json:"success"`
	Message string      `json:"message,omitempty"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
}
