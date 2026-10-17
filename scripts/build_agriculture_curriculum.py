"""Original SSW 1 agriculture course: crop production and livestock."""
from ssw_course_authoring import Course

c=Course('agriculture','agri',[
 dict(title='NCA · Cakupan pertanian tanaman dan peternakan',url='https://asat-nca.jp/asat1/overview/cover'),
 dict(title='NCA · Pedoman ASAT 1, 70 soal sekitar 60 menit',url='https://www.nca.or.jp/upload/asat_guidelines.pdf'),
 dict(title='MAFF · Keselamatan kerja pertanian',url='https://www.maff.go.jp/j/seisan/sien/sizai/s_kikaika/anzen/'),
])
L=c.lesson
c.module('crop-foundations','Tanaman · Dasar, pertumbuhan, dan tanah','crop','Lingkup pekerjaan, kebutuhan tanaman, pengukuran tanah dan drainase.')
L('Rencana produksi dan identitas petak',
 'Produksi tanaman dimulai dari jenis komoditas, varietas, musim, petak, dan tujuan pasar. Kalender budidaya menghubungkan pembibitan, penanaman, pemeliharaan, panen, serta pengiriman. Petak yang tampak sama dapat memiliki perlakuan dan masa tunggu berbeda. Identitas petak mencegah salah panen atau salah aplikasi bahan. Periksa versi rencana yang berlaku dan jangan mengganti varietas hanya karena benih tersedia. Catatan kegiatan harus dapat ditelusuri ke petak serta tanggal sebenarnya.',
 'Baca rencana dan kode petak|Cocokkan varietas dan jadwal|Kerjakan tugas yang disetujui|Catat petak dan tanggal',
 'Dua petak sayur memiliki jadwal panen berbeda tetapi label salah satunya hilang.',
 'Konfirmasikan identitas petak sebelum bekerja|Panen keduanya sekaligus|Pilih berdasarkan ukuran saja|Ganti label tanpa konfirmasi',
 'Apa fungsi identitas petak?~Menghubungkan kegiatan dan hasil dengan lokasi yang benar|Mengganti pemeriksaan tanah|Menentukan cuaca pasti|Menghapus kebutuhan catatan~Petak dan tanggal menjadi dasar keterlacakan serta pemilihan perlakuan.',
 '圃場~ほじょう~petak lahan|品種~ひんしゅ~varietas|計画~けいかく~rencana|収穫~しゅうかく~panen')
L('Pertumbuhan, akar, dan fotosintesis',
 'Akar menyerap air serta unsur hara dan memerlukan kondisi tanah yang mendukung pertukaran udara. Daun menggunakan cahaya untuk fotosintesis, tetapi kebutuhan cahaya, suhu, dan air berbeda menurut komoditas serta fase pertumbuhan. Layu tidak selalu berarti kekurangan air; akar rusak atau tanah terlalu basah juga dapat mengganggu penyerapan. Amati tanaman bersama lingkungannya sebelum mengubah irigasi atau pupuk. Bedakan gejala yang terlihat dari dugaan penyebabnya.',
 'Amati daun dan kondisi akar yang dapat diperiksa|Periksa kelembapan dan lingkungan|Bandingkan dengan standar komoditas|Laporkan perubahan sebelum koreksi',
 'Tanaman layu padahal tanah sangat basah.',
 'Periksa kondisi lingkungan dan laporkan sebelum menambah air|Siram terus tanpa pemeriksaan|Tambahkan pupuk pekat|Cabut semua tanaman segera',
 'Mengapa tanaman layu tidak selalu perlu ditambah air?~Gangguan akar atau tanah terlalu basah juga dapat menyebabkan layu|Semua tanaman tidak memerlukan air|Layu selalu karena benih|Daun tidak berhubungan dengan akar~Gejala perlu diperiksa bersama tanah dan kondisi akar agar tindakan tidak memperburuk masalah.',
 '根~ね~akar|葉~は~daun|光~ひかり~cahaya|成長~せいちょう~pertumbuhan')
L('Struktur tanah, pH, dan pengambilan sampel',
 'Tanah mengandung partikel mineral, bahan organik, air, serta udara. Tekstur dan struktur memengaruhi retensi air, drainase, dan kemudahan akar berkembang. pH menggambarkan keasaman, bukan jumlah semua unsur hara. Kebutuhan pH berbeda menurut tanaman. Sampel dari satu titik dekat tumpukan pupuk tidak mewakili seluruh petak. Gunakan pola pengambilan dan kedalaman yang ditentukan, alat bersih, label jelas, serta metode ukur yang sesuai. Koreksi tanah mengikuti hasil penilaian dan rencana.',
 'Tentukan area dan metode sampel|Ambil sampel representatif dengan alat bersih|Labeli lokasi dan kedalaman|Gunakan hasil untuk arahan perbaikan',
 'Sampel hanya diambil tepat di samping tumpukan pupuk.',
 'Ulangi pengambilan sesuai pola representatif|Anggap mewakili semua petak|Tambahkan kapur tanpa ukur|Campur dengan tanah tak berlabel',
 'Apa yang ditunjukkan pH tanah?~Tingkat keasaman tanah|Berat hasil panen|Jumlah seluruh hama|Tanggal tanam~pH mengukur keasaman; kebutuhan koreksi bergantung komoditas dan hasil pemeriksaan.',
 '土壌~どじょう~tanah budidaya|酸度~さんど~keasaman|試料~しりょう~sampel|測定~そくてい~pengukuran')
L('Pengolahan tanah, bedengan, dan drainase',
 'Pengolahan tanah menyiapkan zona akar dan kondisi penanaman. Tanah terlalu basah dapat memadat bila dilalui mesin, sedangkan bedengan dan saluran harus disesuaikan lahan serta komoditas. Drainase mengalirkan kelebihan air melalui jalur yang diizinkan; jangan membuang air tercemar ke saluran umum. Periksa batu, utilitas, dan orang di sekitar sebelum alat digunakan. Mesin hanya dioperasikan petugas terlatih dengan pengaman utuh. Penyumbatan dibersihkan setelah sumber energi diisolasi sesuai prosedur.',
 'Periksa kondisi tanah dan jalur air|Tentukan bedengan sesuai rencana|Gunakan alat secara aman|Periksa aliran dan catat genangan',
 'Tanah masih jenuh air tetapi alat berat dijadwalkan masuk.',
 'Konfirmasikan penundaan atau metode aman sebelum masuk|Paksa agar jadwal tepat|Tutup saluran drainase|Naikkan kecepatan mesin',
 'Mengapa alat berat perlu dihindari pada tanah terlalu basah?~Risiko pemadatan dan kerusakan struktur|Membuat semua hama hilang|Menambah udara pasti|Mempercepat kematangan buah~Pemadatan mengganggu ruang udara dan perkembangan akar.',
 '排水~はいすい~drainase|畝~うね~bedengan|耕うん~こううん~pengolahan tanah|水路~すいろ~saluran air')

c.module('crop-nutrition','Tanaman · Pupuk, air, dan fasilitas','crop','Kebutuhan hara, perhitungan sederhana, irigasi, dan rumah tanam.')
L('Unsur hara dan pemupukan berimbang',
 'Nitrogen, fosfor, dan kalium merupakan unsur utama pada banyak label pupuk, tetapi kebutuhan tanaman tidak ditentukan oleh tiga unsur saja. Penambahan pupuk mengikuti analisis tanah, komoditas, fase pertumbuhan, dan riwayat aplikasi. Daun menguning tidak cukup untuk memastikan kekurangan nitrogen. Pemupukan berlebih dapat merusak akar, meningkatkan kehilangan unsur ke lingkungan, dan menurunkan mutu hasil. Simpan pupuk berlabel, kering sesuai petunjuk, serta terpisah dari produk yang tidak kompatibel.',
 'Baca rekomendasi dan label pupuk|Pastikan dosis serta petak|Aplikasikan sesuai rencana|Catat jumlah dan respons tanaman',
 'Daun menguning dan seorang rekan ingin menggandakan dosis pupuk.',
 'Periksa gejala dan minta penilaian sebelum mengubah dosis|Gandakan semua pupuk|Campur bahan tanpa label|Sebar pupuk ke saluran air',
 'Dasar penentuan dosis pupuk adalah apa?~Rekomendasi budidaya dan kondisi petak|Warna kemasan saja|Semakin banyak semakin baik|Ukuran sepatu operator~Dosis bergantung analisis, tanaman, fase, dan riwayat; gejala tunggal belum memastikan penyebab.',
 '肥料~ひりょう~pupuk|窒素~ちっそ~nitrogen|施肥~せひ~pemupukan|記録~きろく~catatan')
