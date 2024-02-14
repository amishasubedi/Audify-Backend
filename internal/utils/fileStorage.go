package utils

import (
	"log"
	"os"

	"github.com/cloudinary/cloudinary-go/v2"
)

func CloudSetup() *cloudinary.Cloudinary {
	cld, err := cloudinary.NewFromParams(os.Getenv("CLOUD_NAME"), os.Getenv("CLOUD_KEY"), os.Getenv("CLOUD_SECRET"))
	if err != nil {
		log.Fatalf("Failed to initialize Cloudinary: %v", err)
	}
	return cld
}
