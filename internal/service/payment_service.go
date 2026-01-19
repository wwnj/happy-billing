package service

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/wwnj/happy-billing/internal/models"
	"github.com/wwnj/happy-billing/internal/repository"
	"github.com/wwnj/happy-billing/pkg/errors"
	"github.com/wwnj/happy-billing/pkg/utils"
	"gorm.io/gorm"
)

// PaymentService 支付服务接口
type PaymentService interface {
	// 支付操作
	CreatePayment(ctx context.Context, req *models.CreatePaymentRequest) (*models.Payment, error)
	GetPayment(ctx context.Context, paymentID string) (*models.Payment, error)

	// 余额操作
	GetBalance(ctx context.Context, tenantID string) (*models.AccountBalance, error)
	Recharge(ctx context.Context, tenantID string, amount float64) error
	GetBalanceTransactions(ctx context.Context, tenantID string, page, pageSize int) ([]models.BalanceTransaction, int64, error)
}

type paymentService struct {
	db               *gorm.DB
	redis            *redis.Client
	paymentRepo      repository.PaymentRepository
	billRepo         repository.BillRepository
	orderRepo        repository.OrderRepository
	balanceRepo      repository.AccountBalanceRepository
	balanceTransRepo repository.BalanceTransactionRepository
	currencyService  CurrencyService
}

// NewPaymentService 创建支付服务实例
func NewPaymentService(
	db *gorm.DB,
	redis *redis.Client,
	paymentRepo repository.PaymentRepository,
	billRepo repository.BillRepository,
	orderRepo repository.OrderRepository,
	balanceRepo repository.AccountBalanceRepository,
	balanceTransRepo repository.BalanceTransactionRepository,
	currencyService CurrencyService,
) PaymentService {
	return &paymentService{
		db:               db,
		redis:            redis,
		paymentRepo:      paymentRepo,
		billRepo:         billRepo,
		orderRepo:        orderRepo,
		balanceRepo:      balanceRepo,
		balanceTransRepo: balanceTransRepo,
		currencyService:  currencyService,
	}
}

// ============================================================================
// 支付操作
// ============================================================================

func (s *paymentService) CreatePayment(ctx context.Context, req *models.CreatePaymentRequest) (*models.Payment, error) {
	// 1. 查询账单或订单
	var billID, orderID *string
	var amount float64
	var tenantID string

	if req.BillID != nil {
		// 支付账单
		bill, err := s.billRepo.GetByID(ctx, *req.BillID)
		if err != nil {
			if err == gorm.ErrRecordNotFound {
				return nil, errors.NewWithCode(errors.ErrBillNotFound)
			}
			return nil, errors.NewInternalError("查询账单失败: " + err.Error())
		}

		if bill.Status == models.BillStatusPaid {
			return nil, errors.NewWithCode(errors.ErrBillAlreadyPaid)
		}

		billID = req.BillID
		orderID = &bill.OrderID // 从账单中获取订单ID
		amount = bill.PayableAmount
		tenantID = bill.TenantID
	} else if req.OrderID != nil {
		// 直接支付订单
		order, err := s.orderRepo.GetByID(ctx, *req.OrderID)
		if err != nil {
			if err == gorm.ErrRecordNotFound {
				return nil, errors.NewWithCode(errors.ErrOrderNotFound)
			}
			return nil, errors.NewInternalError("查询订单失败: " + err.Error())
		}

		orderID = req.OrderID
		amount = order.PayableAmount - order.PaidAmount
		tenantID = order.TenantID
	} else {
		return nil, errors.New(errors.ErrInvalidParams, "必须指定账单ID或订单ID")
	}

	// 2. 生成支付ID
	paymentID, err := utils.GeneratePaymentID(ctx, s.redis)
	if err != nil {
		return nil, errors.NewInternalError("生成支付ID失败: " + err.Error())
	}

	// 3. 计算汇率和本位币金额
	paymentCurrency := req.Currency
	if paymentCurrency == "" {
		paymentCurrency = models.CurrencyCNY // 默认人民币
	}

	baseCurrencyAmount, exchangeRate, err := s.currencyService.CalculateBaseCurrencyAmount(ctx, amount, paymentCurrency, nil)
	if err != nil {
		return nil, errors.Wrap(err, "计算本位币金额失败")
	}

	// 4. 创建支付记录
	payment := &models.Payment{
		PaymentID:          paymentID,
		OrderID:            orderID,
		BillID:             billID,
		TenantID:           tenantID,
		UserID:             req.UserID,
		PaymentMethod:      req.PaymentMethod,
		PaymentChannel:     req.PaymentChannel,
		Currency:           paymentCurrency,
		ExchangeRate:       &exchangeRate,
		BaseCurrency:       models.CurrencyCNY,
		BaseCurrencyAmount: &baseCurrencyAmount,
		Amount:             amount,
		Status:             models.PaymentStatusPending,
	}

	// 5. 根据支付方式处理
	if req.PaymentMethod == models.PaymentMethodBalance {
		// 余额支付 - 使用本位币金额扣减
		err = s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
			// 查询账户余额
			balance, err := s.balanceRepo.GetByTenantID(ctx, tenantID)
			if err != nil {
				return err
			}

			// 检查余额是否充足（余额是CNY，使用本位币金额比较）
			if balance.Balance < baseCurrencyAmount {
				return errors.NewWithCode(errors.ErrBalanceInsufficient)
			}

			// 创建支付记录
			if err := s.paymentRepo.Create(ctx, payment); err != nil {
				return err
			}

			// 扣减余额（使用本位币金额）
			if err := s.balanceRepo.UpdateBalance(ctx, tenantID, -baseCurrencyAmount, 0); err != nil {
				return err
			}

			// 记录余额变动
			transID, _ := utils.GeneratePaymentID(ctx, s.redis)
			transaction := &models.BalanceTransaction{
				TransactionID:    transID,
				TenantID:         tenantID,
				TransactionType:  models.TransactionTypeDeduct,
				Amount:           -baseCurrencyAmount,
				BalanceBefore:    balance.Balance,
				BalanceAfter:     balance.Balance - baseCurrencyAmount,
				RelatedPaymentID: &paymentID,
			}

			if billID != nil {
				transaction.RelatedBillID = billID
			}
			if orderID != nil {
				transaction.RelatedOrderID = orderID
			}

			if err := s.balanceTransRepo.Create(ctx, transaction); err != nil {
				return err
			}

			// 更新支付状态
			now := time.Now()
			payment.Status = models.PaymentStatusSuccess
			payment.PaidAt = &now

			if err := s.paymentRepo.UpdateStatus(ctx, paymentID, models.PaymentStatusSuccess); err != nil {
				return err
			}

			// 更新账单状态
			if billID != nil {
				if err := s.billRepo.UpdateStatus(ctx, *billID, models.BillStatusPaid); err != nil {
					return err
				}
			}

			// 更新订单状态
			if orderID != nil {
				if err := s.orderRepo.UpdateStatus(ctx, *orderID, models.OrderStatusPaid); err != nil {
					return err
				}
			}

			return nil
		})

		if err != nil {
			return nil, errors.Wrap(err, "余额支付失败")
		}
	} else {
		// 第三方支付（简化实现，实际需要对接支付网关）
		if err := s.paymentRepo.Create(ctx, payment); err != nil {
			return nil, errors.NewInternalError("创建支付记录失败: " + err.Error())
		}
	}

	return payment, nil
}