L('Menghitung kebutuhan pupuk dan luas',
 'Satuan luas harus dipastikan sebelum menghitung kebutuhan. Satu a sama dengan 100 meter persegi, sehingga 10 a sama dengan 1000 meter persegi. Contoh rencana 2 kilogram per 100 meter persegi untuk petak 300 meter persegi membutuhkan 6 kilogram. Ini hanya latihan hitung, bukan rekomendasi dosis nyata. Bedakan berat produk dengan berat unsur hara pada label. Periksa hasil hitung bersama penanggung jawab sebelum penimbangan dan aplikasi di lapangan.',
 'Pastikan satuan luas dan dosis rencana|Hitung jumlah berdasarkan luas|Periksa ulang hasil dan timbangan|Timbang lalu catat aplikasi',
 'Rencana contoh 2 kg per 100 m² dipakai untuk petak 300 m².',
 'Hitung 6 kg dan konfirmasikan sebelum aplikasi|Gunakan 600 kg|Gunakan 2 kg untuk semua luas|Abaikan satuan luas',
 'Pada contoh 2 kg/100 m², kebutuhan petak 300 m² berapa?~6 kg|60 kg|0,6 kg|300 kg~300 dibagi 100 = 3; tiga kali 2 kg = 6 kg. Dosis nyata tetap mengikuti rencana lapangan.',
 '面積~めんせき~luas|重量~じゅうりょう~berat|計量~けいりょう~penimbangan|単位~たんい~satuan')
L('Irigasi dan pemeriksaan pasokan air',
 'Pemberian air disesuaikan kebutuhan tanaman, media, cuaca, dan fase pertumbuhan. Jadwal tetap tanpa pemeriksaan dapat menyebabkan kekurangan atau kelebihan air. Sistem tetes memerlukan pemeriksaan sumbatan serta distribusi, bukan hanya pompa yang menyala. Pastikan sumber air sesuai penggunaan, terutama bila bersentuhan dengan hasil konsumsi. Perbaikan listrik dan pompa bukan tugas coba-coba. Kebocoran dan genangan dilaporkan, sementara perubahan setelan hanya dilakukan sesuai kewenangan.',
 'Periksa kondisi tanaman dan kelembapan|Cocokkan jadwal serta sumber air|Periksa distribusi dan kebocoran|Catat volume atau durasi sesuai prosedur',
 'Ujung barisan kering walau pompa menyala.',
 'Periksa distribusi dan laporkan sumbatan atau kebocoran|Anggap semua petak mendapat air|Naikkan pupuk otomatis|Buka sambungan listrik basah',
 'Pompa menyala membuktikan seluruh tanaman mendapat air?~Tidak, distribusi dan keluaran perlu diperiksa|Ya selalu|Ya bila bunyinya keras|Ya tanpa melihat ujung barisan~Sumbatan dan kebocoran dapat menyebabkan distribusi tidak merata.',
 '灌水~かんすい~irigasi|水源~すいげん~sumber air|漏水~ろうすい~kebocoran air|湿度~しつど~kelembapan')
L('Rumah tanam, ventilasi, dan panas',
 'Rumah tanam membantu mengendalikan lingkungan, tetapi suhu dan kelembapan dapat berubah cepat. Sensor perlu ditempatkan serta diperiksa menurut prosedur agar tidak memberi pembacaan menyesatkan. Ventilasi, naungan, dan penyiraman tidak boleh diubah tanpa melihat kebutuhan komoditas dan keamanan pekerja. Panas memengaruhi tanaman sekaligus tubuh pekerja. Kenali gejala gangguan panas, jadwal istirahat, akses minum, serta cara meminta bantuan. Jangan menutup rumah tanam dengan pekerja masih terpapar bahan atau panas.',
 'Periksa kondisi dan pembacaan sensor|Konfirmasikan target komoditas|Atur ventilasi sesuai instruksi|Periksa pekerja dan catat perubahan',
 'Suhu rumah tanam meningkat dan rekan merasa pusing.',
 'Hentikan kerja berisiko dan aktifkan bantuan sesuai prosedur panas|Minta rekan bertahan sampai panen|Kunci semua ventilasi|Abaikan karena tanaman sehat',
 'Mengapa kondisi rumah tanam perlu diperiksa berulang?~Suhu dan kelembapan dapat berubah cepat|Sensor selalu benar selamanya|Cuaca tidak berpengaruh|Ventilasi tidak memengaruhi lingkungan~Lingkungan tertutup dapat berubah cepat dan menimbulkan risiko tanaman serta pekerja.',
 '温室~おんしつ~rumah tanam|換気~かんき~ventilasi|温度~おんど~suhu|休憩~きゅうけい~istirahat')

c.module('crop-establishment','Tanaman · Benih dan pembentukan tanaman','crop','Benih, persemaian, pindah tanam, pemangkasan, serta penyerbukan.')
L('Benih, kedalaman semai, dan daya tumbuh',
 'Periksa nama varietas, lot, tanggal, perlakuan benih, dan petunjuk penyimpanan. Benih berperlakuan tidak dipakai sebagai makanan atau pakan. Kedalaman, media, kelembapan, serta suhu perkecambahan bergantung tanaman; tidak semua benih harus ditanam sedalam mungkin. Label tray menjaga varietas tetap terpisah. Daya tumbuh dihitung dari jumlah berkecambah dibanding jumlah yang diuji pada kondisi uji, bukan jaminan setiap benih di lapang akan tumbuh.',
 'Cocokkan varietas dan lot|Siapkan media dan kedalaman sesuai tanaman|Semai lalu labeli tray|Amati dan catat perkecambahan',
 'Dari 100 benih uji, 80 berkecambah pada kondisi uji.',
 'Catat daya tumbuh uji 80 persen|Catat 20 persen|Jamin semua benih lapang tumbuh|Buang label lot',
 '80 benih berkecambah dari 100 berarti berapa persen?~80%|8%|20%|125%~Persentase perkecambahan = jumlah berkecambah dibagi jumlah uji dikali 100.',
 '種子~しゅし~benih|発芽~はつが~perkecambahan|播種~はしゅ~penyemaian|深さ~ふかさ~kedalaman')
L('Persemaian dan penilaian bibit',
 'Bibit dinilai dari pertumbuhan seimbang, kondisi daun, batang, dan akar sesuai komoditas. Bibit tinggi kurus dapat menandakan lingkungan kurang sesuai, bukan otomatis bibit terbaik. Tray bersih, media yang diizinkan, pemisahan lot, dan pemeriksaan rutin membantu pencegahan penyakit. Bibit bergejala dipisahkan sesuai arahan agar tidak menyebarkan masalah. Hindari memindahkan media, air, atau alat kotor antar-kelompok tanpa pembersihan yang ditetapkan.',
 'Periksa tray dan identitas bibit|Amati daun batang dan akar|Pisahkan gejala sesuai arahan|Catat pertumbuhan dan tindakan',
 'Sebagian bibit rebah dan pangkal batang berubah warna.',
 'Pisahkan sesuai SOP dan laporkan gejala serta lot|Campurkan ke semua tray|Tutup gejala dengan pupuk|Pindahkan tanpa pencatatan',
 'Bibit tinggi kurus selalu paling baik?~Tidak, perlu menilai keseimbangan dan kesehatan|Ya tinggi satu-satunya kriteria|Ya walau pangkal busuk|Ya tanpa melihat akar~Mutu bibit meliputi kekuatan dan kesehatan, bukan tinggi saja.',
 '苗~なえ~bibit|育苗~いくびょう~pembibitan|茎~くき~batang|病気~びょうき~penyakit')
