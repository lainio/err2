package err2

type _Bools struct{}

// Bools is a helper variable to generated
// 'type wrappers' to make Try function as fast as Check.
var Bools _Bools

// Try is a helper method to call func() ([]bool, error) functions
// with it and be as fast as Check(err).
func (o _Bools) Try(v []bool, err error) []bool {
	Check(err)
	return v
}
