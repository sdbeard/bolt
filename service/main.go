/*
The MIT License (MIT)

--Copyright (c) 2013 Ben Johnson--
Copyright (c) 2020 Sean Beard

Permission is hereby granted, free of charge, to any person obtaining a copy of
this software and associated documentation files (the "Software"), to deal in
the Software without restriction, including without limitation the rights to
use, copy, modify, merge, publish, distribute, sublicense, and/or sell copies of
the Software, and to permit persons to whom the Software is furnished to do so,
subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY, FITNESS
FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE AUTHORS OR
COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER
IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN
CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.
*/

package main

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"runtime"
	"strings"
	"syscall"
	"time"

	"bitbucket.org/kpsgo/goutil/logging"
	"bitbucket.org/kpslib/golib-eda/unio"
	"github.com/sdbeard/bolt"
	logger "github.com/sirupsen/logrus"
)

/*

Assumptions
-------------
	-

TODO
-------------
	-

*/

var (
	command      string
	paramVersion bool
	paramHelp    bool

	buildDate   = time.Now()
	compileDate = ""
	gitCommit   = ""
	majorver    = "0"
	minorver    = "0"
	buildver    = "0"
	env         = "local"
)

func init() {
	// Set the command line flags
	flag.BoolVar(&paramVersion, "version", false, "determine whether or not to show the version information")
	flag.BoolVar(&paramHelp, "help", false, "determine whether or not to show the help information")

	// Initialize the logging engine
	if err := logging.InitializeDefaultLogging(); err != nil {
		panic(err)
	}
}

func main() {
	buildDate := time.Now()
	if !strings.EqualFold(compileDate, "") {
		buildDate, _ = time.Parse("2006-01-02T15:04:05 -0700", compileDate)
	}

	// Write the startup messages to the log
	logger.Infof("BoltDB Web Service v%s.%s.%s-%s", majorver, minorver, buildver, env)
	logger.Info("Copyright © 2020 Krone Productions")
	logger.WithFields(logger.Fields{
		"App Version": fmt.Sprintf("%s.%s.%s-%s", majorver, minorver, buildver, env),
		"Build":       buildDate.Format("02 Jan 2006 15:04:05 MST"),
		"Environment": env,
		"GO Version":  runtime.Version(),
		"PID":         os.Getpid(),
	}).Infof("Runtime configuration")

	db, err := bolt.Open("test.db", 0600, &bolt.Options{Timeout: 1 * time.Second})
	if err != nil {
		logger.Fatal(err.Error())
		os.Exit(99)
	}
	defer db.Close()

	newRule := &unio.PatternRule{
		Pattern:     make(map[string]interface{}),
		RuleTargets: make([]unio.RuleTarget, 0),
		RuleID:      "01",
		RuleType:    unio.PATTERN,
	}
	ruleString := newRule.ToJSON()
	err = db.Update(func(tx *bolt.Tx) error {
		bucket, err := tx.CreateBucket([]byte(newRule.RuleType.String()))
		if err != nil {
			if errors.Is(err, bolt.ErrBucketExists) {
				return nil
			}
			return fmt.Errorf("create bucket: %s", err)
		}
		return bucket.Put([]byte(newRule.RuleID), []byte(ruleString))
	})
	if err != nil {
		logger.Fatal(err.Error())
		os.Exit(99)
	}

	err = db.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(newRule.RuleType.String()))
		value := bucket.Get([]byte(newRule.RuleID))
		fmt.Println(string(value))
		return nil
	})
	if err != nil {
		logger.Fatal(err.Error())
		os.Exit(99)
	}

	//stop := createStopChannel()

	// Create the API
	// Start the API

	//<-stop
	//close(stop)

	// Exit the service
	os.Exit(0)
}

// createStopChannel created the channel that manages the graceful shutdown of all
// of the processes and threads. This channel captures system messages and allows
// the application to react to shutdown requests from the underlying OS.
func createStopChannel() chan os.Signal {
	stopChannel := make(chan os.Signal, 1)
	signal.Notify(stopChannel, os.Interrupt)
	signal.Notify(stopChannel, syscall.SIGTERM)
	signal.Notify(stopChannel, syscall.SIGKILL)
	signal.Notify(stopChannel, syscall.SIGINT)

	return stopChannel
}
