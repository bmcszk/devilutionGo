#!/usr/bin/env python3
"""beans-lint: validate .beans/devilGo-*.md files. Exit 1 on any error, exit 2 without PyYAML.
Checks: PyYAML is mandatory and frontmatter must parse (an unquoted colon in title fails the parse);
status must be draft|todo|in-progress|completed|scrapped; duplicate bean ids across .beans/ and
.beans/archive/ are reported and the sorted-last file wins; structure law on non-archive beans
(milestone/epic are roots with empty parent; feature's parent must be an epic; task/bug's parent must
be feature/task/epic; parents must exist); blocked_by ids exist; skills paths exist on disk;
Libraries section on todo task/bug/feature beans; no legacy dg-task-NNN prose in todo bodies;
status state machine (draft: no Proof of Work; in-progress: claim evidence; completed: Proof of Work;
todo: Triage-brainer evidence + Acceptance criteria with checkboxes — skills paths on todo bodies are
covered by the global skills check);
Skills for coder check (non-archive task/bug/feature in todo/in-progress: '## Skills for coder' section
listing coding-rules and ponytail SKILL.md paths; go-touching bodies also list go-dev SKILL.md —
archive/ is exempt from body checks, it is scanned only for id-references).
Usage: python3 beans-lint.py [repo_root]
       python3 beans-lint.py --epic-status <bean-id> <pr-number>   (advisory, always exit 0)
"""
import glob, os, re, subprocess, sys

try:
    import yaml
except ImportError:
    print("beans-lint: FATAL: PyYAML required (pip install pyyaml)", file=sys.stderr)
    sys.exit(2)

_cli_args = sys.argv[1:]
EPIC_STATUS_ARGS = None
if "--epic-status" in _cli_args:
    _i = _cli_args.index("--epic-status")
    EPIC_STATUS_ARGS = _cli_args[_i + 1:_i + 3]
    _cli_args = _cli_args[:_i] + _cli_args[_i + 3:]
REPO = _cli_args[0] if _cli_args else os.getcwd()
BEANS = os.path.join(REPO, ".beans")
SKILLS_ROOTS = ["/home/blaze/.agents/skills", os.path.expanduser("~/.hermes/skills")]
LEGACY = re.compile(r"\bdg-task-\d+\b")
errors = []

def err(bean, msg):
    errors.append(f"{bean}: {msg}")

def strip_frontmatter(text):
    m = re.match(r"^---\n.*?\n---\n", text, re.S)
    return text[m.end():] if m else text

def has_claim_evidence(body):
    if "## Progress log" in body or "## Proof of Work" in body:
        return True
    return re.search(r"(?im)^.*\bclaimed\b.*\b20\d\d-\d\d-\d\d", body) is not None

def acceptance_criteria(body):
    m = re.search(r"(?ms)^## (?:Global )?[Aa]cceptance [Cc]riteria[^\n]*\n(.*?)(?=^## |\Z)", body)
    return m.group(1) if m else None

def has_todo_triage(body):
    sections = re.findall(r"(?ims)^## Triage[^\n]*(?:\n|$).*?(?=^## |\Z)", body)
    if any(re.search(r"brainer", section, re.I) for section in sections):
        return True
    return re.search(r"(?im)^\s*TRIAGE\b[^\n]*brainer", body) is not None

def check_todo_acceptance(bid, body):
    ac = acceptance_criteria(body)
    if ac is None:
        err(bid, "todo bean missing ## Acceptance criteria section")
        return
    if not re.search(r"(?m)^\s*-\s+\[.\]", ac):
        err(bid, "todo bean ## Acceptance criteria has no checkbox line")

GO_HINT = re.compile(r"internal/|cmd/|\bgo (?:test|vet)|gofmt")
CODING_TYPES = {"task", "bug", "feature"}
CODING_STATUSES = {"todo", "in-progress"}
REQUIRED_SKILLS = ("coding-rules/SKILL.md", "ponytail/SKILL.md")
GO_SKILL = "go-dev/SKILL.md"

def skill_listed(body, skill):
    return re.search(r"(?:~|/home/blaze)/\.agents/skills/" + re.escape(skill), body) is not None

