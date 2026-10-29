package api

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/lanthora/cacao/logger"
	"github.com/lanthora/cacao/model"
	"gorm.io/gorm"
)

// databaseError separates absence/business conflicts from storage failures.
func databaseError(c *gin.Context, err error, missingCode int) bool {
	if err == nil {
		return false
	}
	switch {
	case errors.Is(err, model.ErrInvalidDeviceIP):
		setErrorCode(c, InvalidIPAddress)
	case errors.Is(err, model.ErrDeviceIPConflict):
		setErrorCode(c, DeviceIPConflict)
	case errors.Is(err, gorm.ErrRecordNotFound):
		setErrorCode(c, missingCode)
	case errors.Is(err, model.ErrUsernameTaken):
		setErrorCode(c, UsernameAlreadyTaken)
	case errors.Is(err, model.ErrNetworkExists):
		setErrorCode(c, NetworkAlreadyExists)
	case errors.Is(err, model.ErrRegistrationDisabled):
		setErrorCode(c, RegistrationDisabled)
	case errors.Is(err, model.ErrRegisterTooOften):
		setErrorCode(c, RegisterTooOften)
	case errors.Is(err, model.ErrCredentialsChanged):
		setErrorCode(c, NotLoggedIn)
	default:
		logger.Warn("database operation %s failed: %v", c.Request.URL.Path, err)
		setErrorCode(c, Unexpected)
	}
	return true
}
func config(c *gin.Context, key, fallback string) (string, bool) {
	value, err := model.GetConfig(key, fallback, c.Request.Context())
	return value, !databaseError(c, err, Unexpected)
}
func saveConfig(c *gin.Context, key, value string) {
	if !databaseError(c, model.SetConfig(key, value, c.Request.Context()), Unexpected) {
		setResponseData(c, nil)
	}
}

// DatabaseContextMiddleware bounds time waiting for the shared connection while
// preserving the GeoDB endpoint's existing five-minute download budget.
func DatabaseContextMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		if !strings.HasPrefix(c.Request.URL.Path, "/api/") {
			c.Next()
			return
		}
		timeout := 10 * time.Second
		if c.Request.URL.Path == "/api/admin/updateGeoDB" {
			timeout = 5 * time.Minute
		}
		ctx, cancel := context.WithTimeout(c.Request.Context(), timeout)
		defer cancel()
		c.Request = c.Request.WithContext(ctx)
		c.Next()
	}
}
