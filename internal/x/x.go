package x

// Whom is exactly same as C/C++ ternary operator. In Go it's implemented with
// generics.
func Whom[T any](b bool, yes, no T) T {
	if b {
		return yes
	}
	return no
}

func SReverse[T any](in []T) (out []T)  {
	length := len(in)
	out = make([]T, length)
	max := length - 1
	for i, v := range in {
		out[max-i] = v
	}
	return out
}

// GetAndSet gets value to the *ptr and returns the old one, after setting new.
func GetAndSet[T any](ptr *T, new T) (old T) {
	old = *ptr
	*ptr = new
	return old
}
// TODO: Swap, Max, ...
