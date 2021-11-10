package cmd

import (
	"fmt"
	"os"
	"testing"
)

// setMinimumEnvVars is a test helper for setting up the minimum env vars for tink-worker to run.
func setMinimumEnvVars(url string, user string, password string, id string) {
	os.Setenv("DOCKER_REGISTRY", url)
	os.Setenv("REGISTRY_USERNAME", user)
	os.Setenv("REGISTRY_PASSWORD", password)
	os.Setenv("ID", id)
}

// unsetMinimumEnvVars is a test helper for cleaning up/clearing env vars that may have been set by setMinimumEnvVars().
func unsetMinimumEnvVars() {
	os.Setenv("DOCKER_REGISTRY", "")
	os.Setenv("REGISTRY_USERNAME", "")
	os.Setenv("REGISTRY_PASSWORD", "")
	os.Setenv("ID", "")
}

// setCompleteEnvVars is a test helper for setting up all env vars that tink-worker uses.
func setCompleteEnvVars(url string, user string, password string, id string, maxFileSize string, maxRetry string, retryInterval string, timeout string) {
	setMinimumEnvVars(url, user, password, id)
	os.Setenv("MAX_FILE_SIZE", maxFileSize)
	os.Setenv("MAX_RETRY", maxRetry)
	os.Setenv("RETRY_INTERVAL", retryInterval)
	os.Setenv("TIMEOUT", timeout)
}

// unsetCompleteEnvVars is a test helper for cleaning up/clearing all env vars that tink-worker uses.
func unsetCompleteEnvVars() {
	unsetMinimumEnvVars()
	os.Setenv("MAX_FILE_SIZE", "")
	os.Setenv("MAX_RETRY", "")
	os.Setenv("RETRY_INTERVAL", "")
	os.Setenv("TIMEOUT", "")
}

// assertIntValuesMatch is a test helper that triggers an error if two int values do not match.
func assertIntValuesMatch(tb testing.TB, got, want int) {
	tb.Helper()
	if got != want {
		tb.Errorf("got %q want %q", got, want)
	}
}

// assertInt64ValuesMatch is a test helper that triggers an error if two int64 values do not match.
func assertInt64ValuesMatch(tb testing.TB, got, want int64) {
	tb.Helper()
	if got != want {
		tb.Errorf("got %q want %q", got, want)
	}
}

// assertStringValuesMatch is a test helper that triggers an error if two strings do not match.
func assertStringValuesMatch(tb testing.TB, got, want string) {
	tb.Helper()
	if got != want {
		tb.Errorf("got %q want %q", got, want)
	}
}

// assertErrorMessage is a test helper that triggers an error if an error object (err) doesn't use a specific string (msg).
func assertErrorMessage(tb testing.TB, err error, msg string) {
	tb.Helper()
	if err == nil {
		tb.Errorf("Error: expecting error but none found")
	}
	assertStringValuesMatch(tb, fmt.Sprint(err), msg)
}

// assertErrNotNil is a test helper that triggers an error if an error object (err) is not nil.
func assertErrNotNil(tb testing.TB, err error) {
	tb.Helper()
	if err != nil {
		tb.Errorf("Error: %s", err)
	}
}

// TestRequiredLongCmdlineFlags ensures required long command line flags are really required.
func TestRequiredLongCmdlineFlags(t *testing.T) {
	t.Run("passing no flags returns an error", func(t *testing.T) {
		args := []string{}
		_, err := CollectCmdlineFlags(args)
		assertErrorMessage(t, err, "missing required flag --docker-registry <url> (or env var DOCKER_REGISTRY)")
	})

	t.Run("passing only --docker-registry localhost:8000 returns an error", func(t *testing.T) {
		args := []string{"--docker-registry", "localhost:8000"}
		_, err := CollectCmdlineFlags(args)
		assertErrorMessage(t, err, "missing required flag --registry-username <username> (or env var REGISTRY_USERNAME)")
	})

	t.Run("passing only --docker-registry localhost:8000 --registry-username user returns an error", func(t *testing.T) {
		args := []string{"--docker-registry", "localhost:8000", "--registry-username", "user"}
		_, err := CollectCmdlineFlags(args)
		assertErrorMessage(t, err, "missing required flag --registry-password <password> (or env var REGISTRY_PASSWORD)")
	})

	t.Run("passing only --docker-registry localhost:8000 --registry-username user --registry-password s3cret returns an error", func(t *testing.T) {
		args := []string{"--docker-registry", "localhost:8000", "--registry-username", "user", "--registry-password", "s3cret"}
		_, err := CollectCmdlineFlags(args)
		assertErrorMessage(t, err, "missing required flag --id <id> (or env var WORKER_ID)")
	})
}

