package member_test

import (
	"net/http"
	"testing"

	testutil2 "github.com/changhyeonkim/pray-together/go-api-server/internal/app/shared/testutil"
	"github.com/changhyeonkim/pray-together/go-api-server/internal/member"
	"github.com/changhyeonkim/pray-together/go-api-server/internal/model"
	"github.com/stretchr/testify/assert"
)

func TestSearchMembers_ReturnsMatches(t *testing.T) {
	memberHandler, db := setupMemberTestEnvironment(t)

	m1 := testutil2.CreateTestMemberWithIndex(t, db, 0)
	m2 := testutil2.CreateTestMemberWithIndex(t, db, 1)
	m3 := testutil2.CreateTestMemberWithIndex(t, db, 2)

	// Customize member data for search scenario
	db.Model(&model.Member{}).Where("id = ?", m1.ID).Updates(map[string]any{
		"name":         "홍길동",
		"phone_number": "010-1234-5678",
	})
	db.Model(&model.Member{}).Where("id = ?", m2.ID).Updates(map[string]any{
		"name":         "홍길순",
		"phone_number": "",
	})
	db.Model(&model.Member{}).Where("id = ?", m3.ID).Updates(map[string]any{
		"name":         "김철수",
		"phone_number": "010-0000-0000",
	})

	router := testutil2.SetupAuthenticatedRouter(m1.ID)
	router.GET("/api/v1/members/search", memberHandler.SearchMembers)

	recorder := testutil2.ExecuteRequest(t, router, testutil2.TestRequest{
		Method: http.MethodGet,
		URL:    "/api/v1/members/search?name=홍",
	})

	assert.Equal(t, http.StatusOK, recorder.Code)

	var response member.SearchMemberResponse
	testutil2.ParseResponse(t, recorder, &response)
	assert.Len(t, response.Members, 2)

	resultMap := map[string]*string{}
	for _, m := range response.Members {
		resultMap[m.Name] = m.PhoneNumberSuffix
	}

	if suffix, ok := resultMap["홍길동"]; assert.True(t, ok) {
		if assert.NotNil(t, suffix) {
			assert.Equal(t, "5678", *suffix)
		}
	}

	if suffix, ok := resultMap["홍길순"]; assert.True(t, ok) {
		assert.Nil(t, suffix)
	}
}

func TestSearchMembers_Unauthorized(t *testing.T) {
	memberHandler, _ := setupMemberTestEnvironment(t)
	router := testutil2.SetupTestRouter()
	router.GET("/api/v1/members/search", memberHandler.SearchMembers)

	recorder := testutil2.ExecuteRequest(t, router, testutil2.TestRequest{
		Method: http.MethodGet,
		URL:    "/api/v1/members/search?name=홍",
	})

	assert.Equal(t, http.StatusUnauthorized, recorder.Code)
}
