package storage_service

import (
	"context"
	"fmt"
	"path/filepath"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/google/uuid"

	core_s3 "storage-service/internal/core/s3"
)

// presignTTL — на сколько действительна ссылка для загрузки. Сама загрузка
// файла с фронтенда происходит сразу после получения URL, 5 минут более
// чем достаточно и ограничивает окно, в которое ссылку можно использовать.
const presignTTL = 5 * time.Minute

// UploadURL отдаётся фронтенду: на UploadURL грузится PUT-ом тело файла
// напрямую в S3 (минуя backend), PublicURL — то, что сохраняется в
// storage-service и показывается потом как <img src>.
type UploadURL struct {
	UploadURL string `json:"upload_url"`
	PublicURL string `json:"public_url"`
}

type Service struct {
	presignClient *s3.PresignClient
	bucket        string
	publicBaseURL string
}

func NewService(ctx context.Context, cfg core_s3.Config) (*Service, error) {
	awsCfg, err := awsconfig.LoadDefaultConfig(ctx,
		awsconfig.WithRegion(cfg.Region),
		awsconfig.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(cfg.AccessKey, cfg.SecretKey, "")),
	)
	if err != nil {
		return nil, fmt.Errorf("load aws config: %w", err)
	}

	client := s3.NewFromConfig(awsCfg, func(o *s3.Options) {
		o.BaseEndpoint = aws.String(cfg.Endpoint)
		// большинство S3-совместимых провайдеров (не сам AWS) ожидают
		// path-style адресацию: https://endpoint/bucket/key, а не
		// https://bucket.endpoint/key
		o.UsePathStyle = true
	})

	return &Service{
		presignClient: s3.NewPresignClient(client),
		bucket:        cfg.Bucket,
		publicBaseURL: fmt.Sprintf("%s/%s", cfg.Endpoint, cfg.Bucket),
	}, nil
}

// NewUploadURL генерирует уникальный ключ объекта в неймспейсе владельца
// (чтобы один пользователь физически не мог перетереть файл другого даже
// при коллизии имён) и presigned PUT-ссылку на него.
func (s *Service) NewUploadURL(ctx context.Context, ownerID uuid.UUID, filename string) (UploadURL, error) {
	key := fmt.Sprintf("listings/%s/%s%s", ownerID, uuid.NewString(), filepath.Ext(filename))

	req, err := s.presignClient.PresignPutObject(ctx, &s3.PutObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
	}, s3.WithPresignExpires(presignTTL))
	if err != nil {
		return UploadURL{}, fmt.Errorf("presign put object: %w", err)
	}

	return UploadURL{
		UploadURL: req.URL,
		PublicURL: fmt.Sprintf("%s/%s", s.publicBaseURL, key),
	}, nil
}