func (s *paymentService) GetPayment(ctx context.Context, paymentID string) (*models.Payment, error) {
	payment, err := s.paymentRepo.GetByID(ctx, paymentID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.New(errors.ErrNotFound, "支付记录不存在")
		}
		return nil, errors.NewInternalError("查询支付记录失败: " + err.Error())
	}
	return payment, nil
}

// ============================================================================
// 余额操作
// ============================================================================

func (s *paymentService) GetBalance(ctx context.Context, tenantID string) (*models.AccountBalance, error) {
	balance, err := s.balanceRepo.GetByTenantID(ctx, tenantID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			// 账户不存在，创建初始余额
			balance = &models.AccountBalance{
				TenantID:      tenantID,
				Balance:       0,
				FrozenBalance: 0,
				CreditLimit:   0,
				Currency:      models.CurrencyCNY,
			}
			if err := s.balanceRepo.Create(ctx, balance); err != nil {
				return nil, errors.NewInternalError("创建账户余额失败: " + err.Error())
			}
		} else {
			return nil, errors.NewInternalError("查询账户余额失败: " + err.Error())
		}
	}
	return balance, nil
}

func (s *paymentService) Recharge(ctx context.Context, tenantID string, amount float64) error {
	if amount <= 0 {
		return errors.New(errors.ErrInvalidParams, "充值金额必须大于0")
	}

	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// 查询当前余额
		balance, err := s.balanceRepo.GetByTenantID(ctx, tenantID)
		if err != nil {
			return err
		}

		// 更新余额
		if err := s.balanceRepo.UpdateBalance(ctx, tenantID, amount, 0); err != nil {
			return err
		}

		// 记录余额变动
		transID, _ := utils.GeneratePaymentID(ctx, s.redis)
		transaction := &models.BalanceTransaction{
			TransactionID:   transID,
			TenantID:        tenantID,
			TransactionType: models.TransactionTypeRecharge,
			Amount:          amount,
			BalanceBefore:   balance.Balance,
			BalanceAfter:    balance.Balance + amount,
		}

		return s.balanceTransRepo.Create(ctx, transaction)
	})
}

// GetBalanceTransactions 获取余额变动记录
func (s *paymentService) GetBalanceTransactions(ctx context.Context, tenantID string, page, pageSize int) ([]models.BalanceTransaction, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	transactions, total, err := s.balanceTransRepo.ListByTenantPaginated(ctx, tenantID, page, pageSize)
	if err != nil {
		return nil, 0, errors.NewInternalError("查询余额变动记录失败: " + err.Error())
	}

	return transactions, total, nil
}
