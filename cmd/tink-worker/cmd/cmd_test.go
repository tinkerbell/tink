package cmd

import (
	"fmt"
	"testing"
)

// Test Helpers

// setMinimumEnvVars is a test helper for setting up the minimum required env vars (unrelated to command line flag env vars).
func setMinimumEnvVars(t *testing.T) {
	t.Helper()
	t.Setenv("TINKERBELL_CERT_URL", "something1")
	t.Setenv("TINKERBELL_GRPC_AUTHORITY", "something2")
}

// setMinimumFlagEnvVars is a test helper for setting up the minimum command line flag env vars for tink-worker to run.
func setMinimumFlagEnvVars(t *testing.T, url string, user string, password string, id string) {
	t.Helper()
	t.Setenv("DOCKER_REGISTRY", url)
	t.Setenv("REGISTRY_USERNAME", user)
	t.Setenv("REGISTRY_PASSWORD", password)
	t.Setenv("ID", id)
}

// setCompleteFlagEnvVars is a test helper for setting up all env vars that tink-worker uses.
func setCompleteFlagEnvVars(t *testing.T, url string, user string, password string, id string, maxFileSize string, maxRetry string, retryInterval string, timeout string) {
	t.Helper()
	setMinimumFlagEnvVars(t, url, user, password, id)
	t.Setenv("MAX_FILE_SIZE", maxFileSize)
	t.Setenv("MAX_RETRY", maxRetry)
	t.Setenv("RETRY_INTERVAL", retryInterval)
	t.Setenv("TIMEOUT", timeout)
}

// assertIntValuesMatch is a test helper that triggers an error if two int values do not match.
func assertIntValuesMatch(t *testing.T, got, want int) {
	t.Helper()
	if got != want {
		t.Errorf("got %q want %q", got, want)
	}
}

// assertInt64ValuesMatch is a test helper that triggers an error if two int64 values do not match.
func assertInt64ValuesMatch(t *testing.T, got, want int64) {
	t.Helper()
	if got != want {
		t.Errorf("got %q want %q", got, want)
	}
}

// assertStringValuesMatch is a test helper that triggers an error if two strings do not match.
func assertStringValuesMatch(t *testing.T, got, want string) {
	t.Helper()
	if got != want {
		t.Errorf("got %q want %q", got, want)
	}
}

// assertErrorMessage is a test helper that triggers an error if an error object (err) doesn't use a specific string (msg).
func assertErrorMessage(t *testing.T, err error, msg string) {
	t.Helper()
	if err == nil {
		t.Errorf("Error: expecting error but none found")
	}
	assertStringValuesMatch(t, fmt.Sprint(err), msg)
}

// assertErrNotNil is a test helper that triggers an error if an error object (err) is not nil.
func assertErrNotNil(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Errorf("Error: %s", err)
	}
}

// CollectFlagEnvSettings() Functional Tests
// These tests validate the code paths through parseFlagEnvSettings() and validateFlagEnvSettings().

// TestRequiredLongCmdlineFlags ensures required long command line flags are really required.
func TestRequiredLongCmdlineFlags(t *testing.T) {
	t.Run("passing no flags returns an error", func(t *testing.T) {
		setMinimumEnvVars(t)
		args := []string{}
		_, err := CollectFlagEnvSettings(args)
		assertErrorMessage(t, err, "missing required flag --docker-registry <url> (or env var DOCKER_REGISTRY)")
	})

	t.Run("passing only --docker-registry localhost:8000 returns an error", func(t *testing.T) {
		setMinimumEnvVars(t)
		args := []string{"--docker-registry", "localhost:8000"}
		_, err := CollectFlagEnvSettings(args)
		assertErrorMessage(t, err, "missing required flag --registry-username <username> (or env var REGISTRY_USERNAME)")
	})

	t.Run("passing only --docker-registry localhost:8000 --registry-username user returns an error", func(t *testing.T) {
		setMinimumEnvVars(t)
		args := []string{"--docker-registry", "localhost:8000", "--registry-username", "user"}
		_, err := CollectFlagEnvSettings(args)
		assertErrorMessage(t, err, "missing required flag --registry-password <password> (or env var REGISTRY_PASSWORD)")
	})

	t.Run("passing only --docker-registry localhost:8000 --registry-username user --registry-password s3cret returns an error", func(t *testing.T) {
		setMinimumEnvVars(t)
		args := []string{"--docker-registry", "localhost:8000", "--registry-username", "user", "--registry-password", "s3cret"}
		_, err := CollectFlagEnvSettings(args)
		assertErrorMessage(t, err, "missing required flag --id <id> (or env var WORKER_ID)")
	})
}

