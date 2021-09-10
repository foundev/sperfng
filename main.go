//  Copyright 2021 Ryan Svihla
//
//  Licensed under the Apache License, Version 2.0 (the "License");
//  you may not use this file except in compliance with the License.
//  You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
//  Unless required by applicable law or agreed to in writing, software
//  distributed under the License is distributed on an "AS IS" BASIS,
//  WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//  See the License for the specific language governing permissions and
//  limitations under the License.

// Package main entry point
package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"runtime/debug"

	"github.com/rsds143/sperfng/cmd"
)

func main() {
	debug.SetMaxThreads(20000)
	argLen := len(os.Args)
	if argLen < 2 {
		flag.Usage()
		os.Exit(1)
	}
	subcmd := os.Args[1]
	var subCmdArgs []string
	if argLen > 2 {
		subCmdArgs = os.Args[2:]
	}
	switch subcmd {
	case cmd.SolrHitsArg:
		if err := cmd.ExecSolrHits(subCmdArgs); err != nil {
			log.Fatal(err)
		}
	case cmd.DropsArg:
		if err := cmd.ExecDrops(subCmdArgs); err != nil {
			log.Fatal(err)
		}
	default:
		fmt.Printf("command '%v' args '%v' used\n", subcmd, subCmdArgs)
		flag.Usage()
		os.Exit(1)
	}
}
