"""Shared application icon rendering for tray, settings UI, and Windows packaging."""

from __future__ import annotations

import sys
from pathlib import Path

from PIL import Image, ImageDraw

STATUS_COLORS = {
    "ok": "#22c55e",
    "error": "#ef4444",
    "idle": "#94a3b8",
}

_BG = "#1a2332"
_ACCENT = "#3b82f6"
_CHART = "#22c55e"
_BASE_SIZE = 32
_ICON_SIZES = (16, 20, 24, 32, 40, 48, 64, 128, 256)


def _scale(value: float, size: int) -> float:
    return value * size / _BASE_SIZE


def _stroke(size: int) -> int:
    if size <= 20:
        return 2
    if size <= 32:
        return 2
    return max(2, int(round(_scale(2, size))))


def _draw_icon(draw: ImageDraw.ImageDraw, size: int, status: str) -> None:
    stroke = _stroke(size)
    radius = max(2, int(_scale(8, size)))

    draw.rounded_rectangle((0, 0, size - 1, size - 1), radius=radius, fill=_BG)

    monitor = (
        _scale(5, size),
        _scale(6, size),
        _scale(27, size),
        _scale(21, size),
    )
    draw.rounded_rectangle(
        monitor,
        radius=max(1, int(_scale(2, size))),
        outline=_ACCENT,
        width=stroke,
    )

    stand_y = _scale(26, size)
    draw.line(
        (_scale(11, size), stand_y, _scale(21, size), stand_y),
        fill=_ACCENT,
        width=stroke,
    )
    draw.line(
        (_scale(16, size), _scale(21, size), _scale(16, size), stand_y),
        fill=_ACCENT,
        width=stroke,
    )

    chart = [
        (_scale(9, size), _scale(16, size)),
        (_scale(13, size), _scale(12, size)),
        (_scale(17, size), _scale(15, size)),
        (_scale(23, size), _scale(9, size)),
    ]
    draw.line(chart, fill=_CHART, width=stroke, joint="curve")

    if size >= 32:
        status_color = STATUS_COLORS.get(status, STATUS_COLORS["idle"])
        dot_radius = max(2, int(round(_scale(4, size))))
        dot_center = (_scale(25, size), _scale(25, size))
        bbox = (
            dot_center[0] - dot_radius,
            dot_center[1] - dot_radius,
            dot_center[0] + dot_radius,
            dot_center[1] + dot_radius,
        )
        draw.ellipse(bbox, fill=status_color, outline=_BG, width=max(1, stroke // 2))


def render_icon(size: int, status: str = "idle") -> Image.Image:
    """Draw the app icon at the exact target size (crisp small sizes, smooth large)."""
    if size >= 64:
        render_size = size * 2
        image = Image.new("RGBA", (render_size, render_size), (0, 0, 0, 0))
        draw = ImageDraw.Draw(image)
        _draw_icon(draw, render_size, status)
        return image.resize((size, size), Image.Resampling.LANCZOS)

    image = Image.new("RGBA", (size, size), (0, 0, 0, 0))
    draw = ImageDraw.Draw(image)
    _draw_icon(draw, size, status)
    return image


def icon_file_path() -> Path | None:
    """Return bundled or packaged .ico path when available."""
    ico_name = "app-icon.ico"
    candidates: list[Path] = []

    if getattr(sys, "frozen", False):
        meipass = getattr(sys, "_MEIPASS", None)
        if meipass:
            candidates.append(Path(meipass) / "assets" / ico_name)
        candidates.append(Path(sys.executable).resolve().parent / ico_name)
    else:
        root = Path(__file__).resolve().parents[2]
        candidates.append(root / "packaging" / "windows" / ico_name)

    for candidate in candidates:
        if candidate.is_file():
            return candidate
    return None


def save_app_icon(path: Path, sizes: tuple[int, ...] = _ICON_SIZES) -> Path:
    """Write a multi-size Windows .ico file (largest size first for Explorer)."""
    path.parent.mkdir(parents=True, exist_ok=True)
    ordered = tuple(sorted(sizes, reverse=True))
    images = [render_icon(size, "ok") for size in ordered]
    images[0].save(
        path,
        format="ICO",
        sizes=[(image.width, image.height) for image in images],
        append_images=images[1:],
    )
    return path


def _ensure_icon_file() -> Path | None:
    path = icon_file_path()
    if path:
        return path

    cache_dir = Path.home() / ".cache" / "system-monitor"
    cache_dir.mkdir(parents=True, exist_ok=True)
    cached = cache_dir / "app-icon.ico"
    if not cached.is_file():
        save_app_icon(cached)
    return cached


def apply_tk_window_icon(root) -> None:
    """Set the Tk window icon on Windows when an .ico file is available."""
    path = _ensure_icon_file()
    if not path:
        return
    try:
        root.iconbitmap(default=str(path))
    except Exception:
        # Non-Windows dev environments may not support .ico window icons.
        pass
