package utils

/*
TryDeref tries to dereference a pointer to type T. If it cannot,
then it returns nil.
*/
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

func NewPointer[T any](v T) *T {
	return &v
}
