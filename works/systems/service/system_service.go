package service

import (
	"github.com/muskong/gopermission/works/systems/service/dto"

	"github.com/gin-gonic/gin"
)

type SystemService interface {
	// 系统监控
	GetSystemMetrics(ctx *gin.Context) (*dto.SystemMetrics, error)
}
