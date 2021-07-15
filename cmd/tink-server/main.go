package main

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"

	"github.com/packethost/pkg/log"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"github.com/tinkerbell/tink/db"
	rpcServer "github.com/tinkerbell/tink/grpc-server"
	httpServer "github.com/tinkerbell/tink/http-server"
	"github.com/tobert/otel-launcher-go/launcher"
)

var (
	// version is set at build time
	version = "devel"
)

// DaemonConfig represents all the values you can configure as part of the tink-server.
// You can change the configuration via environment variable, or file, or command flags.
type DaemonConfig struct {
	Facility              string
	PGDatabase            string
	PGUSer                string
	PGPassword            string
	PGSSLMode             string
	OnlyMigration         bool
	GRPCAuthority         string
	TLSCert               string
	CertDir               string
	HTTPAuthority         string
	HTTPBasicAuthUsername string
	HTTPBasicAuthPassword string
}

func (c *DaemonConfig) AddFlags(fs *pflag.FlagSet) {
	fs.StringVar(&c.Facility, "facility", "deprecated", "This is temporary. It will be removed")
	fs.StringVar(&c.PGDatabase, "postgres-database", "tinkerbell", "The Postgres database name")
	fs.StringVar(&c.PGUSer, "postgres-user", "tinkerbell", "The Postgres database username")
	fs.StringVar(&c.PGPassword, "postgres-password", "tinkerbell", "The Postgres database password")
	fs.StringVar(&c.PGSSLMode, "postgres-sslmode", "disable", "Enable or disable SSL mode in postgres")
	fs.BoolVar(&c.OnlyMigration, "only-migration", false, "When enabled the server applies the migration to postgres database and it exits")
	fs.StringVar(&c.GRPCAuthority, "grpc-authority", ":42113", "The address used to expose the gRPC server")
	fs.StringVar(&c.TLSCert, "tls-cert", "", "")
	fs.StringVar(&c.CertDir, "cert-dir", "", "")
	fs.StringVar(&c.HTTPAuthority, "http-authority", ":42114", "The address used to expose the HTTP server")
}

func (c *DaemonConfig) PopulateFromLegacyEnvVar() {
	if f := os.Getenv("FACILITY"); f != "" {
		c.Facility = f
	}
	if pgdb := os.Getenv("PGDATABASE"); pgdb != "" {
		c.PGDatabase = pgdb
	}
	if pguser := os.Getenv("PGUSER"); pguser != "" {
		c.PGUSer = pguser
	}
	if pgpass := os.Getenv("PGPASSWORD"); pgpass != "" {
		c.PGPassword = pgpass
	}
	if pgssl := os.Getenv("PGSSLMODE"); pgssl != "" {
		c.PGSSLMode = pgssl
	}
	if onlyMigration, isSet := os.LookupEnv("ONLY_MIGRATION"); isSet {
		if b, err := strconv.ParseBool(onlyMigration); err != nil {
			c.OnlyMigration = b
		}
	}
	if tlsCert := os.Getenv("TINKERBELL_TLS_CERT"); tlsCert != "" {
		c.TLSCert = tlsCert
	}
	if certDir := os.Getenv("TINKERBELL_CERTS_DIR"); certDir != "" {
		c.CertDir = certDir
	}
	if grpcAuthority := os.Getenv("TINKERBELL_GRPC_AUTHORITY"); grpcAuthority != "" {
		c.GRPCAuthority = grpcAuthority
	}
	if httpAuthority := os.Getenv("TINKERBELL_HTTP_AUTHORITY"); httpAuthority != "" {
		c.HTTPAuthority = httpAuthority
	}
	if basicAuthUser := os.Getenv("TINK_AUTH_USERNAME"); basicAuthUser != "" {
		c.HTTPBasicAuthUsername = basicAuthUser
	}
	if basicAuthPass := os.Getenv("TINK_AUTH_PASSWORD"); basicAuthPass != "" {
		c.HTTPBasicAuthPassword = basicAuthPass
	}
}

func main() {
	logger, err := log.Init("github.com/tinkerbell/tink")
	if err != nil {
		panic(err)
	}
	defer logger.Close()

	otel := launcher.ConfigureOpentelemetry(
		launcher.WithServiceName("github.com/tinkerbell/tink"),
	)
	defer otel.Shutdown()

	config := &DaemonConfig{}

	cmd := NewRootCommand(config, logger)
	if err := cmd.ExecuteContext(context.Background()); err != nil {
		os.Exit(1)
	}

}

