package lib

func TryDeref[T any](p *T) any /*nil|T*/ {
	if p == nil {
		return nil
	}
	return *p
}

func IsString(v any) bool {
	if _, ok := v.(string); ok {
		return true
	}
	return false
}

func IsInt(v any) bool {
	if _, ok := v.(int); ok {
		return true
	}
	return false
}
