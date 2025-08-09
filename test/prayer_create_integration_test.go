package test

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"

	memberdomain "pray-together/internal/domains/member/domain"
	roomdomain "pray-together/internal/domains/room/domain"
)

// PrayerCreateIntegrationTestSuite tests prayer creation API (matching Java PrayerCreateIntegrateTest)
type PrayerCreateIntegrationTestSuite struct {
	IntegrationTestSuite

	member     *memberdomain.Member
	room       *roomdomain.Room
	memberRoom *roomdomain.RoomMember
	headers    map[string]string

	validRoomID   uint64
	validMemberID uint64
}

// SetupTest runs before each test (matching Java @BeforeEach setup)
func (suite *PrayerCreateIntegrationTestSuite) SetupTest() {
	// Call parent SetupTest
	suite.IntegrationTestSuite.SetupTest()

	// Create member (matching Java)
	suite.member = suite.testUtils.CreateUniqueMember()
	suite.validMemberID = suite.member.ID

	// Create room (matching Java)
	suite.room = suite.testUtils.CreateUniqueRoom()
	suite.validRoomID = suite.room.ID

	// Create member room relationship (matching Java)
	suite.memberRoom = suite.testUtils.CreateUniqueMemberRoomWithMemberAndRoom(suite.member, suite.room)

	// Create auth headers (matching Java)
	suite.headers = suite.testUtils.CreateAuthHeaderWithMember(suite.member)

	// Set up router with prayer routes
}

// TestCreatePrayerWithValidInput tests prayer creation with valid input (matching Java create_prayer_with_valid_input_then_return_201_created)
func (suite *PrayerCreateIntegrationTestSuite) TestCreatePrayerWithValidInput() {
	// Given - Create multiple members for prayer contents
	memberList := []*memberdomain.Member{suite.member} // Start with existing member

	const testCount = 5
	// Create additional members (matching Java logic)
	for i := 2; i < testCount; i++ {
		memberList = append(memberList, suite.testUtils.CreateUniqueMember())
	}

	// Create prayer request contents (matching Java)
	var requestContents []PrayerRequestContent
	for _, memberOne := range memberList {
		content := PrayerRequestContent{
			MemberID:   &memberOne.ID, // Pointer to allow nil
			MemberName: memberOne.Name,
			Content:    fmt.Sprintf("test-prayer-content%d", memberOne.ID),
		}
		requestContents = append(requestContents, content)
	}

	// Add content with memberID == nil (matching Java)
	requestContents = append(requestContents, PrayerRequestContent{
		MemberID:   nil, // nil memberID
		MemberName: "test-memberName-id-null",
		Content:    "test-content-id-null",
	})

	// Create DTO (matching Java)
	requestDTO := PrayerCreateRequest{
		Title:    "test-prayer-title",
		RoomID:   suite.room.ID,
		Contents: requestContents,
	}

	// When - Make POST request
	w := suite.PostJSON(suite.PrayersAPIURL, requestDTO, suite.headers)

	// Then - Assert response
	suite.AssertStatusCode(w, http.StatusCreated, "기도 생성 API 응답 상태 코드가 201 Created가 아닙니다.")

	// Assert database state (matching Java)
	allTitles := suite.testUtils.FindAllPrayerTitles()
	assert.Equal(suite.T(), 1, len(allTitles), "저장된 기도 제목의 개수가 예상과 다릅니다.")

	allContents := suite.testUtils.FindAllPrayerContents()
	assert.Equal(suite.T(), testCount, len(allContents), "저장된 기도 내용의 개수가 예상과 다릅니다.")
}

// TestCreatePrayerTitleOnly tests prayer title only creation (matching Java create_prayer_title_only_then_return_201_created)
func (suite *PrayerCreateIntegrationTestSuite) TestCreatePrayerTitleOnly() {
	// Given - Create request with empty contents
	requestDTO := PrayerCreateRequest{
		Title:    "test-prayer-title-only",
		RoomID:   suite.room.ID,
		Contents: []PrayerRequestContent{}, // Empty contents
	}

	// When - Make POST request
	w := suite.PostJSON(suite.PrayersAPIURL, requestDTO, suite.headers)

	// Then - Assert response
	suite.AssertStatusCode(w, http.StatusCreated, "기도 제목만 생성 API 응답 상태 코드가 201 Created가 아닙니다.")

	// Assert database state
	allTitles := suite.testUtils.FindAllPrayerTitles()
	assert.Equal(suite.T(), 1, len(allTitles), "저장된 기도 제목의 개수가 예상과 다릅니다.")
	assert.Equal(suite.T(), suite.room.ID, allTitles[0].RoomID, "기도 제목의 룸 ID가 예상과 다릅니다.")

	allContents := suite.testUtils.FindAllPrayerContents()
	assert.Equal(suite.T(), 0, len(allContents), "기도 제목만 생성 시 기도 내용은 저장되지 않아야 합니다.")
}

