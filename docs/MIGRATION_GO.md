# Миграция Python → Go

## Перед push (чистая git-история)

```bash
# После прохождения тестов и удаления Python-кода
git checkout --orphan go-main
git add -A
git commit -m "feat: system-monitor on Go (hub, agent, packaging)"
git branch -M main
git push --force origin main
git tag v1.0.0
git push origin v1.0.0
```

Старые теги `v0.0.*` не удалять — они остаются как архив Python-релизов.

## Обновление установок

| Компонент | Действие |
|-----------|----------|
| Hub (Docker) | `docker compose pull && docker compose up -d` |
| Linux agent | Переустановить `.deb` или `apt install --only-upgrade system-monitor-agent` |
| Windows agent | Установить новый `system-monitor-agent_*_setup.exe` |
| `metrics.db` | Совместим без миграции схемы |
