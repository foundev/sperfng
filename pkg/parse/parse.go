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

// Package parse contains logic for parsing a set of files
package parse

import (
	"bufio"
	"fmt"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

type TerminalFileProgress struct {
	errors []error
	m      sync.Mutex
}

func (t *TerminalFileProgress) Open(fileName string) {
	t.m.Lock()
	fmt.Printf(".")
	t.m.Unlock()
}

func (t *TerminalFileProgress) Failure(fileName string, err error) {
	t.m.Lock()
	t.errors = append(t.errors, fmt.Errorf("error opening file %v with error '%v'", fileName, err))
	fmt.Printf("x")
	t.m.Unlock()
}

func (t *TerminalFileProgress) PrintErrors() {
	t.m.Lock()
	fmt.Println("file errors")
	fmt.Println("-----------")
	for i, err := range t.errors {
		fmt.Printf("%v - ", i)
		fmt.Println(err)
	}
	t.m.Unlock()
}

type FileProgress interface {
	Open(fileName string)
	Failure(fileName string, err error)
}

type Rules interface {
	ReadLine(filename, line string) error
	Name() string
}

func maybeConvertToNode(fileName string) string {
	tokens := strings.Split(fileName, string(filepath.Separator))
	var wasFound bool
	for _, token := range tokens {
		if wasFound {
			return token
		}
		if token == "nodes" {
			wasFound = true
		}
	}
	return fileName
}

func parseFile(file string, fileProgress FileProgress, rules []Rules, wg *sync.WaitGroup) {
	defer wg.Done()
	friendlyName := maybeConvertToNode(file)
	fh, err := os.Open(file)
	if err != nil {
		fileProgress.Failure(file, err)
		return
	} else {
		fileProgress.Open(file)
	}
	defer func() {
		err := fh.Close()
		if err != nil {
			log.Printf("WARN unable to close file %v with error '%v'", file, err)
		}
	}()
	reader := bufio.NewReader(fh)
	scanner := bufio.NewScanner(reader)
	counter := 0
	for scanner.Scan() {
		counter++
		line := scanner.Text()
		for _, rule := range rules {
			if err := rule.ReadLine(friendlyName, line); err != nil {
				log.Printf("ERROR unable to read line '%v' for file '%v' using parser '%v' with error '%v'",
					counter, file, rule.Name(), err)
			}
		}
	}
}

func Parse(files []string, fileProgress FileProgress, rules []Rules) {
	var foundFiles []string
	for _, f := range files {
		filesToParse, err := findFiles(f)
		if err != nil {
			log.Fatalf("critical error %v", err)
		}
		foundFiles = append(foundFiles, filesToParse...)
	}
	var wg sync.WaitGroup
	for i := 0; i < len(foundFiles); i++ {
		f := foundFiles[i]
		wg.Add(1)
		go parseFile(f, fileProgress, rules, &wg)
	}
	wg.Wait()
}

func findFiles(fileToSearch string) ([]string, error) {
	var files []string
	return files, filepath.Walk(fileToSearch, func(path string, fileInfo fs.FileInfo, err error) error {
		if !fileInfo.IsDir() && (strings.Contains(path, "system.log") || strings.Contains(path, "debug.log")) {
			files = append(files, path)
		}
		return nil
	})
}
