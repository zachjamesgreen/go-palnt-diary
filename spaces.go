package main

import (
	"fmt"
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
	key := os.Getenv("DO_SPACES_KEY")       // Access key pair. You can create access key pairs using the control panel or API.
	secret := os.Getenv("DO_SPACES_SECRET") // Secret access key defined through an environment variable.

	s3Config = &aws.Config{
		Credentials: credentials.NewStaticCredentials(key, secret, ""), // Specifies your credentials.
		Endpoint:    aws.String("https://sfo3.digitaloceanspaces.com"), // Find your endpoint in the control panel, under Settings. Prepend "https://".
		Region:      aws.String("sfo3"),                                // Must be "us-east-1" when creating new Spaces. Otherwise, use the region in your endpoint, such as "nyc3".
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

	// Step 4: Define the parameters of the object you want to upload.
	object := s3.PutObjectInput{
		ContentType: aws.String(contentType),
		Bucket:      aws.String("persephone/images/"), // The path to the directory you want to upload the object to, starting with your Space name.
		Key:         aws.String(fileName),             // Object key, referenced whenever you want to access this file later.
		Body:        file,                             // The object's contents.
		ACL:         aws.String("public-read"),        // Defines Access-control List (ACL) permissions, such as private or public.
		Metadata: map[string]*string{ // Required. Defines metadata tags.
			"original-filename": aws.String(originalName),
		},
	}

	_, err = s3Client.PutObject(&object)
	if err != nil {
		fmt.Println(err.Error())
		return "", err
	}
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
	// ext := strings.Split(post.CoverImageFile.Filename, ".")[1]
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
		Metadata: map[string]*string{ // Required. Defines metadata tags.
			"id":    aws.String(post.ID),
			"image": aws.String(post.CoverImage),
		},
	}

	_, err = s3Client.PutObject(&postObject)
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	post.Url = "https://persephone.sfo3.digitaloceanspaces.com/posts/" + fmt.Sprintf("%d/%d/%d/", year, month, day) + post.Title + ".md"
	return
}
