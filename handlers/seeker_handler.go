package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"reflect"

	"github.com/HabiMatch/profile-service/models"
)

func (h *ProfileHandler) SeekerProfile(w http.ResponseWriter, r *http.Request) {
	userinfoRaw := r.FormValue("userinfo")
	var input models.Seeker
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
	if result := h.DB.Create(&input); result.Error != nil {
		fmt.Printf("Failed to create seeker: %v\n", result.Error)
		http.Error(w, "Failed to create seeker", http.StatusInternalServerError)
		return
	}

	// Success response
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "Profile created successfully with %d images uploaded")
}

func (h *ProfileHandler) UpdateSeekerProfile(w http.ResponseWriter, r *http.Request) {
	userinfoRaw := r.FormValue("userinfo")
	var input models.Seeker

	// Parse the JSON input
	if err := json.Unmarshal([]byte(userinfoRaw), &input); err != nil {
		http.Error(w, "Invalid JSON input", http.StatusBadRequest)
		return
	}
	if input.UserID == "" {
		http.Error(w, "UserId is required", http.StatusBadRequest)
		return
	}
	// Fetch existing seeker
	var seeker models.Seeker
	if result := h.DB.First(&seeker, "user_id = ?", input.UserID); result.Error != nil {
		http.Error(w, "Profile not found", http.StatusNotFound)
		return
	}

	inputValue := reflect.ValueOf(input)
	profileValue := reflect.ValueOf(&seeker).Elem()
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

	// Save the updated seeker
	if err := h.DB.Save(&seeker).Error; err != nil {
		http.Error(w, "Failed to update seeker", http.StatusInternalServerError)
		return
	}

	// Respond with the updated seeker
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(seeker)

}

func (h *ProfileHandler) DeleteSeekerProfile(w http.ResponseWriter, r *http.Request) {
	userID := r.FormValue("userid")
	if userID == "" {
		http.Error(w, "UserID is required", http.StatusBadRequest)
		return
	}
	if result := h.DB.Delete(&models.Seeker{}, "user_id = ?", userID); result.Error != nil {
		http.Error(w, "Failed to delete seeker", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "Seeker profile deleted successfully")
}
