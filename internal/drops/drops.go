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

// Package drops contains logic for drops reporting
package drops

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"
)

type DropStats struct {
	TimeStampEpochMS    int64
	MessageType         string
	CountLocal          int64
	CountRemote         int64
	MeanLatencyLocalMS  int
	MeanLatencyRemoteMS int
}

type DropParseRules struct {
	m    sync.Mutex
	Data map[string][]DropStats
}

var re = regexp.MustCompile(`INFO  \[(?P<thread>.*)\] (?P<date>.{10} .{12}) *(?P<source_file>[^:]*):(?P<source_line>[0-9]*) - (?P<messageType>\S*) messages were dropped in the last 5 s: (?P<localCount>\d*) internal and (?P<remoteCount>\d*) cross node. Mean internal dropped latency: (?P<localLatency>\d*) ms and Mean cross-node dropped latency: (?P<remoteLatency>\d*) ms`)

const dateLayout = "2006-01-02 15:04:05.000"

func (d *DropParseRules) ReadLine(fileName string, line string) error {
	//INFO  [ScheduledTasks:1] 2021-05-17 11:42:44,114  DroppedMessages.java:156 - MUTATION messages were dropped in the last 5 s: 0 internal and 25 cross node. Mean internal dropped latency: 2952 ms and Mean cross-node dropped latency: 2516 ms
	if !re.MatchString(line) {
		return nil
	}
	paramsMap := make(map[string]string)
	match := re.FindStringSubmatch(line)
	for i, name := range re.SubexpNames() {
		if i > 0 && i <= len(match) {
			paramsMap[name] = match[i]
		}
	}
	dateRaw := paramsMap["date"]
	//because golang parse does not handle our ',' in the logs well
	dateRaw = strings.Replace(dateRaw, ",", ".", 1)
	date, err := time.Parse(dateLayout, dateRaw)
	if err != nil {
		return fmt.Errorf("unable to parse date %v with error '%v', string was '%v'", dateRaw, err, line)
	}

	messageType := paramsMap["messageType"]
	localLatencyRaw := paramsMap["localLatency"]
	localLatency, err := strconv.ParseInt(localLatencyRaw, 10, 32)
	if err != nil {
		return fmt.Errorf("unable to parse local latency %v with error '%v', string was '%v'", localLatencyRaw, err, line)
	}
	remoteLatencyRaw := paramsMap["remoteLatency"]
	remoteLatency, err := strconv.ParseInt(remoteLatencyRaw, 10, 32)
	if err != nil {
		return fmt.Errorf("unable to parse remote latency %v with error '%v', string was '%v'", remoteLatencyRaw, err, line)
	}
	remoteCountRaw := paramsMap["remoteCount"]
	remoteCount, err := strconv.ParseInt(remoteCountRaw, 10, 64)
	if err != nil {
		return fmt.Errorf("unable to parse remote count %v with error '%v', string was '%v'", remoteCountRaw, err, line)
	}
	localCountRaw := paramsMap["localCount"]
	localCount, err := strconv.ParseInt(localCountRaw, 10, 64)
	if err != nil {
		return fmt.Errorf("unable to parse local count %v with error '%v', string was '%v'", localCountRaw, err, line)
	}
	parseStates := DropStats{
		TimeStampEpochMS:    date.Unix(),
		MessageType:         messageType,
		CountLocal:          localCount,
		CountRemote:         remoteCount,
		MeanLatencyLocalMS:  int(localLatency),
		MeanLatencyRemoteMS: int(remoteLatency),
	}
	d.m.Lock()
	dropStats, ok := d.Data[fileName]
	if ok {
		dropStats = append(dropStats, parseStates)
		d.Data[fileName] = dropStats
	} else {
		d.Data[fileName] = []DropStats{parseStates}
	}
	d.m.Unlock()
	return nil
}

func (d *DropParseRules) Name() string {
	return "DropParseRules"
}
