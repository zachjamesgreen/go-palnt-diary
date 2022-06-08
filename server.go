package main

import (
	"encoding/json"
	"fmt"
	"log"
	"mime/multipart"
	"net/http"
	"os"
	"strings"

	"github.com/google/uuid"
)

func main() {
	// create a simple golang server
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		// test path
		fmt.Fprintf(w, "Hello World!")
	})
	http.HandleFunc("/alive", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "alive")
	})
	http.HandleFunc("/upload_image", handler)
	http.HandleFunc("/post", CreatePost)
	fmt.Println("Server started on port 8080")
	http.ListenAndServe(":"+os.Getenv("PORT"), nil)
}

func handler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	var url string
	err := r.ParseMultipartForm(32 << 20)
	if err != nil {
		log.Panicf("Error parsing multipart form: %v", err)
	}
	if files, ok := r.MultipartForm.File["file"]; ok {
		for _, fileHeader := range files {
			fmt.Printf("Uploaded File: %+v\n", fileHeader.Filename)
			fmt.Printf("File Size: %+v\n", fileHeader.Size)
			fmt.Printf("MIME Header: %+v\n", fileHeader.Header["Content-Type"][0])
			url, err = upload(fileHeader)
			if err != nil {
				http.Error(w, "Internal Server Error JSON Marshal", http.StatusInternalServerError)
				return
			}
		}

		res_map := map[string]string{"url": url}
		json.NewEncoder(w).Encode(res_map)
	} else {
		http.Error(w, "Unprocessable Entity", http.StatusUnprocessableEntity)
		return
	}
}

func upload(fileHeader *multipart.FileHeader) (url string, err error) {
	// fileExt := strings.Split(fileHeader.Filename, ".")[1]
	fileExt := fileHeader.Filename[strings.LastIndex(fileHeader.Filename, "."):]
	file, err := fileHeader.Open()
	if err != nil {
		fmt.Println(err.Error())
		return "", err
	}
	fileName := uuid.New().String() + fileExt

	url, err = UploadImage(fileName, file, fileHeader.Header["Content-Type"][0], fileHeader.Filename)
	if err != nil {
		fmt.Println(err.Error())
		return "", err
	}
	return
}
