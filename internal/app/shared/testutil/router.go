package testutil

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http/httptest"
	"strconv"
	"testing"

	sharedHttp "github.com/changhyeonkim/pray-together/go-api-server/internal/app/shared/http"
	"github.com/changhyeonkim/pray-together/go-api-server/internal/app/shared/validator"
	"github.com/gin-gonic/gin"
)

// SetupTestRouter creates a test Gin router without middleware
func SetupTestRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)

	// Register custom validators for testing
	_ = validator.RegisterAll()

	return gin.New()
}

// SetupAuthenticatedRouter creates a test router with memberID set in context
// This simulates the RESULT of JWT middleware (memberID in context) without actual token validation.
// Use this for testing authenticated endpoints without dealing with actual JWT tokens.
//
// In production: Request → JWT Middleware (validate token) → set memberID → Handler
// In test: Request → Mock Middleware (directly set memberID) → Handler
func SetupAuthenticatedRouter(memberID int64) *gin.Engine {
	router := SetupTestRouter()

	// Simulate the result of JWT middleware: memberID in context as string
	// JWT stores memberID as string, so we do the same in tests
	memberIDStr := strconv.FormatInt(int64(memberID), 10)
	router.Use(func(c *gin.Context) {
		c.Set(sharedHttp.MemberIDKey, memberIDStr)
		c.Next()
	})

	return router
}

// MakeRequest is a helper to make HTTP requests in tests
type TestRequest struct {
	Method      string
	URL         string
	Body        interface{}
	AccessToken string // Optional JWT token for authenticated requests
}

// ExecuteRequest executes a test HTTP request and returns the response
func ExecuteRequest(t *testing.T, router *gin.Engine, req TestRequest) *httptest.ResponseRecorder {
	t.Helper()

	var bodyReader io.Reader
	if req.Body != nil {
		bodyBytes, err := json.Marshal(req.Body)
		if err != nil {
			t.Fatalf("Failed to marshal request body: %v", err)
		}
		bodyReader = bytes.NewReader(bodyBytes)
	}

	httpReq := httptest.NewRequest(req.Method, req.URL, bodyReader)
	httpReq.Header.Set("Content-Type", "application/json")

	// Add JWT token if provided
	if req.AccessToken != "" {
		httpReq.Header.Set("Authorization", "Bearer "+req.AccessToken)
	}

	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, httpReq)

	return recorder
}

// ParseResponse parses the JSON response body into the given struct
func ParseResponse(t *testing.T, recorder *httptest.ResponseRecorder, v interface{}) {
	t.Helper()

	if err := json.Unmarshal(recorder.Body.Bytes(), v); err != nil {
		t.Fatalf("Failed to parse response body: %v", err)
	}
}
