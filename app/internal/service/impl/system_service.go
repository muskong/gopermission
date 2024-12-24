package impl

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"runtime"
	"time"

	"github.com/golang-jwt/jwt"
	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/mem"
	"go.uber.org/zap"

	"blackapp/internal/domain/entity"
	"blackapp/internal/domain/repository"
	"blackapp/internal/service/dto"
	"blackapp/pkg/logger"
)

type systemService struct {
	adminRepo    repository.AdminRepository
	loginLogRepo repository.LoginLogRepository
	queryLogRepo repository.QueryLogRepository
	jwtSecret    string
	tokenExpire  time.Duration
}

func NewSystemService(
	adminRepo repository.AdminRepository,
	loginLogRepo repository.LoginLogRepository,
	queryLogRepo repository.QueryLogRepository,
	jwtSecret string,
	tokenExpire time.Duration,
) *systemService {
	return &systemService{
		adminRepo:    adminRepo,
		loginLogRepo: loginLogRepo,
		queryLogRepo: queryLogRepo,
		jwtSecret:    jwtSecret,
		tokenExpire:  tokenExpire,
	}
}

func (s *systemService) GetSystemMetrics(ctx context.Context) (*dto.SystemMetrics, error) {
	metrics := &dto.SystemMetrics{}

	// CPU信息
	cpuPercent, err := cpu.Percent(time.Second, false)
	if err != nil {
		return nil, err
	}
	cpuInfo, err := cpu.Info()
	if err != nil {
		return nil, err
	}
	metrics.CPU.Usage = cpuPercent[0]
	metrics.CPU.Cores = runtime.NumCPU()
	if len(cpuInfo) > 0 {
		metrics.CPU.ModelName = cpuInfo[0].ModelName
	}

	// 内存信息
	memInfo, err := mem.VirtualMemory()
	if err != nil {
		return nil, err
	}
	metrics.Memory.Total = memInfo.Total
	metrics.Memory.Used = memInfo.Used
	metrics.Memory.Free = memInfo.Free
	metrics.Memory.UsageRate = memInfo.UsedPercent

	// Redis信息
	// TODO: 实现Redis信息采集

	// PostgreSQL信息
	// TODO: 实现PostgreSQL信息采集

	return metrics, nil
}

func (s *systemService) AdminLogin(ctx context.Context, req *dto.AdminLoginDTO) (string, error) {
	admin, err := s.adminRepo.FindByUsername(ctx, req.Username)
	if err != nil {
		return "", err
	}

	if admin == nil || admin.Password != hashPassword(req.Password) {
		return "", fmt.Errorf("invalid credentials")
	}

	// 生成JWT Token
	expireTime := time.Now().Add(s.tokenExpire)
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"admin_id": admin.ID,
		"exp":      expireTime.Unix(),
	})

	tokenString, err := token.SignedString([]byte(s.jwtSecret))
	if err != nil {
		return "", err
	}

	// 更新最后登录时间
	admin.LastLogin = time.Now()
	if err := s.adminRepo.Update(ctx, admin); err != nil {
		logger.Logger.Error("更新管理员最后登录时间失败", zap.Error(err))
	}

	return tokenString, nil
}

func (s *systemService) CreateAdmin(ctx context.Context, req *dto.CreateAdminDTO) error {
	admin := &entity.Admin{
		Username:  req.Username,
		Password:  hashPassword(req.Password),
		Name:      req.Name,
		Status:    1,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	return s.adminRepo.Create(ctx, admin)
}

func (s *systemService) UpdateAdmin(ctx context.Context, req *dto.UpdateAdminDTO) error {
	admin, err := s.adminRepo.FindByID(ctx, req.ID)
	if err != nil {
		return err
	}

	admin.Username = req.Username
	if req.Password != "" {
		admin.Password = hashPassword(req.Password)
	}
	admin.Name = req.Name
	admin.Status = req.Status
	admin.UpdatedAt = time.Now()

	return s.adminRepo.Update(ctx, admin)
}

func (s *systemService) DeleteAdmin(ctx context.Context, id int) error {
	return s.adminRepo.Delete(ctx, id)
}

func (s *systemService) GetAdminByID(ctx context.Context, id int) (*dto.AdminDTO, error) {
	admin, err := s.adminRepo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	return toAdminDTO(admin), nil
}

func (s *systemService) ListAdmins(ctx context.Context, page, size int) ([]*dto.AdminDTO, int64, error) {
	admins, total, err := s.adminRepo.List(ctx, page, size)
	if err != nil {
		return nil, 0, err
	}

	dtos := make([]*dto.AdminDTO, len(admins))
	for i, admin := range admins {
		dtos[i] = toAdminDTO(admin)
	}
	return dtos, total, nil
}

func (s *systemService) UpdateAdminStatus(ctx context.Context, id int, status int) error {
	return s.adminRepo.UpdateStatus(ctx, id, status)
}

func (s *systemService) ListLoginLogs(ctx context.Context, userType int, page, size int) ([]*dto.LoginLogDTO, int64, error) {
	logs, total, err := s.loginLogRepo.List(ctx, userType, page, size)
	if err != nil {
		return nil, 0, err
	}

	dtos := make([]*dto.LoginLogDTO, len(logs))
	for i, log := range logs {
		dtos[i] = toLoginLogDTO(log)
	}
	return dtos, total, nil
}

func (s *systemService) ListQueryLogs(ctx context.Context, merchantID int, page, size int) ([]*dto.QueryLogDTO, int64, error) {
	logs, total, err := s.queryLogRepo.List(ctx, merchantID, page, size)
	if err != nil {
		return nil, 0, err
	}

	dtos := make([]*dto.QueryLogDTO, len(logs))
	for i, log := range logs {
		dtos[i] = toQueryLogDTO(log)
	}
	return dtos, total, nil
}

func toAdminDTO(admin *entity.Admin) *dto.AdminDTO {
	return &dto.AdminDTO{
		ID:        admin.ID,
		Username:  admin.Username,
		Name:      admin.Name,
		Status:    admin.Status,
		LastLogin: admin.LastLogin.Format("2006-01-02 15:04:05"),
	}
}

func toLoginLogDTO(log *entity.LoginLog) *dto.LoginLogDTO {
	return &dto.LoginLogDTO{
		ID:        log.ID,
		Type:      log.Type,
		UserID:    log.UserID,
		IP:        log.IP,
		UserAgent: log.UserAgent,
		Status:    log.Status,
		CreatedAt: log.CreatedAt,
	}
}

func toQueryLogDTO(log *entity.QueryLog) *dto.QueryLogDTO {
	return &dto.QueryLogDTO{
		ID:         log.ID,
		MerchantID: log.MerchantID,
		Phone:      log.Phone,
		IDCard:     log.IDCard,
		Name:       log.Name,
		IP:         log.IP,
		UserAgent:  log.UserAgent,
		Exists:     log.Exists,
		CreatedAt:  log.CreatedAt,
	}
}

func hashPassword(password string) string {
	hash := sha256.Sum256([]byte(password))
	return hex.EncodeToString(hash[:])
}
