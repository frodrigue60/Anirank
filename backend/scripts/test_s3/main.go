package main

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

func main() {
	accessKey := "minioadmin"
	secretKey := "miniopassword123"
	region := "us-east-1"
	bucket := "anirank-public"
	endpoint := "http://127.0.0.1:9000"

	opts := []func(*config.LoadOptions) error{
		config.WithRegion(region),
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(accessKey, secretKey, "")),
	}

	cfg, err := config.LoadDefaultConfig(context.Background(), opts...)
	if err != nil {
		log.Fatalf("unable to load SDK config: %v", err)
	}

	client := s3.NewFromConfig(cfg, func(o *s3.Options) {
		o.UsePathStyle = true
		o.BaseEndpoint = aws.String(endpoint)
	})

	_, err = client.PutObject(context.Background(), &s3.PutObjectInput{
		Bucket:      aws.String(bucket),
		Key:         aws.String("test_avatar.svg"),
		Body:        strings.NewReader(`<svg xmlns="http://www.w3.org/2000/svg" width="100" height="100"><circle cx="50" cy="50" r="40" stroke="green" stroke-width="4" fill="yellow" /></svg>`),
		ContentType: aws.String("image/svg+xml"),
	})

	if err != nil {
		log.Fatalf("failed to upload SVG: %v", err)
	}

	fmt.Println("Successfully uploaded test SVG to MinIO")
}
