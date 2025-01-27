package handlers

import (
	"encoding/json"
	"net/http"

	"gorm.io/gorm"
)

type ProfileHandler struct {
	DB *gorm.DB
}

func (h *ProfileHandler) HelloWorld(w http.ResponseWriter, r *http.Request) {
	hello := map[string]string{"hello": "world"}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(hello)
}

func (h *ProfileHandler) ManageProfile(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "multipart/form-data")

	operation := r.FormValue("operation")
	if operation == "" {
		http.Error(w, "Provide operation", http.StatusBadRequest)
		return
	}
	println("Operation: ", operation)
	switch operation {
	case "create_profile":
		h.CreateProfile(w, r)
	case "update_profile":
		h.UpdateProfile(w, r)
	case "update_profile_picture":
		h.UpdateProfilePicture(w, r)
	case "update_geolocation":
		h.UpdateGeolocation(w, r)
	case "keeper_profile":
		h.KeeperProfile(w, r)
	case "update_keeper_profile":
		h.UpdateKeeperProfile(w, r)
	case "delete_keeper_profile":
		h.DeleteKeeperProfile(w, r)
	case "seeker_profile":
		h.SeekerProfile(w, r)
	case "update_seeker_profile":
		h.UpdateSeekerProfile(w, r)
	case "delete_seeker_profile":
		h.DeleteSeekerProfile(w, r)

	default:
		http.Error(w, "Invalid operation", http.StatusBadRequest)
		return
	}

}
