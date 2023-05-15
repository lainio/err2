package x

// Whom is exactly same as C/C++ ternary operator. In Go it's implemented with
// generics.
func Whom[T any](b bool, yes, no T) T {
	if b {
		return yes
	}
	return no
}

// SReverse reverse slice order. Doesn't clone slice. This is Fastest of these.
func SSReverse[S ~[]E, E any](s S) S {
	for i := len(s)/2 - 1; i >= 0; i-- {
		opp := len(s) - 1 - i
		s[i], s[opp] = s[opp], s[i]
	}
	return s
}

// SReverse reverse slice order. Doesn't clone slice. This is Fastest of these.
func SReverse[S ~[]E, E any](s S) S {
	for begin, end := 0, len(s)-1; begin < end; begin, end = begin+1, end-1 {
		s[begin], s[end] = s[end], s[begin]
	}
	return s
}

func SReverseClone[T ~[]U, U any](in T) (out T) {
	out = make(T, len(in))
	//copy(out, in)
	//SReverse(out)
	for begin, end := 0, len(in)-1; begin < end; begin, end = begin+1, end-1 {
		out[begin] = out[end]
	}

	return out
}

func OSReverseClone[T ~[]U, U any](in T) (out T) {
	length := len(in)
	out = make(T, length)
	max := length - 1
	for i, v := range in {
		out[max-i] = v
	}
	return out
}

// GetAndSet gets value to the *ptr and returns the old one, after setting new.
func GetAndSet[T any](ptr *T, val T) (old T) {
	old = *ptr
	*ptr = val
	return old
}

// TODO: Swap, Max, ...
