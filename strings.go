package err2

type _Strings struct{}

// Strings is a helper variable to generated
// 'type wrappers' to make Try function as fast as Check.
var Strings _Strings

// Try is a helper method to call func() ([]string, error) functions
// with it and be as fast as Check(err).
func (o _Strings) Try(v []string, err error) []string {
	Check(err)
	return v
}
