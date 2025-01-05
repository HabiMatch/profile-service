package router

import (
	"net/http"

	"github.com/HabiMatch/profile-service/handlers"
	"github.com/gorilla/mux"
	"gorm.io/gorm"
)

func InitRouter(db *gorm.DB) *mux.Router {
	r := mux.NewRouter()

	profileHandler := handlers.ProfileHandler{DB: db}
	r.HandleFunc("/api/hello", profileHandler.HelloWorld).Methods(http.MethodGet)
	r.HandleFunc("/api/profiles", profileHandler.ManageProfile).Methods(http.MethodPost)
	return r
}
