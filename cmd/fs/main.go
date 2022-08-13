// Copyright (c) 2019-2020 Siemens AG
// Copyright (c) 2019-2021 Jonas Plum
//
// Permission is hereby granted, free of charge, to any person obtaining a copy of
// this software and associated documentation files (the "Software"), to deal in
// the Software without restriction, including without limitation the rights to
// use, copy, modify, merge, publish, distribute, sublicense, and/or sell copies of
// the Software, and to permit persons to whom the Software is furnished to do so,
// subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY, FITNESS
// FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE AUTHORS OR
// COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER
// IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN
// CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.
//
// Author(s): Jonas Plum

// Package fs implements the fs command line tool that has various subcommands
// which imitate unix commands but for nested file system structures.
//
//	cat      Print files
//	file     Determine files types
//	hashsum  Print hashsums
//	ls       List directory contents
//	stat     Display file status
//	strings  Find the printable strings in an object, or other binary, file
//	tree     List contents of directories in a tree-like format
//
// # Usage Examples
//
// Extract the Amcache.hve file from a NTFS image in a zip file:
//
//	fs cat case/evidence.zip/ntfs.dd/Windows/AppCompat/Programs/Amcache.hve > Amcache.hve
//
// Hash all files in a zip file:
//
//	fs hashsum case/evidence.zip/*
package main

import (
	"io/fs"
	"log"

	"github.com/spf13/cobra"

	"github.com/forensicanalysis/fscmd"
	"github.com/forensicanalysis/fslib"
	"github.com/forensicanalysis/recursivefs"
)

func main() {
	fsys := recursivefs.New()
	fsCmd := fscmd.FSCommand(func(_ *cobra.Command, args []string) (fs.FS, []string, error) {
		var names []string
		for _, arg := range args {
			name, err := fslib.ToFSPath(arg)
			if err != nil {
				return nil, nil, err
			}
			names = append(names, name)
		}
		return fsys, names, nil
	})
	fsCmd.Use = "fs"
	fsCmd.Short = "recursive file, filesystem and archive commands"
	err := fsCmd.Execute()
	if err != nil {
		log.Fatal(err)
	}
}
