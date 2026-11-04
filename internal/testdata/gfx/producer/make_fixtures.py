#!/usr/bin/env python3
"""Offline fixture producer for the CEL/CL2 -> CLX conversion parity goldens.

For each case this script:

1. reads a committed CLX hexdump from `internal/testdata/gfx/clx/`;
2. derives an *input* in the legacy format (CEL or CL2) from that CLX data,
   which is what the retail-era loaders would have received;
3. runs the pinned devilutionX C++ converter (built from `oracle_driver.cpp`,
   see README.md) on that input;
4. writes the input, the per-frame widths, and the C++ golden output as
   hexdumps under `internal/testdata/gfx/{cel,cl2}/`.

The derived input and the golden decode to the same pixels as the source CLX
frame; the golden bytes are NOT generally equal to the source `.clx`, because
the CL2 and CEL encoders flush solid runs at different points than the PCX
encoder that produced the bundled files.

This script is never run by `go test` or CI: the Go tests read the committed
bytes only.

Usage:
    ORACLE=/path/to/oracle ./make_fixtures.py [--check]

`--check` re-derives and compares against the committed files instead of
writing them.
"""

import argparse
import os
import pathlib
import struct
import subprocess
import sys

TRANSPARENT = 0xFF  # sentinel used only inside this producer

# Cases: one committed CLX file, converted through each legacy container.
CASES = [
    ("data_health", "cl2"),
    ("data_xpbar", "cl2"),
    ("data_dirtybuc", "cl2"),
    ("data_tabBtnUp", "cl2"),
    ("data_monstertags", "cl2"),
    ("data_hinticons", "cl2"),
    ("data_resistance", "cl2"),
    ("data_health", "cel"),
    ("data_boxleftend", "cel"),
    ("data_boxrightend", "cel"),
    ("data_dirtybucp", "cel"),
    ("data_monstertags", "cel"),
    ("data_panel8bucp", "cel"),
    ("data_stashnavbtns", "cel"),
    ("data_repairAllBtn", "cel"),
    ("data_talkbutton", "cel"),
]

# Multi-group CEL cases: a group offset table plus two frame groups, with a
# decoy frame list in the gap between them. Upstream CEL conversion advances
# by the previous group's last frame end (cel_to_clx.cpp:56,118), so the decoy
# list IS the second group; a group-table lookup would instead read `groups[1]`
# past the gap and produce different bytes. Field order: case name, source CLX,
# frame indices per group, frame indices of the decoy list.
MULTIGROUP_CASES = [
    ("multi_group_stashnavbtns", "data_stashnavbtns", [[0, 1], [2, 3]], [4]),
]

# Surface cases: every committed CLX frame of the source stacked into one
# row-major pixel surface, encoded by SurfaceToClx (surface_to_clx.cpp:20-80).
# Field order: case name, source CLX, transparent palette index or "-".
SURFACE_CASES = [
    ("surface_repairAllBtn", "data_repairAllBtn", "-"),
    ("surface_repairAllBtn_trn", "data_repairAllBtn", "255"),
]


def load_hex(path):
    return bytes.fromhex(path.read_text().replace("\n", "").strip())


def hexdump(data):
    return "\n".join(data[i:i + 16].hex() for i in range(0, len(data), 16)) + "\n"


def dump_hex(path, data):
    path.parent.mkdir(parents=True, exist_ok=True)
    path.write_text(hexdump(data))


def parse_list(buf):
    """Split a CLX list into per-frame payloads (header: count, count+1 offsets)."""
    count = struct.unpack_from("<I", buf, 0)[0]
    offsets = struct.unpack_from("<%dI" % (count + 1), buf, 4)
    return [buf[offsets[i]:offsets[i + 1]] for i in range(count)]


def clx_frame(body):
    """Return (width, height, command stream) of one CLX frame."""
    header_size, width, height = struct.unpack_from("<HHH", body, 0)
    return width, height, body[header_size:]


