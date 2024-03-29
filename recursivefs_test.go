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
	"fmt"
	"io/fs"
	"log"
	"path"
	"reflect"
	"testing"

	"github.com/forensicanalysis/fslib"
	"github.com/forensicanalysis/fslib/bufferfs"
	"github.com/forensicanalysis/fslib/fat16"
	fslibtest "github.com/forensicanalysis/fslib/fstest"
	"github.com/forensicanalysis/fslib/osfs"
)

/*
func TestFS(t *testing.T) {
	osfsys := os.DirFS("testdata/data/document")

	fsys := NewFS(osfsys)

	err := fstest.TestFS(fsys, "Design_of_the_FAT_file_system.xlsx/[Content_Types].xml")
	if err != nil {
		t.Fatal(err)
	}
}
*/

func TestRecursiveFS_OpenRead(t *testing.T) {
	type args struct {
		name string
	}
	tests := []struct {
		name    string
		args    args
		want    []byte
		wantErr bool
	}{
		{"Test zip", args{"testdata/data/container/zip.zip"}, []byte{0x50, 0x4B, 0x03, 0x04, 0x14}, false},
		// {"Test tar", args{"testdata/data/container/tar.tar"}, []byte("READM"), false},
		{"Test 7z", args{"testdata/data/container/7z.7z"}, []byte{0x37, 0x7A, 0xBC, 0xAF, 0x27}, false},
		{"Test deep text", args{"testdata/data/filesystem/mbr_fat16.dd/p0/README.MD"}, []byte("# :ma"), false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := New()
			name, err := fslib.ToFSPath(tt.args.name)
			if err != nil {
				t.Error(err)
				return
			}
			gotF, err := m.Open(name)
			if (err != nil) != tt.wantErr {
				t.Errorf("FS.Open() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotF == nil {
				return
			}
			head := make([]byte, 5)
			_, err = gotF.Read(head)
			if (err != nil) != tt.wantErr {
				t.Errorf("FS.Read() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !reflect.DeepEqual(head, tt.want) {
				t.Errorf("FS.Open() = %v, want %v", head, tt.want)
			}
		})
	}
}

func TestRecursiveFS_OpenDirList(t *testing.T) {
	type args struct {
		name string
	}
	tests := []struct {
		name    string
		args    args
		want    []string
		wantErr bool
	}{
		// {"Test os", args{"testdata/data/filesystem/"}, []string{"7z.7z", "zip.zip"}, false},
		// {"Test tar", args{"testdata/data/container/tar.tar"}, []string{"README.md", "container", "document", "evidence.json", "folder", "image"}, false},
		{"Test zip", args{"testdata/data/container/zip.zip"}, []string{"README.md", "container", "document", "evidence.json", "folder", "image"}, false},
		{"Test deep text", args{"testdata/data/filesystem/mbr_fat16.dd/"}, []string{"p0"}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := New()
			name, err := fslib.ToFSPath(tt.args.name)
			if err != nil {
				t.Error(err)
				return
			}
			gotF, err := m.Open(name)
			if (err != nil) != tt.wantErr {
				t.Errorf("FS.Open() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotF == nil {
				return
			}
			names, err := fslibtest.Readdirnames(gotF, 0)
			if (err != nil) != tt.wantErr {
				t.Errorf("FS.Open() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(names, tt.want) {
				t.Errorf("FS.Open() = %v, want %v", names, tt.want)
			}
		})
	}
}

func TestRecursiveFS_Readdir(t *testing.T) {
	type args struct {
		name string
	}

	containerFiles := map[string]bool{
		"README.md":     false,
		"container":     true,
		"document":      true,
		"evidence.json": false,
		"folder":        true,
		"image":         true,
	}

	tests := []struct {
		name    string
		args    args
		want    map[string]bool
		wantErr bool
	}{
		{"Test folder", args{"testdata/data/container/zip.zip/"}, containerFiles, false},
		{"Test tar folder", args{"testdata/data/container/tar.tar/"}, containerFiles, false},
		{"Test zip", args{"testdata/data/container/zip.zip/container/"}, map[string]bool{"Computer forensics - Wikipedia.zip": true, "Computer forensics - Wikipedia.tar": true}, false}, // TODO fix
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := New()
			name, err := fslib.ToFSPath(tt.args.name)
			if err != nil {
				t.Error(err)
				return
			}
			gotF, err := m.Open(name)
			if (err != nil) != tt.wantErr {
				t.Errorf("FS.Open() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotF == nil {
				return
			}
			fileNames, err := fslibtest.Readdirnames(gotF, 0)
			if (err != nil) != tt.wantErr {
				t.Errorf("FS.Open() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			for _, fileName := range fileNames {
				fi, err := fs.Stat(m, path.Join(name, fileName))
				if err != nil {
					t.Errorf("FS.Stat() error = %v", err)
					return
				}
				if !reflect.DeepEqual(fi.IsDir(), tt.want[fileName]) {
					t.Errorf("FS %s IsDir = %v, want %v", fileName, fi.IsDir(), tt.want[fileName])
				}
			}
		})
	}
}

func TestParseRealPath(t *testing.T) {
	zippath, err := fslib.ToFSPath("testdata/data/container/zip.zip")
	if err != nil {
		t.Error(err)
		return
	}
	fatpath, err := fslib.ToFSPath("testdata/data/filesystem/mbr_fat16.dd")
	if err != nil {
		t.Error(err)
		return
	}

	type args struct {
		sample string
	}
	tests := []struct {
		name      string
		args      args
		wantRpath []element
		wantErr   bool
	}{
		{"Test zip", args{"testdata/data/container/zip.zip/image"}, []element{{&osfs.FS{}, zippath}, {&bufferfs.FS{}, "image"}}, false},
		{"Test fat16", args{"testdata/data/filesystem/mbr_fat16.dd/p0/IMAGE"}, []element{{&osfs.FS{}, fatpath}, {&bufferfs.FS{}, "p0"}, {&fat16.FS{}, "IMAGE"}}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			name, err := fslib.ToFSPath(tt.args.sample)
			if err != nil {
				t.Error(err)
				return
			}
			gotRpath, err := parseRealPath(osfs.New(), name)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseRealPath() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !elementsEqual(gotRpath, tt.wantRpath) {
				t.Errorf("ParseRealPath() = %#v, want %#v", gotRpath, tt.wantRpath)
			}
		})
	}
}

func elementsEqual(rpath []element, rpath2 []element) bool {
	if len(rpath) != len(rpath2) {
		return false
	}
	for i := range rpath {
		if rpath[i].Key != rpath2[i].Key {
			log.Println(rpath[i].Key, rpath2[i].Key)
			return false
		}
		if fmt.Sprintf("%T", rpath[i].FS) != fmt.Sprintf("%T", rpath2[i].FS) {
			log.Println(fmt.Sprintf("%T", rpath[i].FS), fmt.Sprintf("%T", rpath2[i].FS))
			return false
		}
	}
	return true
}
