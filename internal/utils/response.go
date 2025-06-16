package utils

import "github.com/gin-gonic/gin"

func ResponseError(c *gin.Context, status int, err error) {
	c.JSON(status, gin.H{"error": err.Error()})
}
