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

# Merge Runtime And Test Assertions

Especially powerful feature is that even if some assertion violation happens
during the execution of called functions not the test it self. See the above
example. If assertion failure happens inside of the Invite() function instead of
the actual Test function, TestInvite, it's still reported correctly as normal
test failure when TestInvite is executed. It doesn't matter how deep the
recursion is, or if parallel test runs are performed. It works just as you
hoped.

Instead of mocking or other mechanisms we can integrate our preconditions and
raise up quality of our software.

	"Assertions are active comments"

The package offers a convenient way to set preconditions to code which allow us
detect programming errors and API violations faster. Still allowing
production-time error handling if needed. And everything is automatic. You can
set proper asserter according to flag or environment variable. This allows
developer, operator and every-day user share the exact same binary but get the
error messages and diagnostic they need.

	// add formatted caller info for normal errors coming from assertions
	assert.SetDefaultAsserter(AsserterToError | AsserterFormattedCallerInfo)

Please see the code examples for more information.

Note. assert.That's performance has been (<go 1.20) equal to the if-statement.
Most of the generics-based versions are almost as fast. If your algorithm is
performance-critical please run `make bench` in the err2 repo and decide case by
case.

Note. Format string functions need to be own instances because of Go's vet and
test tool integration.
*/
package assert