// TestRequiredShortCmdLineFlags ensures required short command line flags are really required.
func TestRequiredShortCmdlineFlags(t *testing.T) {
	t.Run("passing only -r localhost:8000 returns an error", func(t *testing.T) {
		args := []string{"-r", "localhost:8000"}
		_, err := CollectCmdlineFlags(args)
		assertErrorMessage(t, err, "missing required flag --registry-username <username> (or env var REGISTRY_USERNAME)")
	})

	t.Run("passing only -r localhost:8000 -u user returns an error", func(t *testing.T) {
		args := []string{"-r", "localhost:8000", "-u", "user"}
		_, err := CollectCmdlineFlags(args)
		assertErrorMessage(t, err, "missing required flag --registry-password <password> (or env var REGISTRY_PASSWORD)")
	})

	t.Run("passing only -r localhost:8000 -u user -p s3cret returns an error", func(t *testing.T) {
		args := []string{"-r", "localhost:8000", "-u", "user", "-p", "s3cret"}
		_, err := CollectCmdlineFlags(args)
		assertErrorMessage(t, err, "missing required flag --id <id> (or env var WORKER_ID)")
	})
}

// TestMissingCmdlineArgsCauseErrors ensures missing arguments to command line flags trigger an error.
func TestMissingCmdlineArgsCauseErrors(t *testing.T) {
	t.Run("passing only --docker-registry returns an error", func(t *testing.T) {
		args := []string{"--docker-registry"}
		_, err := CollectCmdlineFlags(args)
		assertErrorMessage(t, err, "error parsing commandline args: flag needs an argument: -docker-registry")
	})
	t.Run("passing only --registry-username returns an error", func(t *testing.T) {
		args := []string{"--registry-username"}
		_, err := CollectCmdlineFlags(args)
		assertErrorMessage(t, err, "error parsing commandline args: flag needs an argument: -registry-username")
	})
	t.Run("passing only --registry-password returns an error", func(t *testing.T) {
		args := []string{"--registry-password"}
		_, err := CollectCmdlineFlags(args)
		assertErrorMessage(t, err, "error parsing commandline args: flag needs an argument: -registry-password")
	})
	t.Run("passing only --id returns an error", func(t *testing.T) {
		args := []string{"--id"}
		_, err := CollectCmdlineFlags(args)
		assertErrorMessage(t, err, "error parsing commandline args: flag needs an argument: -id")
	})
}

