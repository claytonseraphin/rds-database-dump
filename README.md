# rds-database-dump

Small Go utility to dump all non-system databases from a MySQL RDS instance concurrently, then generate per-database JSON reports plus a summary report.

## What it does

- Connects to MySQL RDS and lists available databases.
- Skips system databases:
	- `mysql`
	- `information_schema`
	- `performance_schema`
	- `sys`
- Dumps each remaining database using `mysqldump`.
- Runs dumps concurrently (current code uses a concurrency limit of `2`).
- Creates:
	- one dump directory per RDS host
	- one SQL dump file per database
	- one JSON report per database in `reports/`
	- one `reports/summary.json` after all dumps finish

## Requirements

- Go `1.25+`
- MySQL client tools installed and available in your PATH:
	- `mysql`
	- `mysqldump`
- macOS (the current implementation wraps dumps with `caffeinate`)

## Configuration

The app reads credentials from environment variables and also supports flags.

Expected variables:

- `HOST`
- `PORT`
- `USERNAME`
- `PASSWORD`

Important:

- The code loads a file named `.env` at startup.
- This repository includes a sample file named `dotenv`. Copy or rename it to `.env` before running.

Example `.env`:

```env
HOST=your-instance.your-region.rds.amazonaws.com
PORT=3306
USERNAME=your-db-username
PASSWORD=your-db-password
```

## Run

From the project root:

```bash
go run ./cmd
```

You can also override values using flags:

```bash
go run ./cmd \
	-h your-instance.your-region.rds.amazonaws.com \
	-P 3306 \
	-u your-db-username \
	-p your-db-password
```

## Output structure

After execution, you should see:

```text
<rds-prefix>-dump/
	<database1>_dump.sql
	<database2>_dump.sql
	...

reports/
	<database1>.json
	<database2>.json
	...
	summary.json
```

`<rds-prefix>-dump` is derived from the first segment of the host (for example, host `prod-db.abc123.us-east-1.rds.amazonaws.com` creates `prod-db-dump/`).

## Reports

### Per-database report

Each database gets a JSON report containing:

- database name
- dump status
- dump duration in seconds
- error field (when a dump fails)

### Summary report

`reports/summary.json` includes totals for:

- number of databases processed
- number of successful dumps
- number of failed dumps

## Notes

- If no user databases are found, the program exits with an error.
- The tool currently sets concurrency in code (`ConcurrentDumpDB(..., 2)` in `cmd/main.go`).
