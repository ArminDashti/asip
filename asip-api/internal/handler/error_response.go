package handler

import (
	"errors"
	"net/http"

	"github.com/ArminDashti/as-ip/server/internal/dto"
	"github.com/ArminDashti/as-ip/server/internal/service"
	"github.com/gin-gonic/gin"
)

func respondWithError(c *gin.Context, err error) {
	if c.Request.Context().Err() != nil {
		return
	}
	switch {
	case errors.Is(err, service.ErrBadRequest):
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   "bad_request",
			Message: err.Error(),
		})
	case errors.Is(err, service.ErrNotFound):
		c.JSON(http.StatusNotFound, dto.ErrorResponse{
			Error:   "not_found",
			Message: err.Error(),
		})
	default:
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error:   "internal_error",
			Message: "an unexpected error occurred",
		})
	}
}
