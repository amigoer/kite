package storage

import (
	"encoding/json"
	"fmt"
	"sort"
	"sync"
)

// ConfigMeta is the manager's complete metadata view of a single storage configuration.
// It bundles the driver's StorageConfig together with the priority, capacity, and default
// flag that upstream routing policies need.
type ConfigMeta struct {
	ID                 string
	Name               string
	Driver             string
	Provider           string
	SchemeKey          string
	DriverConfig       StorageConfig
	Priority           int
	CapacityLimitBytes int64
	IsDefault          bool
	IsActive           bool
}

// Manager owns the driver instances for every storage configuration.
// State is replaced wholesale on Reload so the default-storage view cannot drift from the
// database after CRUD operations.
type Manager struct {
	mu        sync.RWMutex
	drivers   map[string]StorageDriver
	metas     map[string]ConfigMeta
	defaultID string
}

// NewManager builds an empty Manager; callers must invoke Reload before use.
func NewManager() *Manager {
	return &Manager{
		drivers: make(map[string]StorageDriver),
		metas:   make(map[string]ConfigMeta),
	}
}

// RawConfig is the raw input to Reload.
// Callers load the rows from the database and pass them in; the manager stays free of a
// dependency on the repo package so the two do not form an import cycle.
type RawConfig struct {
	ID                 string
	Name               string
	Driver             string
	Provider           string
	ConfigJSON         string
	Priority           int
	CapacityLimitBytes int64
	IsDefault          bool
	IsActive           bool
}

// Reload rebuilds the driver and metadata maps from the latest database snapshot.
// The swap is atomic; parse or construction failures are skipped, collected into a single
// aggregate error, and do not prevent the remaining drivers from loading.
func (m *Manager) Reload(rawConfigs []RawConfig) error {
	newDrivers := make(map[string]StorageDriver, len(rawConfigs))
	newMetas := make(map[string]ConfigMeta, len(rawConfigs))
	var defaultID string
	var errs []string

	for _, raw := range rawConfigs {
		if !raw.IsActive {
			continue
		}

		canonicalDriver, provider := CanonicalDriverAndProvider(raw.Driver, &raw.Provider, raw.ConfigJSON)

		scfg, err := ParseConfig(canonicalDriver, json.RawMessage(raw.ConfigJSON))
		if err != nil {
			errs = append(errs, fmt.Sprintf("%s(%s): parse: %v", raw.Name, raw.ID, err))
			continue
		}

		driver, err := NewDriver(scfg)
		if err != nil {
			errs = append(errs, fmt.Sprintf("%s(%s): build: %v", raw.Name, raw.ID, err))
			continue
		}

		newDrivers[raw.ID] = driver
		newMetas[raw.ID] = ConfigMeta{
			ID:                 raw.ID,
			Name:               raw.Name,
			Driver:             canonicalDriver,
			Provider:           provider,
			SchemeKey:          SchemeKeyForStoredConfig(canonicalDriver, &provider, raw.ConfigJSON),
			DriverConfig:       scfg,
			Priority:           raw.Priority,
			CapacityLimitBytes: raw.CapacityLimitBytes,
			IsDefault:          raw.IsDefault,
			IsActive:           raw.IsActive,
		}

		if raw.IsDefault {
			defaultID = raw.ID
		}
	}

	// No row is marked default but at least one is active: fall back to the lowest-priority
	// entry so uploads do not fail purely because the default flag is missing.
	if defaultID == "" && len(newMetas) > 0 {
		sorted := sortMetas(newMetas)
		defaultID = sorted[0].ID
	}

	m.mu.Lock()
	m.drivers = newDrivers
	m.metas = newMetas
	m.defaultID = defaultID
	m.mu.Unlock()

	if len(errs) > 0 {
		return fmt.Errorf("reload storage: %d errors: %v", len(errs), errs)
	}
	return nil
}

// Get returns the driver for the given configuration ID.
func (m *Manager) Get(configID string) (StorageDriver, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	driver, ok := m.drivers[configID]
	if !ok {
		return nil, fmt.Errorf("storage driver not found for config %q", configID)
	}
	return driver, nil
}

// Default returns the default storage driver and its configuration ID.
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

// DefaultID returns the current default configuration ID, or an empty string when none is set.
func (m *Manager) DefaultID() string {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.defaultID
}

// Meta returns a metadata snapshot for the given configuration ID.
func (m *Manager) Meta(configID string) (ConfigMeta, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	meta, ok := m.metas[configID]
	return meta, ok
}

// ActiveMetas returns metadata for every active configuration, sorted by priority ascending
// with a stable ID tiebreaker that mimics creation order.
func (m *Manager) ActiveMetas() []ConfigMeta {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return sortMetas(m.metas)
}

// List returns the IDs of every registered storage configuration.
func (m *Manager) List() []string {
	m.mu.RLock()
	defer m.mu.RUnlock()

	ids := make([]string, 0, len(m.drivers))
	for id := range m.drivers {
		ids = append(ids, id)
	}
	return ids
}

func sortMetas(metas map[string]ConfigMeta) []ConfigMeta {
	out := make([]ConfigMeta, 0, len(metas))
	for _, m := range metas {
		out = append(out, m)
	}
	sort.SliceStable(out, func(i, j int) bool {
		if out[i].Priority != out[j].Priority {
			return out[i].Priority < out[j].Priority
		}
		return out[i].ID < out[j].ID
	})
	return out
}
