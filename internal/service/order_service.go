package service

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/wwnj/happy-billing/internal/models"
	"github.com/wwnj/happy-billing/internal/repository"
	"github.com/wwnj/happy-billing/pkg/errors"
	"github.com/wwnj/happy-billing/pkg/utils"
	"gorm.io/gorm"
)

// OrderService 订单服务接口
type OrderService interface {
	// 订单管理
	CreateOrder(ctx context.Context, req *models.CreateOrderRequest) (*models.Order, error)
	GetOrder(ctx context.Context, orderID string) (*models.Order, error)
	ListOrders(ctx context.Context, req *models.OrderListQueryRequest) (*models.PageResult, error)
	CancelOrder(ctx context.Context, orderID string) error
}

type orderService struct {
	db              *gorm.DB
	redis           *redis.Client
	orderRepo       repository.OrderRepository
	orderItemRepo   repository.OrderItemRepository
	resourceRepo    repository.ResourceInstanceRepository
	billRepo        repository.BillRepository
	skuRepo         repository.ProductSkuRepository
	tenantRepo      repository.TenantRepository
	pricingService  PricingService
	currencyService CurrencyService
}

// NewOrderService 创建订单服务实例
func NewOrderService(
	db *gorm.DB,
	redis *redis.Client,
	orderRepo repository.OrderRepository,
	orderItemRepo repository.OrderItemRepository,
	resourceRepo repository.ResourceInstanceRepository,
	billRepo repository.BillRepository,
	skuRepo repository.ProductSkuRepository,
	tenantRepo repository.TenantRepository,
	pricingService PricingService,
	currencyService CurrencyService,
) OrderService {
	return &orderService{
		db:              db,
		redis:           redis,
		orderRepo:       orderRepo,
		orderItemRepo:   orderItemRepo,
		resourceRepo:    resourceRepo,
		billRepo:        billRepo,
		skuRepo:         skuRepo,
		tenantRepo:      tenantRepo,
		pricingService:  pricingService,
		currencyService: currencyService,
	}
}

// ============================================================================
// 订单管理
// ============================================================================

