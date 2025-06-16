package customErr

import "github.com/gin-gonic/gin"

type AppError struct {
	Code    int                    `json:"code"`
	Message string                 `json:"message"`
	Details map[string]interface{} `json:"details"`
}

var ErrNotFound = &AppError{
	Code:    3,
	Message: "errors.common.notFound",
	Details: map[string]interface{}{},
}

func (e *AppError) Error() string {
	return e.Message
}

func ResponseWithError(c *gin.Context, code int, e *AppError) {
	c.JSON(code, e)
}
