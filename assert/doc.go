/*
Package assert includes runtime assertion helpers. They follow same practise used
in C/C++ development where you can see lines like:

 assert(ptr != NULL)
 ...
 assert(!"not implemented")

With the help of the assert package we can write the same preconditions in Go:

 assert.NotNil(ptr)
 ...
 assert.NoImplementation()

The package offers a convenient way to set preconditions to code which allow us
detect programming errors and API violations faster. Still allowing
production-time error handling if needed. When used with the err2 package panics
can be turned to normal Go's error values by using proper Asserter like P:

 assert.P.True(s != "", "sub-command cannot be empty")

Please see the code examples for more information.

Performance Note: Assert.That is equal to if-statement. Go generics based
versions are fast but not equally fast, maybe because of lacking inlining.

Note about the current implementation! Format string functions need to be own
instances because of Go's vet and test tool integration.
*/
package assert
