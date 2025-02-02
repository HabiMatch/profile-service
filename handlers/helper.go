package handlers

import (
	"fmt"

	"github.com/HabiMatch/profile-service/models"
	"github.com/HabiMatch/profile-service/utils"
)

func serializeProfileDetails(input models.Profile, pictureURL string) (models.Profile, float64, float64, error) {

	if input.UserID == "" {
		return models.Profile{}, 0, 0, fmt.Errorf("Userid is required")
	}
	if input.FirstName == "" {
		return models.Profile{}, 0, 0, fmt.Errorf("FirstName is required")
	}

	if input.LastName == "" {
		return models.Profile{}, 0, 0, fmt.Errorf("LastName is required")
	}
	if input.Selftags == nil {
		input.Selftags = []string{}
	}
	if input.Gender == "" {
		return models.Profile{}, 0, 0, fmt.Errorf("Gender is required")
	}
	if input.Occupation == "" {
		return models.Profile{}, 0, 0, fmt.Errorf("Occupation is required")
	}
	if input.Address == "" {
		return models.Profile{}, 0, 0, fmt.Errorf("Address is required")
	}
	if input.Contactno == "" {
		return models.Profile{}, 0, 0, fmt.Errorf("Contactno is required")
	}
	if input.Description == "" {
		return models.Profile{}, 0, 0, fmt.Errorf("Description is required")
	}
	input.ProfilePicture = pictureURL
	return input, input.Latitude, input.Longitude, nil
}

func cleanupUploadedImages(images []string) {
	for _, image := range images {
		err := utils.DeleteFromS3(image)
		if err != nil {
			fmt.Printf("Error removing image: %v\n", err)
		}
	}
}

func GetProfileType(profileType ProfileType) (interface{}, error) {
	var input interface{}
	switch profileType {
	case KeeperProfileType:
		input = &models.Keeper{}
	case SeekerProfileType:
		input = &models.Seeker{}
	case GeneralProfileType:
		input = &models.Profile{}
	default:
		return nil, fmt.Errorf("Invalid profile type")
	}
	return input, nil

}
