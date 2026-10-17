"""Original SSW 1 fishery course: capture fisheries and aquaculture."""
from ssw_course_authoring import Course

c=Course('fishery','fish',[
 dict(title='Japan Fisheries Agency · Keselamatan kerja perikanan',url='https://www.jfa.maff.go.jp/j/kikaku/anzen.html'),
 dict(title='Dainippon Suisankai · Cakupan dan teks SSW 1',url='https://suisankai.or.jp/skill/'),
 dict(title='Prometric · SSW 1 perikanan tangkap',url='https://www.prometric-jp.com/ssw/test_list/archives/9'),
 dict(title='Prometric · SSW 1 budi daya',url='https://www.prometric-jp.com/ssw/test_list/archives/8bmgetj8jgci'),
])
L=c.lesson
c.module('capture-general','Tangkap · Dasar dan lingkungan operasi','capture','Jenis perikanan, sumber daya, perencanaan operasi, cuaca, dan komunikasi.')
L('Perikanan tangkap dan tanggung jawab awak',
 'Perikanan tangkap mengambil organisme dari perairan dengan alat serta metode yang diizinkan. Pengetahuan jenis ikan, musim, daerah operasi, ukuran, dan ketentuan setempat membantu menjaga mutu serta sumber daya. Tugas awak dibagi sesuai instruksi nakhoda dan kompetensi; materi tertulis tidak memberi izin mengemudikan kapal atau mengoperasikan semua alat. Kenali identitas kapal, titik berkumpul, komunikasi, serta peralatan darurat sebelum berangkat. Jangan mengubah lokasi atau alat tangkap tanpa persetujuan yang berwenang.',
 'Baca rencana dan pembagian tugas kapal|Periksa alat serta perlengkapan pribadi|Ikuti operasi sesuai kewenangan|Catat hasil dan penyimpangan',
 'Awak baru diminta mengoperasikan alat yang belum pernah dipelajarinya.',
 'Jelaskan batas kemampuan dan minta pelatihan serta arahan|Coba sampai bisa saat kapal berjalan|Lepas pengaman agar mudah|Serahkan pada pengunjung',
 'Apa dasar pembagian tugas awak?~Instruksi nakhoda serta kompetensi dan kewenangan|Keinginan mencoba semua alat|Warna pakaian|Jumlah ikan saja~Operasi kapal dan alat memerlukan tanggung jawab serta pelatihan yang sesuai.',
 '漁業~ぎょぎょう~perikanan tangkap|漁船~ぎょせん~kapal penangkap ikan|船長~せんちょう~nakhoda|乗組員~のりくみいん~awak kapal')
L('Sumber daya, hasil samping, dan aturan tangkap',
 'Aturan penangkapan dapat mengatur jenis, ukuran, musim, area, kuota, serta alat. Jangan menganggap semua hasil boleh dibawa hanya karena masuk jaring. Hasil samping dan spesies dilindungi ditangani menurut peraturan serta instruksi kapal, dengan keselamatan manusia diutamakan. Informasi belum pasti perlu dikonfirmasi, bukan ditebak dari foto lama. Pencatatan hasil dan lokasi mendukung keterlacakan serta pengelolaan sumber daya. Limbah alat tangkap tidak dibuang ke laut untuk menghemat ruang.',
 'Periksa ketentuan operasi dan spesies|Identifikasi hasil serta hasil samping|Ikuti penanganan yang diperintahkan|Catat dan laporkan temuan khusus',
 'Ada organisme yang status perlindungannya belum diketahui dalam hasil tangkapan.',
 'Konfirmasikan identitas dan prosedur kepada penanggung jawab|Sembunyikan di kotak lain|Jual tanpa identifikasi|Buang alat tangkap ke laut',
 'Semua organisme yang tertangkap otomatis boleh dijual?~Tidak perlu mematuhi aturan jenis ukuran dan area|Ya karena sudah tertangkap|Ya bila kecil|Ya bila tak ada pembeli lain~Status hasil tangkap bergantung ketentuan yang berlaku, bukan keberadaannya di alat saja.',
 '資源~しげん~sumber daya|混獲~こんかく~hasil tangkapan samping|禁漁~きんりょう~larangan penangkapan|海域~かいいき~wilayah perairan')
L('Cuaca, gelombang, dan keputusan keberangkatan',
 'Cuaca laut, angin, gelombang, arus, dan pasang dapat mengubah risiko operasi. Kondisi saat berangkat tidak menjamin keadaan tetap aman selama bekerja. Gunakan informasi resmi serta penilaian nakhoda, periksa perubahan, dan laporkan bahaya dari posisi aman. Target hasil tidak boleh dijadikan alasan mengabaikan keputusan penghentian. Awak perlu mengetahui jalur evakuasi, tempat perlengkapan, dan cara menerima instruksi darurat. Pekerja tidak menetapkan batas gelombang universal dari materi karena kapal dan kegiatan berbeda.',
 'Periksa prakiraan serta arahan nakhoda|Pastikan komunikasi dan perlengkapan|Pantau perubahan selama operasi|Patuhi keputusan menunda atau menghentikan',
 'Angin menguat dan nakhoda memerintahkan penghentian operasi.',
 'Ikuti penghentian dan pengamanan sesuai instruksi|Lanjutkan demi target|Abaikan komunikasi|Lepas perlengkapan agar cepat',
 'Siapa mengarahkan keputusan operasi kapal saat kondisi memburuk?~Nakhoda dan pihak berwenang sesuai prosedur|Target penjualan saja|Awak berdasarkan tebakan|Pembeli lewat pesan pribadi~Keputusan keselamatan menggunakan penilaian operasi dan kondisi aktual.',
 '天候~てんこう~cuaca|波~なみ~gelombang|風~かぜ~angin|潮流~ちょうりゅう~arus laut')

