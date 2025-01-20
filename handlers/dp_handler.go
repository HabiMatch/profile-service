package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"sync"

	"github.com/HabiMatch/profile-service/models"
	"github.com/HabiMatch/profile-service/utils"
)

func (h *ProfileHandler) UpdateProfilePicture(w http.ResponseWriter, r *http.Request) {
	// profile, _, _ := serializeProfileDetails(r, "pictureURL")

	// w.Header().Set("Content-Type", "application/json")
	// json.NewEncoder(w).Encode(profile)
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
