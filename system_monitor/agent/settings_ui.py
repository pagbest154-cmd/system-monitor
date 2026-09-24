from __future__ import annotations

import socket
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
from .service_control import get_service_state, restart_service


class SettingsApp:
    def __init__(self, config_path: Path | None = None) -> None:
        self.config_path = config_path
        self.config = load_agent_config(config_path)
        self.token = load_agent_token(self.config)

        self.root = tk.Tk()
        self.root.title("system-monitor agent — настройки")
        self.root.resizable(False, False)
        self.root.minsize(420, 320)

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

        buttons = ttk.Frame(frame)
        buttons.grid(row=5, column=0, columnspan=2, sticky="e", pady=(16, 0))
        ttk.Button(buttons, text="Отмена", command=self.root.destroy).grid(row=0, column=0, padx=(0, 8))
        ttk.Button(buttons, text="Сохранить", command=self._save).grid(row=0, column=1)

        self.root.columnconfigure(0, weight=1)
        self.root.rowconfigure(0, weight=1)

    def _add_field(self, parent: ttk.Frame, row: int, label: str, variable: tk.StringVar, show: str | None = None) -> None:
        ttk.Label(parent, text=label).grid(row=row, column=0, sticky="w", pady=4)
        entry = ttk.Entry(parent, textvariable=variable, width=42, show=show)
        entry.grid(row=row, column=1, sticky="ew", pady=4)
        parent.columnconfigure(1, weight=1)

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
        save_agent_config(updated, self.config_path)
        save_agent_token(token, token_file)

        if messagebox.askyesno(
            "Сохранено",
            "Настройки сохранены.\nПерезапустить службу агента?",
            parent=self.root,
        ):
            ok, detail = restart_service()
            if ok:
                messagebox.showinfo("Готово", detail, parent=self.root)
            else:
                messagebox.showwarning(
                    "Перезапуск службы",
                    f"Не удалось перезапустить службу автоматически.\n\n{detail}\n\n"
                    "Перезапустите службу «system-monitor-agent» вручную через services.msc.",
                    parent=self.root,
                )
        self.root.destroy()

    def run(self) -> None:
        self.root.mainloop()


def run_settings(config_path: Path | None = None) -> None:
    SettingsApp(config_path).run()
