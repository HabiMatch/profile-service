package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/HabiMatch/profile-service/models"
)

func (h *ProfileHandler) UpdateProfilePicture(w http.ResponseWriter, r *http.Request) {
	// profile, _, _ := serializeProfileDetails(r, "pictureURL")

	// w.Header().Set("Content-Type", "application/json")
	// json.NewEncoder(w).Encode(profile)
}

func (h *ProfileHandler) DeleteProfilePicture(w http.ResponseWriter, r *http.Request) {
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
