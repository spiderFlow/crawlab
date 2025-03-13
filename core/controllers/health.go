package controllers

import (
	"errors"
	"github.com/gin-gonic/gin"
	"net/http"
)

func GetHealthFn(healthFn func() bool) func(c *gin.Context) error {
	return func(c *gin.Context) (err error) {
		if healthFn() {
			c.Writer.Write([]byte("ok"))
			c.AbortWithStatus(http.StatusOK)
			return
		}
		c.Writer.Write([]byte("not ready"))
		c.AbortWithStatus(http.StatusServiceUnavailable)
		return errors.New("not ready")
	}
}
