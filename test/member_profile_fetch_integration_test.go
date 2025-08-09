package test

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"

	memberdomain "pray-together/internal/domains/member/domain"
)

// MemberProfileFetchIntegrationTestSuite tests member profile fetch API (matching Java MemberProfileFetchIntegrateTest)
type MemberProfileFetchIntegrationTestSuite struct {
	IntegrationTestSuite

	headers map[string]string
	member  *memberdomain.Member
}

// SetupTest runs before each test (matching Java @BeforeEach setup)
func (suite *MemberProfileFetchIntegrationTestSuite) SetupTest() {
	// Call parent SetupTest
	suite.IntegrationTestSuite.SetupTest()

	// 회원 생성
	suite.member = suite.testUtils.CreateUniqueMember()
	// memberRepository.save(member) - already saved in CreateUniqueMember

	// 인증 헤더 생성
	suite.headers = suite.testUtils.CreateAuthHeaderWithMember(suite.member)

	// Set up router with member routes
}

// TearDownTest runs after each test (matching Java @AfterEach cleanup)
func (suite *MemberProfileFetchIntegrationTestSuite) TearDownTest() {
	suite.CleanRepository()
}

// TestFetchMemberProfile tests member profile fetch (matching Java fetch_member_profile_then_return_200_ok_with_profile)
func (suite *MemberProfileFetchIntegrationTestSuite) TestFetchMemberProfile() {
	// Given
	profileURL := suite.MembersAPIURL + "/me"

	// When
	w := suite.GetRequest(profileURL, suite.headers)

	// Then
	suite.AssertStatusCode(w, http.StatusOK, "회원 프로필 조회 API 응답 상태 코드가 200 OK가 아닙니다.")

	var response MemberProfileResponse
	err := suite.UnmarshalResponse(w, &response)
	assert.NoError(suite.T(), err, "응답을 파싱할 수 있어야 합니다")
	assert.NotNil(suite.T(), response, "회원 프로필 조회 API 응답 결과가 NULL 입니다.")

	assert.Equal(suite.T(), suite.member.ID, response.ID, "응답된 회원 ID가 요청한 회원의 ID와 일치하지 않습니다.")
	assert.Equal(suite.T(), suite.member.Name, response.Name, "응답된 회원 이름이 요청한 회원의 이름과 일치하지 않습니다.")
	assert.Equal(suite.T(), suite.member.Email, response.Email, "응답된 회원 이메일이 요청한 회원의 이메일과 일치하지 않습니다.")
}

// MemberProfileResponse represents the member profile response (matching Java MemberProfileResponse)
type MemberProfileResponse struct {
	ID    uint64 `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
}

// TestMemberProfileFetchIntegration runs the member profile fetch integration test suite
func TestMemberProfileFetchIntegration(t *testing.T) {
	suite.Run(t, new(MemberProfileFetchIntegrationTestSuite))
}
