"""Create paged backend flashcards from the lossless, pinned Kotoba mirror."""
import gzip
import html
import json
from pathlib import Path
import random
import re
from urllib.parse import quote

ROOT = Path(__file__).resolve().parents[1] / 'backend/internal/anki/resources'
CHUNK_SIZE = 256
PREFECTURES = set('aichi akita aomori chiba ehime fukui fukuoka fukushima gifu gunma hiroshima hokkaido hyogo hyougo ibaraki ishikawa iwate kagawa kagoshima kanagawa kochi kumamoto kyoto mie miyagi miyazaki nagano nagasaki nara niigata oita okayama okinawa osaka saga saitama shiga shimane shizuoka tochigi tokushima tokyo tottori toyama wakayama yamagata yamaguchi yamanashi'.split())
TOPICS = {
    'buddha': ('Istilah Buddhisme', 'Membaca kosakata dan istilah tradisi Buddhisme dalam bahasa Jepang'),
    'dqn': ('Nama Jepang dengan Bacaan Unik', 'Latihan nama dengan penulisan dan bacaan tidak biasa dari koleksi Kotoba'),
    'kirakira': ('Nama Kirakira', 'Nama Jepang kreatif yang dapat memiliki bacaan tidak langsung dari kanjinya'),
    'onago': ('Nama Perempuan Jepang', 'Bacaan nama perempuan Jepang dari berbagai generasi'),
    'emperors': ('Kaisar dan Shogun', 'Membaca nama kaisar, shogun, serta tokoh terkait dalam sejarah Jepang'),
    'seiyuu': ('Nama Pengisi Suara', 'Membaca nama seiyuu, yaitu pengisi suara Jepang'),
    'hentaigana': ('Hentaigana', 'Bentuk historis hiragana. Karakter langka dapat membutuhkan dukungan font pada perangkat'),
    'hard': ('Bacaan Sulit', 'Kosakata dengan bacaan menantang menurut koleksi Kotoba'),
    'haard': ('Bacaan Sangat Sulit', 'Kelanjutan bacaan sulit dengan pilihan kata yang lebih menantang dari deck Hard'),
    'insane': ('Tantangan Bacaan Ekstrem', 'Koleksi bacaan yang sangat sulit dari sumber Kotoba'),
    'jpdbtop30k': ('JPDB: 30.000 Kosakata', 'Kosakata dari daftar frekuensi JPDB hingga peringkat 30.000 pada sumber Kotoba'),
    'jpdefs': ('Definisi Bahasa Jepang', 'Ingat kosakata berdasarkan penjelasan dalam bahasa Jepang dari Wiktionary sumber'),
    'jsyn': ('Sinonim Bahasa Jepang', 'Padanan kata untuk kosakata umum dari Japanese WordNet'),
    'jsyn_full': ('Sinonim Bahasa Jepang Lengkap', 'Koleksi lebih luas padanan kata dari Japanese WordNet'),
    'k33': ('Koleksi Kosakata K33', 'Koleksi kosakata lanjutan yang dipakai dalam deck pribadi K33 pada sumber Kotoba'),
    'kklc': ('Kodansha Kanji Learner’s Course', 'Latihan bacaan yang dikelompokkan menurut koleksi Kodansha Learner’s Course pada Kotoba'),
    'longwords': ('Kosakata Panjang', 'Kata dengan sedikitnya empat karakter dan mayoritas kanji dari koleksi JMdict sumber'),
    'numbers': ('Angka dalam Kanji', 'Mengingat nilai angka yang dituliskan menggunakan kanji'),
    'places_full': ('Nama Tempat Jepang', 'Koleksi besar nama tempat Jepang dan bacaannya dari ENAMDICT sumber'),
    'pokemon': ('Nama Pokémon', 'Mengenali Pokémon dari gambar dan mengingat nama Jepang atau Inggrisnya'),
    'ranobe': ('Kalimat Light Novel', 'Membaca serta mengingat kalimat dari koleksi novel sumber Kotoba'),
    'sao': ('Kosakata Sword Art Online', 'Kosakata dari volume pertama Sword Art Online: Aincrad pada sumber Kotoba'),
    'trivia': ('Trivia Bahasa Inggris', 'Pertanyaan pengetahuan umum berbahasa Inggris dari arsip Jeopardy sumber'),
    'jptrivia': ('Trivia Bahasa Jepang', 'Pertanyaan pengetahuan umum yang ditulis dalam bahasa Jepang'),
    'vnfreq': ('Kosakata Visual Novel', 'Kosakata yang diurutkan berdasarkan daftar frekuensi visual novel pada sumber Kotoba'),
    'ejtrans': ('Terjemahan Inggris–Jepang', 'Mengingat padanan Jepang untuk kata atau frasa bahasa Inggris'),
    'esyn_easy': ('Sinonim Inggris: Dasar', 'Koleksi kata Inggris untuk latihan padanan kata tingkat dasar'),
    'esyn_med': ('Sinonim Inggris: Menengah', 'Koleksi kata Inggris untuk latihan padanan kata tingkat menengah'),
    'esyn_hard': ('Sinonim Inggris: Sulit', 'Koleksi kata Inggris untuk latihan padanan kata tingkat sulit'),
    'animals': ('Hewan', 'Nama hewan dalam bahasa Jepang'), 'birds': ('Burung', 'Nama dan bacaan berbagai jenis burung'),
    'bugs': ('Serangga', 'Kosakata serangga dan hewan kecil'), 'fish': ('Ikan', 'Nama ikan dan hewan perairan'),
    'plants': ('Tumbuhan', 'Nama tumbuhan dan bacaan kanjinya'), 'vegetables': ('Sayuran', 'Nama sayuran dalam bahasa Jepang'),
    'nature': ('Alam', 'Kosakata alam dan lingkungan'), 'countries': ('Negara', 'Nama negara dalam bahasa Jepang'),
    'cities': ('Kota', 'Nama kota dan cara membacanya'), 'prefectures': ('Prefektur Jepang', 'Nama prefektur Jepang dan bacaannya'),
    'myouji': ('Nama Keluarga Jepang', 'Bacaan nama keluarga Jepang'), 'namae': ('Nama Depan Jepang', 'Bacaan nama pribadi Jepang'),
    'jmndict': ('JMnedict: Nama Jepang', 'Koleksi besar bacaan nama Jepang dari sumber Kotoba'),
    'jmdict': ('JMdict: Kosakata Jepang', 'Koleksi besar kosakata dan bacaan dari sumber Kotoba'),
    'common': ('Kosakata Umum', 'Mengenali bacaan kosakata yang umum digunakan'),
    'common_nojlpt': ('Kosakata Umum di Luar JLPT', 'Kata umum yang tidak dikelompokkan dalam deck JLPT sumber'),
    'hiragana': ('Hiragana', 'Mengenali hiragana dan menuliskan bunyinya dalam romaji'),
    'katakana': ('Katakana', 'Mengenali katakana dan menuliskan bunyinya dalam romaji'),
    'kanawords': ('Kata dalam Kana', 'Membaca kosakata yang ditulis dalam kana'),
    'jouyou': ('Kanji Jōyō', 'Membaca kanji yang digunakan secara umum di Jepang'),
    'radicals': ('Radikal Kanji', 'Mengenali nama komponen dasar kanji'),
    'kokuji': ('Kokuji', 'Kanji yang dibentuk di Jepang'),
    'yojijukugo': ('Yojijukugo', 'Ungkapan Jepang yang tersusun atas empat kanji'),
    'onomato': ('Onomatope Jepang', 'Kata tiruan bunyi dan penggambaran keadaan'),
    '擬音語': ('Giongo', 'Kata tiruan bunyi dalam bahasa Jepang'),
    'igaku': ('Istilah Kedokteran', 'Bacaan istilah bidang kedokteran'),
    'houritsu': ('Istilah Hukum', 'Bacaan istilah bidang hukum'),
    'keizai': ('Istilah Ekonomi', 'Bacaan istilah bidang ekonomi'),
    'kenchiku': ('Istilah Arsitektur', 'Bacaan istilah bangunan dan arsitektur'),
    'shinrigaku': ('Istilah Psikologi', 'Bacaan istilah bidang psikologi'),
    'suugaku': ('Matematika', 'Bacaan istilah matematika'),
    'rikagaku': ('Sains', 'Bacaan istilah bidang sains'),
    'pasokon': ('Komputer', 'Kosakata komputer dan teknologi'),
    'images': ('Tebak Gambar', 'Mengingat kata dari gambar sumber'),
    'strokes_kanken': ('Urutan Goresan Kanken', 'Mengenali nomor goresan kanji yang ditandai'),
}


