package service

import (
	"blackapp/internal/service/dto"

	"github.com/gin-gonic/gin"
)

type BlacklistService interface {
	Create(ctx *gin.Context, req *dto.CreateBlacklistDTO) error
	Update(ctx *gin.Context, req *dto.UpdateBlacklistDTO) error
	Delete(ctx *gin.Context, id int) error
	GetByID(ctx *gin.Context, id int) (*dto.BlacklistDTO, error)
	List(ctx *gin.Context, page, size int) ([]*dto.BlacklistDTO, int64, error)
	UpdateStatus(ctx *gin.Context, id int, status int) error
	Check(ctx *gin.Context, req *dto.CheckBlacklistDTO) (bool, error)
}
