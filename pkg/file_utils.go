package pkg

import (
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/google/uuid"
)

func getProjectRoot() (string, error) {
	projectPath, err := os.Getwd()
	if err != nil {
		return "", err
	}
	return projectPath, nil
}

func MustUploadsInit() {
	root := mustGetProjectRoot()

	mustCreateDir(filepath.Join(root, "uploads"))
	mustCreateDir(filepath.Join(root, "uploads", "private"))
	mustCreateDir(filepath.Join(root, "uploads", "public"))
}

func mustGetProjectRoot() string {
	root, err := getProjectRoot()
	if err != nil {
		panic(err)
	}
	return root
}

func mustCreateDir(path string) {
	if err := os.MkdirAll(path, 0755); err != nil {
		panic(err)
	}
}

func DetectFileType(fileHeader *multipart.FileHeader) (string, error) {

	f, err := fileHeader.Open()

	if err != nil {
		return "", err
	}

	defer f.Close()

	buf := make([]byte, 512)
	_, err = f.Read(buf)

	if err != nil {
		return "", err
	}

	mimeType := http.DetectContentType(buf)

	return mimeType, nil
}

func IsValidImage(mimeType string) (string, bool) {

	isValid := strings.Contains(mimeType, "image/")

	if isValid {
		return "." + strings.TrimPrefix(mimeType, "image/"), true
	}

	return "", false
}

func SavePrivateFile(fileHeader *multipart.FileHeader, ext string) (string, error) {
	file, err := fileHeader.Open()
	if err != nil {
		return "", err
	}
	defer file.Close()

	fileName := uuid.New().String() + ext
	fullPath := filepath.Join(mustGetProjectRoot(), "uploads", "private", fileName)

	dst, err := os.Create(fullPath)
	if err != nil {
		return "", err
	}
	defer dst.Close()

	if _, err := io.Copy(dst, file); err != nil {
		return "", err
	}

	return fileName, nil
}
