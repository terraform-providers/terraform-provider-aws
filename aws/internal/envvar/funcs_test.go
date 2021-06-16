package envvar_test

import (
	"os"
	"testing"

	"github.com/terraform-providers/terraform-provider-aws/aws/internal/envvar"
)

func TestGetWithDefault(t *testing.T) {
	envVar := "TESTENVVAR_GETWITHDEFAULT"

	t.Run("missing", func(t *testing.T) {
		want := "default"

		os.Unsetenv(envVar)

		got := envvar.GetWithDefault(envVar, want)

		if got != want {
			t.Fatalf("expected %s, got %s", want, got)
		}
	})

	t.Run("empty", func(t *testing.T) {
		want := "default"

		os.Setenv(envVar, "")
		defer os.Unsetenv(envVar)

		got := envvar.GetWithDefault(envVar, want)

		if got != want {
			t.Fatalf("expected %s, got %s", want, got)
		}
	})

	t.Run("not empty", func(t *testing.T) {
		want := "notempty"

		os.Setenv(envVar, want)
		defer os.Unsetenv(envVar)

		got := envvar.GetWithDefault(envVar, "default")

		if got != want {
			t.Fatalf("expected %s, got %s", want, got)
		}
	})
}

func TestRequireOneOf(t *testing.T) {
	envVar1 := "TESTENVVAR_REQUIREONEOF1"
	envVar2 := "TESTENVVAR_REQUIREONEOF2"
	envVars := []string{envVar1, envVar2}

	t.Run("missing", func(t *testing.T) {
		for _, envVar := range envVars {
			os.Unsetenv(envVar)
		}

		_, _, err := envvar.RequireOneOf(envVars, "usage")

		if err == nil {
			t.Fatal("expected error")
		}
	})

	t.Run("all empty", func(t *testing.T) {
		os.Setenv(envVar1, "")
		os.Setenv(envVar2, "")
		defer unsetEnvVars(envVars)

		_, _, err := envvar.RequireOneOf(envVars, "usage")

		if err == nil {
			t.Fatal("expected error")
		}
	})

	t.Run("some empty", func(t *testing.T) {
		wantValue := "pickme"

		os.Setenv(envVar1, "")
		os.Setenv(envVar2, wantValue)
		defer unsetEnvVars(envVars)

		gotName, gotValue, err := envvar.RequireOneOf(envVars, "usage")

		if err != nil {
			t.Fatalf("unexpected error: %s", err)
		}

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

		gotName, gotValue, err := envvar.RequireOneOf(envVars, "usage")

		if err != nil {
			t.Fatalf("unexpected error: %s", err)
		}

		if gotName != envVar1 {
			t.Fatalf("expected name: %s, got: %s", envVar1, gotName)
		}

		if gotValue != wantValue {
			t.Fatalf("expected value: %s, got: %s", wantValue, gotValue)
		}
	})
}

func TestRequire(t *testing.T) {
	envVar := "TESTENVVAR_REQUIRE"

	t.Run("missing", func(t *testing.T) {
		os.Unsetenv(envVar)

		_, err := envvar.Require(envVar, "usage")

		if err == nil {
			t.Fatal("expected error")
		}
	})

	t.Run("empty", func(t *testing.T) {
		os.Setenv(envVar, "")
		defer os.Unsetenv(envVar)

		_, err := envvar.Require(envVar, "usage")

		if err == nil {
			t.Fatal("expected error")
		}
	})

	t.Run("not empty", func(t *testing.T) {
		want := "notempty"

		os.Setenv(envVar, want)
		defer os.Unsetenv(envVar)

		got, err := envvar.Require(envVar, "usage")

		if err != nil {
			t.Fatalf("unexpected error: %s", err)
		}

		if got != want {
			t.Fatalf("expected value: %s, got: %s", want, got)
		}
	})
}
