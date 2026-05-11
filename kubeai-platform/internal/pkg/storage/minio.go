package storage

import (
	"context"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"io"
)

type MinIOClient struct {
	Client *minio.Client
	Bucket string
}

func NewMinIOClient(endpoint, accessKey, secretKey, bucket string, useSSL bool) (*MinIOClient, error) {
	client, err := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(accessKey, secretKey, ""),
		Secure: useSSL,
	})
	if err != nil {
		return nil, err
	}
	//确保 bucket存在
	ctx := context.Background()
	exists, err := client.BucketExists(ctx, bucket)
	if err != nil {
		return nil, err
	}
	if !exists {
		err := client.MakeBucket(ctx, bucket, minio.MakeBucketOptions{})
		if err != nil {
			return nil, err
		}
	}
	return &MinIOClient{Client: client, Bucket: bucket}, nil
}

// UploadFile 上传文件到指定路径
func (m *MinIOClient) UploadFile(ctx context.Context, objectPath string, reader io.Reader, size int64,
	contentType string) error {
	_, err := m.Client.PutObject(ctx, m.Bucket, objectPath, reader, size, minio.PutObjectOptions{
		ContentType: contentType,
	})
	return err
}

// DownloadFile 获取文件 reader
func (m *MinIOClient) DownloadFile(ctx context.Context, objectPath string) (io.ReadCloser, error) {
	object, err := m.Client.GetObject(ctx, m.Bucket, objectPath, minio.GetObjectOptions{})
	if err != nil {
		return nil, err
	}
	return object, nil
}
