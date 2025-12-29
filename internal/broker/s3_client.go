package broker

import (
	"bytes"
	"context"
	"io"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

// AWSS3Client implements S3Client using AWS SDK v2
type AWSS3Client struct {
	client *s3.Client
}

// NewAWSS3Client creates a new AWS S3 client
func NewAWSS3Client(ctx context.Context) (*AWSS3Client, error) {
	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		return nil, err
	}

	return &AWSS3Client{
		client: s3.NewFromConfig(cfg),
	}, nil
}

// NewAWSS3ClientWithConfig creates a new AWS S3 client with custom config
func NewAWSS3ClientWithConfig(cfg aws.Config) *AWSS3Client {
	return &AWSS3Client{
		client: s3.NewFromConfig(cfg),
	}
}

// NewAWSS3ClientWithEndpoint creates a new AWS S3 client with custom endpoint (for LocalStack, MinIO, etc.)
func NewAWSS3ClientWithEndpoint(ctx context.Context, endpoint string, region string) (*AWSS3Client, error) {
	cfg, err := config.LoadDefaultConfig(ctx,
		config.WithRegion(region),
	)
	if err != nil {
		return nil, err
	}

	client := s3.NewFromConfig(cfg, func(o *s3.Options) {
		o.BaseEndpoint = aws.String(endpoint)
		o.UsePathStyle = true // Required for LocalStack/MinIO
	})

	return &AWSS3Client{
		client: client,
	}, nil
}

// PutObject stores an object in S3
func (c *AWSS3Client) PutObject(ctx context.Context, bucket, key string, data []byte) error {
	_, err := c.client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:      aws.String(bucket),
		Key:         aws.String(key),
		Body:        bytes.NewReader(data),
		ContentType: aws.String("application/json"),
	})
	return err
}

// GetObject retrieves an object from S3
func (c *AWSS3Client) GetObject(ctx context.Context, bucket, key string) ([]byte, error) {
	result, err := c.client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return nil, err
	}
	defer result.Body.Close()

	return io.ReadAll(result.Body)
}

// DeleteObject removes an object from S3
func (c *AWSS3Client) DeleteObject(ctx context.Context, bucket, key string) error {
	_, err := c.client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	})
	return err
}

// ListObjects lists objects with a given prefix
func (c *AWSS3Client) ListObjects(ctx context.Context, bucket, prefix string) ([]string, error) {
	var keys []string

	paginator := s3.NewListObjectsV2Paginator(c.client, &s3.ListObjectsV2Input{
		Bucket: aws.String(bucket),
		Prefix: aws.String(prefix),
	})

	for paginator.HasMorePages() {
		page, err := paginator.NextPage(ctx)
		if err != nil {
			return nil, err
		}

		for _, obj := range page.Contents {
			keys = append(keys, *obj.Key)
		}
	}

	return keys, nil
}