def check_coder_skills(bid, fm, body):
    """task/bug/feature beans in todo/in-progress must carry a '## Skills for coder' section."""
    if (fm or {}).get("type", "") not in CODING_TYPES or (fm or {}).get("status", "") not in CODING_STATUSES:
        return
    if not re.search(r"(?im)^## Skills for coder\b", body):
        err(bid, f"{fm.get('status')} {fm.get('type')} bean missing ## Skills for coder section")
        return
    for skill in REQUIRED_SKILLS:
        if not skill_listed(body, skill):
            err(bid, f"## Skills for coder missing {skill}")
    if GO_HINT.search(body) and not skill_listed(body, GO_SKILL):
        err(bid, f"## Skills for coder missing {GO_SKILL} (go-touching bean)")

def check_status_state_machine(bid, status, body):
    if status == "draft" and "## Proof of Work" in body:
        err(bid, "draft bean contains ## Proof of Work (smuggled completion)")
    elif status == "in-progress" and not has_claim_evidence(body):
        err(bid, "in-progress bean has no claim evidence (## Progress log, ## Proof of Work, or claimed timestamp line)")
    elif status == "completed" and "## Proof of Work" not in body:
        err(bid, "completed bean missing ## Proof of Work section")
    elif status == "todo":
        if not has_todo_triage(body):
            err(bid, "todo bean missing delegated triage evidence (## Triage citing brainer)")
        check_todo_acceptance(bid, body)

STATUSES = {"draft", "todo", "in-progress", "completed", "scrapped"}
PARENT_TYPES = {
    "feature": {"epic"},
    "task": {"feature", "task", "epic"},
    "bug": {"feature", "task", "epic"},
}
type_of = {}

def check_status_and_parent(bid, fm, archived):
    """Status whitelist for every bean; parent/type structure law for non-archive beans only."""
    status = fm.get("status", "")
    if status not in STATUSES:
        err(bid, f"unknown status: {status}")
    if archived:
        return
    btype = fm.get("type", "")
    parent = fm.get("parent") or ""
    if isinstance(parent, str):
        parent = parent.strip()
    if not parent:
        return
    if parent not in type_of:
        err(bid, f"parent does not exist: {parent}")
    elif btype in ("milestone", "epic"):
        err(bid, f"root bean ({btype}) must have empty parent")
    elif btype == "feature" and type_of[parent] not in PARENT_TYPES["feature"]:
        err(bid, f"feature must be parented under an epic (parent is {type_of[parent]})")
    elif btype in ("task", "bug") and type_of[parent] not in PARENT_TYPES[btype]:
        err(bid, f"task must be parented under feature/task/epic (parent is {type_of[parent]})")

def read_epic_status(bean_id):
    matches = sorted(glob.glob(os.path.join(BEANS, bean_id + "*.md")))
    if not matches:
        return None
    m = re.match(r"^---\n(.*?)\n---\n", open(matches[0], errors="ignore").read(), re.S)
    if not m:
        return None
    return (yaml.safe_load(m.group(1)) or {}).get("status", "")

def pr_commit_count(pr_number):
    proc = subprocess.run(
        ["gh", "pr", "view", str(pr_number), "--json", "commits", "--jq", ".commits|length"],
        capture_output=True, text=True,
    )
    if proc.returncode != 0:
        return None, proc.stderr.strip()
    return int(proc.stdout.strip()), None

def pr_body(pr_number):
    proc = subprocess.run(
        ["gh", "pr", "view", str(pr_number), "--json", "body", "--jq", ".body"],
        capture_output=True, text=True,
    )
    if proc.returncode != 0:
        return None, proc.stderr.strip()
    return proc.stdout.strip(), None

def head_sha(repo):
    proc = subprocess.run(["git", "-C", repo, "rev-parse", "HEAD"], capture_output=True, text=True)
    if proc.returncode != 0:
        return None, proc.stderr.strip()
    return proc.stdout.strip(), None

READY_MARKER = re.compile(r"(?m)^READY-FOR-REVIEW\s+(\S+)\s+@\s+([0-9a-f]{7,40})\s*$")

