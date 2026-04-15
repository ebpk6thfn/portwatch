package notifier

import (
	"fmt"
	"os/exec"
	"runtime"
)

// DesktopNotifier sends native desktop notifications.
type DesktopNotifier struct{}

// NewDesktop creates a DesktopNotifier.
func NewDesktop() *DesktopNotifier {
	return &DesktopNotifier{}
}

// Notify dispatches a desktop notification using the platform-appropriate tool.
func (d *DesktopNotifier) Notify(event Event) error {
	title := fmt.Sprintf("portwatch: port %s", event.Type)
	body := fmt.Sprintf("%s/%d (pid %d %s)", event.Proto, event.Port, event.PID, event.Process)

	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "linux":
		cmd = exec.Command("notify-send", title, body)
	case "darwin":
		script := fmt.Sprintf(`display notification %q with title %q`, body, title)
		cmd = exec.Command("osascript", "-e", script)
	case "windows":
		// PowerShell toast — best-effort on older systems
		ps := fmt.Sprintf(
			`[Windows.UI.Notifications.ToastNotificationManager, Windows.UI.Notifications, ContentType = WindowsRuntime] | Out-Null; `+
				`$t = [Windows.UI.Notifications.ToastNotificationManager]::GetTemplateContent(0); `+
				`$t.GetElementsByTagName('text')[0].AppendChild($t.CreateTextNode('%s %s')) | Out-Null; `+
				`[Windows.UI.Notifications.ToastNotificationManager]::CreateToastNotifier('portwatch').Show($t)`,
			title, body,
		)
		cmd = exec.Command("powershell", "-Command", ps)
	default:
		return fmt.Errorf("desktop notifications not supported on %s", runtime.GOOS)
	}

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("desktop notify: %w", err)
	}
	return nil
}

func (d *DesktopNotifier) Name() string { return "desktop" }
