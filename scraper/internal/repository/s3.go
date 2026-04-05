package repository

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

type S3Storage struct {
	client   *s3.Client
	bucket   string
	endpoint string // Stored to dynamically generate the return URL
}

func NewS3Storage(ctx context.Context, endpoint, bucket, user, pass string) (*S3Storage, error) {
	cfg, err := config.LoadDefaultConfig(ctx,
		config.WithRegion("us-east-1"),
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(user, pass, "")),
	)
	if err != nil {
		return nil, fmt.Errorf("unable to load SDK config: %w", err)
	}

	client := s3.NewFromConfig(cfg, func(o *s3.Options) {
		o.BaseEndpoint = aws.String(endpoint)
		o.UsePathStyle = true
	})

	storage := &S3Storage{
		client:   client,
		bucket:   bucket,
		endpoint: endpoint,
	}

	storage.ensureBucketExists(ctx)

	return storage, nil
}

func (s *S3Storage) ensureBucketExists(ctx context.Context) {
	for i := 0; i < 30; i++ {
		_, err := s.client.HeadBucket(ctx, &s3.HeadBucketInput{Bucket: aws.String(s.bucket)})
		if err == nil {
			log.Printf("S3 Bucket '%s' exists.", s.bucket)
			return
		}

		_, err = s.client.CreateBucket(ctx, &s3.CreateBucketInput{Bucket: aws.String(s.bucket)})
		if err == nil {
			log.Printf("Created missing S3 Bucket '%s'.", s.bucket)
			return
		}

		// Log the actual error so we know why it fails if the network is down
		log.Printf("Waiting for S3 to be ready... (%d/30) - Last error: %v", i+1, err)
		time.Sleep(2 * time.Second)
	}
}

func (s *S3Storage) UploadAvatar(ctx context.Context, username string, imageData []byte) (string, error) {
	// 1. Detect the actual content type of the byte array
	contentType := http.DetectContentType(imageData)

	// 2. Assign the correct file extension based on the content type
	ext := ".jpg" // Default fallback
	switch contentType {
	case "image/png":
		ext = ".png"
	case "image/gif":
		ext = ".gif"
	case "image/webp":
		ext = ".webp"
	}

	key := fmt.Sprintf("%s%s", username, ext)

	// 3. Upload with the correct mime type
	_, err := s.client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:      aws.String(s.bucket),
		Key:         aws.String(key),
		Body:        bytes.NewReader(imageData),
		ContentType: aws.String(contentType),
	})

	if err != nil {
		return "", fmt.Errorf("failed to upload avatar: %w", err)
	}

	// 4. Return the dynamically constructed URL
	return fmt.Sprintf("%s/%s/%s", s.endpoint, s.bucket, key), nil
}
