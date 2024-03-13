/*
Package assert includes runtime assertion helpers both for normal execution as
well as a assertion package for Go's testing. What makes solution unique is its
capable to support both modes with same API. Only thing you need to do is to
add a PushTester line at the beginning of your unit tests:

	func TestInvite(t *testing.T) {
	     defer assert.PushTester(t)() // push testing variable t beginning of any test

		//                 v-----v Invite's control flow includes assertions
	     alice.Node = root1.Invite(alice.Node, root1.Key, alice.PubKey, 1)
	     assert.Equal(alice.Len(), 1) // assert anything normally

# Merge Runtime And Unit Test Assertions

Especially powerful feature is that even if some assertion violation happens
during the execution of called functions, and not the test function itself, they
are catched. See the above example. If assertion failure happens inside the
Invite() function instead of the actual test function, TestInvite, it's still
reported correctly as normal test failure. It doesn't matter how deep the
recursion is, or if parallel test runs are performed. It works just as you
hoped.

This is the actual Invite function implementation's first two lines. Even if the
assertion line is written more for runtime detection and active comment, it
catches all unit test errors as well:

	func (c Chain) Invite(...) {
		assert.That(c.isLeaf(invitersKey), "only leaf can invite")

# Call Stack Traversal During tests

The asserter package has super powerful feature. It allows us track assertion
violations over package and even module boundaries. When using err2 assert
package for runtime Asserts and assert violation happens in whatever package
and module, the whole call stack is brought to unit test logs. Naturally this is
optional. Only thing you need to do is set proper asserter and call PushTester.

	// use unit testing asserter
	assert.SetDefault(assert.TestFull)

With large multi repo environment this has proven to be valuable.

# Why Runtime Asserts Are So Important?

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
	assert.SetDefault(assert.Production)

Please see the code examples for more information.

# Flag Package Support

The assert package supports Go's flags. All you need to do is to call flag.Parse.
And the following flags are supported (="default-value"):

	-asserter="Prod"
	    A name of the asserter Plain, Prod, Dev, Debug
	    See more information from constants: Plain, Production, Development, Debug

And assert package's configuration flags are inserted.

# Performance

assert.That's performance is equal to the if-statement thanks for inlining. And
the most of the generics-based versions are about the equally fast. Practice has
thought that we should prefer other than assert.That because by using detailed
version like assert.Shorter we get precise error messages automatically. Some
also prefer readability of specific asserters.

If your algorithm is performance-critical please run `make bench` in the err2
repo and decide case by case. Also you can make an issue or even PR if you would
like to have something similar like glog.V() function.

# Naming

Because performance has been number one requirement and Go's generics cannot
discrete slices, maps and channels we have used naming prefixes accordingly: S =
slice, M = map, C = channel. No prefix is (currently) for the string type.
*/
package assert