// TestLongCmdlineFlags ensures long command line flags correctly populate the CmdlineFlags struct.
func TestLongCmdlineFlags(t *testing.T) {
	testURL := "localhost:8000"
	testUser := "john.smith"
	testPassword := "super-s3cret"
	testWorkerID := "id-string"
	testMaxFileSize := "1000"
	testMaxRetry := "5"
	testRetryInterval := "10"
	testTimeout := "20"

	minimumLongFlags := []string{"--docker-registry", testURL, "--registry-username", testUser, "--registry-password", testPassword, "--id", testWorkerID}
	completeLongFlags := []string{"--docker-registry", testURL, "--registry-username", testUser, "--registry-password", testPassword, "--id", testWorkerID, "--max-file-size", testMaxFileSize, "--max-retry", testMaxRetry, "--retry-interval", testRetryInterval, "--timeout", testTimeout}
	versionLongFlag := []string{"--version"}

	t.Run("--docker-registry flag is parsed correctly", func(t *testing.T) {
		args := minimumLongFlags
		flags, err := CollectCmdlineFlags(args)
		assertErrNotNil(t, err)

		got := flags.DockerRegistry
		want := testURL
		assertStringValuesMatch(t, got, want)
	})

	t.Run("--registry-username flag is parsed correctly", func(t *testing.T) {
		args := minimumLongFlags
		flags, err := CollectCmdlineFlags(args)
		assertErrNotNil(t, err)

		got := flags.RegistryUsername
		want := testUser
		assertStringValuesMatch(t, got, want)
	})

	t.Run("--registry-password flag is parsed correctly", func(t *testing.T) {
		args := minimumLongFlags
		flags, err := CollectCmdlineFlags(args)
		assertErrNotNil(t, err)

		got := flags.RegistryPassword
		want := testPassword
		assertStringValuesMatch(t, got, want)
	})

	t.Run("--id flag is parsed correctly", func(t *testing.T) {
		args := minimumLongFlags
		flags, err := CollectCmdlineFlags(args)
		assertErrNotNil(t, err)

		got := flags.WorkerID
		want := testWorkerID
		assertStringValuesMatch(t, got, want)
	})

	t.Run("--max-file-size flag is parsed correctly", func(t *testing.T) {
		args := completeLongFlags
		flags, err := CollectCmdlineFlags(args)
		assertErrNotNil(t, err)

		got := fmt.Sprint(flags.MaxFileSize)
		want := testMaxFileSize
		assertStringValuesMatch(t, got, want)
	})

	t.Run("--max-retry flag is parsed correctly", func(t *testing.T) {
		args := completeLongFlags
		flags, err := CollectCmdlineFlags(args)
		assertErrNotNil(t, err)

		got := fmt.Sprint(flags.MaxRetry)
		want := testMaxRetry
		assertStringValuesMatch(t, got, want)
	})

	t.Run("--retry-interval flag is parsed correctly", func(t *testing.T) {
		args := completeLongFlags
		flags, err := CollectCmdlineFlags(args)
		assertErrNotNil(t, err)

		got := fmt.Sprint(flags.RetryInterval)
		want := testRetryInterval
		assertStringValuesMatch(t, got, want)
	})

	t.Run("--timeout flag is parsed correctly", func(t *testing.T) {
		args := completeLongFlags
		flags, err := CollectCmdlineFlags(args)
		assertErrNotNil(t, err)

		got := fmt.Sprint(flags.Timeout)
		want := testTimeout
		assertStringValuesMatch(t, got, want)
	})

	t.Run("--version flag is parsed correctly", func(t *testing.T) {
		args := versionLongFlag
		_, err := CollectCmdlineFlags(args)
		assertErrNotNil(t, err)
	})
}

// TestShortCmdlineFlags ensures short command line flags correctly populate the CmdlineFlags struct.
func TestShortCmdlineFlags(t *testing.T) {
	testURL := "localhost:8000"
	testUser := "john.smith"
	testPassword := "super-s3cret"
	testWorkerID := "id-string"

	minimumShortFlags := []string{"-r", testURL, "-u", testUser, "-p", testPassword, "-i", testWorkerID}
	versionShortFlag := []string{"-v"}

	t.Run("-r flag is parsed correctly", func(t *testing.T) {
		args := minimumShortFlags
		flags, err := CollectCmdlineFlags(args)
		assertErrNotNil(t, err)

		got := flags.DockerRegistry
		want := testURL
		assertStringValuesMatch(t, got, want)
	})

	t.Run("-u flag is parsed correctly", func(t *testing.T) {
		args := minimumShortFlags
		flags, err := CollectCmdlineFlags(args)
		assertErrNotNil(t, err)

		got := flags.RegistryUsername
		want := testUser
		assertStringValuesMatch(t, got, want)
	})

	t.Run("-p flag is parsed correctly", func(t *testing.T) {
		args := minimumShortFlags
		flags, err := CollectCmdlineFlags(args)
		assertErrNotNil(t, err)

		got := flags.RegistryPassword
		want := testPassword
		assertStringValuesMatch(t, got, want)
	})

	t.Run("-i flag is parsed correctly", func(t *testing.T) {
		args := minimumShortFlags
		flags, err := CollectCmdlineFlags(args)
		assertErrNotNil(t, err)

		got := flags.WorkerID
		want := testWorkerID
		assertStringValuesMatch(t, got, want)
	})

	t.Run("-v flag is parsed correctly", func(t *testing.T) {
		args := versionShortFlag
		_, err := CollectCmdlineFlags(args)
		assertErrNotNil(t, err)
	})
}

