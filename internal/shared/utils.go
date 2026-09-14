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
