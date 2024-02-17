package initializers

import (
	"os"

	"github.com/cloudinary/cloudinary-go/v2"
)

/*
* This method initiates cloudinary setup
 */
func SetupCloudinary() (*cloudinary.Cloudinary, error) {
	cldSecret := os.Getenv("CLOUD_SECRET")
	cldName := os.Getenv("CLOUD_NAME")
	cldKey := os.Getenv("CLOUD_KEY")

	cld, err := cloudinary.NewFromParams(cldName, cldKey, cldSecret)
	if err != nil {
		return nil, err
	}

	return cld, nil
}
