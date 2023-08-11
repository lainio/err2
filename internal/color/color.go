package color

import "runtime"

const (
	reset  = "\033[0m"
	red    = "\033[31m"
	green  = "\033[32m"
	yellow = "\033[33m"
	blue   = "\033[34m"
	purple = "\033[35m"
	cyan   = "\033[36m"
	gray   = "\033[37m"
	white  = "\033[97m"
)

var (
	isWindows bool
)

func init() {
	isWindows = runtime.GOOS == "windows"
}

func Reset() string {
	if isWindows {
		return ""
	}
	return reset
}

func Red() string {
	if isWindows {
		return ""
	}
	return red
}

func Green() string {
	if isWindows {
		return ""
	}
	return green
}

func Yellow() string {
	if isWindows {
		return ""
	}
	return yellow
}

func Blue() string {
	if isWindows {
		return ""
	}
	return blue
}

func Purple() string {
	if isWindows {
		return ""
	}
	return purple
}

func Cyan() string {
	if isWindows {
		return ""
	}
	return cyan
}

func Gray() string {
	if isWindows {
		return ""
	}
	return gray
}

func White() string {
	if isWindows {
		return ""
	}
	return white
}
