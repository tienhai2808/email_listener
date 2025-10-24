package main

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

type handler struct {
	svc service
}

func newHandler(svc service) *handler {
	return &handler{svc}
}

func (h *handler) login(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	var req loginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		errMess := handleValidationError(err)
		toApiResponse(c, http.StatusBadRequest, errMess, nil)
		return
	}

	user, err := h.svc.login(ctx, req.Code)
	if err != nil {
		toApiResponse(c, http.StatusInternalServerError, err.Error(), nil)
		return
	}

	toApiResponse(c, http.StatusOK, "Đăng nhập thành công", gin.H{
		"user": user,
	})
}
