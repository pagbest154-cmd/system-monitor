# Документация system-monitor

| Файл | Назначение |
|------|------------|
| [taskmcp-project-context.md](taskmcp-project-context.md) | Полный контекст для **TaskMCP** и AI-агентов (workflow, стек Go, API, ingest prompt). Синхронизация: `../scripts/sync-taskmcp-context.sh` |
| [MIGRATION_GO.md](MIGRATION_GO.md) | История миграции Python → Go, обновление установок |
| [dashboard.svg](dashboard.svg) | Превью UI для README |

Пользовательская документация: корневой [README.md](../README.md), [CHANGELOG.md](../CHANGELOG.md), [android/README.md](../android/README.md).

**Текущий релиз:** `v1.0.72` (hub Docker, agent `.deb` amd64/arm64, Windows installer, `sysmon-1.0.72.apk`) — см. [Releases](https://github.com/pagbest154-cmd/system-monitor/releases/tag/v1.0.72).

Правила Cursor: [`.cursor/rules/`](../.cursor/rules/) — `project.mdc`, `packaging.mdc`, **`debian-agent-install.mdc`** (debconf, conffiles, postinst), `releases.mdc`, `taskmcp.mdc`, `prod-infrastructure.mdc`, `git-commits.mdc`.
