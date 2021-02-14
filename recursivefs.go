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

// Package recursivefs provides an io/fs implementation that can open paths in
// nested file systems recursively. The forensicfs are identified using the filetype
// library. This way e.g. a file in a zip inside a disk image can be accessed.
package recursivefs

import (
	"fmt"
	"github.com/forensicanalysis/fslib/osfs"
	"io/fs"
)

type element struct {
	// Parser *filetype.Filetype
	FS  fs.FS
	Key string
}

// FS implements a read-only meta file system that can access nested file system
// structures.
type FS struct {
	root fs.FS
}

// New creates a new recursive FS.
func New() *FS {
	return &FS{root: osfs.New()}
}

// New creates a new recursive FS.
func NewFS(root fs.FS) *FS {
	return &FS{root: root}
}

// Open returns a File for the given location.
func (fsys *FS) Open(name string) (f fs.File, err error) {
	valid := fs.ValidPath(name)
	if !valid {
		return nil, fmt.Errorf("path %s invalid", name)
	}

	elems, err := parseRealPath(fsys.root, name)
	if err != nil {
		return
	}

	localFS := fsys.root
	var childName = ""
	for _, elem := range elems {
		f, err = elem.FS.Open(elem.Key)
		if err != nil {
			return nil, err
		}

		childName = elem.Key
		localFS = elem.FS
	}

	fi, err := f.Stat()
	if err != nil {
		return nil, err
	}

	if fi.IsDir() {
		return &Item{parentFS: localFS, localPath: childName, internal: f}, nil
	}

	subFS, err := childFS(f, childName)
	if err != nil {
		return nil, err
	}

	return &Item{
		parentFS:  localFS,
		localPath: childName,
		internal:  f,
		childFS:   subFS,
	}, nil
}
