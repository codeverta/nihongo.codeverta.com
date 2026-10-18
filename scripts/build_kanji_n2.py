"""Build the sourced N2 kanji course from the checked-in dictionary snapshot.

Run with: python scripts/build_kanji_n2.py
The generator is offline and deterministic.
"""
import json
from pathlib import Path

ROOT = Path(__file__).resolve().parents[1]
FALLBACK_WORDS = {
    # Common headwords for the small set absent from the app's JLPT reading decks.
    "塗": ("塗る", "ぬる", "to paint; to apply"),
    "帽": ("帽子", "ぼうし", "hat; cap"),
    "珍": ("珍しい", "めずらしい", "rare; unusual"),
    "翌": ("翌日", "よくじつ", "the following day"),
    "誌": ("雑誌", "ざっし", "magazine"),
    "貿": ("貿易", "ぼうえき", "trade; commerce"),
    "隻": ("一隻", "いっせき", "one ship"),
}


def katakana(text):
    return "".join(chr(ord(c) + 0x60) if "\u3041" <= c <= "\u3096" else c for c in text)


def chunks(items):
    for count in range((len(items) + 4) // 5, len(items) // 4 + 1):
        fives = len(items) - 4 * count
        if 0 <= fives <= count:
            sizes = [5] * fives + [4] * (count - fives)
            result, start = [], 0
            for size in sizes:
                result.append(items[start : start + size])
                start += size
            return result
    raise ValueError(f"Cannot partition {len(items)} kanji into 4–5 character lessons")


def source_words():
    words = {}
    for level in ("n1", "n2", "n3", "n4", "n5"):
        payload = json.loads((ROOT / f"backend/data/jlpt/{level}.json").read_text(encoding="utf-8"))
        for card in payload["cards"]:
            for reading in card.get("answer", []):
                word = card.get("question", "").strip()
                if not word or not reading:
                    continue
                key = (word, reading)
                words.setdefault(key, {"japanese": word, "reading": reading, "meaning": card.get("meaning", "")})
    return words


def build():
    reference = json.loads((ROOT / "docs/kanji-n2-reference.json").read_text(encoding="utf-8"))
    dictionary = reference["characters"]
    vocab = source_words()
    rows = {}
    for char, info in dictionary.items():
        candidates = [item for item in vocab.values() if char in item["japanese"] and item["meaning"]]
        candidates.sort(key=lambda item: (char not in item["japanese"], len(item["japanese"]), item["japanese"], item["reading"]))
        if not candidates and char in FALLBACK_WORDS:
            word, reading, meaning = FALLBACK_WORDS[char]
            candidates = [{"japanese": word, "reading": reading, "meaning": meaning}]
        if not candidates:
            raise ValueError(f"No sourced vocabulary example for {char}")
        example = candidates[0]
        en = info.get("readings_on", [])
        kun = info.get("readings_kun", [])
        meanings = info.get("meanings", [])
        meaning = "; ".join(meanings[:2]) if meanings else "see examples"
        note = (
            f"Contoh kata dan gloss mengikuti dataset kosakata JLPT aplikasi. "
            f"Arti kamus ringkas bersumber dari KANJIDIC: {meaning}. "
            "Bacaan ditampilkan dalam kata, bukan sebagai daftar semua kemungkinan bacaan."
        )
        rows[char] = {
            "character": char,
            "meaning": meaning,
            "strokes": info["strokes"],
            "onReading": "、".join(dict.fromkeys(katakana(value) for value in en)) or "—",
            "kunReading": "、".join(dict.fromkeys(kun)) or "—",
            "mnemonic": f"Pelajari {char} melalui kata nyata {example['japanese']} ({example['meaning']}).",
            "note": note,
            "examples": [example],
        }

    # A few rare characters have no entry in the app's reading decks. These are
    # common dictionary headwords with standard readings and concise glosses.
    for char, (word, reading, meaning) in FALLBACK_WORDS.items():
        if char in rows and not any(char in ex["japanese"] for ex in rows[char]["examples"]):
            rows[char]["examples"] = [{"japanese": word, "reading": reading, "meaning": meaning}]

    # Group by source frequency bands. The groups are sequencing aids, not semantic categories.
    ordered = sorted(rows, key=lambda c: (dictionary[c].get("freq", 99999), dictionary[c]["strokes"], c))
    lessons = []
    all_meanings = list(dict.fromkeys(item["meaning"] for item in rows.values()))
    all_readings = list(dict.fromkeys(item["examples"][0]["reading"] for item in rows.values()))
    for index, group in enumerate(chunks(ordered), 1):
        cards = [rows[char] for char in group]
        questions = []
        for offset, card in enumerate(cards):
            choices = [card["meaning"]] + [m for m in all_meanings if m != card["meaning"]][:3]
            if len(choices) < 4:
                raise ValueError("Insufficient distinct meaning distractors")
            correct = (index + offset) % 4
            choices.insert(correct, choices.pop(0))
            questions.append({
                "Prompt": f"Apa arti kamus ringkas {card['character']}?",
                "Choices": choices,
                "Correct": correct,
                "Explanation": f"{card['character']}: {card['meaning']}. Rujuk kata contoh {card['examples'][0]['japanese']}.",
            })
        for offset, card in enumerate(cards):
            example = card["examples"][0]
            choices = [example["reading"]] + [r for r in all_readings if r != example["reading"]][:3]
            correct = (index + offset + 1) % 4
            choices.insert(correct, choices.pop(0))
            questions.append({
                "Prompt": f"Bagaimana membaca「{example['japanese']}」?",
                "Choices": choices,
                "Correct": correct,
                "Explanation": f"{example['japanese']} dibaca {example['reading']}, berarti {example['meaning']}.",
            })
        lessons.append({
            "ID": f"kanji-n2-{index:02d}",
            "Chapter": f"Kosakata dan kanji N2 {index:02d}",
            "Title": f"Kosakata dan kanji N2 {index:02d} · {''.join(group)}",
            "Characters": cards,
            "Questions": questions,
        })

    (ROOT / "backend/data/kanji/n2.json").write_text(json.dumps(lessons, ensure_ascii=False, indent=2) + "\n", encoding="utf-8")
    curriculum = ["| Lesson | Judul | Kanji |", "|---|---|---|"]
    curriculum.extend(f"| {i:02d} | {lesson['Title']} | {' '.join(card['character'] for card in lesson['Characters'])} |" for i, lesson in enumerate(lessons, 1))
    (ROOT / "build/kanji-n2-curriculum.md").write_text("\n".join(curriculum) + "\n", encoding="utf-8")
    print(f"{len(lessons)} lessons, {len(rows)} kanji, {sum(len(lesson['Questions']) for lesson in lessons)} questions")


if __name__ == "__main__":
    build()
