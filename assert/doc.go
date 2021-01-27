/*
Package assert includes runtime assertion helpers. They follow common practise
in C/C++ development where you can see lines like:
 assert(ptr != NULL)
or
 assert(!"not implemented")
With the help of the assert package we can write the same preconditions in Go:
 assert.NotNil(ptr)
or
 assert.NoImplementation()

The assert package is meant to be used for design by contract -type of
development where you set preconditions for your functions. It's not meant to
replace normal error checking but speed up incremental hacking cycle. That's the
reason why default mode is to panic. By panicking developer get immediate
feedback which allows cleanup the code and APIs before actual production
release.

The package offers a convenient way to set preconditions to code which allow us
detect programming errors and API usage violations faster. Still allowing proper
path to production-time error handling. When used with the err2 package panics
can be turned to normal Go's error values by setting:

 assert.ProductionMode = true

Please see the code examples for more information.

Note! Format string functions need to be own instances because of Go's vet and
test tools.
*/
package assert
