package room_test

import (
	"net/http"
	"strings"
	"testing"

	"github.com/changhyeonkim/pray-together/go-api-server/internal/model"
	"github.com/changhyeonkim/pray-together/go-api-server/internal/room"
	sharedError "github.com/changhyeonkim/pray-together/go-api-server/internal/shared/error"
	sharedHttp "github.com/changhyeonkim/pray-together/go-api-server/internal/shared/http"
	"github.com/changhyeonkim/pray-together/go-api-server/internal/shared/testutil"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
)

// setupTestEnvironment creates all dependencies needed for room handler tests
// Returns the handler, database, and a test member
func setupTestEnvironment(t *testing.T) (*room.RoomHandler, *gorm.DB, int64) {
	t.Helper()

	// Setup test database
	db := testutil.SetupTestDB(t)
	t.Cleanup(func() {
		testutil.CleanupTestDB(t, db)
	})

	// Create a test member for authenticated requests
	testMember := testutil.CreateTestMember(t, db)

	// Setup dependencies - need to import member package
	roomRepo := room.NewRoomRepository()
	memberRoomRepo := room.NewMemberRoomRepository()

	// Create member service for validation
	memberRepo := testutil.NewMemberRepository()
	memberService := testutil.NewMemberService(db, memberRepo)

	roomService := room.NewRoomService(db, roomRepo, memberRoomRepo, memberService)
	roomHandler := room.NewRoomHandler(roomService)

	return roomHandler, db, int64(testMember.ID)
}

func TestCreateRoom_Success(t *testing.T) {
	// Given: Setup test environment
	roomHandler, db, memberID := setupTestEnvironment(t)

	router := testutil.SetupAuthenticatedRouter(memberID)
	router.POST("/api/v1/rooms", roomHandler.CreateRoom)

	// Given: Valid create room request
	request := testutil.TestRequest{
		Method: http.MethodPost,
		URL:    "/api/v1/rooms",
		Body: room.CreateRoomRequest{
			Name:        "Test Room",
			Description: "This is a test room",
		},
	}

	// When: Execute create room request
	recorder := testutil.ExecuteRequest(t, router, request)

	// Then: Verify response
	assert.Equal(t, http.StatusCreated, recorder.Code)

	var response sharedHttp.MessageResponse
	testutil.ParseResponse(t, recorder, &response)
	assert.Equal(t, "방 생성을 완료했습니다.", response.Message)

	// Verify room was created in database
	var createdRoom model.Room
	err := db.Table(createdRoom.TableName()).Where("name = ?", "Test Room").First(&createdRoom).Error
	assert.NoError(t, err)
	assert.NotEmpty(t, createdRoom.ID)
	assert.Equal(t, "Test Room", createdRoom.Name)
	assert.Equal(t, "This is a test room", createdRoom.Description)
}

func TestCreateRoom_ValidationError_MissingRequiredFields(t *testing.T) {
	// Given: Setup test environment
	roomHandler, _, memberID := setupTestEnvironment(t)

	router := testutil.SetupAuthenticatedRouter(memberID)
	router.POST("/api/v1/rooms", roomHandler.CreateRoom)

	testCases := []struct {
		name        string
		requestBody map[string]string
		description string
	}{
		{
			name: "Missing name",
			requestBody: map[string]string{
				"description": "This is a test room",
			},
			description: "Should fail when name is missing",
		},
		{
			name: "Missing description",
			requestBody: map[string]string{
				"name": "Test Room",
			},
			description: "Should fail when description is missing",
		},
		{
			name: "Empty name",
			requestBody: map[string]string{
				"name":        "",
				"description": "This is a test room",
			},
			description: "Should fail when name is empty",
		},
		{
			name: "Empty description",
			requestBody: map[string]string{
				"name":        "Test Room",
				"description": "",
			},
			description: "Should fail when description is empty",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// When: Execute request with invalid field
			request := testutil.TestRequest{
				Method: http.MethodPost,
				URL:    "/api/v1/rooms",
				Body:   tc.requestBody,
			}

			recorder := testutil.ExecuteRequest(t, router, request)

			// Then: Verify validation error
			assert.Equal(t, http.StatusBadRequest, recorder.Code, tc.description)

			var errorResponse sharedError.ErrorResponse
			testutil.ParseResponse(t, recorder, &errorResponse)
			assert.NotEmpty(t, errorResponse.Status, tc.description)
			assert.NotEmpty(t, errorResponse.Message, tc.description)
			assert.NotEmpty(t, errorResponse.Code, tc.description)
		})
	}
}

