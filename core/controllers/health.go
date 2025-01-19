package controllers

import (
	"github.com/gin-gonic/gin"
	"net/http"
)

func GetHealthFn(healthFn func() bool) func(c *gin.Context) {
	return func(c *gin.Context) {
		if healthFn() {
			c.Writer.Write([]byte("ok"))
			c.AbortWithStatus(http.StatusOK)
			return
		}
		c.Writer.Write([]byte("not ready"))
		c.AbortWithStatus(http.StatusServiceUnavailable)
	}
}
