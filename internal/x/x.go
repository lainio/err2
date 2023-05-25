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
func SReverse[S ~[]E, E any](s S) S {
	length := len(s)
	for i := length/2 - 1; i >= 0; i-- {
		opp := length - 1 - i
		s[i], s[opp] = s[opp], s[i]
	}
	return s
}

// SSReverse reverse slice order. Doesn't clone slice. This is Fastest of these.
func SSReverse[S ~[]E, E any](s S) S {
	for begin, end := 0, len(s)-1; begin < end; begin, end = begin+1, end-1 {
		s[begin], s[end] = s[end], s[begin]
	}
	return s
}

// GetAndSet gets value to the *ptr and returns the old one, after setting new.
func GetAndSet[T any](ptr *T, val T) (old T) {
	old = *ptr
	*ptr = val
	return old
}

// Swap two values, which must be given as ptr. Returns new lhs, aka rhs.
func Swap[T any](lhs, rhs *T) (nlhs T) {
	swap := *lhs
	*lhs = *rhs
	*rhs = swap
	return *lhs
}

// TODO: , Max, ...
