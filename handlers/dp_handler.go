package handlers

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"sync"

	"github.com/HabiMatch/profile-service/models"
	"github.com/HabiMatch/profile-service/utils"
)

func (h *ProfileHandler) UpdateProfilePicture(w http.ResponseWriter, r *http.Request) {
	userID := r.FormValue("userid")
	imgURL := os.Getenv("PROFILE_PICTURE_URL") + userID + ".jpeg"
	if userID == "" {
		http.Error(w, "userid is required", http.StatusBadRequest)
		return
	}
	file, _, err := r.FormFile("profile_picture")
	if err != nil {
		http.Error(w, "failed to upload profile picture", http.StatusBadRequest)
		return
	}
	defer file.Close()

	// Verify if the file is an image
	isImage, err := utils.IsImageFile(file)
	if err != nil || !isImage {
		http.Error(w, "file is not an image", http.StatusBadRequest)
		return
	}

	errs := utils.DeleteFromS3(imgURL)
	if errs != nil {
		fmt.Printf("Failed to delete image from S3: %v\n", err)
	} else {
		fmt.Printf("Successfully deleted image: %s\n", imgURL)
	}

	// Reset file pointer to the start for processing
	if _, err := file.Seek(0, io.SeekStart); err != nil {
		http.Error(w, "failed to reset file pointer", http.StatusInternalServerError)
		return
	}

	// Convert to JPEG
	convertedFile, convertedFileName, err := utils.ConvertToJPEG(file, userID)
	if err != nil {
		http.Error(w, "failed to convert file to JPEG", http.StatusInternalServerError)
		return
	}
	defer convertedFile.Close()

	// Upload to S3
	pictureURL, err := utils.UploadToS3(convertedFile, os.Getenv("S3_PROFILE_FOLDER_NAME"), convertedFileName)
	if err != nil {
		http.Error(w, "failed to upload file to S3", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"success": "true", "pictureURL": pictureURL})

}

func (h *ProfileHandler) DeleteProfilePicture(w http.ResponseWriter, r *http.Request) {
	UserID := r.FormValue("userid")
	UserID = strings.ReplaceAll(UserID, " ", "")
	if UserID == "" {
		http.Error(w, "userid is required", http.StatusBadRequest)
		return
	}

	url := os.Getenv("PROFILE_PICTURE_URL") + UserID + ".jpeg"
	var deleteErr, dbErr error

	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		defer wg.Done()
		if err := utils.DeleteFromS3(url); err != nil {
			deleteErr = fmt.Errorf("error deleting profile picture from S3: %w", err)
		}
	}()

	go func() {
		defer wg.Done()
		result := h.DB.Model(&models.Profile{}).Where("user_id = ?", UserID).Update("profile_picture", "")
		if result.Error != nil {
			dbErr = fmt.Errorf("failed to update profile picture in database: %w", result.Error)
		}
	}()
	wg.Wait()

	if deleteErr != nil {
		http.Error(w, "failed to delete profile picture from S3", http.StatusInternalServerError)
		return
	}

	if dbErr != nil {
		http.Error(w, "failed to update profile in database", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"success": "true"})
}
