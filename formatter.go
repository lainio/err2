package err2

import (
	"github.com/lainio/err2/formatter"
	fmtstore "github.com/lainio/err2/internal/formatter"
)

func init() {
	SetFormatter(formatter.Decamel)
}

func SetFormatter(f formatter.Interface) {
	fmtstore.SetFormatter(f)
}

func Formatter() formatter.Interface {
	return fmtstore.Formatter()
}
