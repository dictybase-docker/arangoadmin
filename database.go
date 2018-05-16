package main

import (
	"context"
	"fmt"
	"io/ioutil"

	"github.com/arangodb/go-driver"
	"github.com/sirupsen/logrus"
	cli "gopkg.in/urfave/cli.v1"
	"gopkg.in/yaml.v2"
)

func runDatabaseAction(client driver.Client, db *Database, logger *logrus.Entry, c *cli.Context) error {
	var adb driver.Database
	switch db.Action {
	case "create":
		ok, err := client.DatabaseExists(context.Background(), db.Name)
		if err != nil {
			return fmt.Errorf("error in checking existence of database %s %s", db.Name, err)
		}
		if ok {
			logger.Infof("database %s exists, nothing to create", db.Name)
		} else {
			adb, err = client.CreateDatabase(context.Background(), db.Name, nil)
			if err != nil {
				return fmt.Errorf("error in creating database %s %s", db.Name, err)
			}
			logger.Infof("created database %s", db.Name)
		}
	case "delete":
		ok, err := client.DatabaseExists(context.Background(), db.Name)
		if err != nil {
			return fmt.Errorf("error in checking existence of database %s %s", db.Name, err)
		}
		if !ok {
			logger.Infof("database %s does not exists, nothing to delete", db.Name)
			return nil
		}
		adb, err := client.Database(context.Background(), db.Name)
		if err != nil {
			return fmt.Errorf("error in retrieving database %s %s", db.Name, err)
		}
		if err := adb.Remove(context.Background()); err != nil {
			return fmt.Errorf("error in removing database %s %s", db.Name, err)
		}
		logger.Infof("removed database %s", db.Name)
		return nil
	default:
		return fmt.Errorf("unsupported database action %s", db.Action)
	}
	var cusers []User
	for _, u := range db.Allowed {
		ok, err := client.UserExists(context.Background(), u.User)
		if err != nil {
			return fmt.Errorf("error in checking for user %s %s", u.User, err)
		}
		if !ok {
			cusers = append(cusers, u)
			continue
		}
		logger.Infof("user %s exists, skipping", u.User)
		euser, err := client.User(context.Background(), u.User)
		if err != nil {
			return fmt.Errorf("unable to retrieve user %s", err)
		}
		err = euser.SetDatabaseAccess(context.Background(), adb, getGrant(u.Grant))
		if err != nil {
			return fmt.Errorf(
				"error in granting permission %s for user %s %s",
				u.Grant,
				u.User,
				err,
			)
		}
		logger.Infof("successfully granted permission %s to existing user %s", u.Grant, u.User)
	}
	if len(cusers) == 0 {
		return nil
	}
	if !c.IsSet("password-file") {
		return fmt.Errorf("%s\n", "**password-file** is needed for creating user")
	}
	cont, err := ioutil.ReadFile(c.String("password-file"))
	if err != nil {
		return fmt.Errorf("cannot read password file %s", err)
	}
	pw := new(PwList)
	err = yaml.Unmarshal(cont, pw)
	if err != nil {
		return fmt.Errorf("cannot unmarshal yaml for password file %s", err)
	}
	for i, u := range cusers {
		dbUser, err := client.CreateUser(
			context.Background(),
			u.User,
			&driver.UserOptions{
				Password: pw.Passwords[i],
			})
		if err != nil {
			return fmt.Errorf("error in creating user %s %s", u.User, err)
		}
		logger.Infof("successfully created user %s", u.User)
		err = dbUser.SetDatabaseAccess(context.Background(), adb, getGrant(u.Grant))
		if err != nil {
			return fmt.Errorf(
				"error in granting permission %s for user %s %s",
				u.Grant,
				u.User,
				err,
			)
		}
		logger.Infof("successfully granted permission %s existing user %s", u.Grant, u.User)
	}
	return nil
}

func runCollectionAction(client driver.Client, coll *Collection, isEdge bool, logger *logrus.Entry) error {
	ok, err := client.DatabaseExists(context.Background(), coll.Database)
	if err != nil {
		return fmt.Errorf("error in checking for database %s %s", coll.Database, err)
	}
	if !ok {
		return fmt.Errorf("database %s does not exist, cannot create collection", coll.Database)
	}
	db, err := client.Database(context.Background(), coll.Database)
	if err != nil {
		return fmt.Errorf("error in retrieving database %s %s", coll.Database, err)
	}
	switch coll.Action {
	case "create":
		ok, err = db.CollectionExists(context.Background(), coll.Name)
		if err != nil {
			return fmt.Errorf("error in collection lookup %s", err)
		}
		if ok {
			logger.Infof("collection %s exist, nothing to create", coll.Name)
			return nil
		}
		opt := &driver.CreateCollectionOptions{}
		if coll.Key.Increment != 0 {
			opt.KeyOptions = &driver.CollectionKeyOptions{
				Increment: coll.Key.Increment,
				Offset:    coll.Key.Offset,
			}
		}
		if isEdge {
			opt.Type = driver.CollectionTypeEdge
		}
		_, err = db.CreateCollection(context.Background(), coll.Name, opt)
		if err != nil {
			fmt.Errorf("error in creating collection %s %s", coll.Name, err)
		}
		logger.Infof("successfully created collection %s", coll.Name)
	default:
		return fmt.Errorf("unsupported collection action %s", coll.Action)
	}
	return nil
}

func getGrant(g string) driver.Grant {
	var grnt driver.Grant
	switch g {
	case "rw":
		grnt = driver.GrantReadWrite
	case "ro":
		grnt = driver.GrantReadOnly
	default:
		grnt = driver.GrantNone
	}
	return grnt
}
