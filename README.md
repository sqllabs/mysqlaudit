# MySQL Audit

## Overview

MySQL Audit is a SQL audit tool focused exclusively on
MySQL, MariaDB, and Percona Server for MySQL, providing comprehensive
audit trails for all MySQL operations and activities.

## Roadmap

We follow MySQL’s official release roadmap. Whenever MySQL publishes a new
major version, this project upgrades as quickly as possible to support it—
for example, we target MySQL 8.4.x today and will move to MySQL 9.7.x as soon
as it becomes available.

## Community

- Join our Telegram group to discuss SQL Labs projects, ask questions, and share feedback: https://t.me/sqllabs
- Subscribe to the Telegram channel for broadcast announcements and release news: https://t.me/sqllabschannel

## Environment

- Go 1.25.4
- MySQL 8.4.x LTS
- Ubuntu 24.04.x LTS

## Quick Start

```bash
go clean -cache -modcache -testcache -fuzzcache && rm -rf $(go env GOCACHE) $(go env GOMODCACHE)
GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -trimpath -o mysqlaudit tidb-server/main.go && ./mysqlaudit
```

## License

[GNU GPL v3](LICENSE)
