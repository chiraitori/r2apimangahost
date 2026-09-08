package storage

import (
	"archive/zip"
	"bytes"
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"unicode"

	"r2apimangahost/backend/internal/config"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsConfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	s3Types "github.com/aws/aws-sdk-go-v2/service/s3/types"
)

type R2Storage struct {
	client       *s3.Client
	bucket       string
	publicDomain string
}

const (
	maxObjectBytes              int64  = 50 << 20 // 50 MiB per image
	maxArchiveImageCount               = 500
	maxArchiveUncompressedBytes uint64 = 500 << 20 // 500 MiB after extraction
)

func NewR2Storage(cfg *config.Config) (*R2Storage, error) {
	if (cfg.R2AccountID == "" && cfg.S3Endpoint == "") || cfg.R2AccessKeyID == "" || cfg.R2SecretKey == "" {
		log.Println("[Storage] Warning: S3-compatible storage credentials are not fully configured.")
		return &R2Storage{
			bucket:       cfg.R2BucketName,
			publicDomain: strings.TrimRight(cfg.R2PublicDomain, "/"),
		}, nil
	}

	endpoint := cfg.S3Endpoint
	if endpoint == "" {
		endpoint = fmt.Sprintf("https://%s.r2.cloudflarestorage.com", cfg.R2AccountID)
	}
	region := cfg.S3Region
	if region == "" {
		region = "auto"
	}

	customResolver := aws.EndpointResolverWithOptionsFunc(func(service, region string, options ...interface{}) (aws.Endpoint, error) {
		return aws.Endpoint{
			URL:               endpoint,
			SigningRegion:     region,
			HostnameImmutable: true,
		}, nil
	})

	sdkConfig, err := awsConfig.LoadDefaultConfig(context.TODO(),
		awsConfig.WithEndpointResolverWithOptions(customResolver),
		awsConfig.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(
			cfg.R2AccessKeyID,
			cfg.R2SecretKey,
			"",
		)),
		awsConfig.WithRegion(region),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to load AWS config for R2: %w", err)
	}

	client := s3.NewFromConfig(sdkConfig, func(o *s3.Options) {
		o.UsePathStyle = true
	})

	publicDomain := strings.TrimRight(cfg.R2PublicDomain, "/")
	if publicDomain == "" {
		publicDomain = fmt.Sprintf("https://%s.r2.dev", cfg.R2BucketName)
	}

	return &R2Storage{
		client:       client,
		bucket:       cfg.R2BucketName,
		publicDomain: publicDomain,
	}, nil
}

func (r *R2Storage) UploadFile(ctx context.Context, key string, body io.Reader, contentType string) (string, error) {
	if r.client == nil {
		return "", fmt.Errorf("R2 storage is not configured")
	}

	// Bound every object even if a multipart or ZIP header reports a fake size.
	var buf bytes.Buffer
	size, err := io.Copy(&buf, io.LimitReader(body, maxObjectBytes+1))
	if err != nil {
		return "", fmt.Errorf("failed to read file body: %w", err)
	}
	if size > maxObjectBytes {
		return "", fmt.Errorf("object exceeds the 50 MiB image limit")
	}
	if size == 0 {
		return "", fmt.Errorf("cannot upload an empty object")
	}

	detectedType := http.DetectContentType(buf.Bytes()[:min(512, int(size))])
	if strings.HasPrefix(detectedType, "image/") {
		contentType = detectedType
	} else if !(strings.EqualFold(filepath.Ext(key), ".avif") && contentType == "image/avif") {
		return "", fmt.Errorf("object content is not a supported image (detected %s)", detectedType)
	}

	_, err = r.client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:       aws.String(r.bucket),
		Key:          aws.String(key),
		Body:         bytes.NewReader(buf.Bytes()),
		ContentType:  aws.String(contentType),
		CacheControl: aws.String("public, max-age=31536000, immutable"),
	})
	if err != nil {
		return "", fmt.Errorf("failed to upload object to R2: %w", err)
	}

	return r.GetPublicURL(key), nil
}

func (r *R2Storage) DeleteFile(ctx context.Context, key string) error {
	if r.client == nil {
		return nil
	}

	// Clean key if it contains the publicDomain prefix
	if strings.HasPrefix(key, "http://") || strings.HasPrefix(key, "https://") {
		parts := strings.Split(key, r.publicDomain+"/")
		if len(parts) > 1 {
			key = parts[1]
		}
	}

	_, err := r.client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(r.bucket),
		Key:    aws.String(key),
	})
	return err
}

// DeleteFiles removes only objects that belong to this storage's public domain
// and batches them into one R2 operation where possible.
func (r *R2Storage) DeleteFiles(ctx context.Context, locations []string) error {
	if r.client == nil || len(locations) == 0 {
		return nil
	}

	objects := make([]s3Types.ObjectIdentifier, 0, len(locations))
	prefix := r.publicDomain + "/"
	for _, location := range locations {
		key := strings.TrimLeft(location, "/")
		if strings.HasPrefix(location, "http://") || strings.HasPrefix(location, "https://") {
			if !strings.HasPrefix(location, prefix) {
				continue
			}
			key = strings.TrimPrefix(location, prefix)
		}
		if key != "" {
			objects = append(objects, s3Types.ObjectIdentifier{Key: aws.String(key)})
		}
	}

	for start := 0; start < len(objects); start += 1000 {
		end := min(start+1000, len(objects))
		_, err := r.client.DeleteObjects(ctx, &s3.DeleteObjectsInput{
			Bucket: aws.String(r.bucket),
			Delete: &s3Types.Delete{
				Objects: objects[start:end],
				Quiet:   aws.Bool(true),
			},
		})
		if err != nil {
			return err
		}
	}

	return nil
}

