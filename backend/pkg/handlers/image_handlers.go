package handlers

import (
	"encoding/json"
	"fmt"
	"io"

	"net/http"
	"os"
	"path/filepath"
	"social-network/backend/pkg/auth"
	"strings"

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
	buffer := make([]byte, 512)
	n, err := file.Read(buffer)
	if err != nil && err != io.EOF {
		http.Error(w, "Error reading file", http.StatusInternalServerError)
		return
	}
	contentType := http.DetectContentType(buffer[:n])
	if !strings.HasPrefix(contentType, "image") {
		http.Error(w, "the file inserted is not an image", http.StatusBadRequest)
		return
	}
	file.Seek(0, 0)
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

	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var count int
	imagePath := strings.TrimSpace(r.URL.Path)
	token, err := r.Cookie("session_token")

	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
	}
	currentUserID, err := auth.GetUserFromSession(app.DB, token.Value)
	if currentUserID == 0 || err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	stm4, err := app.DB.Prepare(`SELECT COUNT (id) FROM users WHERE avatar = ? `)
	if err != nil {
		http.Error(w, "Image Doesn't Exist", http.StatusNotFound)
		return
	}
	err = stm4.QueryRow(imagePath).Scan(&count)
	if err != nil {
		http.Error(w, "Image Doesn't Exist", http.StatusNotFound)
		return
	}
	if count > 0 {
		http.ServeFile(w, r, "."+r.URL.Path)
		return
	}

	stm, err := app.DB.Prepare(`SELECT COUNT (id) FROM posts WHERE image = ? `)
	if err != nil {
		http.Error(w, "error:"+err.Error(), http.StatusInternalServerError)
		return
	}
	err = stm.QueryRow(imagePath).Scan(&count)
	if err != nil {
		http.Error(w, "error:"+err.Error(), http.StatusInternalServerError)
		return
	}
	if count > 0 {
		app.CanSeePostImage(w, r, imagePath, currentUserID)
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
	if count > 0 {
		app.CanSeeCommentImage(w, r, imagePath, currentUserID)
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
	if count > 0 {
		app.CanSeeGroupPostImage(w, r, imagePath, currentUserID)
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
	if count > 0 {
		app.CanSeeGroupPostCommentImage(w, r, imagePath, currentUserID)
		return
	}
	http.Error(w, "image not found", http.StatusNotFound)
}

func (app *App) CanSeePostImage(w http.ResponseWriter, r *http.Request, imagePath string, currentUserID int64) {

	var status bool
	var pPrivacy string
	var postId int
	var creatorID int64
	stM, err := app.DB.Prepare(`
		SELECT u.profile_is_public , p.privacy , p.id , p.user_id
		FROM posts p 
		INNER JOIN users u ON p.user_id = u.id 
		WHERE p.image = ?
		`)
	if err != nil {

		http.Error(w, "error:"+err.Error(), http.StatusInternalServerError)
		return
	}
	err = stM.QueryRow(imagePath).Scan(&status, &pPrivacy, &postId, &creatorID)
	if err != nil {

		http.Error(w, "error:"+err.Error(), http.StatusInternalServerError)
		return
	}
	is_following, err := app.isFollowing(currentUserID, creatorID)
	if err != nil {

		http.Error(w, "error:"+err.Error(), http.StatusInternalServerError)
		return
	}
	if creatorID == currentUserID || (pPrivacy == "public" && (status || is_following)) || (app.CanSeePost(int(currentUserID), postId)) || (is_following && pPrivacy == "almost private") {
		http.ServeFile(w, r, "."+imagePath)
		return
	} else {
		http.Error(w, "Forbidden Assets", http.StatusForbidden)
	}
}

func (app *App) CanSeeCommentImage(w http.ResponseWriter, r *http.Request, imagePath string, currentUserID int64) {
	var postId int
	var status bool
	var pPrivacy string
	var creatorID int64
	stm, err := app.DB.Prepare(`SELECT post_id FROM comments WHERE image_url = ?`)
	if err != nil {
		http.Error(w, "error:"+err.Error(), http.StatusInternalServerError)
		return
	}
	err = stm.QueryRow(imagePath).Scan(&postId)
	if err != nil {
		http.Error(w, "error:"+err.Error(), http.StatusInternalServerError)
		return
	}

	stM, err := app.DB.Prepare(`
		SELECT u.profile_is_public , p.privacy ,  p.user_id
		FROM posts p 
		INNER JOIN users u ON p.user_id = u.id 
		WHERE p.id  = ?
		`)
	if err != nil {
		http.Error(w, "error:"+err.Error(), http.StatusInternalServerError)
		return
	}
	err = stM.QueryRow(postId).Scan(&status, &pPrivacy, &creatorID)
	if err != nil {
		http.Error(w, "error:"+err.Error(), http.StatusInternalServerError)
		return
	}
	is_following, err := app.isFollowing(currentUserID, creatorID)
	if err != nil {
		http.Error(w, "error:"+err.Error(), http.StatusInternalServerError)
		return
	}
	if creatorID == currentUserID || (pPrivacy == "public" && (status || is_following)) || (app.CanSeePost(int(currentUserID), postId)) || (is_following && pPrivacy == "almost private") {
		http.ServeFile(w, r, "."+imagePath)
		return
	} else {
		http.Error(w, "Forbidden Assets", http.StatusForbidden)
		return
	}
}

func (app *App) CanSeeGroupPostImage(w http.ResponseWriter, r *http.Request, imagePath string, currentUserID int64) {
	var Gid int
	var count int
	stm, err := app.DB.Prepare(`SELECT group_id FROM group_posts WHERE image = ? `)
	if err != nil {

		http.Error(w, "error"+err.Error(), http.StatusInternalServerError)
		return
	}
	err = stm.QueryRow(imagePath).Scan(&Gid)
	if err != nil {

		http.Error(w, "error"+err.Error(), http.StatusInternalServerError)
		return
	}
	stm1, err := app.DB.Prepare(`SELECT count(id) FROM group_members WHERE group_id = ? AND user_id = ? AND status = 'accepted' `)
	if err != nil {

		http.Error(w, "error"+err.Error(), http.StatusInternalServerError)
		return
	}
	err = stm1.QueryRow(Gid, currentUserID).Scan(&count)
	if err != nil {

		http.Error(w, "error"+err.Error(), http.StatusInternalServerError)
		return
	}
	if count > 0 {

		http.ServeFile(w, r, "."+imagePath)
		return
	} else {

		http.Error(w, "Forbidden Assets", http.StatusForbidden)
	}
}

func (app *App) CanSeeGroupPostCommentImage(w http.ResponseWriter, r *http.Request, imagePath string, currentUserID int64) {
	var Gid int
	var count int
	stm, err := app.DB.Prepare(`
	SELECT gp.group_id 
	FROM group_posts gp
	INNER JOIN group_post_comments gpc ON gpc.post_id = gp.id 
	WHERE gpc.image_url = ?
	`)
	if err != nil {
		http.Error(w, "error"+err.Error(), http.StatusInternalServerError)
		return
	}
	err = stm.QueryRow(imagePath).Scan(&Gid)
	if err != nil {
		http.Error(w, "error"+err.Error(), http.StatusInternalServerError)
		return
	}
	stm1, err := app.DB.Prepare(`SELECT count(id) FROM group_members WHERE group_id = ? AND user_id = ? AND status = 'accepted' `)
	if err != nil {
		http.Error(w, "error"+err.Error(), http.StatusInternalServerError)
		return
	}
	err = stm1.QueryRow(Gid, currentUserID).Scan(&count)
	if err != nil {
		http.Error(w, "error"+err.Error(), http.StatusInternalServerError)
		return
	}
	if count > 0 {
		http.ServeFile(w, r, "."+imagePath)
		return
	} else {
		http.Error(w, "Forbidden Assets", http.StatusForbidden)
	}
}

func (app *App) CanSeePost(Userid, postId int) bool {
	var count int
	stm, err := app.DB.Prepare(`SELECT COUNT (*) FROM post_viewers WHERE post_id = ? AND viewer_id = ?`)
	if err != nil {

		return false
	}
	err = stm.QueryRow(postId, Userid).Scan(&count)
	if err != nil {

		return false
	}

	return count > 0

}
