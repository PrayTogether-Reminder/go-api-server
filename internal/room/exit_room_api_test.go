package room_test

import (
	"fmt"
	"log/slog"
	"net/http"
	"testing"

	"github.com/changhyeonkim/pray-together/go-api-server/internal/model"
	sharedError "github.com/changhyeonkim/pray-together/go-api-server/internal/shared/error"
	sharedHttp "github.com/changhyeonkim/pray-together/go-api-server/internal/shared/http"
	"github.com/changhyeonkim/pray-together/go-api-server/internal/shared/testutil"
	"github.com/stretchr/testify/assert"
)

func TestExitRoom_Success(t *testing.T) {
	// Given: Setup test environment
	roomHandler, db, memberID := setupTestEnvironment(t)

	router := testutil.SetupAuthenticatedRouter(memberID)
	router.DELETE("/api/v1/rooms/:roomId", roomHandler.DeleteMemberRoom)

	// Given: Create a test room with member-room relationship
	testRoom := testutil.CreateTestRoom(t, db, memberID, "Test Room", "Test Description")

	// When: Execute exit room request
	request := testutil.TestRequest{
		Method: http.MethodDelete,
		URL:    fmt.Sprintf("/api/v1/rooms/%d", testRoom.ID),
		Body:   nil,
	}

	recorder := testutil.ExecuteRequest(t, router, request)

	// Then: Verify response
	assert.Equal(t, http.StatusOK, recorder.Code)

	var response sharedHttp.MessageResponse
	testutil.ParseResponse(t, recorder, &response)
	assert.Equal(t, "방을 나갔습니다.", response.Message)

	// Verify member_room relationship was deleted from database
	var memberRoom model.MemberRoom
	err := db.Table(memberRoom.TableName()).
		Where("member_id = ? AND room_id = ?", memberID, testRoom.ID).
		First(&memberRoom).Error
	assert.Error(t, err, "member_room relationship should be deleted")
}

func TestExitRoom_InvalidID(t *testing.T) {
	// Given: Setup test environment
	roomHandler, _, memberID := setupTestEnvironment(t)

	router := testutil.SetupAuthenticatedRouter(memberID)
	router.DELETE("/api/v1/rooms/:roomId", roomHandler.DeleteMemberRoom)

	testCases := []struct {
		name        string
		roomID      string
		description string
	}{
		{
			name:        "Negative number",
			roomID:      "-1",
			description: "Should fail when roomId is negative",
		},
		{
			name:        "Zero",
			roomID:      "0",
			description: "Should fail when roomId is zero",
		},
		{
			name:        "String",
			roomID:      "abc",
			description: "Should fail when roomId is a string",
		},
		{
			name:        "Special characters",
			roomID:      "!@#",
			description: "Should fail when roomId contains special characters",
		},
		{
			name:        "Decimal",
			roomID:      "1.5",
			description: "Should fail when roomId is a decimal",
		},
		{
			name:        "Whitespace",
			roomID:      "%20",
			description: "Should fail when roomId is whitespace (URL encoded)",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// When: Execute request with invalid roomId
			request := testutil.TestRequest{
				Method: http.MethodDelete,
				URL:    fmt.Sprintf("/api/v1/rooms/%s", tc.roomID),
				Body:   nil,
			}

			recorder := testutil.ExecuteRequest(t, router, request)

			// Then: Verify bad request error
			assert.Equal(t, http.StatusBadRequest, recorder.Code, tc.description)

			var errorResponse sharedError.ErrorResponse
			testutil.ParseResponse(t, recorder, &errorResponse)
			slog.Info("[Validation]", "message", errorResponse.Message)
			assert.NotEmpty(t, errorResponse.Status, tc.description)
			assert.NotEmpty(t, errorResponse.Message, tc.description)
		})
	}
}

func TestExitRoom_NonexistentRoom(t *testing.T) {
	// Given: Setup test environment
	roomHandler, _, memberID := setupTestEnvironment(t)

	router := testutil.SetupAuthenticatedRouter(memberID)
	router.DELETE("/api/v1/rooms/:roomId", roomHandler.DeleteMemberRoom)

	// Given: Nonexistent room ID
	nonexistentRoomID := int64(999999)

	// When: Execute exit room request with nonexistent room ID
	request := testutil.TestRequest{
		Method: http.MethodDelete,
		URL:    fmt.Sprintf("/api/v1/rooms/%d", nonexistentRoomID),
		Body:   nil,
	}

	recorder := testutil.ExecuteRequest(t, router, request)

	// Then: Verify not found error
	assert.Equal(t, http.StatusNotFound, recorder.Code)

	var errorResponse sharedError.ErrorResponse
	testutil.ParseResponse(t, recorder, &errorResponse)
	assert.Equal(t, http.StatusNotFound, errorResponse.Status)
	assert.Equal(t, "ROOM-002", errorResponse.Code)
	assert.Equal(t, "회원이 속한 방을 찾을 수 없습니다.", errorResponse.Message)
}
