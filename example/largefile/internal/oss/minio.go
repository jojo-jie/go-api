package oss

import (
	"context"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"io"
	"largefile/configs"
	"net/url"
	"time"
)

type Minio struct {
	client *minio.Core
}

func New(c *configs.Config) (*Minio, error) {
	minioClient, err := minio.NewCore(c.Minio.Endpoint, &minio.Options{
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

func (m *Minio) RemoveObject(ctx context.Context, bucketName string, objectName string, removeObjectOptions minio.RemoveObjectOptions) error {
	return m.client.RemoveObject(ctx, bucketName, objectName, removeObjectOptions)
}

func (m *Minio) PreSignedPutObject(ctx context.Context, bucketName string, objectName string, expires time.Duration) (*url.URL, error) {
	return m.client.PresignedPutObject(ctx, bucketName, objectName, expires)
}

func (m *Minio) PreSignedPostPolicy(ctx context.Context, bucketName string, objectName string, expires time.Duration) (*url.URL, map[string]string, error) {
	policy := minio.NewPostPolicy()
	_ = policy.SetBucket(bucketName)
	_ = policy.SetKey(objectName)
	_ = policy.SetExpires(time.Now().Add(expires))
	return m.client.PresignedPostPolicy(ctx, policy)
}

func (m *Minio) StatObject(ctx context.Context, bucketName string, objectName string, opts minio.StatObjectOptions) (minio.ObjectInfo, error) {
	return m.client.StatObject(ctx, bucketName, objectName, opts)
}

func (m *Minio) ListObjects(bucket, prefix, marker, delimiter string, maxKeys int) (minio.ListBucketResult, error) {
	return m.client.ListObjects(bucket, prefix, marker, delimiter, maxKeys)
}

func (m *Minio) GetPolicyUrl(ctx context.Context, bucketName string, objectName string, expires time.Duration, req url.Values) (*url.URL, error) {
	return m.client.PresignedGetObject(ctx, bucketName, objectName, expires, req)
}

func (m *Minio) NewMultipartUpload(ctx context.Context, bucketName string, objectName string, opts minio.PutObjectOptions) (string, error) {
	return m.client.NewMultipartUpload(ctx, bucketName, objectName, opts)
}

func (m *Minio) ListMultipartUploads(ctx context.Context, bucket, prefix, keyMarker, uploadIDMarker, delimiter string, maxUploads int) (result minio.ListMultipartUploadsResult, err error) {
	return m.client.ListMultipartUploads(ctx, bucket, prefix, keyMarker, uploadIDMarker, delimiter, maxUploads)
}

func (m *Minio) PutObjectPart(ctx context.Context, bucket, object, uploadID string, partID int,
	data io.Reader, size int64, opts minio.PutObjectPartOptions) (minio.ObjectPart, error) {
	return m.client.PutObjectPart(ctx, bucket, object, uploadID, partID, data, size, opts)
}

func (m *Minio) CompleteMultipartUploadFinish(ctx context.Context, bucket, object, uploadID string, parts []minio.CompletePart, opts minio.PutObjectOptions) (minio.UploadInfo, error) {
	return m.client.CompleteMultipartUpload(ctx, bucket, object, uploadID, parts, opts)
}
