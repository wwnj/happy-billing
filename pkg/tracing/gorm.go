package tracing

import (
	"context"
	"fmt"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	semconv "go.opentelemetry.io/otel/semconv/v1.17.0"
	"gorm.io/gorm"
)

const (
	gormSpanKey = "gorm_span_key"
	// DBSystemMySQL MySQL 数据库系统标识
	DBSystemMySQL = "mysql"
	// DBSystemClickHouse ClickHouse 数据库系统标识
	DBSystemClickHouse = "clickhouse"
)

// GormTracingPlugin GORM 追踪插件
type GormTracingPlugin struct {
	dbSystem string // mysql, clickhouse
}

// NewGormTracingPlugin 创建 GORM 追踪插件
func NewGormTracingPlugin(dbSystem string) *GormTracingPlugin {
	return &GormTracingPlugin{
		dbSystem: dbSystem,
	}
}

// Name 插件名称
func (p *GormTracingPlugin) Name() string {
	return "otel:tracing"
}

// Initialize 初始化插件
func (p *GormTracingPlugin) Initialize(db *gorm.DB) error {
	// 注册回调
	if err := db.Callback().Create().Before("gorm:create").Register("otel:before_create", p.before); err != nil {
		return err
	}
	if err := db.Callback().Create().After("gorm:create").Register("otel:after_create", p.after); err != nil {
		return err
	}

	if err := db.Callback().Query().Before("gorm:query").Register("otel:before_query", p.before); err != nil {
		return err
	}
	if err := db.Callback().Query().After("gorm:query").Register("otel:after_query", p.after); err != nil {
		return err
	}

	if err := db.Callback().Update().Before("gorm:update").Register("otel:before_update", p.before); err != nil {
		return err
	}
	if err := db.Callback().Update().After("gorm:update").Register("otel:after_update", p.after); err != nil {
		return err
	}

	if err := db.Callback().Delete().Before("gorm:delete").Register("otel:before_delete", p.before); err != nil {
		return err
	}
	if err := db.Callback().Delete().After("gorm:delete").Register("otel:after_delete", p.after); err != nil {
		return err
	}

	if err := db.Callback().Row().Before("gorm:row").Register("otel:before_row", p.before); err != nil {
		return err
	}
	if err := db.Callback().Row().After("gorm:row").Register("otel:after_row", p.after); err != nil {
		return err
	}

	if err := db.Callback().Raw().Before("gorm:raw").Register("otel:before_raw", p.before); err != nil {
		return err
	}
	if err := db.Callback().Raw().After("gorm:raw").Register("otel:after_raw", p.after); err != nil {
		return err
	}

	return nil
}

// before 操作前回调
func (p *GormTracingPlugin) before(db *gorm.DB) {
	// 从上下文获取或创建 Span
	ctx := db.Statement.Context
	if ctx == nil {
		ctx = context.Background()
	}

	// 创建 Span
	operation := extractOperation(db)
	spanName := fmt.Sprintf("gorm:%s", operation)

	ctx, span := StartSpan(ctx, spanName)

	// 设置数据库相关属性
	attrs := []attribute.KeyValue{
		attribute.String("db.system", p.dbSystem),
		attribute.String("db.operation", operation),
		attribute.String("db.sql", db.Statement.SQL.String()),
	}

	// 添加表名
	if db.Statement.Table != "" {
		attrs = append(attrs, semconv.DBSQLTable(db.Statement.Table))
	}

	// 添加业务上下文（租户ID、用户ID等）
	businessCtx := GetBusinessContext(ctx)
	AddBusinessContextToSpan(span, businessCtx)

	span.SetAttributes(attrs...)

	// 保存 Span 到 Context
	db.Statement.Context = ctx
	db.InstanceSet(gormSpanKey, span)
}

// after 操作后回调
func (p *GormTracingPlugin) after(db *gorm.DB) {
	// 获取 Span
	_span, ok := db.InstanceGet(gormSpanKey)
	if !ok {
		return
	}

	// 正确的类型断言：使用 trace.Span 接口
	span, ok := _span.(interface {
		End(...interface{})
		SetStatus(codes.Code, string)
		SetAttributes(...attribute.KeyValue)
		RecordError(error, ...interface{})
	})
	if !ok {
		return
	}
	defer span.End()

	// 记录影响的行数
	if db.Statement.RowsAffected >= 0 {
		span.SetAttributes(attribute.Int64("db.rows_affected", db.Statement.RowsAffected))
	}

	// 如果有错误，记录错误
	if db.Error != nil && db.Error != gorm.ErrRecordNotFound {
		span.RecordError(db.Error)
		span.SetStatus(codes.Error, db.Error.Error())
	} else {
		span.SetStatus(codes.Ok, "")
	}
}

// extractOperation 提取操作类型
func extractOperation(db *gorm.DB) string {
	switch {
	case db.Statement.SQL.String() != "":
		return "raw"
	case db.Statement.Schema != nil:
		switch {
		case db.Statement.ReflectValue.Kind() == 0:
			return "query"
		case db.Statement.Dest != db.Statement.Model:
			return "query"
		default:
			return "exec"
		}
	default:
		return "unknown"
	}
}

// WithContext 为 GORM 添加带追踪的上下文
func WithContext(ctx context.Context, db *gorm.DB) *gorm.DB {
	return db.WithContext(ctx)
}
