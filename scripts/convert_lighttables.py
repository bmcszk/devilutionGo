#!/usr/bin/env python3
"""Convert lighttable goldens from xxd .hex dumps to human-readable .tsv.

One-shot migration (devilGo-d12h): each committed
internal/testdata/gamedata/lighttables/*.hex is an xxd text dump. This
script decodes it (xxd -r semantics) and emits a sibling .tsv with header
comment lines (# table / # dims / # source) followed by one row of 16
decimal byte values per line, tab-separated. The .hex originals are then
deleted — the .tsv files are the committed goldens.
"""

import subprocess
import sys
from pathlib import Path

REPO = Path(__file__).resolve().parent.parent
FIXTURES = REPO / "internal/testdata/gamedata/lighttables"
SOURCE = "devilutionX/Source/engine/light_tables.cpp"
ROW_BYTES = 16


def decode_hex_dump(path: Path) -> bytes:
    """xxd -r semantics: let xxd itself reverse its own dump format."""
    out = subprocess.run(
        ["xxd", "-r"], input=path.read_bytes(), capture_output=True, check=True
    )
    return out.stdout


def convert(path: Path) -> None:
    raw = decode_hex_dump(path)
    rows = len(raw) // ROW_BYTES
    cols = ROW_BYTES
    if rows * cols != len(raw):
        raise SystemExit(f"{path.name}: size {len(raw)} is not a multiple of {ROW_BYTES}")
    tsv = path.with_suffix(".tsv")
    lines = [
        f"# table: {path.stem}",
        f"# dims: {cols}x{rows}",
        f"# source: {SOURCE}",
    ]
    for r in range(rows):
        row = raw[r * cols : (r + 1) * cols]
        lines.append("\t".join(str(b) for b in row))
    tsv.write_text("\n".join(lines) + "\n")
    path.unlink()
    print(f"{tsv.relative_to(REPO)} ({len(raw)} bytes, {rows}x{cols})")


def main() -> int:
    hexes = sorted(FIXTURES.glob("*.hex"))
    if not hexes:
        print("no .hex fixtures found", file=sys.stderr)
        return 1
    for path in hexes:
        convert(path)
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