L('Pindah tanam dan jarak tanam',
 'Pindah tanam menghubungkan bibit dengan lokasi produksi. Pastikan bibit layak, petak sesuai, jarak dan kedalaman menurut komoditas, serta kondisi air dan cuaca memungkinkan. Mengangkat bibit dengan menarik batang dapat merusak akar atau jaringan. Penanganan gumpalan akar mengikuti jenis bibit dan wadah. Jangan memadatkan atau menanam terlalu dalam secara sembarangan. Periksa tanaman setelah penanaman untuk mendeteksi stres dan ketidaksesuaian posisi.',
 'Periksa bibit petak dan jarak rencana|Lepaskan bibit tanpa merusak akar|Tanam sesuai kedalaman dan kebutuhan air|Periksa posisi serta catat jumlah',
 'Bibit tertahan di tray dan rekan menarik batangnya kuat-kuat.',
 'Gunakan cara pelepasan wadah yang sesuai agar akar terlindungi|Tarik lebih keras|Potong semua akar|Tanam bersama tray plastik tanpa arahan',
 'Mengapa jarak tanam mengikuti rencana komoditas?~Mengatur ruang tumbuh dan pemeliharaan|Semua tanaman butuh jarak sama|Menghilangkan semua penyakit|Mengganti kebutuhan air~Kepadatan memengaruhi cahaya, sirkulasi, pertumbuhan, dan akses kerja.',
 '定植~ていしょく~pindah tanam|株間~かぶま~jarak antartanaman|植穴~うえあな~lubang tanam|根鉢~ねばち~gumpalan akar')
L('Pemangkasan, penopang, dan penyerbukan',
 'Pemangkasan, pemilihan tunas, pengikatan, dan penyerbukan berbeda menurut tanaman serta tujuan produksi. Memotong semua tunas atau bunga dapat mengurangi hasil. Alat bersih dan tajam sesuai tugas mengurangi kerusakan; luka dan getah dapat memindahkan masalah jika alat dipakai lintas tanaman tanpa prosedur. Ikatan penopang tidak boleh mencekik batang. Penyerbukan membutuhkan fase bunga dan kondisi yang tepat; periksa instruksi sebelum mengambil tindakan.',
 'Identifikasi tanaman dan bagian sasaran|Siapkan alat serta penopang sesuai prosedur|Lakukan pemotongan atau ikatan yang ditentukan|Periksa luka dan tekanan ikatan',
 'Ikatan penopang mulai menekan batang yang membesar.',
 'Laporkan dan sesuaikan ikatan sesuai prosedur tanaman|Kencangkan agar diam|Potong batang utama|Abaikan sampai putus',
 'Ikatan penopang yang baik harus bagaimana?~Menyangga tanpa mencekik pertumbuhan batang|Sekencang mungkin|Selalu dari kawat tajam|Menutup semua daun~Ikatan perlu diperiksa karena diameter batang bertambah selama pertumbuhan.',
 '整枝~せいし~pengaturan cabang|受粉~じゅふん~penyerbukan|支柱~しちゅう~tiang penopang|誘引~ゆういん~pengarahan ikatan')

c.module('crop-protection','Tanaman · Perlindungan dan keselamatan','crop','Pemantauan hama, pestisida, mesin, serta sistem budidaya padi.')
L('Hama, penyakit, gulma, dan pengamatan terpadu',
 'Pengendalian dimulai dari pengamatan jenis organisme, kerusakan, luas kejadian, cuaca, dan riwayat petak. Tidak semua serangga adalah hama; sebagian bermanfaat. Gulma bersaing dengan tanaman tetapi metode pengendalian harus cocok dengan sistem budidaya. Pencegahan, sanitasi, pengelolaan lingkungan, dan metode nonkimia dapat menjadi bagian pengendalian terpadu. Identifikasi yang belum pasti perlu dikonfirmasi. Menyemprot semua petak tanpa menilai risiko dapat merugikan hasil dan lingkungan.',
 'Amati organisme dan pola kerusakan|Catat lokasi serta tingkat kejadian|Konfirmasikan identifikasi dan rencana|Pantau hasil pengendalian',
 'Ada serangga yang belum dikenali pada satu tanaman.',
 'Catat ciri dan kerusakan lalu konfirmasikan identifikasi|Semprot seluruh kebun segera|Anggap semuanya musuh|Pindahkan ke petak lain',
 'Apa langkah awal pengendalian terpadu?~Pengamatan dan identifikasi masalah|Mencampur semua pestisida|Membakar setiap tanaman|Menaikkan dosis tanpa diagnosis~Pengendalian perlu berdasarkan organisme dan tingkat kerusakan yang benar.',
 '害虫~がいちゅう~hama serangga|雑草~ざっそう~gulma|観察~かんさつ~pengamatan|防除~ぼうじょ~pengendalian')
L('Label pestisida dan interval sebelum panen',
 'Produk pestisida digunakan hanya untuk penggunaan yang diizinkan dengan tanaman, sasaran, dosis, frekuensi, masa tunggu sebelum panen, dan persyaratan keselamatan pada label. APD, cuaca, arah angin, akses orang, dan pencegahan hanyutan perlu diperiksa. Penggunaan berlebih tidak berarti lebih efektif. Catat produk, lot, petak, jumlah, tanggal, serta operator. Jangan menyalin dosis dari tanaman lain atau mencampur produk tanpa izin. Masa tunggu harus dikonfirmasi sebelum hasil dipanen.',
 'Baca label dan izin penggunaan|Konfirmasikan dosis APD dan cuaca|Aplikasikan hanya oleh petugas berwenang|Catat dan patuhi interval panen',
 'Petak akan dipanen tetapi catatan aplikasi dan masa tunggu belum jelas.',
 'Tunda keputusan panen dan konfirmasikan catatan serta label|Panen jika tampak bersih|Cuci untuk menghapus semua masa tunggu|Ganti tanggal catatan',
 'Masa tunggu sebelum panen ditentukan dari apa?~Label produk dan catatan aplikasi pada tanaman terkait|Bau tanaman saja|Warna buah|Perkiraan rekan tanpa data~Keamanan panen membutuhkan kesesuaian penggunaan dan interval yang dapat ditelusuri.',
 '農薬~のうやく~pestisida|使用基準~しようきじゅん~standar penggunaan|飛散~ひさん~hanyutan semprotan|保護具~ほごぐ~alat pelindung')
L('Alat pertanian, titik jepit, dan isolasi energi',
 'Mesin pertanian memiliki bagian berputar, bilah, titik jepit, dan energi tersimpan. Pengaman tidak dilepas demi mempercepat tugas. Pemeriksaan dilakukan sesuai manual dan pelatihan; operator tidak boleh membawa penumpang bila alat tidak dirancang untuk itu. Sebelum membersihkan sumbatan atau mengganti aksesori, hentikan, isolasi energi, dan pastikan kondisi aman sesuai prosedur. Bahan bakar, lereng, jalan umum, serta jarak orang sekitar juga perlu dinilai sebelum operasi.',
 'Periksa izin operator dan kondisi mesin|Amankan area serta pengaman|Operasikan sesuai manual|Isolasi energi sebelum penanganan sumbatan',
 'Tanaman tersangkut pada bilah yang masih bergerak.',
 'Jauhkan orang dan hentikan serta isolasi energi sesuai SOP sebelum penanganan|Tarik dengan tangan|Lepas pengaman saat mesin hidup|Tambah kecepatan bilah',
 'Apa yang diperlukan sebelum membersihkan sumbatan mesin?~Penghentian dan isolasi energi sesuai prosedur|Sarung tipis saja|Rekan memegang bilah|Mesin tetap berputar~Bagian bergerak dan energi tersimpan dapat mencederai walau sumbatan tampak kecil.',
 '機械~きかい~mesin|回転~かいてん~putaran|停止~ていし~penghentian|安全装置~あんぜんそうち~perangkat pengaman')