c.module('capture-safety','Tangkap · Keselamatan kapal dan keadaan darurat','capture','Pelampung, jatuh ke laut, kebakaran, serta ruang tertutup.')
L('Pelampung dan bergerak di geladak',
 'Pelampung atau alat apung pribadi dipilih, dipakai, dan diperiksa sesuai jenis kerja serta aturan kapal. Alat yang tersedia di lemari tidak melindungi orang yang tidak memakainya saat diwajibkan. Geladak basah, ikan, es, serta gerakan kapal meningkatkan risiko terpeleset dan jatuh. Jaga pijakan, jalur lewat, serta pegangan yang aman. Jangan melewati sisi berbahaya atau duduk di pagar. Perlengkapan tidak boleh mengganggu alat bergerak; kesesuaian dan ukuran diperiksa sebelum operasi.',
 'Periksa jenis ukuran dan kondisi pelampung|Pakai serta kencangkan sesuai petunjuk|Gunakan jalur dan pegangan aman|Laporkan kerusakan atau bahaya geladak',
 'Pelampung memiliki pengikat rusak sebelum tugas di geladak.',
 'Laporkan dan ganti dengan alat yang layak sesuai prosedur|Tetap pakai tanpa mengikat|Simpan saja di kabin|Pinjam ukuran yang tidak dapat terpasang',
 'Kapan kesesuaian pelampung perlu diperiksa?~Sebelum pekerjaan sesuai aturan kapal|Hanya setelah jatuh ke laut|Saat sudah rusak berat|Tidak perlu jika tersedia di lemari~Kondisi, ukuran, dan pemakaian benar menentukan fungsi perlengkapan.',
 '救命胴衣~きゅうめいどうい~pelampung keselamatan|甲板~かんぱん~geladak|手すり~てすり~pegangan|転落~てんらく~jatuh dari tempat tinggi')
L('Orang jatuh ke laut dan komunikasi darurat',
 'Respons orang jatuh ke laut harus mengikuti latihan kapal: beri alarm, sampaikan lokasi, pertahankan pengamatan bila aman, dan jalankan tugas yang ditetapkan. Jangan spontan melompat karena dapat menambah korban dan menghilangkan pengamat. Alat bantu dilempar atau dioperasikan hanya menurut prosedur dan posisi aman. Komunikasi harus singkat serta jelas; ulangi instruksi penting. Kenali titik berkumpul dan cara melaporkan orang hilang. Penyelamatan langsung memerlukan kendali nakhoda serta petugas terlatih.',
 'Beri alarm sesuai prosedur kapal|Laporkan lokasi dan pertahankan pengamatan aman|Gunakan alat bantuan sesuai instruksi|Ikuti koordinasi serta pencatatan awak',
 'Rekan jatuh ke laut dan kamu adalah saksi pertama.',
 'Beri alarm serta laporkan lokasi sambil mengamati dari posisi aman|Langsung melompat sendiri|Lanjutkan menarik jaring|Matikan komunikasi',
 'Mengapa tidak spontan melompat saat melihat korban?~Dapat menambah korban dan mengganggu respons terkoordinasi|Air laut selalu dangkal|Korban tidak perlu bantuan|Pelampung tidak pernah berguna~Penyelamatan perlu prosedur kapal dan petugas dengan perlindungan yang sesuai.',
 '落水~らくすい~jatuh ke air|救助~きゅうじょ~penyelamatan|警報~けいほう~alarm|位置~いち~posisi')
L('Kebakaran, ruang tertutup, dan alat darurat',
 'Bahan bakar, listrik, mesin panas, dan pekerjaan tertentu dapat menimbulkan kebakaran. Jenis alat pemadam dan penggunaannya mengikuti pelatihan serta sumber bahaya; jangan memakai air pada instalasi listrik bertegangan. Ruang tertutup atau tangki dapat memiliki gas berbahaya maupun kekurangan oksigen. Bau bukan bukti udara aman. Jangan masuk atau melakukan penyelamatan spontan tanpa prosedur ruang terbatas. Ketahui cara memberi alarm, menghubungi nakhoda, dan meninggalkan area melalui jalur yang ditetapkan.',
 'Kenali bahaya dan lokasi alat darurat|Beri alarm dari posisi aman|Ikuti isolasi atau evakuasi sesuai tugas|Tunggu petugas terlatih untuk tindakan khusus',
 'Tercium bau aneh dari ruang tertutup dan rekan ingin masuk memeriksa.',
 'Larang masuk tanpa prosedur serta penilaian berwenang dan laporkan|Uji dengan mencium lebih dekat|Masuk sambil menahan napas|Nyalakan api untuk melihat',
 'Ruang tertutup yang tidak berbau pasti aman?~Tidak gas dan oksigen memerlukan penilaian sesuai prosedur|Ya jika pintunya kecil|Ya bila masuk sebentar|Ya jika memakai sepatu bot~Bahaya udara tidak dinilai dari penciuman atau lamanya rencana masuk.',
 '火災~かさい~kebakaran|消火器~しょうかき~alat pemadam|酸素~さんそ~oksigen|避難~ひなん~evakuasi')

c.module('capture-equipment','Tangkap · Tali, alat angkat, dan jaring','capture','Titik jepit, tali bertegangan, winch, serta prinsip jaring insang.')
L('Tali, simpul, dan zona tarikan',
 'Tali dapat bergerak tiba-tiba, menjerat kaki, atau mencambuk ketika bertegangan dan putus. Kenali gulungan, lingkaran tali, arah tarikan, serta zona larangan. Jangan berdiri dalam lingkaran tali atau di jalur tali tegang. Simpul dipilih sesuai tugas dan dilatih dengan alat aman sebelum digunakan dalam operasi. Tali aus, terpotong, atau tidak sesuai kapasitas dilaporkan. Penyimpanan teratur membantu mencegah tersandung serta salah penggunaan; jangan menganggap semua simpul aman untuk semua beban.',
 'Periksa kondisi tali dan jalur tarikan|Tentukan posisi di luar zona bahaya|Gunakan simpul sesuai tugas terlatih|Rapikan serta laporkan kerusakan',
 'Kaki seorang awak berada di dalam lingkaran tali sebelum penarikan.',
 'Beri peringatan dan hentikan operasi sesuai prosedur sampai posisi aman|Mulai tarik perlahan|Ikat kaki pada tali|Anggap aman karena memakai bot',
 'Di mana posisi yang harus dihindari saat tali akan ditarik?~Di dalam lingkaran atau jalur tali tegang|Di zona aman yang ditetapkan|Di tempat pengamatan resmi|Di balik pembatas aman~Gerakan tali dapat menjerat atau menyeret orang secara mendadak.',
 'ロープ~ろーぷ~tali|結び方~むすびかた~cara membuat simpul|張力~ちょうりょく~tegangan tarik|輪~わ~lingkaran')
