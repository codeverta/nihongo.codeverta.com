"""Build the sourced JLPT N1 kanji curriculum from its checked-in snapshot.

The level assignment is an attributed study mapping, not an official JLPT list.
Kanji metadata comes from KANJIDIC through kanji-data; examples are JMdict
headwords where available. The script is deterministic and needs no network.
"""
import json
from pathlib import Path

ROOT = Path(__file__).resolve().parents[1]


def katakana(value):
    return ''.join(chr(ord(c) + 0x60) if '\u3041' <= c <= '\u3096' else c for c in value)


def chunks(chars):
    for count in range((len(chars) + 4) // 5, len(chars) // 4 + 1):
        fives = len(chars) - 4 * count
        if 0 <= fives <= count:
            result, start = [], 0
            for size in [5] * fives + [4] * (count - fives):
                result.append(chars[start:start + size])
                start += size
            return result
    raise ValueError(f'Curriculum cannot be partitioned into 4–5 character lessons: {len(chars)}')


def build():
    reference = json.loads((ROOT / 'docs/kanji-n1-reference.json').read_text(encoding='utf-8'))
    dictionary = reference['characters']
    rows = {}
    for char, info in dictionary.items():
        meanings = info['meanings']
        examples = info['examples']
        if (not meanings or info['strokes'] < 1 or len(examples) != 2
                or any(char not in ex['word'] or not ex['reading'] or not ex['meaning'] for ex in examples)):
            raise ValueError(f'Invalid sourced character entry: {char}')

        on_readings = [r for r in info['readings_on'] if r and not r.startswith('-')][:2]
        kun_readings = [r for r in info['readings_kun'] if r and not r.startswith('-')][:2]
        on = '、'.join(katakana(r) for r in on_readings) or '—'
        kun = '、'.join(kun_readings) or '—'
        word = examples[0]
        source_note = 'Contoh kata dan bacaannya mengikuti JMdict; makna contoh memakai gloss bahasa Inggris sumber.'
        if any(ex.get('source', '').startswith('KANJIDIC fallback') for ex in examples):
            source_note += ' Untuk kanji yang tidak mempunyai headword JMdict pada snapshot ini, contoh satu-karakter mengikuti entri KANJIDIC dan ditandai pada dokumentasi.'
        rows[char] = {
            'character': char,
            'meaning': '; '.join(meanings[:5]),
            'strokes': info['strokes'],
            'onReading': on,
            'kunReading': kun,
            'mnemonic': f'Pelajari arti {char} melalui kata {word["word"]}（{word["reading"]}）. Satu kanji dapat memiliki beberapa bacaan menurut katanya.',
            'note': source_note + ' Bacaan yang tercantum merupakan pilihan dari sumber, bukan daftar seluruh bacaan.',
            'examples': [
                {'japanese': ex['word'], 'reading': ex['reading'], 'meaning': ex['meaning']}
                for ex in examples
            ],
        }

    groups = chunks(list(rows))
    lessons = []
    all_meanings = list(dict.fromkeys(row['meaning'] for row in rows.values()))
    all_readings = list(dict.fromkeys(row['examples'][0]['reading'] for row in rows.values()))
    for index, chars in enumerate(groups, 1):
        cards = [rows[c] for c in chars]
        questions = []
        for i, card in enumerate(cards):
            distractors = list(dict.fromkeys(
                [r['meaning'] for r in cards if r['meaning'] != card['meaning']]
                + [m for m in all_meanings if m != card['meaning']]
            ))[:3]
            choices = [card['meaning']] + distractors
            correct = (index + i) % 4
            choices.insert(correct, choices.pop(0))
            questions.append({
                'Prompt': f'Apa gloss sumber untuk kanji {card["character"]}?',
                'Choices': choices,
                'Correct': correct,
                'Explanation': f'{card["character"]}: {card["meaning"]}. Contoh JMdict: {card["examples"][0]["japanese"]}（{card["examples"][0]["reading"]}）— {card["examples"][0]["meaning"]}.',
            })
        for i, card in enumerate(cards):
            example = card['examples'][0]
            distractors = list(dict.fromkeys(
                [r['examples'][0]['reading'] for r in cards if r['examples'][0]['reading'] != example['reading']]
                + [reading for reading in all_readings if reading != example['reading']]
            ))[:3]
            choices = [example['reading']] + distractors
            correct = (index + i + 1) % 4
            choices.insert(correct, choices.pop(0))
            questions.append({
                'Prompt': f'Bagaimana membaca「{example["japanese"]}」?',
                'Choices': choices,
                'Correct': correct,
                'Explanation': f'{example["japanese"]} dibaca {example["reading"]} — {example["meaning"]}. Bacaan mengikuti entri JMdict.',
            })
        lessons.append({
            'ID': f'kanji-n1-{index:03d}',
            'Chapter': 'Kanji N1',
            'Title': f'Kanji N1 {index:03d} · {"".join(chars)}',
            'Characters': cards,
            'Questions': questions,
        })

    flattened = [card['character'] for lesson in lessons for card in lesson['Characters']]
    if len(flattened) != len(set(flattened)) or set(flattened) != set(dictionary):
        raise ValueError('N1 curriculum must cover every reference character exactly once')
    destination = ROOT / 'backend/data/kanji/n1.json'
    destination.write_text(json.dumps(lessons, ensure_ascii=False, indent=2) + '\n', encoding='utf-8')
    curriculum = ['| Lesson | Kanji |', '|---|---|'] + [
        f'| {i:03d} | {" ".join(c["character"] for c in lesson["Characters"])} |'
        for i, lesson in enumerate(lessons, 1)
    ]
    (ROOT / 'docs/kanji-n1-curriculum.md').write_text('\n'.join(curriculum) + '\n', encoding='utf-8')
    print(f'{len(lessons)} lessons, {len(rows)} kanji, {sum(len(l["Questions"]) for l in lessons)} questions')


if __name__ == '__main__':
    build()
