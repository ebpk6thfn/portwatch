package daemon

import (
	"context"
	"log"
	"time"

	"github.com/user/portwatch/internal/config"
	"github.com/user/portwatch/internal/notifier"
	"github.com/user/portwatch/internal/portscanner"
)

// Daemon orchestrates periodic port scanning, diffing, rate-limiting,
// and notification dispatch.
type Daemon struct {
	cfg         *config.Config
	scanner     *portscanner.Scanner
	filter      *portscanner.Filter
	state       *portscanner.StateStore
	rateLimiter *portscanner.RateLimiter
	notifier    notifier.Notifier
}

// New constructs a Daemon from the provided config and notifier.
func New(cfg *config.Config, n notifier.Notifier) *Daemon {
	filterOpts := []portscanner.FilterOption{
		portscanner.WithProtocols(cfg.Protocols),
		portscanner.WithExcludePorts(cfg.ExcludePorts),
	}
	if cfg.ExcludeLoopback {
		filterOpts = append(filterOpts, portscanner.WithExcludeLoopback())
	}
	if cfg.ExcludePrivate {
		filterOpts = append(filterOpts, portscanner.WithExcludePrivate())
	}

	return &Daemon{
		cfg:         cfg,
		scanner:     portscanner.NewScanner(),
		filter:      portscanner.NewFilter(filterOpts...),
		state:       portscanner.NewStateStore(cfg.StateFile),
		rateLimiter: portscanner.NewRateLimiter(cfg.RateLimitCooldown),
		notifier:    n,
	}
}

// Run starts the daemon loop. It blocks until ctx is cancelled.
func (d *Daemon) Run(ctx context.Context) error {
	ticker := time.NewTicker(d.cfg.Interval)
	defer ticker.Stop()

	if err := d.tick(ctx); err != nil {
		log.Printf("[portwatch] initial scan error: %v", err)
	}

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			if err := d.tick(ctx); err != nil {
				log.Printf("[portwatch] scan error: %v", err)
			}
			d.rateLimiter.Purge()
		}
	}
}

func (d *Daemon) tick(ctx context.Context) error {
	entries, err := d.scanner.Scan()
	if err != nil {
		return err
	}

	filtered := d.filter.Apply(entries)
	current := portscanner.NewSnapshot(filtered)

	previous, _ := d.state.Load()
	events := portscanner.Diff(previous, current.ToMap())
	events = d.rateLimiter.Filter(events)

	for _, e := range events {
		if err := d.notifier.Notify(ctx, e); err != nil {
			log.Printf("[portwatch] notify error: %v", err)
		}
	}

	return d.state.Save(current.ToMap())
}