L('Winch, roller, dan isolasi alat',
 'Winch serta roller memiliki bagian bergerak yang dapat menarik pakaian, sarung tangan, dan tubuh. Pengaman serta tombol berhenti darurat tidak menggantikan posisi kerja yang aman dan pelatihan. Perintah mulai harus dipahami semua awak yang terkait. Jangan menyentuh sumbatan atau memperbaiki alat saat energi belum diisolasi sesuai prosedur. Muatan dan alat angkat harus sesuai kapasitas. Beban tergantung tidak dilintasi atau dijadikan tempat berdiri. Kelainan bunyi, gerakan, atau kebocoran perlu dihentikan dan dilaporkan.',
 'Periksa operator alat dan zona aman|Konfirmasikan sinyal mulai dan berhenti|Operasikan dengan pengaman lengkap|Isolasi energi sebelum penanganan sumbatan',
 'Jaring tersangkut dekat roller yang masih berputar.',
 'Hentikan dan isolasi energi sesuai prosedur sebelum penanganan|Tarik dengan tangan|Lepas pengaman sambil jalan|Berdiri di bawah beban',
 'Sebelum menangani sumbatan roller perlu apa?~Penghentian dan isolasi energi sesuai SOP|Sarung tangan saja|Motor tetap hidup|Rekan memegang tali tegang~Bagian bergerak dan energi tersimpan dapat menyebabkan cedera berat.',
 '巻揚機~まきあげき~winch|回転部~かいてんぶ~bagian berputar|挟まれ~はさまれ~terjepit|停止~ていし~penghentian')
L('Jaring insang dan pemeriksaan alat tangkap',
 'Jaring insang menangkap ikan melalui hubungan tubuh ikan dengan mata jaring, sehingga spesifikasi dan aturan alat penting. Tali atas, pemberat, pelampung, serta jaring diperiksa sebelum digunakan. Kerusakan dapat mengurangi hasil atau menyebabkan alat hilang. Penebaran dan pengangkatan mengikuti komando; jangan berdiri pada jaring atau tali yang akan bergerak. Hasil tangkap dilepas dengan cara yang menjaga keselamatan tangan serta mutu. Alat hilang atau hasil samping dilaporkan sesuai ketentuan, bukan disembunyikan.',
 'Periksa spesifikasi serta kondisi jaring|Siapkan posisi awak dan jalur alat|Tebar atau angkat sesuai komando|Lepaskan hasil dan catat kerusakan',
 'Ditemukan robekan dan sambungan tali lemah sebelum penebaran.',
 'Laporkan dan perbaiki atau ganti sesuai keputusan berwenang|Tebar tanpa laporan|Tutup dengan sampah|Ubah ukuran alat tanpa izin',
 'Mengapa spesifikasi mata jaring penting?~Berkaitan selektivitas dan ketentuan alat tangkap|Menentukan warna ikan|Menggantikan pengaman mesin|Menghilangkan semua hasil samping~Alat harus sesuai rencana penangkapan dan aturan yang berlaku.',
 '刺網~さしあみ~jaring insang|網目~あみめ~mata jaring|浮子~うき~pelampung alat|沈子~おもり~pemberat alat')

c.module('capture-methods','Tangkap · Metode jaring dan pancing','capture','Pukat cincin, trawl, rawai, serta penanganan kail.')
L('Pukat cincin dan koordinasi awak',
 'Pukat cincin mengurung kelompok ikan dan menutup bagian bawah dengan sistem tali sesuai rancangan alat. Operasi melibatkan gerakan kapal, jaring, tali, serta alat bantu yang harus dikoordinasikan. Awak tidak mengubah posisi atau menginjak jaring demi mempercepat proses. Pastikan sinyal dan jalur penarikan diketahui. Beban serta gerakan ikan dapat berubah cepat, sehingga pemeriksaan posisi aman terus dilakukan. Keputusan mengejar kelompok atau menghentikan operasi menjadi kewenangan nakhoda sesuai keadaan.',
 'Dengarkan rencana dan sinyal operasi|Pastikan zona awak aman dari jaring tali|Ikuti urutan penutupan sesuai komando|Periksa alat serta laporkan perubahan',
 'Dua awak menerima perintah mulai yang berbeda saat jaring akan ditarik.',
 'Hentikan ketidakjelasan dan konfirmasikan satu komando yang dipahami|Mulai dengan tebakan masing-masing|Matikan radio|Tarik tali sendiri',
 'Apa yang penting pada operasi pukat cincin?~Komando yang jelas dan koordinasi gerakan alat serta awak|Semua orang bertindak sendiri|Menginjak jaring agar tidak bergerak|Mengabaikan kondisi kapal~Operasi saling terkait; salah sinyal dapat menempatkan orang di jalur bahaya.',
 'まき網~まきあみ~pukat cincin|魚群~ぎょぐん~kelompok ikan|合図~あいず~sinyal|連携~れんけい~koordinasi')
L('Trawl dan kendali muatan alat',
 'Trawl menarik jaring melalui air sesuai rancangan alat dan ketentuan operasi. Kabel, papan pembuka, jaring, serta kantong hasil memiliki risiko beban, ayunan, dan titik jepit. Jangan berdiri pada jalur atau di bawah alat tergantung. Pemeriksaan alat, kapasitas, dan sinyal dilakukan sebelum penanganan. Spesifikasi dasar, daerah, serta metode tangkap tidak diubah tanpa persetujuan. Hasil dan alat diperiksa setelah diangkat, dengan perhatian pada ikan berbahaya atau benda asing yang mungkin masuk.',
 'Periksa rencana spesifikasi dan posisi alat|Amankan zona beban serta komunikasi|Ikuti penarikan sesuai komando|Periksa hasil dan kerusakan alat',
 'Kantong hasil tergantung dan seorang awak hendak melintas di bawahnya.',
 'Cegah melintas dan gunakan jalur aman sesuai komando|Lewat cepat di bawah beban|Berdiri untuk mengarahkan dengan kepala|Lepas pengikat tanpa izin',
 'Posisi di bawah alat atau hasil tergantung boleh digunakan untuk lewat?~Tidak gunakan jalur aman yang ditetapkan|Ya asal cepat|Ya jika hasil sedikit|Ya memakai helm saja~Beban dapat bergerak atau jatuh; APD tidak menghilangkan zona bahaya.',
 '底びき網~そこびきあみ~trawl dasar|曳網~えいもう~penarikan jaring|荷重~かじゅう~beban|吊り荷~つりに~muatan tergantung')