func TestCreateRoom_ValidationError_NameTooLong(t *testing.T) {
	// Given: Setup test environment
	roomHandler, _, memberID := setupTestEnvironment(t)

	router := testutil.SetupAuthenticatedRouter(memberID)
	router.POST("/api/v1/rooms", roomHandler.CreateRoom)

	// Given: Request with name exceeding 50 characters
	longName := strings.Repeat("a", 51) // 51 characters
	request := testutil.TestRequest{
		Method: http.MethodPost,
		URL:    "/api/v1/rooms",
		Body: room.CreateRoomRequest{
			Name:        longName,
			Description: "This is a test room",
		},
	}

	// When: Execute request
	recorder := testutil.ExecuteRequest(t, router, request)

	// Then: Verify validation error
	assert.Equal(t, http.StatusBadRequest, recorder.Code)

	var errorResponse sharedError.ErrorResponse
	testutil.ParseResponse(t, recorder, &errorResponse)
	assert.NotEmpty(t, errorResponse.Message)
}

func TestCreateRoom_ValidationError_DescriptionTooLong(t *testing.T) {
	// Given: Setup test environment
	roomHandler, _, memberID := setupTestEnvironment(t)

	router := testutil.SetupAuthenticatedRouter(memberID)
	router.POST("/api/v1/rooms", roomHandler.CreateRoom)

	// Given: Request with description exceeding 200 characters
	longDescription := strings.Repeat("a", 201) // 201 characters
	request := testutil.TestRequest{
		Method: http.MethodPost,
		URL:    "/api/v1/rooms",
		Body: room.CreateRoomRequest{
			Name:        "Test Room",
			Description: longDescription,
		},
	}

	// When: Execute request
	recorder := testutil.ExecuteRequest(t, router, request)

	// Then: Verify validation error
	assert.Equal(t, http.StatusBadRequest, recorder.Code)

	var errorResponse sharedError.ErrorResponse
	testutil.ParseResponse(t, recorder, &errorResponse)
	assert.NotEmpty(t, errorResponse.Message)
}

func TestCreateRoom_Validation_NameAtMaxLength(t *testing.T) {
	// Given: Setup test environment
	roomHandler, _, memberID := setupTestEnvironment(t)

	router := testutil.SetupAuthenticatedRouter(memberID)
	router.POST("/api/v1/rooms", roomHandler.CreateRoom)

	// Given: Request with name at exactly 50 characters (should succeed)
	maxLengthName := strings.Repeat("a", 50) // Exactly 50 characters
	request := testutil.TestRequest{
		Method: http.MethodPost,
		URL:    "/api/v1/rooms",
		Body: room.CreateRoomRequest{
			Name:        maxLengthName,
			Description: "This is a test room",
		},
	}

	// When: Execute request
	recorder := testutil.ExecuteRequest(t, router, request)

	// Then: Verify success
	assert.Equal(t, http.StatusCreated, recorder.Code)
}

func TestCreateRoom_Validation_DescriptionAtMaxLength(t *testing.T) {
	// Given: Setup test environment
	roomHandler, _, memberID := setupTestEnvironment(t)

	router := testutil.SetupAuthenticatedRouter(memberID)
	router.POST("/api/v1/rooms", roomHandler.CreateRoom)

	// Given: Request with description at exactly 200 characters (should succeed)
	maxLengthDescription := strings.Repeat("a", 200) // Exactly 200 characters
	request := testutil.TestRequest{
		Method: http.MethodPost,
		URL:    "/api/v1/rooms",
		Body: room.CreateRoomRequest{
			Name:        "Test Room",
			Description: maxLengthDescription,
		},
	}

	// When: Execute request
	recorder := testutil.ExecuteRequest(t, router, request)

	// Then: Verify success
	assert.Equal(t, http.StatusCreated, recorder.Code)
}
