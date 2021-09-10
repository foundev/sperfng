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
	"os"
	"sort"
	"text/tabwriter"

	"github.com/rsds143/sperfng/internal/drops"
	"github.com/rsds143/sperfng/pkg/parse"
	"golang.org/x/text/message"
)

const (
	DropsArg = "drops"
)

func ExecDrops(args []string) error {
	dropsCmd := flag.NewFlagSet("drops", flag.ExitOnError)
	if err := dropsCmd.Parse(args); err != nil {
		return err
	}
	fp := &parse.TerminalFileProgress{}
	dropsRules := &drops.DropParseRules{
		Data: make(map[string][]drops.DropStats),
	}
	rules := []parse.Rules{dropsRules}
	files := dropsCmd.Args()
	parse.Parse(files, fp, rules)
	fmt.Println("")
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 1, ' ', tabwriter.AlignRight|tabwriter.Debug)
	fmt.Println("report complete")
	p := message.NewPrinter(message.MatchLanguage("en_US", "en_UK", "fr", "it"))
	p.Fprintln(w, "host\tmessage_type\tavg freq in log\tlocal count\tremote count\tavg local latency ms\tremote latency ms")
	p.Fprintln(w, "----\t------------\t---------------\t-----------\t------------\t--------------------\t-----------------")

	var nodes []string
	for k := range dropsRules.Data {
		nodes = append(nodes, k)
	}
	sort.Strings(nodes)
	for _, node := range nodes {
		var first int64
		var last int64
		var totalReportBlock reportBlock
		messageTypesBlocks := make(map[string]reportBlock)
		dropStats := dropsRules.Data[node]
		for _, dropStat := range dropStats {
			if dropStat.TimeStampEpochMS > last {
				last = dropStat.TimeStampEpochMS
			}
			if dropStat.TimeStampEpochMS < first || first == 0 {
				first = dropStat.TimeStampEpochMS
			}
			totalReportBlock.Freq += 1
			totalReportBlock.TotalLocal += dropStat.CountLocal
			totalReportBlock.TotalRemote += dropStat.CountRemote
			totalReportBlock.TotalLocalLatencyMS += int64(dropStat.MeanLatencyLocalMS)
			totalReportBlock.TotalRemoteLatencyMS += int64(dropStat.MeanLatencyRemoteMS)
			var messageReportBlock reportBlock
			existingMessageReportBlock, ok := messageTypesBlocks[dropStat.MessageType]
			if ok {
				messageReportBlock = existingMessageReportBlock
			}
			messageReportBlock.Freq += 1
			messageReportBlock.TotalLocal += dropStat.CountLocal
			messageReportBlock.TotalRemote += dropStat.CountRemote
			messageReportBlock.TotalLocalLatencyMS += int64(dropStat.MeanLatencyLocalMS)
			messageReportBlock.TotalRemoteLatencyMS += int64(dropStat.MeanLatencyRemoteMS)
			messageTypesBlocks[dropStat.MessageType] = messageReportBlock
		}
		totalTime := last - first
		avgFreq := float64(totalTime) / float64(totalReportBlock.Freq)
		avgLocalLatency := float64(totalReportBlock.TotalLocalLatencyMS) / float64(totalReportBlock.Freq)
		avgRemoteLatency := float64(totalReportBlock.TotalRemoteLatencyMS) / float64(totalReportBlock.Freq)
		p.Fprintf(w, "%v\t%v\t%.2f ms\t%v\t%v\t%.2f\t%.2f\n", node, "-total-", avgFreq, totalReportBlock.TotalLocal, totalReportBlock.TotalRemote, avgLocalLatency, avgRemoteLatency)
		var messageTypeForNode []string
		for k := range messageTypesBlocks {
			messageTypeForNode = append(messageTypeForNode, k)
		}
		sort.Strings(messageTypeForNode)
		for _, messageType := range messageTypeForNode {
			messageTypeStats := messageTypesBlocks[messageType]
			avgLocalLatency := float64(messageTypeStats.TotalLocalLatencyMS) / float64(messageTypeStats.Freq)
			avgRemoteLatency := float64(messageTypeStats.TotalRemoteLatencyMS) / float64(messageTypeStats.Freq)
			p.Fprintf(w, "\t%v\t--\t%v\t%v\t%.2fs\t%.2f\n", messageType, messageTypeStats.TotalLocal, messageTypeStats.TotalRemote, avgLocalLatency, avgRemoteLatency)
		}
	}
	w.Flush()
	return nil
}

type reportBlock struct {
	Freq                 int64
	TotalRemoteLatencyMS int64
	TotalLocalLatencyMS  int64
	TotalLocal           int64
	TotalRemote          int64
}
