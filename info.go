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
	"io/fs"
	"time"
)

// Info wraps the fs.FileInfo.
type Info struct {
	internal fs.FileInfo
	isFS     bool
}

func (m *Info) Type() fs.FileMode {
	return m.Mode() & fs.ModeType
}

func (m *Info) Info() (fs.FileInfo, error) {
	return m, nil
}

func (m *Info) Name() string {
	return m.internal.Name()
}

func (m *Info) Size() int64 {
	return m.internal.Size()
}

func (m *Info) Mode() fs.FileMode {
	if m.IsDir() {
		// return fs.ModeDir
		return m.internal.Mode() | fs.ModeDir
	}
	// return 0
	return m.internal.Mode()
}

func (m *Info) ModTime() time.Time {
	// return time.Date(2020, 1, 1, 12, 12, 12, 12, time.UTC)
	return m.internal.ModTime()
}

func (m *Info) Sys() interface{} {
	return m.internal.Sys()
}

// IsDir returns if the item is a directory. Returns true for files that are file
// systems (e.g. zip archives).
func (m *Info) IsDir() bool {
	if m.isFS {
		return true
	}
	return m.internal.IsDir()
}
