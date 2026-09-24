from __future__ import annotations

import socket
import threading
import tkinter as tk
from pathlib import Path
from tkinter import messagebox, ttk

from ..config_loader import (
    AgentFileConfig,
    load_agent_config,
    load_agent_token,
    save_agent_config,
    save_agent_token,
)
from ..paths import AGENT_TOKEN_FILE
from .branding import apply_tk_window_icon
from .service_control import get_service_state, restart_service
from .win_ui import enable_entry_clipboard
from .updates import check_for_updates, format_update_message, get_installed_version, open_update_page


class SettingsApp:
    def __init__(self, config_path: Path | None = None) -> None:
        self.config_path = config_path
        self.config = load_agent_config(config_path)
        self.token = load_agent_token(self.config)

        self.root = tk.Tk()
        self.root.title("system-monitor agent — настройки")
        self.root.resizable(False, False)
        self.root.minsize(420, 320)
        apply_tk_window_icon(self.root)

        frame = ttk.Frame(self.root, padding=16)
        frame.grid(row=0, column=0, sticky="nsew")

        self._hub_url = tk.StringVar(value=self.config.hub_url)
        self._agent_id = tk.StringVar(value=self.config.agent_id or socket.gethostname())
        self._token = tk.StringVar(value=self.token)
        self._interval = tk.StringVar(value=str(self.config.interval_sec))

        self._add_field(frame, 0, "Hub URL", self._hub_url)
        self._add_field(frame, 1, "Agent ID", self._agent_id)
        self._add_field(frame, 2, "Token", self._token, show="*")
        self._add_field(frame, 3, "Интервал (с)", self._interval)

        status = get_service_state()
        status_text = {
            "running": "Служба запущена",
            "stopped": "Служба остановлена",
            "missing": "Служба не установлена",
        }.get(status, "Статус службы неизвестен")
        ttk.Label(frame, text=status_text).grid(row=4, column=0, columnspan=2, sticky="w", pady=(12, 0))

        version_frame = ttk.Frame(frame)
        version_frame.grid(row=5, column=0, columnspan=2, sticky="ew", pady=(12, 0))
        ttk.Label(version_frame, text=f"Версия: {get_installed_version()}").grid(row=0, column=0, sticky="w")
        self._update_button = ttk.Button(version_frame, text="Проверить обновления", command=self._check_updates)
        self._update_button.grid(row=0, column=1, sticky="e", padx=(12, 0))
        version_frame.columnconfigure(1, weight=1)

        buttons = ttk.Frame(frame)
        buttons.grid(row=6, column=0, columnspan=2, sticky="e", pady=(16, 0))
        ttk.Button(buttons, text="Отмена", command=self.root.destroy).grid(row=0, column=0, padx=(0, 8))
        ttk.Button(buttons, text="Сохранить", command=self._save).grid(row=0, column=1)

        self.root.columnconfigure(0, weight=1)
        self.root.rowconfigure(0, weight=1)

    def _add_field(self, parent: ttk.Frame, row: int, label: str, variable: tk.StringVar, show: str | None = None) -> None:
        ttk.Label(parent, text=label).grid(row=row, column=0, sticky="w", pady=4)
        entry = ttk.Entry(parent, textvariable=variable, width=42, show=show)
        entry.grid(row=row, column=1, sticky="ew", pady=4)
        enable_entry_clipboard(entry)
        parent.columnconfigure(1, weight=1)

    def _check_updates(self) -> None:
        self._update_button.configure(state="disabled")

        def _run() -> None:
            result = check_for_updates(force=True)
            self.root.after(0, lambda: self._update_button.configure(state="normal"))
            if result.update_available:
                if messagebox.askyesno(
                    "Доступно обновление",
                    format_update_message(result) + "\n\nОткрыть страницу загрузки?",
                    parent=self.root,
                ):
                    open_update_page(result)
            elif result.error:
                messagebox.showerror("Обновления", format_update_message(result), parent=self.root)
            else:
                messagebox.showinfo("Обновления", format_update_message(result), parent=self.root)

        threading.Thread(target=_run, daemon=True).start()

    def _save(self) -> None:
        hub_url = self._hub_url.get().strip()
        agent_id = self._agent_id.get().strip()
        token = self._token.get().strip()
        interval_raw = self._interval.get().strip()

        if not hub_url:
            messagebox.showerror("Ошибка", "Укажите Hub URL", parent=self.root)
            return
        if not agent_id:
            messagebox.showerror("Ошибка", "Укажите Agent ID", parent=self.root)
            return
        try:
            interval_sec = max(1, int(interval_raw))
        except ValueError:
            messagebox.showerror("Ошибка", "Интервал должен быть целым числом", parent=self.root)
            return

        token_file = self.config.token_file.strip() or str(AGENT_TOKEN_FILE)
        updated = AgentFileConfig(
            hub_url=hub_url,
            agent_id=agent_id,
            token="",
            token_file=token_file,
            interval_sec=interval_sec,
            transport=self.config.transport or "http",
        )
        try:
            save_agent_config(updated, self.config_path)
            save_agent_token(token, token_file)
        except OSError as exc:
            messagebox.showerror(
                "Ошибка",
                "Не удалось сохранить настройки.\n\n"
                f"{exc}\n\n"
                "Запустите «Настройки агента» от имени администратора.",
                parent=self.root,
            )
            return

        messagebox.showinfo(
            "Сохранено",
            "Настройки сохранены.\n"
            "Служба подхватит изменения в течение нескольких секунд.",
            parent=self.root,
        )
        if messagebox.askyesno(
            "Перезапуск службы",
            "Перезапустить службу агента сейчас?\n(может потребоваться запуск от администратора)",
            parent=self.root,
        ):
            ok, detail = restart_service()
            if ok:
                messagebox.showinfo("Готово", detail, parent=self.root)
            else:
                messagebox.showwarning(
                    "Перезапуск службы",
                    f"Не удалось перезапустить службу автоматически.\n\n{detail}\n\n"
                    "Изменения уже сохранены — служба применит их автоматически.",
                    parent=self.root,
                )
        self.root.destroy()

    def run(self) -> None:
        self.root.mainloop()


def run_settings(config_path: Path | None = None) -> None:
    SettingsApp(config_path).run()
