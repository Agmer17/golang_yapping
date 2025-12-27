package pkg

import (
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/google/uuid"
)

var allowedMimes = map[string]bool{
	// image
	"image/jpeg": true,
	"image/png":  true,
	"image/webp": true,

	// video
	"video/mp4":       true,
	"video/webm":      true,
	"video/quicktime": true,

	// audio
	"audio/mpeg": true,
	"audio/wav":  true,
	"audio/ogg":  true,

	// document
	"application/pdf":    true,
	"application/msword": true,
	"application/vnd.openxmlformats-officedocument.wordprocessingml.document": true,
	"text/plain": true,
}

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

func IsTypeSupportted(mimeType string) (string, bool) {

	if !allowedMimes[mimeType] {
		return "", false
	}

	_, after, _ := strings.Cut(mimeType, "/")

	fileExt := "." + after

	return fileExt, true
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

func DeletePrivateFile(fname string) {

	deletePath := filepath.Join(mustGetProjectRoot(), "uploads", "private", fname)

	if err := os.Remove(deletePath); err != nil {
		if !os.IsNotExist(err) {
			log.Printf("failed to remove file %s: %v", deletePath, err)
		}
	}

}

func DeleteAllPrivateFile(fNameList []string) {
	for _, v := range fNameList {
		DeletePrivateFile(v)
	}

}

func GetMediaType(mime string) string {
	switch {
	case strings.HasPrefix(mime, "image/"):
		return "IMAGE"
	case strings.HasPrefix(mime, "video/"):
		return "VIDEO"
	case strings.HasPrefix(mime, "audio/"):
		return "AUDIO"
	case strings.HasPrefix(mime, "application/"):
		return "DOCUMENT"
	default:
		return ""
	}
}
