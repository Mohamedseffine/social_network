package handlers

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"

	"github.com/gofrs/uuid"
)

func (app *App) UploadImageHandler(w http.ResponseWriter, r *http.Request) {
	r.ParseMultipartForm(10 << 20) // 10 MB

	file, handler, err := r.FormFile("image")
	if err != nil {
		http.Error(w, "Error retrieving the file", http.StatusBadRequest)
		return
	}
	defer file.Close()

	// Create a unique filename
	ext := filepath.Ext(handler.Filename)
	newUUID, err := uuid.NewV4()
	if err != nil {
		http.Error(w, "Failed to generate UUID", http.StatusInternalServerError)
		return
	}
	filename := newUUID.String() + ext

	// Create the file
	dst, err := os.Create(filepath.Join("uploads", filename))
	if err != nil {
		http.Error(w, "Failed to create the file", http.StatusInternalServerError)
		return
	}
	defer dst.Close()

	// Copy the uploaded file to the destination file
	if _, err := io.Copy(dst, file); err != nil {
		http.Error(w, "Failed to save the file", http.StatusInternalServerError)
		return
	}

	// Return the file path
	filePath := fmt.Sprintf("/uploads/%s", filename)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"path": filePath})
}

func (app *App) GetImageHandler(w http.ResponseWriter, r *http.Request) {
	var count int
	imagePath := r.URL.Path
	currentUserID := ForContext(r.Context())
	if currentUserID == 0 {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	stm, err := app.DB.Prepare(`SELECT COUNT (id) FROM posts WHERE image = ? `)
	if err != nil {
		http.Error(w, "Image Doesn't Exist", http.StatusNotFound)
		return
	}
	err = stm.QueryRow(imagePath).Scan(&count)
	if err != nil {
		http.Error(w, "Image Doesn't Exist", http.StatusNotFound)
		return
	}
	if count>0 {
		
		return
	}
	stm1, err := app.DB.Prepare(`SELECT COUNT (id) FROM comments WHERE image_url = ? `)
	if err != nil {
		http.Error(w, "Image Doesn't Exist", http.StatusNotFound)
		return
	}
	err = stm1.QueryRow(imagePath).Scan(&count)
	if err != nil {
		http.Error(w, "Image Doesn't Exist", http.StatusNotFound)
		return
	}
	if count>0 {
		
		return
	}
	stm2, err := app.DB.Prepare(`SELECT COUNT (id) FROM group_post_comments WHERE image_url = ? `)
	if err != nil {
		http.Error(w, "Image Doesn't Exist", http.StatusNotFound)
		return
	}
	err = stm2.QueryRow(imagePath).Scan(&count)
	if err != nil {
		http.Error(w, "Image Doesn't Exist", http.StatusNotFound)
		return
	}
	if count>0 {
		
		return
	}
	stm3, err := app.DB.Prepare(`SELECT COUNT (id) FROM group_posts WHERE image = ? `)
	if err != nil {
		http.Error(w, "Image Doesn't Exist", http.StatusNotFound)
		return
	}
	err = stm3.QueryRow(imagePath).Scan(&count)
	if err != nil {
		http.Error(w, "Image Doesn't Exist", http.StatusNotFound)
		return
	}
	if count>0 {
		
		return
	}
}