L('Budidaya padi dan pengelolaan air sawah',
 'Padi memiliki tahapan dari persiapan benih dan bibit, penanaman, pemeliharaan, pembentukan malai, hingga panen. Kebutuhan penggenangan atau pengeringan mengikuti fase, tanah, dan rencana setempat; kedalaman air tidak berlaku sama sepanjang musim. Periksa pematang, saluran, gulma, serta distribusi air tanpa merusak lahan tetangga. Catatan pertumbuhan dan anakan membantu pengamatan. Mesin tanam dan panen memerlukan pelatihan khusus, terutama dekat saluran atau tanah lunak.',
 'Identifikasi fase dan rencana air|Periksa pematang serta saluran|Atur air sesuai arahan petak|Catat kondisi tanaman dan kebocoran',
 'Pematang bocor dan air mengalir ke petak lain.',
 'Laporkan dan tangani sesuai koordinasi pengairan|Buka semua saluran tanpa izin|Tutup saluran tetangga diam-diam|Anggap tidak perlu dicatat',
 'Kebutuhan air sawah mengikuti apa?~Fase tanaman dan rencana kondisi petak|Satu kedalaman tetap untuk semua fase|Warna traktor|Jadwal libur operator~Tahap pertumbuhan dan kondisi lahan menentukan pengelolaan air yang sesuai.',
 '稲~いね~padi|水田~すいでん~sawah|田植え~たうえ~penanaman padi|畦~あぜ~pematang')

c.module('crop-harvest','Tanaman · Buah, panen, dan penanganan hasil','crop','Pohon buah, kematangan, sortasi, penyimpanan, dan pengiriman.')
L('Pohon buah dan penjarangan',
 'Tanaman buah memerlukan pengamatan tunas, bunga, buah, serta keseimbangan pertumbuhan menurut spesies dan musim. Penjarangan mengatur beban buah, tetapi waktu dan jumlah yang dipertahankan mengikuti rencana komoditas. Jangan memilih buah hanya dari ukuran tanpa memperhatikan posisi dan kondisi. Alat tajam harus dikendalikan, dan tangga atau pekerjaan tinggi memerlukan akses aman yang disetujui. Catat cabang atau kelompok yang telah dikerjakan agar tidak dijarangkan berulang.',
 'Periksa spesies dan target penjarangan|Pilih bagian sesuai instruksi|Kerjakan dengan alat serta akses aman|Tandai dan catat bagian selesai',
 'Buah di cabang tinggi tidak terjangkau dari posisi kerja aman.',
 'Hentikan dan minta metode akses yang disetujui|Naik di ember|Lompat sambil membawa gunting|Tarik cabang sampai patah',
 'Penjarangan buah bertujuan apa?~Mengatur beban buah sesuai rencana komoditas|Menghilangkan seluruh hasil|Mengganti semua pemupukan|Membuat semua jenis buah sama~Beban buah dan waktu penjarangan ditentukan untuk komoditas serta tujuan mutu.',
 '果樹~かじゅ~pohon buah|摘果~てきか~penjarangan buah|枝~えだ~cabang|開花~かいか~pembungaan')
L('Kematangan panen dan higiene pekerja',
 'Kriteria panen mencakup tahap pertumbuhan, warna, ukuran, kekerasan, atau parameter lain menurut komoditas dan pasar. Warna yang sama tidak selalu berarti kematangan sama. Periksa status masa tunggu bahan dan identitas petak sebelum memanen. Tangan, wadah, serta alat harus sesuai standar higiene agar hasil tidak tercemar. Hasil yang jatuh atau terkena pencemar dipisahkan menurut prosedur, bukan disatukan agar jumlah lebih banyak. Penanganan lembut mengurangi memar dan luka.',
 'Konfirmasikan petak dan izin panen|Nilai kematangan sesuai komoditas|Panen dengan alat serta wadah bersih|Pisahkan hasil rusak dan catat',
 'Sebagian sayuran jatuh ke area tercemar saat panen.',
 'Pisahkan dan ikuti prosedur penilaian hasil tercemar|Campurkan ke produk bersih|Cukup beri label baru|Sembunyikan di dasar kotak',
 'Kriteria panen perlu disesuaikan dengan apa?~Komoditas dan tujuan mutu pasar|Warna seragam pekerja|Jumlah kotak kosong saja|Keinginan selesai cepat~Kematangan dan persyaratan penggunaan bahan harus sesuai produk yang dipanen.',
 '適期~てっき~waktu yang sesuai|成熟~せいじゅく~kematangan|衛生~えいせい~higiene|容器~ようき~wadah')
L('Sortasi, penimbangan, dan label lot',
 'Sortasi memisahkan hasil menurut standar mutu, ukuran, kerusakan, dan tujuan pemasaran. Timbangan perlu diperiksa dan bobot wadah diperhitungkan sesuai prosedur; berat bruto berbeda dari netto. Jangan mengubah lot agar produk tidak sesuai tampak memenuhi syarat. Kemasan melindungi hasil tanpa tekanan berlebih atau ventilasi yang salah. Label menghubungkan produk dengan petak, tanggal, serta proses. Hasil meragukan dipisahkan sampai ada keputusan yang berwenang.',
 'Baca standar sortasi dan lot|Periksa timbangan serta bobot wadah|Pisahkan dan kemas sesuai mutu|Cocokkan label dan catat berat',
 'Bruto kotak 12 kg dan bobot wadah 2 kg pada contoh latihan.',
 'Catat netto 10 kg setelah memeriksa penimbangan|Catat netto 14 kg|Abaikan bobot wadah|Ganti label lot lain',
 'Bruto 12 kg dan wadah 2 kg menghasilkan netto berapa?~10 kg|14 kg|24 kg|6 kg~Netto adalah bruto dikurangi bobot wadah, yaitu 12 minus 2 = 10 kg.',
 '選別~せんべつ~sortasi|正味~しょうみ~berat netto|包装~ほうそう~pengemasan|出荷~しゅっか~pengiriman')
L('Penyimpanan, suhu, dan keterlacakan hasil',
 'Penyimpanan memperlambat penurunan mutu, tetapi suhu dan kelembapan sesuai tidak sama untuk semua hasil. Beberapa komoditas sensitif terhadap suhu terlalu rendah atau gas dari produk lain. Jangan menebak satu suhu universal. Hindari kondensasi, tekanan tumpukan, hama gudang, serta pencampuran bahan kimia dengan pangan. Catat penerimaan, kondisi ruang, lot, dan pengeluaran. Bila rantai kondisi terputus, pisahkan dan laporkan sebelum pengiriman agar keputusan berdasarkan data.',
 'Cocokkan kebutuhan produk dan ruang simpan|Periksa kondisi serta lot masuk|Pantau suhu dan kondisi tumpukan|Catat pengeluaran serta penyimpangan',
 'Catatan suhu gudang hilang dan ada hasil membusuk dalam satu lot.',
 'Pisahkan lot terdampak dan laporkan untuk penilaian|Kirim bersama hasil baik|Ubah tanggal masuk|Tambahkan pewangi',
 'Mengapa suhu simpan tidak boleh disamaratakan?~Kebutuhan dan sensitivitas berbeda menurut komoditas|Semua hasil identik|Suhu tidak memengaruhi mutu|Label lot menggantikan pendinginan~Kondisi simpan harus sesuai spesifikasi produk dan dapat dipantau.',
 '貯蔵~ちょぞう~penyimpanan|低温~ていおん~suhu rendah|品質~ひんしつ~mutu|追跡~ついせき~penelusuran')

