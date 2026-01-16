package pricers

import (
	"fmt"
	"sync"

	"github.com/wwnj/happy-billing/internal/domain/pricing"
)

// Registry 定价器注册中心(单例)
type Registry struct {
	mu      sync.RWMutex
	pricers map[string]pricing.PricerPlugin
}

var (
	once     sync.Once
	instance *Registry
)

// GetRegistry 获取注册中心单例
func GetRegistry() *Registry {
	once.Do(func() {
		instance = &Registry{
			pricers: make(map[string]pricing.PricerPlugin),
		}
	})
	return instance
}

// Register 注册定价器插件
func (r *Registry) Register(plugin pricing.PricerPlugin) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	name := plugin.Name()
	if name == "" {
		return fmt.Errorf("plugin name cannot be empty")
	}

	if _, exists := r.pricers[name]; exists {
		return fmt.Errorf("pricer plugin %s already registered", name)
	}

	r.pricers[name] = plugin
	return nil
}

// Get 获取指定名称的定价器插件
func (r *Registry) Get(name string) (pricing.PricerPlugin, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	plugin, ok := r.pricers[name]
	if !ok {
		return nil, fmt.Errorf("pricer plugin %s not found", name)
	}

	return plugin, nil
}

// List 列出所有已注册的定价器名称
func (r *Registry) List() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	names := make([]string, 0, len(r.pricers))
	for name := range r.pricers {
		names = append(names, name)
	}
	return names
}

// Has 检查指定名称的定价器是否存在
func (r *Registry) Has(name string) bool {
	r.mu.RLock()
	defer r.mu.RUnlock()

	_, ok := r.pricers[name]
	return ok
}

// Unregister 注销定价器插件
func (r *Registry) Unregister(name string) {
	r.mu.Lock()
	defer r.mu.Unlock()

	delete(r.pricers, name)
}

// Count 返回已注册的定价器数量
func (r *Registry) Count() int {
	r.mu.RLock()
	defer r.mu.RUnlock()

	return len(r.pricers)
}

// 便捷函数(直接使用全局单例)

// Register 注册定价器插件(全局)
func Register(plugin pricing.PricerPlugin) error {
	return GetRegistry().Register(plugin)
}

// Get 获取定价器插件(全局)
func Get(name string) (pricing.PricerPlugin, error) {
	return GetRegistry().Get(name)
}

// List 列出所有定价器(全局)
func List() []string {
	return GetRegistry().List()
}

// Has 检查定价器是否存在(全局)
func Has(name string) bool {
	return GetRegistry().Has(name)
}