def plain(value):
    text = str(value or '')
    text = re.sub(r'\[([^\]]+)\]\([^)]*\)', r'\1', text)
    text = re.sub(r'<[^>]+>', '', text)
    return html.unescape(text.replace('**', '').replace('__', '').replace('||', '')).strip()


def metadata(id, source):
    category = 'Kosakata'
    level = ''
    title, intro = TOPICS.get(id, (source.get('name', id).removesuffix(' Quiz'), ''))
    if id in PREFECTURES:
        category = 'Nama & Geografi'
        title = 'Nama Tempat ' + id.capitalize()
        intro = (f'{id.capitalize()} adalah salah satu prefektur di Jepang. Deck ini berisi nama kota, kawasan, dan lokasi dari prefektur tersebut. Latih membaca nama tempat dalam kanji lalu cocokkan dengan bacaan kana di belakang kartu.')
    elif re.fullmatch(r'j?(10|[1-9])k', id):
        category = 'Kana & Kanji'
        grade = id.removeprefix('j').removesuffix('k')
        stage = ('Pra-' if id.startswith('j') else '') + grade
        title = f'Kanken Tingkat {stage}'
        intro = f'Latihan bacaan kanji tingkat {stage} pada Kanji Kentei menurut sumber Kotoba. Penomoran Kanken berbeda dari JLPT: tingkat 10 menjadi fondasi, sedangkan tingkat 1 jauh lebih lanjut.'
    elif id.startswith('stations_'):
        category = 'Stasiun'
        place = id.removeprefix('stations_').capitalize()
        title = 'Stasiun Jepang' if id == 'stations_full' else f'Stasiun {place}'
        intro = f'Latihan membaca nama stasiun {"di berbagai wilayah Jepang" if id == "stations_full" else "di wilayah " + place}. Nama stasiun sering memiliki bacaan khusus; hafalkan sebagai nama utuh bersama bacaan kananya.'
    elif re.fullmatch(r'(k_)?n[1-5]m?', id):
        level = re.search(r'n[1-5]', id)[0].upper()
        category = 'JLPT'
        kind = 'Penggunaan Kanji' if id.startswith('k_') else 'Arti Kosakata' if id.endswith('m') else 'Bacaan Kosakata'
        title = f'{level} · {kind}'
        intro = f'Kumpulan {kind.lower()} yang dikelompokkan sebagai JLPT {level} oleh Kotoba. Pelajari bagian depan, ingat jawabannya, lalu cocokkan semua variasi jawaban di belakang kartu.'
    elif id in {'hiragana', 'katakana', 'hentaigana', 'jouyou', 'radicals', 'kokuji', 'kklc', 'k33', 'strokes_kanken'}:
        category = 'Kana & Kanji'
    elif id in {'myouji', 'namae', 'jmndict', 'cities', 'countries', 'prefectures', 'places_full', 'emperors', 'seiyuu', 'dqn', 'kirakira', 'onago'}:
        category = 'Nama & Geografi'
    elif id.startswith(('anagrams', 'esyn')) or id in {'trivia', 'ejtrans'}:
        category = 'Inggris & Trivia'
    elif id in {'pokemon', 'ranobe', 'sao', 'jptrivia'}:
        category = 'Budaya & Trivia'
    if id.startswith('anagrams'):
        title = f'Anagram Inggris · {id.removeprefix("anagrams")} Huruf'
        intro = 'Susun kembali huruf di bagian depan menjadi kata bahasa Inggris. Bagian belakang memuat variasi jawaban dari sumber.'
    if intro and not id in PREFECTURES and not id.startswith(('stations_', 'k_n')) and not level and id in TOPICS:
        intro += '. Gunakan preview untuk mengenali bentuk kartu, lalu latih mengingat jawabannya sebelum membalik kartu.'
    if not intro:
        intro = f'{title}. Kumpulan kartu bertema {plain(source.get("name", id))} dari Kotoba. Gunakan preview di bawah untuk mengenali bentuk pertanyaan dan jenis jawaban sebelum belajar.'
    instruction = ('Sebutkan contoh kata yang mengandung kanji ini. Cocokkan dengan contoh jawaban dari sumber di belakang kartu.' if id.startswith('k_n') else
                   'Baca kalimat di depan, lalu bandingkan dengan transkripsi sumber di belakang kartu.' if id == 'ranobe' else
                   'Sebutkan sinonim kata Jepang di depan kartu.' if id in {'jsyn','jsyn_full'} else
                   'Terjemahkan kata bahasa Inggris di depan ke bahasa Jepang.' if id == 'ejtrans' else
                   'Ingat kosakata Jepang yang sesuai dengan definisi di depan.' if id == 'jpdefs' else
                   'Ingat angka yang dituliskan dalam kanji ini.' if id == 'numbers' else
                   'Ingat nama objek pada gambar.' if source.get('questionCreationStrategy') == 'IMAGE_URI' else
                   'Susun huruf menjadi kata.' if id.startswith('anagrams') else
                   'Ingat arti kata sebelum membuka jawaban.' if id.endswith('m') and level else
                   'Ingat jawaban sebelum membalik kartu; cocokkan dengan semua bacaan atau jawaban yang tersedia.')
    if id.startswith('k_n'):
        intro = f'Latihan penggunaan kanji JLPT {level} menurut pengelompokan Kotoba. Lihat satu kanji, sebutkan contoh kata yang memuatnya, lalu buka daftar contoh jawaban. Fokus deck ini adalah menghubungkan kanji dengan kosakata.'
    return dict(id=id, title=title, category=category, level=level, description=intro,
                instructions=instruction, originalName=source.get('name', id),
                originalDescription=plain(source.get('description', '')),
                sourceFile=id + '.json', sourceUrl='https://github.com/mistval/kotoba',
                meaningLanguage='Mengikuti sumber; kebanyakan arti menggunakan bahasa Inggris.')


