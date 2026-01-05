package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/jt-chihara/yakusoku/internal/broker"
)

func main() {
	port := flag.Int("port", 8080, "Port to listen on")
	storageType := flag.String("storage", "memory", "Storage type: memory or s3")
	s3Bucket := flag.String("s3-bucket", "", "S3 bucket name (required when storage=s3)")
	s3Prefix := flag.String("s3-prefix", "yakusoku/", "S3 key prefix")
	s3Endpoint := flag.String("s3-endpoint", "", "S3 endpoint URL (for LocalStack/MinIO)")
	s3Region := flag.String("s3-region", "ap-northeast-1", "AWS region")
	flag.Parse()

	token := os.Getenv("YAKUSOKU_BROKER_TOKEN")
	if token == "" {
		log.Fatal("YAKUSOKU_BROKER_TOKEN environment variable is required")
	}

	var storage broker.Storage

	switch *storageType {
	case "memory":
		storage = broker.NewMemoryStorage()
		fmt.Println("Using in-memory storage (data will be lost on restart)")
	case "s3":
		if *s3Bucket == "" {
			log.Fatal("--s3-bucket is required when using S3 storage")
		}

		ctx := context.Background()
		var s3Client *broker.AWSS3Client
		var err error

		if *s3Endpoint != "" {
			// Use custom endpoint (LocalStack/MinIO)
			s3Client, err = broker.NewAWSS3ClientWithEndpoint(ctx, *s3Endpoint, *s3Region)
			if err != nil {
				log.Fatalf("Failed to create S3 client: %v", err)
			}
			fmt.Printf("Using S3 storage with custom endpoint: %s\n", *s3Endpoint)
		} else {
			// Use default AWS credentials
			s3Client, err = broker.NewAWSS3Client(ctx)
			if err != nil {
				log.Fatalf("Failed to create S3 client: %v", err)
			}
			fmt.Println("Using S3 storage with AWS credentials")
		}

		storage = broker.NewS3Storage(s3Client, *s3Bucket, *s3Prefix)
		fmt.Printf("S3 bucket: %s, prefix: %s\n", *s3Bucket, *s3Prefix)
	default:
		log.Fatalf("Unknown storage type: %s (use 'memory' or 's3')", *storageType)
	}

	api := broker.NewAPI(storage)
	handler := broker.WrapWithAuth(token, api.Handler())

	addr := fmt.Sprintf(":%d", *port)
	fmt.Printf("\nYakusoku Broker starting on %s\n", addr)
	fmt.Println("Authentication: enabled (Bearer token required)")
	fmt.Println("\nAvailable endpoints:")
	fmt.Println("  GET  /pacts                                                    - List all contracts")
	fmt.Println("  GET  /pacts/provider/{provider}                                - Get contracts by provider")
	fmt.Println("  GET  /pacts/provider/{provider}/consumer/{consumer}/version/{version} - Get specific contract")
	fmt.Println("  GET  /pacts/provider/{provider}/consumer/{consumer}/latest     - Get latest contract")
	fmt.Println("  POST /pacts/provider/{provider}/consumer/{consumer}/version/{version} - Publish contract")
	fmt.Println("  DELETE /pacts/provider/{provider}/consumer/{consumer}/version/{version} - Delete contract")
	fmt.Println("  POST /pacts/provider/{provider}/consumer/{consumer}/version/{version}/verification-results - Record verification")
	fmt.Println("  GET  /matrix                                                   - Can I deploy check")
	fmt.Println("")
	fmt.Println("  GET  /ui                                                       - Web UI")

	if err := http.ListenAndServe(addr, handler); err != nil {
		log.Fatal(err)
	}
}
