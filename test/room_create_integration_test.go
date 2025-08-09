package test

import (
	"net/http"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"

	memberdomain "pray-together/internal/domains/member/domain"
	roomdomain "pray-together/internal/domains/room/domain"
)

// RoomCreateIntegrationTestSuite tests room creation API (matching Java RoomCreateIntegrateTest)
type RoomCreateIntegrationTestSuite struct {
	IntegrationTestSuite

	member  *memberdomain.Member
	headers map[string]string
}

// SetupTest runs before each test (matching Java @BeforeEach setup)
func (suite *RoomCreateIntegrationTestSuite) SetupTest() {
	// Call parent SetupTest
	suite.IntegrationTestSuite.SetupTest()

	// Create member (matching Java)
	suite.member = suite.testUtils.CreateUniqueMember()
	// No need to save again as CreateUniqueMember already saves

	// Create auth headers (matching Java)
	suite.headers = suite.testUtils.CreateAuthHeaderWithMember(suite.member)

	// Set up router with room routes
}

// TearDownTest runs after each test (matching Java @AfterEach cleanup)
func (suite *RoomCreateIntegrationTestSuite) TearDownTest() {
	suite.CleanRepository()
}

// TestCreateRoomWithValidInput tests room creation with valid input (matching Java create_room_with_valid_input_then_return_201_created)
func (suite *RoomCreateIntegrationTestSuite) TestCreateRoomWithValidInput() {
	// Given - Request Body 준비
	requestDTO := RoomCreateRequest{
		Name:        "테스트 방",
		Description: "테스트를 위한 방입니다.",
	}

	// When - API 요청
	w := suite.PostJSON(suite.RoomsAPIURL, requestDTO, suite.headers)

	// Then
	// API 응답 검증
	suite.AssertStatusCode(w, http.StatusCreated, "Room 생성 시 201 Created를 반환해야 합니다")

	// 생성된 Room 확인
	var allRooms []roomdomain.Room
	suite.db.Find(&allRooms)
	assert.NotEmpty(suite.T(), allRooms, "Room이 생성되어야 합니다")

	createdRoom := allRooms[0]
	assert.Equal(suite.T(), "테스트 방", createdRoom.RoomName)
	assert.Equal(suite.T(), "테스트를 위한 방입니다.", getDescription(&createdRoom)) // Note: Go struct에 따라 수정 필요

	// 생성된 Member-Room 확인
	var memberRooms []roomdomain.RoomMember
	suite.db.Find(&memberRooms)
	assert.NotEmpty(suite.T(), memberRooms, "MemberRoom 관계가 생성되어야 합니다")
	assert.Equal(suite.T(), suite.member.ID, memberRooms[0].MemberID)
}

// TestCreateRoomWithInvalidInput tests room creation with invalid parameters (matching Java @ParameterizedTest)
func (suite *RoomCreateIntegrationTestSuite) TestCreateRoomWithInvalidInput() {
	// Test cases matching Java provideInvalidRoomCreateParameters
	testCases := []struct {
		name        string
		roomName    *string
		description *string
		expectedMsg string
	}{
		{
			name:        "방 이름 null",
			roomName:    nil,
			description: stringPtr("정상적인 방 설명입니다."),
			expectedMsg: "방 이름이 null일 때 400 Bad Request를 반환해야 합니다",
		},
		{
			name:        "방 이름 빈 문자열",
			roomName:    stringPtr(""),
			description: stringPtr("정상적인 방 설명입니다."),
			expectedMsg: "방 이름이 빈 문자열일 때 400 Bad Request를 반환해야 합니다",
		},
		{
			name:        "방 이름 공백만 포함",
			roomName:    stringPtr("   "),
			description: stringPtr("정상적인 방 설명입니다."),
			expectedMsg: "방 이름이 공백만 포함할 때 400 Bad Request를 반환해야 합니다",
		},
		{
			name:        "방 이름 최대 길이 초과(51자)",
			roomName:    stringPtr(strings.Repeat("a", 51)),
			description: stringPtr("정상적인 방 설명입니다."),
			expectedMsg: "방 이름이 51자를 초과할 때 400 Bad Request를 반환해야 합니다",
		},
		{
			name:        "방 설명 null",
			roomName:    stringPtr("정상적인 방 이름"),
			description: nil,
			expectedMsg: "방 설명이 null일 때 400 Bad Request를 반환해야 합니다",
		},
		{
			name:        "방 설명 빈 문자열",
			roomName:    stringPtr("정상적인 방 이름"),
			description: stringPtr(""),
			expectedMsg: "방 설명이 빈 문자열일 때 400 Bad Request를 반환해야 합니다",
		},
		{
			name:        "방 설명 최대 길이 초과(201자)",
			roomName:    stringPtr("정상적인 방 이름"),
			description: stringPtr(strings.Repeat("a", 201)),
			expectedMsg: "방 설명이 201자를 초과할 때 400 Bad Request를 반환해야 합니다",
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			// Given
			requestDTO := RoomCreateRequest{}

			if tc.roomName != nil {
				requestDTO.Name = *tc.roomName
			}
			if tc.description != nil {
				requestDTO.Description = *tc.description
			}

			// When
			w := suite.PostJSON(suite.RoomsAPIURL, requestDTO, suite.headers)

			// Then
			suite.AssertStatusCode(w, http.StatusBadRequest, tc.expectedMsg)

			// Response body 검증
			var errorResponse map[string]interface{}
			err := suite.UnmarshalResponse(w, &errorResponse)
			assert.NoError(suite.T(), err, "응답을 파싱할 수 있어야 합니다")
			assert.NotNil(suite.T(), errorResponse, "에러 응답이 있어야 합니다")

			// 방이 생성되지 않았는지 확인
			var allRooms []roomdomain.Room
			suite.db.Find(&allRooms)
			assert.Empty(suite.T(), allRooms, tc.name+": 방이 생성되면 안됩니다")
		})
	}
}

// RoomCreateRequest represents the request to create a room (matching Java RoomCreateRequest)
type RoomCreateRequest struct {
	Name        string `json:"name" binding:"required,min=1,max=50"`
	Description string `json:"description" binding:"required,max=200"`
}

// Helper function to get description (adjust based on actual Room struct)
func getDescription(room *roomdomain.Room) string {
	return room.Description
}

// TestRoomCreateIntegration runs the room create integration test suite
func TestRoomCreateIntegration(t *testing.T) {
	suite.Run(t, new(RoomCreateIntegrationTestSuite))
}
