package handlers

import (
	"encoding/json"
	"errors"
	"net/http"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"unicode"

	"r2apimangahost/backend/internal/models"
)

const (
	maxJSONBodyBytes       int64 = 1 << 20   // 1 MiB
	maxCoverRequestBytes   int64 = 20 << 20  // 20 MiB
	maxCoverImageBytes     int64 = 16 << 20  // 16 MiB
	maxChapterRequestBytes int64 = 220 << 20 // 220 MiB
	maxArchiveBytes        int64 = 200 << 20 // 200 MiB compressed
	maxPageImageBytes      int64 = 50 << 20  // 50 MiB per image
)

func limitRequestBody(w http.ResponseWriter, r *http.Request, maxBytes int64) {
	r.Body = http.MaxBytesReader(w, r.Body, maxBytes)
}

func writeParseError(w http.ResponseWriter, err error, fallback string) {
	var maxBytesError *http.MaxBytesError
	if errors.As(err, &maxBytesError) {
		writeError(w, http.StatusRequestEntityTooLarge, "Request body is too large")
		return
	}
	writeError(w, http.StatusBadRequest, fallback+": "+err.Error())
}

func isAllowedImageFilename(filename string) bool {
	switch strings.ToLower(filepath.Ext(filename)) {
	case ".jpg", ".jpeg", ".png", ".webp", ".gif", ".avif", ".bmp":
		return true
	default:
		return false
	}
}

func writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(data)
}

func writeSuccess(w http.ResponseWriter, data interface{}, message ...string) {
	msg := ""
	if len(message) > 0 {
		msg = message[0]
	}
	writeJSON(w, http.StatusOK, models.APIResponse{
		Success: true,
		Message: msg,
		Data:    data,
	})
}

func writeError(w http.ResponseWriter, status int, err string) {
	writeJSON(w, status, models.APIResponse{
		Success: false,
		Error:   err,
	})
}

var nonAlphanumericRegex = regexp.MustCompile(`[^a-z0-9]+`)

func slugify(s string) string {
	s = strings.ToLower(strings.TrimSpace(s))
	s = nonAlphanumericRegex.ReplaceAllString(s, "-")
	return strings.Trim(s, "-")
}

// NaturalLess provides natural sorting for filenames (e.g. page2.jpg < page10.jpg)
func naturalLess(a, b string) bool {
	chunksA := splitChunks(a)
	chunksB := splitChunks(b)

	for i := 0; i < len(chunksA) && i < len(chunksB); i++ {
		chunkA := chunksA[i]
		chunkB := chunksB[i]

		numA, errA := strconv.Atoi(chunkA)
		numB, errB := strconv.Atoi(chunkB)

		if errA == nil && errB == nil {
			if numA != numB {
				return numA < numB
			}
		} else {
			if strings.ToLower(chunkA) != strings.ToLower(chunkB) {
				return strings.ToLower(chunkA) < strings.ToLower(chunkB)
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
