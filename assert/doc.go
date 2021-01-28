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

The package offers a convenient way to set preconditions to code which allow us
detect programming errors and API usage violations faster. Still allowing proper
path to production-time error handling if needed. When used with the err2 package panics
can be turned to normal Go's error values by using proper Asserter like P:

 assert.True(a > b)

Please see the code examples for more information.

Note! Format string functions need to be own instances because of Go's vet and
test tool integration.
*/
package assert
