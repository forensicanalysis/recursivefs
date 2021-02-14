module github.com/forensicanalysis/recursivefs

go 1.16

require (
	github.com/forensicanalysis/fslib v0.14.10-0.20210130143731-88588c4c3f19
	github.com/forensicanalysis/zipfs v0.0.0-20210202205655-81fcfd84e063
	github.com/h2non/filetype v1.1.1
	github.com/nlepage/go-tarfs v1.0.4
	github.com/spf13/cobra v1.1.1
	github.com/xlab/treeprint v1.0.0
)

replace github.com/forensicanalysis/fslib => ../fslib
