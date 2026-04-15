package daemon

import (
	"context"
	"log"
	"time"

	"github.com/user/portwatch/internal/config"
	"github.com/user/portwatch/internal/notifier"
	"github.com/user/portwatch/internal/portscanner"
)

// Daemon orchestrates periodic port scanning and change notification.
type Daemon struct {
	cfg     *config.Config
	scanner *portscanner.Scanner
	store   *portscanner.StateStore
	notify  notifier.Notifier
}

// New creates a Daemon wired with the provided dependencies.
func New(cfg *config.Config, scanner *portscanner.Scanner, store *portscanner.StateStore, n notifier.Notifier) *Daemon {
	return &Daemon{
		cfg:    cfg,
		scanner: scanner,
		store:  store,
		notify: n,
	}
}

// Run starts the scan loop and blocks until ctx is cancelled.
func (d *Daemon) Run(ctx context.Context) error {
	ticker := time.NewTicker(d.cfg.Interval)
	defer ticker.Stop()

	// Perform an initial scan on startup to establish baseline.
	if err := d.tick(); err != nil {
		log.Printf("[portwatch] initial scan error: %v", err)
	}

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			if err := d.tick(); err != nil {
				log.Printf("[portwatch] scan error: %v", err)
			}
		}
	}
}

func (d *Daemon) tick() error {
	prev, err := d.store.Load()
	if err != nil {
		return err
	}

	current, err := d.scanner.Scan()
	if err != nil {
		return err
	}

	events := portscanner.Diff(prev, current)
	for _, ev := range events {
		if err := d.notify.Notify(ev); err != nil {
			log.Printf("[portwatch] notify error: %v", err)
		}
	}

	return d.store.Save(current)
}
