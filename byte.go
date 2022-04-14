package err2

type _Byte struct{}

// Byte is a helper variable to generated
// 'type wrappers' to make Try function as fast as Check.
var Byte _Byte

// Try is a helper method to call func() (byte, error) functions
// with it and be as fast as Check(err).
func (o _Byte) Try(v byte, err error) byte {
	Check(err)
	return v
}
