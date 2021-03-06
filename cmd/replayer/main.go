// Copyright 2015 PingCAP, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"flag"
	"fmt"
	"time"

	"github.com/juju/errors"
	"github.com/ngaut/log"
	"github.com/pingcap/mysqlrelay/relay"
	"github.com/pingcap/tidb"
)

var (
	store     = flag.String("store", "memory", "registered store name, [memory, goleveldb, boltdb]")
	storePath = flag.String("path", "/tmp/tidb", "tidb storage path")
	logLevel  = flag.String("L", "error", "log level: info, debug, warn, error, fatal")
	lease     = flag.Int("lease", 1, "schema lease seconds, very dangerous to change only if you know what you do")
	relayPath = flag.String("relay_path", "/tmp/tidb_relay.log", "relay log file for replaying")
	check     = flag.Bool("check", true, "check when in replaying")
)

func main() {
	flag.Parse()

	if *lease < 0 {
		log.Fatalf("invalid lease seconds %d", *lease)
	}

	tidb.SetSchemaLease(time.Duration(*lease) * time.Second)

	log.SetLevelByString(*logLevel)
	store, err := tidb.NewStore(fmt.Sprintf("%s://%s", *store, *storePath))
	if err != nil {
		log.Fatal(err)
	}

	var driver relay.IDriver
	driver = relay.NewTiDBDriver(store)
	replayer, err := relay.NewReplayer(driver, *relayPath, *check)
	if err != nil {
		log.Fatal(err)
	}

	replayer.OnRecordRead = func(rec *relay.Record) {
		fmt.Printf("%s\n", rec)
	}

	err = replayer.Run()
	if err != nil {
		log.Fatal(errors.ErrorStack(err))
	}
}
