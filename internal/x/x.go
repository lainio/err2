package x

// Whom is exactly same as C/C++ ternary operator. In Go it's implemented with
// generics.
func Whom[T any](b bool, yes, no T) T {
	if b {
		return yes
	}
	return no
}
