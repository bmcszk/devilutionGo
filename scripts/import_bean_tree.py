#!/usr/bin/env python3
"""Import a fenced-JSON bean tree into a beans ledger (orchestrator = single writer).

Tree format: markdown with ```json blocks, each {"id","type","title","parent","blocked_by","tags","priority","body"}.
Pass 1 creates draft beans (type map: epic->epic, task/subtask->task) recording plan-id -> real-id;
pass 2 wires --blocked-by via the map. Idempotent via the plan-id map json (--idmap, read+updated):
batches already in the map are skipped, so re-running repairs partial imports.
"""
import argparse
import json
import os
import re
import subprocess
import sys
import tempfile

TYPE_MAP = {"epic": "epic", "task": "task", "subtask": "task"}


def blocks(path):
    text = open(path).read()
    for m in re.finditer(r"```json\s*(\{.*?\})\s*```", text, re.S):
        yield json.loads(m.group(1), strict=False)


def beans(cwd, args):
    r = subprocess.run(["beans"] + list(args), cwd=cwd, capture_output=True, text=True)
    if r.returncode != 0:
        print("FAIL", " ".join(args[:3]), "->", r.stderr.strip()[:200])
        return None
    return r.stdout


def main():
    ap = argparse.ArgumentParser()
    ap.add_argument("tree")
    ap.add_argument("--cwd", default=".")
    ap.add_argument("--idmap", help="json plan-id -> real-bean-id (read + updated in place)")
    ap.add_argument("--root", help="real bean id that parentless epics hang from")
    a = ap.parse_args()
    idmap = json.load(open(a.idmap)) if a.idmap and os.path.exists(a.idmap) else {}
    all_beans = list(blocks(a.tree))
    created = {}

    for b in all_beans:
        if b["id"] in idmap or b["id"] in created:
            continue
        parent = b.get("parent")
        real_parent = None
        if parent:
            real_parent = created.get(parent) or idmap.get(parent)
            if not real_parent:
                print("UNMAPPED PARENT:", b["id"], "->", parent)
                continue
        elif a.root and b["type"] == "epic":
            real_parent = a.root
        with tempfile.NamedTemporaryFile("w", suffix=".md", delete=False) as f:
            f.write(b["body"])
            bf = f.name
        args = ["create", b["title"], "-t", TYPE_MAP[b["type"]], "-s", "draft", "--body-file", bf]
        if real_parent:
            args += ["--parent", real_parent]
        for t in b.get("tags", []):
            args += ["--tag", t]
        out = beans(a.cwd, args + ["--json"])
        os.unlink(bf)
        if out:
            created[b["id"]] = json.loads(out)["bean"]["id"]
            print(f"{b['id']} -> {created[b['id']]} ({TYPE_MAP[b['type']]})")

    wired = 0
    for b in all_beans:
        rid = created.get(b["id"]) or idmap.get(b["id"])
        for bl in b.get("blocked_by") or []:
            real = created.get(bl) or idmap.get(bl)
            if rid and real:
                if beans(a.cwd, ["update", rid, "--blocked-by", real]):
                    wired += 1
            else:
                print("BLOCKER UNMAPPED:", b["id"], "->", bl)
    print(f"created={len(created)}/{len(all_beans)} blockers_wired={wired}")
    if a.idmap:
        json.dump({**idmap, **created}, open(a.idmap, "w"), indent=1)
    return 0


if __name__ == "__main__":
    sys.exit(main())
