package main

import (
	"fmt"
	"log"
	"mime/multipart"
	"os"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/google/uuid"
)

func createConfig() (s3Config *aws.Config) {
	key := os.Getenv("DO_SPACES_KEY")
	secret := os.Getenv("DO_SPACES_SECRET")

	s3Config = &aws.Config{
		Credentials: credentials.NewStaticCredentials(key, secret, ""),
		Endpoint:    aws.String(os.Getenv("SPACES_URL")),
		Region:      aws.String("sfo3"),
	}
	return
}

func UploadImage(fileName string, file multipart.File, contentType string, originalName string) (url string, err error) {
	err = nil
	s3Config := createConfig()
	newSession, err := session.NewSession(s3Config)
	if err != nil {
		fmt.Println(err.Error())
		return "", err
	}
	s3Client := s3.New(newSession)

	object := s3.PutObjectInput{
		ContentType: aws.String(contentType),
		Bucket:      aws.String("images/"),
		Key:         aws.String(fileName),
		Body:        file,
		ACL:         aws.String("public-read"),
		Metadata: map[string]*string{
			"original-filename": aws.String(originalName),
		},
	}

	_, err = s3Client.PutObject(&object)
	if err != nil {
		fmt.Println(err.Error())
		return "", err
	}
	log.Println("Uploaded image to S3 <spaces name> <original name> " + fileName + " " + originalName)
	url = os.Getenv("SPACES_URL") + "/images/" + fileName
	return
}

func UploadPost(post *Post) (err error) {
	t := time.Now()
	year := t.Year()
	month := t.Month()
	day := t.Day()
	s3Config := createConfig()
	newSession, err := session.NewSession(s3Config)
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	s3Client := s3.New(newSession)
	image, err := post.CoverImageFile.Open()
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	ext := post.CoverImageFile.Filename[strings.LastIndex(post.CoverImageFile.Filename, "."):]
	imageFileName := uuid.New().String() + "." + ext
	imageUrl, err := UploadImage(imageFileName, image, post.CoverImageFile.Header["Content-Type"][0], post.CoverImageFile.Filename)
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	post.CoverImage = imageUrl

	postObject := s3.PutObjectInput{
		ContentType: aws.String("text/markdown"),
		Bucket:      aws.String("persephone/posts/" + fmt.Sprintf("%d/%d/%d/", year, month, day)),
		Key:         aws.String(post.Title + ".md"),
		Body:        strings.NewReader(post.Body),
		ACL:         aws.String("public-read"),
		Metadata: map[string]*string{
			"id":    aws.String(post.ID),
			"image": aws.String(post.CoverImage),
		},
	}

	_, err = s3Client.PutObject(&postObject)
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	log.Println("Uploaded post to S3 <spaces name> " + post.Title)
	post.Url = os.Getenv("SPACES_URL") + "/posts/" + fmt.Sprintf("%d/%d/%d/", year, month, day) + post.Title + ".md"
	return
}
