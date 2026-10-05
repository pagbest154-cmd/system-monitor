# Политика безопасности

## Поддерживаемые версии

| Версия | Поддержка |
|--------|-----------|
| 1.0.x  | Активная (текущие релизы Go) |
| 0.0.x  | Архив (Python), без обновлений |

Резервная копия hub: `GET /api/hub/backup` (авторизованный доступ) или `VACUUM INTO` по [документации SQLite](https://www.sqlite.org/lang_vacuum.html) при остановленном hub.

## Сообщить об уязвимости

Если вы нашли проблему безопасности, **не создавайте публичный issue**.

Напишите через [GitHub Security Advisories](https://github.com/pagbest154-cmd/system-monitor/security/advisories/new)
или откройте приватный issue с пометкой «Security».
