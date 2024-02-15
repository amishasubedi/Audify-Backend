package utils

import (
	"context"
	"log"
	"os"

	"github.com/cloudinary/cloudinary-go/v2"
	"github.com/cloudinary/cloudinary-go/v2/api/uploader"
)

/*
* This method initiates cloudinary setup
 */
func CloudSetup() *cloudinary.Cloudinary {
	cld, err := cloudinary.NewFromParams(os.Getenv("CLOUD_NAME"), os.Getenv("CLOUD_KEY"), os.Getenv("CLOUD_SECRET"))
	if err != nil {
		log.Fatalf("Failed to initialize Cloudinary: %v", err)
	}
	return cld
}

/*
* This method uploads a file to Cloudinary and returns the URL
 */
func UploadFileToCloudinary(cld *cloudinary.Cloudinary, filePath string) (string, error) {
	ctx := context.Background() // Create a background context
	uploadResult, err := cld.Upload.Upload(ctx, filePath, uploader.UploadParams{})
	if err != nil {
		return "", err
	}
	return uploadResult.SecureURL, nil
}
