package try_test

import (
	"fmt"
	"os"
	"strconv"

	"github.com/lainio/err2"
	"github.com/lainio/err2/try"
)

func convTwoStr(s1, s2 string) (_ int, _ int, err error) {
	defer err2.Handle(&err)

	return try.To1(strconv.Atoi(s1)), try.To1(strconv.Atoi(s2)), nil
}

func ExampleResult2_Logf() {
	// Set log tracing to stdout that we can see it in Example output. In
	// normal cases that would be a Logging stream or stderr.
	err2.SetLogTracer(os.Stdout)

	countSomething := func(s1, s2 string) (int, int) {
		r := try.Out2(convTwoStr(s1, s2)).Logf().Def1(10).Def2(10)
		v1, v2 := r.Val1, r.Val2
		return v1 + v2, v2
	}
	_, _ = countSomething("1", "2")
	num1, num2 := countSomething("WRONG", "2")
	fmt.Printf("results: %d, %d", num1, num2)
	err2.SetLogTracer(nil)
	// Output: testing: run example: strconv.Atoi: parsing "WRONG": invalid syntax
	// results: 20, 10
}