L('Rawai, pancing, dan risiko kail',
 'Rawai menggunakan tali utama, cabang, serta banyak kail; pancing lain memiliki rancangan dan cara operasi berbeda. Kail tajam dan tali bergerak menimbulkan risiko tertusuk serta terseret. Umpan disiapkan dengan alat, posisi, dan higiene yang sesuai. Jangan memegang kail atau ikan berbahaya tanpa metode yang diajarkan. Penebaran serta penarikan mengikuti komando, dengan jalur bebas kaki dan pakaian. Cedera kail dilaporkan dan ditangani sesuai pertolongan yang diajarkan, bukan ditarik paksa tanpa penilaian.',
 'Periksa tali kail dan tugas awak|Siapkan umpan serta posisi aman|Tebar atau tarik mengikuti komando|Amankan kail bekas dan catat cedera',
 'Kail tersangkut pada tangan rekan.',
 'Hentikan bahaya sekitar dan minta bantuan sesuai prosedur cedera|Tarik paksa dengan tali|Biarkan tetap bekerja|Potong kulit sendiri',
 'Risiko utama kail dan tali bergerak adalah apa?~Tusukan dan tarikan yang dapat menyeret|Membuat ikan selalu aman dimakan|Mengganti kebutuhan pelampung|Tidak ada bila tali kecil~Alat tajam dan gerakan tali harus dikendalikan dengan posisi serta prosedur aman.',
 '延縄~はえなわ~rawai|釣針~つりばり~kail|餌~えさ~umpan|幹縄~みきなわ~tali utama rawai')

c.module('capture-products','Tangkap · Mutu hasil dan bahasa kerja','capture','Pendinginan, pencatatan, ikan berbahaya, dan instruksi operasi.')
L('Kesegaran, es, dan kebersihan hasil tangkap',
 'Kesegaran dipengaruhi waktu, suhu, kerusakan fisik, serta pencemaran. Gunakan wadah dan es yang sesuai penggunaan, hindari kontak produk bersih dengan bahan bakar, limbah, atau alat kotor. Pendinginan dilakukan sesuai spesies serta metode fasilitas; satu rasio es tidak berlaku untuk semua kondisi. Jangan menganggap bau normal membuktikan produk aman. Ikan berduri atau berbahaya memerlukan penanganan khusus. Pisahkan hasil rusak atau meragukan untuk penilaian berwenang sebelum pengiriman.',
 'Periksa identitas hasil dan wadah bersih|Tangani lembut sesuai spesies|Dinginkan menurut prosedur produk|Catat waktu serta penyimpangan',
 'Kotak hasil terkena tumpahan bahan bakar.',
 'Pisahkan hasil terdampak dan laporkan untuk keputusan produk|Cuci lalu campur tanpa penilaian|Tutup bau dengan es|Pindahkan label kotak',
 'Pendinginan menggantikan pengendalian pencemaran bahan bakar?~Tidak produk terdampak harus dipisahkan dan dinilai|Ya jika es banyak|Ya bila ikan tampak segar|Ya setelah label diganti~Suhu rendah tidak otomatis menghilangkan kontaminan kimia.',
 '鮮度~せんど~kesegaran|氷~こおり~es|冷却~れいきゃく~pendinginan|汚染~おせん~kontaminasi')
L('Sortasi, hasil bersih, dan keterlacakan kapal',
 'Sortasi memisahkan hasil menurut jenis, mutu, ukuran, dan tujuan produk sesuai ketentuan. Timbang dengan alat yang diperiksa serta bedakan berat wadah dan hasil. Identitas lot menghubungkan kapal, waktu, area, serta perlakuan penanganan yang diperlukan. Jangan mencampur lot meragukan atau mengganti tanggal agar lolos standar. Catat hasil samping dan kerusakan secara faktual. Informasi yang benar memudahkan penelusuran bila ada keluhan atau penahanan produk, dan membantu serah terima ke petugas darat.',
 'Periksa standar sortasi serta lot|Pisahkan jenis dan mutu|Timbang dengan memperhitungkan wadah|Labeli dan serahkan catatan lengkap',
 'Berat bruto hasil beserta kotak 25 kg dan kotak kosong 5 kg.',
 'Catat netto 20 kg setelah memeriksa data|Catat 30 kg netto|Abaikan kotak|Ubah lot agar berat naik',
 'Bruto 25 kg dengan wadah 5 kg menghasilkan netto berapa?~20 kg|30 kg|125 kg|5 kg~Berat netto = 25 dikurangi 5 = 20 kg pada contoh penghitungan.',
 '選別~せんべつ~sortasi|水揚げ~みずあげ~pendaratan hasil|重量~じゅうりょう~berat|記録~きろく~catatan')
L('Instruksi Jepang dan laporan operasi tangkap',
 'Komunikasi di kapal harus singkat dan dipahami, terutama perintah berhenti, bahaya, arah, serta lokasi awak. 止まって berarti berhenti dan 危険 berarti bahaya. Jika instruksi tidak terdengar, minta pengulangan; jangan mengira semua teriakan meminta mempercepat. Laporan berisi kondisi alat, lokasi, waktu, dan tindakan yang telah diambil. Bahasa Jepang dipelajari bersama demonstrasi aman serta latihan kapal, bukan menggantikan pelatihan navigasi atau radio. Dalam ujian resmi perikanan, bahasa Jepang menjadi bagian penting pemahaman soal.',
 'Dengarkan perintah dan identifikasi tugas|Ulangi bagian yang belum jelas|Laksanakan hanya setelah dipahami|Laporkan hasil serta kondisi berbahaya',
 'Terdengar 「止まってください」 ketika alat sedang dioperasikan.',
 'Lakukan penghentian sesuai prosedur dan konfirmasikan bahaya|Tambah kecepatan|Abaikan karena tidak menyebut nama|Matikan semua alat tanpa koordinasi',
 '「止まってください」 dalam instruksi kerja berarti apa?~Tolong berhenti|Tambahkan kecepatan|Mulai makan|Buang hasil~止まって meminta berhenti; pelaksanaannya tetap mengikuti prosedur penghentian aman alat.',
 '指示~しじ~instruksi|停止~ていし~penghentian|危険~きけん~bahaya|報告~ほうこく~laporan')

