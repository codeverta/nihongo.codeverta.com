# JLPT N5 dan N4 Paket 02–10 (dengan furigana)

Tambahan latihan ujian **gratis** untuk N5 dan N4: N5 Paket 02–10 (masing-masing 38 soal) serta
N4 Paket 02–10 (masing-masing 57 soal). Soal orisinal Nihongo Master; hanya struktur dan jenis soal
yang mengikuti pedoman resmi JLPT. Listening dilewati, dan N5 juga melewati dokkai seperti Paket 01.

## Struktur

| Level | 文字・語彙 | 文法 (・読解) | Timer |
|---|---|---|---|
| N5 | 21 soal (7 bacaan, 5 penulisan, 6 konteks, 3 padanan) | 17 soal (9 grammar, 4 susunan ★, 4 grammar teks) | 20 + 20 menit |
| N4 | 28 soal (7, 5, 8, 4 padanan, 4 penggunaan) | 29 soal (13 grammar, 4 ★, 4 grammar teks, 8 reading) | 25 + 55 menit |

## Furigana

Teks soal memakai markup `{漢字|かんじ}` yang ditampilkan Flutter sebagai ruby di atas kanji
(`SimulationText`). Markup `[[...]]` (garis bawah untuk kata yang diuji) tetap berlaku dan dapat digabung.
Aturan isi:

- Kanji yang diuji pada soal 漢字の読み tidak diberi furigana (jawabannya akan terbuka).
- Pilihan jawaban soal 表記 yang berupa kanji tidak diberi furigana.
- Kanji lain pada instruksi, teks bacaan, kalimat soal, dan pilihan diberi furigana.
- Pembahasan (Bahasa Indonesia) tidak memakai markup.

## ID dan level

| Paket | Level API | Exam ID | Paper ID |
|---|---|---|---|
| N5 Paket 02–10 | `N5-2` … `N5-10` | `jlpt-n5-simulation-2` … `-10` | `n5-language-02` … `-10` |
| N4 Paket 02–10 | `N4-2` … `N4-10` | `jlpt-n4-simulation-2` … `-10` | `n4-language-02` … `-10` |

Rute Flutter: `/exams/n5-{2..10}/simulation` dan `/exams/n4-{2..10}/simulation`. Overview: `GET /simulation/{level}`.

## Akses

Semua simulasi JLPT N5 dan N4 (Paket 01–10) gratis (`jlptSimulationFree`, `IsFreeJLPTSimulationExam`,
dan `RequiredTier: TierFree` pada katalog ujian). N3, N2, dan N1 tetap premium.

## Bahasa Inggris

Soal dan instruksi berbahasa Jepang dan tidak diterjemahkan. Pembahasan jawaban Paket 02–05 memiliki
terjemahan Inggris eksak di `backend/internal/api/english_jlpt_simulation_extra.go`, yang dikunci oleh teks
pembahasan Indonesia. Paket 06–10 menyimpan terjemahan Inggris langsung di JSON (`explanationEn`) dan
diambil lewat `store.SimulationJSONExplanationsEN()`. Tes `TestEnglishCatalogCoversExtraJLPTPapers`
memastikan semua paket terpetakan.

## Penyimpanan konten: JSON untuk Paket 06 ke atas

Paket 02–05 berupa literal Go. Paket 06 dan seterusnya disimpan sebagai JSON di
`backend/internal/store/simulation_data/jlpt_<n5|n4>_paper<NN>.json` dan dibundel ke binary dengan
`//go:embed` (`simulation_json.go`), sehingga tidak perlu berkas eksternal saat deploy. Isi dapat diedit
tanpa menyentuh kode Go. Struktur berkas:

```json
{
  "level": "N5",
  "paper": 6,
  "items": [
    {
      "section": "vocabulary",
      "group": "...", "instructions": "...", "passage": "", "prompt": "...",
      "choices": ["jawaban benar", "salah 1", "salah 2", "salah 3"],
      "explanation": "Pembahasan Indonesia (tanpa markup)",
      "explanationEn": "English explanation"
    }
  ]
}
```

Pilihan pertama selalu jawaban benar (diacak server per percobaan). Loader memvalidasi 4 pilihan dan
kedua pembahasan per soal; berkas rusak membuat server panic saat pertama dimuat, dan
`TestSimulationJSONPapersLoad` menangkapnya saat `go test`. Untuk menambah Paket 11: tambahkan berkas JSON,
perluas `jlptLastPaper` di `simulation.go`, katalog ujian di `memory.go`, serta loop paket di router dan `exam_hub_screen.dart`.

## Berkas

- `backend/internal/store/simulation_json.go` dan `simulation_data/*.json` (Paket 06–10)
- `backend/internal/store/simulation_json_test.go`
- `backend/internal/api/english_jlpt_simulation_json.go`

- `backend/internal/store/simulation_n5_additional_content.go`, `simulation_n4_additional_content.go`
- `backend/internal/api/english_jlpt_simulation_extra.go` dan `english_jlpt_simulation_extra_test.go`
- `backend/internal/store/simulation_jlpt_extra_test.go`
- `lib/features/exam/presentation/widgets/simulation_text.dart` dan `test/features/exam/simulation_furigana_test.dart`
