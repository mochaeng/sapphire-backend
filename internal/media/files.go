package media

import (
	"errors"
	"mime"
	"net/http"
	"os"
	"path/filepath"

	"github.com/google/uuid"
)

var (
	ErrInvalidFile     = errors.New("file is invalid")
	ErrInvalidFileType = errors.New("cannot read file type")
	ErrFileTooBig      = errors.New("file is too big")
	ErrWriteFile       = errors.New("not possible to save file")
)

func SaveFileToServer(fileBytes []byte, folderPath string) (string, error) {
	fileType := http.DetectContentType(fileBytes)
	fileName := uuid.New().String()
	fileEndings, err := mime.ExtensionsByType(fileType)
	if err != nil {
		return "", ErrInvalidFileType
	}
	newFileName := fileName + fileEndings[0]
	workDir, _ := os.Getwd()
	newPath := filepath.Join(workDir, folderPath, newFileName)
	newFile, err := os.Create(newPath)
	if err != nil {
		return "", ErrWriteFile
	}
	defer newFile.Close()
	if _, err := newFile.Write(fileBytes); err != nil || newFile.Close() != nil {
		return "", ErrWriteFile
	}
	return newFileName, nil
}
