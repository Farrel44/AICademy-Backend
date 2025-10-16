package utils

import (
	"context"
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/google/uuid"
)

type R2Client struct {
	s3Client *s3.Client
	bucket   string
	baseURL  string
}

func NewR2Client() (*R2Client, error) {
	bucketName := os.Getenv("R2_BUCKET_NAME")
	accountId := os.Getenv("R2_ACCOUNT_ID")
	accessKeyId := os.Getenv("R2_KEY_ID")
	accessKeySecret := os.Getenv("ACCESS_KEY_SECRET")
	baseURL := os.Getenv("OBJECT_STORAGE_URL")

	cfg, err := config.LoadDefaultConfig(context.TODO(),
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(accessKeyId, accessKeySecret, "")),
		config.WithRegion("auto"),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to load AWS config: %v", err)
	}

	client := s3.NewFromConfig(cfg, func(o *s3.Options) {
		o.BaseEndpoint = aws.String(fmt.Sprintf("https://%s.r2.cloudflarestorage.com", accountId))
	})

	return &R2Client{
		s3Client: client,
		bucket:   bucketName,
		baseURL:  baseURL,
	}, nil
}

func (r *R2Client) UploadFile(file *multipart.FileHeader, folder string) (string, error) {
	src, err := file.Open()
	if err != nil {
		return "", err
	}
	defer src.Close()

	ext := filepath.Ext(file.Filename)
	filename := fmt.Sprintf("%s/%s%s", folder, uuid.New().String(), ext)

	_, err = r.s3Client.PutObject(context.TODO(), &s3.PutObjectInput{
		Bucket: aws.String(r.bucket),
		Key:    aws.String(filename),
		Body:   src,
	})
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("%s/%s", r.baseURL, filename), nil
}

func (r *R2Client) UploadFileFromReader(reader io.Reader, filename string, folder string) (string, error) {
	key := fmt.Sprintf("%s/%s", folder, filename)

	_, err := r.s3Client.PutObject(context.TODO(), &s3.PutObjectInput{
		Bucket: aws.String(r.bucket),
		Key:    aws.String(key),
		Body:   reader,
	})
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("%s/%s", r.baseURL, key), nil
}

func (r *R2Client) DeleteFile(key string) error {
	_, err := r.s3Client.DeleteObject(context.TODO(), &s3.DeleteObjectInput{
		Bucket: aws.String(r.bucket),
		Key:    aws.String(key),
	})
	return err
}
