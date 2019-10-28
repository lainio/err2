package err2

type _Bytes struct{}

// Bytes is a helper variable to generated
// 'type wrappers' to make Try function as fast as Check.
var Bytes _Bytes

// Try is a helper method to call func() ([]byte, error) functions
// with it and be as fast as Check(err).
func (o _Bytes) Try(v []byte, err error) []byte {
	Check(err)
	return v
}
