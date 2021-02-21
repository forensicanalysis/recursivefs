<h1 align="center">recursive fs</h1>

<p  align="center">
 <a href="https://godocs.io/github.com/forensicanalysis/recursivefs"><img src="https://godocs.io/github.com/forensicanalysis/recursivefs?status.svg" alt="doc" /></a>
</p>

A recursive file system that processes container files according to their file type. 
You can use it e.g. to read a pdf from a zip file on an NTFS disk image (s. below). 
It also provides the `fs` command line tool do use the functionality from the command line.
recursivefs implements [io/fs.FS](https://tip.golang.org/pkg/io/fs).


## Example

``` golang
func main() {
	// Read the pdf header from a zip file on an NTFS disk image.

	// parse the file system
	fsys := recursivefs.New()

	// create fslib path
	wd, _ := os.Getwd()
	fpath, _ := fslib.ToFSPath(path.Join(wd, "testdata/data/filesystem/ntfs.dd/container/Computer forensics - Wikipedia.zip/Computer forensics - Wikipedia.pdf"))

	// get handle the README.md
	file, err := fsys.Open(fpath)
	if err != nil {
		panic(err)
	}

	// get content
	content, _ := io.ReadAll(file)

	// print content
	fmt.Println(string(content[0:4]))
	// Output: %PDF
}
```

---

## The fs command

The fs command line tool that has various subcommands which imitate unix commands
but for nested file system structures.

- **fs cat**: Print files
- **fs file**: Determine files types
- **fs hashsum**: Print hashsums
- **fs ls**: List directory contents
- **fs stat**: Display file status
- **fs tree**: List contents of directories in a tree-like format


#### Download

https://github.com/forensicanalysis/recursivefs/releases

#### Usage Examples

List all files in a zip file:
```
fs ls test.zip
```

Extract the Amcache.hve file from a NTFS image in a zip file:

```
fs cat case/evidence.zip/ntfs.dd/Windows/AppCompat/Programs/Amcache.hve > Amcache.hve
```

Hash all files in a zip file:
```
fs hashsum case/evidence.zip/*
```