// TestDefaultValuesWithMinimumLongFlags ensures default values are used when passing the minimum required set of long flags.
func TestDefaultValuesWithMinimumLongFlags(t *testing.T) {
	testURL := "localhost:8000"
	testUser := "john.smith"
	testPassword := "super-s3cret"
	testWorkerID := "id-string"
	// testMaxFileSize := "1000"
	// testMaxRetry := "5"
	// testRetryInterval := "10"
	// testTimeout := "20"

	minimumLongFlags := []string{"--docker-registry", testURL, "--registry-username", testUser, "--registry-password", testPassword, "--id", testWorkerID}
	// minimumShortFlags := []string{"-r", testURL, "-u", testUser, "-p", testPassword, "-i", testWorkerID}

	// long flags
	t.Run("default max file size is used", func(t *testing.T) {
		args := minimumLongFlags
		flags, err := CollectCmdlineFlags(args)
		assertErrNotNil(t, err)

		got := flags.MaxFileSize
		want := defaultMaxFileSize
		assertInt64ValuesMatch(t, got, want)
	})

	t.Run("default max retry is used", func(t *testing.T) {
		args := minimumLongFlags
		flags, err := CollectCmdlineFlags(args)
		assertErrNotNil(t, err)

		got := flags.MaxRetry
		want := defaultMaxRetry
		assertIntValuesMatch(t, got, want)
	})

	t.Run("default retry interval is used", func(t *testing.T) {
		args := minimumLongFlags
		flags, err := CollectCmdlineFlags(args)
		assertErrNotNil(t, err)

		got := flags.RetryInterval
		want := defaultRetryInterval
		assertIntValuesMatch(t, got, want)
	})

	t.Run("default timeout is used", func(t *testing.T) {
		args := minimumLongFlags
		flags, err := CollectCmdlineFlags(args)
		assertErrNotNil(t, err)

		got := flags.Timeout
		want := defaultTimeoutMinutes
		assertIntValuesMatch(t, got, want)
	})
}

// TestDefaultValuesWithMinimumShortFlags ensures default values are used when passing the minimum required set of short flags.
func TestDefaultValuesWithMinimumShortFlags(t *testing.T) {
	testURL := "localhost:8000"
	testUser := "john.smith"
	testPassword := "super-s3cret"
	testWorkerID := "id-string"

	minimumShortFlags := []string{"-r", testURL, "-u", testUser, "-p", testPassword, "-i", testWorkerID}

	t.Run("default max file size is used", func(t *testing.T) {
		args := minimumShortFlags
		flags, err := CollectCmdlineFlags(args)
		assertErrNotNil(t, err)

		got := flags.MaxFileSize
		want := defaultMaxFileSize
		assertInt64ValuesMatch(t, got, want)
	})

	t.Run("default max retry is used", func(t *testing.T) {
		args := minimumShortFlags
		flags, err := CollectCmdlineFlags(args)
		assertErrNotNil(t, err)

		got := flags.MaxRetry
		want := defaultMaxRetry
		assertIntValuesMatch(t, got, want)
	})

	t.Run("default retry interval is used", func(t *testing.T) {
		args := minimumShortFlags
		flags, err := CollectCmdlineFlags(args)
		assertErrNotNil(t, err)

		got := flags.RetryInterval
		want := defaultRetryInterval
		assertIntValuesMatch(t, got, want)
	})

	t.Run("default timeout is used", func(t *testing.T) {
		args := minimumShortFlags
		flags, err := CollectCmdlineFlags(args)
		assertErrNotNil(t, err)

		got := flags.Timeout
		want := defaultTimeoutMinutes
		assertIntValuesMatch(t, got, want)
	})
}

