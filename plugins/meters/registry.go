package meters

import (
	"fmt"
	"sync"

	"github.com/wwnj/happy-billing/internal/domain/meter"
)

// Registry 计量器注册中心(单例)
type Registry struct {
	mu     sync.RWMutex
	meters map[string]meter.MeterPlugin
}

var (
	once     sync.Once
	instance *Registry
)

// GetRegistry 获取注册中心单例
func GetRegistry() *Registry {
	once.Do(func() {
		instance = &Registry{
			meters: make(map[string]meter.MeterPlugin),
		}
	})
	return instance
}

// Register 注册计量器插件
func (r *Registry) Register(plugin meter.MeterPlugin) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	name := plugin.Name()
	if name == "" {
		return fmt.Errorf("plugin name cannot be empty")
	}

	if _, exists := r.meters[name]; exists {
		return fmt.Errorf("meter plugin %s already registered", name)
	}

	r.meters[name] = plugin
	return nil
}

// Get 获取指定名称的计量器插件
func (r *Registry) Get(name string) (meter.MeterPlugin, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	plugin, ok := r.meters[name]
	if !ok {
		return nil, fmt.Errorf("meter plugin %s not found", name)
	}

	return plugin, nil
}

// List 列出所有已注册的计量器名称
func (r *Registry) List() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	names := make([]string, 0, len(r.meters))
	for name := range r.meters {
		names = append(names, name)
	}
	return names
}

// Has 检查指定名称的计量器是否存在
func (r *Registry) Has(name string) bool {
	r.mu.RLock()
	defer r.mu.RUnlock()

	_, ok := r.meters[name]
	return ok
}

// Unregister 注销计量器插件
func (r *Registry) Unregister(name string) {
	r.mu.Lock()
	defer r.mu.Unlock()

	delete(r.meters, name)
}

// Count 返回已注册的计量器数量
func (r *Registry) Count() int {
	r.mu.RLock()
	defer r.mu.RUnlock()

	return len(r.meters)
}

// 便捷函数(直接使用全局单例)

// Register 注册计量器插件(全局)
func Register(plugin meter.MeterPlugin) error {
	return GetRegistry().Register(plugin)
}

// Get 获取计量器插件(全局)
func Get(name string) (meter.MeterPlugin, error) {
	return GetRegistry().Get(name)
}

// List 列出所有计量器(全局)
func List() []string {
	return GetRegistry().List()
}

// Has 检查计量器是否存在(全局)
func Has(name string) bool {
	return GetRegistry().Has(name)
}
