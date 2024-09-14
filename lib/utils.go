package lib

func TryDeref[T any](p *T) any {
	if p == nil {
		return nil
	}
	return *p
}
