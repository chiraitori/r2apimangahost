package handlers

import (
	"encoding/json"
	"net/http"

	"r2apimangahost/backend/internal/auth"
	"r2apimangahost/backend/internal/config"
	"r2apimangahost/backend/internal/db"
	"r2apimangahost/backend/internal/models"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"golang.org/x/crypto/bcrypt"
)

type AuthHandler struct {
	db          *db.Database
	authService *auth.AuthService
	cfg         *config.Config
}

func NewAuthHandler(db *db.Database, authService *auth.AuthService, cfg *config.Config) *AuthHandler {
	return &AuthHandler{
		db:          db,
		authService: authService,
		cfg:         cfg,
	}
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	limitRequestBody(w, r, 64<<10)
	var req models.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeParseError(w, err, "Invalid request payload")
		return
	}

	if req.Username == "" || req.Password == "" {
		writeError(w, http.StatusBadRequest, "Username and password are required")
		return
	}

	var user models.User
	err := h.db.Users.FindOne(r.Context(), bson.M{"username": req.Username}).Decode(&user)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			writeError(w, http.StatusUnauthorized, "Invalid username or password")
			return
		}
		writeError(w, http.StatusInternalServerError, "Database error: "+err.Error())
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		writeError(w, http.StatusUnauthorized, "Invalid username or password")
		return
	}

	token, err := h.authService.GenerateToken(user.ID.Hex(), user.Username, user.Role)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to generate token")
		return
	}

	writeSuccess(w, models.LoginResponse{
		Token:    token,
		Username: user.Username,
		Role:     user.Role,
	}, "Login successful")
}

func (h *AuthHandler) Me(w http.ResponseWriter, r *http.Request) {
	claims, ok := r.Context().Value(auth.UserContextKey).(*auth.Claims)
	if !ok || claims == nil {
		writeError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	writeSuccess(w, map[string]interface{}{
		"userId":   claims.UserID,
		"username": claims.Username,
		"role":     claims.Role,
	})
}