def decode_pixels(stream, width, height):
    """Decode a CLX command stream into a width*height grid of pixel indices.

    Runs advance an (x, y) cursor and wrap at the right edge, exactly like the
    C++ decoder (GetSkipSize, Source/utils/clx_decode.hpp:46-60). Pixels the
    stream never reaches stay TRANSPARENT.
    """
    pixels = [TRANSPARENT] * (width * height)
    x = y = src = 0
    while src < len(stream):
        control = stream[src]
        src += 1
        if control >= 0xBF:
            run = 65 if control == 0xBF else 256 - control
            values = list(stream[src:src + run])
            src += run
        elif control >= 0x80:
            run = 0xBF - control
            values = [stream[src]] * run
            src += 1
        else:
            run = control
            values = [TRANSPARENT] * run
        for value in values:
            if y >= height:
                raise SystemExit("run overruns the frame height")
            pixels[y * width + x] = value
            x += 1
            y += x // width
            x %= width
    if (x, y) != (0, height):
        raise SystemExit(f"decoded to ({x},{y}), expected (0,{height})")

    return pixels


def cel_lines(pixels, width, height):
    """Encode a pixel grid as CEL line runs; every line sums to exactly width."""
    out = bytearray()
    for y in range(height):
        row = pixels[y * width:(y + 1) * width]
        x = 0
        while x < width:
            color = row[x]
            run = 1
            while x + run < width and row[x + run] == color and run < 128:
                run += 1
            if color == TRANSPARENT:
                out.append((256 - run) & 0xFF)
            else:
                while run > 127:
                    out.append(127)
                    out.extend(row[x:x + 127])
                    x += 127
                    run -= 127
                out.append(run)
                out.extend(row[x:x + run])
            x += run

    return bytes(out)


def build_frame_list(frames, frame_header):
    """Assemble count + offsets + frames; the first offset is the header size."""
    count = len(frames)
    header_size = 4 * (count + 2)
    offsets = [header_size]
    body = b""
    for frame in frames:
        body += frame_header + frame
        offsets.append(header_size + len(body))

    return struct.pack("<I", count) + struct.pack("<%dI" % (count + 1), *offsets) + body


def derive(clx, kind):
    """Derive a CEL or CL2 fixture input from one committed CLX file."""
    bodies = parse_list(clx)
    frames = []
    widths = []
    for body in bodies:
        width, height, stream = clx_frame(body)
        widths.append(width)
        if kind == "cl2":
            # CL2 frames carry a header whose size is its first u16
            # (Source/utils/cl2_to_clx.cpp:71).
            frames.append(struct.pack("<HH", 4, 0) + stream)
        else:
            pixels = decode_pixels(stream, width, height)
            # CEL frames carry a 10-byte header when its first u16 is 10
            # (Source/utils/cel_to_clx.cpp:80-84); the fields are unused.
            frames.append(b"\x0a\x00" + b"\x00" * 8 + cel_lines(pixels, width, height))

    return widths, build_frame_list(frames, b"")


def cel_group_frames(clx, indices):
    """Derive CEL frames (10-byte header + line runs) for the given CLX indices."""
    bodies = parse_list(clx)
    frames = []
    widths = []
    for index in indices:
        width, height, stream = clx_frame(bodies[index])
        pixels = decode_pixels(stream, width, height)
        frames.append(b"\x0a\x00" + b"\x00" * 8 + cel_lines(pixels, width, height))
        widths.append(width)

    return widths, frames

def derive_multigroup(clx, groups, gap):
    """Derive a two-group CEL: table, group 0, decoy list, group 1."""
    if len(groups) != 2:
        raise SystemExit("multi-group cases are defined for exactly two groups")

    group_widths = []
    group_lists = []
    for indices in groups:
        widths, frames = cel_group_frames(clx, indices)
        group_widths.extend(widths)
        group_lists.append(build_frame_list(frames, b""))
    gap_widths, gap_frames = cel_group_frames(clx, gap)
    gap_list = build_frame_list(gap_frames, b"")

    if len(set(group_widths + gap_widths)) != 1:
        raise SystemExit(f"multi-group case needs one width, got {group_widths + gap_widths}")

    # The table is numGroups*4 bytes; slot 0 holds that size, which is also the
    # address of the first group (cel_to_clx.cpp:44-49, cl2_to_clx.cpp:24-28).
    table_size = 4 * len(groups)
    payload = bytearray(struct.pack("<II", table_size, 0))
    payload += group_lists[0]
    payload += gap_list
    struct.pack_into("<I", payload, 4, len(payload))
    payload += group_lists[1]

    # The single-group heuristic must not fire: it compares the u32 at
    # 4*(table_size)+4 against the file size (cel_to_clx.cpp:53-58).
    probe = struct.unpack_from("<I", payload, table_size * 4 + 4)[0]
    if probe == len(payload):
        raise SystemExit("multi-group probe collides with the file size")

    return group_widths[0], bytes(payload)

