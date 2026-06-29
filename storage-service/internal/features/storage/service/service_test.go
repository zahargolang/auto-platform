package storage_service

import (
	"context"
	"strings"
	"testing"

	"github.com/google/uuid"

	core_s3 "storage-service/internal/core/s3"
)

// Presigning — чистое локальное HMAC-подписывание URL, без сетевого
// похода в S3, поэтому тестируется без реального endpoint'а/credentials.
func testService(t *testing.T) *Service {
	t.Helper()

	svc, err := NewService(context.Background(), core_s3.Config{
		Endpoint:  "http://localhost:9000",
		Region:    "us-east-1",
		Bucket:    "test-bucket",
		AccessKey: "test-access-key",
		SecretKey: "test-secret-key",
	})
	if err != nil {
		t.Fatalf("NewService() error = %v", err)
	}
	return svc
}

func TestNewUploadURL_PublicURLShape(t *testing.T) {
	svc := testService(t)
	ownerID := uuid.New()

	result, err := svc.NewUploadURL(context.Background(), ownerID, "photo.jpg")
	if err != nil {
		t.Fatalf("NewUploadURL() error = %v", err)
	}

	wantPrefix := "http://localhost:9000/test-bucket/listings/" + ownerID.String() + "/"
	if !strings.HasPrefix(result.PublicURL, wantPrefix) {
		t.Fatalf("PublicURL = %q, want prefix %q", result.PublicURL, wantPrefix)
	}
	if !strings.HasSuffix(result.PublicURL, ".jpg") {
		t.Fatalf("PublicURL = %q, want suffix .jpg", result.PublicURL)
	}
}

func TestNewUploadURL_IsPresigned(t *testing.T) {
	svc := testService(t)

	result, err := svc.NewUploadURL(context.Background(), uuid.New(), "photo.png")
	if err != nil {
		t.Fatalf("NewUploadURL() error = %v", err)
	}

	if !strings.Contains(result.UploadURL, "X-Amz-Signature") {
		t.Fatalf("UploadURL = %q, want a presigned (X-Amz-Signature) URL", result.UploadURL)
	}
}

func TestNewUploadURL_UniqueKeysPerCall(t *testing.T) {
	svc := testService(t)
	ownerID := uuid.New()

	first, err := svc.NewUploadURL(context.Background(), ownerID, "a.jpg")
	if err != nil {
		t.Fatalf("NewUploadURL() error = %v", err)
	}
	second, err := svc.NewUploadURL(context.Background(), ownerID, "a.jpg")
	if err != nil {
		t.Fatalf("NewUploadURL() error = %v", err)
	}

	if first.PublicURL == second.PublicURL {
		t.Fatalf("two calls produced the same key: %q", first.PublicURL)
	}
}
