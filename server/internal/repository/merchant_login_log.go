package repository

import (
	"blacklist/internal/model"
	"blacklist/internal/pkg/database"
	"context"
)

type MerchantLoginLogRepository interface {
	Create(ctx context.Context, log *model.MerchantLoginLog) error
	FindByMerchantID(ctx context.Context, merchantID uint, page, pageSize int) ([]model.MerchantLoginLog, int64, error)
}

type merchantLoginLogRepository struct {
	db *database.PostgresDB
}

func NewMerchantLoginLogRepository(db *database.PostgresDB) MerchantLoginLogRepository {
	return &merchantLoginLogRepository{db: db}
}

func (r *merchantLoginLogRepository) Create(ctx context.Context, log *model.MerchantLoginLog) error {
	return r.db.WithContext(ctx).Create(log).Error
}

func (r *merchantLoginLogRepository) FindByMerchantID(ctx context.Context, merchantID uint, page, pageSize int) ([]model.MerchantLoginLog, int64, error) {
	var logs []model.MerchantLoginLog
	var total int64

	offset := (page - 1) * pageSize

	err := r.db.WithContext(ctx).Model(&model.MerchantLoginLog{}).
		Where("merchant_id = ?", merchantID).
		Count(&total).
		Order("login_time DESC").
		Offset(offset).
		Limit(pageSize).
		Find(&logs).Error

	return logs, total, err
}