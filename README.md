# arangoadmin

## Command line

```
NAME:
   arangoadmin - cli for creating databases and users in arangodb

USAGE:
   arangoadmin [global options] command [command options] [arguments...]

VERSION:
   1.0.0

COMMANDS:
     create-database  create a new arangodb database
     create-user      create a new user for accessing arangodb
     help, h          Shows a list of commands or help for one command

GLOBAL OPTIONS:
   --host value        arangodb host address (default: "arangodb") [$ARANGODB_SERVICE_HOST]
   --port value        arangodb port (default: "8529") [$ARANGODB_SERVICE_PORT]
   --log-level value   log level for the application (default: "info")
   --log-format value  format of the logging out, either of json or text (default: "json")
   --is-secure         connect through a secure endpoint
   --help, -h          show help
   --version, -v       print the version
```

### Subcommands

```
NAME:
   arangoadmin create-database - create a new arangodb database

USAGE:
   arangoadmin create-database [command options] [arguments...]

OPTIONS:
   --admin-user value, --au value      arangodb admin user
   --admin-password value, --ap value  arangodb admin password
   --database value, --db value        name of arangodb database
   --user value, -u value              arangodb user
   --password value, --pw value      arangodb password for new user
   --grant value, -g value             level of access for arangodb user, could be one of ro,rw or none (default: "rw")
```

```
NAME:
   arangoadmin create-user - create a new user for accessing arangodb

USAGE:
   arangoadmin create-user [command options] [arguments...]

OPTIONS:
   --admin-user value, --au value      arangodb admin user (default: "root")
   --admin-password value, --ap value  arangodb admin password
   --user value, -u value              arangodb user
   --password value, --pw value        arangodb password for new user
```