func (r *R2Storage) DeletePrefix(ctx context.Context, prefix string) error {
	if r.client == nil {
		return nil
	}

	// List objects with prefix
	paginator := s3.NewListObjectsV2Paginator(r.client, &s3.ListObjectsV2Input{
		Bucket: aws.String(r.bucket),
		Prefix: aws.String(prefix),
	})

	for paginator.HasMorePages() {
		page, err := paginator.NextPage(ctx)
		if err != nil {
			return err
		}

		if len(page.Contents) == 0 {
			continue
		}

		var objectsToDelete []s3Types.ObjectIdentifier
		for _, obj := range page.Contents {
			objectsToDelete = append(objectsToDelete, s3Types.ObjectIdentifier{
				Key: obj.Key,
			})
		}

		_, err = r.client.DeleteObjects(ctx, &s3.DeleteObjectsInput{
			Bucket: aws.String(r.bucket),
			Delete: &s3Types.Delete{
				Objects: objectsToDelete,
				Quiet:   aws.Bool(true),
			},
		})
		if err != nil {
			return err
		}
	}
	return nil
}

func (r *R2Storage) GetPublicURL(key string) string {
	key = strings.TrimLeft(key, "/")
	return fmt.Sprintf("%s/%s", r.publicDomain, key)
}

// ExtractAndUploadZip handles zip/cbz extraction and streams images into Cloudflare R2 directly
func (r *R2Storage) ExtractAndUploadZip(ctx context.Context, zipBytes []byte, basePath string) ([]string, error) {
	reader, err := zip.NewReader(bytes.NewReader(zipBytes), int64(len(zipBytes)))
	if err != nil {
		return nil, fmt.Errorf("invalid zip archive: %w", err)
	}

	// Filter valid image files
	var imageFiles []*zip.File
	for _, f := range reader.File {
		if f.FileInfo().IsDir() {
			continue
		}
		// Ignore hidden files and MacOS __MACOSX metadata
		if strings.HasPrefix(filepath.Base(f.Name), ".") || strings.Contains(f.Name, "__MACOSX") {
			continue
		}
		ext := strings.ToLower(filepath.Ext(f.Name))
		if isImageExtension(ext) {
			imageFiles = append(imageFiles, f)
		}
	}

	if len(imageFiles) == 0 {
		return nil, fmt.Errorf("no valid image files found inside zip archive")
	}
	if len(imageFiles) > maxArchiveImageCount {
		return nil, fmt.Errorf("archive contains too many images (maximum %d)", maxArchiveImageCount)
	}

	var totalUncompressed uint64
	for _, file := range imageFiles {
		if file.UncompressedSize64 == 0 || file.UncompressedSize64 > uint64(maxObjectBytes) {
			return nil, fmt.Errorf("image %s exceeds the 50 MiB limit or is empty", file.Name)
		}
		totalUncompressed += file.UncompressedSize64
		if totalUncompressed > maxArchiveUncompressedBytes {
			return nil, fmt.Errorf("archive expands beyond the 500 MiB safety limit")
		}
	}

	// Natural sort files by filename so pages are properly ordered 1, 2, ... 10, etc.
	sort.Slice(imageFiles, func(i, j int) bool {
		return naturalLess(imageFiles[i].Name, imageFiles[j].Name)
	})

	var pageURLs []string
	basePath = strings.Trim(basePath, "/")

	for idx, file := range imageFiles {
		rc, err := file.Open()
		if err != nil {
			return nil, fmt.Errorf("failed to read zip entry %s: %w", file.Name, err)
		}

		ext := strings.ToLower(filepath.Ext(file.Name))
		key := fmt.Sprintf("%s/%03d%s", basePath, idx+1, ext)

		contentType := mimeFromExt(ext)
		url, err := r.UploadFile(ctx, key, rc, contentType)
		rc.Close()
		if err != nil {
			return nil, fmt.Errorf("failed to upload page %d (%s) to R2: %w", idx+1, file.Name, err)
		}

		pageURLs = append(pageURLs, url)
	}

	return pageURLs, nil
}

func isImageExtension(ext string) bool {
	switch ext {
	case ".jpg", ".jpeg", ".png", ".webp", ".gif", ".avif", ".bmp":
		return true
	default:
		return false
	}
}

func mimeFromExt(ext string) string {
	switch ext {
	case ".jpg", ".jpeg":
		return "image/jpeg"
	case ".png":
		return "image/png"
	case ".webp":
		return "image/webp"
	case ".gif":
		return "image/gif"
	case ".avif":
		return "image/avif"
	default:
		return "application/octet-stream"
	}
}

// naturalLess implements natural alphanumeric sorting (e.g., page2.jpg < page10.jpg)
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

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
