package auth

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"net/url"

	abpkg "sambhav/pkg/authboss"

	"github.com/gin-gonic/gin"
)

// AuthHandler provides API endpoints that proxy requests into the
// authboss router so clients can authenticate via JSON-based API calls
// rather than HTML forms.
type AuthHandler struct{}

func NewAuthHandler() *AuthHandler { return &AuthHandler{} }

// Login accepts JSON {"email":"...", "password":"..."} and forwards
// it to the authboss login handler. It copies status, headers and body back
// to the Gin response, including cookies.
func (h *AuthHandler) Login(c *gin.Context) {
	// Read raw body
	bodyBuf := new(bytes.Buffer)
	_, _ = bodyBuf.ReadFrom(c.Request.Body)

	// Build a new internal request to the authboss login path.
	req := httptest.NewRequest(http.MethodPost, "/login", bytes.NewReader(bodyBuf.Bytes()))
	req.Header = c.Request.Header.Clone()
	// Ensure content-type
	if req.Header.Get("Content-Type") == "" {
		req.Header.Set("Content-Type", "application/json")
	}

	rec := httptest.NewRecorder()
	// Use the authboss router directly
	abpkg.Router().ServeHTTP(rec, req)

	// Copy headers (including Set-Cookie)
	for k, vals := range rec.HeaderMap {
		for _, v := range vals {
			c.Writer.Header().Add(k, v)
		}
	}

	c.Status(rec.Code)
	c.Writer.Write(rec.Body.Bytes())
}

// GoogleCallback accepts JSON {"code":"...","state":"..."} and proxies
// to the authboss oauth2 callback handler for the google provider.
func (h *AuthHandler) GoogleCallback(c *gin.Context) {
	var payload struct {
		Code  string `json:"code"`
		State string `json:"state"`
	}
	if err := c.BindJSON(&payload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid payload"})
		return
	}

	// Construct query params as the authboss oauth2 callback expects GET with code+state
	params := url.Values{}
	params.Set("code", payload.Code)
	if payload.State != "" {
		params.Set("state", payload.State)
	}

	req := httptest.NewRequest(http.MethodGet, "/oauth2/google/callback?"+params.Encode(), nil)
	req.Header = c.Request.Header.Clone()

	rec := httptest.NewRecorder()
	abpkg.Router().ServeHTTP(rec, req)

	for k, vals := range rec.HeaderMap {
		for _, v := range vals {
			c.Writer.Header().Add(k, v)
		}
	}

	c.Status(rec.Code)
	c.Writer.Write(rec.Body.Bytes())
}
