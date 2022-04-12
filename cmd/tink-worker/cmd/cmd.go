package cmd

import (
	"context"
	"flag"
	"fmt"
	"os"
	"strings"
	"text/tabwriter"

	dockercli "github.com/docker/docker/client"
	"github.com/go-logr/logr"
	"github.com/go-logr/zerologr"
	"github.com/go-playground/validator/v10"
	"github.com/peterbourgon/ff/v3"
	"github.com/peterbourgon/ff/v3/ffcli"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
	"github.com/tinkerbell/tink/client"
	"github.com/tinkerbell/tink/cmd/tink-worker/worker"
	pb "github.com/tinkerbell/tink/protos/workflow"
	"google.golang.org/grpc"
)

// Command represents the tink worker environment variables/flags.
type Command struct {
	// Log is the logging implementation.
	Log logr.Logger
	// LogLevel defines the logging level.
	LogLevel string
	Registry string
	RegUser  string
	RegPass  string
	ID       string `validate:"required,uuid4"`
}

// Execute is an opinionated way to run the tink-worker.
// Flags are registered, cli/env vars are parsed, the Command struct is validated,
// and the tink worker is run.
func Execute(ctx context.Context, args []string) error {
	c := &Command{}
	fs := flag.NewFlagSet("tink-worker", flag.ExitOnError)
	cmd := newCLI(c, fs)
	return cmd.ParseAndRun(ctx, args)
}

func newCLI(c *Command, fs *flag.FlagSet) *ffcli.Command {
	c.RegisterFlags(fs)
	return &ffcli.Command{
		Name:       "tink-worker",
		ShortUsage: "Run Tink Worker",
		FlagSet:    fs,
		Options:    []ff.Option{ff.WithEnvVarPrefix("TINK_WORKER")},
		UsageFunc:  customUsageFunc,
		Exec: func(ctx context.Context, args []string) error {
			c.Log = defaultLogger(c.LogLevel)
			c.Log = c.Log.WithName("tink-worker")
			if err := c.Validate(); err != nil {
				return err
			}
			c.Log.Info("Starting tink worker")
			return c.Run(ctx)
		},
	}
}

// Validate checks the Command struct for validation errors.
func (c *Command) Validate() error {
	return validator.New().Struct(c)
}

// RegisterFlags registers a flag set for the tink worker command.
func (c *Command) RegisterFlags(f *flag.FlagSet) {
	f.StringVar(&c.Registry, "registry", "", "Container registry from which to pull images.")
	f.StringVar(&c.RegUser, "reg-user", "", "Container registry username.")
	f.StringVar(&c.RegPass, "reg-pass", "", "Container registry password.")
	f.StringVar(&c.ID, "id", "", "Worker ID.")
	f.StringVar(&c.LogLevel, "log-level", "info", "Logging level.")
}

// Run Tink worker.
func (c *Command) Run(ctx context.Context) error {
	if err := client.Setup(); err != nil {
		return err
	}

	conn, err := tryClientConnection()
	if err != nil {
		return err
	}
	rClient := pb.NewWorkflowServiceClient(conn)

	dockerClient, err := dockercli.NewClientWithOpts(dockercli.FromEnv, dockercli.WithAPIVersionNegotiation())
	if err != nil {
		return err
	}
	containerManager := worker.NewContainerManager(
		c.Log,
		dockerClient,
		worker.RegistryConnDetails{
			Registry: c.Registry,
			Username: c.RegUser,
			Password: c.RegPass,
		})

	logCapturer := worker.NewDockerLogCapturer(dockerClient, c.Log, os.Stdout)

	w := worker.NewWorker(
		c.ID,
		rClient,
		containerManager,
		logCapturer,
		c.Log,
	)

	err = w.ProcessWorkflowActions(ctx)
	if err != nil {
		return errors.Wrap(err, "worker Finished with error")
	}
	return nil
}

// defaultLogger is a zerolog logr implementation.
func defaultLogger(level string) logr.Logger {
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnixMs
	zerologr.NameFieldName = "logger"
	zerologr.NameSeparator = "/"

	zl := zerolog.New(os.Stdout)
	zl = zl.With().Caller().Timestamp().Logger()
	var l zerolog.Level
	switch level {
	case "debug":
		l = zerolog.DebugLevel
	default:
		l = zerolog.InfoLevel
	}
	zl = zl.Level(l)

	return zerologr.New(&zl)
}

func tryClientConnection() (*grpc.ClientConn, error) {
	c, err := client.GetConnection()
	if err != nil {
		return nil, err
	}

	return c, nil
}

// customUsageFunc is a custom UsageFunc (cli help message) used for all commands.
func customUsageFunc(c *ffcli.Command) string {
	var b strings.Builder

	fmt.Fprintf(&b, "USAGE\n")
	if c.ShortUsage != "" {
		fmt.Fprintf(&b, "  %s\n", c.ShortUsage)
	} else {
		fmt.Fprintf(&b, "  %s\n", c.Name)
	}
	fmt.Fprintf(&b, "\n")

	if c.LongHelp != "" {
		fmt.Fprintf(&b, "%s\n\n", c.LongHelp)
	}

	if len(c.Subcommands) > 0 {
		fmt.Fprintf(&b, "SUBCOMMANDS\n")
		tw := tabwriter.NewWriter(&b, 0, 2, 2, ' ', 0)
		for _, subcommand := range c.Subcommands {
			fmt.Fprintf(tw, "  %s\t%s\n", subcommand.Name, subcommand.ShortHelp)
		}
		tw.Flush()
		fmt.Fprintf(&b, "\n")
	}

	if countFlags(c.FlagSet) > 0 {
		fmt.Fprintf(&b, "FLAGS\n")
		tw := tabwriter.NewWriter(&b, 0, 2, 2, ' ', 0)
		c.FlagSet.VisitAll(func(f *flag.Flag) {
			format := "  -%s\t%s\n"
			values := []interface{}{f.Name, f.Usage}
			if def := f.DefValue; def != "" {
				format = "  -%s\t%s (default %q)\n"
				values = []interface{}{f.Name, f.Usage, def}
			}
			fmt.Fprintf(tw, format, values...)
		})
		_ = tw.Flush()
		fmt.Fprintf(&b, "\n")
	}

	return strings.TrimSpace(b.String()) + "\n"
}

func countFlags(fs *flag.FlagSet) (n int) {
	fs.VisitAll(func(*flag.Flag) { n++ })

	return n
}