// TestEnvVars ensures env vars correctly populate the CmdlineFlags struct.
func TestEnvVars(t *testing.T) {
	testURL := "localhost:8000"
	testUser := "john.smith"
	testPassword := "super-s3cret"
	testWorkerID := "id-string"
	testMaxFileSize := "1000"
	testMaxRetry := "5"
	testRetryInterval := "10"
	testTimeout := "20"

	t.Run("DOCKER_REGISTRY env var is used", func(t *testing.T) {
		args := []string{}
		setMinimumEnvVars(testURL, testUser, testPassword, testWorkerID)
		flags, err := CollectCmdlineFlags(args)
		unsetMinimumEnvVars()
		assertErrNotNil(t, err)

		got := flags.DockerRegistry
		want := testURL
		assertStringValuesMatch(t, got, want)
	})

	t.Run("REGISTRY_USERNAME env var is used", func(t *testing.T) {
		args := []string{}
		setMinimumEnvVars(testURL, testUser, testPassword, testWorkerID)
		flags, err := CollectCmdlineFlags(args)
		unsetMinimumEnvVars()
		assertErrNotNil(t, err)

		got := flags.RegistryUsername
		want := testUser
		assertStringValuesMatch(t, got, want)
	})

	t.Run("REGISTRY_PASSWORD env var is used", func(t *testing.T) {
		args := []string{}
		setMinimumEnvVars(testURL, testUser, testPassword, testWorkerID)
		flags, err := CollectCmdlineFlags(args)
		unsetMinimumEnvVars()
		assertErrNotNil(t, err)

		got := flags.RegistryPassword
		want := testPassword
		assertStringValuesMatch(t, got, want)
	})

	t.Run("ID env var is used", func(t *testing.T) {
		args := []string{}
		setMinimumEnvVars(testURL, testUser, testPassword, testWorkerID)
		flags, err := CollectCmdlineFlags(args)
		unsetMinimumEnvVars()
		assertErrNotNil(t, err)

		got := flags.WorkerID
		want := testWorkerID
		assertStringValuesMatch(t, got, want)
	})

	t.Run("MAX_FILE_SIZE env var is used", func(t *testing.T) {
		args := []string{}
		setCompleteEnvVars(testURL, testUser, testPassword, testWorkerID, testMaxFileSize, testMaxRetry, testRetryInterval, testTimeout)
		flags, err := CollectCmdlineFlags(args)
		unsetCompleteEnvVars()
		assertErrNotNil(t, err)

		got := fmt.Sprint(flags.MaxFileSize)
		want := testMaxFileSize
		assertStringValuesMatch(t, got, want)
	})

	t.Run("MAX_RETRY env var is used", func(t *testing.T) {
		args := []string{}
		setCompleteEnvVars(testURL, testUser, testPassword, testWorkerID, testMaxFileSize, testMaxRetry, testRetryInterval, testTimeout)
		flags, err := CollectCmdlineFlags(args)
		unsetCompleteEnvVars()
		assertErrNotNil(t, err)

		got := fmt.Sprint(flags.MaxRetry)
		want := testMaxRetry
		assertStringValuesMatch(t, got, want)
	})

	t.Run("RETRY_INTERVAL env var is used", func(t *testing.T) {
		args := []string{}
		setCompleteEnvVars(testURL, testUser, testPassword, testWorkerID, testMaxFileSize, testMaxRetry, testRetryInterval, testTimeout)
		flags, err := CollectCmdlineFlags(args)
		unsetCompleteEnvVars()
		assertErrNotNil(t, err)

		got := fmt.Sprint(flags.RetryInterval)
		want := testRetryInterval
		assertStringValuesMatch(t, got, want)
	})

	t.Run("TIMEOUT env var is used", func(t *testing.T) {
		args := []string{}
		setCompleteEnvVars(testURL, testUser, testPassword, testWorkerID, testMaxFileSize, testMaxRetry, testRetryInterval, testTimeout)
		flags, err := CollectCmdlineFlags(args)
		unsetCompleteEnvVars()
		assertErrNotNil(t, err)

		got := fmt.Sprint(flags.Timeout)
		want := testTimeout
		assertStringValuesMatch(t, got, want)
	})
}

