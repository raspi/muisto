package procpidmaps

import "fmt"

type MemoryMap struct {
	pid          int   // set internally, use GetPID()
	AddressStart int64 // Address space starting offset
	AddressEnd   int64 // Address space ending offset
	Permissions  Permissions
	Inode        int64    // Possible inode to open handle
	InodeOffset  int64    // Possible inode offset to open handle
	Device       string   // Possible device address
	Path         FileInfo // a path to possible open file handle
}

func (i MemoryMap) Size() int64 {
	return i.AddressEnd - i.AddressStart
}

func (i MemoryMap) String() string {
	return fmt.Sprintf(`start:%d end:%d size:%d perms:[%v] offset:%v dev:%v inode:%d path:%v`,
		i.AddressStart, i.AddressEnd, i.Size(), i.Permissions, i.InodeOffset, i.Device, i.Inode, i.Path)
}

func (i MemoryMap) GetPID() int {
	return i.pid
}