c.module('livestock-foundations','Peternakan · Dasar dan biosekuriti','livestock','Perilaku, identitas hewan, kebersihan masuk, pakan, serta air minum.')
L('Jenis ternak, perilaku, dan penanganan aman',
 'Sapi, babi, dan unggas memiliki perilaku, kebutuhan, serta risiko penanganan berbeda. Pekerja perlu memahami jalur bergerak, respons hewan terhadap suara atau gerakan, dan posisi aman. Jangan berdiri di titik jepit antara hewan dan pagar, memukul agar cepat bergerak, atau mendekati tanpa jalur keluar. Identitas individu atau kelompok menghubungkan pakan, kesehatan, dan produksi. Pemindahan menggunakan metode yang disetujui dan jumlah pekerja sesuai risiko, bukan improvisasi sendiri.',
 'Identifikasi hewan dan perilakunya|Siapkan jalur dan posisi aman|Pindahkan dengan metode yang disetujui|Periksa jumlah serta identitas',
 'Sapi gelisah berada di antara pekerja dan pagar sempit.',
 'Jaga jarak aman dan minta bantuan untuk mengatur jalur|Masuk ke celah sempit|Pukul agar bergerak cepat|Berdiri di belakang tanpa jalur keluar',
 'Mengapa identitas ternak diperlukan?~Menghubungkan perawatan dan catatan dengan hewan yang benar|Menggantikan pakan|Mencegah semua penyakit otomatis|Menentukan cuaca~Kesalahan identitas dapat menyebabkan salah penanganan, pencatatan, atau produk.',
 '家畜~かちく~ternak|個体~こたい~individu hewan|行動~こうどう~perilaku|移動~いどう~pemindahan')
L('Zona bersih, alas kaki, dan biosekuriti',
 'Biosekuriti mencegah masuk dan menyebarnya agen penyakit melalui hewan, orang, alat, kendaraan, pakan, atau satwa liar. Batas zona, pakaian kerja, alas kaki, pembersihan, serta pencatatan pengunjung mengikuti fasilitas. Disinfektan tidak efektif hanya karena baunya kuat; permukaan kotor dan kondisi penggunaan memengaruhi hasil. Jangan melompati prosedur masuk karena kunjungan singkat. Peralatan tidak dipindahkan antar-kelompok tanpa pengolahan yang ditetapkan, terutama setelah kontak dengan hewan sakit.',
 'Periksa aturan masuk dan zona|Ganti perlengkapan sesuai batas zona|Bersihkan serta disinfeksi sesuai SOP|Catat kunjungan dan perpindahan',
 'Rekan hendak memakai sepatu kandang sakit ke kandang sehat.',
 'Hentikan perpindahan dan ikuti prosedur zona bersih|Semprot pewangi lalu masuk|Anggap aman bila sebentar|Tutup jejak dengan jerami',
 'Apa tujuan pemisahan zona kandang?~Mengurangi perpindahan pencemar dan penyakit|Mengurangi kebutuhan air hewan|Mengubah jenis ternak|Menghapus catatan pengunjung~Perpindahan orang dan alat merupakan jalur yang perlu dikendalikan.',
 '防疫~ぼうえき~biosekuriti|畜舎~ちくしゃ~kandang ternak|消毒~しょうどく~disinfeksi|長靴~ながぐつ~sepatu bot')
L('Pakan, penyimpanan, dan perubahan ransum',
 'Pakan memenuhi kebutuhan menurut jenis hewan, umur, fase produksi, dan kondisi kesehatan. Bahan yang berjamur, basah, atau salah identitas tidak dicampurkan untuk menghabiskan stok. Pakan baru dan perubahan ransum memerlukan rencana agar konsumsi serta pencernaan tetap stabil. Timbang atau ukur sesuai prosedur, periksa tanggal dan lot, lalu simpan terhindar dari hama dan pencemar. Produk kimia dan obat tidak disimpan bersama bahan pakan tanpa pengendalian yang tepat.',
 'Cocokkan ransum dengan kelompok|Periksa mutu lot dan jumlah|Berikan sesuai jadwal yang disetujui|Catat konsumsi serta sisa',
 'Karung pakan lembap dan tampak berjamur.',
 'Pisahkan dan laporkan agar tidak diberikan sebelum penilaian|Campur sedikit ke semua pakan|Beri lebih banyak|Ubah label tanggal',
 'Ransum ternak ditentukan berdasarkan apa?~Jenis umur fase produksi dan rencana pemeliharaan|Jumlah karung sisa saja|Warna kandang|Satu formula untuk semua hewan~Kebutuhan nutrisi berbeda; perubahan harus mengikuti rencana yang disetujui.',
 '飼料~しりょう~pakan|給与~きゅうよ~pemberian pakan|残飼~ざんし~sisa pakan|配合~はいごう~komposisi ransum')
L('Air minum dan pemeriksaan tempat minum',
 'Air minum yang sesuai mutu dan tersedia cukup mendukung kesehatan serta produksi. Kebutuhan berubah menurut suhu, jenis ternak, konsumsi pakan, dan fase produksi. Tangki penuh tidak membuktikan semua titik minum berfungsi. Periksa aliran, kebersihan, posisi, dan akses kelompok, lalu catat kebocoran atau sumbatan. Jangan menambah obat atau bahan kimia ke air tanpa instruksi berwenang. Penurunan konsumsi bisa menjadi tanda gangguan dan perlu dilaporkan bersama data lingkungan.',
 'Periksa sumber dan akses air|Uji fungsi titik minum sesuai SOP|Bersihkan atau laporkan sumbatan|Catat gangguan dan respons kelompok',
 'Tangki penuh tetapi salah satu tempat minum tidak mengalir.',
 'Laporkan dan pastikan akses air aman sesuai prosedur|Anggap hewan sudah mendapat air|Kurangi semua air|Tambahkan obat sendiri',
 'Mengapa titik minum perlu diperiksa walau tangki penuh?~Distribusi dapat tersumbat atau tidak terjangkau|Tangki menentukan umur ternak|Semua katup pasti berfungsi|Air tak memengaruhi kesehatan~Ketersediaan di sumber berbeda dari akses aktual setiap kelompok.',
 '飲水~いんすい~air minum|給水~きゅうすい~pasokan air|水槽~すいそう~tangki air|流量~りゅうりょう~laju aliran')

c.module('livestock-environment','Peternakan · Lingkungan dan pengamatan','livestock','Alas kandang, ventilasi, perubahan kesehatan, reproduksi, dan pelaporan.')
L('Alas kandang, kebersihan, dan kenyamanan',
 'Alas kandang membantu kenyamanan, kebersihan, dan pengendalian kelembapan sesuai sistem pemeliharaan. Alas basah atau kotor dapat meningkatkan paparan pencemar dan gangguan kaki atau kulit. Pergantian dilakukan dengan memperhatikan posisi hewan, ventilasi, debu, serta alat yang dipakai. Jangan menyiram kandang berisi hewan tanpa rencana keselamatan. Pisahkan alur alat kotor dari area pakan dan produk. Periksa lantai rusak, benda tajam, dan titik licin saat area aman.',
 'Amankan hewan dan area kerja|Periksa kebasahan serta kondisi lantai|Ganti alas sesuai sistem kandang|Catat kerusakan dan kebersihan akhir',
 'Alas sangat basah di sekitar kebocoran tempat minum.',
 'Laporkan kebocoran dan ganti alas sesuai prosedur|Tutup lapisan basah tanpa penilaian|Biarkan sampai bau hilang|Campurkan alas ke pakan',
 'Mengapa alas basah perlu ditangani?~Mengganggu kenyamanan dan meningkatkan paparan pencemar|Selalu menambah nutrisi|Menggantikan air minum|Mencegah semua hama~Kelembapan alas memengaruhi kondisi lingkungan dan kesehatan hewan.',
 '敷料~しきりょう~alas kandang|清掃~せいそう~pembersihan|床~ゆか~lantai|乾燥~かんそう~pengeringan')
