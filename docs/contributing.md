# Участие в разработке (solo / PR)

Репозиторий ведётся в основном через **прямые коммиты в `main`** и релизы по тегу `v1.0.N`. Для **GitHub Achievements** (в т.ч. **Pull Shark**) имеет смысл периодически вливать изменения через **pull request** — критерии задаёт GitHub, пороги не публикуются; ориентир — несколько **merged** PR на аккаунт.

## Рекомендуемый цикл

1. Актуальный `main`:
   ```bash
   git fetch origin
   git checkout main
   git pull origin main
   ```
2. Ветка от `main`:
   ```bash
   git checkout -b fix/краткое-описание
   ```
3. Правки, коммит (автор: `pagbest154-cmd`, без `Co-authored-by` от ботов).
4. Push и PR:
   ```bash
   git push -u origin HEAD
   gh pr create --title "…" --body "…"
   ```
5. **Merge** PR в `main` (squash или merge commit — на достижения не влияет).
6. Локально: `git checkout main && git pull`.

Релизы по-прежнему: `CHANGELOG`, версии, коммит в `main`, тег `v1.0.N`, push тега (см. `.cursor/rules/releases.mdc`).

## Что не стоит делать ради значков

| Действие | Зачем не гоняться |
|----------|-------------------|
| **YOLO** | merge без review — риск для `main`; при solo review необязателен, но осознанно |
| **Quickdraw** | закрыть issue/PR за ~5 минут — случайность, не процесс |
| **Pair Extraordinaire** | `Co-authored-by` в коммитах — в проекте **не** добавляем co-author Cursor |

## Другие достижения

- **Starstruck** — звёзды на публичных репозиториях аккаунта.
- **Galaxy Brain** — принятые ответы в **Discussions** (нужно включить Discussions в настройках репо).
- **Public Sponsor** — GitHub Sponsors.

Скрыть блок Achievements: GitHub → Settings → Profile.

## CI

На каждый push в PR и в `main` запускается workflow **CI** (тесты на push в `main`; на теге `v*` — полная сборка релиза). Перед merge убедитесь, что проверки для PR зелёные, если workflow настроен на `pull_request`.
