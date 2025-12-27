package service

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

type FileStorage struct {
	Root    string
	Public  string
	Private string
}

func mustGetProjectRoot() string {
	exe, err := os.Executable()
	if err != nil {
		panic(err)
	}
	return filepath.Dir(exe)
}

func mustCreateDir(path string) {
	if err := os.MkdirAll(path, 0755); err != nil {
		panic(err)
	}
}

func NewFileService() *FileStorage {

	root := mustGetProjectRoot()
	uploadsDir := filepath.Join(root, "uploads")

	privateDir := filepath.Join(uploadsDir, "private")
	publicDir := filepath.Join(uploadsDir, "public")

	mustCreateDir(uploadsDir) // uploads
	mustCreateDir(privateDir) // private
	mustCreateDir(publicDir)  // public

	fileStoreage := FileStorage{
		Root:    root,
		Public:  publicDir,
		Private: privateDir,
	}

	return &fileStoreage

}

func (storage *FileStorage) SavePublicFile(
	fileHeader *multipart.FileHeader,
	ext string,
	place ...string) (string, error) {

	file, err := fileHeader.Open()
	if err != nil {
		return "", err
	}
	defer file.Close()

	fileName := uuid.New().String() + ext

	parts := []string{
		storage.Public,
	}

	parts = append(parts, place...)

	parts = append(parts, fileName)

	fullPath := filepath.Join(parts...)

	if err := os.MkdirAll(filepath.Dir(fullPath), 0755); err != nil {
		return "", err
	}

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

func (storage *FileStorage) SavePrivateFile(
	fileHeader *multipart.FileHeader,
	ext string,
	place ...string,
) (string, error) {

	file, err := fileHeader.Open()
	if err != nil {
		return "", err
	}
	defer file.Close()

	fileName := uuid.New().String() + ext

	parts := []string{
		storage.Private,
	}

	parts = append(parts, place...)

	parts = append(parts, fileName)

	fullPath := filepath.Join(parts...)

	if err := os.MkdirAll(filepath.Dir(fullPath), 0755); err != nil {
		return "", err
	}

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

func (storage *FileStorage) DeletePrivateFile(fname string, place ...string) {

	parts := []string{storage.Private}
	parts = append(parts, place...)
	parts = append(parts, fname)

	deletePath := filepath.Join(parts...)

	if err := os.Remove(deletePath); err != nil {
		if !os.IsNotExist(err) {
			log.Printf("failed to remove file %s: %v", deletePath, err)
		}
	}

}

func (storage *FileStorage) DeleteAllPrivateFile(fNameList []string) {
	for _, v := range fNameList {
		storage.DeletePrivateFile(v)
	}

}

func (storage *FileStorage) DetectFileType(fileHeader *multipart.FileHeader) (string, error) {

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

func (storage *FileStorage) IsTypeSupportted(mimeType string) (string, bool) {

	if !allowedMimes[mimeType] {
		return "", false
	}

	_, after, _ := strings.Cut(mimeType, "/")

	fileExt := "." + after

	return fileExt, true
}

func (storage *FileStorage) GetMediaType(mime string) string {
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
