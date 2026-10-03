# JLPT N5 Simulation · Paket 01

Simulasi ini berisi **soal orisinal** dengan bentuk soal, instruksi Jepang,
empat pilihan jawaban, timer per bagian, serta penguncian bagian seperti ujian.
Dokkai (pemahaman bacaan) dan listening dilewati sesuai cakupan produk.
Grammar teks tetap termasuk bunpou: empat soal memilih bentuk atau penghubung
dalam satu teks, bukan pertanyaan pemahaman dokkai.

## Struktur

| Bagian | Jenis soal | Jumlah |
|---|---|---:|
| 文字・語彙 | Bacaan kanji | 7 |
| 文字・語彙 | Penulisan kanji/katakana | 5 |
| 文字・語彙 | Kosakata dalam konteks | 6 |
| 文字・語彙 | Padanan makna | 3 |
| 文法 | Memilih bentuk grammar | 9 |
| 文法 | Menyusun kalimat dan memilih posisi ★ | 4 |
| 文法 | Grammar teks | 4 |

Total 38 soal, mengikuti perkiraan jumlah soal setelah revisi resmi 2020.
Jumlah ujian asli dapat sedikit berbeda pada setiap penyelenggaraan.

Timer kosakata: 20 menit. Timer grammar tanpa dokkai: **20 menit sebagai
penyesuaian aplikasi**. Durasi resmi 40 menit untuk bagian kedua mencakup grammar
sekaligus dokkai, sehingga simulasi ini tidak mengklaim memiliki durasi ujian
lengkap. Kedua bagian berjalan berurutan tanpa jeda otomatis.

Referensi resmi:

- [Komposisi dan waktu JLPT](https://www.jlpt.jp/e/guideline/testsections.html)
- [Revisi jumlah soal N5](https://jlpt.jp/sp/e/topics/202009091599642827.html)
- [Tabel jumlah soal sesudah revisi](https://www.jlpt.jp/userdata/image/N5eng.jpg)
- [Tujuan masing-masing jenis soal N5](https://jlpt.jp/e/guideline/pdf/n5_e_revised.pdf)
- [Skor berskala dan hasil resmi](https://www.jlpt.jp/e/guideline/results.html)
- [FAQ sistem skor JLPT](https://www.jlpt.jp/e/faq/)

## Alur pengguna

Masuk akun → tab **Ujian** → **Buka simulasi N5** → baca aturan → mulai.
Selama bagian aktif, pengguna bisa berpindah soal, mengganti/menghapus pilihan,
dan menandai soal ragu. Peta hanya menampilkan bagian aktif.

Pengiriman bagian meminta konfirmasi dan menghitung soal kosong. Mengirim
kosakata lebih awal langsung memulai timer grammar 20 menit. Setelah bagian
dikunci, pengguna tidak dapat kembali atau mengubahnya. Kunci dan pembahasan
Indonesia baru tersedia setelah grammar selesai.

Riwayat menampilkan sepuluh hasil terbaru. **Ulangi paket 01** memakai soal yang
sama dengan urutan pilihan baru, bukan paket soal baru. Hasil memuat persentase
benar, jumlah benar/kosong, durasi, performa tujuh jenis soal, serta pembahasan
yang dapat difilter ke jawaban salah atau belum dijawab. Tidak ada keputusan
lulus/gagal atau konversi ke skor resmi 0–180: JLPT resmi memakai skor berskala
dan simulasi ini tidak mencakup semua bagian.

## Timer, resume, dan koneksi

Backend memegang waktu dan status yang otoritatif. Bagian yang kedaluwarsa
berpindah pada deadline aslinya; membuka aplikasi kembali tidak menambah waktu.
Contoh: sesi ditinggal sejak awal dan dibuka menit ke-25 → grammar tersisa sekitar
15 menit. Jika baru dibuka menit ke-50 → seluruh sesi sudah dinilai pada deadline
40 menit. Saat offline, UI mengunci jawaban setelah timer lokal habis dan mencoba
sinkronisasi lagi. Server menolak jawaban setelah batas waktu, termasuk jawaban
untuk bagian lama atau bagian berikutnya.

Jawaban/tanda ragu baru ditampilkan tersimpan setelah server mengonfirmasi.
Jika koneksi putus saat menyimpan, aplikasi mengambil snapshot server untuk
memastikan pilihan yang benar-benar tersimpan. Tidak ada antrean jawaban offline
yang dapat dikirim melewati deadline.

Menutup aplikasi tidak membatalkan sesi. Pemanggilan start berulang melanjutkan
sesi aktif yang sama. Pengiriman menyertakan bagian yang diharapkan sehingga
retry pengiriman kosakata tidak dapat tanpa sengaja mengirim grammar.

**Batas penyimpanan saat ini:** mengikuti backend proyek yang masih memakai
MemoryStore, sesi, riwayat, dan autentikasi berada di memori proses. Resume
bertahan saat aplikasi mobile ditutup, tetapi tidak setelah proses backend
direstart. Untuk deployment produksi dengan restart/multiple replicas,
persistensikan session/attempt/answer ke database dan gunakan penyimpanan auth
yang terpusat. Ini belum merupakan penyimpanan database.

## Kontrak API

Seluruh endpoint simulasi memerlukan token login pengguna:

| Method | Endpoint | Fungsi |
|---|---|---|
| GET | `/v1/simulation/n5` | Struktur, riwayat, sesi aktif |
| POST | `/v1/exam-attempts` | Mulai/resume; `examId: jlpt-n5-simulation-1` |
| GET | `/v1/simulation/attempts/{id}` | Snapshot bagian aktif atau hasil akhir |
| PUT | `/v1/simulation/attempts/{id}/answers` | `questionId`, `optionId`, `marked`; opsi kosong menghapus pilihan |
| POST | `/v1/simulation/attempts/{id}/submit-section` | `section: vocabulary` atau `grammar` |

Snapshot aktif tidak berisi kunci, penjelasan, atau soal bagian berikutnya.
ID opsi dibuat acak dan tidak membocorkan posisi jawaban benar. Pilihan diacak
sekali per attempt dan tetap stabil ketika dilanjutkan. Riwayat ringkas tidak
memuat semua pembahasan; buka attempt selesai untuk membaca hasil lengkap.

Katalog lama `jlpt-n5-full-1` tidak ditampilkan lagi. Permintaan dari client lama
tetap mempertahankan pemeriksaan premium sebelum diarahkan ke simulasi baru;
ID baru tersedia untuk pengguna yang sudah login tanpa premium.

## Source dan pemeriksaan

- Konten: `backend/internal/store/simulation_content.go`.
- State machine dan penilaian: `backend/internal/store/simulation.go`.
- Endpoint: `backend/internal/api/simulation.go` dan `backend/openapi.yaml`.
- Flutter: `lib/features/exam/data/simulation_repository.dart` dan layar `simulation_*`.

```powershell
cd backend
go test ./internal/store ./internal/api -run TestSimulation
go vet ./...
cd ..
flutter test test/features/exam/simulation_test.dart
flutter analyze
flutter build apk --debug
```

Pengujian mencakup integritas jumlah/jenis soal, akses milik pengguna,
kunci yang tidak bocor, penguncian bagian, penghitungan salah/kosong,
retry idempotent, deadline/resume, kegagalan jaringan, tampilan ponsel kecil,
serta respons jaringan yang tiba setelah halaman dibuang.
