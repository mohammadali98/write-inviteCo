package webui

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

const (
	csrfCookieName = "csrf_token"
	csrfFieldName  = "csrf_token"
)

func EnsureCSRFToken(c *gin.Context) string {
	token, err := c.Cookie(csrfCookieName)
	if err == nil && strings.TrimSpace(token) != "" {
		return token
	}

	token = generateCSRFToken()
	setCSRFCookie(c, token)
	return token
}

func ValidateCSRF(c *gin.Context) bool {
	cookieToken, err := c.Cookie(csrfCookieName)
	if err != nil || strings.TrimSpace(cookieToken) == "" {
		return false
	}

	formToken := strings.TrimSpace(c.PostForm(csrfFieldName))
	if formToken == "" {
		return false
	}

	return subtle.ConstantTimeCompare([]byte(cookieToken), []byte(formToken)) == 1
}

func CSRFFieldName() string {
	return csrfFieldName
}

func RenderError(c *gin.Context, statusCode int, title string, message string) {
	c.HTML(statusCode, "error.html", gin.H{
		"statusCode": statusCode,
		"title":      title,
		"message":    message,
	})
}

func generateCSRFToken() string {
	buffer := make([]byte, 32)
	if _, err := rand.Read(buffer); err != nil {
		return base64.RawURLEncoding.EncodeToString([]byte("writeandinviteco-csrf-fallback-token"))
	}
	return base64.RawURLEncoding.EncodeToString(buffer)
}

func setCSRFCookie(c *gin.Context, token string) {
	secure := c.Request.TLS != nil || strings.EqualFold(c.GetHeader("X-Forwarded-Proto"), "https")
	c.SetSameSite(http.SameSiteLaxMode)
	c.SetCookie(csrfCookieName, token, 60*60*12, "/", "", secure, true)
}
