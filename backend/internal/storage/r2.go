package storage

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
)

type R2Client struct {
	client     *s3.Client
	bucketName string
	endpoint   string
}

func NewR2Client() (*R2Client, error) {
	accountID := os.Getenv("R2_ACCOUNT_ID")
	accessKey := os.Getenv("R2_ACCESS_KEY_ID")
	secretKey := os.Getenv("R2_SECRET_ACCESS_KEY")
	bucketName := os.Getenv("R2_BUCKET_NAME")

	if accountID == "" || accessKey == "" || secretKey == "" || bucketName == "" {
		return nil, fmt.Errorf("R2 credentials not configured: R2_ACCOUNT_ID, R2_ACCESS_KEY_ID, R2_SECRET_ACCESS_KEY, and R2_BUCKET_NAME are required")
	}

	endpoint := fmt.Sprintf("https://%s.r2.cloudflarestorage.com", accountID)

	cfg, err := config.LoadDefaultConfig(context.Background(),
		config.WithRegion("auto"),
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(accessKey, secretKey, "")),
		config.WithEndpointResolverWithOptions(aws.EndpointResolverWithOptionsFunc(func(service, region string, options ...interface{}) (aws.Endpoint, error) {
			return aws.Endpoint{
				URL:               endpoint,
				HostnameImmutable: true,
				Source:            aws.EndpointSourceCustom,
			}, nil
		})),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to load AWS config: %w", err)
	}

	client := s3.NewFromConfig(cfg, func(o *s3.Options) {
		o.UsePathStyle = true
	})

	return &R2Client{
		client:     client,
		bucketName: bucketName,
		endpoint:   endpoint,
	}, nil
}

func (c *R2Client) UploadFile(ctx context.Context, key string, data []byte, contentType string) error {
	if c == nil {
		return fmt.Errorf("R2 client not initialized")
	}

	input := &s3.PutObjectInput{
		Bucket:      aws.String(c.bucketName),
		Key:         aws.String(key),
		Body:        bytes.NewReader(data),
		ContentType: aws.String(contentType),
	}

	_, err := c.client.PutObject(ctx, input)
	if err != nil {
		return fmt.Errorf("failed to upload file: %w", err)
	}

	return nil
}

func (c *R2Client) DeleteFile(ctx context.Context, key string) error {
	if c == nil {
		return fmt.Errorf("R2 client not initialized")
	}

	input := &s3.DeleteObjectInput{
		Bucket: aws.String(c.bucketName),
		Key:    aws.String(key),
	}

	_, err := c.client.DeleteObject(ctx, input)
	if err != nil {
		return fmt.Errorf("failed to delete file: %w", err)
	}

	return nil
}

func (c *R2Client) GetPresignedURL(ctx context.Context, key string, expiryMinutes int) (string, error) {
	if c == nil {
		return "", fmt.Errorf("R2 client not initialized")
	}

	presignClient := s3.NewPresignClient(c.client)

	input := &s3.GetObjectInput{
		Bucket: aws.String(c.bucketName),
		Key:    aws.String(key),
	}

	req, err := presignClient.PresignGetObject(ctx, input, s3.WithPresignExpires(time.Duration(expiryMinutes)*time.Minute))
	if err != nil {
		return "", fmt.Errorf("failed to generate presigned URL: %w", err)
	}

	return req.URL, nil
}

func (c *R2Client) FileExists(ctx context.Context, key string) (bool, error) {
	if c == nil {
		return false, fmt.Errorf("R2 client not initialized")
	}

	input := &s3.HeadObjectInput{
		Bucket: aws.String(c.bucketName),
		Key:    aws.String(key),
	}

	_, err := c.client.HeadObject(ctx, input)
	if err != nil {
		var notFoundErr *types.NotFound
		var noSuchKey *types.NoSuchKey
		if errors.As(err, &notFoundErr) || errors.As(err, &noSuchKey) {
			return false, nil
		}
		return false, fmt.Errorf("failed to check file existence: %w", err)
	}

	return true, nil
}
