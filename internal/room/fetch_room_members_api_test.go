package room_test

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/changhyeonkim/pray-together/go-api-server/internal/room"
	sharedError "github.com/changhyeonkim/pray-together/go-api-server/internal/shared/error"
	"github.com/changhyeonkim/pray-together/go-api-server/internal/shared/testutil"
	"github.com/stretchr/testify/assert"
)

func TestFetchRoomMembers_Success(t *testing.T) {
	// Given: Setup test environment
	roomHandler, db, ownerID := setupTestEnvironment(t)
	router := testutil.SetupAuthenticatedRouter(ownerID)
	router.GET("/api/v1/rooms/:roomId/members", roomHandler.FetchRoomMembers)

	// Create room and add members
	testRoom := testutil.CreateTestRoom(t, db, ownerID, "Test Room", "Test Description")
	const totalMembers = 10
	const additionalMembersCount = totalMembers - 1 // exclude owner

	// Add additional members to room
	testutil.AddMembersToRoom(t, db, testRoom.ID, additionalMembersCount)

	// When: Execute fetch room members request
	request := testutil.TestRequest{
		Method: http.MethodGet,
		URL:    fmt.Sprintf("/api/v1/rooms/%d/members", testRoom.ID),
		Body:   nil,
	}
	recorder := testutil.ExecuteRequest(t, router, request)

	// Then: Verify response
	assert.Equal(t, http.StatusOK, recorder.Code)

	var response room.FetchRoomMemberResponse
	testutil.ParseResponse(t, recorder, &response)

	// Verify member count
	assert.NotNil(t, response.RoomMembers, "room members should not be null")
	assert.Len(t, response.RoomMembers, totalMembers, "expected 10 room members")

	// Verify owner is included
	ownerFound := false
	for _, member := range response.RoomMembers {
		if member.ID == ownerID {
			ownerFound = true
			break
		}
	}
	assert.True(t, ownerFound, "owner should be included in room members")

	// Verify all members have phone number suffix
	for _, member := range response.RoomMembers {
		assert.NotEmpty(t, member.PhoneNumberSuffix, "phone number suffix should not be empty")
		assert.Len(t, member.PhoneNumberSuffix, 4, "phone number suffix should be 4 digits")
	}

	// Verify members are sorted by name (ascending)
	for i := 0; i < len(response.RoomMembers)-1; i++ {
		assert.LessOrEqual(t, response.RoomMembers[i].Name, response.RoomMembers[i+1].Name,
			"members should be sorted by name in ascending order")
	}
}

func TestFetchRoomMembers_InvalidRoomID(t *testing.T) {
	// Given: Setup test environment
	roomHandler, _, ownerID := setupTestEnvironment(t)
	router := testutil.SetupAuthenticatedRouter(ownerID)
	router.GET("/api/v1/rooms/:roomId/members", roomHandler.FetchRoomMembers)

	testCases := []struct {
		name          string
		roomID        string
		expectedError string
	}{
		{
			name:          "Negative room ID",
			roomID:        "-1",
			expectedError: "Key: 'FetchRoomMemberRequest.RoomID' Error:Field validation for 'RoomID' failed on the 'gt' tag",
		},
		{
			name:          "Zero room ID",
			roomID:        "0",
			expectedError: "Key: 'FetchRoomMemberRequest.RoomID' Error:Field validation for 'RoomID' failed on the 'gt' tag",
		},
		{
			name:          "Invalid string room ID",
			roomID:        "invalid",
			expectedError: "invalid URI",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// When: Execute request with invalid room ID
			request := testutil.TestRequest{
				Method: http.MethodGet,
				URL:    fmt.Sprintf("/api/v1/rooms/%s/members", tc.roomID),
				Body:   nil,
			}
			recorder := testutil.ExecuteRequest(t, router, request)

			// Then: Verify error response
			assert.Equal(t, http.StatusBadRequest, recorder.Code)
		})
	}
}

func TestFetchRoomMembers_NonexistentRoom(t *testing.T) {
	// Given: Setup test environment
	roomHandler, _, ownerID := setupTestEnvironment(t)
	router := testutil.SetupAuthenticatedRouter(ownerID)
	router.GET("/api/v1/rooms/:roomId/members", roomHandler.FetchRoomMembers)

	nonexistentRoomID := int64(99999)

	// When: Execute request for nonexistent room
	request := testutil.TestRequest{
		Method: http.MethodGet,
		URL:    fmt.Sprintf("/api/v1/rooms/%d/members", nonexistentRoomID),
		Body:   nil,
	}
	recorder := testutil.ExecuteRequest(t, router, request)

	// Then: Verify error response
	assert.Equal(t, http.StatusNotFound, recorder.Code)

	var errorResponse sharedError.ErrorResponse
	testutil.ParseResponse(t, recorder, &errorResponse)

	assert.Equal(t, "ROOM-002", errorResponse.Code)
	assert.Equal(t, "회원이 속한 방을 찾을 수 없습니다.", errorResponse.Message)
}

func TestFetchRoomMembers_NotMemberOfRoom(t *testing.T) {
	// Given: Setup test environment
	roomHandler, db, ownerID := setupTestEnvironment(t)

	// Create another member who is not part of the room
	otherMember := testutil.CreateTestMemberWithIndex(t, db, 1)
	router := testutil.SetupAuthenticatedRouter(int64(otherMember.ID))
	router.GET("/api/v1/rooms/:roomId/members", roomHandler.FetchRoomMembers)

	// Create room with owner
	testRoom := testutil.CreateTestRoom(t, db, ownerID, "Test Room", "Test Description")

	// When: Other member tries to fetch room members
	request := testutil.TestRequest{
		Method: http.MethodGet,
		URL:    fmt.Sprintf("/api/v1/rooms/%d/members", testRoom.ID),
		Body:   nil,
	}
	recorder := testutil.ExecuteRequest(t, router, request)

	// Then: Verify error response
	assert.Equal(t, http.StatusNotFound, recorder.Code)

	var errorResponse sharedError.ErrorResponse
	testutil.ParseResponse(t, recorder, &errorResponse)

	assert.Equal(t, "ROOM-002", errorResponse.Code)
	assert.Equal(t, "회원이 속한 방을 찾을 수 없습니다.", errorResponse.Message)
}

func TestFetchRoomMembers_EmptyRoom(t *testing.T) {
	// Given: Setup test environment with empty database
	roomHandler, db, ownerID := setupTestEnvironment(t)
	router := testutil.SetupAuthenticatedRouter(ownerID)
	router.GET("/api/v1/rooms/:roomId/members", roomHandler.FetchRoomMembers)

	// Create room with only owner (no additional members)
	testRoom := testutil.CreateTestRoom(t, db, ownerID, "Empty Room", "Room with only owner")

	// When: Execute fetch room members request
	request := testutil.TestRequest{
		Method: http.MethodGet,
		URL:    fmt.Sprintf("/api/v1/rooms/%d/members", testRoom.ID),
		Body:   nil,
	}
	recorder := testutil.ExecuteRequest(t, router, request)

	// Then: Verify response
	assert.Equal(t, http.StatusOK, recorder.Code)

	var response room.FetchRoomMemberResponse
	testutil.ParseResponse(t, recorder, &response)

	// Verify only owner is in the room
	assert.NotNil(t, response.RoomMembers)
	assert.Len(t, response.RoomMembers, 1, "expected 1 room member (owner only)")

	// Verify it's the owner
	assert.Equal(t, ownerID, response.RoomMembers[0].ID)
}
