package err2

type _String struct{}

// String is a helper variable to generated
// 'type wrappers' to make Try function as fast as Check.
var String _String

// Try is a helper method to call func() (string, error) functions
// with it and be as fast as Check(err).
func (o _String) Try(v string, err error) string {
	Check(err)
	return v
}
