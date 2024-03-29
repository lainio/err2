/*
Package assert includes runtime assertion helpers both for normal execution as
well as a assertion package for Go's testing. What makes solution unique is its
capable to support both modes with the same API. Only thing you need to do is to
add a [PushTester] line at the beginning of your unit tests and its
sub-gouroutines:

	func TestInvite(t *testing.T) {
	     defer assert.PushTester(t)() // push testing variable t beginning of any test

	     //                 v-----v Invite's control flow includes assertions
	     alice.Node = root1.Invite(alice.Node, root1.Key, alice.PubKey, 1)
	     assert.Equal(alice.Len(), 1) // assert anything normally
	     ...
	     go func() {
	          assert.PushTester(t) // <-- Needs to do again for a new goroutine

# Merge Runtime And Unit Test Assertions

The next block is the actual Invite function's first two lines. Even if the
assertion line is written more from a runtime detection point of view, it catches
all assert violations in the unit tests as well:

	func (c Chain) Invite(...) {
	     assert.That(c.isLeaf(invitersKey), "only leaf can invite")

If some assertion violation occurs in the deep call stack, they are still
reported as a test failure. See the above code blocks. If assertion failure
happens somewhere inside the Invite function's call stack, it's still reported
correctly as a test failure of the TestInvite unit test. It doesn't matter how
deep the recursion is or if parallel test runs are performed. The failure report
includes all the locations of the meaningful call stack steps. See the next
chapter.

# Call Stack Traversal During Tests

The Assert package allows us to track assertion violations over the package and
even module boundaries. When an assertion fails during the unit testing, the
whole call stack is brought to unit test logs. And some help with your IDE, such
as transferring output to a location list, for example, in Neovim/Vim. For
example, you can find a proper test result parser like [nvim-go] (fork)

With a sizeable multi-repo environment, this has proven to be valuable.

# Why Runtime Asserts Are So Important?

Instead of mocking or other mechanisms we can integrate our preconditions and
raise up quality of our software.

	"Assertions are active comments"

The assert package offers a convenient way to set preconditions to code which
allow us detect programming errors and API violations faster. Still allowing
production-time error handling if needed. And everything is automatic. You can
set asserter with [SetDefault] function or --asserter flag if Go's flag package is
in use. This allows developer, operator and every-day user share the exact same
binary but get the error messages and diagnostic they need.

	// Production asserter adds formatted caller info to normal errors.
	// Information is transported thru error values when err2.Handle is in use.
	assert.SetDefault(assert.Production)

Please see the code examples for more information.

# Flag Package Support

The assert package supports Go's flags. All you need to do is to call
[flag.Parse]. And the following flags are supported (="default-value"):

	-asserter="Prod"
	    A name of the asserter Plain, Prod, Dev, Debug
	    See more information from constants: Plain, Production, Development, Debug

And assert package's configuration flags are inserted.

# Performance

[assert.That]'s performance is equal to the if-statement thanks for inlining. And
the most of the generics-based versions are about the equally fast. Practice has
thought that we should prefer other than [assert.That] because by using detailed
version like [assert.Shorter] we get precise error messages automatically. Some
also prefer readability of specific asserters.

If your algorithm is performance-critical please run `make bench` in the err2
repo and decide case by case. Also you can make an issue or even PR if you would
like to have something similar like [glog.V] function.

# Naming

Because performance has been number one requirement and Go's generics cannot
discrete slices, maps and channels we have used naming prefixes accordingly: S =
slice, M = map, C = channel. No prefix is (currently) for the string type.

[nvim-go]: https://github.com/lainio/nvim-go
*/
package assert
