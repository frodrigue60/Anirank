package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"

	"anirank/api/internal/infrastructure"
	"github.com/joho/godotenv"
)

func main() {
	// Try to load .env from common locations
	_ = godotenv.Load(".env")       // backend/.env
	_ = godotenv.Load("../.env")    // root/.env (if running from backend)
	_ = godotenv.Load("../../.env") // root/.env (if running from backend/scripts/...)

	accountID := os.Getenv("R2_ACCOUNT_ID")
	accessKey := os.Getenv("R2_ACCESS_KEY_ID")
	secretKey := os.Getenv("R2_SECRET_ACCESS_KEY")
	bucket := os.Getenv("R2_BUCKET_NAME")
	publicURL := os.Getenv("R2_PUBLIC_URL")

	if accountID == "" || accessKey == "" || secretKey == "" || bucket == "" {
		cwd, _ := os.Getwd()
		log.Fatalf("Missing required R2 environment variables. (CWD: %s). Please check your .env file and ensure R2_ACCOUNT_ID, R2_ACCESS_KEY_ID, R2_SECRET_ACCESS_KEY, and R2_BUCKET_NAME are set.", cwd)
	}

	endpoint := fmt.Sprintf("https://%s.r2.cloudflarestorage.com", accountID)
	region := "auto" // Cloudflare R2 uses "auto" for region

	fmt.Printf("--- R2 Configuration Test ---\n")
	fmt.Printf("Account ID: %s\n", accountID)
	fmt.Printf("Bucket:     %s\n", bucket)
	fmt.Printf("Endpoint:   %s\n", endpoint)
	fmt.Printf("-----------------------------\n\n")

	ctx := context.Background()
	storage, err := infrastructure.NewS3Storage(ctx, accessKey, secretKey, region, bucket, endpoint, publicURL)
	if err != nil {
		log.Fatalf("Failed to initialize R2 storage: %v", err)
	}

	// 1. Test Upload
	testFileName := "r2-test-connection.txt"
	content := "AniRank R2 Connection Test - " + fmt.Sprint(os.Getpid())
	fmt.Printf("[1/3] Uploading test file: %s... ", testFileName)
	
	_, err = storage.UploadFile(ctx, testFileName, strings.NewReader(content), int64(len(content)), "text/plain")
	if err != nil {
		fmt.Printf("FAILED\n")
		log.Fatalf("Error: %v", err)
	}
	fmt.Printf("SUCCESS\n")

	// 2. Test Listing
	fmt.Printf("[2/3] Listing files in bucket... ")
	files, err := storage.ListFiles(ctx, "r2-test-")
	if err != nil {
		fmt.Printf("FAILED\n")
		log.Fatalf("Error: %v", err)
	}
	
	found := false
	for _, f := range files {
		if f == testFileName {
			found = true
			break
		}
	}

	if found {
		fmt.Printf("SUCCESS (Found %d file(s))\n", len(files))
	} else {
		fmt.Printf("FAILED (File not found in list)\n")
		log.Fatalf("File listing did not include the test file.")
	}

	// 3. Test URL Generation
	genURL := storage.GetURL(testFileName)
	fmt.Printf("Generated URL: %s\n", genURL)
	if publicURL != "" && !strings.HasPrefix(genURL, publicURL) {
		fmt.Printf("WARNING: Generated URL does not match R2_PUBLIC_URL prefix (%s)\n", publicURL)
	}

	// 4. Cleanup
	fmt.Printf("[3/3] Cleaning up test file... ")
	err = storage.DeleteFile(ctx, testFileName)
	if err != nil {
		fmt.Printf("FAILED\n")
		log.Printf("Warning: Failed to delete test file: %v", err)
	} else {
		fmt.Printf("SUCCESS\n")
	}

	fmt.Println("\n--- R2 TEST PASSED SUCCESSFULLY ---")
}
