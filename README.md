# BriefPush

BriefPush pulls RSS/Atom feeds, uses an LLM to summarize recent items, and sends scheduled digest reports via configured notification providers.

## Features

- Polls configured feeds every 30 minutes.
- Generates AI summaries and a combined report.
- Sends daily reports at `report_hour`.
- Supports multiple notification providers (Email, Feishu).

## Requirements

- Go 1.26+
- Network access to feed sources
- LLM API endpoint and key

## Quick Start

1. Copy and edit config:
   ```bash
   cp config.example.json config.json
   ```
2. Update `config.json`:
   - `base_url`, `api_key`, `model`
   - `feeds` list
   - `notification.providers`
3. Run:
   ```bash
   go run .
   ```

## Configuration Notes

- `json_file_path`: local feed storage file (default: `feeds.json`).
- `report_hour`: daily report hour (0-23).
- `notification.providers`: enable one or more providers.

### Email provider

Set `type` to `email`, `enabled` to `true`, and fill:

- `from`
- `to` (at least one recipient)
- `smtp.host`, `smtp.port`, `smtp.username`, `smtp.password`, `smtp.use_tls`

### Feishu provider

Set `type` to `feishu`, `enabled` to `true`, and fill:

- `feishu.webhook_url`

## Project Structure

- `main.go`: app entrypoint
- `feed/`: feed fetching, scheduling, report generation
- `ai/`: LLM integration for summary/report generation
- `notify/`: notifier dispatcher and provider implementations
- `store/`: JSON-based feed storage
- `config/`: config model and loader
