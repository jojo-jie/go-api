package minio

import (
	"context"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"largefile/configs"
	"log"
	"net/url"
	"time"
)

type Minio struct {
	client *minio.Client
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
	return m.client.ListBuckets(ctx)
}

func (m *Minio) BucketExists(ctx context.Context, bucketName string) (bool, error) {
	return m.client.BucketExists(ctx, bucketName)
}

func (m *Minio) MakeBucket(ctx context.Context, bucketName string) error {
	return m.client.MakeBucket(ctx, bucketName, minio.MakeBucketOptions{})
}

func (m *Minio) RemoveBucket(ctx context.Context, bucketName string) error {
	return m.client.RemoveBucket(ctx, bucketName)
}

func (m *Minio) PreSignedPutObject(ctx context.Context, bucketName string, objectName string, expires time.Duration) (*url.URL, error) {
	return m.client.PresignedPutObject(ctx, bucketName, objectName, expires)
}

func (m *Minio) PreSignedPostPolicy(ctx context.Context, bucketName string, objectName string, expires time.Duration) (*url.URL, error) {
	policy := minio.NewPostPolicy()
	_ = policy.SetBucket(bucketName)
	_ = policy.SetKey(objectName)
	_ = policy.SetExpires(time.Now().Add(expires))
	presignedPostPolicyURL, fordata, err := m.client.PresignedPostPolicy(ctx, policy)
	log.Println(fordata)
	return presignedPostPolicyURL, err
}

func (m *Minio) GetPolicyUrl(ctx context.Context, bucketName string, objectName string, expires time.Duration, req url.Values) (*url.URL, error) {
	return m.client.PresignedGetObject(ctx, bucketName, objectName, expires, req)
}
