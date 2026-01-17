package service

import (
	"context"
	"errors"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/Agmer17/golang_yapping/internal/model"
	"github.com/google/uuid"
	"golang.org/x/sync/errgroup"
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

	fullPath := storage.GetPathPrivateFile(fileName, place...)

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

func (storage *FileStorage) SaveAllPrivateFiles(
	context context.Context,
	files []*multipart.FileHeader,
	filesExt []string,
	place ...string) ([]string, error) {

	if len(files) != len(filesExt) {
		return []string{}, errors.New("The files len and files ext len aren't the same!")
	}

	results := make([]string, len(files))

	tpool, ctx := errgroup.WithContext(context)

	tpool.SetLimit(runtime.NumCPU())

	// for loop save file multithread

	for index, file := range files {

		tpool.Go(func() error {

			select {
			case <-ctx.Done():
				return ctx.Err()
			default:
			}

			filename, err := storage.SavePrivateFile(file, filesExt[index], place...)

			if err != nil {
				return err
			}

			results[index] = filename

			return nil
		})

	}

	if err := tpool.Wait(); err != nil {
		return results, err
	}

	return results, nil
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

func (storage *FileStorage) GetPathPrivateFile(filename string, place ...string) string {

	parts := []string{
		storage.Private,
	}

	parts = append(parts, place...)
	parts = append(parts, filename)

	return path.Join(parts...)

}

func (storage *FileStorage) DeleteAllPrivateFile(fNameList []string, place ...string) {
	for _, v := range fNameList {
		storage.DeletePrivateFile(v, place...)
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
		return model.TypeImage
	case strings.HasPrefix(mime, "video/"):
		return model.TypeVideo
	case strings.HasPrefix(mime, "audio/"):
		return model.TypeAudio
	case strings.HasPrefix(mime, "application/"):
		return model.TypeDocument
	default:
		return ""
	}
}

func (storage *FileStorage) GetVideoDurationPVT(filenames string, place ...string) (time.Duration, error) {

	fullpath := storage.GetPathPrivateFile(filenames, place...)

	cmd := exec.Command(
		"ffprobe",
		"-v", "error",
		"-show_entries", "format=duration",
		"-of", "default=noprint_wrappers=1:nokey=1",
		fullpath,
	)

	out, err := cmd.Output()

	if err != nil {
		return 0, err
	}

	seconds, err := strconv.ParseFloat(strings.TrimSpace(string(out)), 64)
	if err != nil {
		return 0, err
	}

	return time.Duration(seconds * float64(time.Second)), nil

}
