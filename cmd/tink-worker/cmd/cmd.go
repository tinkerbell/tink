package cmd

import (
	"errors"
	"flag"
	"fmt"

	"github.com/peterbourgon/ff"
)

// Default values used when cmdline flags/env vars are not specified.
const (
	defaultMaxFileSize    int64 = 10 * 1024 * 1024 // 10MB
	defaultMaxRetry             = 3
	defaultRetryInterval        = 3 // 3s
	defaultTimeoutMinutes       = 60
	programVersion              = "devel"
)

type CmdlineFlags struct {
	DockerRegistry   string
	RegistryUsername string
	RegistryPassword string
	WorkerID         string
	MaxFileSize      int64
	MaxRetry         int
	RetryInterval    int
	Timeout          int
	Version          bool
}

// Usage prints the help screen.
func Usage() {
	usageHelp := fmt.Sprintf(`Usage:
   tink-worker [flags]
Flags:
-r, --docker-registry string     Sets the Docker registry url (DOCKER_REGISTRY)
-h, --help                       help for tink-worker
-i, --id string                  Sets the worker id (ID)
    --max-file-size int          Maximum file size (bytes) (MAX_FILE_SIZE) (default %d)
    --max-retry int              Maximum number of retries to attempt (MAX_RETRY) (default %d)
-p, --registry-password string   Sets the registry-password (REGISTRY_PASSWORD)
-u, --registry-username string   Sets the registry username (REGISTRY_USERNAME)
    --retry-interval duration    Retry interval (seconds) (RETRY_INTERVAL) (default %ds)
    --timeout duration           Max duration (minutes) to wait for worker to complete (TIMEOUT) (default %dm)
-v, --version                    Print version string
`, defaultMaxFileSize, defaultMaxRetry, defaultRetryInterval, defaultTimeoutMinutes)

	fmt.Print(usageHelp)
}

// CollectCmdlineFlags parses and validates command line flags/env vars.
func CollectCmdlineFlags(args []string) (CmdlineFlags, error) {
	flags, err := parseCmdlineFlags(args)
	if err != nil {
		return flags, err
	}
	return flags, validateFlags(flags)
}

// parseCmdlineFlags parses flags from the command line and returns a populated CmdlineFlags struct.
func parseCmdlineFlags(args []string) (CmdlineFlags, error) {
	var flags CmdlineFlags

	fs := flag.NewFlagSet("tink-worker", flag.ContinueOnError)
	fs.Usage = Usage

	// flags that have long and short options
	fs.StringVar(&flags.DockerRegistry, "docker-registry", "", "")
	fs.StringVar(&flags.DockerRegistry, "r", "", "")
	fs.StringVar(&flags.RegistryUsername, "registry-username", "", "")
	fs.StringVar(&flags.RegistryUsername, "u", "", "")
	fs.StringVar(&flags.RegistryPassword, "registry-password", "", "")
	fs.StringVar(&flags.RegistryPassword, "p", "", "")
	fs.StringVar(&flags.WorkerID, "id", "", "")
	fs.StringVar(&flags.WorkerID, "i", "", "")
	fs.BoolVar(&flags.Version, "version", false, "")
	fs.BoolVar(&flags.Version, "v", false, "")

	// flags that are long-options only
	fs.Int64Var(&flags.MaxFileSize, "max-file-size", defaultMaxFileSize, "")
	fs.IntVar(&flags.MaxRetry, "max-retry", defaultMaxRetry, "")
	fs.IntVar(&flags.RetryInterval, "retry-interval", defaultRetryInterval, "")
	fs.IntVar(&flags.Timeout, "timeout", defaultTimeoutMinutes, "")

	err := ff.Parse(fs, args, ff.WithEnvVarNoPrefix())
	return flags, err
}

// validateFlags performs command line flag validation.
func validateFlags(flags CmdlineFlags) error {
	// specifying the version flag will print the version and ignore everything else
	if flags.Version {
		fmt.Println(programVersion)
		return nil
	}

	// check for required flags/env vars and immediately return an error when found
	if flags.DockerRegistry == "" {
		return errors.New("missing required flag --docker-registry <url> (or env var DOCKER_REGISTRY)")
	}
	if flags.RegistryUsername == "" {
		return errors.New("missing required flag --registry-username <username> (or env var REGISTRY_USERNAME)")
	}
	if flags.RegistryPassword == "" {
		return errors.New("missing required flag --registry-password <password> (or env var REGISTRY_PASSWORD)")
	}
	if flags.WorkerID == "" {
		return errors.New("missing required flag --id <id> (or env var WORKER_ID)")
	}

	return nil
}
