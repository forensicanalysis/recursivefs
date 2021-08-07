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

package main_test

import (
	"bytes"
	"io"
	"io/fs"
	"io/ioutil"
	"os"
	"reflect"
	"regexp"
	"testing"

	"github.com/spf13/cobra"

	"github.com/forensicanalysis/fscmd"
	"github.com/forensicanalysis/fslib"
	"github.com/forensicanalysis/recursivefs"
)

func stdout(f func()) []byte {
	old := os.Stdout // keep backup of the real stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	f()

	outC := make(chan []byte)
	// copy the output in a separate goroutine so printing can't block indefinitely
	go func() {
		var buf bytes.Buffer
		io.Copy(&buf, r) // nolint
		outC <- buf.Bytes()
	}()

	// back to normal state
	w.Close()
	os.Stdout = old // restoring the real stdout
	return <-outC

}

func Test_cat(t *testing.T) {
	b, _ := ioutil.ReadFile("../../testdata/data/document/Digital forensics.txt")

	type args struct {
		url string
	}
	tests := []struct {
		name     string
		args     args
		wantData []byte
	}{
		{"cat", args{"document/Digital forensics.txt"}, b},
		{"cat zip", args{"container/zip.zip/document/Digital forensics.txt"}, b},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotData := stdout(func() { fscmd.CatCmd(testParse)(nil, []string{"../../testdata/data/" + tt.args.url}) })

			re := regexp.MustCompile(`\r?\n`) // TODO: improve newline handling
			gotDataString := re.ReplaceAllString(string(gotData), "")
			wantData := re.ReplaceAllString(string(tt.wantData), "")

			if len(gotDataString) != len(wantData) {
				t.Errorf("cat() len = %d, want %d", len(gotData), len(tt.wantData))
			}

			if !reflect.DeepEqual(gotDataString, wantData) {
				t.Errorf("cat() = %s, want %s", gotData, tt.wantData)
				t.Errorf("cat() = %x, want %x", gotData, tt.wantData)
			}
		})
	}
}

func Test_ls(t *testing.T) {
	type args struct {
		url string
	}
	tests := []struct {
		name     string
		args     args
		wantData []byte
	}{
		{"ls", args{"container/zip.zip"}, []byte("README.md\ncontainer/\ndocument/\nevidence.json\nfolder/\nimage/\n")},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotData := stdout(func() { fscmd.LsCmd(testParse)(nil, []string{"../../testdata/data/" + tt.args.url}) })
			if !reflect.DeepEqual(string(gotData), string(tt.wantData)) {
				t.Errorf("ls() = %s, want %s", gotData, tt.wantData)
				t.Errorf("ls() = %x, want %x", gotData, tt.wantData)
			}
		})
	}
}

func Test_file(t *testing.T) {
	type args struct {
		url string
	}
	tests := []struct {
		name     string
		args     args
		wantData []byte
	}{
		{"file", args{"container/zip.zip"}, []byte(": application/zip\n")},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			name, _ := fslib.ToFSPath("../../testdata/data/" + tt.args.url)
			gotData := stdout(func() { fscmd.FileCmd(testParse)(nil, []string{"../../testdata/data/" + tt.args.url}) })
			if !reflect.DeepEqual(string(gotData), name+string(tt.wantData)) {
				t.Errorf("file() = %s, want %s", gotData, tt.wantData)
			}
		})
	}
}

func Test_hashsum(t *testing.T) {
	type args struct {
		url string
	}
	tests := []struct {
		name     string
		args     args
		wantData []byte
	}{
		{"hashsum", args{"container/zip.zip"}, []byte("MD5: 1d62df4bab8bb2ec2fefcf21cd509347\nSHA1: 880e3e47458ae264aebf2b42706ed0bac8831360\nSHA256: 82c38c2c6a5cb6b88d60c8de05bfea04ac16ac91b0e919786c5acf2f3bca2314\nSHA512: dde022a4c538bc802fa50aac473761aa3eaf965cf801136c736f4bbc89562423c9079a05da5de09c3f66d58c1f643a319d6bc33d8b1f6a9913fdff141a5c756f\n")},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotData := stdout(func() { fscmd.HashsumCmd(testParse)(nil, []string{"../../testdata/data/" + tt.args.url}) })
			if !reflect.DeepEqual(string(gotData), string(tt.wantData)) {
				t.Errorf("hashsum() = %s, want %s", gotData, tt.wantData)
			}
		})
	}
}

func Test_stat(t *testing.T) {
	result := `Name: image
Size: 0
IsDir: true
Mode: drwxr-xr-x
Modified: 2018-03-31 19:48:36 +0000 UTC
`
	type args struct {
		url string
	}
	tests := []struct {
		name     string
		args     args
		wantData []byte
	}{
		{"stat", args{"container/zip.zip/image"}, []byte(result)},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotData := stdout(func() { fscmd.StatCmd(testParse)(nil, []string{"../../testdata/data/" + tt.args.url}) })
			if !reflect.DeepEqual(string(gotData), string(tt.wantData)) {
				// t.Errorf("stat() = '%s', want '%s'", gotData, tt.wantData) // TODO https://github.com/golang/go/issues/43872
			}
		})
	}
}

