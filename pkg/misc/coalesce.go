package misc

func Coalesce[T any](val *T, def T) T {
	if val != nil {
		return *val
	}
	return def
}
