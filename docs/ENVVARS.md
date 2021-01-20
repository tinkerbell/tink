# Environment Variables

This describes environment variables in Tink server.

| Name                                                                                           | Type   | Service(s) | Description                                                                                                          |
| ---------------------------------------------------------------------------------------------- | ------ | ---------- | -------------------------------------------------------------------------------------------------------------------- |
| `TINK_TLS_ENABLED=false`                                                                       | bool   | server/cli | toggles whether TLS will be terminated by Tink server or whether the cli will use TLS for communication              |
| `TINK_AUTH_USERNAME=tink`                                                                      | string | server     | username to use for basic auth to http endpoints                                                                     |
| `TINK_AUTH_PASSWORD=tink`                                                                      | string | server     | password to use for basic auth to http endpoints                                                                     |
| `TINKERBELL_CERT_URL=http://127.0.0.1:42114/cert`                                              | string | cli        | string url from which to get a TLS certificate                                                                       |
| `TINKERBELL_CERTS_DIR=/certs`                                                                  | string | server     | a directory which contains the `bundle.pem` and `server-key.pem` files                                               |
| `TINKERBELL_TLS_CERT="-----BEGIN RSA PRIVATE KEY-----\n....\n-----END RSA PRIVATE KEY-----\n"` | string | server     | a TLS certificate for use with Tink server                                                                           |
| `TINKERBELL_GRPC_AUTHORITY=127.0.0.1:42113`                                                    | string | server/cli | string url of the Tink gRPC server                                                                                   |
| `TINKERBELL_HTTP_AUTHORITY=127.0.0.1:42114`                                                    | string | server     | string url of the Tink HTTP server                                                                                   |
| `FACILITY=onprem`                                                                              | string | server/cli | location for which the Tink server serves                                                                            |
| `PGDATABASE=tinkerbell`                                                                        | string | server     | name of the PostgreSQL database for use in the Tink server                                                           |
| `PGUSER=tink`                                                                                  | string | server     | PostgreSQL username for connecting to the DB                                                                         |
| `PGPASSWORD=tink`                                                                              | string | server     | PostgreSQL password for connecting to the DB                                                                         |
| `PGSSLMODE=disable`                                                                            | string | server     | sets the PostgreSQL SSL priority [docs](https://www.postgresql.org/docs/10/libpq-connect.html#LIBPQ-CONNECT-SSLMODE) |
| `MAX_WORKFLOW_DATA_VERSIONS=`                                                                  | int    | server     | maximum number of workflow data versions to be kept in database                                                      |
| `EVENTS_TTL=60`                                                                                | string | server     | purges the events in the events table that have passed this TTL in minutes                                           |
| `ONLY_MIGRATION=true`                                                                          | bool   | server     | when true, applies DB migrations and then exits                                                                      |
