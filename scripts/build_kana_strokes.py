"""Convert KanjiVG kana SVGs into a compact polyline JSON for the app.

Usage: python -I build_kana.py <svg_dir> <out_json>
"""
import json
import math
import re
import sys
from pathlib import Path

TOKEN = re.compile(r"[MmCcSsLlHhVvZz]|-?\d*\.?\d+(?:e-?\d+)?")
STEP = 1.6  # resample spacing in the 109x109 KanjiVG box


def cubic(p0, p1, p2, p3, n=24):
    out = []
    for i in range(1, n + 1):
        t = i / n
        u = 1 - t
        out.append((
            u**3 * p0[0] + 3 * u * u * t * p1[0] + 3 * u * t * t * p2[0] + t**3 * p3[0],
            u**3 * p0[1] + 3 * u * u * t * p1[1] + 3 * u * t * t * p2[1] + t**3 * p3[1],
        ))
    return out


def parse(d):
    toks = TOKEN.findall(d)
    i = 0
    cmd = None
    cur = (0.0, 0.0)
    start = cur
    last_ctrl = None
    pts = []

    def num():
        nonlocal i
        v = float(toks[i])
        i += 1
        return v

    while i < len(toks):
        if re.match(r"[A-Za-z]", toks[i]):
            cmd = toks[i]
            i += 1
        rel = cmd.islower()
        c = cmd.upper()
        ox, oy = cur if rel else (0.0, 0.0)
        if c == "M":
            cur = (ox + num(), oy + num())
            start = cur
            pts.append(cur)
            last_ctrl = None
            cmd = "l" if rel else "L"
        elif c == "C":
            p1 = (ox + num(), oy + num())
            p2 = (ox + num(), oy + num())
            p3 = (ox + num(), oy + num())
            pts += cubic(cur, p1, p2, p3)
            last_ctrl, cur = p2, p3
        elif c == "S":
            p1 = (2 * cur[0] - last_ctrl[0], 2 * cur[1] - last_ctrl[1]) if last_ctrl else cur
            p2 = (ox + num(), oy + num())
            p3 = (ox + num(), oy + num())
            pts += cubic(cur, p1, p2, p3)
            last_ctrl, cur = p2, p3
        elif c == "L":
            cur = (ox + num(), oy + num())
            pts.append(cur)
            last_ctrl = None
        elif c == "H":
            cur = ((ox if rel else 0) + num(), cur[1])
            pts.append(cur)
            last_ctrl = None
        elif c == "V":
            cur = (cur[0], (oy if rel else 0) + num())
            pts.append(cur)
            last_ctrl = None
        elif c == "Z":
            cur = start
            pts.append(cur)
            last_ctrl = None
        else:
            raise ValueError(f"unsupported command {cmd} in {d}")
    return pts


def resample(pts):
    out = [pts[0]]
    carry = 0.0
    for a, b in zip(pts, pts[1:]):
        seg = math.dist(a, b)
        if seg == 0:
            continue
        pos = STEP - carry
        while pos <= seg:
            t = pos / seg
            out.append((a[0] + (b[0] - a[0]) * t, a[1] + (b[1] - a[1]) * t))
            pos += STEP
        carry = seg - (pos - STEP)
    if math.dist(out[-1], pts[-1]) > 0.3:
        out.append(pts[-1])
    return out


def main(svg_dir, out_json):
    data = {}
    for svg in sorted(Path(svg_dir).glob("*.svg")):
        text = svg.read_text(encoding="utf-8")
        char = chr(int(svg.stem, 16))
        strokes = re.findall(r'<path id="kvg:[0-9a-f]+-s(\d+)"[^>]*\sd="([^"]+)"', text)
        strokes.sort(key=lambda s: int(s[0]))
        if not strokes:
            raise SystemExit(f"no strokes in {svg}")
        data[char] = [
            [round(v, 1) for p in resample(parse(d)) for v in p] for _, d in strokes
        ]
    payload = {
        "source": "KanjiVG (https://kanjivg.tagaini.net) by Ulrich Apel, CC BY-SA 3.0",
        "viewBox": 109,
        "strokes": data,
    }
    Path(out_json).write_text(
        json.dumps(payload, ensure_ascii=False, separators=(",", ":")), encoding="utf-8"
    )
    print(len(data), "characters,", Path(out_json).stat().st_size, "bytes")


if __name__ == "__main__":
    main(sys.argv[1], sys.argv[2])
