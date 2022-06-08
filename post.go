package main

import (
	"encoding/json"
	"log"
	"mime/multipart"
	"net/http"
)

type Post struct {
	ID             string                `json:"id"`
	Title          string                `json:"title"`
	Body           string                `json:"body"`
	CoverImage     string                `json:"coverImage"`
	Published      bool                  `json:"published"`
	Url            string                `json:"url"`
	Slug           string                `json:"slug"`
	CoverImageFile *multipart.FileHeader `json:"-"`
}

func CreatePost(w http.ResponseWriter, r *http.Request) {
	var post Post
	var err error
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	err = r.ParseMultipartForm(32 << 20)
	if err != nil {
		log.Panicf("Error parsing multipart form: %v", err)
	}
	files, ok := r.MultipartForm.File["coverImageFile"]
	if !ok {
		http.Error(w, "Unprocessable Entity", http.StatusUnprocessableEntity)
		return
	}
	file := files[0]
	title := r.FormValue("title")
	slug := r.FormValue("slug")
	body := r.FormValue("body")
	published := r.FormValue("published")
	if title == "" || body == "" || published == "" {
		http.Error(w, "Unprocessable Entity", http.StatusUnprocessableEntity)
		return
	}
	post = Post{
		Title:          title,
		Slug:           slug,
		Body:           body,
		CoverImageFile: file,
		Published:      published == "true",
	}
	err = post.Save()
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(post)
}

func (post *Post) Save() (err error) {
	err = UploadPost(post)
	if err != nil {
		return
	}

	return nil
}
