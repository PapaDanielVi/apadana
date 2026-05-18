package mt

// lazyIniter is satisfied by configs that can defer SDK initialization.
type lazyIniter interface {
	LazyInit() bool
}

// lazyInit returns true if v implements lazyIniter and its LazyInit returns true.
func lazyInit(v any) bool {
	li, ok := v.(lazyIniter)
	if ok {
		return li.LazyInit()
	}
	return true
}

// centralizedSDKMer is satisfied by configs for centralized SDKs.
type centralizedSDKMer interface {
	IsCentralized() bool
}

// isCentralized returns true if v implements centralizedSDKMer and its IsCentralized returns true.
func isCentralized(v any) bool {
	cs, ok := v.(centralizedSDKMer)
	if ok {
		return cs.IsCentralized()
	}
	return false
}
