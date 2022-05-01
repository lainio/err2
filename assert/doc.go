/*
Package assert includes runtime assertion helpers. They follow same practise used
in C/C++ development where you can see lines like:

 assert(ptr != NULL)
 ...
 assert(!"not implemented")

With the help of the assert package we can write the same preconditions in Go:

 assert.NotNil(ptr)
 ...
 assert.NotImplemented()

The package offers a convenient way to set preconditions to code which allow us
detect programming errors and API violations faster. Still allowing
production-time error handling if needed. When used with the err2 package panics
can be turned to normal Go's error values by using proper Asserter like P:

 assert.P.True(s != "", "sub-command cannot be empty")

Please see the code examples for more information.

Note. assert.That's performance is equal to the if-statement. Most of the
generics-based versions are as fast, but some of them (Equal, SLen, MLen)
aren't. If your algorithm is performance-critical please run `make bench` in the
err2 repo and decide case by case.

Note. Format string functions need to be own instances because of Go's vet and
test tool integration. */
package assert
