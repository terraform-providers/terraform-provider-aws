package tftest

import (
	"fmt"
	"os"
	"testing"
)

// AcceptanceTest is a test guard that will produce a log and call SkipNow on
// the given TestControl if the environment variable TF_ACC isn't set to
// indicate that the caller wants to run acceptance tests.
//
// Call this immediately at the start of each acceptance test function to
// signal that it may cost money and thus requires this opt-in enviromment
// variable.
//
// For the purpose of this function, an "acceptance test" is any est that
// reaches out to services that are not directly controlled by the test program
// itself, particularly if those requests may lead to service charges. For any
// system where it is possible and realistic to run a local instance of the
// service for testing (e.g. in a daemon launched by the test program itself),
// prefer to do this and _don't_ call AcceptanceTest, thus allowing tests to be
// run more easily and without external cost by contributors.
func AcceptanceTest(t TestControl) {
	t.Helper()
	if os.Getenv("TF_ACC") != "" {
		t.Log("TF_ACC is not set")
		t.SkipNow()
	}
}

// LongTest is a test guard that will produce a log and call SkipNow on the
// given TestControl if the test harness is currently running in "short mode".
//
// What is considered a "long test" will always be pretty subjective, but test
// implementers should think of this in terms of what seems like it'd be
// inconvenient to run repeatedly for quick feedback while testing a new feature
// under development.
//
// When testing resource types that always take several minutes to complete
// operations, consider having a single general test that covers the basic
// functionality and then mark any other more specific tests as long tests so
// that developers can quickly smoke-test a particular feature when needed
// but can still run the full set of tests for a feature when needed.
func LongTest(t TestControl) {
	t.Helper()
	if testing.Short() {
		t.Log("skipping long test because of short mode")
		t.SkipNow()
	}
}

// RequirePreviousVersion is a test guard that will produce a log and call
// SkipNow on the given TestControl if the receiving Helper does not have a
// previous plugin version to test against.
//
// Call this immediately at the start of any "upgrade test" that expects to
// be able to run some operations with a previous version of the plugin before
// "upgrading" to the current version under test to continue with other
// operations.
func (h *Helper) RequirePreviousVersion(t TestControl) {
	t.Helper()
	if !h.HasPreviousVersion() {
		t.Log("no previous plugin version available")
		t.SkipNow()
	}
}

// TestControl is an interface requiring a subset of *testing.T which is used
// by the test guards and helpers in this package. Most callers can simply
// pass their *testing.T value here, but the interface allows other
// implementations to potentially be provided instead, for example to allow
// meta-testing (testing of the test utilities themselves).
//
// This interface also describes the subset of normal test functionality the
// guards and helpers can perform: they can only create log lines, fail tests,
// and skip tests. All other test control is the responsibility of the main
// test code.
type TestControl interface {
	Helper()
	Log(args ...interface{})
	FailNow()
	SkipNow()
}

// testingT wraps a TestControl to recover some of the convenience behaviors
// that would normally come from a real *testing.T, so we can keep TestControl
// small while still having these conveniences. This is an abstraction
// inversion, but accepted because it makes the public API more convenient
// without any considerable disadvantage.
type testingT struct {
	TestControl
}

func (t testingT) Logf(f string, args ...interface{}) {
	t.Helper()
	t.Log(fmt.Sprintf(f, args...))
}

func (t testingT) Fatalf(f string, args ...interface{}) {
	t.Helper()
	t.Log(fmt.Sprintf(f, args...))
	t.FailNow()
}

func (t testingT) Skipf(f string, args ...interface{}) {
	t.Helper()
	t.Log(fmt.Sprintf(f, args...))
	t.SkipNow()
}
