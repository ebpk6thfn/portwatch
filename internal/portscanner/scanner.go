package portscanner

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
)

// PortEntry represents a single open port on the host.
type PortEntry struct {
	Protocol string
	LocalAddr string
	Port     int
	PID      int
	State    string
}

// Scanner reads current open ports from the OS.
type Scanner struct{}

// NewScanner creates a new Scanner instance.
func NewScanner() *Scanner {
	return &Scanner{}
}

// Scan returns a map of port number to PortEntry for all currently open ports.
func (s *Scanner) Scan() (map[int]PortEntry, error) {
	result := make(map[int]PortEntry)

	for _, proto := range []string{"tcp", "tcp6", "udp", "udp6"} {
		path := fmt.Sprintf("/proc/net/%s", proto)
		entries, err := parseProcNet(path, proto)
		if err != nil {
			// Not all systems have all files; skip missing ones.
			continue
		}
		for _, e := range entries {
			result[e.Port] = e
		}
	}

	return result, nil
}

// parseProcNet parses a /proc/net/{tcp,udp} file and returns port entries.
func parseProcNet(path, proto string) ([]PortEntry, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var entries []PortEntry
	scanner := bufio.NewScanner(f)
	// Skip header line.
	scanner.Scan()

	for scanner.Scan() {
		fields := strings.Fields(scanner.Text())
		if len(fields) < 4 {
			continue
		}
		localAddr := fields[1]
		state := fields[3]

		// Only include LISTEN (0A) for TCP, or all for UDP.
		if strings.HasPrefix(proto, "tcp") && state != "0A" {
			continue
		}

		port, err := parsePort(localAddr)
		if err != nil {
			continue
		}

		entries = append(entries, PortEntry{
			Protocol:  proto,
			LocalAddr: localAddr,
			Port:      port,
			State:     state,
		})
	}

	return entries, scanner.Err()
}

// parsePort extracts the decimal port from a hex "addr:port" string.
func parsePort(addrPort string) (int, error) {
	parts := strings.Split(addrPort, ":")
	if len(parts) != 2 {
		return 0, fmt.Errorf("invalid addr:port %q", addrPort)
	}
	port64, err := strconv.ParseInt(parts[1], 16, 32)
	if err != nil {
		return 0, err
	}
	return int(port64), nil
}
