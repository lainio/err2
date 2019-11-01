package err2

type _StrStr struct{}

// StrStr is a helper variable to generated
// 'type wrappers' to make Try function as fast as Check.
var StrStr _StrStr

// Try is a helper method to call func() (string, error) functions
// with it and be as fast as Check(err).
func (o _StrStr) Try(v string, v2 string, err error) (string, string) {
	Check(err)
	return v, v2
}
