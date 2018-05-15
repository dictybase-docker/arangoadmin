package main

// KeyOptions for how keys in the collections are created
type KeyOptions struct {
	Increment int `yaml:"increment"`
	Offset    int `yaml:"offset"`
}

// User the data used to update a user account
type User struct {
	User  string `yaml:"user"`
	Pass  string `yaml:"password"`
	Grant string `yaml:"grant"`
}

// Database the YAML struct for configuring a database migration.
// Expected yaml
//
//    type: database
//    action: create(or delete)
//    name: mydb
//	  allowed:
//		- user: borat
//        password: zara
//        grant: rw
//
type Database struct {
	Operation string `yaml:"type"`
	Action    string `yaml:"action"`
	Allowed   []User `yaml:"allowed,omitempty"`
	Name      string `yaml:"name"`
}

// Collection the YAML struct for configuring a collection migration.
//
//  type: collection
//	action: create
//	name: zumba
//	database: mydb
//	key:
//	 increment: 3
//	 offset 1258
//
type Collection struct {
	Operation string     `yaml:"type"`
	Action    string     `yaml:"action"`
	Name      string     `yaml:"name"`
	Key       KeyOptions `yaml:"key,omitempty"`
	Database  string     `yaml:"database"`
}

type PwList struct {
	Passwords []string `yaml:"passwords"`
}
