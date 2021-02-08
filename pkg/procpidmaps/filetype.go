package procpidmaps

type FileType uint8

const (
	EmptyPath FileType = iota
	Path
	Special
	Stack
	Heap
	VirtualDynamicallyLinkedSharedObject
	VirtualVariable
)

func (ft FileType) String() string {
	switch ft {
	case Path:
		return `path`
	case EmptyPath:
		return `emptypath`
	case Special:
		return `special`
	case Stack:
		return `stack`
	case Heap:
		return `heap`
	case VirtualDynamicallyLinkedSharedObject:
		return `vdso`
	case VirtualVariable:
		return `vvar`
	default:
		return `???`
	}
}
