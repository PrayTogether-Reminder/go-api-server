package test

import (
	"fmt"
	"net/http"
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"

	memberdomain "pray-together/internal/domains/member/domain"
	roomdomain "pray-together/internal/domains/room/domain"
)

// RoomDeleteIntegrationTestSuite tests room delete API (matching Java RoomDeleteIntegrateTest)
type RoomDeleteIntegrationTestSuite struct {
	IntegrationTestSuite

	member   *memberdomain.Member
	headers  map[string]string
	testRoom *roomdomain.Room
}

// SetupTest runs before each test (matching Java @BeforeEach setup)
func (suite *RoomDeleteIntegrationTestSuite) SetupTest() {
	// Call parent SetupTest
	suite.IntegrationTestSuite.SetupTest()

	suite.member = suite.testUtils.CreateUniqueMember()
	// memberRepository.save(member) - already saved in CreateUniqueMember
	suite.headers = suite.testUtils.CreateAuthHeaderWithMember(suite.member)

	// Set up router with room routes
}

// TearDownTest runs after each test (matching Java @AfterEach cleanup)
func (suite *RoomDeleteIntegrationTestSuite) TearDownTest() {
	suite.CleanRepository()
}

// TestDeleteRoomWhenRoomExistsThenReturn200OK tests deleting existing room (matching Java delete_room_when_room_exists_then_return_200_ok)
func (suite *RoomDeleteIntegrationTestSuite) TestDeleteRoomWhenRoomExistsThenReturn200OK() {
	// Given
	// 방 생성
	requestDto := RoomCreateRequest{
		Name:        "삭제 예정 방",
		Description: "테스트를 위해 삭제하려는 방 입니다.",
	}

	w := suite.PostJSON(suite.RoomsAPIURL, requestDto, suite.headers)
	assert.Equal(suite.T(), http.StatusCreated, w.Code)

	// 방 정보 획득
	var allRooms []roomdomain.Room
	suite.db.Find(&allRooms)
	suite.testRoom = &allRooms[0]

	// When
	deleteURL := fmt.Sprintf("%s/%d", suite.RoomsAPIURL, suite.testRoom.ID)
	w = suite.DeleteRequest(deleteURL, suite.headers)

	// Then
	// 삭제 응답 상태 검증
	assert.Equal(suite.T(), http.StatusOK, w.Code)

	// memberRoom 삭제 확인
	var allMemberRooms []roomdomain.RoomMember
	suite.db.Find(&allMemberRooms)
	assert.Empty(suite.T(), allMemberRooms)
}

// TestDeleteRoomWithInvalidIdThenReturn400BadRequest tests deleting room with invalid ID (matching Java delete_room_with_invalid_id_then_return_400_bad_request)
func (suite *RoomDeleteIntegrationTestSuite) TestDeleteRoomWithInvalidIdThenReturn400BadRequest() {
	testCases := []struct {
		name       string
		encodedURL string
	}{
		{"음수 ID", url.QueryEscape("-1")},
		{"0 ID", url.QueryEscape("0")},
		{"문자열 ID", url.QueryEscape("abc")},
		{"특수문자 ID", url.QueryEscape("!@#")},
		{"소수점 ID", url.QueryEscape("1.5")},
		{"공백 ID", url.QueryEscape(" ")},
		{"null", "null"},
	}

	for _, tc := range testCases {
		suite.T().Run(tc.name, func(t *testing.T) {
			// Given
			deleteURL := fmt.Sprintf("%s/%s", suite.RoomsAPIURL, tc.encodedURL)

			// When
			w := suite.DeleteRequest(deleteURL, suite.headers)

			// Then
			assert.Equal(t, http.StatusBadRequest, w.Code)

			var exceptionResponse ExceptionResponse
			err := suite.UnmarshalResponse(w, &exceptionResponse)
			assert.NoError(t, err)
			assert.NotNil(t, exceptionResponse)
		})
	}
}

// TestDeleteRoomWithNonexistentIdThenReturn404NotFound tests deleting non-existent room (matching Java delete_room_with_nonexistent_id_then_return_404_not_found)
func (suite *RoomDeleteIntegrationTestSuite) TestDeleteRoomWithNonexistentIdThenReturn404NotFound() {
	// Given
	nonExistentID := int64(999999) // 존재하지 않는 ID
	deleteURL := fmt.Sprintf("%s/%d", suite.RoomsAPIURL, nonExistentID)

	// When
	w := suite.DeleteRequest(deleteURL, suite.headers)

	// Then
	assert.Equal(suite.T(), http.StatusNotFound, w.Code)

	var errorResponse MessageResponse
	err := suite.UnmarshalResponse(w, &errorResponse)
	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), errorResponse)
}

// TestRoomDeleteIntegration runs the room delete integration test suite
func TestRoomDeleteIntegration(t *testing.T) {
	suite.Run(t, new(RoomDeleteIntegrationTestSuite))
}
