package procpidmaps

import "fmt"

type FileInfo struct {
	Path string   // location to possible file, can be empty for non-files
	Type FileType // type of path
}

func (fi FileInfo) String() string {
	return fmt.Sprintf(`%q (%v)`, fi.Path, fi.Type)
}
