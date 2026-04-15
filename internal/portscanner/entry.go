package portscanner

import "fmt"

// Protocol represents the network protocol of a port entry.
type Protocol string

const (
	ProtoTCP  Protocol = "tcp"
	ProtoTCP6 Protocol = "tcp6"
	ProtoUDP  Protocol = "udp"
	ProtoUDP6 Protocol = "udp6"
)

// Entry represents a single listening port captured during a scan.
type Entry struct {
	// LocalAddr is the IP address the port is bound to (e.g. "0.0.0.0", "127.0.0.1").
	LocalAddr string `json:"local_addr"`

	// Port is the port number.
	Port uint16 `json:"port"`

	// Protocol is the network protocol (tcp, udp, etc.).
	Protocol Protocol `json:"protocol"`

	// PID is the owning process ID, if resolvable; 0 if unknown.
	PID int `json:"pid,omitempty"`

	// ProcessName is the human-readable process name, if resolvable.
	ProcessName string `json:"process_name,omitempty"`
}

// Key returns a unique string key for the entry, suitable for use in maps.
func (e Entry) Key() string {
	return fmt.Sprintf("%s:%s:%d", e.Protocol, e.LocalAddr, e.Port)
}

// String returns a human-readable representation of the entry.
func (e Entry) String() string {
	if e.ProcessName != "" {
		return fmt.Sprintf("%s %s:%d (pid=%d, %s)", e.Protocol, e.LocalAddr, e.Port, e.PID, e.ProcessName)
	}
	return fmt.Sprintf("%s %s:%d", e.Protocol, e.LocalAddr, e.Port)
}

// ChangeEvent describes a detected change between two scans.
type ChangeEvent struct {
	// Type is either "opened" or "closed".
	Type string `json:"type"`

	// Entry is the port entry that changed.
	Entry Entry `json:"entry"`
}
