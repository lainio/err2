package try

// To is a helper method to call functions which returns (error)
// and check the error value. If error occurs it panics the error where err2
// handlers can catch it if needed.
func To(err error) {
	if err != nil {
		panic(err)
	}
}

// To1 is a helper method to call functions which returns (any, error)
// and check the error value. If error occurs it panics the error where err2
// handlers can catch it if needed.
func To1[T any](v T, err error) T {
	To(err)
	return v
}

// To2 is a helper method to call functions which returns (any, any, error)
// and check the error value. If error occurs it panics the error where err2
// handlers can catch it if needed.
func To2[T, U any](v1 T, v2 U, err error) (T, U) {
	To(err)
	return v1, v2
}

// To3 is a helper method to call functions which returns (any, any, any, error)
// and check the error value. If error occurs it panics the error where err2
// handlers can catch it if needed.
func To3[T, U, V any](v1 T, v2 U, v3 V, err error) (T, U, V) {
	To(err)
	return v1, v2, v3
}
