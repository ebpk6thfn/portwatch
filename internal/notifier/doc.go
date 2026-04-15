// Package notifier provides pluggable notification backends for portwatch.
//
// Supported backends:
//   - WebhookNotifier: POSTs JSON-encoded Event payloads to an HTTP endpoint.
//   - DesktopNotifier: Sends native OS desktop notifications via notify-send
//     (Linux), osascript (macOS), or PowerShell (Windows).
//
// Multiple backends can be combined using Multi, which fans out each event
// to all registered notifiers and aggregates any errors.
//
// Example usage:
//
//	wh := notifier.NewWebhook("https://example.com/hook")
//	dt := notifier.NewDesktop()
//	multi := notifier.NewMulti(wh, dt)
//
//	event := notifier.Event{Type: "opened", Port: 8080, Proto: "tcp"}
//	if err := multi.Notify(event); err != nil {
//		log.Println("notify error:", err)
//	}
package notifier

import "fmt"

// ensure fmt is used by notifier.go (Multi.Notify references it).
var _ = fmt.Sprintf