func (s *orderService) CreateOrder(ctx context.Context, req *models.CreateOrderRequest) (*models.Order, error) {
	// 1. 查询SKU信息
	sku, err := s.skuRepo.GetByCode(ctx, req.SkuCode)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.New(errors.ErrSKUNotFound, "SKU不存在")
		}
		return nil, errors.NewInternalError("查询SKU失败: " + err.Error())
	}

	// 2. 计算价格
	priceReq := &models.PriceCalculateRequest{
		SkuCode:   req.SkuCode,
		TenantID:  req.TenantID,
		Quantity:  req.Quantity,
		StartTime: time.Now(),
		EndTime:   time.Now().Add(24 * time.Hour), // 默认1天
	}
	if req.PeriodStart != nil && req.PeriodEnd != nil {
		priceReq.StartTime = *req.PeriodStart
		priceReq.EndTime = *req.PeriodEnd
	}

	priceResult, err := s.pricingService.CalculatePrice(ctx, priceReq)
	if err != nil {
		return nil, errors.Wrap(err, "计算价格失败")
	}

	// 2.5 多币种处理：获取租户偏好币种和汇率
	tenant, err := s.tenantRepo.GetByID(ctx, req.TenantID)
	if err != nil {
		return nil, errors.NewInternalError("查询租户失败: " + err.Error())
	}

	currency := models.Currency(tenant.PreferredCurrency)
	if currency == "" {
		currency = models.CurrencyCNY // 默认人民币
	}

	// 计算本位币金额和汇率
	baseCurrencyAmount, exchangeRate, err := s.currencyService.CalculateBaseCurrencyAmount(ctx, priceResult.FinalPrice, currency, nil)
	if err != nil {
		return nil, errors.Wrap(err, "计算本位币金额失败")
	}

	// 3. 生成订单ID和订单号
	orderID, err := utils.GenerateOrderID(ctx, s.redis)
	if err != nil {
		return nil, errors.NewInternalError("生成订单ID失败: " + err.Error())
	}

	orderNo := fmt.Sprintf("ORD%s", time.Now().Format("20060102150405"))

	// 4. 创建订单
	order := &models.Order{
		OrderID:            orderID,
		OrderNo:            orderNo,
		TenantID:           req.TenantID,
		OrganizationID:     req.OrganizationID,
		ProjectID:          req.ProjectID,
		UserID:             req.UserID,
		OrderType:          req.OrderType,
		SpuCode:            sku.SpuCode,
		SkuCode:            req.SkuCode,
		Currency:           currency,
		ExchangeRate:       &exchangeRate,
		BaseCurrency:       models.CurrencyCNY,
		BaseCurrencyAmount: &baseCurrencyAmount,
		OriginalAmount:     priceResult.OriginalPrice,
		DiscountAmount:     priceResult.DiscountAmount,
		PayableAmount:      priceResult.FinalPrice,
		PaidAmount:         0,
		PeriodStart:        req.PeriodStart,
		PeriodEnd:          req.PeriodEnd,
		Status:             models.OrderStatusPending,
		OrderDetail:        (*models.OrderDetail)(&req.OrderDetail),
	}

	// 5. 使用事务创建订单和订单明细
	err = s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// 创建订单
		if err := s.orderRepo.Create(ctx, order); err != nil {
			return err
		}

		// 创建订单明细
		orderItem := models.OrderItem{
			OrderID:   orderID,
			ItemNo:    fmt.Sprintf("%s-001", orderNo),
			SpuCode:   sku.SpuCode,
			SkuCode:   req.SkuCode,
			SkuName:   sku.SkuName,
			SkuSpec:   &sku.SpecValues,
			Quantity:  req.Quantity,
			UnitPrice: priceResult.OriginalPrice / req.Quantity,
			Amount:    priceResult.OriginalPrice,
		}
		if priceResult.PriceRule != nil {
			orderItem.PriceRuleID = &priceResult.PriceRule.RuleID
		}

		if err := s.orderItemRepo.Create(ctx, &orderItem); err != nil {
			return err
		}

		// 如果是预付费，创建账单
		if req.OrderType == models.OrderTypePrepaid {
			billID, err := utils.GenerateBillID(ctx, s.redis)
			if err != nil {
				return err
			}

			baseCurrency := models.CurrencyCNY
			bill := &models.Bill{
				BillID:             billID,
				OrderID:            orderID,
				TenantID:           req.TenantID,
				ProjectID:          req.ProjectID,
				BillType:           "SUBSCRIPTION",
				Currency:           order.Currency,
				ExchangeRate:       order.ExchangeRate,
				BaseCurrency:       &baseCurrency,
				BaseCurrencyAmount: order.BaseCurrencyAmount,
				OriginalAmount:     order.OriginalAmount,
				DiscountAmount:     order.DiscountAmount,
				PayableAmount:      order.PayableAmount,
				Status:             models.BillStatusUnpaid,
			}

			if err := s.billRepo.Create(ctx, bill); err != nil {
				return err
			}
		}

		return nil
	})

	if err != nil {
		return nil, errors.NewInternalError("创建订单失败: " + err.Error())
	}

	return order, nil
}

func (s *orderService) GetOrder(ctx context.Context, orderID string) (*models.Order, error) {
	order, err := s.orderRepo.GetByID(ctx, orderID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.NewWithCode(errors.ErrOrderNotFound)
		}
		return nil, errors.NewInternalError("查询订单失败: " + err.Error())
	}
	return order, nil
}

func (s *orderService) ListOrders(ctx context.Context, req *models.OrderListQueryRequest) (*models.PageResult, error) {
	orders, total, err := s.orderRepo.List(ctx, req)
	if err != nil {
		return nil, errors.NewInternalError("查询订单列表失败: " + err.Error())
	}

	return &models.PageResult{
		Total:    total,
		Page:     req.Page,
		PageSize: req.PageSize,
		Data:     orders,
	}, nil
}

func (s *orderService) CancelOrder(ctx context.Context, orderID string) error {
	// 查询订单
	order, err := s.orderRepo.GetByID(ctx, orderID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return errors.NewWithCode(errors.ErrOrderNotFound)
		}
		return errors.NewInternalError("查询订单失败: " + err.Error())
	}

	// 只有待支付状态的订单可以取消
	if order.Status != models.OrderStatusPending {
		return errors.New(errors.ErrOrderCannotCancel, "订单状态不允许取消")
	}

	// 更新订单状态
	if err := s.orderRepo.UpdateStatus(ctx, orderID, models.OrderStatusCancelled); err != nil {
		return errors.NewInternalError("取消订单失败: " + err.Error())
	}

	return nil
}