// TestCmdlineFlagsOverrideEnvVars ensures cmdline flags override env vars.
func TestCmdlineFlagsOverrideEnvVars(t *testing.T) {
	// Env values
	testURLEnv := "localhost:1234"
	testUserEnv := "jane.doe"
	testPasswordEnv := "env-password"
	testWorkerIDEnv := "env-id"
	testMaxFileSizeEnv := "5000"
	testMaxRetryEnv := "8"
	testRetryIntervalEnv := "22"
	testTimeoutEnv := "50"

	// Cmdline values
	testURL := "localhost:8000"
	testUser := "john.smith"
	testPassword := "super-s3cret"
	testWorkerID := "id-string"
	testMaxFileSize := "1000"
	testMaxRetry := "5"
	testRetryInterval := "10"
	testTimeout := "20"

	minimumLongFlags := []string{"--docker-registry", testURL, "--registry-username", testUser, "--registry-password", testPassword, "--id", testWorkerID}
	completeLongFlags := []string{"--docker-registry", testURL, "--registry-username", testUser, "--registry-password", testPassword, "--id", testWorkerID, "--max-file-size", testMaxFileSize, "--max-retry", testMaxRetry, "--retry-interval", testRetryInterval, "--timeout", testTimeout}

	t.Run("DOCKER_REGISTRY env var is overridden by --docker-registry", func(t *testing.T) {
		args := minimumLongFlags
		setMinimumEnvVars(testURLEnv, testUserEnv, testPasswordEnv, testWorkerIDEnv)
		flags, err := CollectCmdlineFlags(args)
		unsetMinimumEnvVars()
		assertErrNotNil(t, err)

		got := flags.DockerRegistry
		want := testURL
		assertStringValuesMatch(t, got, want)
	})

	t.Run("REGISTRY_USERNAME env var is overridden by --registry-username", func(t *testing.T) {
		args := minimumLongFlags
		setMinimumEnvVars(testURLEnv, testUserEnv, testPasswordEnv, testWorkerIDEnv)
		flags, err := CollectCmdlineFlags(args)
		unsetMinimumEnvVars()
		assertErrNotNil(t, err)

		got := flags.RegistryUsername
		want := testUser
		assertStringValuesMatch(t, got, want)
	})

	t.Run("REGISTRY_PASSWORD env var is overridden by --registry-password", func(t *testing.T) {
		args := minimumLongFlags
		setMinimumEnvVars(testURLEnv, testUserEnv, testPasswordEnv, testWorkerIDEnv)
		flags, err := CollectCmdlineFlags(args)
		unsetMinimumEnvVars()
		assertErrNotNil(t, err)

		got := flags.RegistryPassword
		want := testPassword
		assertStringValuesMatch(t, got, want)
	})

	t.Run("ID env var is overridden by --id", func(t *testing.T) {
		args := minimumLongFlags
		setMinimumEnvVars(testURLEnv, testUserEnv, testPasswordEnv, testWorkerIDEnv)
		flags, err := CollectCmdlineFlags(args)
		unsetMinimumEnvVars()
		assertErrNotNil(t, err)

		got := flags.WorkerID
		want := testWorkerID
		assertStringValuesMatch(t, got, want)
	})

	t.Run("MAX_FILE_SIZE env var is overridden by --max-file-size", func(t *testing.T) {
		args := completeLongFlags
		setCompleteEnvVars(testURLEnv, testUserEnv, testPasswordEnv, testWorkerIDEnv, testMaxFileSizeEnv, testMaxRetryEnv, testRetryIntervalEnv, testTimeoutEnv)
		flags, err := CollectCmdlineFlags(args)
		unsetCompleteEnvVars()
		assertErrNotNil(t, err)

		got := fmt.Sprint(flags.MaxFileSize)
		want := testMaxFileSize
		assertStringValuesMatch(t, got, want)
	})

	t.Run("MAX_RETRY env var is overridden by --max-retry", func(t *testing.T) {
		args := completeLongFlags
		setCompleteEnvVars(testURLEnv, testUserEnv, testPasswordEnv, testWorkerIDEnv, testMaxFileSizeEnv, testMaxRetryEnv, testRetryIntervalEnv, testTimeoutEnv)
		flags, err := CollectCmdlineFlags(args)
		unsetCompleteEnvVars()
		assertErrNotNil(t, err)

		got := fmt.Sprint(flags.MaxRetry)
		want := testMaxRetry
		assertStringValuesMatch(t, got, want)
	})

	t.Run("RETRY_INTERVAL env var is overridden by --retry-interval", func(t *testing.T) {
		args := completeLongFlags
		setCompleteEnvVars(testURLEnv, testUserEnv, testPasswordEnv, testWorkerIDEnv, testMaxFileSizeEnv, testMaxRetryEnv, testRetryIntervalEnv, testTimeoutEnv)
		flags, err := CollectCmdlineFlags(args)
		unsetCompleteEnvVars()
		assertErrNotNil(t, err)

		got := fmt.Sprint(flags.RetryInterval)
		want := testRetryInterval
		assertStringValuesMatch(t, got, want)
	})

	t.Run("TIMEOUT env var is overridden by --timeout", func(t *testing.T) {
		args := completeLongFlags
		setCompleteEnvVars(testURLEnv, testUserEnv, testPasswordEnv, testWorkerIDEnv, testMaxFileSizeEnv, testMaxRetryEnv, testRetryIntervalEnv, testTimeoutEnv)
		flags, err := CollectCmdlineFlags(args)
		unsetCompleteEnvVars()
		assertErrNotNil(t, err)

		got := fmt.Sprint(flags.Timeout)
		want := testTimeout
		assertStringValuesMatch(t, got, want)
	})
}
