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

// Package main contains logic to delete files in a directory
package main

import (
	"fmt"
	"io/fs"
	"log"
	"os"
	"path/filepath"
)

func main() {
	dir := os.Args[1]
	err := filepath.Walk(dir, func(path string, info fs.FileInfo, err error) error {
		if err != nil {
			return fmt.Errorf("while walking %v found previous error %v, exiting path", path, err)
		}
		if info.IsDir() && path == dir {
			log.Printf("entering parent directory %v", path)
			return nil
		}
		if info.IsDir() {
			log.Printf("removing directory %v", path)
			if err := os.RemoveAll(path); err != nil {
				return fmt.Errorf("trying to remove directory %v resulted in error %v", path, err)
			}
			return nil
		}
		log.Printf("removing file %v", path)
		if err := os.Remove(path); err != nil {
			return fmt.Errorf("trying to remove file %v resulted in error %v", path, err)
		}
		return nil
	})
	if err != nil {
		log.Fatalf("error walking the path %q: %v", dir, err)
	}
}
