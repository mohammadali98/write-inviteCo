package webui

import (
	"errors"
	"io"
	"mime/multipart"
	"net/http"
	"path/filepath"
	"strings"
)

// ErrInvalidUpload is returned by ValidateUploadedFile when a file fails the size cap, extension
// whitelist, or content-sniffing check.
var ErrInvalidUpload = errors.New("invalid upload")

// ValidateUploadedFile enforces a size cap and an extension whitelist, then sniffs the first 512
// bytes of the file to confirm its real content type matches what the extension claims. It returns
// the canonical (sniffed) extension to use when saving the file.
//
// allowedExt maps a lowercase file extension (e.g. ".jpg") to the MIME type it is expected to be.
// detectedExt maps a sniffed MIME type back to its canonical extension.
func ValidateUploadedFile(fileHeader *multipart.FileHeader, maxSize int64, allowedExt map[string]string, detectedExt map[string]string) (string, error) {
	if fileHeader == nil || fileHeader.Size <= 0 || fileHeader.Size > maxSize {
		return "", ErrInvalidUpload
	}

	ext := strings.ToLower(filepath.Ext(strings.TrimSpace(fileHeader.Filename)))
	expectedContentType, ok := allowedExt[ext]
	if !ok {
		return "", ErrInvalidUpload
	}

	file, err := fileHeader.Open()
	if err != nil {
		return "", ErrInvalidUpload
	}
	defer file.Close()

	buffer := make([]byte, 512)
	readBytes, readErr := io.ReadFull(file, buffer)
	if readErr != nil && !errors.Is(readErr, io.ErrUnexpectedEOF) {
		return "", ErrInvalidUpload
	}
	if readBytes == 0 {
		return "", ErrInvalidUpload
	}

	detectedContentType := http.DetectContentType(buffer[:readBytes])
	canonicalExt, ok := detectedExt[detectedContentType]
	if !ok || detectedContentType != expectedContentType {
		return "", ErrInvalidUpload
	}

	return canonicalExt, nil
}