func NewRootCommand(config *DaemonConfig, logger log.Logger) *cobra.Command {
	cmd := &cobra.Command{
		Use: "tink-server",
		PreRunE: func(cmd *cobra.Command, args []string) error {
			viper, err := createViper(logger)
			if err != nil {
				return err
			}
			return applyViper(viper, cmd)
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			// I am not sure if it is right for this to be here,
			// but as last step I want to keep compatibility with
			// what we have for a little bit and I thinik that's
			// the most aggressive way we have to guarantee that
			// the old way works as before.
			config.PopulateFromLegacyEnvVar()

			logger.Info("starting version " + version)

			sigs := make(chan os.Signal, 1)
			signal.Notify(sigs, syscall.SIGINT, syscall.SIGQUIT, syscall.SIGTERM)
			ctx, closer := context.WithCancel(cmd.Context())
			defer closer()
			// TODO(gianarb): I think we can do better in terms of
			// graceful shutdown and error management but I want to
			// figure this out in another PR
			errCh := make(chan error, 2)

			// TODO(gianarb): I moved this up because we need to be sure that both
			// connection, the one used for the resources and the one used for
			// listening to events and notification are coming in the same way.
			// BUT we should be using the right flags
			connInfo := fmt.Sprintf("dbname=%s user=%s password=%s sslmode=%s",
				config.PGDatabase,
				config.PGUSer,
				config.PGPassword,
				config.PGSSLMode,
			)

			dbCon, err := sql.Open("postgres", connInfo)
			if err != nil {
				return err
			}
			tinkDB := db.Connect(dbCon, logger)

			if config.OnlyMigration {
				logger.Info("Applying migrations. This process will end when migrations will take place.")
				numAppliedMigrations, err := tinkDB.Migrate()
				if err != nil {
					return err
				}
				logger.With("num_applied_migrations", numAppliedMigrations).Info("Migrations applied successfully")
				return nil
			}

			numAvailableMigrations, err := tinkDB.CheckRequiredMigrations()
			if err != nil {
				return err
			}
			if numAvailableMigrations != 0 {
				logger.Info("Your database schema is not up to date. Please apply migrations running tink-server with env var ONLY_MIGRATION set.")
			}

			cert, modT := rpcServer.SetupGRPC(ctx, logger, &rpcServer.ConfigGRPCServer{
				Facility:      config.Facility,
				TLSCert:       config.TLSCert,
				GRPCAuthority: config.GRPCAuthority,
				DB:            tinkDB,
			}, errCh)

			httpServer.SetupHTTP(ctx, logger, &httpServer.HTTPServerConfig{
				CertPEM:               cert,
				ModTime:               modT,
				GRPCAuthority:         config.GRPCAuthority,
				HTTPAuthority:         config.HTTPAuthority,
				HTTPBasicAuthUsername: config.HTTPBasicAuthUsername,
				HTTPBasicAuthPassword: config.HTTPBasicAuthPassword,
			}, errCh)

			<-ctx.Done()
			select {
			case err = <-errCh:
				logger.Error(err)
			case sig := <-sigs:
				logger.With("signal", sig.String()).Info("signal received, stopping servers")
			}

			// wait for grpc server to shutdown
			err = <-errCh
			if err != nil {
				return err
			}
			err = <-errCh
			if err != nil {
				return err
			}
			return nil
		},
	}
	config.AddFlags(cmd.Flags())
	return cmd
}

func createViper(logger log.Logger) (*viper.Viper, error) {
	v := viper.New()
	v.AutomaticEnv()
	v.SetConfigName("tink-server")
	v.AddConfigPath("/etc/tinkerbell")
	v.AddConfigPath(".")
	v.SetEnvKeyReplacer(strings.NewReplacer("-", "_"))

	// If a config file is found, read it in.
	if err := v.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			logger.With("configFile", v.ConfigFileUsed()).Error(err, "could not load config file")
			return nil, err
		}
		logger.Info("no config file found")
	} else {
		logger.With("configFile", v.ConfigFileUsed()).Info("loaded config file")
	}

	return v, nil
}

func applyViper(v *viper.Viper, cmd *cobra.Command) error {
	errors := []error{}

	cmd.Flags().VisitAll(func(f *pflag.Flag) {
		if !f.Changed && v.IsSet(f.Name) {
			val := v.Get(f.Name)
			if err := cmd.Flags().Set(f.Name, fmt.Sprintf("%v", val)); err != nil {
				errors = append(errors, err)
				return
			}
		}
	})

	if len(errors) > 0 {
		errs := []string{}
		for _, err := range errors {
			errs = append(errs, err.Error())
		}
		return fmt.Errorf(strings.Join(errs, ", "))
	}

	return nil
}
