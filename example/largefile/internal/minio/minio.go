package minio

import (
	"context"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"largefile/configs"
)

type Minio struct {
	Client *minio.Client
}

func New(c *configs.Config) (*Minio, error) {
	minioClient, err := minio.New(c.Minio.Endpoint, &minio.Options{
		Creds: credentials.NewStaticV4(c.Minio.AccessKey, c.Minio.SecretKey, ""),
	})
	if err != nil {
		return nil, err
	}
	return &Minio{minioClient}, nil
}

func (m *Minio) ListBuckets(ctx context.Context) ([]minio.BucketInfo, error) {
	return m.Client.ListBuckets(ctx)
}

func (m *Minio) BucketExists(ctx context.Context, bucketName string) (bool, error) {
	return m.Client.BucketExists(ctx, bucketName)
}

func (m *Minio) MakeBucket(ctx context.Context, bucketName string) error {
	return m.Client.MakeBucket(ctx, bucketName, minio.MakeBucketOptions{})
}

func (m *Minio) removeBucket(ctx context.Context, bucketName string) error {
	return m.Client.RemoveBucket(ctx, bucketName)
}
