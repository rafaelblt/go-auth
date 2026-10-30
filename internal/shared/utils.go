// Package shared holds generic helpers used across packages: the EntityID
// every identity wraps, a Set, pointer helpers, and ShortDuration.
//
// See docs/architecture/domain.md#shared.
package shared

func Ptr[T any](value T) *T {
	return &value
}

func ClonePtr[T any](p *T) *T {
	if p == nil {
		return nil
	}

	v := *p
	return &v
}

func PtrFromOk[T any](v T, ok bool) *T {
	if !ok {
		return nil
	}
	return &v
}

func Deref[T any](p *T) T {
	var zero T
	return DerefOr(p, zero)
}

func DerefOr[T any](p *T, fallback T) T {
	if p == nil {
		return fallback
	}
	return *p
}