c.module('aqua-general','Budi daya · Sistem, lokasi, dan kualitas air','aquaculture','Budi daya berpakan dan tanpa pakan, lokasi, parameter air, serta oksigen.')
L('Sistem budi daya dan siklus produksi',
 'Budi daya memelihara organisme melalui pengaturan benih, lingkungan, pemeliharaan, panen, serta catatan produksi. Sistem berpakan seperti banyak ikan berbeda dari sistem tanpa pemberian pakan seperti rumput laut dan sejumlah kerang. Tanpa pakan bukan berarti tanpa pemantauan atau risiko. Jenis organisme dan kondisi lokasi menentukan metode; satu prosedur tidak dapat diterapkan ke semua spesies. Pekerja harus memahami unit pemeliharaan, identitas kelompok, alat, serta batas kewenangan sebelum melakukan perubahan.',
 'Identifikasi spesies sistem dan kelompok|Baca rencana siklus produksi|Kerjakan pemeliharaan sesuai jalur|Catat perubahan dan hasil',
 'Rekan menganggap kerang tidak diberi pakan sehingga tak perlu pemeriksaan.',
 'Tetap periksa lingkungan kondisi organisme dan fasilitas|Hentikan semua pemantauan|Tambahkan pakan ikan sembarangan|Hilangkan identitas kelompok',
 'Sistem tanpa pemberian pakan berarti apa?~Tidak diberi pakan tambahan tetapi tetap perlu pemeliharaan|Tidak memerlukan air|Tidak memiliki risiko penyakit|Tidak perlu pencatatan~Kerang dan rumput laut tetap membutuhkan kondisi lingkungan serta pengawasan sesuai sistem.',
 '養殖~ようしょく~budi daya|給餌~きゅうじ~pemberian pakan|無給餌~むきゅうじ~tanpa pemberian pakan|生産~せいさん~produksi')
L('Lokasi, sarana, arus, dan keamanan akses',
 'Lokasi kolam, tangki, atau keramba memengaruhi pertukaran air, akses kerja, paparan pencemar, dan risiko cuaca. Sarana harus diperiksa menurut rancangan, termasuk jalur orang, pagar, pelampung, pengikatan, serta alat darurat. Jangan masuk air atau menyelam untuk memperbaiki alat tanpa kewenangan dan prosedur khusus. Kerusakan yang tampak kecil dapat membesar karena arus atau beban. Catat unit terdampak dan laporkan sebelum meneruskan kegiatan yang membuat pekerja atau organisme terpapar bahaya.',
 'Periksa unit akses dan cuaca|Amati sarana serta jalur air|Laporkan kerusakan sebelum operasi berisiko|Ikuti perbaikan oleh petugas berwenang',
 'Pengikat keramba rusak dan arus menguat.',
 'Batasi kegiatan berisiko dan laporkan untuk penanganan terkoordinasi|Masuk air sendiri|Ikat tubuh ke keramba|Abaikan sampai putus',
 'Pekerja boleh menyelam spontan untuk memperbaiki sarana?~Tidak diperlukan pelatihan kewenangan dan prosedur khusus|Ya jika pandai berenang|Ya bila sebentar|Ya tanpa alat darurat~Kemampuan berenang tidak menggantikan pengendalian risiko kerja di air atau penyelaman.',
 '施設~しせつ~fasilitas|生簀~いけす~keramba|潮流~ちょうりゅう~arus laut|係留~けいりゅう~penambatan')
L('Parameter air dan oksigen terlarut',
 'Suhu, oksigen terlarut, pH, salinitas, serta parameter limbah dipantau sesuai spesies dan sistem. Satu nilai tidak menjelaskan seluruh kondisi. Alat ukur harus diperiksa atau dikalibrasi sesuai manual, lokasi dan waktu sampel dicatat agar hasil dapat dibandingkan. Ikan berkumpul di permukaan atau perubahan konsumsi dapat mengisyaratkan gangguan, tetapi diagnosis memerlukan data. Target oksigen dan tindakan koreksi mengikuti rencana fasilitas, bukan angka universal; jangan menambah bahan kimia berdasarkan tebakan.',
 'Periksa alat serta titik sampel|Ukur dan catat waktu parameter|Bandingkan dengan target spesies|Laporkan penyimpangan dan ikuti respons',
 'Ikan berkumpul di permukaan dan pembacaan oksigen turun dari pola biasa.',
 'Laporkan segera dan jalankan rencana respons kualitas air|Tambahkan pakan sebanyak mungkin|Matikan semua aerasi|Tambahkan obat sendiri',
 'Mengapa waktu dan titik pengukuran air dicatat?~Agar perubahan dapat dibandingkan dengan benar|Mengganti kalibrasi alat|Menentukan warna ikan|Menghilangkan semua penyakit~Kondisi air dapat berbeda menurut lokasi dan waktu, sehingga konteks pengukuran penting.',
 '溶存酸素~ようぞんさんそ~oksigen terlarut|水質~すいしつ~kualitas air|塩分~えんぶん~salinitas|測定~そくてい~pengukuran')

c.module('aqua-stock','Budi daya · Benih, kepadatan, dan pertumbuhan','aquaculture','Suhu, aklimatisasi, penerimaan benih, pencatatan jumlah, dan grading.')
L('Suhu, perubahan mendadak, dan aklimatisasi',
 'Perubahan suhu atau salinitas yang mendadak dapat menyebabkan stres. Penerimaan organisme dari lokasi lain memerlukan prosedur aklimatisasi menurut spesies, ukuran, serta sistem pengangkutan. Jangan menuangkan semua air angkut ke unit budidaya tanpa prosedur biosekuriti. Periksa kondisi wadah, waktu kedatangan, data lingkungan, dan status kelompok. Bila ada kematian atau gangguan, hentikan pemindahan yang tidak aman dan minta arahan. Proses adaptasi tidak memiliki satu durasi universal untuk semua organisme.',
 'Periksa kondisi organisme dan data asal|Bandingkan lingkungan tujuan|Lakukan aklimatisasi sesuai SOP|Catat pemindahan dan kelainan',
 'Air wadah angkut dan unit tujuan berbeda jauh suhunya.',
 'Tahan pelepasan langsung dan ikuti prosedur aklimatisasi|Tuang semua seketika|Tambahkan air panas sesuka hati|Abaikan perbedaan',
 'Mengapa pelepasan benih langsung pada lingkungan berbeda berisiko?~Perubahan mendadak dapat menyebabkan stres|Semua organisme kebal perubahan|Air angkut selalu steril|Suhu tidak berpengaruh~Aklimatisasi disesuaikan spesies dan kondisi angkut serta tujuan.',
 '水温~すいおん~suhu air|馴致~じゅんち~aklimatisasi|輸送~ゆそう~pengangkutan|変化~へんか~perubahan')
