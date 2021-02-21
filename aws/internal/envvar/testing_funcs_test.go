package envvar_test

import (
	"os"
	"testing"

	testingiface "github.com/mitchellh/go-testing-interface"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/envvar"
)

func TestTestFailIfAllEmpty(t *testing.T) {
	envVar1 := "TESTENVVAR_FAILIFALLEMPTY1"
	envVar2 := "TESTENVVAR_FAILIFALLEMPTY2"
	envVars := []string{envVar1, envVar2}

	t.Run("missing", func(t *testing.T) {
		defer testingifaceRecover()

		for _, envVar := range envVars {
			os.Unsetenv(envVar)
		}

		envvar.TestFailIfAllEmpty(&testingiface.RuntimeT{}, envVars, "usage")

		t.Fatal("expected to fail previously")
	})

	t.Run("all empty", func(t *testing.T) {
		defer testingifaceRecover()

		os.Setenv(envVar1, "")
		os.Setenv(envVar2, "")
		defer unsetEnvVars(envVars)

		envvar.TestFailIfAllEmpty(&testingiface.RuntimeT{}, envVars, "usage")

		t.Fatal("expected to fail previously")
	})

	t.Run("some empty", func(t *testing.T) {
		wantValue := "pickme"

		os.Setenv(envVar1, "")
		os.Setenv(envVar2, wantValue)
		defer unsetEnvVars(envVars)

		gotName, gotValue := envvar.TestFailIfAllEmpty(&testingiface.RuntimeT{}, envVars, "usage")

		if gotName != envVar2 {
			t.Fatalf("expected name: %s, got: %s", envVar2, gotName)
		}

		if gotValue != wantValue {
			t.Fatalf("expected value: %s, got: %s", wantValue, gotValue)
		}
	})

	t.Run("all not empty", func(t *testing.T) {
		wantValue := "pickme"

		os.Setenv(envVar1, wantValue)
		os.Setenv(envVar2, "other")
		defer unsetEnvVars(envVars)

		gotName, gotValue := envvar.TestFailIfAllEmpty(&testingiface.RuntimeT{}, envVars, "usage")

		if gotName != envVar1 {
			t.Fatalf("expected name: %s, got: %s", envVar1, gotName)
		}

		if gotValue != wantValue {
			t.Fatalf("expected value: %s, got: %s", wantValue, gotValue)
		}
	})
}

func TestTestFailIfEmpty(t *testing.T) {
	envVar := "TESTENVVAR_FAILIFEMPTY"

	t.Run("missing", func(t *testing.T) {
		defer testingifaceRecover()

		os.Unsetenv(envVar)

		envvar.TestFailIfEmpty(&testingiface.RuntimeT{}, envVar, "usage")

		t.Fatal("expected to fail previously")
	})

	t.Run("empty", func(t *testing.T) {
		defer testingifaceRecover()

		os.Setenv(envVar, "")
		defer os.Unsetenv(envVar)

		envvar.TestFailIfEmpty(&testingiface.RuntimeT{}, envVar, "usage")

		t.Fatal("expected to fail previously")
	})

	t.Run("not empty", func(t *testing.T) {
		want := "notempty"

		os.Setenv(envVar, want)
		defer os.Unsetenv(envVar)

		got := envvar.TestFailIfEmpty(&testingiface.RuntimeT{}, envVar, "usage")

		if got != want {
			t.Fatalf("expected value: %s, got: %s", want, got)
		}
	})
}

func TestTestSkipIfEmpty(t *testing.T) {
	envVar := "TESTENVVAR_SKIPIFEMPTY"

	t.Run("missing", func(t *testing.T) {
		mockT := &testingiface.RuntimeT{}

		os.Unsetenv(envVar)

		envvar.TestSkipIfEmpty(mockT, envVar, "usage")

		if !mockT.Skipped() {
			t.Fatal("expected to skip previously")
		}
	})

	t.Run("empty", func(t *testing.T) {
		mockT := &testingiface.RuntimeT{}

		os.Setenv(envVar, "")
		defer os.Unsetenv(envVar)

		envvar.TestSkipIfEmpty(mockT, envVar, "usage")

		if !mockT.Skipped() {
			t.Fatal("expected to skip previously")
		}
	})

	t.Run("not empty", func(t *testing.T) {
		want := "notempty"

		os.Setenv(envVar, want)
		defer os.Unsetenv(envVar)

		got := envvar.TestSkipIfEmpty(&testingiface.RuntimeT{}, envVar, "usage")

		if got != want {
			t.Fatalf("expected value: %s, got: %s", want, got)
		}
	})
}

func testingifaceRecover() {
	r := recover()

	// this string is hardcoded in github.com/mitchellh/go-testing-interface
	if s, ok := r.(string); !ok || s != "testing.T failed, see logs for output (if any)" {
		panic(r)
	}
}

func unsetEnvVars(envVars []string) {
	for _, envVar := range envVars {
		os.Unsetenv(envVar)
	}
}
