package storage

import (
	"context"
	"fmt"
	"sync/atomic"
)

// Upload-policy constants. Stored in the settings table under the key `storage.upload_policy`.
const (
	PolicySingle          = "single"           // use only the default storage; failure is terminal
	PolicyPrimaryFallback = "primary_fallback" // walk storages by priority, falling back when full or on write error
	PolicyRoundRobin      = "round_robin"      // round-robin across all active storages
	PolicyMirror          = "mirror"           // write the primary synchronously and replicate to the rest in the background
)

// UsageFn returns the current used bytes for a storage configuration; the repo layer provides the implementation.
type UsageFn func(ctx context.Context, configID string) (int64, error)

// PolicyFn returns the current upload-policy string read from the settings table.
type PolicyFn func(ctx context.Context) (string, error)

// Router decides which storages receive an upload based on the Manager's driver cache and the runtime policy.
type Router struct {
	mgr      *Manager
	usageFn  UsageFn
	policyFn PolicyFn
	rrCursor atomic.Uint64
}

// NewRouter builds a Router.
// usageFn drives capacity filtering and policyFn reads the upload policy; both are required.
func NewRouter(mgr *Manager, usageFn UsageFn, policyFn PolicyFn) *Router {
	return &Router{mgr: mgr, usageFn: usageFn, policyFn: policyFn}
}

// Target is a single upload target consisting of metadata and its driver instance.
type Target struct {
	Meta   ConfigMeta
	Driver StorageDriver
}

// Plan is the routing plan for a single upload.
// Mode dictates how callers consume Targets:
//   - single / round_robin: write to Targets[0]; failure is terminal.
//   - primary_fallback: try Targets[0..n] in order; the first success wins.
//   - mirror: Targets[0] is the primary and must succeed; Targets[1..] are replicas written concurrently in the background.
type Plan struct {
	Mode    string
	Targets []Target
}

// Plan picks targets for a size-byte upload according to the current policy, priority, and capacity.
// Returns an error when no active storage has enough remaining capacity.
func (r *Router) Plan(ctx context.Context, size int64) (*Plan, error) {
	mode, err := r.policyFn(ctx)
	if err != nil || mode == "" {
		mode = PolicySingle
	}

	metas := r.mgr.ActiveMetas()
	if len(metas) == 0 {
		return nil, fmt.Errorf("no active storage configured")
	}

	switch mode {
	case PolicyMirror:
		// Every active storage is a target; the primary is the first eligible one and the rest follow priority order.
		all := r.buildTargets(metas)
		eligible := r.filterByCapacity(ctx, all, size)
		if len(eligible) == 0 {
			return nil, fmt.Errorf("no storage has enough capacity for %d bytes", size)
		}
		// Guarantee the primary (Targets[0]) has capacity; replicas include the remaining eligible storages.
		primary := eligible[0]
		var replicas []Target
		for _, t := range eligible[1:] {
			replicas = append(replicas, t)
		}
		return &Plan{Mode: PolicyMirror, Targets: append([]Target{primary}, replicas...)}, nil

	case PolicyPrimaryFallback:
		all := r.buildTargets(metas)
		eligible := r.filterByCapacity(ctx, all, size)
		if len(eligible) == 0 {
			return nil, fmt.Errorf("no storage has enough capacity for %d bytes", size)
		}
		return &Plan{Mode: PolicyPrimaryFallback, Targets: eligible}, nil

	case PolicyRoundRobin:
		all := r.buildTargets(metas)
		eligible := r.filterByCapacity(ctx, all, size)
		if len(eligible) == 0 {
			return nil, fmt.Errorf("no storage has enough capacity for %d bytes", size)
		}
		idx := int(r.rrCursor.Add(1)-1) % len(eligible)
		return &Plan{Mode: PolicyRoundRobin, Targets: []Target{eligible[idx]}}, nil

	default: // PolicySingle
		defID := r.mgr.DefaultID()
		if defID == "" {
			return nil, fmt.Errorf("no default storage configured")
		}
		driver, err := r.mgr.Get(defID)
		if err != nil {
			return nil, err
		}
		meta, _ := r.mgr.Meta(defID)
		// The single policy still honours the capacity cap: exceeding it fails fast rather than silently
		// switching to another storage, avoiding unpredictable behaviour.
		if meta.CapacityLimitBytes > 0 {
			used, _ := r.usageFn(ctx, defID)
			if used+size > meta.CapacityLimitBytes {
				return nil, fmt.Errorf("default storage %q is full", meta.Name)
			}
		}
		return &Plan{Mode: PolicySingle, Targets: []Target{{Meta: meta, Driver: driver}}}, nil
	}
}

// buildTargets turns metas into (meta, driver) pairs; the input order already reflects priority sorting.
func (r *Router) buildTargets(metas []ConfigMeta) []Target {
	out := make([]Target, 0, len(metas))
	for _, m := range metas {
		driver, err := r.mgr.Get(m.ID)
		if err != nil {
			continue
		}
		out = append(out, Target{Meta: m, Driver: driver})
	}
	return out
}

// filterByCapacity drops storages that are full; CapacityLimitBytes==0 means unlimited.
// A usage-query failure is treated as "allow" so that transient errors do not block uploads.
func (r *Router) filterByCapacity(ctx context.Context, targets []Target, size int64) []Target {
	out := make([]Target, 0, len(targets))
	for _, t := range targets {
		if t.Meta.CapacityLimitBytes <= 0 {
			out = append(out, t)
			continue
		}
		used, err := r.usageFn(ctx, t.Meta.ID)
		if err != nil {
			out = append(out, t)
			continue
		}
		if used+size <= t.Meta.CapacityLimitBytes {
			out = append(out, t)
		}
	}
	return out
}