L('Ventilasi kandang dan stres panas',
 'Ventilasi membantu mengeluarkan panas, kelembapan, dan gas yang tidak diinginkan, tetapi aliran udara tidak boleh ditentukan tanpa melihat jenis serta umur hewan. Kondensasi, napas terengah, perubahan perilaku, dan konsumsi dapat menjadi tanda lingkungan perlu diperiksa. Sensor serta kipas perlu pemeliharaan aman. Jangan masuk ruang tertutup berpotensi gas berbahaya atau mengaduk limbah tanpa prosedur khusus. Pekerja juga berisiko gangguan panas dan harus mengikuti pengaturan istirahat serta bantuan darurat.',
 'Amati hewan dan kondisi lingkungan|Periksa data suhu serta ventilasi|Minta koreksi sesuai rencana fasilitas|Pantau perubahan dan catat',
 'Unggas terengah dan suhu kandang meningkat.',
 'Laporkan segera dan jalankan rencana penanganan lingkungan|Tutup semua ventilasi|Tambah kepadatan kandang|Abaikan karena kipas berbunyi',
 'Kipas berbunyi menjamin ventilasi memadai?~Tidak perlu memeriksa aliran dan kondisi hewan|Ya selalu|Ya walau saluran tertutup|Ya tanpa pemeriksaan suhu~Fungsi alat, distribusi udara, dan respons hewan perlu diperiksa bersama.',
 '換気~かんき~ventilasi|暑熱~しょねつ~panas lingkungan|呼吸~こきゅう~pernapasan|湿気~しっけ~kelembapan udara')
L('Pengamatan kesehatan dan isolasi sesuai arahan',
 'Pengamatan rutin membandingkan nafsu makan, aktivitas, napas, kotoran, posisi tubuh, dan produksi dengan kondisi biasa. Gejala tidak cukup untuk diagnosis sendiri. Catat hewan atau kelompok, waktu, perubahan, serta lingkungan, kemudian laporkan ke penanggung jawab. Pemisahan hewan sakit dilakukan sesuai rencana dan biosekuriti; jangan membawa alat tercemar ke kelompok sehat. Obat serta keputusan pengobatan memerlukan petugas berwenang dan pencatatan, termasuk batas penggunaan produk yang relevan.',
 'Identifikasi individu atau kelompok|Amati dan catat perubahan objektif|Laporkan untuk penilaian berwenang|Ikuti isolasi dan pemantauan yang diarahkan',
 'Satu hewan menolak pakan dan sulit berdiri.',
 'Laporkan identitas gejala serta waktu dan ikuti arahan|Berikan obat sisa sendiri|Paksa berjalan tanpa penilaian|Campur ke kelompok sehat',
 'Tugas pekerja saat melihat gejala baru adalah apa?~Mencatat dan melapor untuk penilaian|Menentukan obat sendiri|Menghapus identitas hewan|Menunda sampai semua sakit~Pengamatan objektif mendukung keputusan penanggung jawab dan tenaga kesehatan hewan.',
 '健康~けんこう~kesehatan|食欲~しょくよく~nafsu makan|異常~いじょう~kelainan|隔離~かくり~isolasi')
L('Reproduksi, pencatatan, dan batas kewenangan',
 'Reproduksi melibatkan siklus, perkawinan atau inseminasi, kebuntingan, dan kelahiran yang berbeda menurut jenis hewan. Pekerja mengamati tanda sesuai pelatihan dan mencatat tanggal serta identitas, bukan memastikan kebuntingan hanya dari penampilan. Tindakan reproduksi, penggunaan hormon, dan pemeriksaan invasif merupakan kewenangan petugas yang ditunjuk. Kesalahan identitas atau tanggal dapat mengganggu keputusan pemeliharaan. Sampaikan perubahan perilaku dan kondisi tanpa menambahkan diagnosis yang belum dikonfirmasi.',
 'Periksa identitas dan catatan reproduksi|Amati tanda yang diajarkan|Laporkan waktu serta perubahan|Ikuti instruksi petugas berwenang',
 'Hewan menunjukkan perubahan perilaku yang diduga tanda berahi.',
 'Catat identitas waktu dan tanda lalu laporkan|Berikan hormon sendiri|Pastikan bunting tanpa pemeriksaan|Ubah catatan hewan lain',
 'Mengapa tanggal dan identitas reproduksi harus tepat?~Mendukung keputusan pemeliharaan hewan yang benar|Menggantikan semua pemeriksaan|Menentukan harga pakan|Menghapus kebutuhan pengamatan~Keputusan reproduksi harus berdasarkan hewan dan waktu yang dapat ditelusuri.',
 '繁殖~はんしょく~reproduksi|発情~はつじょう~berahi|妊娠~にんしん~kebuntingan|分娩~ぶんべん~kelahiran')

c.module('livestock-cattle','Peternakan · Sapi dan produksi susu','livestock','Kelahiran, pedet, pemerahan, higiene susu, dan sapi potong.')
L('Persiapan kelahiran dan pengawasan induk',
 'Persiapan kelahiran meliputi area bersih, alat sesuai protokol, catatan induk, serta jalur menghubungi petugas berwenang. Amati tanda sesuai pelatihan dan jangan melakukan penarikan paksa atau tindakan invasif tanpa kewenangan. Kesulitan atau perubahan mendadak perlu segera dilaporkan sesuai ambang fasilitas, bukan menunggu angka waktu dari ingatan. Setelah kelahiran, pengamatan induk serta anak mengikuti rencana. Jaga ketenangan, kebersihan, dan keselamatan dari gerakan hewan yang dapat berubah cepat.',
 'Periksa identitas induk dan rencana kelahiran|Siapkan area serta kontak bantuan|Amati sesuai pelatihan tanpa tindakan invasif|Laporkan kelainan dan catat kejadian',
 'Induk tampak kesulitan melahirkan dan pekerja ingin menarik anak kuat-kuat.',
 'Segera hubungi petugas berwenang dan ikuti protokol|Tarik sekuat tenaga|Gunakan alat apa saja|Tunda laporan sampai shift selesai',
 'Siapa menentukan tindakan saat kelahiran bermasalah?~Petugas berwenang sesuai protokol|Siapa pun tanpa pelatihan|Pengunjung kandang|Pekerja berdasarkan tebakan~Intervensi kelahiran membutuhkan kompetensi dan kewenangan, dengan pelaporan cepat.',
 '母畜~ぼちく~induk ternak|出産~しゅっさん~persalinan|準備~じゅんび~persiapan|連絡~れんらく~kontak pelaporan')
L('Pedet, kolostrum, dan kebersihan peralatan',
 'Pedet memerlukan pengawasan awal, lingkungan sesuai, dan pemberian kolostrum mengikuti rencana fasilitas serta arahan tenaga berwenang. Mutu, waktu, jumlah, dan cara pemberian tidak ditebak dari materi umum karena kondisi individu penting. Botol dan alat harus bersih sesuai prosedur agar tidak menjadi jalur penyakit. Amati kemampuan minum, aktivitas, kotoran, serta kondisi tubuh. Pedet lemah atau tidak mampu minum memerlukan laporan segera; jangan memaksa memasukkan cairan tanpa pelatihan.',
 'Periksa identitas dan kondisi pedet|Konfirmasikan rencana kolostrum serta alat bersih|Berikan hanya dengan metode terlatih|Catat asupan dan laporkan kesulitan',
 'Pedet sangat lemah dan tidak mampu minum seperti biasa.',
 'Laporkan segera dan ikuti arahan penanganan|Paksa cairan masuk dengan alat improvisasi|Biarkan tanpa catatan|Campur obat sendiri ke botol',
 'Pemberian kolostrum perlu mengikuti apa?~Rencana fasilitas dan kondisi individu|Satu volume tebakan untuk semua|Warna botol|Jumlah pekerja~Kolostrum dan cara pemberiannya perlu prosedur yang mempertimbangkan kondisi pedet.',
 '子牛~こうし~pedet|初乳~しょにゅう~kolostrum|哺乳~ほにゅう~pemberian susu|器具~きぐ~peralatan')
