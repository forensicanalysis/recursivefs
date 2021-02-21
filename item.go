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

package recursivefs

import (
	"io"
	"io/fs"
	"path"
	"sort"

	"github.com/forensicanalysis/fslib"
)

// Item describes files and directories in the file system.
type Item struct {
	parentFS  fs.FS
	localPath string

	internal fs.File
	childFS  fs.FS

	dirOffset int
}

func (i *Item) Read(bytes []byte) (int, error) {
	return i.internal.Read(bytes)
}

func (i *Item) Close() error {
	return i.internal.Close()
}

// ReadDir returns up to n child items of a directory.
func (i *Item) ReadDir(n int) (entries []fs.DirEntry, err error) {
	if i.childFS != nil {
		entries, err = fs.ReadDir(i.childFS, ".")
		if err != nil {
			return nil, err
		}
		entries, err = recEntries(entries, ".", i.childFS)
		if err != nil {
			return nil, err
		}
	} else {
		entries, err = fslib.ReadDir(i.internal, -1)
		if err != nil {
			return nil, err
		}
		entries, err = recEntries(entries, i.localPath, i.parentFS)
		if err != nil {
			return nil, err
		}
	}

	entries, offset, err := dirEntries(n, entries, i.dirOffset)
	i.dirOffset += offset

	return entries, err
}

func recEntries(ditems []fs.DirEntry, p string, fsys fs.FS) (items []fs.DirEntry, err error) {
	for _, item := range ditems {
		info, err := item.Info()
		if err != nil {
			return nil, err
		}
		isFS := false
		if !item.IsDir() {
			f, err := fsys.Open(path.Join(p, item.Name()))
			if err != nil {
				return nil, err
			}
			cfsys, err := childFS(f, item.Name())
			if err != nil {
				return nil, err
			}

			isFS = cfsys != nil
		}

		items = append(items, &Info{info, isFS})
	}
	return items, nil
}

func dirEntries(n int, items []fs.DirEntry, dirOffset int) ([]fs.DirEntry, int, error) {
	sort.Slice(items, func(i, j int) bool {
		return items[i].Name() < items[j].Name()
	})

	// directory already exhausted
	if n <= 0 && dirOffset >= len(items) {
		return nil, 0, nil
	}

	var err error
	// read till end
	if n > 0 && dirOffset+n > len(items) {
		err = io.EOF
		if dirOffset > len(items) {
			return nil, 0, err
		}
	}

	offset := 0
	if n > 0 && dirOffset+n <= len(items) {
		items = items[dirOffset : dirOffset+n]
		offset += n
	} else {
		items = items[dirOffset:]
		offset += len(items)
	}

	return items, offset, err
}

// Stat return an fs.FileInfo object that describes a file.
func (i *Item) Stat() (fs.FileInfo, error) {
	info, err := i.internal.Stat()
	return &Info{info, i.childFS != nil}, err
}
