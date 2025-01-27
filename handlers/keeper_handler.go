package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"reflect"
	"sync"

	"github.com/HabiMatch/profile-service/models"
	"github.com/HabiMatch/profile-service/utils"
)

func (h *ProfileHandler) KeeperProfile(w http.ResponseWriter, r *http.Request) {
	// Parse JSON from the "userinfo" field
	userinfoRaw := r.FormValue("userinfo")
	fmt.Printf("Userinfo: %s\n", userinfoRaw)
	var input models.Keeper
	err := json.Unmarshal([]byte(userinfoRaw), &input)
	if err != nil {
		fmt.Printf("Error parsing userinfo JSON: %v\n", err)
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}
	if input.UserID == "" {
		http.Error(w, "UserId is required", http.StatusBadRequest)
		return
	}

	imageChannel := make(chan string, 3) // Buffer size for 3 images
	errorChannel := make(chan error, 1)  // To capture errors during upload
	var wg sync.WaitGroup

	imageCount := 0 // Counter for valid images provided

	for i := 0; i < 3; i++ { // Assuming 3 images
		f, fh, err := r.FormFile(fmt.Sprintf("room_image%d", i+1))
		if err == nil && f != nil && fh != nil {
			imageCount++
		}
		if f != nil {
			f.Close()
		}
	}

	if imageCount < 3 {
		http.Error(w, "Please provide at least 3 images", http.StatusBadRequest)
		return
	}
	// Reset form to process files again for upload
	err = r.ParseMultipartForm(10 << 20)
	if err != nil {
		http.Error(w, "Failed to parse multipart form", http.StatusInternalServerError)
		return
	}
	for i := 0; i < 3; i++ { // Assuming 3 images
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			f, fh, errs := r.FormFile(fmt.Sprintf("room_image%d", i+1))
			if errs != nil || f == nil || fh == nil {
				errorChannel <- fmt.Errorf("error reading file for room_image%d: %v", i+1, errs)
				return
			}
			defer f.Close()
			convertedFile, convertedFileName, err := utils.ConvertToJPEG(f, fmt.Sprintf("%s_%d", input.UserID, i+1))
			if err != nil {
				errorChannel <- fmt.Errorf("failed to convert file to JPEG: %v", err)
				return
			}
			defer convertedFile.Close()
			// Upload the image to S3
			pictureURL, err := utils.UploadToS3(convertedFile, os.Getenv("S3_ROOM_FOLDER_NAME"), convertedFileName)
			if err != nil {
				errorChannel <- fmt.Errorf("failed to upload to S3: %v", err)
				return
			}
			imageChannel <- pictureURL
		}(i)
	}

	go func() {
		wg.Wait()
		close(imageChannel)
		close(errorChannel)
	}()

	done := false
	for !done {
		select {
		case img, ok := <-imageChannel:
			if ok {
				fmt.Printf("Image URL: %s\n", img)
				input.FlatImages = append(input.FlatImages, img)
			} else {
				imageChannel = nil
			}
		case err, ok := <-errorChannel:
			if ok {
				fmt.Printf("Error during upload: %v\n", err)
				cleanupUploadedImages(input.FlatImages)
				http.Error(w, "Failed to upload images", http.StatusInternalServerError)
				return
			} else {
				errorChannel = nil
			}
		}

		if imageChannel == nil && errorChannel == nil {
			done = true
		}
	}

	if result := h.DB.Create(&input); result.Error != nil {
		fmt.Printf("Failed to create keeper: %v\n", result.Error)
		cleanupUploadedImages(input.FlatImages)
		http.Error(w, "Failed to create keeper", http.StatusInternalServerError)
		return
	}

	// Success response
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "Profile created successfully with %d images uploaded", len(input.FlatImages))
}

func cleanupUploadedImages(uploadedImages []string) {
	for _, imgURL := range uploadedImages {
		err := utils.DeleteFromS3(imgURL)
		if err != nil {
			fmt.Printf("Failed to delete image from S3: %v\n", err)
		} else {
			fmt.Printf("Successfully deleted image: %s\n", imgURL)
		}
	}
}

func (h *ProfileHandler) UpdateKeeperProfile(w http.ResponseWriter, r *http.Request) {
	userinfoRaw := r.FormValue("userinfo")
	var input models.Keeper
	// Parse the JSON input
	if err := json.Unmarshal([]byte(userinfoRaw), &input); err != nil {
		http.Error(w, "Invalid JSON input", http.StatusBadRequest)
		return
	}
	if input.UserID == "" {
		http.Error(w, "UserId is required", http.StatusBadRequest)
		return
	}
	// Fetch existing keeper
	var keeper models.Keeper
	if result := h.DB.First(&keeper, "user_id = ?", input.UserID); result.Error != nil {
		http.Error(w, "Profile not found", http.StatusNotFound)
		return
	}
	inputValue := reflect.ValueOf(input)
	profileValue := reflect.ValueOf(&keeper).Elem()
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

	// Save the updated keeper
	if err := h.DB.Save(&keeper).Error; err != nil {
		http.Error(w, "Failed to update keeper", http.StatusInternalServerError)
		return
	}

	// Respond with the updated keeper
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(keeper)
}

func (h *ProfileHandler) DeleteKeeperProfile(w http.ResponseWriter, r *http.Request) {
	userID := r.FormValue("userid")
	if userID == "" {
		http.Error(w, "UserId is required", http.StatusBadRequest)
		return
	}
	var keeper models.Keeper
	if result := h.DB.First(&keeper, "user_id = ?", userID); result.Error != nil {
		http.Error(w, "Profile not found", http.StatusNotFound)
		return
	}
	if err := h.DB.Delete(&keeper).Error; err != nil {
		http.Error(w, "Failed to delete keeper", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "Profile deleted successfully")
}
