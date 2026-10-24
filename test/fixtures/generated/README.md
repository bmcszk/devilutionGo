# generated/ — probe-produced fixtures (local, not committed)

Local copy of one `dvg_probe` run (devilutionX pinned at
`ca91d778fba92a1f680afb1fb2cdf09337ff1b42`; see tools/probe/README.md).
The binaries here are regenerable artifacts and are **gitignored** — refresh:
`just fixture-refresh`; verify against the committed manifest:
`just fixture-verify`.

The only committed text is `../manifest.txt` (sha256 of every artifact below
at refresh time). Subdirs: `dun/`, `frames/`, `lighttables/`, `rng/`,
`save/` — see each subdir's README for what it contains. Root files are the
probe inputs copied from `tools/probe/`: `demo_0.dmo` (scripted timedemo),
`seeds.txt` (seed list).
