#!/usr/bin/env python3
"""devilGo-1spo fixture storage: refresh + verify (docs/fixtures.md).

fixture-refresh — keep test/fixtures/generated/ in lockstep with a probe run:
  1. Re-run the C++ probe only when tools/probe/out/ is absent (the caller is
     responsible for having built dvg_probe; this script never wires the
     devilutionX build — see tools/probe/README.md).
  2. Copy the probe artifacts into test/fixtures/generated/ (same relative
     layout; manifest.txt overwrites the committed one).
  3. Diff old-vs-new manifest rows and print a drift report
     (added / removed / changed paths).
  4. Regenerate the "## Committed artifact inventory" section of
     docs/fixtures.md (file counts per class + pinned commit).

fixture-verify — hash-only: recompute every manifest row's sha256 against
test/fixtures/generated/ and exit nonzero on any mismatch or missing file.
"""

import hashlib
import shutil
import sys
from pathlib import Path

REPO = Path(__file__).resolve().parent.parent
PROBE_OUT = REPO / "tools/probe/out"
GENERATED = REPO / "test/fixtures/generated"
DOCS = REPO / "docs/fixtures.md"
INVENTORY_HEADING = "## Committed artifact inventory"
CLASSES = ("dun", "frames", "lighttables", "rng", "save")


def read_manifest(path: Path) -> dict[str, str]:
    """sha256sum-format manifest -> {relpath: hash} (rows only)."""
    rows: dict[str, str] = {}
    for line in path.read_text().splitlines():
        if not line or line.startswith("dvg-probe") or line.startswith("devilutionX commit:"):
            continue
        digest, relpath = line.split("  ", 1)
        rows[relpath] = digest
    return rows


def pinned_commit(path: Path) -> str:
    for line in path.read_text().splitlines():
        if line.startswith("devilutionX commit: "):
            return line.removeprefix("devilutionX commit: ")
    raise SystemExit(f"no pinned commit in {path}")


def refresh() -> int:
    if not PROBE_OUT.exists():
        print(f"fixture-refresh: {PROBE_OUT} absent — re-run the probe first.")
        print("Reproduction steps: tools/probe/README.md (Build + Subcommands).")
        return 1

    old_rows = read_manifest(GENERATED / "manifest.txt")
    shutil.copytree(PROBE_OUT, GENERATED, dirs_exist_ok=True)
    for name in ("demo_0.dmo", "seeds.txt"):
        shutil.copy(REPO / "tools/probe" / name, GENERATED / name)

    new_rows = read_manifest(GENERATED / "manifest.txt")
    added = sorted(set(new_rows) - set(old_rows))
    removed = sorted(set(old_rows) - set(new_rows))
    changed = sorted(p for p in set(old_rows) & set(new_rows) if old_rows[p] != new_rows[p])

    print(f"drift: {len(added)} added, {len(removed)} removed, {len(changed)} changed")
    for label, paths in (("added", added), ("removed", removed), ("changed", changed)):
        for relpath in paths:
            print(f"  {label}: {relpath}")
    if not (added or removed or changed):
        print("drift: 0 changed (probe output byte-identical to committed fixtures)")

    write_inventory()
    return 0


def write_inventory() -> None:
    text = DOCS.read_text()
    if INVENTORY_HEADING not in text:
        raise SystemExit(f"{INVENTORY_HEADING} missing from {DOCS}")
    commit = pinned_commit(GENERATED / "manifest.txt")
    lines = [
        INVENTORY_HEADING,
        "",
        f"Pinned devilutionX commit: `{commit}` "
        "(probe run: `tools/probe/out/` per tools/probe/README.md).",
        "",
    ]
    rows = read_manifest(GENERATED / "manifest.txt")
    lines.append("| Class | Files | Contents |")
    lines.append("|---|---|---|")
    contents = {
        "dun": "seeded DRLG dumps: `.dun` + `.meta` per (level, entry, seed)",
        "frames": "timedemo frames: `.idx.bin` index buffer + `.pal.bin` palette",
        "lighttables": "MakeLightTable() outputs: 4 leveltype tables, falloffs, cone",
        "rng": "1000 raw `GenerateRandomNumber()` outputs per seed",
        "save": "vanilla spawn `.sv` hero archive",
    }
    for klass in CLASSES:
        count = sum(1 for relpath in rows if relpath.startswith(klass + "/"))
        lines.append(f"| `{klass}/` | {count} | {contents[klass]} |")
    extras = sorted(p.name for p in GENERATED.iterdir()
                    if p.is_file() and p.name != "README.md")
    lines.append(f"| root files | {len(extras)} | {', '.join(f'`{e}`' for e in extras)} |")
    lines.append("")

    start = text.index(INVENTORY_HEADING)
    end = len(text)
    for heading in ("\n## ", "\n# "):
        nxt = text.find(heading, start + len(INVENTORY_HEADING))
        if nxt != -1:
            end = nxt + 1
            break
    DOCS.write_text(text[:start] + "\n".join(lines) + text[end:])
    print(f"regenerated {INVENTORY_HEADING} in {DOCS}")


def verify() -> int:
    rows = read_manifest(GENERATED / "manifest.txt")
    bad = 0
    for relpath, digest in sorted(rows.items()):
        path = GENERATED / relpath
        if not path.exists():
            print(f"MISSING: {relpath}")
            bad += 1
            continue
        actual = hashlib.sha256(path.read_bytes()).hexdigest()
        if actual != digest:
            print(f"MISMATCH: {relpath} (manifest {digest}, file {actual})")
            bad += 1
    if bad:
        print(f"fixture-verify: {bad} failed of {len(rows)}")
        return 1
    print(f"fixture-verify: all {len(rows)} hashes match")
    return 0


if __name__ == "__main__":
    mode = sys.argv[1] if len(sys.argv) > 1 else ""
    if mode == "refresh":
        raise SystemExit(refresh())
    if mode == "verify":
        raise SystemExit(verify())
    raise SystemExit("usage: fixture_fixtures.py refresh|verify")
