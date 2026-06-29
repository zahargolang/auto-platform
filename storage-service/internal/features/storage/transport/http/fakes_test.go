package transport_http

import (
	"context"

	"github.com/google/uuid"

	storage_service "storage-service/internal/features/storage/service"
)

type fakeService struct {
	newUploadURLFunc func(ctx context.Context, ownerID uuid.UUID, filename string) (storage_service.UploadURL, error)
}

func (f *fakeService) NewUploadURL(ctx context.Context, ownerID uuid.UUID, filename string) (storage_service.UploadURL, error) {
	return f.newUploadURLFunc(ctx, ownerID, filename)
}
