package orderpresentation

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	neturl "net/url"
	"time"
)

// cloudinaryProofUploader uploads payment proof files to Cloudinary using the
// same account credentials and HTTP Basic Auth pattern already used by
// cmd/cloudinary_sync for Admin API reads (api_key:api_secret). Cloudinary's
// upload endpoint accepts this Basic Auth in place of a computed signature.
type cloudinaryProofUploader struct {
	cloudName  string
	apiKey     string
	apiSecret  string
	httpClient *http.Client
	// apiBaseURL overrides the Cloudinary API host; only ever set by tests
	// to point at a local mock server instead of the real Cloudinary API.
	apiBaseURL string
}

func newCloudinaryProofUploader(cloudName string, apiKey string, apiSecret string) *cloudinaryProofUploader {
	return &cloudinaryProofUploader{
		cloudName:  cloudName,
		apiKey:     apiKey,
		apiSecret:  apiSecret,
		httpClient: &http.Client{Timeout: 30 * time.Second},
	}
}

func (u *cloudinaryProofUploader) configured() bool {
	return u != nil && u.cloudName != "" && u.apiKey != "" && u.apiSecret != ""
}

func (u *cloudinaryProofUploader) apiBase() string {
	if u.apiBaseURL != "" {
		return u.apiBaseURL
	}
	return "https://api.cloudinary.com"
}

type cloudinaryUploadResponse struct {
	SecureURL string `json:"secure_url"`
	PublicID  string `json:"public_id"`
}

// upload streams file into Cloudinary under the given folder and returns the
// resulting secure_url and public_id.
func (u *cloudinaryProofUploader) upload(file io.Reader, filename string, folder string) (secureURL string, publicID string, err error) {
	var body bytes.Buffer
	writer := multipart.NewWriter(&body)

	if err := writer.WriteField("folder", folder); err != nil {
		return "", "", err
	}
	part, err := writer.CreateFormFile("file", filename)
	if err != nil {
		return "", "", err
	}
	if _, err := io.Copy(part, file); err != nil {
		return "", "", err
	}
	if err := writer.Close(); err != nil {
		return "", "", err
	}

	url := fmt.Sprintf("%s/v1_1/%s/auto/upload", u.apiBase(), u.cloudName)
	req, err := http.NewRequest(http.MethodPost, url, &body)
	if err != nil {
		return "", "", err
	}
	req.SetBasicAuth(u.apiKey, u.apiSecret)
	req.Header.Set("Content-Type", writer.FormDataContentType())

	resp, err := u.httpClient.Do(req)
	if err != nil {
		return "", "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		var errBody bytes.Buffer
		errBody.ReadFrom(resp.Body)
		return "", "", fmt.Errorf("cloudinary upload failed: %s: %s", resp.Status, errBody.String())
	}

	var uploadResp cloudinaryUploadResponse
	if err := json.NewDecoder(resp.Body).Decode(&uploadResp); err != nil {
		return "", "", err
	}
	if uploadResp.SecureURL == "" {
		return "", "", fmt.Errorf("cloudinary upload response missing secure_url")
	}

	return uploadResp.SecureURL, uploadResp.PublicID, nil
}

// destroy best-effort removes an uploaded asset (e.g. when the surrounding
// order submission fails after the file already reached Cloudinary). Errors
// are intentionally swallowed: a failed cleanup just leaves one orphaned
// asset, which is preferable to failing the customer-facing request over it.
func (u *cloudinaryProofUploader) destroy(publicID string) {
	if u == nil || publicID == "" {
		return
	}

	url := fmt.Sprintf("%s/v1_1/%s/resources/image/upload?public_ids[]=%s", u.apiBase(), u.cloudName, neturl.QueryEscape(publicID))
	req, err := http.NewRequest(http.MethodDelete, url, nil)
	if err != nil {
		return
	}
	req.SetBasicAuth(u.apiKey, u.apiSecret)

	resp, err := u.httpClient.Do(req)
	if err != nil {
		return
	}
	resp.Body.Close()
}
