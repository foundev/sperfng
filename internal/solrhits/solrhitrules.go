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

// Package solrhits contains logic for solr hits reporting
package solrhits

import (
	"fmt"
	"regexp"
	"strconv"
	"sync"
)

type SolrHintParseRules struct {
	m    sync.Mutex
	Data map[string][]int64
}

var re = regexp.MustCompile(`WARN  \[\S*MessageServer query worker - (?P<thread_id>[0-9]*)\] (?P<date>.{10} .{12}) *(?P<source_file>[^:]*):(?P<source_line>[0-9]*) - slow: \[(?P<table>\S*)\]  hits=(?P<hits>\d*) status=(?P<status>\d) QTime=(?P<qtime>\d*)`)

func (s *SolrHintParseRules) ReadLine(fileName string, line string) error {
	//WARN  [RemoteMessageServer query worker - 52] 2021-05-19 01:10:49,847  SolrCore.java:2208 - slow: [cybs_rtd_search.transaction_search]  hits=408877851 status=0 QTime=4499
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
	hitsRaw := paramsMap["hits"]
	hits, err := strconv.ParseInt(hitsRaw, 10, 64)
	if err != nil {
		return fmt.Errorf("unable to parse hits %v with error '%v', string was '%v'", hitsRaw, err, line)
	}
	s.m.Lock()
	queries, ok := s.Data[fileName]
	if ok {
		queries = append(queries, hits)
		s.Data[fileName] = queries
	} else {
		s.Data[fileName] = []int64{hits}
	}
	s.m.Unlock()
	return nil
}

func (s *SolrHintParseRules) Name() string {
	return "SolrHintParseRules"
}