func Test_tree(t *testing.T) {
	type args struct {
		url string
	}
	tests := []struct {
		name     string
		args     args
		wantData []byte
	}{
		{"tree", args{"container/zip.zip"}, []byte(`
├── README.md
├── container
│   ├── Computer forensics - Wikipedia.7z
│   ├── Computer forensics - Wikipedia.pdf.gz
│   ├── Computer forensics - Wikipedia.tar
│   │   └── Computer forensics - Wikipedia.pdf
│   └── Computer forensics - Wikipedia.zip
│       └── Computer forensics - Wikipedia.pdf
├── document
│   ├── Computer forensics - Wikipedia.pdf
│   ├── Design_of_the_FAT_file_system.xlsx
│   │   ├── [Content_Types].xml
│   │   ├── _rels
│   │   │   └── .rels
│   │   ├── docProps
│   │   │   ├── app.xml
│   │   │   └── core.xml
│   │   └── xl
│   │       ├── _rels
│   │       │   └── workbook.xml.rels
│   │       ├── printerSettings
│   │       │   └── printerSettings1.bin
│   │       ├── sharedStrings.xml
│   │       ├── styles.xml
│   │       ├── theme
│   │       │   └── theme1.xml
│   │       ├── workbook.xml
│   │       └── worksheets
│   │           ├── _rels
│   │           │   └── sheet1.xml.rels
│   │           └── sheet1.xml
│   ├── Digital forensics.docx
│   │   ├── [Content_Types].xml
│   │   ├── _rels
│   │   │   └── .rels
│   │   ├── docProps
│   │   │   ├── app.xml
│   │   │   └── core.xml
│   │   └── word
│   │       ├── _rels
│   │       │   └── document.xml.rels
│   │       ├── document.xml
│   │       ├── fontTable.xml
│   │       ├── media
│   │       │   └── image1.jpeg
│   │       ├── numbering.xml
│   │       ├── settings.xml
│   │       ├── styles.xml
│   │       ├── theme
│   │       │   └── theme1.xml
│   │       └── webSettings.xml
│   ├── Digital forensics.txt
│   └── NTFS.pptx
│       ├── [Content_Types].xml
│       ├── _rels
│       │   └── .rels
│       ├── docProps
│       │   ├── app.xml
│       │   ├── core.xml
│       │   └── thumbnail.jpeg
│       └── ppt
│           ├── _rels
│           │   └── presentation.xml.rels
│           ├── presProps.xml
│           ├── presentation.xml
│           ├── slideLayouts
│           │   ├── _rels
│           │   │   ├── slideLayout1.xml.rels
│           │   │   ├── slideLayout10.xml.rels
│           │   │   ├── slideLayout11.xml.rels
│           │   │   ├── slideLayout2.xml.rels
│           │   │   ├── slideLayout3.xml.rels
│           │   │   ├── slideLayout4.xml.rels
│           │   │   ├── slideLayout5.xml.rels
│           │   │   ├── slideLayout6.xml.rels
│           │   │   ├── slideLayout7.xml.rels
│           │   │   ├── slideLayout8.xml.rels
│           │   │   └── slideLayout9.xml.rels
│           │   ├── slideLayout1.xml
│           │   ├── slideLayout10.xml
│           │   ├── slideLayout11.xml
│           │   ├── slideLayout2.xml
│           │   ├── slideLayout3.xml
│           │   ├── slideLayout4.xml
│           │   ├── slideLayout5.xml
│           │   ├── slideLayout6.xml
│           │   ├── slideLayout7.xml
│           │   ├── slideLayout8.xml
│           │   └── slideLayout9.xml
│           ├── slideMasters
│           │   ├── _rels
│           │   │   └── slideMaster1.xml.rels
│           │   └── slideMaster1.xml
│           ├── slides
│           │   ├── _rels
│           │   │   ├── slide1.xml.rels
│           │   │   └── slide2.xml.rels
│           │   ├── slide1.xml
│           │   └── slide2.xml
│           ├── tableStyles.xml
│           ├── theme
│           │   └── theme1.xml
│           └── viewProps.xml
├── evidence.json
├── folder
│   ├── file.txt
│   └── subfolder
│       ├── subfile.txt
│       └── subsubfolder
│           └── subsubfile.txt
└── image
    ├── alps.jpg
    ├── alps.png
    └── alps.tiff
`)},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			name, _ := fslib.ToFSPath("../../testdata/data/" + tt.args.url)
			gotData := stdout(func() { fscmd.TreeCmd(testParse)(nil, []string{"../../testdata/data/" + tt.args.url}) })
			if !reflect.DeepEqual(string(gotData), name + string(tt.wantData)) {
				t.Errorf("tree() = '%s', want '%s'", gotData, tt.wantData)
				t.Errorf("tree() = '%x', want '%x'", gotData, tt.wantData)
			}
		})
	}
}

func testParse(_ *cobra.Command, args []string) (fs.FS, []string, error) {
	var names []string
	for _, arg := range args {
		name, err := fslib.ToFSPath(arg)
		if err != nil {
			return nil, nil, err
		}
		names = append(names, name)
	}
	return recursivefs.New(), names, nil
}
