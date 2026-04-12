package storage

import (
	"fmt"
	"sync"
)

// Manager 管理多个存储配置的驱动实例。
// 通过存储配置 ID 索引驱动，支持热加载和默认存储切换。
type Manager struct {
	mu       sync.RWMutex
	drivers  map[string]StorageDriver // configID -> driver
	configs  map[string]StorageConfig // configID -> config
	defaultID string                  // 默认存储配置 ID
}

// NewManager 创建存储管理器。
func NewManager() *Manager {
	return &Manager{
		drivers: make(map[string]StorageDriver),
		configs: make(map[string]StorageConfig),
	}
}

// Register 注册一个存储驱动实例。
func (m *Manager) Register(configID string, cfg StorageConfig, driver StorageDriver) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.drivers[configID] = driver
	m.configs[configID] = cfg
}

// LoadAndRegister 加载配置并注册驱动。
func (m *Manager) LoadAndRegister(configID string, cfg StorageConfig) error {
	driver, err := NewDriver(cfg)
	if err != nil {
		return fmt.Errorf("load storage %q: %w", configID, err)
	}
	m.Register(configID, cfg, driver)
	return nil
}

// Get 获取指定配置 ID 的存储驱动。
func (m *Manager) Get(configID string) (StorageDriver, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	driver, ok := m.drivers[configID]
	if !ok {
		return nil, fmt.Errorf("storage driver not found for config %q", configID)
	}
	return driver, nil
}

// Default 获取默认存储驱动。
func (m *Manager) Default() (StorageDriver, string, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if m.defaultID == "" {
		return nil, "", fmt.Errorf("no default storage configured")
	}

	driver, ok := m.drivers[m.defaultID]
	if !ok {
		return nil, "", fmt.Errorf("default storage driver %q not loaded", m.defaultID)
	}
	return driver, m.defaultID, nil
}

// SetDefault 设置默认存储配置 ID。
func (m *Manager) SetDefault(configID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, ok := m.drivers[configID]; !ok {
		return fmt.Errorf("cannot set default: storage %q not registered", configID)
	}
	m.defaultID = configID
	return nil
}

// Remove 移除指定配置 ID 的存储驱动。
func (m *Manager) Remove(configID string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	delete(m.drivers, configID)
	delete(m.configs, configID)
	if m.defaultID == configID {
		m.defaultID = ""
	}
}

// List 列出所有已注册的存储配置 ID。
func (m *Manager) List() []string {
	m.mu.RLock()
	defer m.mu.RUnlock()

	ids := make([]string, 0, len(m.drivers))
	for id := range m.drivers {
		ids = append(ids, id)
	}
	return ids
}