L('Penerimaan benih dan identitas kelompok',
 'Benih yang diterima diperiksa asal, spesies, ukuran, jumlah, kondisi, serta dokumen yang dipersyaratkan. Perbedaan label dan isi perlu dikonfirmasi sebelum dicampurkan. Biosekuriti dapat mensyaratkan pemisahan, penilaian kesehatan, atau prosedur lain sesuai fasilitas. Jangan menganggap semua organisme dari pemasok sama mempunyai status seragam. Catatan lot memudahkan pengamatan pertumbuhan, mortalitas, dan perlakuan. Penanganan dilakukan lembut dengan alat yang sesuai agar kerusakan fisik dan waktu di luar kondisi aman diminimalkan.',
 'Cocokkan dokumen lot dan spesies|Periksa jumlah ukuran serta kondisi|Tempatkan sesuai prosedur penerimaan|Catat asal dan unit tujuan',
 'Label benih tidak cocok dengan dokumen pengiriman.',
 'Pisahkan penerimaan dan konfirmasikan identitas sebelum dicampur|Campur dahulu agar cepat|Ubah label sendiri|Buang dokumen',
 'Apa fungsi catatan lot benih?~Menghubungkan asal dan kondisi dengan kelompok yang benar|Mengganti semua pengamatan|Menentukan cuaca|Menghapus kebutuhan biosekuriti~Keterlacakan dimulai saat penerimaan dan pemindahan organisme.',
 '種苗~しゅびょう~benih organisme|受入れ~うけいれ~penerimaan|由来~ゆらい~asal|群~むれ~kelompok')
L('Kepadatan, grading, dan penghitungan jumlah',
 'Kepadatan memengaruhi akses pakan, kualitas air, pertumbuhan, dan risiko stres. Target bergantung spesies, ukuran, sarana, dan kapasitas pengelolaan air. Grading memisahkan ukuran bila diperlukan dengan alat dan waktu penanganan yang sesuai. Jumlah hidup dihitung dari data awal, penambahan, pemindahan, serta kematian yang dicatat; angka perkiraan harus dibedakan dari hitungan pasti. Jangan menaikkan kepadatan hanya karena ruang tampak masih luas atau menghapus mortalitas agar catatan terlihat baik.',
 'Baca jumlah serta target kepadatan|Periksa ukuran dan kebutuhan grading|Tangani pemisahan sesuai SOP|Perbarui jumlah dan unit kelompok',
 'Contoh unit berisi 1000 ekor, 20 mati dan 100 dipindahkan tanpa penambahan.',
 'Catat sisa menurut contoh 880 ekor|Catat 1120 ekor|Hapus data kematian|Tetap catat 1000 ekor',
 '1000 ekor dikurangi 20 mati dan 100 pindah menyisakan berapa?~880 ekor|1080 ekor|920 ekor|1120 ekor~Jumlah contoh = 1000 minus 20 minus 100 = 880, dengan asumsi tidak ada perubahan lain.',
 '密度~みつど~kepadatan|選別~せんべつ~grading atau pemilahan|尾数~びすう~jumlah ekor ikan|成長~せいちょう~pertumbuhan')

c.module('aqua-feed','Budi daya · Pakan dan kesehatan','aquaculture','Respons makan, biomassa, perhitungan ransum contoh, penyakit, dan biosekuriti.')
L('Pemberian pakan dan respons makan',
 'Pakan dipilih menurut spesies, ukuran, fase, dan rencana produksi. Kondisi air, cuaca, kesehatan, serta respons makan memengaruhi pelaksanaan harian. Pakan berlebih bukan cara pasti mempercepat pertumbuhan; sisa dapat memperburuk lingkungan. Simpan pakan berlabel dan terlindung dari air serta hama. Jangan memberikan produk berjamur atau obat tanpa kewenangan. Catat jumlah aktual dan respons organisme, bukan hanya jatah rencana. Penurunan respons perlu dilaporkan bersama data lingkungan.',
 'Periksa rencana pakan dan kondisi air|Pastikan mutu pakan serta kelompok|Berikan sesuai SOP sambil mengamati respons|Catat aktual sisa dan perubahan',
 'Respons makan turun dan banyak pakan tersisa.',
 'Laporkan dan ikuti penyesuaian yang disetujui berdasarkan kondisi|Tambahkan dua kali lipat|Tutup sisa dengan pakan baru|Abaikan catatan kualitas air',
 'Apa akibat potensial pemberian pakan berlebih?~Sisa pakan dapat memperburuk kualitas lingkungan|Selalu menghilangkan penyakit|Mengganti kebutuhan oksigen|Menjamin pertumbuhan tanpa batas~Jumlah aktual perlu mempertimbangkan respons makan dan kondisi sistem.',
 '飼料~しりょう~pakan|摂餌~せつじ~aktivitas makan|残餌~ざんじ~sisa pakan|保管~ほかん~penyimpanan')
