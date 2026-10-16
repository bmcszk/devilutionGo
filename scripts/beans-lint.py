#!/usr/bin/env python3
"""beans-lint: validate .beans/devilGo-*.md files. Exit 1 on any error.
Checks: YAML frontmatter parses; title has no unquoted colon; blocked_by ids exist;
skills paths exist on disk; Libraries section on todo beans; no legacy dg-task-NNN prose in todo bodies.
Usage: python3 beans-lint.py [repo_root]
"""
import glob, os, re, sys

try:
    import yaml
except ImportError:
    yaml = None

REPO = sys.argv[1] if len(sys.argv) > 1 else os.getcwd()
BEANS = os.path.join(REPO, ".beans")
SKILLS_ROOTS = ["/home/blaze/.agents/skills", os.path.expanduser("~/.hermes/skills")]
LEGACY = re.compile(r"\bdg-task-\d+\b")
errors = []

def err(bean, msg):
    errors.append(f"{bean}: {msg}")

files = sorted(glob.glob(os.path.join(BEANS, "devilGo-*.md")))
if not files:
    print(f"beans-lint: no bean files found in {BEANS}")
    sys.exit(1)

ids = set()
raw = {}
for f in files:
    bid = os.path.basename(f).split("--")[0]
    ids.add(bid)
    raw[bid] = open(f, errors="ignore").read()

for bid, t in raw.items():
    m = re.match(r"^---\n(.*?)\n---\n", t, re.S)
    if not m:
        err(bid, "no frontmatter block")
        continue
    fm_text = m.group(1)
    fm = None
    if yaml is not None:
        try:
            fm = yaml.safe_load(fm_text)
        except yaml.YAMLError as e:
            err(bid, f"frontmatter YAML parse: {e}")
            continue
    else:
        # minimal fallback: title colon check
        for line in fm_text.splitlines():
            if line.startswith("title:") and line.count(":") > 1 and not line.startswith('title: "') and not line.startswith("title: '"):
                err(bid, f"unquoted colon in title line: {line[:60]}")
    if fm and "status" not in fm:
        err(bid, "frontmatter missing status")

    status = (fm or {}).get("status", "")
    # blocked_by targets exist
    for dep in re.findall(r"devilGo-[a-z0-9]{4}", t):
        pass  # prose refs are allowed; only frontmatter deps are checked below
    if fm:
        deps = fm.get("blocked_by") or []
        if isinstance(deps, str):
            deps = [deps]
        for d in deps:
            if d not in ids:
                err(bid, f"blocked_by target does not exist: {d}")

    # legacy task ids must not survive in todo bodies
    if status == "todo" and LEGACY.search(t):
        err(bid, "todo body still references legacy dg-task-NNN ids")

    # todo consumer beans need a Libraries section (lib-binding policy)
    if status == "todo" and "## Libraries" not in t:
        # only warn-classified via stderr-ish marker: keep as error per policy
        err(bid, "todo bean missing ## Libraries section (lib-binding policy)")

# skills paths exist
for bid, t in raw.items():
    for sp in re.findall(r"(/home/blaze/\.agents/skills/[\w\-/]+|~/.hermes/skills/[\w\-/]+)", t):
        p = os.path.expanduser(sp)
        if not (os.path.isdir(p) or os.path.isfile(p + ".md") or os.path.isfile(p + "/SKILL.md")):
            err(bid, f"skills path missing on disk: {sp}")

if errors:
    print(f"beans-lint: {len(errors)} error(s)")
    for e in errors:
        print("  " + e)
    sys.exit(1)
print(f"beans-lint: OK ({len(files)} beans)")
