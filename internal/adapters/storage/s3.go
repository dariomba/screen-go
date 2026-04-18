package storage

import (
	"context"
	"fmt"
	"io"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/feature/s3/manager"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/dariomba/screen-go/internal/domain"
	"github.com/dariomba/screen-go/internal/logger"
	"github.com/dariomba/screen-go/internal/ports"
)

type S3Config struct {
	Bucket    string
	Endpoint  string
	AccessKey string
	SecretKey string
	Region    string
}

type S3Storage struct {
	client *s3.Client
	bucket string
}

func NewS3Storage(cfg S3Config) (*S3Storage, error) {
	var awsCfg aws.Config
	var err error

	if cfg.Endpoint != "" {
		awsCfg, err = config.LoadDefaultConfig(context.Background(),
			config.WithRegion(cfg.Region),
			config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(
				cfg.AccessKey,
				cfg.SecretKey,
				"",
			)),
		)
		if err != nil {
			return nil, fmt.Errorf("failed to load s3 config: %w", err)
		}
	} else {
		awsCfg, err = config.LoadDefaultConfig(context.Background(),
			config.WithRegion(cfg.Region),
		)
		if err != nil {
			return nil, fmt.Errorf("failed to load AWS config: %w", err)
		}
	}

	client := s3.NewFromConfig(awsCfg, func(o *s3.Options) {
		if cfg.Endpoint != "" {
			o.BaseEndpoint = aws.String(cfg.Endpoint)
			o.UsePathStyle = true
		}
	})

	return &S3Storage{
		client: client,
		bucket: cfg.Bucket,
	}, nil
}

func (s *S3Storage) Get(ctx context.Context, key string) (io.ReadCloser, error) {
	logger.Ctx(ctx).Debug().
		Str("bucket", s.bucket).
		Str("key", key).
		Msg("Getting object from S3")

	output, err := s.client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
	})

	if err != nil {
		// Check if it's a NoSuchKey error
		if strings.Contains(err.Error(), "NoSuchKey") {
			logger.Ctx(ctx).Debug().Msg("Object not found in S3")
			return nil, domain.ErrScreenshotNotFound
		}
		return nil, fmt.Errorf("failed to get object from S3: %w", err)
	}

	return output.Body, nil
}

func (s *S3Storage) Save(ctx context.Context, input *ports.SaveScreenshotInput) (*ports.SaveScreenshotResult, error) {
	logger.Ctx(ctx).Debug().
		Str("bucket", s.bucket).
		Str("key", input.Key).
		Str("contentType", input.ContentType).
		Msg("Uploading object to S3")

	uploader := manager.NewUploader(s.client)

	_, err := uploader.Upload(ctx, &s3.PutObjectInput{
		Bucket:      aws.String(s.bucket),
		Key:         aws.String(input.Key),
		Body:        input.Body,
		ContentType: aws.String(input.ContentType),
	})

	if err != nil {
		return nil, fmt.Errorf("failed to upload object to S3: %w", err)
	}

	headOutput, err := s.client.HeadObject(ctx, &s3.HeadObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(input.Key),
	})
	if err != nil || headOutput.ContentLength == nil {
		return nil, fmt.Errorf("failed to get object metadata from S3: %w", err)
	}

	return &ports.SaveScreenshotResult{
		Key:  input.Key,
		Size: *headOutput.ContentLength,
	}, nil
}
