package err2

// Try1 is a helper method to call functions which returns (any, error)
// and be as fast as Check(err).
func Try1[T any](v T, err error) T {
	Check(err)
	return v
}

// Try2 is a helper method to call functions which returns (any, error)
// and be as fast as Check(err).
func Try2[T1, T2 any](v1 T1, v2 T2, err error) (T1, T2) {
	Check(err)
	return v1, v2
}

// Try3 is a helper method to call functions which returns (any, error)
// and be as fast as Check(err).
func Try3[T1, T2, T3 any](v1 T1, v2 T2, v3 T3, err error) (T1, T2, T3) {
	Check(err)
	return v1, v2, v3
}

