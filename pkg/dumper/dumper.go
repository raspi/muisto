package dumper

import (
	"fmt"
	"github.com/raspi/muisto/pkg/procpidmaps"
	"io"
	"io/ioutil"
	"os"
	"path"
)

type MemDump struct {
	l   []procpidmaps.MemoryMap
	pid int
}

func New(pid int) (m MemDump, err error) {
	m = MemDump{
		pid: pid,
	}

	m.l, err = procpidmaps.ParseMapFile(m.pid)
	if err != nil {
		return m, err
	}

	return m, nil
}

func (m *MemDump) GetList() []procpidmaps.MemoryMap {
	return m.l
}

func (m *MemDump) Dump(i procpidmaps.MemoryMap) (newFilePath string, err error) {
	if !i.Permissions.Read {
		return ``, fmt.Errorf(`no read permissions`)
	}

	f, err := os.Open(fmt.Sprintf(`/proc/%d/mem`, i.GetPID()))
	if err != nil {
		return ``, err
	}
	defer f.Close()

	_, err = f.Seek(i.AddressStart, io.SeekStart)
	if err != nil {
		return ``, fmt.Errorf(`seek error start:%d device:%s inode:%d path:%v`, i.AddressStart, i.Device, i.Inode, i.Path)
	}

	buffer := make([]byte, i.Size())

	_, err = f.Read(buffer)
	if err != nil {
		return ``, fmt.Errorf(`read error start:%d device:%s inode:%d path:%v`, i.AddressStart, i.Device, i.Inode, i.Path)
	}

	tmpf, err := ioutil.TempFile(`.`, `.dump-`)
	if err != nil {
		return ``, err
	}

	// Write contents to file
	_, err = tmpf.Write(buffer)
	if err != nil {
		return ``, err
	}

	err = tmpf.Close()
	if err != nil {
		return ``, err
	}

	fname := fmt.Sprintf(`%d-%s-%d.dump`, m.pid, i.Path.Type, i.AddressStart)

	newFilePath = path.Join(`.`, fname)

	// Rename temporary file to proper filename
	err = os.Rename(tmpf.Name(), newFilePath)
	if err != nil {
		return ``, err
	}

	return newFilePath, nil
}
