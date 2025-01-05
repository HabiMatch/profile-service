package handlers

import (
	"net/http"
)

func (h *ProfileHandler) UpdateGeolocation(w http.ResponseWriter, r *http.Request) {
	// profileID := r.FormValue("profile_id")
	// if profileID == "" {
	// 	http.Error(w, "Provide profile_id", http.StatusBadRequest)
	// 	return
	// }

	// latitude := r.FormValue("latitude")
	// if latitude == "" {
	// 	http.Error(w, "Provide latitude", http.StatusBadRequest)
	// 	return
	// }

	// longitude := r.FormValue("longitude")
	// if longitude == "" {
	// 	http.Error(w, "Provide longitude", http.StatusBadRequest)
	// 	return
	// }

	// profile := models.Profile{}
	// if err := h.DB.Where("id = ?", profileID).First(&profile).Error; err != nil {
	// 	http.Error(w, "Profile not found", http.StatusNotFound)
	// 	return
	// }

	// profile.Latitude = latitude
	// profile.Longitude = longitude

	// if err := h.DB.Save(&profile).Error; err != nil {
	// 	http.Error(w, "Error updating profile", http.StatusInternalServerError)
	// 	return
	// }
}
