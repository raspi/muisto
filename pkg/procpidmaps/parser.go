package procpidmaps

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
)

// ParseMapFile parses `/proc/<pid>/maps` file to proper structures
func ParseMapFile(pid int) (maps []MemoryMap, err error) {
	f, err := os.Open(fmt.Sprintf(`/proc/%d/maps`, pid))
	if err != nil {
		return maps, err
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)

	// See: `man 5 proc` for file format
	for scanner.Scan() {

		info := MemoryMap{
			pid: pid,
		}

		// Read a line and split it into parts
		for idx, l := range strings.SplitN(scanner.Text(), ` `, 6) {
			switch idx {
			case 0: // address
				info.AddressStart, info.AddressEnd = readAddressRange(l)
			case 1: // perms
				info.Permissions = readPermissions(l)
			case 2: // offset (to inode)
				info.InodeOffset = hex2Int(l)
			case 3: // device
				info.Device = l
			case 4: // inode
				info.Inode = hex2Int(l)
			case 5: // filepath
				l = strings.TrimLeft(l, ` `)
				fi := FileInfo{
					Path: l,
				}

				if strings.HasPrefix(l, `[`) && strings.HasSuffix(l, `]`) {
					fi.Type = Special

					switch l {
					case `[stack]`:
						fi.Type = Stack
					case `[heap]`:
						fi.Type = Heap
					case `[vdso]`:
						fi.Type = VirtualDynamicallyLinkedSharedObject
					case `[vvar]`:
						fi.Type = VirtualVariable
					}

				} else if l == `` {
					fi.Type = EmptyPath
				} else {
					fi.Type = Path
				}

				info.Path = fi
			default:
				continue
			}
		}

		// Add to list
		maps = append(maps, info)
	}

	return maps, nil
}

func hex2Int(line string) int64 {
	start, err := strconv.ParseInt(line, 16, 0)
	if err != nil {
		panic(err)
	}

	return start
}

func readAddressRange(line string) (int64, int64) {
	addresses := strings.SplitN(line, `-`, 2)

	l := len(addresses)
	if l != 2 {
		panic(fmt.Errorf(`invalid length: %d for %q`, l, line))
	}

	addresses[0] = strings.TrimLeft(addresses[0], "0")
	addresses[1] = strings.TrimLeft(addresses[1], "0")

	return hex2Int(addresses[0]), hex2Int(addresses[1])
}