// TestRequiredShortCmdLineFlags ensures required short command line flags are really required.
func TestRequiredShortCmdlineFlags(t *testing.T) {
	t.Run("passing only -r localhost:8000 returns an error", func(t *testing.T) {
		setMinimumEnvVars(t)
		args := []string{"-r", "localhost:8000"}
		_, err := CollectFlagEnvSettings(args)
		assertErrorMessage(t, err, "missing required flag --registry-username <username> (or env var REGISTRY_USERNAME)")
	})

	t.Run("passing only -r localhost:8000 -u user returns an error", func(t *testing.T) {
		setMinimumEnvVars(t)
		args := []string{"-r", "localhost:8000", "-u", "user"}
		_, err := CollectFlagEnvSettings(args)
		assertErrorMessage(t, err, "missing required flag --registry-password <password> (or env var REGISTRY_PASSWORD)")
	})

	t.Run("passing only -r localhost:8000 -u user -p s3cret returns an error", func(t *testing.T) {
		setMinimumEnvVars(t)
		args := []string{"-r", "localhost:8000", "-u", "user", "-p", "s3cret"}
		_, err := CollectFlagEnvSettings(args)
		assertErrorMessage(t, err, "missing required flag --id <id> (or env var WORKER_ID)")
	})
}

// TestMissingCmdlineArgsCauseErrors ensures missing arguments to command line flags trigger an error.
func TestMissingCmdlineArgsCauseErrors(t *testing.T) {
	t.Run("passing only --docker-registry returns an error", func(t *testing.T) {
		setMinimumEnvVars(t)
		args := []string{"--docker-registry"}
		_, err := CollectFlagEnvSettings(args)
		assertErrorMessage(t, err, "error parsing commandline args: flag needs an argument: -docker-registry")
	})
	t.Run("passing only --registry-username returns an error", func(t *testing.T) {
		setMinimumEnvVars(t)
		args := []string{"--registry-username"}
		_, err := CollectFlagEnvSettings(args)
		assertErrorMessage(t, err, "error parsing commandline args: flag needs an argument: -registry-username")
	})
	t.Run("passing only --registry-password returns an error", func(t *testing.T) {
		setMinimumEnvVars(t)
		args := []string{"--registry-password"}
		_, err := CollectFlagEnvSettings(args)
		assertErrorMessage(t, err, "error parsing commandline args: flag needs an argument: -registry-password")
	})
	t.Run("passing only --id returns an error", func(t *testing.T) {
		setMinimumEnvVars(t)
		args := []string{"--id"}
		_, err := CollectFlagEnvSettings(args)
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
		setMinimumEnvVars(t)
		args := minimumLongFlags
		flags, err := CollectFlagEnvSettings(args)
		assertErrNotNil(t, err)

		got := flags.DockerRegistry
		want := testURL
		assertStringValuesMatch(t, got, want)
	})

	t.Run("--registry-username flag is parsed correctly", func(t *testing.T) {
		setMinimumEnvVars(t)
		args := minimumLongFlags
		flags, err := CollectFlagEnvSettings(args)
		assertErrNotNil(t, err)

		got := flags.RegistryUsername
		want := testUser
		assertStringValuesMatch(t, got, want)
	})

	t.Run("--registry-password flag is parsed correctly", func(t *testing.T) {
		setMinimumEnvVars(t)
		args := minimumLongFlags
		flags, err := CollectFlagEnvSettings(args)
		assertErrNotNil(t, err)

		got := flags.RegistryPassword
		want := testPassword
		assertStringValuesMatch(t, got, want)
	})

	t.Run("--id flag is parsed correctly", func(t *testing.T) {
		setMinimumEnvVars(t)
		args := minimumLongFlags
		flags, err := CollectFlagEnvSettings(args)
		assertErrNotNil(t, err)

		got := flags.WorkerID
		want := testWorkerID
		assertStringValuesMatch(t, got, want)
	})

	t.Run("--max-file-size flag is parsed correctly", func(t *testing.T) {
		setMinimumEnvVars(t)
		args := completeLongFlags
		flags, err := CollectFlagEnvSettings(args)
		assertErrNotNil(t, err)

		got := fmt.Sprint(flags.MaxFileSize)
		want := testMaxFileSize
		assertStringValuesMatch(t, got, want)
	})

	t.Run("--max-retry flag is parsed correctly", func(t *testing.T) {
		setMinimumEnvVars(t)
		args := completeLongFlags
		flags, err := CollectFlagEnvSettings(args)
		assertErrNotNil(t, err)

		got := fmt.Sprint(flags.MaxRetry)
		want := testMaxRetry
		assertStringValuesMatch(t, got, want)
	})

	t.Run("--retry-interval flag is parsed correctly", func(t *testing.T) {
		setMinimumEnvVars(t)
		args := completeLongFlags
		flags, err := CollectFlagEnvSettings(args)
		assertErrNotNil(t, err)

		got := fmt.Sprint(flags.RetryInterval)
		want := testRetryInterval
		assertStringValuesMatch(t, got, want)
	})

	t.Run("--timeout flag is parsed correctly", func(t *testing.T) {
		setMinimumEnvVars(t)
		args := completeLongFlags
		flags, err := CollectFlagEnvSettings(args)
		assertErrNotNil(t, err)

		got := fmt.Sprint(flags.Timeout)
		want := testTimeout
		assertStringValuesMatch(t, got, want)
	})

	t.Run("--version flag is parsed correctly", func(t *testing.T) {
		setMinimumEnvVars(t)
		args := versionLongFlag
		_, err := CollectFlagEnvSettings(args)
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
		setMinimumEnvVars(t)
		args := minimumShortFlags
		flags, err := CollectFlagEnvSettings(args)
		assertErrNotNil(t, err)

		got := flags.DockerRegistry
		want := testURL
		assertStringValuesMatch(t, got, want)
	})

	t.Run("-u flag is parsed correctly", func(t *testing.T) {
		setMinimumEnvVars(t)
		args := minimumShortFlags
		flags, err := CollectFlagEnvSettings(args)
		assertErrNotNil(t, err)

		got := flags.RegistryUsername
		want := testUser
		assertStringValuesMatch(t, got, want)
	})

	t.Run("-p flag is parsed correctly", func(t *testing.T) {
		setMinimumEnvVars(t)
		args := minimumShortFlags
		flags, err := CollectFlagEnvSettings(args)
		assertErrNotNil(t, err)

		got := flags.RegistryPassword
		want := testPassword
		assertStringValuesMatch(t, got, want)
	})

	t.Run("-i flag is parsed correctly", func(t *testing.T) {
		setMinimumEnvVars(t)
		args := minimumShortFlags
		flags, err := CollectFlagEnvSettings(args)
		assertErrNotNil(t, err)

		got := flags.WorkerID
		want := testWorkerID
		assertStringValuesMatch(t, got, want)
	})

	t.Run("-v flag is parsed correctly", func(t *testing.T) {
		setMinimumEnvVars(t)
		args := versionShortFlag
		_, err := CollectFlagEnvSettings(args)
		assertErrNotNil(t, err)
	})
}

