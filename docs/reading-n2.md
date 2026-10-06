# Reading N2

Buka **Belajar → N2 → Reading** setelah login. Tersedia 12 lesson orisinal,
36 bagian bacaan, 72 entri kosakata kontekstual, dan 60 soal dengan pembahasan
Indonesia. Materi ini latihan buatan aplikasi, bukan soal resmi JLPT.

Setiap lesson menyediakan teks Jepang, bacaan kana, terjemahan Indonesia,
dan dua kosakata penting per bagian. Coba baca tanpa bantuan dahulu, lalu
buka kana atau terjemahan untuk meninjau bagian yang sulit.

Bacaan dibuat lebih panjang: **844–915 karakter Jepang per lesson**
(termasuk tanda baca, tanpa baris baru), dibagi menjadi tiga bagian dengan
dua paragraf per bagian. Estimasi belajar adalah **22 menit per lesson**,
termasuk membaca dan menjawab kuis. Paragraf tambahan memperluas contoh,
alasan, dan batas kesimpulan; bantuan kana dan terjemahan mencakup seluruh
teks. Soal tetap mengikuti informasi dalam bacaan.

| Lesson | Fokus |
| --- | --- |
| 01 Perpustakaan dan kesenjangan digital | Kesimpulan, batas klaim, akses layanan |
| 02 Harga murah dan nilai pasar lokal | Rujukan kata, nilai di luar harga |
| 03 Rapat yang benar-benar terbuka | Penyebab, inferensi, kebebasan formal |
| 04 Program pengurangan sisa makanan | Membaca data dan faktor pembanding |
| 05 Surat dan makna menunggu | Perubahan pandangan narator |
| 06 Ketentuan perubahan reservasi | Tenggat, pengecualian, pencarian informasi |
| 07 Kota yang ramah sepeda | Kepentingan berbeda dan kompromi |
| 08 Notifikasi dan kendali perhatian | Hubungan sebab-akibat dan batas solusi |
| 09 Latihan bencana | Kendala partisipasi dan akses pengetahuan |
| 10 Dua pendapat tentang komunikasi | Pemahaman terpadu teks A dan B |
| 11 Daur ulang dan konsumsi berlebihan | Paradoks dan argumen penulis |
| 12 Perubahan kecil di toko roti | Bukti, evaluasi, batas generalisasi |

## Penilaian dan progress

Jawaban dinilai di backend. Empat jawaban benar dari lima (80%) menyelesaikan
lesson. Progress N2 dihitung dari jumlah lesson yang lulus dibagi 12, dibulatkan
ke bawah. Nilai terbaik dan status selesai tetap dipertahankan jika hasil retry
lebih rendah. XP diberikan sekali pada kelulusan pertama, sesuai alur reading
yang sudah ada. Progress pengguna dan level lain terpisah.

## Integrasi developer

- Konten: `backend/data/reading/n2.json`, di-embed ketika Go dibuild.
- Katalog: `GET /v1/reading/course?level=N2` (memerlukan bearer token).
- Detail: `GET /v1/reading/lessons/reading-n2-01` sampai `reading-n2-12`.
- Mulai kuis, kirim jawaban, dan complete memakai endpoint reading yang sama
  seperti level sebelumnya, terdokumentasi di `backend/openapi.yaml`.
- Restart backend setelah mengubah JSON karena kontennya di-embed.
- Ilustrasi pendukung memakai kembali aset N3 yang tersedia. Tidak ada aset N2
  baru atau upload COS dalam penambahan ini. URL COS mengikuti konfigurasi
  asset base URL yang sudah ada, dengan fallback ke aset lokal di Flutter.
- Penyimpanan progress mengikuti implementasi store aplikasi yang sudah ada;
  penambahan kurikulum ini tidak mengganti mekanisme persistence.

Jalankan pemeriksaan dari folder `backend`:

```powershell
go test ./internal/store ./internal/api -run ReadingN2
```
