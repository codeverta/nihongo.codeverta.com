"""Mirror every quiz_data JSON from a pinned Kotoba revision, verify Git blobs.

The original JSON bytes are retained losslessly as gzip. Study chunks are derived
separately, with stable IDs based on source file and ordinal. No runtime scraping.
"""
import concurrent.futures
import gzip
import hashlib
import json
from pathlib import Path
import time
import urllib.parse
import urllib.request

REVISION = "a0f8f984bf5dd8014527c6608d36b479ab622b07"
ROOT = Path(__file__).resolve().parents[1] / "backend/internal/anki/resources"


def fetch(url):
    for attempt in range(4):
        try:
            with urllib.request.urlopen(urllib.request.Request(url, headers={"User-Agent": "NihongoMaster-resource-import", "Accept-Encoding": "gzip"}), timeout=90) as response:
                data = response.read()
                return gzip.decompress(data) if response.headers.get("Content-Encoding") == "gzip" else data
        except Exception:
            if attempt == 3:
                raise
            time.sleep(1 + attempt)


def mirror(item):
    destination = ROOT / "source" / (item["name"] + ".gz")
    if destination.exists():
        data = gzip.decompress(destination.read_bytes())
    else:
        data = fetch("https://raw.githubusercontent.com/mistval/kotoba/" + REVISION + "/resources/quiz_data/" + urllib.parse.quote(item["name"]))
    blob = hashlib.sha1(b"blob " + str(len(data)).encode() + b"\0" + data).hexdigest()
    if blob != item["sha"]:
        raise ValueError("Git blob mismatch: " + item["name"])
    destination.parent.mkdir(parents=True, exist_ok=True)
    if not destination.exists():
        destination.write_bytes(gzip.compress(data, mtime=0))
    parsed = json.loads(data)
    return dict(file=item["name"], gitBlob=blob, sha256=hashlib.sha256(data).hexdigest(), bytes=len(data)), parsed


def main():
    ROOT.mkdir(parents=True, exist_ok=True)
    entries = json.loads(fetch(f"https://api.github.com/repos/mistval/kotoba/contents/resources/quiz_data?ref={REVISION}"))
    if any(x["type"] != "file" or not x["name"].endswith(".json") for x in entries):
        raise ValueError("Unexpected source directory entry; review before importing")
    manifests, audit = [], []
    with concurrent.futures.ThreadPoolExecutor(max_workers=8) as executor:
        futures = {executor.submit(mirror, item): item for item in entries}
        for future in concurrent.futures.as_completed(futures):
            manifest, data = future.result()
            manifests.append(manifest)
            cards = data["cards"]
            audit.append(dict(file=manifest["file"], name=data.get("name"), count=len(cards),
                              strategy=data.get("cardPreprocessingStrategy"),
                              questionStrategy=data.get("questionCreationStrategy"),
                              empty=sum(not isinstance(c, dict) or not c.get("answer") for c in cards),
                              description=data.get("description", ""),
                              example=cards[:2]))
            if len(manifests) % 20 == 0:
                print(f"Mirrored {len(manifests)}/{len(entries)}", flush=True)
    (ROOT / "KOTOBA-LICENSE.txt").write_bytes(fetch(f"https://raw.githubusercontent.com/mistval/kotoba/{REVISION}/LICENSE"))
    (ROOT / "manifest.json").write_text(json.dumps(dict(repository="https://github.com/mistval/kotoba", revision=REVISION, files=sorted(manifests, key=lambda x:x["file"])), indent=2) + "\n", encoding="utf-8")
    report = Path(__file__).resolve().parents[1] / "build/kotoba-audit.json"
    report.write_text(json.dumps(sorted(audit, key=lambda x:x["file"]), ensure_ascii=False, indent=2), encoding="utf-8")
    print(f"Mirrored all {len(manifests)} source files; inspect build/kotoba-audit.json", flush=True)


if __name__ == "__main__":
    main()
