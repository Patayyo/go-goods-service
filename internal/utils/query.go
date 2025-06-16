package utils

import (
	"errors"
	"strconv"

	"github.com/gin-gonic/gin"
)

var (
	ErrInvalidID        = errors.New("invalid id")
	ErrInvalidProjectID = errors.New("invalid project id")
	ErrInvalidPriority  = errors.New("invalid priority")
	ErrInvalidSort      = errors.New("invalid sort")
	ErrInvalidLimit     = errors.New("invalid limit")
	ErrInvalidOffset    = errors.New("invalid offset")
)

func GetID(c *gin.Context) (int, error) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil || id <= 0 {
		return 0, ErrInvalidID
	}

	return id, nil
}

func GetProjectID(c *gin.Context) (int, error) {
	projectStr := c.Query("project_id")
	id, err := strconv.Atoi(projectStr)
	if err != nil || id <= 0 {
		return 0, ErrInvalidProjectID
	}

	return id, nil
}

func GetSort(c *gin.Context) (string, error) {
	sort := c.DefaultQuery("sort", "asc")
	if sort != "asc" && sort != "desc" {
		return "", ErrInvalidSort
	}

	return sort, nil
}

func GetLimit(c *gin.Context) (int, error) {
	limitStr := c.DefaultQuery("limit", "20")
	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 {
		return 0, ErrInvalidLimit
	}

	return limit, nil
}

func GetOffset(c *gin.Context) (int, error) {
	offsetStr := c.DefaultQuery("offset", "0")
	offset, err := strconv.Atoi(offsetStr)
	if err != nil || offset < 0 {
		return 0, ErrInvalidOffset
	}

	return offset, nil
}
