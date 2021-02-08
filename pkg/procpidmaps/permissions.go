package procpidmaps

import "fmt"

// Permissions for given memory address space
type Permissions struct {
	Read    bool
	Write   bool
	Execute bool
	Shared  bool
	Private bool
}

func (p Permissions) String() string {
	return fmt.Sprintf(`read:%v write:%v shared:%v private:%v`,
		p.Read, p.Write, p.Shared, p.Private)
}

func readPermissions(l string) (p Permissions) {
	for _, c := range l {
		switch c {
		case 'r':
			p.Read = true
		case 'w':
			p.Write = true
		case 'x':
			p.Execute = true
		case 's':
			p.Shared = true
		case 'p':
			p.Private = true
		}
	}

	return p
}
