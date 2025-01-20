package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/HabiMatch/profile-service/models"
)

func serializeProfileDetails(r *http.Request, pictureURL string) (models.Profile, float64, float64, error) {
	userinfoRaw := r.FormValue("userinfo")
	var input models.Profile
	err := json.Unmarshal([]byte(userinfoRaw), &input)
	if err != nil {
		fmt.Printf("Error parsing userinfo JSON: %v\n", err)
		return models.Profile{}, 0, 0, err
	}
	if input.UserID == "" {
		if r.FormValue("userid") == "" {
			return models.Profile{}, 0, 0, fmt.Errorf("userid is required")
		}
		input.UserID = r.FormValue("userid")
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
	if input.Gender == false {
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