def main():
    decks = []
    for path in sorted((ROOT / 'source').glob('*.json.gz')):
        source = json.loads(gzip.decompress(path.read_bytes()))
        id = path.name.removesuffix('.json.gz').lower()
        deck = metadata(id, source)
        deck['sourceFile'] = path.name.removesuffix('.gz')
        deck['sourceUrl'] = 'https://github.com/mistval/kotoba/blob/a0f8f984bf5dd8014527c6608d36b479ab622b07/resources/quiz_data/' + quote(deck['sourceFile'])
        deck['sourceCount'] = len(source['cards'])
        reason = ''
        strategy = source.get('cardPreprocessingStrategy', 'NONE')
        if strategy == 'THESAURUS_SYNONYMS':
            reason = 'Data sumber tidak menyertakan jawaban sinonim. Kotoba membentuk jawabannya melalui layanan thesaurus eksternal; sesi belajar belum tersedia.'
        elif strategy == 'HIGHLIGHT_STROKE':
            reason = 'Deck ini memerlukan diagram goresan serta jawaban dinamis dari mesin Kotoba. Data sumber tersimpan; sesi belajar belum tersedia.'
        cards = []
        if not reason:
            for source_index, raw in enumerate(source['cards']):
                if not isinstance(raw, dict):
                    continue
                front = plain(raw.get('question'))
                answers = [plain(a) for a in raw.get('answer', []) if isinstance(a, (str, int, float)) and plain(a)]
                if not front or not answers:
                    continue
                if strategy == 'RANDOMIZE_QUESTION_CHARACTERS':
                    chars = list(front)
                    generator = random.Random(f'{id}:{source_index}')
                    for _ in range(30):
                        generator.shuffle(chars)
                        if ''.join(chars) not in answers:
                            break
                    front = ''.join(chars)
                image = front if source.get('questionCreationStrategy') == 'IMAGE_URI' and front.startswith('https://') else ''
                cards.append(dict(index=len(cards), sourceIndex=source_index, front='Apa yang ada pada gambar?' if image else front,
                                  answers=answers, meaning=plain(raw.get('meaning')), imageUrl=image))
        deck.update(cardCount=len(cards), available=bool(cards), unavailableReason=reason,
                    skippedCount=deck['sourceCount'] - len(cards), preview=cards[:3], chunkSize=CHUNK_SIZE)
        if not cards and not reason:
            deck['unavailableReason'] = 'Belum ada pasangan pertanyaan dan jawaban yang valid untuk dipelajari.'
        for start in range(0, len(cards), CHUNK_SIZE):
            chunk = ROOT / 'cards' / id / f'{start // CHUNK_SIZE:06d}.json.gz'
            chunk.parent.mkdir(parents=True, exist_ok=True)
            data = json.dumps(cards[start:start+CHUNK_SIZE], ensure_ascii=False, separators=(',', ':')).encode()
            chunk.write_bytes(gzip.compress(data, mtime=0))
        decks.append(deck)
        print(f'{id}: {len(cards)}/{deck["sourceCount"]}', flush=True)
    if len(decks) != 191:
        raise ValueError('Expected all 191 source decks before publishing the catalog')
    (ROOT / 'catalog.json').write_text(json.dumps(dict(revision='a0f8f984bf5dd8014527c6608d36b479ab622b07', decks=decks), ensure_ascii=False, separators=(',', ':')) + '\n', encoding='utf-8')
    print('Ready:',len(decks),'decks;',sum(d['cardCount'] for d in decks),'study cards')


if __name__ == '__main__':
    import sys
    if '--metadata-only' in sys.argv:
        path = ROOT / 'catalog.json'
        catalog = json.loads(path.read_text(encoding='utf-8'))
        for deck in catalog['decks']:
            source = json.loads(gzip.decompress((ROOT / 'source' / (deck['sourceFile'] + '.gz')).read_bytes()))
            update = metadata(deck['id'], source)
            for field in ['title', 'category', 'level', 'description', 'instructions']:
                deck[field] = update[field]
            deck['sourceUrl'] = 'https://github.com/mistval/kotoba/blob/a0f8f984bf5dd8014527c6608d36b479ab622b07/resources/quiz_data/' + quote(deck['sourceFile'])
        path.write_text(json.dumps(catalog, ensure_ascii=False, separators=(',', ':')) + '\n', encoding='utf-8')
        print('Updated catalog metadata without rewriting study chunks')
    else:
        main()
