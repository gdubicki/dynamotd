"""Converts the colored terminal output of dynamotd to an ImageMagick drawing (MVG),
in the style of the macOS Terminal with the Monaco font."""

import os
import re
import sys

COLORS = {"34": "#2E43C0", "32": "#57BF38", "33": "#C7C43F", "31": "#D7473D", None: "#DCDCDC"}
BACKGROUND = "#15191E"
FONT = os.environ.get("SCREENSHOTS_FONT", "/System/Library/Fonts/Monaco.ttf")
POINTSIZE, CHAR_WIDTH, LINE_HEIGHT, ASCENT, PADDING = 36, 22.1, 44, 33, 24

ansi_path, mvg_path = sys.argv[1], sys.argv[2]
lines = open(ansi_path, encoding="utf-8").read().rstrip("\n").split("\n")

def segments(line):
    """yields (column, text, color) for each run of text in one color"""
    column, color = 0, None
    for part in re.split(r"(\x1b\[[0-9;]*m)", line):
        match = re.fullmatch(r"\x1b\[([0-9;]*)m", part)
        if match:
            color = None if match.group(1) in ("0", "") else match.group(1)
        elif part:
            yield column, part, color
            column += len(part)

def escape(text):
    return text.replace("\\", "\\\\").replace("'", "\\'")

width = int(max(sum(len(t) for _, t, _ in segments(l)) for l in lines) * CHAR_WIDTH + 2 * PADDING)
height = len(lines) * LINE_HEIGHT + 2 * PADDING - (LINE_HEIGHT - ASCENT) // 2

mvg = [f"viewbox 0 0 {width} {height}", f"fill '{BACKGROUND}'", f"rectangle 0,0 {width},{height}",
       f"font '{FONT}'", f"font-size {POINTSIZE}", "text-antialias 1"]
for row, line in enumerate(lines):
    y = PADDING + row * LINE_HEIGHT + ASCENT
    for column, text, color in segments(line):
        mvg.append(f"fill '{COLORS[color]}'")
        mvg.append(f"text {PADDING + column * CHAR_WIDTH:.1f},{y} '{escape(text)}'")
open(mvg_path, "w").write("\n".join(mvg) + "\n")
print(width, height)
