package err2

type _Ints struct{}

// Ints is a helper variable to generated
// 'type wrappers' to make Try function as fast as Check.
var Ints _Ints

// Try is a helper method to call func() ([]int, error) functions
// with it and be as fast as Check(err).
func (o _Ints) Try(v []int, err error) []int {
	Check(err)
	return v
}
