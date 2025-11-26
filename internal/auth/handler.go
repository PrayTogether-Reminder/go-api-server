package auth

import (
	"net/http"

	sharedError "github.com/changhyeonkim/pray-together/go-api-server/internal/shared/error"
	sharedHttp "github.com/changhyeonkim/pray-together/go-api-server/internal/shared/http"
	"github.com/gin-gonic/gin"
)

type AuthHandler struct {
	authUseCase *AuthUseCase
}

func NewAuthHandler(authUseCase *AuthUseCase) *AuthHandler {
	return &AuthHandler{
		authUseCase: authUseCase,
	}
}

func (a *AuthHandler) Login(c *gin.Context) {
	var request LoginRequest

	// Parse and validate JSON request
	if !sharedHttp.BindJSON(c, &request) {
		return
	}

	response, err := a.authUseCase.Login(c.Request.Context(), &request)
	if err != nil {
		if resp, ok := sharedError.ResolveDomainError(err); ok {
			sharedHttp.RespondError(c, err, resp)
			return
		}

		sharedHttp.RespondError(c, err, sharedError.InternalServerError)
		return
	}

	c.JSON(http.StatusOK, response)
}

func (a *AuthHandler) Signup(c *gin.Context) {
	var request SignupRequest

	// Parse and validate JSON request
	if !sharedHttp.BindJSON(c, &request) {
		return
	}

	err := a.authUseCase.Signup(c.Request.Context(), &request)
	if err != nil {
		if resp, ok := sharedError.ResolveDomainError(err); ok {
			sharedHttp.RespondError(c, err, resp)
			return
		}

		sharedHttp.RespondError(c, err, sharedError.InternalServerError)
		return
	}
	c.JSON(http.StatusCreated, gin.H{})
}

func (a *AuthHandler) RequestEmailOTP(c *gin.Context) {
	var request EmailOtpRequest

	if !sharedHttp.BindJSON(c, &request) {
		return
	}

	response, err := a.authUseCase.RequestEmailOTP(c.Request.Context(), &request)
	if err != nil {
		if resp, ok := sharedError.ResolveDomainError(err); ok {
			sharedHttp.RespondError(c, err, resp)
			return
		}

		sharedHttp.RespondError(c, err, sharedError.InternalServerError)
		return
	}

	c.JSON(http.StatusOK, response)
}
