package utils

import (
	"backend/internal/initializers"
	"context"
	"mime/multipart"

	"github.com/google/uuid"

	"github.com/cloudinary/cloudinary-go/v2/api/uploader"
)

/*
* This method helps in file upload in cloudinary
 */
func UploadToCloudinary(file multipart.File, filePath string) (string, string, error) {
	ctx := context.Background()
	cld, err := initializers.SetupCloudinary()
	if err != nil {
		return "", "", err
	}

	fileBaseName := uuid.New().String()

	uploadParams := uploader.UploadParams{
		PublicID: fileBaseName,
	}

	result, err := cld.Upload.Upload(ctx, file, uploadParams)
	if err != nil {
		return "", "", err
	}

	imageUrl := result.SecureURL
	return imageUrl, result.PublicID, nil
}
