from __future__ import annotations

import tkinter as tk
from tkinter import ttk


def enable_entry_clipboard(entry: ttk.Entry) -> None:
    menu = tk.Menu(entry, tearoff=0)

    def paste() -> None:
        try:
            text = entry.clipboard_get()
        except tk.TclError:
            return
        try:
            if entry.selection_present():
                entry.delete("sel.first", "sel.last")
        except tk.TclError:
            pass
        entry.insert(entry.index("insert"), text)

    def copy() -> None:
        try:
            if entry.selection_present():
                entry.clipboard_clear()
                entry.clipboard_append(entry.selection_get())
        except tk.TclError:
            pass

    def cut() -> None:
        copy()
        try:
            if entry.selection_present():
                entry.delete("sel.first", "sel.last")
        except tk.TclError:
            pass

    menu.add_command(label="Вставить", command=paste)
    menu.add_command(label="Копировать", command=copy)
    menu.add_command(label="Вырезать", command=cut)

    def show_menu(event: tk.Event) -> None:
        menu.tk_popup(event.x_root, event.y_root)

    entry.bind("<Button-3>", show_menu)
    entry.bind("<Control-v>", lambda _event: paste())
    entry.bind("<Control-V>", lambda _event: paste())
    entry.bind("<Shift-Insert>", lambda _event: paste())
    entry.bind("<Control-c>", lambda _event: copy())
    entry.bind("<Control-C>", lambda _event: copy())
    entry.bind("<Control-x>", lambda _event: cut())
    entry.bind("<Control-X>", lambda _event: cut())
    entry.bind("<Control-a>", lambda _event: (entry.select_range(0, tk.END), "break"))
    entry.bind("<Control-A>", lambda _event: (entry.select_range(0, tk.END), "break"))
