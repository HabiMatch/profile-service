package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"path/filepath"
	"sync"

	"github.com/HabiMatch/profile-service/models"
	"github.com/HabiMatch/profile-service/utils"
)

func (h *ProfileHandler) CreateProfile(w http.ResponseWriter, r *http.Request) {
	f, fh, errs := r.FormFile("profile_picture")
	println("File: ", f)
	println("FileHeader: ", fh)
	println("Errors: ", errs)
	if errs != nil || f == nil || fh == nil {
		return
	}
	err := r.ParseMultipartForm(10 << 20) // 10 MB
	if err != nil {
		http.Error(w, "Unable to parse form data", http.StatusBadRequest)
		return
	}
	var (
		profile        models.Profile
		lat, lon       float64
		errChan        = make(chan error, 3)
		wg             sync.WaitGroup
		pictureURLChan = make(chan string)
		geoReady       = make(chan struct{})
	)

	// Goroutine 1: Upload profile picture to S3
	wg.Add(1)
	go func() {
		defer wg.Done()
		file, handler, err := r.FormFile("profile_picture")
		if err != nil {
			errChan <- fmt.Errorf("failed to upload profile picture: %v", err)
			close(pictureURLChan)
			return
		}
		defer file.Close()

		userID := r.FormValue("userid")
		if userID == "" {
			errChan <- fmt.Errorf("userid is required")
			close(pictureURLChan)
			return
		}

		fileExt := filepath.Ext(handler.Filename)
		if fileExt == "" {
			errChan <- fmt.Errorf("file must have an extension")
			close(pictureURLChan)
			return
		}
		fileName := fmt.Sprintf("profile_pictures/%s%s", userID, fileExt)
		pictureURL, err := utils.UploadToS3(file, fileName)
		if err != nil {
			errChan <- fmt.Errorf("failed to upload to S3: %v", err)
			close(pictureURLChan)
			return
		}

		pictureURLChan <- pictureURL // Send pictureURL to the channel
		close(pictureURLChan)        // Close the channel after sending
	}()

	// Goroutine 2: Serialize profile details and store profile in the database
	wg.Add(1)
	go func() {
		defer wg.Done()
		pictureURL := <-pictureURLChan // Receive pictureURL from the channel
		if pictureURL == "" {
			errChan <- fmt.Errorf("failed to receive pictureURL")
			return
		}
		profile, lat, lon, err = serializeProfileDetails(r, pictureURL)
		if err != nil {
			errChan <- fmt.Errorf("failed to serialize profile details: %v", err)
			return
		}
		fmt.Printf("Parsed JSON profile: %+v\n", profile)
		if result := h.DB.Create(&profile); result.Error != nil {
			errChan <- fmt.Errorf("failed to create profile: %v", result.Error)
			return
		}

		close(geoReady) // Signal that lat, lon are ready
		errChan <- nil
	}()

	// Goroutine 3: Store geolocation
	wg.Add(1)
	go func() {
		defer wg.Done()
		<-geoReady // Wait for lat, lon to be ready
		err := utils.StoreGeolocation(h.DB, profile.UserID, lat, lon)
		if err != nil {
			errChan <- fmt.Errorf("failed to store geolocation: %v", err)
			return
		}
		errChan <- nil
	}()

	// Wait for all goroutines to finish
	go func() {
		wg.Wait()
		close(errChan)
	}()

	for err := range errChan {
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(profile)
}

func (h *ProfileHandler) UpdateProfile(w http.ResponseWriter, r *http.Request) {
	var input models.Profile
	err := json.NewDecoder(r.Body).Decode(&input)
	if err != nil {
		http.Error(w, "failed to parse JSON", http.StatusBadRequest)
		return
	}

	// Get the profile ID from the URL
	profileID := r.FormValue("profile_id")
	if profileID == "" {
		http.Error(w, "profile_id is required", http.StatusBadRequest)
		return
	}

	// Retrieve the profile from the database
	var profile models.Profile
	if result := h.DB.First(&profile, profileID); result.Error != nil {
		http.Error(w, "failed to fetch profile", http.StatusInternalServerError)
		return
	}

	// Update the profile with the new data
	profile.Name = input.Name
	profile.Smoking = input.Smoking
	profile.Drinking = input.Drinking
	profile.Cleanliness = input.Cleanliness
	profile.WorkingHours = input.WorkingHours
	profile.Hobbies = input.Hobbies
	profile.FavoriteActivities = input.FavoriteActivities
	profile.MusicPreferences = input.MusicPreferences
	profile.GenderPreference = input.GenderPreference
	profile.PetFriendly = input.PetFriendly
	profile.FoodHabits = input.FoodHabits
	profile.Location = input.Location
	profile.Description = input.Description

	// Save the updated profile
	if result := h.DB.Save(&profile); result.Error != nil {
		http.Error(w, "failed to update profile", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(profile)
}

func (h *ProfileHandler) DeleteProfile(w http.ResponseWriter, r *http.Request) {
	// Get the profile ID from the URL
	profileID := r.FormValue("profile_id")
	if profileID == "" {
		http.Error(w, "profile_id is required", http.StatusBadRequest)
		return
	}

	// Delete the profile from the database
	if result := h.DB.Delete(&models.Profile{}, profileID); result.Error != nil {
		http.Error(w, "failed to delete profile", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"success": "true"})
}
