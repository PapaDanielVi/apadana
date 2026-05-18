package mt

import (
	"context"
	"fmt"
	"maps"
	"sync"
)

type (
	InitFn[S any, C any]               func(context.Context, C) S
	InitFnE[S any, C any]              func(context.Context, C) (S, error)
	InitFnWMet[S any, C any, M any]    func(context.Context, C, M) S
)

// ISDK is the interface for SDK managers returning T.
type ISDK[T any] interface {
	Get(ctx context.Context) T
	Map(func(string, T))
}

// SDKMgr manages per-tenant SDK instances.
type SDKMgr[S any, C any] struct {
	configs       map[string]C
	sdks          map[string]S
	mu            sync.RWMutex
	initFn        InitFn[S, C]
	isCentralized bool
}

func (m *SDKMgr[S, C]) Map(f func(string, S)) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	for tenant, sdk := range m.sdks {
		f(tenant, sdk)
	}
}

// NewSDKMgr creates a new SDK manager.
// configs maps tenant ID to SDK config.
// initFn is the factory function to create SDK instances.
func NewSDKMgr[S any, C any](configs map[string]C, initFn InitFn[S, C]) ISDK[S] {
	copied := make(map[string]C, len(configs))
	maps.Copy(copied, configs)

	sm := &SDKMgr[S, C]{
		configs: copied,
		sdks:    make(map[string]S, len(configs)),
		initFn:  initFn,
	}

	defOpt, ok := copied[fallbackTID]
	if ok && isCentralized(defOpt) {
		_ = sm.Get(InjectTID(context.Background(), fallbackTID))
		sm.isCentralized = true
		return sm
	}

	for tenant, opt := range configs {
		if !lazyInit(opt) {
			_ = sm.Get(InjectTID(context.Background(), tenant))
		}
	}
	return sm
}

func (s *SDKMgr[S, C]) Get(ctx context.Context) S {
	if s.isCentralized {
		return s.sdks[fallbackTID]
	}
	tid := ExtractTID(ctx)
	s.mu.RLock()
	sdk, exists := s.sdks[tid]
	s.mu.RUnlock()
	if exists {
		return sdk
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	cfg := s.configs[tid]
	newSDK := s.initFn(ctx, cfg)
	s.sdks[tid] = newSDK
	return newSDK
}

// ISDKE is the interface for SDK managers returning (T, error).
type ISDKE[T any] interface {
	Get(ctx context.Context) (T, error)
	IsCentralized() bool
}

// SDKMgrE manages per-tenant SDK instances with error handling.
type SDKMgrE[S any, C any] struct {
	configs       map[string]C
	sdks          map[string]S
	mu            sync.RWMutex
	initFn        InitFnE[S, C]
	isCentralized bool
}

func (m *SDKMgrE[S, C]) Map(f func(string, S)) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	for tenant, sdk := range m.sdks {
		f(tenant, sdk)
	}
}

// NewSDKMgrE creates a new SDK manager with error handling.
func NewSDKMgrE[S any, C any](configs map[string]C, initFn InitFnE[S, C]) (ISDKE[S], error) {
	copied := make(map[string]C, len(configs))
	maps.Copy(copied, configs)

	sm := &SDKMgrE[S, C]{
		configs: copied,
		sdks:    make(map[string]S, len(configs)),
		initFn:  initFn,
	}

	defOpt, ok := copied[fallbackTID]
	if ok && isCentralized(defOpt) {
		_, err := sm.Get(InjectTID(context.Background(), fallbackTID))
		if err != nil {
			return nil, nil
		}
		sm.isCentralized = true
		return sm, nil
	}

	for tenant, opt := range configs {
		if !lazyInit(opt) {
			_, err := sm.Get(InjectTID(context.Background(), tenant))
			if err != nil {
				return nil, fmt.Errorf("init sdk for %s: %w", tenant, err)
			}
		}
	}
	return sm, nil
}

func (s *SDKMgrE[S, C]) IsCentralized() bool {
	return s.isCentralized
}

func (s *SDKMgrE[S, C]) Get(ctx context.Context) (S, error) {
	if s.isCentralized {
		return s.sdks[fallbackTID], nil
	}
	tid := ExtractTID(ctx)
	s.mu.RLock()
	sdk, exists := s.sdks[tid]
	s.mu.RUnlock()
	if exists {
		return sdk, nil
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	cfg := s.configs[tid]
	newSDK, err := s.initFn(ctx, cfg)
	if err != nil {
		return newSDK, err
	}
	s.sdks[tid] = newSDK
	return newSDK, nil
}

// SDKMgrWMet manages per-tenant SDK instances with metrics.
type SDKMgrWMet[S any, C any, M any] struct {
	configs       map[string]C
	sdks          map[string]S
	metrics       M
	mu            sync.RWMutex
	initFn        InitFnWMet[S, C, M]
	isCentralized bool
}

func (m *SDKMgrWMet[S, C, M]) Map(f func(string, S)) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	for tenant, sdk := range m.sdks {
		f(tenant, sdk)
	}
}

// NewSDKMgrWMet creates a new SDK manager with metrics support.
func NewSDKMgrWMet[S any, C any, M any](configs map[string]C, initFn InitFnWMet[S, C, M], metrics M) ISDK[S] {
	copied := make(map[string]C, len(configs))
	maps.Copy(copied, configs)

	sm := &SDKMgrWMet[S, C, M]{
		configs: copied,
		sdks:    make(map[string]S, len(configs)),
		initFn:  initFn,
		metrics: metrics,
	}

	defOpt, ok := copied[fallbackTID]
	if ok && isCentralized(defOpt) {
		_ = sm.Get(InjectTID(context.Background(), fallbackTID))
		sm.isCentralized = true
		return sm
	}

	for tenant, opt := range configs {
		if !lazyInit(opt) {
			_ = sm.Get(InjectTID(context.Background(), tenant))
		}
	}
	return sm
}

func (s *SDKMgrWMet[S, C, M]) Get(ctx context.Context) S {
	if s.isCentralized {
		return s.sdks[fallbackTID]
	}
	tid := ExtractTID(ctx)
	s.mu.RLock()
	sdk, exists := s.sdks[tid]
	s.mu.RUnlock()
	if exists {
		return sdk
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	cfg := s.configs[tid]
	newSDK := s.initFn(ctx, cfg, s.metrics)
	s.sdks[tid] = newSDK
	return newSDK
}
