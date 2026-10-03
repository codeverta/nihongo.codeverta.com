# Vocabulary N5 bergambar

## Isi materi

20 lesson bertema, masing-masing 8 entri kosakata, 16 contoh kalimat, dan 24 soal. Total: 160 entri, 320 contoh, dan 480 soal. Kata yang sama dapat muncul pada tema berbeda, seperti 魚 dalam makanan dan hewan. Paket ini merupakan fondasi N5; angka tersebut tidak menyatakan seluruh kosakata N5 telah tercakup.

Tema: perkenalan, keluarga, sekolah, rumah, makanan, minuman dan buah, aktivitas sehari-hari, belajar dan bekerja, tempat di kota, transportasi, waktu, hari, cuaca dan alam, hewan, tubuh, pakaian, warna, ukuran dan sifat, posisi dan arah, serta hobi.

## Alur belajar

1. Buka **Belajar → Vocabulary → N5**. Pilih lesson atau cari berdasarkan tema.
2. Baca kata Jepang, kana, arti, serta ilustrasinya. Setiap kata memiliki dua contoh dengan bantuan kana dan terjemahan yang dapat disembunyikan.
3. Buka catatan pemakaian jika perlu, lalu lanjutkan hingga kata kedelapan.
4. Kerjakan 24 soal: 8 arti, 8 cara baca/penulisan, dan 8 pemahaman kalimat. Kata yang sudah ditulis sepenuhnya dalam kana memakai soal penulisan dari arti, agar pertanyaannya tidak sekadar menyalin tulisan yang sama.
5. Baca penjelasan setelah mengirim jawaban. Nilai minimal kelulusan 80%; kelulusan pertama memberi 50 XP. Nilai terbaik tersimpan per pengguna.

Bantuan kana adalah satu baris bacaan lengkap, bukan furigana di atas setiap kanji. Pada gangguan jaringan, layar kuis mempertahankan pilihan dan soal agar pengiriman dapat dicoba lagi.

## Sumber dan gambar

- Materi: `backend/data/vocabulary/n5.json`, dimuat dengan Go embed saat backend dibangun.
- Soal dan penilaian: `backend/internal/store/vocabulary.go`. API tidak mengirim kunci jawaban ketika kuis dimulai.
- Gambar: `assets/images/vocabulary/n5-theme-01.png` hingga `n5-theme-20.png`, dibuat menggunakan imagegen bawaan. Setiap lembar berisi delapan ilustrasi dalam susunan 4 kolom × 2 baris; aplikasi mengambil panel sesuai kata.
- Prompt lengkap dan pemetaan aset: `docs/vocabulary-n5-art.json`.
- Gambar lokal dibundel dengan aplikasi. URL aset backend, bila dikonfigurasi, memiliki fallback ke gambar lokal.

## Menjalankan pembaruan

Mulai ulang backend setelah mengubah JSON karena kontennya di-embed:

```powershell
cd backend
go run ./cmd/api
```

Jalankan ulang Flutter untuk menyertakan aset baru:

```powershell
flutter run -d emulator-5556 --dart-define=API_BASE_URL=http://10.0.2.2:8080/v1
```

Sesuaikan ID emulator dengan `adb devices`. Backend saat ini menyimpan sesi, attempt, dan progres dalam memori; restart backend menghapus data tersebut dan pengguna perlu login kembali. Panduan lengkap ada di [running-project.md](running-project.md).
