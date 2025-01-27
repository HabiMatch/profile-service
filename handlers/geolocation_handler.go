package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/HabiMatch/profile-service/models"
	"github.com/HabiMatch/profile-service/utils"
)

func (h *ProfileHandler) UpdateGeolocation(w http.ResponseWriter, r *http.Request) {
	userinfoRaw := r.FormValue("userinfo")
	var input models.Geolocation

	if err := json.Unmarshal([]byte(userinfoRaw), &input); err != nil {
		http.Error(w, "Invalid JSON input", http.StatusBadRequest)
		return
	}
	if input.UserID == "" {
		http.Error(w, "Provide UserID", http.StatusBadRequest)
		return
	}
	err := utils.UpdateGeolocation(h.DB, input.UserID, input.Latitude, input.Longitude)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	h.DB.Exec("UPDATE profiles SET latitude = ?, longitude = ?, updated_at = NOW() WHERE user_id = ?", input.Latitude, input.Longitude, input.UserID)
	w.WriteHeader(http.StatusOK)
}
