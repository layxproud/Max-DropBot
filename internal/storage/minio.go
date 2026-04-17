package storage

import (
	"context"
	"era-dropbot/internal/shortener"
	"io"
	"time"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

type MinioStorage struct {
	uploadClient *minio.Client // nginx:80 — used for PutObject
	urlClient    *minio.Client // localhost:80 — used only for signing, no network calls
	bucket       string
	shortener    *shortener.Shortener
}

func NewMinio(endpoint, accessKey, secretKey, bucket, publicEndpoint string, shortener *shortener.Shortener) (*MinioStorage, error) {
	newClient := func(ep string) (*minio.Client, error) {
		return minio.New(ep, &minio.Options{
			Creds:  credentials.NewStaticV4(accessKey, secretKey, ""),
			Secure: false,
			Region: "us-east-1",
		})
	}

	uploadClient, err := newClient(endpoint)
	if err != nil {
		return nil, err
	}

	urlClient, err := newClient(publicEndpoint)
	if err != nil {
		return nil, err
	}

	ctx := context.Background()
	exists, err := uploadClient.BucketExists(ctx, bucket)
	if err != nil {
		return nil, err
	}
	if !exists {
		if err = uploadClient.MakeBucket(ctx, bucket, minio.MakeBucketOptions{}); err != nil {
			return nil, err
		}
	}

	return &MinioStorage{
		uploadClient: uploadClient,
		urlClient:    urlClient,
		bucket:       bucket,
		shortener:    shortener,
	}, nil
}

func (s *MinioStorage) Upload(ctx context.Context, objectName string, reader io.Reader, size int64, contentType string) (string, error) {
	_, err := s.uploadClient.PutObject(ctx, s.bucket, objectName, reader, size, minio.PutObjectOptions{
		ContentType: contentType,
	})
	if err != nil {
		return "", err
	}

	presigned, err := s.urlClient.PresignedGetObject(ctx, s.bucket, objectName, 24*time.Hour, nil)
	if err != nil {
		return "", err
	}

	return s.shortener.Shorten(ctx, presigned.String())
}
