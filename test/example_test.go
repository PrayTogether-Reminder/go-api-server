package test

import (
	"net/http"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

// ExampleTestSuite demonstrates basic integration test structure
type ExampleTestSuite struct {
	IntegrationTestSuite
}

// SetupTest runs before each test
func (suite *ExampleTestSuite) SetupTest() {
	// Call parent SetupTest
	suite.IntegrationTestSuite.SetupTest()

	// Set up router for example
	suite.router = gin.New()

	// Add a simple test route
	suite.router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	suite.router.POST("/api/v1/test", func(c *gin.Context) {
		var body map[string]interface{}
		if err := c.ShouldBindJSON(&body); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusCreated, gin.H{
			"message": "테스트가 성공했습니다.",
			"data":    body,
		})
	})
}

// TestHealthCheck tests the health check endpoint
func (suite *ExampleTestSuite) TestHealthCheck() {
	// When
	w := suite.GetRequest("/health", nil)

	// Then
	suite.AssertStatusCode(w, http.StatusOK, "Health check should return 200 OK")

	expected := map[string]interface{}{
		"status": "ok",
	}
	suite.AssertJSONResponse(w, expected, "Health check should return correct JSON")
}

// TestCreateEndpoint tests a simple create endpoint
func (suite *ExampleTestSuite) TestCreateEndpoint() {
	// Given
	requestBody := map[string]interface{}{
		"name":  "test",
		"value": 123,
	}

	// When
	w := suite.PostJSON("/api/v1/test", requestBody, map[string]string{
		"Content-Type": "application/json",
	})

	// Then
	suite.AssertStatusCode(w, http.StatusCreated, "Create endpoint should return 201 Created")
	suite.AssertMessageResponse(w, "테스트가 성공했습니다.", "Should return Korean success message")
}

// TestDatabaseOperations tests basic database operations
func (suite *ExampleTestSuite) TestDatabaseOperations() {
	// Given - Create a test member
	member := suite.testUtils.CreateUniqueMember()

	// Then - Assert member was created
	assert.NotZero(suite.T(), member.ID, "Member should have a non-zero ID")
	assert.NotEmpty(suite.T(), member.Email, "Member should have an email")
	assert.NotEmpty(suite.T(), member.Name, "Member should have a name")

	// Assert count in database
	suite.testUtils.AssertMemberCount(1)

	// Create another member
	member2 := suite.testUtils.CreateUniqueMember()
	assert.NotEqual(suite.T(), member.ID, member2.ID, "Members should have different IDs")
	assert.NotEqual(suite.T(), member.Email, member2.Email, "Members should have different emails")

	// Assert count in database
	suite.testUtils.AssertMemberCount(2)
}

// TestAuthHeaders tests JWT authentication headers
func (suite *ExampleTestSuite) TestAuthHeaders() {
	// Given
	member := suite.testUtils.CreateUniqueMember()

	// When - Create auth headers
	headers := suite.testUtils.CreateAuthHeaderWithMember(member)

	// Then
	assert.Contains(suite.T(), headers, "Authorization", "Should contain Authorization header")
	assert.Contains(suite.T(), headers, "Content-Type", "Should contain Content-Type header")

	authHeader := headers["Authorization"]
	assert.Contains(suite.T(), authHeader, "Bearer ", "Authorization header should start with 'Bearer '")
	assert.Greater(suite.T(), len(authHeader), len("Bearer "), "Authorization header should contain a token")
}

// TestCleanup tests that cleanup works properly between tests
func (suite *ExampleTestSuite) TestCleanup() {
	// This test should start with empty database due to cleanup
	suite.testUtils.AssertMemberCount(0)
	suite.testUtils.AssertRoomCount(0)

	// Create some data
	member := suite.testUtils.CreateUniqueMember()
	room := suite.testUtils.CreateUniqueRoom()

	assert.NotZero(suite.T(), member.ID, "Member should be created")
	assert.NotZero(suite.T(), room.ID, "Room should be created")

	suite.testUtils.AssertMemberCount(1)
	suite.testUtils.AssertRoomCount(1)
}

// TestExampleIntegration runs the example integration test suite
func TestExampleIntegration(t *testing.T) {
	suite.Run(t, new(ExampleTestSuite))
}