L('Biomassa dan latihan menghitung pakan',
 'Biomassa adalah perkiraan total berat organisme dalam unit. Pengukuran contoh harus mengikuti metode yang ditetapkan agar tidak memilih hanya individu terbesar. Contoh 100 ekor dengan berat rata-rata 0,2 kilogram menghasilkan biomassa 20 kilogram. Bila rencana latihan menggunakan 2 persen biomassa per hari, hasil hitung 0,4 kilogram pakan. Angka ini contoh matematika, bukan rekomendasi untuk semua spesies atau kondisi. Konfirmasikan satuan, jumlah hidup, metode sampling, dan rencana aktual sebelum pemberian.',
 'Periksa jumlah hidup dan data berat sampel|Hitung biomassa dengan satuan tepat|Gunakan persentase rencana yang dikonfirmasi|Periksa ulang dan catat pemberian aktual',
 'Biomassa contoh 20 kg dan rencana latihan 2 persen per hari.',
 'Hitung 0,4 kg pakan dan konfirmasikan rencana aktual|Hitung 40 kg|Anggap 2 kg selalu benar|Ganti kilogram ke ekor',
 '2 persen dari biomassa 20 kg adalah berapa?~0,4 kg|4 kg|40 kg|0,04 kg~20 dikali 0,02 = 0,4 kg. Persentase nyata mengikuti rencana spesies dan kondisi.',
 '魚体重~ぎょたいじゅう~berat tubuh ikan|平均~へいきん~rata-rata|給餌量~きゅうじりょう~jumlah pakan|割合~わりあい~persentase atau proporsi')
L('Gejala penyakit dan mortalitas',
 'Perubahan berenang, nafsu makan, warna, luka, dan kematian dipantau bersama kualitas air serta riwayat kelompok. Gejala tidak cukup untuk menentukan penyakit atau obat sendiri. Mortalitas dicatat menurut unit dan waktu; organisme mati ditangani melalui prosedur fasilitas agar tidak menyebarkan masalah. Pisahkan alat atau kelompok sesuai arahan. Penggunaan obat memerlukan kewenangan serta catatan dan batas penggunaan produk yang berlaku. Jangan mengirim hasil yang statusnya belum dipastikan hanya karena tampak normal dari luar.',
 'Amati kelompok dan catat gejala|Periksa data air serta riwayat|Laporkan untuk penilaian berwenang|Ikuti pengendalian dan penanganan kematian',
 'Kematian meningkat tajam dan beberapa ikan berenang tidak normal.',
 'Laporkan segera dengan data unit waktu gejala dan kualitas air|Berikan obat sisa tanpa arahan|Buang bangkai ke unit lain|Hapus angka kematian',
 'Apa yang dilakukan sebelum memilih obat pada gejala baru?~Penilaian petugas berwenang berdasarkan data|Tebak dari satu foto|Pakai obat kelompok lain|Campur semua bahan~Penyebab gejala bisa beragam; pengobatan dan keamanan produk membutuhkan keputusan berwenang.',
 '疾病~しっぺい~penyakit|死亡~しぼう~kematian|症状~しょうじょう~gejala|診断~しんだん~diagnosis')

c.module('aqua-systems','Budi daya · Sarana, kerang, dan rumput laut','aquaculture','Jaring keramba, fouling, pemeliharaan kerang, serta budidaya tanpa pakan.')
L('Keramba, jaring, dan pengendalian organisme penempel',
 'Jaring keramba menjaga organisme tetap di unit serta memungkinkan pertukaran air. Robekan menimbulkan risiko lepas, sedangkan organisme penempel dapat mengurangi pertukaran dan menambah beban. Pemeriksaan mencakup jaring, sambungan, pelampung, tali, serta kondisi sekitar dengan metode aman. Pembersihan atau penggantian mengikuti rencana agar organisme tidak terpapar stres atau pencemar. Jangan menyelam atau menggunakan bahan antifouling tanpa pelatihan dan izin. Catat kerusakan serta tindakan untuk unit yang tepat.',
 'Periksa jaring pengikat dan akses aman|Catat robekan serta penempel|Koordinasikan perawatan sesuai SOP|Periksa unit dan jumlah setelah tindakan',
 'Robekan jaring terdeteksi saat cuaca memburuk.',
 'Laporkan dan ikuti rencana pengamanan tanpa masuk air spontan|Terjun memperbaiki sendiri|Abaikan sampai panen|Tambahkan bahan kimia acak',
 'Mengapa organisme penempel pada jaring dipantau?~Dapat mengurangi pertukaran air dan menambah beban|Selalu menjadi pakan cukup|Menggantikan jaring|Menjamin ikan tidak sakit~Perawatan jaring mendukung aliran air dan integritas sarana.',
 '網~あみ~jaring|付着生物~ふちゃくせいぶつ~organisme penempel|破損~はそん~kerusakan|交換~こうかん~penggantian')
L('Kerang, pakan alami, dan area panen',
 'Banyak kerang memperoleh makanan dengan menyaring partikel dari air, sehingga kualitas lokasi sangat penting. Sistem tanpa pakan tambahan tetap memerlukan benih, sarana, pengamatan pertumbuhan, kepadatan, dan risiko lingkungan. Penutupan area panen atau peringatan biotoksin harus dipatuhi; mencuci kerang tidak otomatis menghilangkan seluruh bahaya. Jangan menilai keamanan hanya dari cangkang yang tampak normal. Pemindahan, pemanenan, serta pengiriman mengikuti status area dan rencana fasilitas dengan identitas lot yang terjaga.',
 'Periksa status area serta lot kerang|Amati sarana dan kondisi pertumbuhan|Panen hanya sesuai izin serta rencana|Catat asal waktu dan penanganan',
 'Ada pemberitahuan resmi penutupan panen pada area kerang.',
 'Hentikan panen sesuai ketentuan dan konfirmasikan status|Panen lalu cuci agar aman|Campur dari area lain tanpa label|Ganti tanggal asal',
 'Cangkang normal membuktikan kerang aman saat area ditutup?~Tidak status area dan ketentuan panen harus dipatuhi|Ya selalu|Ya setelah dibilas|Ya jika pembeli setuju~Bahaya lingkungan atau biotoksin tidak dinilai hanya dari penampilan.',
 '貝類~かいるい~kerang|濾過~ろか~penyaringan|採苗~さいびょう~pengumpulan benih|海域~かいいき~area perairan')
