# portwatch

A lightweight CLI daemon that monitors port usage changes on a host and alerts via webhook or desktop notification.

---

## Installation

```bash
go install github.com/yourusername/portwatch@latest
```

Or build from source:

```bash
git clone https://github.com/yourusername/portwatch.git && cd portwatch && go build -o portwatch .
```

---

## Usage

Start the daemon with a polling interval and optional webhook URL:

```bash
portwatch --interval 10s --webhook https://hooks.example.com/notify
```

Use desktop notifications instead:

```bash
portwatch --interval 5s --notify desktop
```

Filter monitoring to specific port ranges:

```bash
portwatch --interval 30s --range 3000-9000 --webhook https://hooks.example.com/notify
```

### Flags

| Flag | Default | Description |
|------------|---------|--------------------------------------|
| `--interval` | `15s` | How often to poll port usage |
| `--webhook` | `""` | Webhook URL to POST change events |
| `--notify` | `""` | Notification method (`desktop`) |
| `--range` | all | Port range to monitor (e.g. `80-8080`) |
| `--verbose` | false | Enable verbose logging |

### Example Webhook Payload

```json
{
  "event": "port_opened",
  "port": 8080,
  "protocol": "tcp",
  "timestamp": "2024-05-01T12:00:00Z"
}
```

---

## License

[MIT](LICENSE)