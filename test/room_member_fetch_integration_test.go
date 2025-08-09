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

// RoomMemberFetchIntegrationTestSuite tests room member fetch API (matching Java RoomMemberFetchIntegrateTest)
type RoomMemberFetchIntegrationTestSuite struct {
	IntegrationTestSuite

	member      *memberdomain.Member
	headers     map[string]string
	membersURL  string
	room        *roomdomain.Room
	memberCount int
}

// SetupTest runs before each test (matching Java @BeforeEach setup)
func (suite *RoomMemberFetchIntegrationTestSuite) SetupTest() {
	// Call parent SetupTest
	suite.IntegrationTestSuite.SetupTest()

	suite.membersURL = "/members"
	suite.memberCount = 10

	// 회원 생성 및 JWT 설정
	suite.member = suite.testUtils.CreateUniqueMember()
	// memberRepository.save(member) - already saved in CreateUniqueMember
	suite.headers = suite.testUtils.CreateAuthHeaderWithMember(suite.member)

	// 방 생성
	createRequest := RoomCreateRequest{
		Name:        "test-name",
		Description: "test-description",
	}
	w := suite.PostJSON(suite.RoomsAPIURL, createRequest, suite.headers)
	assert.Equal(suite.T(), http.StatusCreated, w.Code)

	// 방 정보 획득
	var allRoom []roomdomain.Room
	suite.db.Find(&allRoom)
	suite.room = &allRoom[0]

	// 방 참가자 생성 (본인 포함 총 memberCount 명)
	memberRoomList := []roomdomain.RoomMember{}
	memberList := []*memberdomain.Member{}

	for i := 0; i < suite.memberCount-1; i++ {
		newMember := suite.testUtils.CreateUniqueMember()
		memberList = append(memberList, newMember)

		memberRoom := roomdomain.RoomMember{
			MemberID:       newMember.ID,
			RoomID:         suite.room.ID,
			Role:           roomdomain.RoleMember,
			IsNotification: true,
		}
		memberRoomList = append(memberRoomList, memberRoom)
	}

	// Save all member rooms
	for _, mr := range memberRoomList {
		suite.db.Create(&mr)
	}

	// Set up router with room routes
}

// TearDownTest runs after each test (matching Java @AfterEach cleanup)
func (suite *RoomMemberFetchIntegrationTestSuite) TearDownTest() {
	suite.CleanRepository()
}

// TestFetchRoomMembersThenReturn200OK tests fetching room members (matching Java fetch_room_members_then_return_200_ok)
func (suite *RoomMemberFetchIntegrationTestSuite) TestFetchRoomMembersThenReturn200OK() {
	// When
	// API 요청
	url := fmt.Sprintf("%s/%d%s", suite.RoomsAPIURL, suite.room.ID, suite.membersURL)
	w := suite.GetRequest(url, suite.headers)

	// Then
	// 응답 상태 코드 검증
	suite.AssertStatusCode(w, http.StatusOK, "방 참가자 조회 API 응답 상태 코드가 200 OK가 아닙니다.")

	var response RoomMemberResponse
	err := suite.UnmarshalResponse(w, &response)
	assert.NoError(suite.T(), err, "응답을 파싱할 수 있어야 합니다")

	// 응답 결과 검증
	members := response.Members
	assert.NotNil(suite.T(), members, "방 참가자 조회 API 응답 결과가 NULL 입니다.")
	assert.Equal(suite.T(), suite.memberCount, len(members),
		"방 참가자 조회 API 응답 결과, 방 참가자 수가 예상과 다릅니다.")

	// Check if owner member is included
	ownerMember := MemberIdName{
		ID:   suite.member.ID,
		Name: suite.member.Name,
	}

	found := false
	for _, m := range members {
		if m.ID == ownerMember.ID && m.Name == ownerMember.Name {
			found = true
			break
		}
	}
	assert.True(suite.T(), found, "방을 생성한 Member가 요청 방에 포함되지 않고 있습니다.")
}

// MemberIdName represents member ID and name (matching Java MemberIdName)
type MemberIdName struct {
	ID   uint64 `json:"id"`
	Name string `json:"name"`
}

// RoomMemberResponse represents room member response (matching Java RoomMemberResponse)
type RoomMemberResponse struct {
	Members []MemberIdName `json:"members"`
}

// TestRoomMemberFetchIntegration runs the room member fetch integration test suite
func TestRoomMemberFetchIntegration(t *testing.T) {
	suite.Run(t, new(RoomMemberFetchIntegrationTestSuite))
}