L('Rumput laut, cahaya, dan pemeliharaan tali',
 'Rumput laut tumbuh dengan kondisi cahaya, suhu, salinitas, arus, dan unsur lingkungan yang sesuai spesies. Budidaya tanpa pakan tetap memerlukan penyiapan benih, pemasangan, pengaturan kepadatan, serta pemeriksaan tali dan organisme penempel. Penempatan terlalu padat dapat memengaruhi cahaya dan aliran, tetapi koreksi mengikuti rencana, bukan satu angka universal. Perubahan warna atau jaringan rusak dicatat dan dilaporkan. Panen serta pengeringan atau penyimpanan disesuaikan produk, dengan pencegahan pencemaran dari alat dan tempat kerja.',
 'Periksa spesies fase dan sarana tali|Amati kondisi lingkungan serta kepadatan|Rawat sesuai rencana komoditas|Catat pertumbuhan serta perubahan jaringan',
 'Rumput laut berubah warna dan penempel meningkat pada unit padat.',
 'Catat dan laporkan kondisi untuk penilaian lingkungan serta pemeliharaan|Tambahkan pakan ikan|Semprot obat tanpa izin|Abaikan karena tanpa pakan',
 'Apa kebutuhan penting rumput laut selain sarana?~Cahaya dan kondisi air sesuai spesies|Pakan ikan untuk semua jenis|Tidak memerlukan pemantauan|Air tanpa pertukaran selamanya~Fotosintesis dan pertumbuhan bergantung lingkungan; kebutuhan spesifik mengikuti komoditas.',
 '海藻~かいそう~rumput laut|光~ひかり~cahaya|養成~ようせい~pemeliharaan pembesaran|収穫~しゅうかく~panen')

c.module('aqua-products','Budi daya · Panen, laporan, dan integrasi','aquaculture','Kesiapan panen, higiene produk, keterlacakan, dan bahasa Jepang.')
L('Persiapan panen dan status perlakuan',
 'Kesiapan panen memerlukan kondisi organisme, target mutu, status perlakuan, kebutuhan pasar, serta sarana penanganan. Catatan obat dan masa henti yang berlaku harus dikonfirmasi sebelum produk dikirim; penampilan sehat tidak cukup. Penghentian pakan sebelum panen, bila diperlukan, mengikuti rencana spesies dan fasilitas. Jangan menganggap satu lama penghentian berlaku untuk semua. Siapkan alat serta wadah bersih dan posisi pekerja aman, lalu tangani organisme dengan metode terlatih untuk mengurangi cedera serta penurunan mutu.',
 'Konfirmasikan lot mutu dan status perlakuan|Siapkan rencana alat serta wadah panen|Tangani sesuai metode yang diizinkan|Catat jumlah dan penanganan hasil',
 'Catatan perlakuan kelompok belum lengkap menjelang panen.',
 'Tunda keputusan pelepasan produk sampai status dikonfirmasi|Kirim bila ikan tampak sehat|Hapus catatan lama|Campur ke lot tanpa perlakuan',
 'Penampilan sehat menggantikan pemeriksaan status perlakuan sebelum panen?~Tidak catatan dan ketentuan produk tetap perlu dipastikan|Ya selalu|Ya jika jumlah sedikit|Ya setelah diganti wadah~Keputusan produk melibatkan riwayat perlakuan dan persyaratan yang tidak terlihat secara visual.',
 '出荷~しゅっか~pengiriman|休薬期間~きゅうやくきかん~masa henti obat|品質~ひんしつ~mutu|準備~じゅんび~persiapan')
L('Higiene hasil dan rantai penyimpanan',
 'Hasil budidaya ditangani dengan wadah, air, es, serta area yang sesuai produk. Jalur bersih dipisahkan dari limbah, bahan kimia, dan alat tercemar. Pendinginan atau proses lain mengikuti kebutuhan komoditas; bahan hidup, ikan segar, kerang, dan rumput laut dapat memerlukan penanganan berbeda. Jangan menyamakan suhu atau metode untuk semuanya. Catatan waktu dan kondisi mendukung penilaian jika terjadi penyimpangan. Lot yang meragukan dipisahkan agar keputusan tidak hanya berdasarkan bau atau warna.',
 'Cocokkan produk dengan metode penanganan|Periksa kebersihan alat dan area|Pantau kondisi penyimpanan sesuai spesifikasi|Labeli lot serta laporkan penyimpangan',
 'Wadah bersih tercampur dengan alat bekas menangani kematian.',
 'Pisahkan dan proses alat sesuai higiene sebelum kontak produk|Gunakan asal tidak berbau|Tutupi dengan es|Campur ke semua lot',
 'Mengapa jalur hasil bersih dan alat tercemar dipisahkan?~Mencegah perpindahan pencemar ke produk|Menambah berat hasil|Menggantikan label|Membuat semua suhu sama~Pemisahan dan pengolahan alat membantu mengendalikan kontaminasi silang.',
 '衛生~えいせい~higiene|容器~ようき~wadah|保冷~ほれい~penjagaan suhu dingin|汚染~おせん~kontaminasi')
L('Laporan budi daya dan instruksi Jepang',
 'Laporan unit memuat spesies atau kelompok, waktu, parameter air, pakan aktual, mortalitas, pemindahan, serta tindakan yang telah disetujui. 給餌量 adalah jumlah pakan dan 水温 adalah suhu air. Ketidakjelasan satuan dapat mengubah keputusan; minta pengulangan dan konfirmasikan angka sebelum tindakan. Bedakan data ukur dari perkiraan. Ujian resmi budi daya menggunakan bahasa Jepang, sedangkan materi ini memberi jembatan Indonesia dengan kosakata dan kasus. Pemahaman tertulis tidak menggantikan pelatihan pengukuran, alat, atau keselamatan kerja di air.',
 'Baca unit waktu parameter dan satuan|Konfirmasikan instruksi yang belum jelas|Catat hasil aktual secara objektif|Serahkan laporan dan tugas tindak lanjut',
 'Instruksi berbunyi 「水温を確認してください」.',
 'Periksa suhu air sesuai metode yang ditetapkan|Tambah pakan langsung|Ganti label kelompok|Ukur berat wadah saja',
 '「水温を確認してください」 meminta apa?~Memeriksa suhu air|Menambah pakan|Menghitung harga kapal|Membuka semua jaring~水温 berarti suhu air dan 確認 berarti memeriksa atau mengonfirmasi.',
 '水温~すいおん~suhu air|給餌量~きゅうじりょう~jumlah pakan|報告~ほうこく~laporan|単位~たんい~satuan')

if __name__=='__main__': c.write()
