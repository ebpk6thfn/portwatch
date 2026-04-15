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
	notify  notifier.Notifier
}

// New creates a new Daemon with the provided config and notifier.
func New(cfg *config.Config, n notifier.Notifier) (*Daemon, error) {
	s, err := portscanner.NewScanner()
	if err != nil {
		return nil, err
	}
	return &Daemon{
		cfg:    cfg,
		scanner: s,
		notify: n,
	}, nil
}

// Run starts the daemon loop, scanning at the configured interval until ctx is cancelled.
func (d *Daemon) Run(ctx context.Context) error {
	prev, err := d.scanner.Scan()
	if err != nil {
		return err
	}
	log.Printf("portwatch: initial scan found %d open ports", len(prev))

	ticker := time.NewTicker(d.cfg.Interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			log.Println("portwatch: shutting down")
			return nil
		case <-ticker.C:
			curr, err := d.scanner.Scan()
			if err != nil {
				log.Printf("portwatch: scan error: %v", err)
				continue
			}

			events := portscanner.Diff(prev, curr)
			for _, ev := range events {
				if err := d.notify.Notify(ctx, ev); err != nil {
					log.Printf("portwatch: notify error: %v", err)
				}
			}
			prev = curr
		}
	}
}