// TestDefaultValuesWithMinimumLongFlags ensures default values are used when passing the minimum required set of long flags.
func TestDefaultValuesWithMinimumLongFlags(t *testing.T) {
	testURL := "localhost:8000"
	testUser := "john.smith"
	testPassword := "super-s3cret"
	testWorkerID := "id-string"

	minimumLongFlags := []string{"--docker-registry", testURL, "--registry-username", testUser, "--registry-password", testPassword, "--id", testWorkerID}

	t.Run("default max file size is used", func(t *testing.T) {
		setMinimumEnvVars(t)
		args := minimumLongFlags
		flags, err := CollectFlagEnvSettings(args)
		assertErrNotNil(t, err)

		got := flags.MaxFileSize
		want := defaultMaxFileSize
		assertInt64ValuesMatch(t, got, want)
	})

	t.Run("default max retry is used", func(t *testing.T) {
		setMinimumEnvVars(t)
		args := minimumLongFlags
		flags, err := CollectFlagEnvSettings(args)
		assertErrNotNil(t, err)

		got := flags.MaxRetry
		want := defaultMaxRetry
		assertIntValuesMatch(t, got, want)
	})

	t.Run("default retry interval is used", func(t *testing.T) {
		setMinimumEnvVars(t)
		args := minimumLongFlags
		flags, err := CollectFlagEnvSettings(args)
		assertErrNotNil(t, err)

		got := flags.RetryInterval
		want := defaultRetryInterval
		assertIntValuesMatch(t, got, want)
	})

	t.Run("default timeout is used", func(t *testing.T) {
		setMinimumEnvVars(t)
		args := minimumLongFlags
		flags, err := CollectFlagEnvSettings(args)
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
		setMinimumEnvVars(t)
		args := minimumShortFlags
		flags, err := CollectFlagEnvSettings(args)
		assertErrNotNil(t, err)

		got := flags.MaxFileSize
		want := defaultMaxFileSize
		assertInt64ValuesMatch(t, got, want)
	})

	t.Run("default max retry is used", func(t *testing.T) {
		setMinimumEnvVars(t)
		args := minimumShortFlags
		flags, err := CollectFlagEnvSettings(args)
		assertErrNotNil(t, err)

		got := flags.MaxRetry
		want := defaultMaxRetry
		assertIntValuesMatch(t, got, want)
	})

	t.Run("default retry interval is used", func(t *testing.T) {
		setMinimumEnvVars(t)
		args := minimumShortFlags
		flags, err := CollectFlagEnvSettings(args)
		assertErrNotNil(t, err)

		got := flags.RetryInterval
		want := defaultRetryInterval
		assertIntValuesMatch(t, got, want)
	})

	t.Run("default timeout is used", func(t *testing.T) {
		setMinimumEnvVars(t)
		args := minimumShortFlags
		flags, err := CollectFlagEnvSettings(args)
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
		setMinimumEnvVars(t)
		setMinimumFlagEnvVars(t, testURL, testUser, testPassword, testWorkerID)
		args := []string{}
		flags, err := CollectFlagEnvSettings(args)
		assertErrNotNil(t, err)

		got := flags.DockerRegistry
		want := testURL
		assertStringValuesMatch(t, got, want)
	})

	t.Run("REGISTRY_USERNAME env var is used", func(t *testing.T) {
		setMinimumEnvVars(t)
		setMinimumFlagEnvVars(t, testURL, testUser, testPassword, testWorkerID)
		args := []string{}
		flags, err := CollectFlagEnvSettings(args)
		assertErrNotNil(t, err)

		got := flags.RegistryUsername
		want := testUser
		assertStringValuesMatch(t, got, want)
	})

	t.Run("REGISTRY_PASSWORD env var is used", func(t *testing.T) {
		setMinimumEnvVars(t)
		setMinimumFlagEnvVars(t, testURL, testUser, testPassword, testWorkerID)
		args := []string{}
		flags, err := CollectFlagEnvSettings(args)
		assertErrNotNil(t, err)

		got := flags.RegistryPassword
		want := testPassword
		assertStringValuesMatch(t, got, want)
	})

	t.Run("ID env var is used", func(t *testing.T) {
		setMinimumEnvVars(t)
		setMinimumFlagEnvVars(t, testURL, testUser, testPassword, testWorkerID)
		args := []string{}
		flags, err := CollectFlagEnvSettings(args)
		assertErrNotNil(t, err)

		got := flags.WorkerID
		want := testWorkerID
		assertStringValuesMatch(t, got, want)
	})

	t.Run("MAX_FILE_SIZE env var is used", func(t *testing.T) {
		setMinimumEnvVars(t)
		setCompleteFlagEnvVars(t, testURL, testUser, testPassword, testWorkerID, testMaxFileSize, testMaxRetry, testRetryInterval, testTimeout)
		args := []string{}
		flags, err := CollectFlagEnvSettings(args)
		assertErrNotNil(t, err)

		got := fmt.Sprint(flags.MaxFileSize)
		want := testMaxFileSize
		assertStringValuesMatch(t, got, want)
	})

	t.Run("MAX_RETRY env var is used", func(t *testing.T) {
		setMinimumEnvVars(t)
		setCompleteFlagEnvVars(t, testURL, testUser, testPassword, testWorkerID, testMaxFileSize, testMaxRetry, testRetryInterval, testTimeout)
		args := []string{}
		flags, err := CollectFlagEnvSettings(args)
		assertErrNotNil(t, err)

		got := fmt.Sprint(flags.MaxRetry)
		want := testMaxRetry
		assertStringValuesMatch(t, got, want)
	})

	t.Run("RETRY_INTERVAL env var is used", func(t *testing.T) {
		setMinimumEnvVars(t)
		setCompleteFlagEnvVars(t, testURL, testUser, testPassword, testWorkerID, testMaxFileSize, testMaxRetry, testRetryInterval, testTimeout)
		args := []string{}
		flags, err := CollectFlagEnvSettings(args)
		assertErrNotNil(t, err)

		got := fmt.Sprint(flags.RetryInterval)
		want := testRetryInterval
		assertStringValuesMatch(t, got, want)
	})

	t.Run("TIMEOUT env var is used", func(t *testing.T) {
		setMinimumEnvVars(t)
		setCompleteFlagEnvVars(t, testURL, testUser, testPassword, testWorkerID, testMaxFileSize, testMaxRetry, testRetryInterval, testTimeout)
		args := []string{}
		flags, err := CollectFlagEnvSettings(args)
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
		setMinimumEnvVars(t)
		setMinimumFlagEnvVars(t, testURLEnv, testUserEnv, testPasswordEnv, testWorkerIDEnv)
		args := minimumLongFlags
		flags, err := CollectFlagEnvSettings(args)
		assertErrNotNil(t, err)

		got := flags.DockerRegistry
		want := testURL
		assertStringValuesMatch(t, got, want)
	})

	t.Run("REGISTRY_USERNAME env var is overridden by --registry-username", func(t *testing.T) {
		setMinimumEnvVars(t)
		setMinimumFlagEnvVars(t, testURLEnv, testUserEnv, testPasswordEnv, testWorkerIDEnv)
		args := minimumLongFlags
		flags, err := CollectFlagEnvSettings(args)
		assertErrNotNil(t, err)

		got := flags.RegistryUsername
		want := testUser
		assertStringValuesMatch(t, got, want)
	})

	t.Run("REGISTRY_PASSWORD env var is overridden by --registry-password", func(t *testing.T) {
		setMinimumEnvVars(t)
		setMinimumFlagEnvVars(t, testURLEnv, testUserEnv, testPasswordEnv, testWorkerIDEnv)
		args := minimumLongFlags
		flags, err := CollectFlagEnvSettings(args)
		assertErrNotNil(t, err)

		got := flags.RegistryPassword
		want := testPassword
		assertStringValuesMatch(t, got, want)
	})

	t.Run("ID env var is overridden by --id", func(t *testing.T) {
		setMinimumEnvVars(t)
		setMinimumFlagEnvVars(t, testURLEnv, testUserEnv, testPasswordEnv, testWorkerIDEnv)
		args := minimumLongFlags
		flags, err := CollectFlagEnvSettings(args)
		assertErrNotNil(t, err)

		got := flags.WorkerID
		want := testWorkerID
		assertStringValuesMatch(t, got, want)
	})

	t.Run("MAX_FILE_SIZE env var is overridden by --max-file-size", func(t *testing.T) {
		setMinimumEnvVars(t)
		setCompleteFlagEnvVars(t, testURLEnv, testUserEnv, testPasswordEnv, testWorkerIDEnv, testMaxFileSizeEnv, testMaxRetryEnv, testRetryIntervalEnv, testTimeoutEnv)
		args := completeLongFlags
		flags, err := CollectFlagEnvSettings(args)
		assertErrNotNil(t, err)

		got := fmt.Sprint(flags.MaxFileSize)
		want := testMaxFileSize
		assertStringValuesMatch(t, got, want)
	})

	t.Run("MAX_RETRY env var is overridden by --max-retry", func(t *testing.T) {
		setMinimumEnvVars(t)
		setCompleteFlagEnvVars(t, testURLEnv, testUserEnv, testPasswordEnv, testWorkerIDEnv, testMaxFileSizeEnv, testMaxRetryEnv, testRetryIntervalEnv, testTimeoutEnv)
		args := completeLongFlags
		flags, err := CollectFlagEnvSettings(args)
		assertErrNotNil(t, err)

		got := fmt.Sprint(flags.MaxRetry)
		want := testMaxRetry
		assertStringValuesMatch(t, got, want)
	})

	t.Run("RETRY_INTERVAL env var is overridden by --retry-interval", func(t *testing.T) {
		setMinimumEnvVars(t)
		setCompleteFlagEnvVars(t, testURLEnv, testUserEnv, testPasswordEnv, testWorkerIDEnv, testMaxFileSizeEnv, testMaxRetryEnv, testRetryIntervalEnv, testTimeoutEnv)
		args := completeLongFlags
		flags, err := CollectFlagEnvSettings(args)
		assertErrNotNil(t, err)

		got := fmt.Sprint(flags.RetryInterval)
		want := testRetryInterval
		assertStringValuesMatch(t, got, want)
	})

	t.Run("TIMEOUT env var is overridden by --timeout", func(t *testing.T) {
		setMinimumEnvVars(t)
		setCompleteFlagEnvVars(t, testURLEnv, testUserEnv, testPasswordEnv, testWorkerIDEnv, testMaxFileSizeEnv, testMaxRetryEnv, testRetryIntervalEnv, testTimeoutEnv)
		args := completeLongFlags
		flags, err := CollectFlagEnvSettings(args)
		assertErrNotNil(t, err)

		got := fmt.Sprint(flags.Timeout)
		want := testTimeout
		assertStringValuesMatch(t, got, want)
	})
}

