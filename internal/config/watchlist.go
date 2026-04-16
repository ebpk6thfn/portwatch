package config

import (
	"fmt"

	"github.com/user/portwatch/internal/portscanner"
)

// WatchlistEntry is the serialisable form of a watchlist rule used in config
// files. Action must be either "alert" or "ignore".
type WatchlistEntry struct {
	Port   uint16 `toml:"port"   json:"port"`
	Action string `toml:"action" json:"action"`
}

// ToRule converts a WatchlistEntry to a portscanner.WatchlistRule.
// An error is returned when the action string is not recognised.
func (e WatchlistEntry) ToRule() (portscanner.WatchlistRule, error) {
	switch portscanner.WatchlistAction(e.Action) {
	case portscanner.ActionAlert, portscanner.ActionIgnore:
		return portscanner.WatchlistRule{
			Port:   e.Port,
			Action: portscanner.WatchlistAction(e.Action),
		}, nil
	default:
		return portscanner.WatchlistRule{}, fmt.Errorf(
			"watchlist: unknown action %q for port %d (must be \"alert\" or \"ignore\")",
			e.Action, e.Port,
		)
	}
}

// BuildWatchlist converts a slice of WatchlistEntry values from config into a
// *portscanner.Watchlist. The first conversion error causes an early return.
func BuildWatchlist(entries []WatchlistEntry) (*portscanner.Watchlist, error) {
	rules := make([]portscanner.WatchlistRule, 0, len(entries))
	for _, e := range entries {
		r, err := e.ToRule()
		if err != nil {
			return nil, err
		}
		rules = append(rules, r)
	}
	return portscanner.NewWatchlist(rules), nil
}
