package auth

import (
	sharedError "github.com/changhyeonkim/pray-together/go-api-server/internal/shared/error"
	sharedHttp "github.com/changhyeonkim/pray-together/go-api-server/internal/shared/http"
	"github.com/gin-gonic/gin"
)

type AuthHandler struct {
	authService *AuthService
}

func NewAuthHandler(authService *AuthService) *AuthHandler {
	return &AuthHandler{
		authService: authService,
	}
}

func (a *AuthHandler) Login(c *gin.Context) {
	var request LoginRequest

	// Parse and validate JSON request
	if !sharedHttp.BindJSON(c, &request) {
		return
	}

	response, err := a.authService.Login(c.Request.Context(), &request)
	if err != nil {
		if resp, ok := sharedError.ResolveDomainError(err); ok {
			sharedHttp.RespondError(c, err, resp)
			return
		}

		sharedHttp.RespondError(c, err, sharedError.InternalServerError)
		return
	}

	c.JSON(200, response)
}

func (a *AuthHandler) Signup(c *gin.Context) {
	var request SignupRequest

	// Parse and validate JSON request
	if !sharedHttp.BindJSON(c, &request) {
		return
	}

	err := a.authService.Signup(c.Request.Context(), &request)
	if err != nil {
		if resp, ok := sharedError.ResolveDomainError(err); ok {
			sharedHttp.RespondError(c, err, resp)
			return
		}

		sharedHttp.RespondError(c, err, sharedError.InternalServerError)
		return
	}
	c.JSON(201, gin.H{})
}
