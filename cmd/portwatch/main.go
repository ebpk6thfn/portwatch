package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/user/portwatch/internal/config"
	"github.com/user/portwatch/internal/daemon"
	"github.com/user/portwatch/internal/notifier"
	"github.com/user/portwatch/internal/portscanner"
)

var version = "dev"

func main() {
	var (
		configPath  = flag.String("config", "", "path to config file (TOML)")
		showVersion = flag.Bool("version", false, "print version and exit")
		dryRun      = flag.Bool("dry-run", false, "scan once, print events, and exit")
	)
	flag.Parse()

	if *showVersion {
		fmt.Printf("portwatch %s\n", version)
		os.Exit(0)
	}

	// Load and validate configuration.
	cfg, err := config.Load(*configPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error loading config: %v\n", err)
		os.Exit(1)
	}
	if err := config.Validate(cfg); err != nil {
		fmt.Fprintf(os.Stderr, "invalid config: %v\n", err)
		os.Exit(1)
	}

	// Build notifier chain.
	var notifiers []notifier.Notifier
	if cfg.Webhook.URL != "" {
		notifiers = append(notifiers, notifier.NewWebhook(cfg.Webhook.URL, cfg.Webhook.Secret))
	}
	if cfg.Desktop.Enabled {
		notifiers = append(notifiers, notifier.NewDesktop(cfg.Desktop.AppName))
	}
	multi := notifier.NewMulti(notifiers...)

	// Build the port scanner pipeline.
	scanner := portscanner.NewScanner(cfg.ProcRoot)
	pipeline := portscanner.NewPipeline(scanner, cfg, multi)

	if *dryRun {
		events, err := pipeline.RunOnce(context.Background())
		if err != nil {
			fmt.Fprintf(os.Stderr, "scan error: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("portwatch dry-run: %d event(s)\n", len(events))
		for _, e := range events {
			fmt.Println(" ", e)
		}
		os.Exit(0)
	}

	// Start the daemon.
	d := daemon.New(cfg, pipeline, multi)

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	fmt.Printf("portwatch %s starting (interval: %s)\n", version, cfg.Interval)

	if err := d.Run(ctx); err != nil {
		fmt.Fprintf(os.Stderr, "daemon error: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("portwatch stopped")
}
