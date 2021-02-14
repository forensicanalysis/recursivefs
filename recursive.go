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
	"errors"
	"io"
	"io/fs"
	"os"
	"path"
	"strings"

	"github.com/forensicanalysis/fslib/bufferfs"
	"github.com/forensicanalysis/fslib/fat16"
	"github.com/forensicanalysis/fslib/fsio"
	"github.com/forensicanalysis/fslib/gpt"
	"github.com/forensicanalysis/fslib/mbr"
	"github.com/forensicanalysis/fslib/ntfs"
	"github.com/forensicanalysis/recursivefs/filetype"
	"github.com/forensicanalysis/zipfs"
	"github.com/nlepage/go-tarfs"
)

func parseRealPath(fsys fs.FS, sample string) (rpath []element, err error) {
	parts := strings.Split(sample, "/")

	if len(parts) == 0 {
		return []element{{fsys, "."}}, nil
	}

	key := "."
	for len(parts) > 0 {
		key = path.Join(key, parts[0])
		parts = parts[1:]
		info, err := fs.Stat(fsys, key)
		if err != nil {
			return nil, err
		}

		if !info.IsDir() {
			rpath = append(rpath, element{fsys, key})
			f, err := fsys.Open(key)
			if err != nil {
				return nil, err
			}
			cfsys, err := childFS(f, key)
			if err != nil || cfsys == nil {
				continue
			}
			fsys = cfsys

			key = "."
		} else if len(parts) == 0 {
			rpath = append(rpath, element{fsys, key})
		}
	}
	return rpath, nil
}

// func childFS(fsys fs.FS, name string) (fs.FS, error) {
func childFS(r io.Reader, name string) (fs.FS, error) {
	t, err := filetype.DetectReaderByExtension(r, path.Ext(name))
	if err != nil && err != io.EOF {
		return nil, err
	}

	readSeekerAt, ok := r.(fsio.ReadSeekerAt)
	if !ok {
		return nil, errors.New("files must be ReadSeekerAt")
	}
	_, _ = readSeekerAt.Seek(0, os.SEEK_SET)

	switch t {
	case filetype.Zip, filetype.Xlsx, filetype.Pptx, filetype.Docx:
		zipfsys, err := zipfs.New(readSeekerAt)
		if err != nil {
			return nil, err
		}
		return bufferfs.New(zipfsys), nil
	case filetype.Tar:
		tarfsys, err := tarfs.New(readSeekerAt)
		if err != nil {
			return nil, err
		}
		return bufferfs.New(tarfsys), nil
	case filetype.FAT16:
		return fat16.New(readSeekerAt)
	case filetype.MBR:
		return mbr.New(readSeekerAt)
	case filetype.GPT:
		return gpt.New(readSeekerAt)
	case filetype.NTFS:
		ntfsys, err := ntfs.New(readSeekerAt)
		if err != nil {
			return nil, err
		}
		return bufferfs.New(ntfsys), nil
	default:
		return nil, nil
	}
}
