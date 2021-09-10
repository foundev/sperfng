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

// Package cmd contains all the cli commands
package cmd

import (
	"flag"
	"fmt"
	"log"
	"os"
	"sort"
	"sync"
	"text/tabwriter"

	"golang.org/x/text/message"

	"github.com/HdrHistogram/hdrhistogram-go"
	"github.com/rsds143/sperfng/internal/solrhits"
	"github.com/rsds143/sperfng/pkg/parse"
)

const (
	TPCArg = "tpc"
)

func ExecTPC(args []string) error {
	tpcCmd := flag.NewFlagSet("tpc", flag.ExitOnError)
	if err := tpcCmd.Parse(args); err != nil {
		return err
	}
	fp := &parse.TerminalFileProgress{}
	solrRules := &solrhits.SolrHintParseRules{
		Data: make(map[string][]int64),
	}
	rules := []parse.Rules{solrRules}
	files := tpcCmd.Args()
	parse.Parse(files, fp, rules)

	var wg sync.WaitGroup
	var m sync.Mutex
	var keys []string
	hitsByNode := make(map[string]*hdrhistogram.Histogram)
	for k := range solrRules.Data {
		keys = append(keys, k)
		wg.Add(1)
		go func(index string) {
			v := solrRules.Data[index]
			histogram := hdrhistogram.New(1, 9990000000, 1)
			for _, q := range v {
				if err := histogram.RecordValue(q); err != nil {
					log.Printf("ERROR recording histogram with number %v and error was '%v'", q, err)
				}
			}
			m.Lock()
			hitsByNode[index] = histogram
			m.Unlock()
		}(k)
	}
	wg.Done()
	fmt.Println("")
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 1, ' ', tabwriter.AlignRight|tabwriter.Debug)
	fmt.Println("report complete")
	p := message.NewPrinter(message.MatchLanguage("en_US", "en_UK", "fr", "it"))
	p.Fprintln(w, "host\tp25\tp50\tP99\tmax\tcount")
	p.Fprintln(w, "----\t---\t---\t----\t---\t-----")
	sort.Strings(keys)
	for _, k := range keys {
		v := hitsByNode[k]
		if v == nil {
			log.Printf("node %v has no histogram", k)
			continue
		}
		p25 := v.ValueAtQuantile(25)
		p50 := v.ValueAtQuantile(50)
		p99 := v.ValueAtQuantile(99)
		p.Fprintf(w, "%v\t%v\t%v\t%v\t%v\t%v\n", k, p25, p50, p99, v.Max(), v.TotalCount())
	}
	w.Flush()
	return nil
}
