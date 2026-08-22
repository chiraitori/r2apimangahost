package handlers

import (
	"encoding/json"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"unicode"

	"r2apimangahost/backend/internal/models"
)

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
