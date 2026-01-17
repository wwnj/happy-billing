package tracing

import (
	"context"
	"fmt"
	"net"
	"strings"

	"github.com/redis/go-redis/v9"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	semconv "go.opentelemetry.io/otel/semconv/v1.17.0"
)

// RedisTracingHook Redis 追踪 Hook
type RedisTracingHook struct {
	addr string
}

// NewRedisTracingHook 创建 Redis 追踪 Hook
func NewRedisTracingHook(addr string) *RedisTracingHook {
	return &RedisTracingHook{
		addr: addr,
	}
}

// DialHook 实现 redis.DialHook
func (h *RedisTracingHook) DialHook(next redis.DialHook) redis.DialHook {
	return func(ctx context.Context, network, addr string) (net.Conn, error) {
		return next(ctx, network, addr)
	}
}

// ProcessHook 实现 redis.ProcessHook
func (h *RedisTracingHook) ProcessHook(next redis.ProcessHook) redis.ProcessHook {
	return func(ctx context.Context, cmd redis.Cmder) error {
		// 创建 Span
		cmdName := cmd.Name()
		spanName := fmt.Sprintf("redis:%s", cmdName)

		ctx, span := StartSpan(ctx, spanName)
		defer span.End()

		// 设置 Redis 相关属性
		attrs := []attribute.KeyValue{
			attribute.String("db.system", "redis"),
			attribute.String("db.operation", cmdName),
			attribute.String("db.statement", formatCommand(cmd)),
		}

		// 添加 Redis 地址
		if h.addr != "" {
			attrs = append(attrs, semconv.NetPeerName(h.addr))
		}

		// 添加业务上下文（租户ID、用户ID等）
		businessCtx := GetBusinessContext(ctx)
		AddBusinessContextToSpan(span, businessCtx)

		span.SetAttributes(attrs...)

		// 执行命令
		err := next(ctx, cmd)

		// 记录错误
		if err != nil && err != redis.Nil {
			span.RecordError(err)
			span.SetStatus(codes.Error, err.Error())
		} else {
			span.SetStatus(codes.Ok, "")
		}

		return err
	}
}

// ProcessPipelineHook 实现 redis.ProcessPipelineHook
func (h *RedisTracingHook) ProcessPipelineHook(next redis.ProcessPipelineHook) redis.ProcessPipelineHook {
	return func(ctx context.Context, cmds []redis.Cmder) error {
		// 创建 Span
		spanName := fmt.Sprintf("redis:pipeline(%d commands)", len(cmds))

		ctx, span := StartSpan(ctx, spanName)
		defer span.End()

		// 设置 Redis 相关属性
		attrs := []attribute.KeyValue{
			attribute.String("db.system", "redis"),
			attribute.String("db.operation", "pipeline"),
			attribute.Int("db.pipeline.size", len(cmds)),
		}

		// 添加 Redis 地址
		if h.addr != "" {
			attrs = append(attrs, semconv.NetPeerName(h.addr))
		}

		// 添加业务上下文
		businessCtx := GetBusinessContext(ctx)
		AddBusinessContextToSpan(span, businessCtx)

		// 添加命令列表
		if len(cmds) <= 10 {
			cmdNames := make([]string, len(cmds))
			for i, cmd := range cmds {
				cmdNames[i] = cmd.Name()
			}
			attrs = append(attrs, attribute.String("db.pipeline.commands", strings.Join(cmdNames, ",")))
		}

		span.SetAttributes(attrs...)

		// 执行 Pipeline
		err := next(ctx, cmds)

		// 记录错误
		if err != nil && err != redis.Nil {
			span.RecordError(err)
			span.SetStatus(codes.Error, err.Error())
		} else {
			span.SetStatus(codes.Ok, "")
		}

		return err
	}
}

// formatCommand 格式化 Redis 命令
func formatCommand(cmd redis.Cmder) string {
	args := cmd.Args()
	if len(args) == 0 {
		return ""
	}

	// 构建命令字符串
	parts := make([]string, len(args))
	for i, arg := range args {
		parts[i] = fmt.Sprintf("%v", arg)
	}

	// 限制长度
	cmdStr := strings.Join(parts, " ")
	if len(cmdStr) > 100 {
		cmdStr = cmdStr[:100] + "..."
	}

	return cmdStr
}

// InstallRedisHook 为 Redis 客户端安装追踪 Hook
func InstallRedisHook(client *redis.Client, addr string) {
	hook := NewRedisTracingHook(addr)
	client.AddHook(hook)
}
