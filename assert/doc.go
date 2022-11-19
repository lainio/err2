/*
Package assert includes runtime assertion helpers both for normal execution as
well as a assertion package for Go's testing. What makes solution unique is its
capable to support both modes with same API. Only thing you need to do is to
add following two lines at the beginning of your unit tests:

	func TestInvite(t *testing.T) {
	     assert.PushTester(t) // push testing variable t beginning of any test
	     defer assert.PopTester()

	     alice.Node = root1.Invite(alice.Node, root1.Key, alice.PubKey, 1)
	     assert.Equal(alice.Len(), 1) // assert anything normally

Especially powerful feature is that even if some assertion violation happens
during the execution of called functions like inside of the Invite() function
instead of the actual Test function, it's reported correctly as normal test
failure!

Instead of mocking or other mechanisms we can integrate our preconditions and
raise up quality of our software.

	"Assertsions are active comments"

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
test tool integration.
*/
package assert
