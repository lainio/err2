package try

// To is a helper method to call functions which returns (any, error)
// and be as fast as Check(err).
func To(err error) {
	check(err)
}

// To1 is a helper method to call functions which returns (any, error)
// and be as fast as Check(err).
func To1[T any](v T, err error) T {
	check(err)
	return v
}

// To2 is a helper method to call functions which returns (any, error)
// and be as fast as Check(err).
func To2[T, U any](v1 T, v2 U, err error) (T, U) {
	check(err)
	return v1, v2
}

// To3 is a helper method to call functions which returns (any, error)
// and be as fast as Check(err).
func To3[T, U, V any](v1 T, v2 U, v3 V, err error) (T, U, V) {
	check(err)
	return v1, v2, v3
}

// check implements err nil check for this package. It's identical to in err2
// version, but it's needed here to avoid cyclic dependencies and help compiler
// optimization.
func check(err error) {
	if err != nil {
		panic(err)
	}
}
