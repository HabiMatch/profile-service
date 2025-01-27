package handlers

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"reflect"
	"strings"
	"sync"

	"github.com/HabiMatch/profile-service/models"
	"github.com/HabiMatch/profile-service/utils"
)

func (h *ProfileHandler) CreateProfile(w http.ResponseWriter, r *http.Request) {
	f, fh, errs := r.FormFile("profile_picture")
	if errs != nil || f == nil || fh == nil {
		http.Error(w, "Provide Profile Picture", http.StatusBadRequest)
		return
	}
	err := r.ParseMultipartForm(10 << 20) // 10 MB
	if err != nil {
		http.Error(w, "Unable to parse form data", http.StatusBadRequest)
		return
	}
	// Parse UserInfo
	userinfoRaw := r.FormValue("userinfo")
	var Profile models.Profile
	erro := json.Unmarshal([]byte(userinfoRaw), &Profile)
	if erro != nil {
		fmt.Printf("Error parsing userinfo JSON: %v\n", err)
		http.Error(w, "Unable to parse userinfo", http.StatusBadRequest)
		return
	}
	if Profile.UserID == "" {
		http.Error(w, "UserId is required", http.StatusBadRequest)
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
		file, _, err := r.FormFile("profile_picture")
		if err != nil {
			errChan <- fmt.Errorf("failed to upload profile picture: %v", err)
			close(pictureURLChan)
			return
		}
		defer file.Close()

		userID := Profile.UserID
		userID = strings.ReplaceAll(userID, " ", "")
		if userID == "" {
			errChan <- fmt.Errorf("userid is required")
			close(pictureURLChan)
			return
		}

		// Verify if the file is an image
		isImage, err := utils.IsImageFile(file)
		if err != nil || !isImage {
			errChan <- fmt.Errorf("uploaded file is not a valid image")
			close(pictureURLChan)
			return
		}

		// Reset file pointer to the start for processing
		if _, err := file.Seek(0, io.SeekStart); err != nil {
			errChan <- fmt.Errorf("failed to reset file pointer: %v", err)
			close(pictureURLChan)
			return
		}

		// Convert to JPEG
		convertedFile, convertedFileName, err := utils.ConvertToJPEG(file, userID)
		if err != nil {
			errChan <- fmt.Errorf("failed to convert file to JPEG: %v", err)
			close(pictureURLChan)
			return
		}
		defer convertedFile.Close()

		// Upload to S3
		pictureURL, err := utils.UploadToS3(convertedFile, os.Getenv("S3_PROFILE_FOLDER_NAME"), convertedFileName)
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
		profile, lat, lon, err = serializeProfileDetails(Profile, pictureURL)
		if err != nil {
			errChan <- fmt.Errorf("failed to serialize profile details: %v", err)
			return
		}
		fmt.Printf("Parsed JSON profile: %+v\n", profile)
		if result := h.DB.Create(&profile); result.Error != nil {
			errChan <- fmt.Errorf("failed to create profile: %v", result.Error)
			var temp []string
			temp = append(temp, pictureURL)
			cleanupUploadedImages(temp)
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
	userinfoRaw := r.FormValue("userinfo")
	var input models.Profile

	// Parse the JSON input
	if err := json.Unmarshal([]byte(userinfoRaw), &input); err != nil {
		http.Error(w, "Invalid JSON input", http.StatusBadRequest)
		return
	}
	if input.UserID == "" {
		http.Error(w, "UserId is required", http.StatusBadRequest)
		return
	}
	// Fetch existing profile
	var profile models.Profile
	if result := h.DB.First(&profile, "user_id = ?", input.UserID); result.Error != nil {
		http.Error(w, "Profile not found", http.StatusNotFound)
		return
	}

	inputValue := reflect.ValueOf(input)
	profileValue := reflect.ValueOf(&profile).Elem()
	inputType := inputValue.Type()

	for i := 0; i < inputValue.NumField(); i++ {
		fieldValue := inputValue.Field(i)
		fieldType := inputType.Field(i)

		// Skip zero (default) values for non-pointer fields
		if fieldValue.Kind() == reflect.Ptr || !fieldValue.IsZero() {
			profileField := profileValue.FieldByName(fieldType.Name)
			if profileField.IsValid() && profileField.CanSet() {
				profileField.Set(fieldValue)
			}
		}
	}

	// Save the updated profile
	if err := h.DB.Save(&profile).Error; err != nil {
		http.Error(w, "Failed to update profile", http.StatusInternalServerError)
		return
	}

	// Respond with the updated profile
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(profile)
}

// DeleteProfile function is for server its not exposed to the client
func (h *ProfileHandler) DeleteProfile(w http.ResponseWriter, r *http.Request) {
	// Get the profile ID from the URL
	profileID := r.FormValue("user_id")
	if profileID == "" {
		http.Error(w, "userid is required", http.StatusBadRequest)
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
