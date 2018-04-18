# arangomanager

## Command line
```
NAME:
   arangomanager - cli for managing databases, users and collection in arangodb

USAGE:
   arangomanager [global options] command [command options] [arguments...]

VERSION:
   0.0.0

COMMANDS:
     create-db-user     create database, users and database level access
     create-collection  create collections in a database
     help, h            Shows a list of commands or help for one command

GLOBAL OPTIONS:
   --host value        arangodb host address (default: "arangodb") [$ARANGODB_SERVICE_HOST]
   --port value        arangodb port [$ARANGODB_SERVICE_PORT]
   --log-level value   log level for the application (default: "info")
   --log-format value  format of the logging out, either of json or text (default: "json")
   --help, -h          show help
   --version, -v       print the version
```

### Subcommands
```
NAME:
   arangomanager create-db-user - create database, users and database level access

USAGE:
   arangomanager create-db-user [command options] [arguments...]

OPTIONS:
   --admin-user value, --au value      arangodb admin user
   --admin-password value, --ap value  arangodb admin password
   --database value, --db value        database to create, skip if it already exists
   --user value, -u value              user to create, skip if the user already exists
   --password value, --pass value      password for the user
   --grant value, -g value             access level of user, could be one of ro,rw or none (default: "rw")
```

```
NAME:
   arangomanager create-collection - create collections in a database

USAGE:
   arangomanager create-collection [command options] [arguments...]

OPTIONS:
   --database value, --db value    database where the collections will be created
   --collection value, -c value    name of collection to create, skip if it already exist
   --is-edge                       flag to create an edge collection
   --user value, -u value          user to create collection, should have required grant to create collection
   --password value, --pass value  password for the user
```
