package handlers

import (
	"encoding/json"
	"net/http"

	"gorm.io/gorm"
)

type ProfileHandler struct {
	DB *gorm.DB
}

const (
	KeeperProfileType ProfileType = iota
	SeekerProfileType
	GeneralProfileType
)

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
		h.CreateProfileHandler(w, r, GeneralProfileType)
	case "update_profile":
		h.UpdateProfileHandler(w, r, GeneralProfileType)
	case "update_profile_picture":
		h.UpdateProfilePicture(w, r)
	case "update_geolocation":
		h.UpdateGeolocation(w, r)
	case "keeper_profile":
		h.CreateProfileHandler(w, r, KeeperProfileType)
	case "update_keeper_profile":
		h.UpdateProfileHandler(w, r, KeeperProfileType)
	case "delete_keeper_profile":
		h.DeleteProfileHandler(w, r, KeeperProfileType)
	case "seeker_profile":
		h.CreateProfileHandler(w, r, SeekerProfileType)
	case "update_seeker_profile":
		h.UpdateProfileHandler(w, r, SeekerProfileType)
	case "delete_seeker_profile":
		h.DeleteProfileHandler(w, r, SeekerProfileType)

	default:
		http.Error(w, "Invalid operation", http.StatusBadRequest)
		return
	}

}
