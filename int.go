package err2

type _Int struct{}

// Int is a helper variable to generated
// 'type wrappers' to make Try function as fast as Check.
// Note! Deprecated, use try package
var Int _Int

// Try is a helper method to call func() (int, error) functions
// with it and be as fast as Check(err).
func (o _Int) Try(v int, err error) int {
	Check(err)
	return v
}
