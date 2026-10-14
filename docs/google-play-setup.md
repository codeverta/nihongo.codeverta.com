# Setup Google Play — Nihongo Master

Halaman di bawah dilayani langsung oleh binary backend Go, tanpa frontend/Docker tambahan, login, cookie, atau form yang mengubah database. File HTML masuk ke binary melalui `go:embed`. Kontak publik: `contact@codeverta.com`, dikonfirmasi pemilik layanan. Tanggal kebijakan: 14 Oktober 2026.

| Keperluan | URL produksi |
|---|---|
| Privacy policy di Store listing / App content | https://nihongo.codeverta.com/privacy-policy |
| URL account deletion di Data safety | https://nihongo.codeverta.com/delete-account |
| Terms of service | https://nihongo.codeverta.com/terms-of-service |
| Pembelian / refund | https://nihongo.codeverta.com/refund-policy |
| Website dukungan | https://nihongo.codeverta.com/support |
| Ringkasan data untuk pengguna | https://nihongo.codeverta.com/data-safety |
| Indeks halaman | https://nihongo.codeverta.com/legal |

Alias `/privacy`, `/terms`, `/refunds`, `/account-deletion` redirect permanen ke halaman canonical. Root `/` menampilkan indeks. Halaman ini tidak menggantikan formulir Play Console.

## Checklist Play Console

1. **Privacy policy**: masukkan URL privasi di atas. Kebijakan menamai Nihongo Master/Codeverta, menjelaskan akun, aktivitas belajar, pembelian, kueri kamus, metadata jaringan, provider, kontak, keamanan, penghapusan, dan retensi.
2. **Data safety**: jawab berdasarkan AAB final beserta seluruh SDK, bukan hanya server. Email, nama, ID akun, foto profil Google opsional, aktivitas belajar, dan riwayat pembelian diproses untuk fungsi aplikasi. Kueri kamus diteruskan ke penyedia dan di-cache; metadata jaringan diterima infrastruktur. Jangan memilih “tidak mengumpulkan data”. Tentukan status collected/shared, optional/required, dan ephemeral menurut definisi Google per data dan jalur; transfer ke service provider dapat memiliki pengecualian “sharing”, tetapi tetap perlu diungkapkan dalam privasi. Tidak ada ad/behavior analytics SDK dalam source yang diperiksa. HTTPS publik harus aktif.
3. **Account deletion**: masukkan URL `/delete-account`. Jalur aplikasi: Profil → Hapus akun → konfirmasi. Jalur web: kirim email ke kontak dengan identitas akun, tanpa perlu memasang aplikasi. Permintaan email harus benar-benar dipantau, kepemilikan diverifikasi, dan diproses operator; halaman tidak mengeksekusi deletion otomatis. Jangan memasukkan password/token dalam permintaan. Bukti pembelian minimum tetap disimpan tanpa nama/email/ID akun lama. Backup belum memiliki expiry otomatis: tentukan operasional retensi dan penanganan backup sebelum menyatakan batas hari tertentu pada Console.
4. **App access / sign in**: berikan akun reviewer khusus yang benar-benar dapat dipakai beserta langkah untuk fitur berlogin/premium. Simpan kredensial hanya di bagian reviewer Play Console, bukan repository atau halaman publik. Uji login, purchase/restore, dan hapus akun pada akun test. Katalog tamu tidak berarti seluruh aplikasi bebas login.
5. **Ads**: source saat ini tidak memasang SDK iklan. Konfirmasi semua SDK pada AAB final sebelum menjawab “No”.
6. **Content rating**: isi questionnaire IARC berdasarkan seluruh materi, fitur, dan interaksi sebenarnya. Jangan menentukan rating hanya dari halaman kebijakan.
7. **Target audience**: pilih rentang usia berdasarkan desain/konten dan strategi distribusi yang benar. Source tidak meminta tanggal lahir; itu tidak menetapkan bahwa aplikasi termasuk/tidak termasuk Families. Bila memilih anak, tinjau persyaratan Families, SDK, consent, dan privasi tambahan.
8. **Government apps / Financial features / Health**: aplikasi yang diperiksa adalah pendidikan bahasa, bukan layanan pemerintah/keuangan/kesehatan. Jawab deklarasi berdasarkan fitur final; in-app purchase sendiri tidak membuktikan adanya produk keuangan. Lengkapi form yang diminta Console.
9. **Category & contact**: gunakan kategori Education bila sesuai listing, email `contact@codeverta.com`, dan URL support. Isi identitas/alamat/telepon developer sesuai identitas pemilik di Console.
10. **Store listing**: isi nama Nihongo Master, deskripsi, icon, feature graphic, screenshot asli, negara/bahasa, dan informasi premium yang konsisten. Tidak boleh mengklaim sebagai aplikasi resmi JLPT/JFT/NAT.
11. **Testing & release**: lengkapi persyaratan closed testing yang ditampilkan pada akun Console, signing, versionCode unik, dan test store menggunakan license tester. Review manual kebijakan dan Data safety harus konsisten dengan AAB yang dikirim.

## Tautan dalam aplikasi

Profil menampilkan tautan privasi, ketentuan, refund, penghapusan data, dan bantuan untuk akun maupun tamu. Tautan membuka browser eksternal tanpa token login. Default origin `https://nihongo.codeverta.com`; override saat build dengan `--dart-define=LEGAL_BASE_URL=https://...` hanya bila domain berubah. Perubahan Flutter baru tersedia kepada pengguna setelah AAB berikutnya dibangun/didistribusikan; deploy backend tidak mengubah APK yang sudah terpasang. Tidak perlu membangun Android untuk memasang halaman backend ini.

## Operasional dan perubahan kebijakan

- Jaga mailbox kontak aktif. Permintaan web menggunakan email; tidak ada inbox database atau email keluar otomatis dari backend legal pages.
- Jangan menjanjikan purge backup otomatis atau batas retensi yang belum diterapkan. Deployment membuat backup dan mempertahankan rilis lama. Restore database membutuhkan penerapan ulang permintaan penghapusan yang telah diproses.
- Perubahan penggunaan data/SDK membutuhkan pembaruan kebijakan, tanggal berlaku, serta jawaban Data safety. Template publik tidak boleh berisi secrets atau detail reviewer.
- Pembayaran diproses store; jangan membuat aturan “refund selalu ditolak” atau menjamin refund. Halaman mengarahkan ke proses Google/Apple dan mempertahankan hak wajib konsumen.
- Halaman tersedia sebagai HTML publik di HTTPS origin yang sama. Cloudflare Access, geofencing, atau challenge login tidak boleh diterapkan pada URL legal.

## Sumber resmi

- [Google Play User Data](https://support.google.com/googleplay/android-developer/answer/10144311?hl=en)
- [Persyaratan penghapusan akun](https://support.google.com/googleplay/android-developer/answer/13327111?hl=en)
- [Form Data safety](https://support.google.com/googleplay/android-developer/answer/10787469?hl=en)
- [Refund Google Play](https://support.google.com/googleplay/answer/2479637?hl=en)

Persyaratan Console dapat berubah. Halaman refund/terms/support membantu transparansi, sedangkan privacy policy dan jalur penghapusan adalah bagian utama persyaratan data pengguna untuk aplikasi berakun.
