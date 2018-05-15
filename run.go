package main

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"github.com/sirupsen/logrus"
	"gopkg.in/urfave/cli.v1"
	"gopkg.in/yaml.v2"
)

var collection = regexp.MustCompile(`^type:\scollection`)
var edgeCollection = regexp.MustCompile(`^type:\sedgeCollection`)
var database = regexp.MustCompile(`^type:\sdatabase`)

type ByNumFileName []os.FileInfo

func (nf ByNumFileName) Len() int {
	return len(nf)
}
func (nf ByNumFileName) Swap(i, j int) {
	nf[i], nf[j] = nf[j], nf[i]
}
func (nf ByNumFileName) Less(i, j int) bool {
	//Grab integer value
	fileA := nf[i].Name()
	fileB := nf[j].Name()

	a, errA := strconv.ParseInt(strings.Split(fileA, "_")[0], 10, 64)
	b, errB := strconv.ParseInt(strings.Split(fileB, "_")[0], 10, 64)
	if errA != nil || errB != nil {
		return fileA < fileB
	}
	return a < b
}

func Run(c *cli.Context) error {
	var dir string
	if !c.IsSet("dir") {
		d, err := os.Getwd()
		if err != nil {
			return cli.NewExitError(
				fmt.Sprintf("could not get current working dir %s", err),
				2,
			)
		}
		dir = d
	} else {
		dir = c.String("dir")
	}
	logger := getLogger(c)
	client, err := getClient(
		c.GlobalString("host"),
		c.GlobalString("port"),
		c.String("admin-user"),
		c.String("admin-password"),
		c.Bool("is-secure"),
	)
	if err != nil {
		return cli.NewExitError(err.Error(), 2)
	}
	v, err := client.Version(context.Background())
	if err != nil {
		return cli.NewExitError(
			fmt.Sprintf("unable to get version %s", err),
			2,
		)
	}
	logger.Infof("Got version %s", v.Version)
	// read the input dir
	f, err := os.Open(dir)
	if err != nil {
		return cli.NewExitError(
			fmt.Sprintf("error in reading directory %s %s", dir, err),
			2,
		)
	}
	allInfo, err := f.Readdir(-1)
	sort.Sort(ByNumFileName(allInfo))
	if err != nil {
		return cli.NewExitError(
			fmt.Sprintf("unable to read content from the directory %s", err),
			2,
		)
	}
	// go through all yaml files
	for _, finfo := range allInfo {
		fullPath := filepath.Join(dir, finfo.Name())
		if finfo.IsDir() {
			logger.Debugf("skipped dir %s", fullPath)
			continue
		}
		if !strings.HasSuffix(finfo.Name(), "yaml") {
			logger.Debugf("skipped non yaml file %s", fullPath)
			continue
		}
		cont, err := ioutil.ReadFile(fullPath)
		if err != nil {
			return cli.NewExitError(
				fmt.Sprintf("unable to read file %s %s", fullPath, err),
				2,
			)
		}
		logger.Debugf("going to process file %s", fullPath)
		switch {
		case database.Match(cont):
			db := new(Database)
			err := yaml.UnmarshalStrict(cont, db)
			if err != nil {
				return cli.NewExitError(
					fmt.Sprintf("error in unmarshalling yaml %s", err.Error()),
					2,
				)
			}
			logger.Debug("going to process type database")
			if err := runDatabaseAction(client, db, logger, c); err != nil {
				return cli.NewExitError(err.Error(), 2)
			}
			logger.Debug("processed type database")
		case collection.Match(cont):
			coll := new(Collection)
			err := yaml.UnmarshalStrict(cont, coll)
			if err != nil {
				return cli.NewExitError(
					fmt.Sprintf("error in unmarshalling yaml %s", err.Error()),
					2,
				)
			}
			logger.Debug("going to process type collection")
			if err := runCollectionAction(client, coll, false, logger); err != nil {
				return cli.NewExitError(err.Error(), 2)
			}
			logger.Debug("processed type collection")
		case edgeCollection.Match(cont):
			coll := new(Collection)
			err := yaml.UnmarshalStrict(cont, coll)
			if err != nil {
				return cli.NewExitError(
					fmt.Sprintf("error in unmarshalling yaml %s", err.Error()),
					2,
				)
			}
			logger.Debug("going to process type edgeCollection")
			if err := runCollectionAction(client, coll, true, logger); err != nil {
				return cli.NewExitError(err.Error(), 2)
			}
		default:
			return cli.NewExitError("yaml type is not supported", 2)
		}
		logger.Infof("processed file %s", fullPath)
	}
	return nil
}

func getLogger(c *cli.Context) *logrus.Entry {
	log := logrus.New()
	log.Out = os.Stderr
	switch c.GlobalString("log-format") {
	case "text":
		log.Formatter = &logrus.TextFormatter{
			TimestampFormat: "02/Jan/2006:15:04:05",
		}
	case "json":
		log.Formatter = &logrus.JSONFormatter{
			TimestampFormat: "02/Jan/2006:15:04:05",
		}
	}
	l := c.GlobalString("log-level")
	switch l {
	case "debug":
		log.Level = logrus.DebugLevel
	case "warn":
		log.Level = logrus.WarnLevel
	case "error":
		log.Level = logrus.ErrorLevel
	case "fatal":
		log.Level = logrus.FatalLevel
	case "panic":
		log.Level = logrus.PanicLevel
	}
	return logrus.NewEntry(log)
}