def print_epic_status(bean_id, pr_number):
    status = read_epic_status(bean_id)
    if status is None:
        print(f"beans-lint: epic-status: bean not found: {bean_id}")
        sys.exit(1)
    count, gh_err = pr_commit_count(pr_number)
    if gh_err is not None:
        print(f"beans-lint: epic-status: gh pr view {pr_number} failed: {gh_err}")
        sys.exit(1)
    body, body_err = pr_body(pr_number)
    if body_err is not None:
        print(f"beans-lint: epic-status: gh pr view {pr_number} failed: {body_err}")
        sys.exit(1)
    head, git_err = head_sha(REPO)
    if git_err is not None:
        print(f"beans-lint: epic-status: git rev-parse HEAD failed: {git_err}")
        sys.exit(1)
    # ready = epic completed AND its READY-FOR-REVIEW marker sha matches current HEAD
    ready = status == "completed"
    if ready:
        marker_shas = [m.group(2) for m in READY_MARKER.finditer(body or "") if m.group(1) == bean_id]
        ready = any(head.startswith(sha) for sha in marker_shas)
    print(f"epic-status: {bean_id} status={status} pr-commits={count} ready={'true' if ready else 'false'}")
    sys.exit(0)

files = sorted(glob.glob(os.path.join(BEANS, "devilGo-*.md"))) + sorted(glob.glob(os.path.join(BEANS, "archive", "devilGo-*.md")))
if EPIC_STATUS_ARGS is not None:
    if len(EPIC_STATUS_ARGS) < 2:
        print("beans-lint: usage: beans-lint.py --epic-status <bean-id> <pr-number>")
        sys.exit(1)
    print_epic_status(*EPIC_STATUS_ARGS)
if not files:
    print(f"beans-lint: no bean files found in {BEANS}")
    sys.exit(1)

ids = set()
raw = {}
non_archive = set()
parsed = {}
by_bid = {}
for f in files:
    by_bid.setdefault(os.path.basename(f).split("--")[0], []).append(f)
# duplicate ids: report, then keep only the sorted-last file (same as the old dict-overwrite)
for bid, paths in by_bid.items():
    if len(paths) > 1:
        err(bid, f"duplicate bean id: {bid} ({len(paths)} files: {', '.join(paths)}); later file shadows earlier")
for f in files:
    bid = os.path.basename(f).split("--")[0]
    if f != by_bid[bid][-1]:
        continue
    ids.add(bid)
    raw[bid] = open(f, errors="ignore").read()
    # archived beans are scanned only for id-references; the skills-for-coder rule is
    # not applied to them (archive predates the rule)
    if "/archive/" not in f:
        non_archive.add(bid)

for bid, t in raw.items():
    m = re.match(r"^---\n(.*?)\n---\n", t, re.S)
    if not m:
        err(bid, "no frontmatter block")
        continue
    fm_text = m.group(1)
    fm = None
    try:
        fm = yaml.safe_load(fm_text)
    except yaml.YAMLError as e:
        err(bid, f"frontmatter YAML parse: {e}")
        continue
    # (the deleted no-yaml fallback's title-colon check is dead: safe_load errors on an unquoted colon)
    parsed[bid] = fm
    if isinstance(fm, dict):
        type_of[bid] = fm.get("type", "")
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

    # lib-binding policy: only consumer beans (task/bug/feature) need a Libraries section
    if status == "todo" and (fm or {}).get("type") in ("task", "bug", "feature") and "## Libraries" not in t:
        # only warn-classified via stderr-ish marker: keep as error per policy
        err(bid, "todo bean missing ## Libraries section (lib-binding policy)")

    check_status_state_machine(bid, status, strip_frontmatter(t))
    if bid in non_archive:
        check_coder_skills(bid, fm, strip_frontmatter(t))

# structure law after the parse loop so parent types are known
for bid, fm in parsed.items():
    if isinstance(fm, dict):
        check_status_and_parent(bid, fm, bid not in non_archive)

# skills paths exist (local machines only — skip when no skills root is present, e.g. CI checkout)
if any(os.path.isdir(os.path.expanduser(root)) for root in SKILLS_ROOTS):
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
