package service

import (
	"context"

	"github.com/wwnj/happy-billing/internal/models"
	"github.com/wwnj/happy-billing/internal/repository"
	"github.com/wwnj/happy-billing/pkg/errors"
	"gorm.io/gorm"
)

// BillService 账单服务接口
type BillService interface {
	// 账单查询
	GetBill(ctx context.Context, billID string) (*models.Bill, error)
	ListBills(ctx context.Context, req *models.BillListQueryRequest) (*models.PageResult, error)
	GetBillsByOrderID(ctx context.Context, orderID string) ([]models.Bill, error)
}

type billService struct {
	billRepo repository.BillRepository
}

// NewBillService 创建账单服务实例
func NewBillService(billRepo repository.BillRepository) BillService {
	return &billService{
		billRepo: billRepo,
	}
}

// ============================================================================
// 账单查询
// ============================================================================

func (s *billService) GetBill(ctx context.Context, billID string) (*models.Bill, error) {
	bill, err := s.billRepo.GetByID(ctx, billID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.NewWithCode(errors.ErrBillNotFound)
		}
		return nil, errors.NewInternalError("查询账单失败: " + err.Error())
	}
	return bill, nil
}

func (s *billService) ListBills(ctx context.Context, req *models.BillListQueryRequest) (*models.PageResult, error) {
	bills, total, err := s.billRepo.List(ctx, req)
	if err != nil {
		return nil, errors.NewInternalError("查询账单列表失败: " + err.Error())
	}

	return &models.PageResult{
		Total:    total,
		Page:     req.Page,
		PageSize: req.PageSize,
		Data:     bills,
	}, nil
}

func (s *billService) GetBillsByOrderID(ctx context.Context, orderID string) ([]models.Bill, error) {
	bills, err := s.billRepo.GetByOrderID(ctx, orderID)
	if err != nil {
		return nil, errors.NewInternalError("查询订单账单失败: " + err.Error())
	}
	return bills, nil
}