// TestCreatePrayerWithInvalidInput tests prayer creation with invalid input (matching Java @ParameterizedTest)
func (suite *PrayerCreateIntegrationTestSuite) TestCreatePrayerWithInvalidInput() {
	// Test cases matching Java provideInvalidPrayerCreateArguments
	testCases := []struct {
		name        string
		roomID      *uint64
		title       *string
		contents    []PrayerRequestContent
		expectedMsg string
	}{
		{
			name:        "roomId가 0일 때",
			roomID:      uintPtr(0),
			title:       stringPtr("valid-title"),
			contents:    []PrayerRequestContent{},
			expectedMsg: "roomId가 0일 때: 응답 상태 코드가 400 Bad Request가 아닙니다.",
		},
		{
			name:        "roomId가 null일 때",
			roomID:      nil,
			title:       stringPtr("valid-title"),
			contents:    []PrayerRequestContent{},
			expectedMsg: "roomId가 null일 때: 응답 상태 코드가 400 Bad Request가 아닙니다.",
		},
		{
			name:        "title이 empty일 때",
			roomID:      &suite.validRoomID,
			title:       stringPtr(""),
			contents:    []PrayerRequestContent{},
			expectedMsg: "title이 empty일 때: 응답 상태 코드가 400 Bad Request가 아닙니다.",
		},
		{
			name:        "title이 null일 때",
			roomID:      &suite.validRoomID,
			title:       nil,
			contents:    []PrayerRequestContent{},
			expectedMsg: "title이 null일 때: 응답 상태 코드가 400 Bad Request가 아닙니다.",
		},
		{
			name:        "title이 50자 초과일 때",
			roomID:      &suite.validRoomID,
			title:       stringPtr(generateString(51)), // 51 characters
			contents:    []PrayerRequestContent{},
			expectedMsg: "title이 50자 초과일 때: 응답 상태 코드가 400 Bad Request가 아닙니다.",
		},
		{
			name:   "contents의 content가 empty일 때",
			roomID: &suite.validRoomID,
			title:  stringPtr("valid-title"),
			contents: []PrayerRequestContent{
				{
					MemberID:   &suite.validMemberID,
					MemberName: "valid-name",
					Content:    "", // Empty content
				},
			},
			expectedMsg: "contents의 content가 empty일 때: 응답 상태 코드가 400 Bad Request가 아닙니다.",
		},
		{
			name:   "contents의 memberName이 empty일 때",
			roomID: &suite.validRoomID,
			title:  stringPtr("valid-title"),
			contents: []PrayerRequestContent{
				{
					MemberID:   &suite.validMemberID,
					MemberName: "", // Empty member name
					Content:    "valid-content",
				},
			},
			expectedMsg: "contents의 memberName이 empty일 때: 응답 상태 코드가 400 Bad Request가 아닙니다.",
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			// Given
			requestDTO := PrayerCreateRequest{
				Contents: tc.contents,
			}

			if tc.roomID != nil {
				requestDTO.RoomID = *tc.roomID
			}
			if tc.title != nil {
				requestDTO.Title = *tc.title
			}

			// When
			w := suite.PostJSON(suite.PrayersAPIURL, requestDTO, suite.headers)

			// Then
			suite.AssertStatusCode(w, http.StatusBadRequest, tc.expectedMsg)

			// Assert no data was saved
			allTitles := suite.testUtils.FindAllPrayerTitles()
			assert.Equal(suite.T(), 0, len(allTitles), tc.name+": 예외 발생 시 기도 제목이 저장되면 안됩니다.")

			allContents := suite.testUtils.FindAllPrayerContents()
			assert.Equal(suite.T(), 0, len(allContents), tc.name+": 예외 발생 시 기도 내용이 저장되면 안됩니다.")
		})
	}
}

// Helper types for request/response matching Java DTOs

// PrayerCreateRequest represents the request to create a prayer (matching Java PrayerCreateRequest)
type PrayerCreateRequest struct {
	Title    string                 `json:"title" binding:"required,min=1,max=50"`
	RoomID   uint64                 `json:"roomId" binding:"required"`
	Contents []PrayerRequestContent `json:"contents"`
}

// PrayerRequestContent represents prayer content in request (matching Java PrayerRequestContent)
type PrayerRequestContent struct {
	MemberID   *uint64 `json:"memberId,omitempty"`
	MemberName string  `json:"memberName" binding:"required"`
	Content    string  `json:"content" binding:"required"`
}

// Helper functions

func stringPtr(s string) *string {
	return &s
}

func uintPtr(u uint64) *uint64 {
	return &u
}

func generateString(length int) string {
	result := make([]byte, length)
	for i := 0; i < length; i++ {
		result[i] = 'a'
	}
	return string(result)
}

// TestPrayerCreateIntegration runs the prayer create integration test suite
func TestPrayerCreateIntegration(t *testing.T) {
	suite.Run(t, new(PrayerCreateIntegrationTestSuite))
}