def derive_surface(clx):
    """Stack every CLX frame of `clx` into one row-major pixel surface."""
    bodies = parse_list(clx)
    pixels = []
    widths = []
    for body in bodies:
        width, height, stream = clx_frame(body)
        widths.append(width)
        pixels.extend(decode_pixels(stream, width, height))
    if len(set(widths)) != 1:
        raise SystemExit(f"surface case needs one width, got {widths}")

    return widths[0], len(bodies), bytes(pixels)

def emit(args, failures, files):
    """Write (or in --check mode compare) one case's fixture files."""
    for path, data in files.items():
        if args.check:
            expected = hexdump(data) if path.suffix == ".hex" else data.decode()
            if not path.exists() or path.read_text() != expected:
                failures.append(str(path))
        elif path.suffix == ".hex":
            dump_hex(path, data)
        else:
            path.write_bytes(data)

def run_oracle(kind, in_path, out_path, widths):
    oracle = os.environ.get("ORACLE")
    if not oracle:
        raise SystemExit("set ORACLE to the built oracle binary (see README.md)")
    subprocess.run([oracle, kind, str(in_path), str(out_path)] + [str(w) for w in widths],
                   check=True, capture_output=True)

    return out_path.read_bytes()


def main():
    parser = argparse.ArgumentParser()
    parser.add_argument("--check", action="store_true", help="compare instead of writing")
    args = parser.parse_args()

    gfx = pathlib.Path(__file__).resolve().parents[1]
    scratch = pathlib.Path(os.environ.get("SCRATCH", "/tmp/dgo-assets-fixtures"))
    scratch.mkdir(parents=True, exist_ok=True)

    failures = []
    cases = [(name, kind, derive(load_hex(gfx / "clx" / f"{name}.clx.hex"), kind))
             for name, kind in CASES]
    for name, source, groups, gap in MULTIGROUP_CASES:
        clx = load_hex(gfx / "clx" / f"{source}.clx.hex")
        width, payload = derive_multigroup(clx, groups, gap)
        cases.append((name, "cel", ([width], payload)))

    for name, kind, (widths, payload) in cases:
        in_path = scratch / f"{name}.{kind}"
        in_path.write_bytes(payload)
        golden = run_oracle(kind, in_path, scratch / f"{name}.{kind}.golden.clx", widths)

        emit(args, failures, {
            gfx / kind / f"{name}.{kind}.hex": payload,
            gfx / kind / f"{name}.{kind}.clx.hex": golden,
            gfx / kind / f"{name}.widths": (",".join(str(w) for w in widths) + "\n").encode(),
        })
        print(f"{kind} {name}: frames={len(widths)} in={len(payload)} golden={len(golden)}")

    for name, source, transparent in SURFACE_CASES:
        width, numFrames, pixels = derive_surface(load_hex(gfx / "clx" / f"{source}.clx.hex"))
        in_path = scratch / f"{name}.surface"
        in_path.write_bytes(pixels)
        golden = run_oracle("surface", in_path, scratch / f"{name}.surface.golden.clx",
                            (width, numFrames, transparent))
        emit(args, failures, {
            gfx / "surface" / f"{name}.surface.hex": pixels,
            gfx / "surface" / f"{name}.surface.clx.hex": golden,
            gfx / "surface" / f"{name}.surface.args":
                f"{width},{numFrames},{transparent}\n".encode(),
        })
        print(f"surface {name}: {width}x{numFrames} frames in={len(pixels)} golden={len(golden)}")

    if failures:
        print("stale fixtures:", *failures, sep="\n  ")
        return 1

    return 0


if __name__ == "__main__":
    sys.exit(main())