L('Alur pemerahan dan pengamatan ambing',
 'Pemerahan yang baik memadukan identitas sapi, kebersihan, kondisi ambing, operasi alat, dan pemisahan produk yang tidak memenuhi ketentuan. Periksa catatan hewan termasuk status pengobatan sebelum susu digabungkan. Prosedur persiapan, pemasangan, pelepasan, dan perawatan setelah pemerahan mengikuti sistem fasilitas. Jangan mengubah setelan vakum atau membiarkan alat rusak karena susu tetap keluar. Perubahan susu atau ambing dilaporkan sebagai fakta, bukan didiagnosis dan diobati sendiri.',
 'Cocokkan sapi dengan catatan dan status produk|Lakukan persiapan higiene sesuai SOP|Perah memakai alat yang diperiksa|Pisahkan kelainan serta catat hasil',
 'Catatan menunjukkan susu satu sapi harus dipisahkan setelah pengobatan.',
 'Pisahkan sesuai ketentuan dan konfirmasikan statusnya|Campurkan agar tak rugi|Ubah identitas sapi|Anggap aman jika warna normal',
 'Susu berwarna normal menjamin boleh digabung setelah pengobatan?~Tidak status penggunaan dan catatan harus dipastikan|Ya selalu|Ya bila dingin|Ya jika jumlah sedikit~Warna tidak membuktikan pemenuhan batas penggunaan produk setelah pengobatan.',
 '搾乳~さくにゅう~pemerahan|乳房~にゅうぼう~ambing|乳頭~にゅうとう~puting|牛乳~ぎゅうにゅう~susu sapi')
L('Pendinginan susu dan higiene jalur produk',
 'Susu mudah mengalami penurunan mutu bila higiene dan kondisi penyimpanan tidak terkendali. Tangki, pipa, serta alat dibersihkan melalui prosedur fasilitas dengan produk dan parameter yang sesuai. Jangan mencampurkan residu bahan pembersih dengan produk. Kondisi pendinginan dipantau serta dicatat; angka target mengikuti persyaratan fasilitas dan produk. Gangguan alat, listrik, atau penyimpangan suhu memerlukan pemisahan serta keputusan berwenang sebelum pengiriman, bukan sekadar menurunkan suhu kembali lalu menghapus kejadian.',
 'Periksa kebersihan jalur dan status tangki|Pastikan produk yang boleh masuk|Pantau pendinginan sesuai spesifikasi|Catat penyimpangan dan tahan bila diarahkan',
 'Pendingin tangki berhenti beberapa waktu dan data suhu tidak lengkap.',
 'Laporkan dan tahan keputusan pengiriman sampai dinilai|Hapus catatan gangguan|Campur susu baru untuk menutupi|Kirim asal tampak putih',
 'Apa tindakan pada penyimpangan pendinginan susu?~Catat laporkan dan ikuti keputusan produk|Hapus bukti|Nilai dari warna saja|Tambah pewangi~Keputusan keamanan produk membutuhkan data proses, bukan penampilan saja.',
 '冷却~れいきゃく~pendinginan|貯乳~ちょにゅう~penyimpanan susu|洗浄~せんじょう~pencucian|温度~おんど~suhu')

c.module('livestock-production','Peternakan · Sapi potong, babi, dan unggas','livestock','Manajemen kelompok, kelahiran babi, unggas pedaging, serta telur.')
L('Sapi potong dan pemantauan kelompok',
 'Pemeliharaan sapi potong memperhatikan fase pertumbuhan, ransum, kondisi tubuh, kesehatan kaki, serta lingkungan. Perubahan bobot atau konsumsi ditafsirkan dari catatan dan kondisi, bukan dengan menambah konsentrat secara mendadak. Pemindahan dan penimbangan memerlukan jalur aman serta alat sesuai. Hindari titik jepit dan kepadatan tidak sesuai. Kelompok yang berbeda status atau ukuran tidak dicampur tanpa rencana karena perilaku dan persaingan dapat berubah. Catatan identitas memastikan keputusan tepat individu atau kelompok.',
 'Periksa kelompok dan catatan fase|Amati konsumsi serta kondisi tubuh|Lakukan pemindahan atau ukur dengan aman|Catat hasil dan perubahan',
 'Konsumsi turun setelah kelompok dicampur tanpa rencana.',
 'Laporkan perubahan kelompok serta konsumsi untuk penilaian|Tambah konsentrat berlipat sendiri|Paksa semua makan|Hapus identitas kelompok',
 'Pemantauan produksi sapi perlu membandingkan apa?~Catatan konsumsi pertumbuhan dan kondisi kelompok|Jumlah pagar saja|Warna gudang|Satu pengamatan tanpa identitas~Data berurutan membantu membedakan perubahan nyata dan dugaan penyebab.',
 '肉牛~にくぎゅう~sapi potong|増体~ぞうたい~pertambahan bobot|群~むれ~kelompok hewan|体重~たいじゅう~bobot tubuh')
L('Babi, persaingan pakan, dan pemindahan',
 'Babi dapat bereaksi kuat terhadap perubahan kelompok, suhu, serta perlakuan kasar. Kepadatan, akses pakan dan air, lantai, serta lingkungan mengikuti fase dan rencana fasilitas. Pemindahan memerlukan jalur dan alat bantu yang disetujui; jangan menarik telinga atau melakukan kekerasan. Perkelahian, luka, dan penurunan konsumsi dicatat lalu dilaporkan. Jangan mencampur kelompok sehat dengan hewan yang status kesehatannya belum jelas. Pembersihan kandang kosong dan perpindahan kelompok harus mematuhi biosekuriti.',
 'Periksa status serta rencana kelompok babi|Siapkan jalur dan alat pemindahan|Pindahkan secara tenang sesuai pelatihan|Amati luka serta akses pakan air',
 'Setelah pemindahan ada babi terluka dan tidak mendapat akses pakan.',
 'Laporkan dan ikuti pengaturan kelompok sesuai SOP|Biarkan agar terbiasa|Kurangi semua pakan|Campur ke kelompok sakit',
 'Apa yang penting setelah perubahan kelompok babi?~Pengamatan perkelahian luka dan akses sumber daya|Warna pintu|Menghapus semua catatan|Menutup akses air~Perubahan kelompok dapat meningkatkan persaingan dan risiko gangguan.',
 '豚~ぶた~babi|豚舎~とんしゃ~kandang babi|給餌器~きゅうじき~tempat pakan|傷~きず~luka')
L('Induk babi dan anak babi',
 'Pemeliharaan induk dan anak babi membutuhkan pengamatan kelahiran, perilaku induk, akses menyusu, serta kondisi lingkungan sesuai fase. Anak yang lemah, tidak menyusu, atau terjepit memerlukan respons mengikuti protokol dengan keselamatan pekerja dan induk. Pengaturan pemanas harus dipantau dan tidak diasumsikan cocok untuk semua hewan. Tindakan teknis pada anak babi serta pemberian obat memerlukan pelatihan dan kewenangan. Catat identitas induk, waktu, jumlah anak, dan kelainan secara faktual.',
 'Periksa induk dan area kelahiran|Amati anak serta akses menyusu|Minta bantuan untuk kelainan sesuai protokol|Catat jumlah dan kondisi',
 'Beberapa anak babi lemah dan tidak tampak menyusu.',
 'Segera laporkan dan jalankan protokol penanganan anak|Berikan obat acak|Naikkan panas tanpa pantau|Pindah ke induk lain tanpa identitas',
 'Catatan kelahiran anak babi sebaiknya memuat apa?~Induk waktu jumlah dan kondisi anak|Warna baju operator|Harga truk|Dugaan tanpa pengamatan~Catatan yang tepat mendukung pemantauan induk dan anak serta keputusan petugas.',
 '母豚~ぼとん~induk babi|子豚~こぶた~anak babi|授乳~じゅにゅう~menyusui|保温~ほおん~penghangatan')