// TestRequiredEnvVars ensures required env vars (unrelated to command line flags) are really required.
func TestRequiredEnvVars(t *testing.T) {
	// Cmdline values
	testURL := "localhost:8000"
	testUser := "john.smith"
	testPassword := "super-s3cret"
	testWorkerID := "id-string"

	minimumLongFlags := []string{"--docker-registry", testURL, "--registry-username", testUser, "--registry-password", testPassword, "--id", testWorkerID}

	t.Run("TINKERBELL_CERT_URL env var is required", func(t *testing.T) {
		setMinimumEnvVars(t)
		t.Setenv("TINKERBELL_CERT_URL", "")
		_, err := CollectFlagEnvSettings(minimumLongFlags)
		assertErrorMessage(t, err, "required env var TINKERBELL_CERT_URL is undefined")
	})

	t.Run("TINKERBELL_GRPC_AUTHORITY env var is required", func(t *testing.T) {
		setMinimumEnvVars(t)
		t.Setenv("TINKERBELL_GRPC_AUTHORITY", "")
		_, err := CollectFlagEnvSettings(minimumLongFlags)
		assertErrorMessage(t, err, "required env var TINKERBELL_GRPC_AUTHORITY is undefined")
	})

	t.Run("no errors are reported when required env vars are set", func(t *testing.T) {
		setMinimumEnvVars(t)
		_, err := CollectFlagEnvSettings(minimumLongFlags)
		if err != nil {
			t.Errorf("Error: this test should not have generated an error. Error is %v", err)
		}
	})
}
