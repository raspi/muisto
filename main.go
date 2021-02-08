package main

import (
	"flag"
	"fmt"
	"github.com/raspi/muisto/pkg/dumper"
	"github.com/raspi/muisto/pkg/procpidmaps"
	"github.com/raspi/muisto/pkg/units"
	"log"
	"os"
	"strconv"
	"strings"
	"syscall"
)

var (
	// These are set with Makefile -X=main.VERSION, etc
	VERSION   = `v0.0.0`
	BUILD     = `dev`
	BUILDDATE = `0000-00-00T00:00:00+00:00`
)

const (
	AUTHOR   = `Pekka Järvinen`
	HOMEPAGE = `https://github.com/raspi/muisto`
	YEAR     = 2021
)

func main() {
	versionArg := flag.Bool(`version`, false, `Show version information`)
	signalArg := flag.Bool(`stop`, true, `Stop process before dumping`)

	maxAddressArg := flag.String(`maxaddress`, "0", `Max address start offset (0 = no limit)`)
	minAddressArg := flag.String(`minaddress`, "0", `Min address start offset`)
	maxSizeArg := flag.String(`maxsize`, "100MiB", `Max address size (0 = no limit)`)
	minSizeArg := flag.String(`minsize`, "1KiB", `Min address size`)

	pidArg := flag.Int(`pid`, 0, `Program ID (PID)`)

	flag.Usage = func() {
		f := os.Args[0]
		_, _ = fmt.Fprintf(os.Stdout, `muisto - process memory address space dumper %v (%v)`+"\n", VERSION, BUILDDATE)
		_, _ = fmt.Fprintf(os.Stdout, `(c) %v %v- [ %v ]`+"\n", AUTHOR, YEAR, HOMEPAGE)
		_, _ = fmt.Fprintf(os.Stdout, "\n")

		_, _ = fmt.Fprintf(os.Stdout, "Parameters:\n")

		// Calculate padding
		paramMaxLen := 0
		flag.VisitAll(func(f *flag.Flag) {
			l := len(f.Name)
			if l > paramMaxLen {
				paramMaxLen = l
			}
		})

		flag.VisitAll(func(f *flag.Flag) {
			padding := strings.Repeat(` `, paramMaxLen-len(f.Name))
			_, _ = fmt.Fprintf(os.Stdout, "  -%s%s   %s   default: %q\n", f.Name, padding, f.Usage, f.DefValue)
		})

		_, _ = fmt.Fprintf(os.Stdout, "\n")

		_, _ = fmt.Fprintf(os.Stdout, "Examples:\n")
		_, _ = fmt.Fprintf(os.Stdout, "  Dump addresses which has size between 8 KiB - 100 MiB at address offsets between 512 MiB - 1 GiB:\n")
		_, _ = fmt.Fprintf(os.Stdout, `    %v -pid 4321 -minaddress 512MiB -maxaddress 1GiB -maxsize 100MiB -minsize 8KiB`+"\n", f)

		_, _ = fmt.Fprintf(os.Stdout, "\n")

		_, _ = fmt.Fprintf(os.Stdout, "See:\n")
		_, _ = fmt.Fprintf(os.Stdout, "  `man 5 proc`, `cat /proc/<pid>/maps`\n")

		_, _ = fmt.Fprintf(os.Stdout, "\n")

	}

	flag.Parse()

	if *versionArg {
		_, _ = fmt.Fprintf(os.Stdout, `Version %v %v %v`+"\n", VERSION, BUILD, BUILDDATE)
		return
	}

	if *pidArg == 0 {
		_, _ = fmt.Fprintf(os.Stderr, `No PID given. See --help`+"\n")
		os.Exit(1)
	}

	pid := *pidArg

	proc, err := os.FindProcess(pid)
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, `invalid PID %q`, pid)
		os.Exit(1)
	}

	maxSize, err := units.Parse(*maxSizeArg)
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, `error: %v`, err)
		os.Exit(1)
	}

	minSize, err := units.Parse(*minSizeArg)
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, `error: %v`, err)
		os.Exit(1)
	}

	maxAddress, err := units.Parse(*maxAddressArg)
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, `error: %v`, err)
		os.Exit(1)
	}

	minAddress, err := units.Parse(*minAddressArg)
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, `error: %v`, err)
		os.Exit(1)
	}

	if *signalArg {
		// Send SIGSTOP to halt the process
		err = proc.Signal(syscall.SIGSTOP)
		if err != nil {
			_, _ = fmt.Fprintf(os.Stderr, `couldn't send SIGSTOP to PID %q`, pid)
			os.Exit(1)
		}

		// Continue process after memory addresses have been dumped
		defer proc.Signal(syscall.SIGCONT)
	}

	r, err := dumper.New(pid)
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, `error: %v`, err)
		os.Exit(1)
	}

	// Iterate through memory space addresses
	for _, i := range r.GetList() {
		if i.Permissions.Read != true {
			// Non-readable
			continue
		}

		if i.Path.Type == procpidmaps.Stack {
			continue
		}

		if i.Path.Type == procpidmaps.Heap {
			continue
		}

		if i.Path.Type == procpidmaps.VirtualVariable {
			continue
		}

		if i.Path.Type == procpidmaps.VirtualDynamicallyLinkedSharedObject {
			continue
		}

		if i.Path.Type == procpidmaps.Path {
			// Files can be read directly from the filesystem, so no sense to dump them
			continue
		}

		if maxAddress != 0 && i.AddressStart <= maxAddress {
			log.Printf(`Skipped: (max address offset %d) %v`, maxAddress, i)
			continue
		}

		if minAddress > 0 && i.AddressStart <= minAddress {
			log.Printf(`Skipped: (min address offset %d) %v`, minAddress, i)
			continue
		}

		if maxSize != 0 && i.Size() <= maxSize {
			log.Printf(`Skipped: (max size %d) %v`, maxSize, i)
			continue
		}

		if minSize > 0 && i.Size() <= minSize {
			log.Printf(`Skipped: (min size %d) %v`, minSize, i)
			continue
		}

		// Dump address to file
		fname, err := r.Dump(i)
		if err != nil {
			_, _ = fmt.Fprintf(os.Stderr, `error: %v`, err)
			continue // non-fatal
		}

		if os.Getuid() == 0 {
			// We're root, try to ease life so that files are owned by proper user that ran this app

			uid, err := strconv.Atoi(os.Getenv("SUDO_UID"))
			if err != nil {
				_, _ = fmt.Fprintf(os.Stderr, `error: %v`, err)
				os.Exit(1)
			}

			gid, err := strconv.Atoi(os.Getenv("SUDO_GID"))
			if err != nil {
				_, _ = fmt.Fprintf(os.Stderr, `error: %v`, err)
				os.Exit(1)
			}

			// Change owner and group
			err = os.Chown(fname, uid, gid)
			if err != nil {
				_, _ = fmt.Fprintf(os.Stderr, `error: %v`, err)
				continue // non-fatal
			}
		}

		log.Printf(`Dumped address 0x%x size %d byte(s) to %q path:%v`, i.AddressStart, i.Size(), fname, i.Path)
	}

	log.Printf(`Done.`)
}
