package handlers

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"reflect"
	"strconv"
	"strings"
	"sync"

	"github.com/HabiMatch/profile-service/models"
	"github.com/HabiMatch/profile-service/utils"
	"github.com/jackc/pgx/v5/pgconn"
	"gorm.io/gorm"
)

type ProfileType int

func (h *ProfileHandler) CreateProfileHandler(w http.ResponseWriter, r *http.Request, profileType ProfileType) {
	userinfoRaw := r.FormValue("userinfo")
	var input interface{}

	input, _ = GetProfileType(profileType)

	err := json.Unmarshal([]byte(userinfoRaw), &input)
	if err != nil {
		fmt.Printf("Error parsing userinfo JSON: %v\n", err)
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	userIDValue := reflect.ValueOf(input).Elem().FieldByName("UserID").String()
	if userIDValue == "" {
		http.Error(w, "UserId is required", http.StatusBadRequest)
		return
	}

	var result *gorm.DB

	switch profileType {
	case KeeperProfileType:
		if result := h.DB.First(&input, "user_id = ?", userIDValue); result.Error == nil {
			http.Error(w, "Keeper Already Exists", http.StatusNotFound)
			return
		}
		roomImagesCount, err := strconv.Atoi(os.Getenv("ROOM_IMAGES_COUNT"))
		if err != nil {
			print("Error converting ROOM_IMAGES_COUNT to int")
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		// for i := 1; i <= roomImagesCount; i++ {
		// 	file, _, err := r.FormFile("room_image" + strconv.Itoa(i))
		// 	isImage, err := utils.IsImageFile(file)
		// 	print("isImage: ", isImage)
		// 	if err != nil {
		// 		http.Error(w, "Provide Profile Picture", http.StatusBadRequest)
		// 		return
		// 	}
		// }

		imageChannel := make(chan string, roomImagesCount)
		errorChannel := make(chan error, 1)
		var wg sync.WaitGroup
		err = r.ParseMultipartForm(10 << 20) // Parse multipart form for file uploads
		if err != nil {
			http.Error(w, "Failed to parse multipart form", http.StatusInternalServerError)
			return
		}

		for i := 0; i < roomImagesCount; i++ {
			wg.Add(1)
			go func(i int) {
				defer wg.Done()
				f, fh, errs := r.FormFile(fmt.Sprintf("room_image%d", i+1))
				if errs != nil || f == nil || fh == nil {
					errorChannel <- fmt.Errorf("error reading file for room_image%d: %v", i+1, errs)
					return
				}
				defer f.Close()

				convertedFile, convertedFileName, err := utils.ConvertToJPEG(f, fmt.Sprintf("%s_%d", userIDValue, i+1))
				if err != nil {
					errorChannel <- fmt.Errorf("failed to convert file to JPEG: %v", err)
					return
				}
				defer convertedFile.Close()

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

		for imgURL := range imageChannel {
			reflect.ValueOf(input).Elem().FieldByName("FlatImages").Set(reflect.Append(reflect.ValueOf(input).Elem().FieldByName("FlatImages"), reflect.ValueOf(imgURL)))
		}

		if len(errorChannel) > 0 {
			err := <-errorChannel
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		result = h.DB.Create(input)
		if result.Error != nil {
			http.Error(w, "Failed to create profile", http.StatusInternalServerError)
			// if pgErr, ok := result.Error.(*pgconn.PgError); ok {
			// 	if pgErr.Code == "23505" {
			// 		http.Error(w, "User Already Exists", http.StatusBadRequest)
			// 		return
			// 	}
			// }
			return
		}

		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, "Profile created successfully")
		return

	case SeekerProfileType:
		result = h.DB.Create(input)
		if result.Error != nil {
			http.Error(w, "Failed to create profile", http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, "Profile created successfully")
		return

	case GeneralProfileType:
		var usercredentials models.UserCredentials
		usercredentials.UserID = userIDValue
		if result := h.DB.Create(&usercredentials); result.Error != nil {
			if pgErr, ok := result.Error.(*pgconn.PgError); ok {
				if pgErr.Code == "23505" {
					http.Error(w, "User Already Exists", http.StatusBadRequest)
					return
				}
			}
		}
		f, fh, errs := r.FormFile("profile_picture")
		if errs != nil || f == nil || fh == nil {
			http.Error(w, "Provide Profile Picture", http.StatusBadRequest)
			return
		}
		defer f.Close()

		err := r.ParseMultipartForm(10 << 20)
		if err != nil {
			http.Error(w, "Unable to parse form data", http.StatusBadRequest)
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

			userID := userIDValue // Use userIDValue from above
			userID = strings.ReplaceAll(userID, " ", "")
			if userID == "" {
				errChan <- fmt.Errorf("userid is required")
				close(pictureURLChan)
				return
			}

			isImage, err := utils.IsImageFile(file)
			if err != nil || !isImage {
				errChan <- fmt.Errorf("uploaded file is not a valid image")
				close(pictureURLChan)
				return
			}

			if _, err := file.Seek(0, io.SeekStart); err != nil {
				errChan <- fmt.Errorf("failed to reset file pointer: %v", err)
				close(pictureURLChan)
				return
			}

			convertedFile, convertedFileName, err := utils.ConvertToJPEG(file, userID)
			if err != nil {
				errChan <- fmt.Errorf("failed to convert file to JPEG: %v", err)
				close(pictureURLChan)
				return
			}
			defer convertedFile.Close()

			pictureURL, err := utils.UploadToS3(convertedFile, os.Getenv("S3_PROFILE_FOLDER_NAME"), convertedFileName)
			if err != nil {
				errChan <- fmt.Errorf("failed to upload to S3: %v", err)
				close(pictureURLChan)
				return
			}

			pictureURLChan <- pictureURL
			close(pictureURLChan)
		}()

		wg.Add(1)
		go func() {
			defer wg.Done()
			pictureURL := <-pictureURLChan
			if pictureURL == "" {
				errChan <- fmt.Errorf("failed to receive pictureURL")
				return
			}

			profile, lat, lon, err = serializeProfileDetails(*input.(*models.Profile), pictureURL)
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

			close(geoReady)
			errChan <- nil
		}()

		wg.Add(1)
		go func() {
			defer wg.Done()
			<-geoReady
			err := utils.StoreGeolocation(h.DB, profile.UserID, lat, lon)
			if err != nil {
				errChan <- fmt.Errorf("failed to store geolocation: %v", err)
				return
			}
			errChan <- nil
		}()

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

}

func (h *ProfileHandler) UpdateProfileHandler(w http.ResponseWriter, r *http.Request, profileType ProfileType) {
	userinfoRaw := r.FormValue("userinfo")
	var input interface{}
	var profile interface{}
	switch profileType {
	case KeeperProfileType:
		input = &models.Keeper{}
		profile = &models.Keeper{}
	case SeekerProfileType:
		input = &models.Seeker{}
		profile = &models.Seeker{}
	case GeneralProfileType:
		input = &models.Profile{}
		profile = &models.Profile{}
	default:
		http.Error(w, "Invalid profile type", http.StatusBadRequest)
		return
	}

	// Parse the JSON input
	if err := json.Unmarshal([]byte(userinfoRaw), &input); err != nil {
		http.Error(w, "Invalid JSON input", http.StatusBadRequest)
		return
	}
	userIDValue := reflect.ValueOf(input).Elem().FieldByName("UserID").String()
	if userIDValue == "" {
		http.Error(w, "UserId is required", http.StatusBadRequest)
		return
	}
	// Fetch existing profile

	if result := h.DB.First(&profile, "user_id = ?", userIDValue); result.Error != nil {
		http.Error(w, fmt.Sprintf("Profile Not found %s profile: %v", input, result.Error), http.StatusNotFound)
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
		http.Error(w, fmt.Sprintf("Failed to update %s profile: %v", input, err), http.StatusInternalServerError)
		return
	}

	// Respond with the updated profile
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(profile)
}

func (h *ProfileHandler) DeleteProfileHandler(w http.ResponseWriter, r *http.Request, profileType ProfileType) {
	var input interface{}

	switch profileType {
	case KeeperProfileType:
		input = &models.Keeper{}
	case SeekerProfileType:
		input = &models.Seeker{}
	case GeneralProfileType:
		input = &models.Profile{}
	default:
		http.Error(w, "Invalid profile type", http.StatusBadRequest)
		return
	}
	userID := r.FormValue("userid")
	if userID == "" {
		http.Error(w, "UserId is required", http.StatusBadRequest)
		return
	}
	if result := h.DB.First(&input, "user_id = ?", userID); result.Error != nil {
		http.Error(w, fmt.Sprintf("Failed to find %s profile: %v", input, result.Error), http.StatusNotFound)
		return
	}
	if err := h.DB.Delete(&input).Error; err != nil {
		http.Error(w, fmt.Sprintf("Failed to delete %s profile: %v", input, err), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "Profile deleted successfully")
}
