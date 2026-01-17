package tracing

import (
	"context"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

// 业务上下文 Key
type contextKey string

const (
	// ContextKeyTenantID 租户ID上下文键
	ContextKeyTenantID contextKey = "tenant_id"
	// ContextKeyUserID 用户ID上下文键
	ContextKeyUserID contextKey = "user_id"
	// ContextKeyOrgID 组织ID上下文键
	ContextKeyOrgID contextKey = "org_id"
	// ContextKeyProjectID 项目ID上下文键
	ContextKeyProjectID contextKey = "project_id"
	// ContextKeyOrderID 订单ID上下文键
	ContextKeyOrderID contextKey = "order_id"
	// ContextKeyBillID 账单ID上下文键
	ContextKeyBillID contextKey = "bill_id"
	// ContextKeyRequestID 请求ID上下文键
	ContextKeyRequestID contextKey = "request_id"
)

// Span 属性键名（用于 OpenTelemetry）
const (
	AttrTenantID  = "tenant.id"
	AttrUserID    = "user.id"
	AttrOrgID     = "org.id"
	AttrProjectID = "project.id"
	AttrOrderID   = "order.id"
	AttrBillID    = "bill.id"
	AttrRequestID = "request.id"
)

// HTTP Header 键名
const (
	HeaderTenantID  = "X-Tenant-ID"
	HeaderUserID    = "X-User-ID"
	HeaderOrgID     = "X-Org-ID"
	HeaderProjectID = "X-Project-ID"
	HeaderRequestID = "X-Request-ID"
)

// BusinessContext 业务上下文
type BusinessContext struct {
	TenantID  string
	UserID    string
	OrgID     string
	ProjectID string
	OrderID   string
	BillID    string
	RequestID string
}

// SetTenantID 设置租户ID到上下文和Span
func SetTenantID(ctx context.Context, tenantID string) context.Context {
	if tenantID == "" {
		return ctx
	}
	// 设置到 Span
	SetAttributes(ctx, attribute.String(AttrTenantID, tenantID))
	// 设置到 Context
	return context.WithValue(ctx, ContextKeyTenantID, tenantID)
}

// GetTenantID 从上下文获取租户ID
func GetTenantID(ctx context.Context) string {
	if v := ctx.Value(ContextKeyTenantID); v != nil {
		if tenantID, ok := v.(string); ok {
			return tenantID
		}
	}
	return ""
}

// SetUserID 设置用户ID到上下文和Span
func SetUserID(ctx context.Context, userID string) context.Context {
	if userID == "" {
		return ctx
	}
	SetAttributes(ctx, attribute.String(AttrUserID, userID))
	return context.WithValue(ctx, ContextKeyUserID, userID)
}

// GetUserID 从上下文获取用户ID
func GetUserID(ctx context.Context) string {
	if v := ctx.Value(ContextKeyUserID); v != nil {
		if userID, ok := v.(string); ok {
			return userID
		}
	}
	return ""
}

// SetOrgID 设置组织ID到上下文和Span
func SetOrgID(ctx context.Context, orgID string) context.Context {
	if orgID == "" {
		return ctx
	}
	SetAttributes(ctx, attribute.String(AttrOrgID, orgID))
	return context.WithValue(ctx, ContextKeyOrgID, orgID)
}

// GetOrgID 从上下文获取组织ID
func GetOrgID(ctx context.Context) string {
	if v := ctx.Value(ContextKeyOrgID); v != nil {
		if orgID, ok := v.(string); ok {
			return orgID
		}
	}
	return ""
}

// SetProjectID 设置项目ID到上下文和Span
func SetProjectID(ctx context.Context, projectID string) context.Context {
	if projectID == "" {
		return ctx
	}
	SetAttributes(ctx, attribute.String(AttrProjectID, projectID))
	return context.WithValue(ctx, ContextKeyProjectID, projectID)
}

// GetProjectID 从上下文获取项目ID
func GetProjectID(ctx context.Context) string {
	if v := ctx.Value(ContextKeyProjectID); v != nil {
		if projectID, ok := v.(string); ok {
			return projectID
		}
	}
	return ""
}

// SetOrderID 设置订单ID到上下文和Span
func SetOrderID(ctx context.Context, orderID string) context.Context {
	if orderID == "" {
		return ctx
	}
	SetAttributes(ctx, attribute.String(AttrOrderID, orderID))
	return context.WithValue(ctx, ContextKeyOrderID, orderID)
}

// GetOrderID 从上下文获取订单ID
func GetOrderID(ctx context.Context) string {
	if v := ctx.Value(ContextKeyOrderID); v != nil {
		if orderID, ok := v.(string); ok {
			return orderID
		}
	}
	return ""
}

// SetBillID 设置账单ID到上下文和Span
func SetBillID(ctx context.Context, billID string) context.Context {
	if billID == "" {
		return ctx
	}
	SetAttributes(ctx, attribute.String(AttrBillID, billID))
	return context.WithValue(ctx, ContextKeyBillID, billID)
}

// GetBillID 从上下文获取账单ID
func GetBillID(ctx context.Context) string {
	if v := ctx.Value(ContextKeyBillID); v != nil {
		if billID, ok := v.(string); ok {
			return billID
		}
	}
	return ""
}

// SetRequestID 设置请求ID到上下文和Span
func SetRequestID(ctx context.Context, requestID string) context.Context {
	if requestID == "" {
		return ctx
	}
	SetAttributes(ctx, attribute.String(AttrRequestID, requestID))
	return context.WithValue(ctx, ContextKeyRequestID, requestID)
}

// GetRequestID 从上下文获取请求ID
func GetRequestID(ctx context.Context) string {
	if v := ctx.Value(ContextKeyRequestID); v != nil {
		if requestID, ok := v.(string); ok {
			return requestID
		}
	}
	return ""
}

// SetBusinessContext 批量设置业务上下文
func SetBusinessContext(ctx context.Context, bc *BusinessContext) context.Context {
	if bc == nil {
		return ctx
	}

	if bc.TenantID != "" {
		ctx = SetTenantID(ctx, bc.TenantID)
	}
	if bc.UserID != "" {
		ctx = SetUserID(ctx, bc.UserID)
	}
	if bc.OrgID != "" {
		ctx = SetOrgID(ctx, bc.OrgID)
	}
	if bc.ProjectID != "" {
		ctx = SetProjectID(ctx, bc.ProjectID)
	}
	if bc.OrderID != "" {
		ctx = SetOrderID(ctx, bc.OrderID)
	}
	if bc.BillID != "" {
		ctx = SetBillID(ctx, bc.BillID)
	}
	if bc.RequestID != "" {
		ctx = SetRequestID(ctx, bc.RequestID)
	}

	return ctx
}

// GetBusinessContext 获取完整业务上下文
func GetBusinessContext(ctx context.Context) *BusinessContext {
	return &BusinessContext{
		TenantID:  GetTenantID(ctx),
		UserID:    GetUserID(ctx),
		OrgID:     GetOrgID(ctx),
		ProjectID: GetProjectID(ctx),
		OrderID:   GetOrderID(ctx),
		BillID:    GetBillID(ctx),
		RequestID: GetRequestID(ctx),
	}
}

// AddBusinessContextToSpan 将业务上下文添加到 Span 属性
func AddBusinessContextToSpan(span trace.Span, bc *BusinessContext) {
	if bc == nil {
		return
	}

	attrs := make([]attribute.KeyValue, 0, 7)
	if bc.TenantID != "" {
		attrs = append(attrs, attribute.String(AttrTenantID, bc.TenantID))
	}
	if bc.UserID != "" {
		attrs = append(attrs, attribute.String(AttrUserID, bc.UserID))
	}
	if bc.OrgID != "" {
		attrs = append(attrs, attribute.String(AttrOrgID, bc.OrgID))
	}
	if bc.ProjectID != "" {
		attrs = append(attrs, attribute.String(AttrProjectID, bc.ProjectID))
	}
	if bc.OrderID != "" {
		attrs = append(attrs, attribute.String(AttrOrderID, bc.OrderID))
	}
	if bc.BillID != "" {
		attrs = append(attrs, attribute.String(AttrBillID, bc.BillID))
	}
	if bc.RequestID != "" {
		attrs = append(attrs, attribute.String(AttrRequestID, bc.RequestID))
	}

	if len(attrs) > 0 {
		span.SetAttributes(attrs...)
	}
}
