package portscanner

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
)

// Resolver maps process IDs to human-readable process names by reading
// /proc/<pid>/comm and correlating socket inodes via /proc/<pid>/fd.
type Resolver struct {
	procRoot string
}

// NewResolver creates a Resolver. procRoot is typically "/proc".
func NewResolver(procRoot string) *Resolver {
	if procRoot == "" {
		procRoot = "/proc"
	}
	return &Resolver{procRoot: procRoot}
}

// InodeToProcess returns the PID and process name owning the given socket
// inode, or empty strings if not found.
func (r *Resolver) InodeToProcess(inode uint64) (pid int, name string) {
	entries, err := os.ReadDir(r.procRoot)
	if err != nil {
		return 0, ""
	}
	target := fmt.Sprintf("socket:[%d]", inode)
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		p, err := strconv.Atoi(e.Name())
		if err != nil {
			continue
		}
		fdDir := fmt.Sprintf("%s/%d/fd", r.procRoot, p)
		links, err := os.ReadDir(fdDir)
		if err != nil {
			continue
		}
		for _, fd := range links {
			link, err := os.Readlink(fmt.Sprintf("%s/%s", fdDir, fd.Name()))
			if err != nil {
				continue
			}
			if link == target {
				commPath := fmt.Sprintf("%s/%d/comm", r.procRoot, p)
				n, err := readComm(commPath)
				if err != nil {
					n = fmt.Sprintf("pid:%d", p)
				}
				return p, n
			}
		}
	}
	return 0, ""
}

func readComm(path string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer f.Close()
	scanner := bufio.NewScanner(f)
	if scanner.Scan() {
		return strings.TrimSpace(scanner.Text()), nil
	}
	return "", fmt.Errorf("empty comm file")
}