L('Unggas pedaging dan pemeriksaan flok',
 'Unggas dipelihara menurut fase, tipe produksi, kepadatan, serta lingkungan. Sebaran anak ayam, konsumsi air dan pakan, kondisi alas, serta kematian memberi informasi penting. Flok mengumpul pada satu sisi tidak selalu normal; periksa cahaya, panas, angin, atau gangguan tanpa menebak satu penyebab. Pintu dan alat perlu pengendalian biosekuriti. Unggas mati ditangani melalui prosedur khusus, dicatat, dan tidak dicampurkan ke pakan atau produk. Penurunan tajam kondisi flok harus segera dilaporkan.',
 'Periksa identitas flok dan fase|Amati sebaran konsumsi serta mortalitas|Periksa lingkungan dan laporkan kelainan|Catat tindakan serta hasil pemantauan',
 'Banyak anak ayam bergerombol di satu sudut dan konsumsi turun.',
 'Periksa lingkungan dan laporkan pola flok serta konsumsi|Anggap semua kondisi normal|Tambah kepadatan|Tutup semua sumber udara',
 'Mengapa sebaran unggas diamati?~Dapat menunjukkan ketidaknyamanan lingkungan atau gangguan|Menentukan jenis pupuk|Mengganti semua pengukuran|Tidak ada kaitan pemeliharaan~Pola kelompok dibaca bersama suhu, aliran udara, konsumsi, dan kondisi lain.',
 '鶏~にわとり~ayam|鶏群~けいぐん~flok unggas|雛~ひな~anak ayam|死亡数~しぼうすう~jumlah kematian')

c.module('livestock-products','Peternakan · Produk, limbah, dan bahasa kerja','livestock','Telur, kotoran ternak, keselamatan ruang terbatas, serta laporan terintegrasi.')
L('Ayam petelur dan penanganan telur',
 'Mutu telur dipengaruhi kesehatan flok, lingkungan, kebersihan pengumpulan, kerusakan cangkang, serta penyimpanan. Telur retak, sangat kotor, atau tidak sesuai standar dipisahkan sesuai sistem fasilitas. Jangan mencuci atau mengoles bahan sembarangan karena metode penanganan harus mengikuti ketentuan produk. Kumpulkan dengan lembut, hindari tekanan dan pencampuran lot. Catat produksi, jumlah tidak sesuai, serta perubahan mendadak. Penurunan produksi perlu dilaporkan bersama konsumsi dan kondisi flok, bukan ditutupi dengan label lama.',
 'Periksa lot dan standar telur|Kumpulkan dengan wadah bersih|Pisahkan retak atau tidak sesuai|Catat produksi dan simpan sesuai ketentuan',
 'Ditemukan telur retak saat pengumpulan.',
 'Pisahkan sesuai prosedur telur tidak sesuai|Campur ke telur utuh|Tutup retakan dengan perekat|Ganti label agar lolos',
 'Mengapa telur retak dipisahkan?~Memenuhi pengendalian mutu dan risiko kontaminasi|Menambah berat telur|Mengubah warna alami|Mengganti pakan ayam~Cangkang rusak memerlukan penanganan sesuai ketentuan produk.',
 '鶏卵~けいらん~telur ayam|産卵~さんらん~bertelur|卵殻~らんかく~cangkang telur|破卵~はらん~telur pecah')
L('Kotoran ternak dan pengomposan',
 'Kotoran ternak dikelola agar tidak mencemari air, pakan, produk, atau lingkungan sekitar. Penyimpanan serta pengomposan membutuhkan rencana bahan, kelembapan, aerasi, dan pemantauan proses yang ditetapkan. Tidak semua tumpukan yang sudah lama menjadi kompos matang. Pisahkan alur kotor dari area produksi dan gunakan alat khusus. Aliran limpasan serta bau yang berubah perlu dilaporkan. Pekerja tidak mencampur bahan kimia atau masuk tempat penampungan limbah untuk mempercepat pekerjaan.',
 'Periksa rencana dan area limbah|Pisahkan alur kotor dari pakan produk|Pantau proses sesuai SOP|Catat hasil serta cegah limpasan',
 'Air limbah mengalir menuju area penyimpanan pakan.',
 'Laporkan dan kendalikan pencemaran sesuai prosedur aman|Biarkan karena berasal dari ternak|Campur ke pakan|Tutup catatan kejadian',
 'Umur tumpukan saja membuktikan kompos matang?~Tidak proses dan indikator harus dinilai|Ya semua tumpukan lama aman|Ya bila dekat kandang|Ya tanpa pemantauan~Kematangan bergantung proses dan indikator yang ditetapkan, bukan usia saja.',
 'ふん尿~ふんにょう~kotoran dan urine|堆肥~たいひ~kompos|汚水~おすい~air tercemar|処理~しょり~pengolahan')
L('Gas limbah, listrik, dan keadaan darurat kandang',
 'Tempat limbah dan ruang tertutup dapat mengandung gas beracun atau kekurangan oksigen walaupun tampak biasa. Bau tidak menjadi alat ukur keselamatan. Jangan masuk atau mencoba menyelamatkan dengan ikut turun tanpa tim serta prosedur khusus. Mesin, kipas, pompa, dan pemanas juga memiliki risiko listrik serta kebakaran. Kenali jalur evakuasi dan cara meminta bantuan. Penghentian pekerjaan, pengamanan dari lokasi aman, dan laporan cepat lebih penting daripada menyelesaikan target produksi saat ada bahaya.',
 'Kenali area terbatas dan kontak darurat|Jaga orang dari area berbahaya|Aktifkan bantuan sesuai SOP|Catat kejadian setelah situasi aman',
 'Rekan jatuh tak sadar di dalam lubang penampungan limbah.',
 'Panggil bantuan darurat dan jangan ikut masuk tanpa tim terlatih|Langsung turun sendiri|Uji udara dengan mencium|Nyalakan api untuk melihat',
 'Bau normal membuktikan ruang limbah aman dimasuki?~Tidak diperlukan prosedur ruang terbatas dan penilaian berwenang|Ya selalu|Ya jika cepat|Ya bila memakai sepatu bot~Gas berbahaya dan kekurangan oksigen tidak dapat dinilai dengan penciuman.',
 '有毒ガス~ゆうどくがす~gas beracun|酸欠~さんけつ~kekurangan oksigen|避難~ひなん~evakuasi|救助~きゅうじょ~penyelamatan')
L('Catatan peternakan dan instruksi Jepang',
 'Laporan kerja menghubungkan identitas, waktu, jumlah, satuan, kondisi, serta tindakan yang telah dilakukan. 飼料 berarti pakan, 給水 pasokan air, dan 異常 kelainan. Jangan menjawab paham bila angka atau kelompok belum jelas. Ulangi instruksi penting sebelum bekerja dan tanyakan ketidakjelasan. Contoh kebutuhan pakan 2 kg per ekor untuk 5 ekor adalah 10 kg dalam latihan hitung, bukan rekomendasi ransum nyata. Catatan pengobatan dan penggunaan produk tidak diubah tanpa kewenangan.',
 'Baca identitas jumlah dan satuan|Konfirmasikan instruksi yang belum jelas|Kerjakan rencana dan catat kondisi aktual|Serahkan laporan objektif ke penanggung jawab',
 'Instruksi contoh meminta 2 kg pakan per ekor untuk 5 ekor.',
 'Hitung 10 kg lalu pastikan kelompok dan rencana|Berikan 2 kg total tanpa periksa|Berikan 100 kg|Ganti satuan kg menjadi liter',
 'Contoh 2 kg per ekor untuk 5 ekor menghasilkan berapa?~10 kg|2 kg|25 kg|0,4 kg~Kebutuhan total contoh = 2 dikali 5 = 10 kg; dosis nyata berasal dari rencana pemeliharaan.',
 '指示~しじ~instruksi|頭数~とうすう~jumlah ternak|数量~すうりょう~kuantitas|報告~ほうこく~laporan')

if __name__=='__main__': c.write()
