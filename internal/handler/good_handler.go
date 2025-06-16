package handler

import (
	"go-test/internal/customErr"
	"go-test/internal/dto"
	"go-test/internal/model"
	"go-test/internal/service"
	"go-test/internal/utils"
	"net/http"

	"github.com/gin-gonic/gin"
)

type GoodHandler struct {
	service service.GoodService
}

func NewGoodHandler(s service.GoodService) *GoodHandler {
	return &GoodHandler{service: s}
}

func (h *GoodHandler) Router(r *gin.Engine) {
	r.POST("/good/create", h.Create)
	r.GET("/good/:id", h.GetByID)
	r.PATCH("/good/update/:id", h.Update)
	r.DELETE("/good/remove/:id", h.Delete)
	r.GET("/goods/list", h.List)
	r.PATCH("/goods/:id/reprioritize", h.Reprioritize)
}

func (h *GoodHandler) Create(c *gin.Context) {
	var input dto.CreateGoodInput

	projectID, err := utils.GetProjectID(c)
	if err != nil {
		utils.ResponseError(c, http.StatusBadRequest, err)
		return
	}

	ctx := c.Request.Context()

	if err := c.ShouldBindJSON(&input); err != nil {
		utils.ResponseError(c, http.StatusBadRequest, err)
		return
	}

	good := model.Good{
		ProjectID: projectID,
		Name:      input.Name,
	}

	if err := h.service.Create(ctx, &good); err != nil {
		utils.ResponseError(c, http.StatusBadRequest, err)
		return
	}

	c.JSON(http.StatusCreated, good)
}

func (h *GoodHandler) GetByID(c *gin.Context) {
	id, err := utils.GetID(c)
	if err != nil {
		utils.ResponseError(c, http.StatusBadRequest, err)
		return
	}

	ctx := c.Request.Context()

	g, err := h.service.GetByID(ctx, id)
	if err != nil {
		utils.ResponseError(c, http.StatusBadRequest, err)
		return
	}

	c.JSON(http.StatusOK, g)
}

func (h *GoodHandler) Update(c *gin.Context) {
	var input dto.UpdateGoodInput

	id, err := utils.GetID(c)
	if err != nil {
		utils.ResponseError(c, http.StatusBadRequest, err)
		return
	}

	projectID, err := utils.GetProjectID(c)
	if err != nil {
		utils.ResponseError(c, http.StatusBadRequest, err)
		return
	}

	ctx := c.Request.Context()

	if err := c.ShouldBindJSON(&input); err != nil {
		utils.ResponseError(c, http.StatusBadRequest, err)
		return
	}

	existing, err := h.service.GetByID(ctx, id)
	if err != nil {
		customErr.ResponseWithError(c, http.StatusNotFound, customErr.ErrNotFound)
		return
	}

	good := model.Good{
		ID:          id,
		ProjectID:   projectID,
		Name:        input.Name,
		Description: input.Description,
		Priority:    existing.Priority,
		Removed:     existing.Removed,
		CreatedAt:   existing.CreatedAt,
	}

	if err := h.service.Update(ctx, &good); err != nil {
		customErr.ResponseWithError(c, http.StatusNotFound, customErr.ErrNotFound)
		return
	}

	c.JSON(http.StatusOK, good)
}

func (h *GoodHandler) Delete(c *gin.Context) {
	id, err := utils.GetID(c)
	if err != nil {
		utils.ResponseError(c, http.StatusBadRequest, err)
		return
	}

	projectID, err := utils.GetProjectID(c)
	if err != nil {
		utils.ResponseError(c, http.StatusBadRequest, err)
		return
	}

	ctx := c.Request.Context()

	g, err := h.service.Delete(ctx, id, projectID)
	if err != nil {
		customErr.ResponseWithError(c, http.StatusNotFound, customErr.ErrNotFound)
		return
	}

	c.JSON(http.StatusOK, g)
}

func (h *GoodHandler) List(c *gin.Context) {
	projectID, err := utils.GetProjectID(c)
	if err != nil {
		utils.ResponseError(c, http.StatusBadRequest, err)
		return
	}

	sort, err := utils.GetSort(c)
	if err != nil {
		utils.ResponseError(c, http.StatusBadRequest, err)
		return
	}

	limit, err := utils.GetLimit(c)
	if err != nil {
		utils.ResponseError(c, http.StatusBadRequest, err)
		return
	}

	offset, err := utils.GetOffset(c)
	if err != nil {
		utils.ResponseError(c, http.StatusBadRequest, err)
		return
	}

	ctx := c.Request.Context()

	goods, totalCount, removedCount, err := h.service.List(ctx, projectID, limit, offset, sort)
	if err != nil {
		utils.ResponseError(c, http.StatusBadRequest, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"goods":   goods,
		"total":   totalCount,
		"removed": removedCount,
	})
}

type ReprioritizeInput struct {
	NewPriority int `json:"newPriority"`
}

func (h *GoodHandler) Reprioritize(c *gin.Context) {
	id, err := utils.GetID(c)
	if err != nil {
		utils.ResponseError(c, http.StatusBadRequest, err)
		return
	}

	projectID, err := utils.GetProjectID(c)
	if err != nil {
		utils.ResponseError(c, http.StatusBadRequest, err)
		return
	}

	var input ReprioritizeInput
	if err := c.ShouldBindJSON(&input); err != nil {
		utils.ResponseError(c, http.StatusBadRequest, err)
		return
	}

	ctx := c.Request.Context()
	goods, err := h.service.Reprioritize(ctx, id, projectID, input.NewPriority)
	if err != nil {
		utils.ResponseError(c, http.StatusBadRequest, err)
		return
	}

	type priorityResp struct {
		ID       int `json:"id"`
		Priority int `json:"priority"`
	}
	var priorities []priorityResp
	for _, g := range goods {
		priorities = append(priorities, priorityResp{ID: g.ID, Priority: g.Priority})
	}

	c.JSON(http.StatusOK, gin.H{"priorities": priorities})
}
