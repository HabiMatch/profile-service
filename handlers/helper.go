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
	if input.Name == "" {
		return models.Profile{}, 0, 0, fmt.Errorf("name is required")
	}

	if input.WorkingHours == nil {
		input.WorkingHours = []byte("[]")
	}
	if input.Hobbies == nil {
		input.Hobbies = []byte("[]")
	}
	if input.FavoriteActivities == nil {
		input.FavoriteActivities = []byte("[]")
	}
	if input.MusicPreferences == nil {
		input.MusicPreferences = []byte("[]")
	}
	if input.GenderPreference == nil {
		input.GenderPreference = []byte("[]")
	}
	if input.FoodHabits == nil {
		input.FoodHabits = []byte("[]")
	}

	input.ProfilePicture = pictureURL
	return input, input.Latitude, input.Longitude, nil
}
