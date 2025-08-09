package test

import (
	"bytes"
	"encoding/json"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	authdomain "pray-together/internal/domains/auth/domain"
	invitationdomain "pray-together/internal/domains/invitation/domain"
	memberdomain "pray-together/internal/domains/member/domain"
	notificationdomain "pray-together/internal/domains/notification/domain"
	prayerdomain "pray-together/internal/domains/prayer/domain"
	roomdomain "pray-together/internal/domains/room/domain"
)

// PrayerCompletionNotification represents prayer completion notification for tests
type PrayerCompletionNotification struct {
	ID            uint64 `gorm:"primaryKey;autoIncrement;column:id" json:"id"`
	SenderID      uint64 `gorm:"column:sender_id;not null" json:"senderId"`
	ReceiverID    uint64 `gorm:"column:receiver_id;not null" json:"receiverId"`
	PrayerTitleID uint64 `gorm:"column:prayer_title_id;not null" json:"prayerTitleId"`
	Message       string `gorm:"column:message;type:text" json:"message"`
}

// TableName specifies the table name
func (PrayerCompletionNotification) TableName() string {
	return "prayer_completion_notification"
}

// IntegrationTestSuite is the base test suite for integration tests
type IntegrationTestSuite struct {
	suite.Suite
	db     *gorm.DB
	router *gin.Engine

	// Test utilities
	testUtils *TestUtils

	// API URLs matching Java constants
	APIVersion        string
	RoomsAPIURL       string
	PrayersAPIURL     string
	MembersAPIURL     string
	InvitationsAPIURL string
}

// SetupSuite runs once before all tests in the suite
func (suite *IntegrationTestSuite) SetupSuite() {
	// Set Gin to test mode
	gin.SetMode(gin.TestMode)

	// Set API URLs matching Java
	suite.APIVersion = "/api/v1"
	suite.RoomsAPIURL = suite.APIVersion + "/rooms"
	suite.PrayersAPIURL = suite.APIVersion + "/prayers"
	suite.MembersAPIURL = suite.APIVersion + "/members"
	suite.InvitationsAPIURL = suite.APIVersion + "/invitations"
}

// SetupTest runs before each individual test
func (suite *IntegrationTestSuite) SetupTest() {
	// Set up in-memory SQLite database for testing (new for each test)
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	require.NoError(suite.T(), err)

	suite.db = db

	// Migrate all tables
	err = suite.db.AutoMigrate(
		&invitationdomain.Invitation{},
		&memberdomain.Member{},
		&roomdomain.Room{},
		&roomdomain.RoomMember{},
		&prayerdomain.PrayerTitle{},
		&prayerdomain.PrayerContent{},
		&prayerdomain.PrayerCompletion{},
		&authdomain.RefreshToken{},
		&notificationdomain.Notification{},
		&PrayerCompletionNotification{}, // Test-specific table
		// Add other models as needed
	)
	require.NoError(suite.T(), err)

	// Initialize router with actual handlers
	suite.router = SetupTestRouter(suite.db)

	// Initialize test utilities
	suite.testUtils = NewTestUtils(suite.db)
}

// TearDownTest runs after each individual test
func (suite *IntegrationTestSuite) TearDownTest() {
	// Clean up after each test
	suite.CleanRepository()
}

// CleanRepository cleans all database tables in the correct order
// Order is very important to avoid foreign key constraint violations
func (suite *IntegrationTestSuite) CleanRepository() {
	// Delete in reverse dependency order
	suite.db.Exec("DELETE FROM prayer_completion_notification")
	suite.db.Exec("DELETE FROM notification")
	suite.db.Exec("DELETE FROM prayer_completion")
	suite.db.Exec("DELETE FROM invitation")
	suite.db.Exec("DELETE FROM prayer_content")
	suite.db.Exec("DELETE FROM prayer_title")
	suite.db.Exec("DELETE FROM member_room")
	suite.db.Exec("DELETE FROM room")
	suite.db.Exec("DELETE FROM member")
	suite.db.Exec("DELETE FROM refresh_token")
}

// Helper methods for HTTP requests

// PostJSON makes a POST request with JSON body
func (suite *IntegrationTestSuite) PostJSON(url string, body interface{}, headers map[string]string) *httptest.ResponseRecorder {
	jsonBody, err := json.Marshal(body)
	require.NoError(suite.T(), err)

	req := httptest.NewRequest("POST", url, bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")

	// Add custom headers
	for key, value := range headers {
		req.Header.Set(key, value)
	}

	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	return w
}

// GetRequest makes a GET request
func (suite *IntegrationTestSuite) GetRequest(url string, headers map[string]string) *httptest.ResponseRecorder {
	req := httptest.NewRequest("GET", url, nil)

	// Add custom headers
	for key, value := range headers {
		req.Header.Set(key, value)
	}

	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	return w
}

// PutJSON makes a PUT request with JSON body
func (suite *IntegrationTestSuite) PutJSON(url string, body interface{}, headers map[string]string) *httptest.ResponseRecorder {
	jsonBody, err := json.Marshal(body)
	require.NoError(suite.T(), err)

	req := httptest.NewRequest("PUT", url, bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")

	// Add custom headers
	for key, value := range headers {
		req.Header.Set(key, value)
	}

	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	return w
}

// DeleteRequest makes a DELETE request
func (suite *IntegrationTestSuite) DeleteRequest(url string, headers map[string]string) *httptest.ResponseRecorder {
	req := httptest.NewRequest("DELETE", url, nil)

	// Add custom headers
	for key, value := range headers {
		req.Header.Set(key, value)
	}

	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	return w
}

// AssertStatusCode asserts the HTTP status code
func (suite *IntegrationTestSuite) AssertStatusCode(w *httptest.ResponseRecorder, expectedCode int, message string) {
	assert.Equal(suite.T(), expectedCode, w.Code, message)
}

// AssertJSONResponse asserts the JSON response body
func (suite *IntegrationTestSuite) AssertJSONResponse(w *httptest.ResponseRecorder, expected interface{}, message string) {
	var actual interface{}
	err := json.Unmarshal(w.Body.Bytes(), &actual)
	require.NoError(suite.T(), err, "Failed to unmarshal response JSON")

	assert.Equal(suite.T(), expected, actual, message)
}

// AssertMessageResponse asserts the message response format
func (suite *IntegrationTestSuite) AssertMessageResponse(w *httptest.ResponseRecorder, expectedMessage string, message string) {
	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(suite.T(), err, "Failed to unmarshal response JSON")

	actualMessage, exists := response["message"]
	require.True(suite.T(), exists, "Response should contain 'message' field")
	assert.Equal(suite.T(), expectedMessage, actualMessage, message)
}

// PatchJSON makes a PATCH request with JSON body
func (suite *IntegrationTestSuite) PatchJSON(url string, body interface{}, headers map[string]string) *httptest.ResponseRecorder {
	jsonBody, err := json.Marshal(body)
	require.NoError(suite.T(), err)

	req := httptest.NewRequest("PATCH", url, bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")

	// Add custom headers
	for key, value := range headers {
		req.Header.Set(key, value)
	}

	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	return w
}

// UnmarshalResponse unmarshals the response body into the provided interface
func (suite *IntegrationTestSuite) UnmarshalResponse(w *httptest.ResponseRecorder, target interface{}) error {
	return json.Unmarshal(w.Body.Bytes(), target)
}

// RunIntegrationTest runs a test suite
func RunIntegrationTest(t *testing.T, testSuite suite.TestingSuite) {
	suite.Run(t, testSuite)
}
