"""Compile the original Indonesian Kaigo course into an offline Flutter asset.

Run: python scripts/build_kaigo_curriculum.py
Content is authored below; no network requests or external textbook reproduction.
"""
import json
from pathlib import Path

modules = []


def module(id, title, description):
    modules.append(dict(id=id, title=title, description=description, lessons=[]))


def q(prompt, options, answer, explanation):
    return dict(prompt=prompt, options=options.split("|"), answer=answer,
                explanation=explanation)


def lesson(id, title, objectives, concept, practice, steps, caution,
           scenario, response, words, phrase, checks):
    def word(value):
        japanese, reading, meaning = value.split("~")
        return dict(japanese=japanese, reading=reading, meaning=meaning)
    modules[-1]["lessons"].append(dict(
        id=id, title=title, minutes=20,
        objectives=objectives.split("|"),
        sections=[dict(title="Pahami konsep", body=concept),
                  dict(title="Penerapan di tempat kerja", body=practice)],
        steps=steps.split("|"), caution=caution, scenario=scenario,
        response=response, vocabulary=[word(w) for w in words.split("|")],
        dialogue=[word(p) for p in phrase.split("|")], checks=checks))


module("foundations", "Dasar Kaigo, martabat, dan etika",
       "Memahami peran caregiver, pilihan pengguna layanan, serta batas kewenangan.")
lesson("kaigo-01", "Kaigo dan kehidupan yang bermakna",
       "Membedakan dukungan hidup dengan mengambil alih seluruh kegiatan|Mengidentifikasi ADL, IADL, dan kualitas hidup",
       "Kaigo (介護) adalah dukungan agar seseorang dapat menjalani kehidupan sehari-hari sesuai kemampuan dan keinginannya. Pengguna layanan disebut riyōsha, bukan sekadar kumpulan penyakit. Kebutuhan mereka mencakup tubuh, perasaan, hubungan sosial, kebiasaan, serta makna hidup. Bantuan yang cepat belum tentu menjadi bantuan yang tepat bila pengguna kehilangan pilihan dan kesempatan melakukan sesuatu sendiri.\n\nADL adalah kegiatan dasar seperti makan, berpakaian, berpindah, mandi, dan menggunakan toilet. IADL mencakup kegiatan yang lebih kompleks, misalnya berbelanja, mengelola jadwal, dan menggunakan transportasi. QOL atau kualitas hidup menanyakan apakah kehidupan terasa nyaman dan berarti bagi orang tersebut. Ketiganya membantu caregiver melihat kebutuhan secara utuh.",
       "Awali dengan mengenali kebiasaan: jam bangun, makanan yang disukai, bahasa, kegiatan, serta bagian tubuh yang masih dapat digunakan. Amati kemampuan nyata, lalu cocokkan dengan rencana perawatan. Dukungan dapat berupa menata lingkungan, memberi petunjuk, mengawasi, atau membantu sebagian. Tidak setiap kebutuhan memerlukan bantuan penuh.\n\nMisalnya pengguna mampu menyisir rambut bila sisir diletakkan di tangan kanannya. Menyiapkan sisir dan cermin menjaga kemampuan serta rasa percaya diri. Catatan yang baik menyebut bantuan yang diperlukan dan respons pengguna; jangan hanya menulis 'sudah dibantu'.",
       "Baca rencana dukungan dan kebiasaan pengguna.|Tanyakan kegiatan yang ingin dilakukan.|Amati kemampuan, hambatan, dan kelelahan.|Bantu hanya bagian yang diperlukan lalu evaluasi respons.",
       "Kecepatan kerja tidak boleh menjadi alasan mengambil alih kemampuan pengguna.",
       "Pak Sato makan perlahan tetapi dapat memegang sendok. Petugas baru ingin langsung menyuapi agar jadwal selesai.",
       "Sediakan waktu dan posisi nyaman, dekatkan peralatan yang sesuai, lalu bantu bagian yang sulit sesuai rencana. Nilai kemampuan dan keinginan Pak Sato sebelum memilih bantuan penuh.",
       "介護~かいご~dukungan perawatan|利用者~りようしゃ~pengguna layanan|日常生活~にちじょうせいかつ~kehidupan sehari-hari|生活の質~せいかつのしつ~kualitas hidup",
       "ご自分でできますか。~ごじぶんでできますか。~Apakah Anda dapat melakukannya sendiri?|難しいところをお手伝いします。~むずかしいところをおてつだいします。~Saya akan membantu bagian yang sulit.",
       [q("Pengguna masih mampu menyisir rambut. Bantuan yang mendukung kemandirian adalah…", "Menyisir tanpa bertanya|Menyiapkan sisir dan membantu bagian sulit|Melarang pengguna mencoba", 1, "Lingkungan dan bantuan sebagian mempertahankan kemampuan yang masih dimiliki."),
        q("Contoh IADL adalah…", "Mengelola jadwal belanja|Menelan makanan|Mengubah posisi tidur", 0, "IADL merupakan kegiatan kehidupan yang lebih kompleks dibanding ADL dasar.")])
lesson("kaigo-02", "Martabat, persetujuan, dan kemandirian",
       "Meminta persetujuan sebelum bantuan|Menjaga pilihan dan privasi saat kebutuhan tubuh dibantu",
       "Martabat berarti setiap orang tetap berharga, termasuk ketika membutuhkan bantuan penuh. Menggunakan nama yang disukai, berbicara sebagai orang dewasa, dan menjelaskan tindakan merupakan bentuk penghormatan yang nyata. Kemandirian tidak berarti memaksa seseorang melakukan semuanya sendiri. Kemandirian juga berarti menentukan pilihan, menyampaikan keinginan, dan menerima dukungan yang sesuai.\n\nPersetujuan adalah proses komunikasi, bukan sekadar satu kata 'ya'. Jelaskan apa yang akan dilakukan, alasan, dan pilihan yang tersedia dengan bahasa yang dapat dipahami. Orang yang sulit bicara tetap dapat menunjukkan persetujuan atau keberatan melalui gerakan, ekspresi, dan perilaku. Pastikan pemahaman dan perhatikan perubahan respons selama bantuan.",
       "Saat membantu mengganti pakaian, tutup pintu atau tirai dan tutupi bagian tubuh yang tidak sedang dirawat. Berikan pilihan yang sederhana dan nyata, misalnya pakaian biru atau hijau. Jangan menawarkan pilihan yang sebenarnya tidak akan dihormati.\n\nJika pengguna menolak mandi, cari alasan: kedinginan, malu, nyeri, takut jatuh, atau waktu yang tidak cocok. Dengarkan, tawarkan alternatif yang diperbolehkan, dan laporkan kepada tim bila kebutuhan belum terpenuhi. Kesulitan memahami keputusan perlu dukungan tim, bukan alasan caregiver menetapkan semua keputusan sendiri.",
       "Sapa pengguna dan jelaskan tujuan bantuan.|Tawarkan pilihan yang aman dan mudah dipahami.|Konfirmasi persetujuan serta tutup area privat.|Amati respons sepanjang bantuan; hentikan dan diskusikan bila menolak.",
       "Jangan mengancam, memaksa, atau mempermalukan pengguna untuk memperoleh kepatuhan.",
       "Ibu Tanaka menolak mandi sore. Ia berkata ruangan terasa dingin.",
       "Akui rasa dinginnya, periksa lingkungan bersama tim, dan tawarkan waktu atau bentuk kebersihan lain sesuai rencana. Catat alasan penolakan serta alternatif yang diterima.",
       "尊厳~そんげん~martabat|自立~じりつ~kemandirian|同意~どうい~persetujuan|意思~いし~kehendak",
       "今、お着替えしてもよろしいですか。~いま、おきがえしてもよろしいですか。~Bolehkah kita berganti pakaian sekarang?|どちらがいいですか。~どちらがいいですか。~Yang mana yang Anda pilih?",
       [q("Saat pengguna menolak mandi, langkah pertama adalah…", "Memaksa agar jadwal selesai|Mengabaikan selamanya|Mendengarkan alasan dan mencari alternatif", 2, "Penolakan dapat menunjukkan kebutuhan atau rasa tidak nyaman yang perlu dipahami."),
        q("Kemandirian juga mencakup…", "Hak memilih dengan dukungan|Harus selalu tanpa bantuan|Tidak boleh menggunakan alat bantu", 0, "Pilihan dan penggunaan dukungan tetap merupakan bagian dari kemandirian.")])
lesson("kaigo-03", "Etika profesi, kerahasiaan, dan kerja tim",
       "Melindungi informasi pengguna|Mengenali batas pekerjaan dan meminta bantuan yang tepat",
       "Caregiver bekerja dalam tim yang dapat melibatkan perawat, dokter, pekerja sosial, terapis, ahli gizi, dan pengelola layanan. Peran caregiver terutama mendukung kehidupan harian, mengamati, mendokumentasikan, serta menyampaikan perubahan. Rencana yang sama membantu petugas memberikan dukungan konsisten. Kompetensi dan kewenangan ditetapkan oleh aturan tempat kerja dan pelatihan, bukan oleh keberanian mencoba.\n\nKerahasiaan melindungi identitas, kesehatan, keluarga, foto, serta percakapan pribadi pengguna. Informasi hanya dibagikan kepada pihak berwenang yang memerlukannya untuk pelayanan. Membicarakan pengguna di lift, mengirim foto melalui akun pribadi, atau membagikan cerita dengan identitas tersamar yang masih mudah dikenali dapat melanggar privasi.",
       "Jika instruksi tidak jelas, ulangi inti instruksi dan minta penjelasan sebelum bertindak. Ketika tugas berada di luar pelatihan, sampaikan keterbatasan dan hubungi penanggung jawab. Jangan mengubah obat, dosis, atau keputusan klinis sendiri. Bantuan terkait obat hanya dilakukan sesuai aturan, kewenangan, dan supervisi yang berlaku.\n\nGunakan sistem catatan resmi, kunci perangkat, serta periksa penerima sebelum mengirim laporan. Permintaan informasi dari keluarga juga perlu verifikasi identitas dan kewenangan. Hubungan profesional tetap ramah, tetapi hadiah, uang, dan urusan pribadi mengikuti kebijakan fasilitas.",
       "Pastikan instruksi, rencana, dan kewenangan tugas.|Sampaikan bila belum dilatih atau membutuhkan supervisi.|Bagikan informasi melalui saluran resmi kepada tim terkait.|Simpan dokumen dan perangkat dengan aman.",
       "Jangan melakukan tindakan medis atau mengubah obat berdasarkan materi aplikasi ini.",
       "Kerabat yang belum dikenal menelepon meminta diagnosis pengguna. Petugas juga diminta memberi dosis obat tambahan.",
       "Verifikasi permintaan informasi melalui penanggung jawab dan kebijakan privasi. Untuk obat tambahan, hubungi perawat; jangan menentukan dosis sendiri.",
       "守秘義務~しゅひぎむ~kewajiban menjaga rahasia|連携~れんけい~koordinasi|看護師~かんごし~perawat|個人情報~こじんじょうほう~informasi pribadi",
       "看護師に確認します。~かんごしにかくにんします。~Saya akan memastikannya kepada perawat.|もう一度説明していただけますか。~もういちどせつめいしていただけますか。~Bisa dijelaskan sekali lagi?",
       [q("Foto pengguna sebaiknya…", "Dikirim ke grup pribadi|Dikelola sesuai persetujuan dan kebijakan resmi|Diunggah bila wajah diburamkan saja", 1, "Persetujuan, tujuan, penerima, dan kebijakan penggunaan tetap perlu diperiksa."),
        q("Instruksi obat tidak jelas. Tindakan yang tepat adalah…", "Menebak dari kemasan|Mengikuti pengalaman teman|Mengonfirmasi kepada petugas berwenang", 2, "Keputusan obat memerlukan kewenangan dan instruksi yang jelas.")])

module("body", "Tubuh, penuaan, dan disabilitas",
       "Menghubungkan fungsi tubuh dengan kebutuhan dukungan, tanpa mendiagnosis.")
lesson("kaigo-04", "Fungsi tubuh dan pengamatan harian",
       "Menghubungkan organ dengan aktivitas sehari-hari|Membedakan pengamatan dengan diagnosis",
       "Tubuh bekerja sebagai satu kesatuan. Otak dan saraf mengatur gerakan, perhatian, serta sensasi; otot, tulang, dan sendi menopang posisi dan perpindahan. Jantung dan pembuluh darah mengedarkan darah, paru-paru mendukung pernapasan, sementara sistem pencernaan serta ginjal mengolah makanan dan mengatur pembuangan. Gangguan satu fungsi dapat mengubah cara seseorang makan, berjalan, berbicara, atau menggunakan toilet.\n\nCaregiver tidak perlu menetapkan diagnosis untuk menyampaikan pengamatan yang berguna. Bandingkan keadaan hari ini dengan keadaan biasanya: cara bicara, warna wajah, napas, nafsu makan, aktivitas, dan pola buang air. Keluhan pengguna merupakan data penting, bukan sesuatu yang boleh diabaikan hanya karena pemeriksaan tampak normal.",
       "Gunakan urutan melihat, mendengar, bertanya, dan mencatat. Bila ditugaskan mengukur tanda vital, gunakan alat serta metode yang telah dilatih dan ikuti batas pelaporan fasilitas. Angka harus disertai waktu, kondisi saat pengukuran, dan gejala; jangan menentukan keamanan hanya dari satu angka.\n\nPerubahan mendadak seperti sulit dibangunkan, kesulitan bernapas, atau kelemahan baru memerlukan respons segera sesuai prosedur. Hindari menyimpulkan 'hanya tua'. Sampaikan fakta dan panggil petugas berwenang sebelum melanjutkan aktivitas yang berisiko.",
       "Kenali kondisi dasar pengguna dari catatan dan tim.|Amati aktivitas, napas, ekspresi, serta keluhan.|Catat waktu dan perubahan yang terlihat.|Laporkan perubahan dan ikuti instruksi penanggung jawab.",
       "Pengamatan caregiver tidak menggantikan diagnosis atau penilaian klinis.",
       "Pak Mori biasanya aktif berbicara. Pagi ini ia mengantuk tidak biasa dan sulit mengikuti percakapan.",
       "Hentikan aktivitas berisiko, pastikan keselamatan, dan segera laporkan perubahan kesadaran beserta waktu mulai serta kondisi sebelumnya. Jangan langsung memberi label demensia.",
       "呼吸~こきゅう~pernapasan|脈拍~みゃくはく~denyut nadi|意識~いしき~kesadaran|観察~かんさつ~pengamatan",
       "いつから苦しいですか。~いつからくるしいですか。~Sejak kapan terasa sesak atau tidak nyaman?|いつもと様子が違います。~いつもとようすがちがいます。~Keadaannya berbeda dari biasanya.",
       [q("Laporan pengamatan yang paling jelas adalah…", "Beliau aneh|Mungkin penyakit berat|Pukul 08.00 sulit dibangunkan, berbeda dari biasanya", 2, "Waktu, perilaku teramati, dan perbandingan dengan kondisi dasar membantu tim menilai perubahan."),
        q("Satu angka tanda vital harus…", "Ditafsirkan bersama gejala dan prosedur|Selalu dianggap normal|Digunakan untuk menentukan diagnosis sendiri", 0, "Konteks serta jalur pelaporan diperlukan untuk pengamatan yang aman.")])
lesson("kaigo-05", "Penuaan, kebutuhan psikologis, dan kemampuan tersisa",
       "Mengenali variasi penuaan|Menyesuaikan lingkungan tanpa merendahkan pengguna",
       "Penuaan dapat disertai perubahan kekuatan, keseimbangan, penglihatan, pendengaran, kulit, dan kecepatan merespons. Namun perubahan tidak sama pada semua orang. Umur saja tidak menjelaskan kemampuan, minat, atau keputusan seseorang. Perhatikan riwayat hidup, lingkungan, penyakit yang telah dinilai tenaga kesehatan, serta bantuan yang biasa digunakan.\n\nKehilangan pasangan, pindah tempat tinggal, atau berkurangnya peran sosial dapat memengaruhi perasaan aman dan harga diri. Pengguna mungkin ingin tetap mengurus tanaman atau memilih pakaian meskipun membutuhkan bantuan fisik. Memberi kesempatan ikut kegiatan yang berarti mendukung kehidupan, bukan hanya memenuhi jadwal perawatan.",
       "Sesuaikan pencahayaan, kurangi kebisingan, letakkan barang yang sering digunakan dalam jangkauan, dan beri waktu merespons. Pastikan kacamata atau alat bantu dengar digunakan sesuai keinginan dan rencana. Jangan berteriak secara otomatis atau berbicara seperti kepada anak kecil.\n\nAmati perubahan baru secara terpisah dari perubahan yang sudah biasa. Hilang nafsu makan, menarik diri, atau nyeri baru perlu disampaikan kepada tim. Jangan menentukan bahwa semua lansia pasti lupa, depresi, atau tidak mampu belajar.",
       "Tanyakan kebiasaan, minat, serta alat bantu yang biasa digunakan.|Sesuaikan cahaya, suara, dan jangkauan barang.|Berikan waktu dan bantuan sebagian sesuai kemampuan.|Catat perubahan baru serta respons terhadap penyesuaian.",
       "Perubahan mendadak tidak boleh dianggap sebagai akibat usia tanpa penilaian tim.",
       "Ibu Yamada sering tidak menjawab di ruang makan yang ramai, tetapi merespons ketika diajak bicara di tempat tenang.",
       "Kurangi kebisingan, hadapkan wajah, dan periksa kebutuhan alat bantu sesuai rencana. Laporkan pengamatan spesifik; jangan menganggap ia sengaja tidak mau berkomunikasi.",
       "老化~ろうか~penuaan|視力~しりょく~kemampuan melihat|聴力~ちょうりょく~kemampuan mendengar|残存能力~ざんぞんのうりょく~kemampuan yang masih dimiliki",
       "ゆっくりで大丈夫です。~ゆっくりでだいじょうぶです。~Tidak apa-apa, perlahan saja.|静かな場所でお話ししましょう。~しずかなばしょでおはなししましょう。~Mari berbicara di tempat yang tenang.",
       [q("Cara mendukung pengguna yang membutuhkan waktu menjawab adalah…", "Memberi waktu dan mengurangi gangguan|Menjawab untuknya setiap saat|Mengulang pertanyaan dengan marah", 0, "Penyesuaian komunikasi memberi ruang bagi kemampuan yang masih dimiliki."),
        q("Penuaan pada setiap pengguna…", "Selalu sama|Perlu dinilai secara individual|Selalu menyebabkan demensia", 1, "Kemampuan dan kebutuhan tidak dapat diputuskan hanya berdasarkan umur.")])
lesson("kaigo-06", "Disabilitas, hemiparesis, dan aksesibilitas",
       "Menyesuaikan dukungan dengan fungsi dan lingkungan|Menggunakan informasi sisi tubuh dari rencana perawatan",
       "Disabilitas dapat berkaitan dengan gerakan, penglihatan, pendengaran, komunikasi, atau fungsi kognitif. Hambatan juga muncul dari lingkungan: pintu sempit, informasi yang tidak terbaca, atau alat yang sulit dijangkau. Dukungan yang baik mengurangi hambatan dan memperhatikan apa yang ingin dilakukan pengguna.\n\nHemiparesis berarti kelemahan pada satu sisi tubuh. Kondisi dan kemampuan masing-masing pengguna berbeda, sehingga sisi yang lebih kuat, nyeri, rentang gerak, dan cara transfer harus diketahui dari penilaian tim. Jangan menarik lengan yang lemah atau menganggap sisi tersebut tidak mempunyai sensasi. Gangguan bicara juga tidak otomatis berarti pengguna tidak memahami percakapan.",
       "Pada bantuan berpakaian untuk kelemahan satu sisi, prinsip umum adalah memasukkan sisi yang membutuhkan bantuan terlebih dahulu dan melepas dari sisi yang lebih mudah terlebih dahulu. Sesuaikan dengan rencana, jenis pakaian, nyeri, dan latihan petugas. Jelaskan tiap tahap dan tetap libatkan tangan atau kemampuan yang dapat digunakan.\n\nGunakan petunjuk visual, tulisan, gambar, atau waktu tambahan untuk komunikasi sesuai kebutuhan. Penempatan alat makan, bel panggil, dan kursi harus mengikuti hasil penilaian pengguna, bukan kebiasaan petugas. Konsultasikan perubahan alat bantu kepada tim.",
       "Periksa kemampuan, sisi tubuh, nyeri, dan rencana bantuan.|Tanyakan cara yang biasa dirasakan nyaman.|Atur alat serta lingkungan agar dapat diakses.|Bantu bertahap tanpa menarik anggota tubuh yang lemah.",
       "Teknik transfer untuk kelemahan satu sisi harus dipelajari langsung dengan supervisi.",
       "Pak Ito memahami pembicaraan tetapi sulit mengucapkan kata. Petugas baru berbicara hanya kepada keluarganya.",
       "Tetap ajak Pak Ito berkomunikasi langsung, gunakan pilihan singkat atau alat komunikasi yang sesuai, dan beri waktu. Libatkan keluarga sebagai dukungan, bukan pengganti semua keputusan.",
       "障害~しょうがい~disabilitas|片麻痺~かたまひ~kelumpuhan atau kelemahan satu sisi|健側~けんそく~sisi yang tidak terdampak|患側~かんそく~sisi yang terdampak",
       "痛いところはありますか。~いたいところはありますか。~Apakah ada bagian yang sakit?|こちらの手を使えますか。~こちらのてをつかえますか。~Apakah tangan ini dapat digunakan?",
       [q("Kesulitan bicara berarti…", "Pasti tidak memahami|Tidak perlu diajak berbicara|Kemampuan memahami tetap perlu diperhatikan", 2, "Kemampuan berbicara dan memahami tidak selalu terganggu dengan cara yang sama."),
        q("Saat berpakaian dengan kelemahan satu sisi, prinsip umum memasukkan pakaian adalah…", "Sisi yang lebih kuat selalu dahulu|Sisi yang memerlukan bantuan dahulu, sesuai rencana|Menarik lengan agar cepat", 1, "Urutan bantuan bertujuan mengurangi kesulitan; tetap sesuaikan kondisi dan rencana.")])

module("dementia", "Demensia dan dukungan perilaku",
       "Memahami perubahan kognitif, merespons emosi, serta mengenali perubahan akut.")
lesson("kaigo-07", "Demensia: memori, orientasi, dan persepsi",
       "Membedakan demensia dari stereotip penuaan|Mengenali pengaruh kognitif terhadap aktivitas sehari-hari",
       "Demensia adalah kondisi yang dapat memengaruhi ingatan, berpikir, komunikasi, serta aktivitas harian. Demensia bukan bagian yang pasti terjadi pada setiap proses penuaan. Diagnosis dilakukan tenaga kesehatan. Untuk caregiver, yang penting adalah memahami kebutuhan individual dan membantu orang merasa aman, bukan menguji atau mempermalukan ingatannya.\n\nKesulitan mengingat kejadian baru dapat membuat pengguna mengulang pertanyaan. Gangguan orientasi dapat menyebabkan bingung tentang tempat atau waktu. Kesulitan menjalankan urutan kegiatan dapat membuat pakaian atau peralatan tampak sulit digunakan. Respons dukungan harus terkait hambatan yang diamati, bukan semata label diagnosis.",
       "Gunakan jadwal yang mudah dikenali, penanda sederhana, dan petunjuk satu tahap. Jika pengguna mengulang pertanyaan tentang waktu makan, jawab dengan tenang dan tunjukkan petunjuk yang sesuai. Jangan memaksa pengguna mengakui bahwa ia sudah menanyakan hal yang sama.\n\nDukung pilihan yang masih dapat dibuat, misalnya memilih minuman dari dua pilihan yang aman. Perhatikan bahasa, kebiasaan, dan riwayat hidup. Koordinasikan perubahan dukungan dengan tim agar petugas memberikan pendekatan konsisten dan tidak membingungkan.",
       "Kenali pola kesulitan dan kemampuan yang masih ada.|Gunakan kalimat pendek serta petunjuk satu langkah.|Atur lingkungan dengan tanda yang mudah dikenali.|Evaluasi apakah dukungan membuat pengguna lebih tenang.",
       "Jangan menyimpulkan diagnosis demensia dari satu kejadian lupa.",
       "Ibu Suzuki bertanya berulang kali kapan makan siang. Petugas mulai berkata 'sudah saya jawab'.",
       "Jawab singkat dengan nada tenang, tunjukkan penanda jadwal, dan amati apakah ia lapar atau cemas. Pengulangan pertanyaan perlu dukungan, bukan teguran.",
       "認知症~にんちしょう~demensia|記憶~きおく~ingatan|見当識~けんとうしき~orientasi waktu tempat dan orang|不安~ふあん~kecemasan",
       "昼ご飯は十二時です。~ひるごはんはじゅうにじです。~Makan siang pukul dua belas.|一緒に確認しましょう。~いっしょにかくにんしましょう。~Mari kita periksa bersama.",
       [q("Pengguna mengulang pertanyaan. Respons yang tepat adalah…", "Menjawab tenang dan memberi petunjuk|Menguji ingatan dengan paksa|Mempermalukan di depan orang lain", 0, "Petunjuk dan ketenangan membantu kebutuhan yang muncul dari kesulitan mengingat."),
        q("Demensia…", "Selalu dialami setiap lansia|Memerlukan penilaian profesional dan dukungan individual|Dapat didiagnosis dari satu kali lupa", 1, "Usia atau satu kejadian lupa tidak cukup untuk menentukan diagnosis.")])
lesson("kaigo-08", "Emosi, penolakan, dan perilaku yang menantang",
       "Mencari kebutuhan di balik perilaku|Memilih respons yang menenangkan tanpa berdebat",
       "Perilaku yang tampak sulit sering merupakan bentuk komunikasi kebutuhan yang belum dipahami. Pengguna bisa merasa takut, kesepian, malu, tidak nyaman, atau bingung. Sebelum memberi label 'rewel', amati apa yang terjadi sebelum, selama, dan sesudah perilaku tersebut. Lingkungan ramai, perubahan petugas, serta instruksi terlalu cepat dapat menjadi konteks penting.\n\nMemvalidasi emosi berarti mengakui perasaan, bukan membenarkan semua isi pernyataan. Bila pengguna berkata ingin pulang, petugas dapat berkata bahwa ia tampak merindukan rumah dan menanyakan hal yang membuatnya nyaman. Berdebat panjang tentang fakta sering tidak membantu hubungan atau keselamatan.",
       "Catat pemicu, ucapan pengguna, respons petugas, dan hasilnya. Coba pendekatan sederhana: turunkan suara, beri ruang, tawarkan kegiatan yang disukai, atau ulangi penjelasan pada waktu yang lebih tepat. Bila kebutuhan tubuh atau kondisi baru dicurigai, laporkan kepada tim.\n\nGunakan pengawasan dan penyesuaian lingkungan sesuai rencana ketika pengguna ingin berjalan. Mengunci, mengikat, atau memberi obat untuk kenyamanan petugas bukan keputusan caregiver. Risiko harus ditangani oleh tim melalui prosedur yang menghormati hak pengguna.",
       "Pastikan keselamatan dan kurangi keramaian.|Dengarkan perasaan tanpa menghakimi.|Cari pemicu serta kebutuhan yang belum terpenuhi.|Tawarkan pilihan, amati hasil, dan laporkan pola kepada tim.",
       "Pembatasan fisik atau pemberian obat bukan solusi mandiri caregiver.",
       "Pak Kato menolak berganti pakaian ketika dua petugas berbicara keras dan bergerak cepat.",
       "Kurangi jumlah suara, jelaskan satu tahap dengan tenang, jaga privasi, dan beri kesempatan menjawab. Catat apakah pendekatan ini mengurangi penolakan.",
       "拒否~きょひ~penolakan|行動~こうどう~perilaku|安心~あんしん~rasa aman|気持ち~きもち~perasaan",
       "心配なことがありますか。~しんぱいなことがありますか。~Apakah ada yang membuat Anda khawatir?|少し休みましょう。~すこしやすみましょう。~Mari beristirahat sebentar.",
       [q("Memvalidasi emosi berarti…", "Mengakui perasaan pengguna|Selalu menyetujui semua fakta|Memaksa pengguna berhenti bicara", 0, "Perasaan dapat diakui tanpa berdebat atau membenarkan informasi yang tidak tepat."),
        q("Catatan perilaku sebaiknya memuat…", "Label 'nakal' saja|Pendapat tanpa waktu|Pemicu, perilaku, respons, dan hasil", 2, "Urutan kejadian membantu tim menentukan dukungan yang sesuai.")])
lesson("kaigo-09", "Perubahan akut dan dukungan keluarga",
       "Melaporkan perubahan mendadak dari kondisi dasar|Mengumpulkan riwayat kebiasaan melalui keluarga secara tepat",
       "Kebingungan yang muncul mendadak atau berubah-ubah tidak boleh langsung dianggap sebagai pola demensia yang biasa. Caregiver perlu menyampaikan perubahan perhatian, kesadaran, aktivitas, atau perilaku kepada petugas klinis. Tugasnya adalah mengamati serta menjaga keselamatan; membedakan penyebab klinis memerlukan penilaian profesional.\n\nKeluarga dapat membantu mengenali kebiasaan, panggilan yang disukai, pekerjaan masa lalu, serta hal yang menenangkan pengguna. Informasi ini dipakai untuk dukungan yang personal. Namun pengguna tetap menjadi pusat percakapan dan pilihan; keluarga tidak otomatis berwenang menerima setiap informasi pribadi atau memutuskan semua hal.",
       "Saat perubahan terjadi, sebutkan kondisi biasanya, waktu terakhir tampak seperti biasa, perubahan yang terlihat, dan tindakan yang sudah dilakukan. Tanyakan informasi tambahan hanya jika tidak menunda panggilan bantuan. Catat tanpa menetapkan diagnosis sendiri.\n\nDalam percakapan keluarga, ajukan pertanyaan terbuka tentang rutinitas dan preferensi. Ulangi pemahaman untuk memastikan informasi akurat. Sampaikan kekhawatiran melalui tim dan hindari menjanjikan hasil yang tidak dapat dipastikan. Petugas juga perlu berdiskusi ketika dukungan keluarga dan keinginan pengguna tampak berbeda.",
       "Bandingkan keadaan dengan kondisi dasar pengguna.|Laporkan perubahan akut segera melalui jalur fasilitas.|Kumpulkan waktu, keluhan, dan tindakan yang telah dilakukan.|Gunakan riwayat keluarga untuk dukungan setelah keselamatan ditangani.",
       "Jangan menunda eskalasi perubahan akut demi menyelesaikan catatan atau wawancara.",
       "Pengguna dengan demensia biasanya mengikuti sapaan. Sore ini ia tiba-tiba tidak dapat memusatkan perhatian dan jauh lebih mengantuk.",
       "Segera panggil petugas berwenang, pastikan pengguna aman, dan laporkan perubahan dari kondisi dasar beserta waktunya. Jangan menganggapnya sekadar demensia bertambah.",
       "急変~きゅうへん~perubahan kondisi mendadak|家族~かぞく~keluarga|普段~ふだん~biasanya|報告~ほうこく~pelaporan",
       "午後から急に様子が変わりました。~ごごからきゅうにようすがかわりました。~Sejak siang keadaannya tiba-tiba berubah.|普段の生活について教えてください。~ふだんのせいかつについておしえてください。~Mohon ceritakan kebiasaan hidupnya.",
       [q("Kebingungan mendadak pada pengguna dengan demensia perlu…", "Diabaikan sebagai usia|Dilaporkan sebagai perubahan baru|Ditunggu sampai shift besok", 1, "Perubahan dari kondisi dasar memerlukan penilaian, bukan asumsi."),
        q("Informasi keluarga paling berguna untuk dukungan personal adalah…", "Kebiasaan dan hal yang menenangkan|Gosip tetangga|Perintah yang selalu mengganti pilihan pengguna", 0, "Riwayat hidup membantu tim memahami kebutuhan, dengan pengguna tetap menjadi pusat dukungan.")])

module("communication", "Komunikasi yang menghormati pengguna",
       "Latihan mendengarkan, menjelaskan bantuan, dan mengatasi hambatan komunikasi.")
lesson("kaigo-10", "Mendengarkan aktif dan komunikasi nonverbal",
       "Menggunakan pertanyaan terbuka dan konfirmasi|Membaca ekspresi tanpa menganggapnya sebagai diagnosis",
       "Komunikasi bukan hanya menyampaikan instruksi. Petugas perlu memahami apa yang dirasakan, dibutuhkan, dan diinginkan pengguna. Mendengarkan aktif dilakukan dengan perhatian, jeda, serta pengulangan inti ucapan menggunakan kata sendiri. Pertanyaan terbuka seperti 'bagian mana yang terasa tidak nyaman?' memberi ruang lebih luas daripada menebak jawaban.\n\nEkspresi, tatapan, posisi tubuh, dan gerakan dapat melengkapi ucapan. Wajah meringis mungkin menunjukkan ketidaknyamanan, tetapi penyebabnya perlu ditanyakan atau dilaporkan. Diam tidak selalu berarti setuju. Hindari menafsirkan perilaku berdasarkan budaya atau usia saja; konfirmasikan dengan pengguna dan informasi tim.",
       "Duduk atau berdiri pada posisi yang membuat wajah terlihat tanpa mengintimidasi. Perkenalkan diri, gunakan nama yang disukai, dan kurangi gangguan. Setelah pengguna bicara, rangkum lalu tanyakan apakah pemahaman sudah benar. Jika pengguna sulit bicara, gunakan alat komunikasi yang telah disepakati.\n\nJangan terburu-buru mengisi semua jeda. Beri waktu untuk menyusun jawaban. Saat dukungan tubuh dilakukan, tetap amati ekspresi dan tanyakan rasa nyaman. Catat ucapan penting secara akurat, bukan menggantinya dengan asumsi seperti 'tidak kooperatif'.",
       "Sapa dan pastikan pengguna dapat melihat serta mendengar.|Ajukan satu pertanyaan yang jelas.|Dengarkan sampai selesai, termasuk jeda.|Rangkum isi dan konfirmasikan pemahaman.",
       "Sentuhan harus sesuai persetujuan, preferensi pengguna, dan konteks perawatan.",
       "Ibu Arai berkata 'saya tidak suka ruang makan'. Petugas hendak menyimpulkan ia tidak lapar.",
       "Tanyakan apa yang terasa tidak nyaman dan dengarkan jawabannya. Bisa jadi suara, posisi kursi, atau pengalaman tertentu yang menjadi hambatan, sehingga asumsi soal nafsu makan belum tepat.",
       "傾聴~けいちょう~mendengarkan aktif|表情~ひょうじょう~ekspresi wajah|確認~かくにん~konfirmasi|会話~かいわ~percakapan",
       "どんなことが気になりますか。~どんなことがきになりますか。~Hal apa yang mengganggu Anda?|椅子が低いということですね。~いすがひくいということですね。~Jadi, kursinya terlalu rendah, ya.",
       [q("Pengguna diam ketika ditawari bantuan. Petugas sebaiknya…", "Langsung menganggap setuju|Mengonfirmasi pemahaman dan respons|Menganggap menolak selamanya", 1, "Diam dapat memiliki banyak makna; persetujuan perlu dikonfirmasi."),
        q("Mendengarkan aktif mencakup…", "Memotong jawaban|Menebak tanpa bertanya|Merangkum dan memastikan pemahaman", 2, "Konfirmasi mencegah salah memahami kebutuhan pengguna.")])
lesson("kaigo-11", "Menjelaskan bantuan dan memberi arahan sederhana",
       "Menggunakan satu instruksi per tahap|Meminta izin dan memeriksa kenyamanan",
       "Penjelasan bantuan yang baik menjawab apa, mengapa, dan bagaimana pengguna dapat ikut serta. Gunakan bahasa sehari-hari sebelum istilah teknis. Kalimat yang pendek memudahkan pengguna mengikuti, terutama ketika perhatian atau pendengaran terbatas. Nada tenang dan tempo yang sesuai lebih berguna daripada volume besar.\n\nArahan harus terkait tindakan saat itu. 'Mari pegang sandaran ini' lebih mudah diikuti daripada beberapa perintah sekaligus. Hindari janji seperti 'pasti tidak sakit'. Sebaliknya jelaskan bahwa pengguna dapat memberi tahu bila terasa sakit atau ingin berhenti. Pertanyaan 'sudah mengerti?' dapat dilengkapi dengan meminta pengguna menunjukkan atau menyampaikan bagian yang dipahami.",
       "Sebelum bantuan, jelaskan tujuan lalu minta izin. Saat bantuan berjalan, beritahu tindakan berikut dan tunggu respons. Jika pengguna tampak bingung, ulangi dengan susunan lebih sederhana, tunjukkan alat, atau gunakan contoh gerak yang aman. Jangan memaksa pengulangan yang membuatnya malu.\n\nSetelah bantuan, tanyakan kenyamanan dan letakkan bel panggil dalam jangkauan. Umpan balik pengguna membantu penyesuaian berikutnya. Dalam bahasa Jepang, gunakan ungkapan sopan yang jelas; kalimat terlalu panjang atau keigo rumit dapat mengurangi pemahaman.",
       "Jelaskan satu tujuan sebelum mulai.|Minta izin dengan pilihan yang sesuai.|Berikan satu arahan, lalu tunggu respons.|Periksa kenyamanan selama dan sesudah bantuan.",
       "Arahan gerak harus sesuai rencana serta teknik yang telah dilatih; kata-kata tidak menggantikan penilaian transfer.",
       "Petugas mengatakan 'pegang, berdiri, putar, duduk' dalam satu kalimat. Pengguna berhenti dan tampak takut.",
       "Berhenti, tenangkan pengguna, lalu ulangi satu tahap sesuai rencana dengan konfirmasi. Bila bantuan belum aman, minta petugas terlatih mendampingi.",
       "説明~せつめい~penjelasan|声かけ~こえかけ~sapaan atau arahan verbal|介助~かいじょ~bantuan|手順~てじゅん~urutan tindakan",
       "これから椅子に移ります。~これからいすにうつります。~Sekarang kita akan berpindah ke kursi.|痛かったら教えてください。~いたかったらおしえてください。~Beri tahu saya bila terasa sakit.",
       [q("Cara memberi arahan yang mudah diikuti adalah…", "Beberapa perintah cepat|Satu tahap lalu menunggu respons|Hanya berbicara kepada rekan", 1, "Pengguna perlu waktu memahami dan merespons satu tahap."),
        q("Sebelum bantuan tubuh, petugas perlu…", "Menjelaskan dan meminta persetujuan|Langsung menyentuh|Menjanjikan pasti tidak sakit", 0, "Penjelasan dan persetujuan membangun rasa aman dan menghormati pengguna.")])
lesson("kaigo-12", "Hambatan bahasa, pendengaran, dan komunikasi",
       "Menyesuaikan komunikasi tanpa menebak|Menggunakan klarifikasi dan bantuan tim",
       "Hambatan komunikasi dapat berasal dari bahasa, pendengaran, penglihatan, atau kesulitan menyampaikan kata. Pilih dukungan berdasarkan kebutuhan orang tersebut. Tulisan besar membantu sebagian pengguna, tetapi tidak semua dapat membaca dengan baik. Alat bantu dengar pun tidak menghilangkan semua gangguan, terutama di ruang ramai.\n\nCaregiver yang belajar bahasa Jepang perlu membedakan kata yang tidak dikenal dari instruksi yang tidak jelas. Mengangguk untuk menutupi ketidakpahaman dapat menimbulkan kesalahan. Klarifikasi adalah bagian kerja profesional: minta pengulangan, ulangi isi, atau tunjukkan istilah yang perlu dipastikan. Untuk informasi penting, gunakan penerjemah atau bantuan resmi sesuai kebijakan.",
       "Berbicara dengan wajah terlihat, jarak yang nyaman, dan suara jelas. Jangan berteriak dari belakang. Pada pengguna yang sulit menyampaikan kata, gunakan pilihan ya/tidak atau papan komunikasi yang telah disepakati, sambil tetap memastikan preferensi. Jangan menganggap jawaban singkat berarti tidak ingin dilibatkan.\n\nSaat menerima instruksi, ulangi waktu, nama, jumlah, dan tindakan. Bila masih tidak paham, hentikan sebelum tugas berisiko dimulai. Catat istilah baru untuk dipelajari setelah keselamatan pekerjaan dipastikan.",
       "Identifikasi hambatan serta alat komunikasi yang biasa digunakan.|Kurangi kebisingan dan pastikan wajah terlihat.|Gunakan kalimat pendek, tulisan, atau gambar sesuai kebutuhan.|Konfirmasi isi; minta bantuan bila informasi penting belum jelas.",
       "Jangan mengandalkan tebakan atau terjemahan otomatis saja untuk instruksi klinis penting.",
       "Petugas mendengar instruksi Jepang tentang waktu makan, tetapi tidak yakin apakah sebelum atau sesudah pemeriksaan.",
       "Minta pengulangan dan ulangi urutan kegiatan kepada pemberi instruksi. Jangan memulai berdasarkan kata yang hanya sebagian dipahami.",
       "聞き取り~ききとり~pemahaman saat mendengar|補聴器~ほちょうき~alat bantu dengar|筆談~ひつだん~komunikasi tertulis|失語症~しつごしょう~afasia",
       "もう少しゆっくり話してください。~もうすこしゆっくりはなしてください。~Mohon bicara sedikit lebih perlahan.|食事の前ですか、後ですか。~しょくじのまえですか、あとですか。~Apakah sebelum atau sesudah makan?",
       [q("Instruksi penting belum dipahami. Tindakan yang tepat…", "Mengangguk lalu menebak|Mengabaikan rincian waktu|Meminta klarifikasi sebelum bertindak", 2, "Mengonfirmasi mencegah kesalahan yang timbul dari pemahaman sebagian."),
        q("Pengguna kurang mendengar paling terbantu oleh…", "Wajah terlihat dan kebisingan berkurang|Teriakan dari belakang|Bicara lebih cepat", 0, "Konteks visual dan lingkungan yang tenang dapat membantu komunikasi.")])

module("safety", "Keselamatan, infeksi, dan keadaan darurat",
       "Mengenali risiko, menjaga kebersihan, dan mengeskalasi keadaan yang berubah.")
lesson("kaigo-13", "Pencegahan infeksi dan kebersihan tangan",
       "Menerapkan kebersihan tangan pada momen yang tepat|Membedakan sarung tangan dengan kebersihan tangan",
       "Pencegahan infeksi berlaku pada setiap pengguna, bukan hanya orang yang sudah diketahui sakit. Risiko dinilai dari kegiatan dan kemungkinan kontak dengan cairan tubuh, kulit yang tidak utuh, atau benda terkontaminasi. Kebersihan tangan dan alat pelindung yang sesuai memutus perpindahan kotoran di antara kegiatan.\n\nSarung tangan tidak menggantikan kebersihan tangan. Bersihkan tangan sesuai metode fasilitas sebelum dan setelah kontak yang memerlukannya serta setelah melepas sarung tangan. Tangan yang tampak kotor perlu dicuci dengan sabun dan air. Aturan tambahan untuk kondisi infeksi tertentu mengikuti petugas penanggung jawab; jangan menentukan jenis isolasi sendiri.",
       "Pisahkan alur bersih dan kotor: persiapkan alat bersih sebelum memulai, letakkan barang terpakai di tempat yang ditentukan, dan jangan menyentuh gagang pintu atau telepon dengan sarung tangan kotor. Ganti sarung tangan antara pengguna dan ketika berpindah dari kegiatan kotor ke kegiatan bersih.\n\nPilih alat pelindung berdasarkan penilaian tugas serta pelatihan. Gunakan urutan memakai dan melepas yang diajarkan agar tidak mencemari diri. Ikuti produk, pengenceran, dan waktu kontak pembersihan yang disetujui fasilitas; jangan mencampur bahan pembersih.",
       "Siapkan alat bersih dan tempat barang kotor.|Lakukan kebersihan tangan serta gunakan pelindung sesuai tugas.|Pisahkan alur bersih dan kotor selama bantuan.|Lepas pelindung dengan benar, lalu bersihkan tangan dan alat.",
       "Ikuti SOP infeksi fasilitas dan jangan memakai sarung tangan yang sama untuk dua pengguna.",
       "Setelah bantuan toilet, petugas masih memakai sarung tangan saat menyiapkan sendok makan pengguna lain.",
       "Hentikan alur yang mencemari. Lepas sarung tangan, bersihkan tangan, dan tangani peralatan yang mungkin terkontaminasi menurut SOP sebelum menyiapkan bantuan bersih.",
       "感染~かんせん~infeksi|手洗い~てあらい~mencuci tangan|手袋~てぶくろ~sarung tangan|消毒~しょうどく~disinfeksi",
       "手を洗ってからお手伝いします。~てをあらってからおてつだいします。~Saya mencuci tangan dahulu sebelum membantu.|手袋を交換します。~てぶくろをこうかんします。~Saya akan mengganti sarung tangan.",
       [q("Sarung tangan…", "Menggantikan cuci tangan|Boleh dipakai antar pengguna|Tetap memerlukan kebersihan tangan", 2, "Sarung tangan dapat terkontaminasi dan tangan perlu dibersihkan pada momen yang tepat."),
        q("Setelah bantuan toilet sebelum menyiapkan makanan, petugas harus…", "Melepas pelindung dan melakukan kebersihan tangan|Langsung menggunakan sarung tangan sama|Hanya mengelap sarung tangan", 0, "Alur kotor dan bersih perlu dipisahkan untuk mencegah kontaminasi.")])
lesson("kaigo-14", "Risiko jatuh, lingkungan, dan near miss",
       "Mengidentifikasi bahaya yang dapat diubah|Melaporkan kejadian hampir celaka tanpa menyalahkan",
       "Risiko jatuh dipengaruhi pengguna, aktivitas, dan lingkungan. Lantai basah, kabel, pencahayaan buruk, alas kaki tidak sesuai, serta alat yang jauh dari jangkauan dapat memperbesar risiko. Kebutuhan toilet mendesak atau perubahan kondisi juga perlu diperhatikan. Satu penilaian lama tidak cukup bila keadaan berubah.\n\nNear miss atau hiyari-hatto adalah kejadian hampir celaka yang belum menimbulkan cedera. Pelaporan membantu tim memperbaiki sistem sebelum kecelakaan terjadi. Tujuan utamanya memahami penyebab dan pencegahan, bukan mencari siapa yang dapat disalahkan. Pengamatan perlu objektif: apa yang terjadi, di mana, kapan, dan apa kondisi saat itu.",
       "Periksa jalur berjalan, rem alat bila relevan, bel panggil, serta posisi barang sebelum bantuan. Pastikan alat bantu yang digunakan sesuai rencana dan pengguna tahu cara meminta bantuan. Jangan menghilangkan kesempatan bergerak hanya karena petugas merasa takut; diskusikan cara yang aman dengan tim.\n\nJika pengguna jatuh, jangan langsung mengangkat atau menariknya. Panggil bantuan, amankan lingkungan, dan ikuti prosedur penilaian serta respons fasilitas. Catat fakta setelah kebutuhan segera ditangani. Perubahan berulang pada waktu atau tempat tertentu dapat menunjukkan perlunya penyesuaian rencana.",
       "Periksa jalur, cahaya, lantai, dan barang dalam jangkauan.|Cocokkan aktivitas dengan penilaian terbaru.|Laporkan bahaya serta kejadian hampir celaka.|Evaluasi perbaikan bersama tim dan dokumentasikan hasil.",
       "Setelah jatuh, jangan memindahkan pengguna secara otomatis sebelum respons dan penilaian sesuai prosedur.",
       "Pengguna hampir tersandung kabel charger di samping tempat tidur. Tidak ada cedera.",
       "Amankan kabel dengan cara yang disetujui, pastikan pengguna nyaman, lalu laporkan near miss beserta lokasi dan konteks. Tidak adanya cedera bukan alasan untuk melewati pelaporan.",
       "転倒~てんとう~jatuh|危険~きけん~bahaya|予防~よぼう~pencegahan|事故~じこ~kecelakaan",
       "床が濡れています。~ゆかがぬれています。~Lantainya basah.|歩く前に確認しましょう。~あるくまえにかくにんしましょう。~Mari kita periksa sebelum berjalan.",
       [q("Kejadian hampir tersandung tanpa cedera perlu…", "Disembunyikan|Dilaporkan untuk pencegahan|Diabaikan karena belum cedera", 1, "Near miss menyediakan informasi untuk memperbaiki keselamatan."),
        q("Pengguna sudah jatuh. Langkah yang tepat adalah…", "Panggil bantuan dan ikuti prosedur penilaian|Langsung tarik agar berdiri|Suruh berjalan untuk menguji", 0, "Keselamatan dan penilaian sesuai prosedur didahulukan sebelum pemindahan.")])
lesson("kaigo-15", "Perubahan mendadak, panggilan bantuan, dan evakuasi",
       "Memprioritaskan keselamatan dan eskalasi|Menyampaikan lokasi serta kondisi secara ringkas",
       "Saat terjadi keadaan mendadak, petugas perlu mengenali bahaya, memanggil bantuan, dan mengikuti prosedur yang telah dilatih. Kesulitan bernapas, tidak responsif, atau perubahan fungsi yang tiba-tiba harus segera dieskalasi. Jangan menunggu diagnosis sendiri. Di tempat kerja, ketahui tombol panggil, jalur kontak, alamat, dan penanggung jawab tiap shift.\n\nPanggilan bantuan yang baik menyebut lokasi, identitas pengguna sesuai sistem, kondisi teramati, waktu kejadian, serta bantuan yang dibutuhkan. Jika ada bahaya di lingkungan, sampaikan juga. Penanganan kegawatdaruratan seperti resusitasi memerlukan pelatihan langsung dan instruksi yang berlaku; aplikasi ini tidak menggantikan pelatihan tersebut.",
       "Dalam kebakaran atau bencana, ikuti rencana evakuasi fasilitas dan arahan petugas. Kenali rute aman, titik berkumpul, daftar pengguna, serta kebutuhan alat bantu sebelum kejadian. Hindari tindakan spontan yang dapat menambah bahaya, misalnya menggunakan rute yang belum dinyatakan aman.\n\nSetelah kejadian, serahkan informasi kepada tim penerima dan catat urutan faktual. Jangan meninggalkan pengguna berisiko tanpa pengawasan untuk mencari dokumen; minta bantuan rekan dan atur pembagian tugas. Tinjau latihan berkala agar respons tidak hanya bergantung pada hafalan.",
       "Pastikan keselamatan diri, pengguna, dan lingkungan.|Panggil bantuan segera melalui jalur yang ditetapkan.|Sampaikan lokasi, kondisi, waktu, dan kebutuhan bantuan.|Ikuti arahan serta tindakan yang telah dilatih, lalu serah terima.",
       "Jangan memberikan makanan, minuman, atau obat kepada orang yang tidak responsif.",
       "Saat kegiatan pagi, Pak Ono tiba-tiba sulit merespons dan tampak kesulitan bernapas.",
       "Hentikan kegiatan, panggil bantuan segera dengan lokasi yang jelas, amankan pengguna, dan ikuti instruksi serta prosedur darurat. Dokumentasi dilakukan setelah respons segera terorganisasi.",
       "緊急~きんきゅう~darurat|避難~ひなん~evakuasi|応援~おうえん~bantuan tambahan|安全~あんぜん~keselamatan",
       "すぐに来てください。~すぐにきてください。~Mohon segera datang.|二階の食堂です。呼吸が苦しそうです。~にかいのしょくどうです。こきゅうがくるしそうです。~Di ruang makan lantai dua. Tampak sulit bernapas.",
       [q("Saat keadaan darurat, catatan panjang sebaiknya…", "Diselesaikan sebelum memanggil|Menggantikan panggilan bantuan|Dikerjakan setelah respons segera ditangani", 2, "Eskalasi dan keselamatan tidak boleh tertunda oleh administrasi."),
        q("Panggilan bantuan harus mencakup…", "Hanya 'tolong'|Lokasi dan kondisi yang diamati|Diagnosis yang ditebak", 1, "Lokasi serta kondisi membantu tim memberi respons yang tepat.")])

module("mobility", "Posisi tubuh, transfer, dan mobilitas",
       "Menilai kesiapan, menggunakan alat sesuai pelatihan, dan mempertahankan aktivitas.")
lesson("kaigo-16", "Body mechanics dan penanganan yang aman",
       "Menjelaskan prinsip pusat gravitasi dan tumpuan|Memilih bantuan berdasarkan penilaian, bukan kekuatan petugas",
       "Body mechanics memanfaatkan posisi dan gerak tubuh agar kerja lebih stabil. Tumpuan yang cukup, kedekatan dengan aktivitas, serta menghindari memutar punggung sambil membawa beban membantu mengurangi gerakan yang tidak efisien. Namun teknik tubuh saja tidak menjadikan semua pemindahan pengguna aman. Beban, kemampuan pengguna, alat, ruang, dan jumlah petugas tetap menentukan cara bantuan.\n\nPenanganan pengguna berbeda dari mengangkat barang. Orang dapat bergerak tiba-tiba, merasa nyeri, atau kehilangan keseimbangan. Ikuti penilaian transfer dan kebijakan penggunaan alat. Pengguna yang tidak dapat menopang tubuh tidak boleh diangkat sendirian hanya karena petugas merasa kuat.",
       "Sebelum bantuan, baca rencana dan cek perubahan kondisi. Pastikan alat tersedia, berfungsi, dan sesuai pengguna. Penggunaan hoist, sling, atau alat transfer memerlukan pelatihan khusus serta pemeriksaan kecocokan alat. Jumlah petugas mengikuti penilaian dan instruksi alat, bukan asumsi bahwa dua orang selalu cukup.\n\nJelaskan cara pengguna dapat ikut serta dan koordinasikan satu pemimpin arahan. Bila situasi tidak sesuai rencana, hentikan dan minta penilaian ulang. Merasa terburu-buru atau kekurangan petugas bukan alasan melewati pemeriksaan.",
       "Cek kemampuan pengguna dan rencana penanganan terbaru.|Siapkan ruang, alat, serta jumlah petugas yang ditetapkan.|Jelaskan tahap dan koordinasikan arahan.|Hentikan bila kemampuan atau kondisi tidak sesuai penilaian.",
       "Jangan mencoba teknik angkat atau hoist pertama kali hanya dari bacaan; praktikkan dengan instruktur.",
       "Pak Endo yang biasanya membantu berdiri hari ini tidak mampu menumpu. Petugas hendak tetap memakai cara kemarin.",
       "Hentikan transfer yang tidak sesuai kemampuan hari ini dan minta penilaian tim. Siapkan bantuan serta alat berdasarkan rencana yang diperbarui.",
       "重心~じゅうしん~pusat gravitasi|支持基底面~しじきていめん~bidang tumpuan|移乗~いじょう~transfer|福祉用具~ふくしようぐ~alat bantu kesejahteraan",
       "今日は立てそうですか。~きょうはたてそうですか。~Apakah hari ini tampaknya bisa berdiri?|二人で確認してから行います。~ふたりでかくにんしてからおこないます。~Kami akan melakukannya setelah memeriksa bersama.",
       [q("Body mechanics yang baik…", "Menjamin semua angkat aman|Menggantikan alat bantu|Tetap memerlukan penilaian dan alat sesuai kebutuhan", 2, "Kemampuan pengguna, alat, ruang, serta petugas harus dipertimbangkan bersama."),
        q("Kemampuan menumpu berubah hari ini. Petugas perlu…", "Memakai cara kemarin tanpa perubahan|Menghentikan dan meminta penilaian ulang|Menarik pengguna lebih kuat", 1, "Rencana bantuan harus sesuai kondisi aktual.")])
lesson("kaigo-17", "Posisi tidur, kenyamanan, dan perlindungan kulit",
       "Mengenali posisi dasar dan kebutuhan dukungan|Mengamati kulit serta melaporkan perubahan",
       "Posisi telentang, miring, dan duduk memiliki kegunaan serta risiko yang berbeda. Posisi yang nyaman harus memperhatikan kemampuan bergerak, nyeri, kondisi pernapasan, dan rencana dukungan pengguna. Bantal atau alat penopang digunakan sesuai penilaian; tujuan bukan sekadar membuat tubuh tampak rapi.\n\nTekanan berkepanjangan, gesekan, kelembapan, dan gaya geser dapat merusak kulit. Orang yang sulit mengubah posisi memerlukan penilaian risiko serta rencana individual. Tidak ada satu jadwal perubahan posisi yang otomatis cocok untuk semua pengguna. Ikuti rencana dan laporkan bila tidak dapat dilaksanakan atau pengguna merasa tidak nyaman.",
       "Saat membantu posisi, jaga privasi dan gunakan metode yang telah dilatih untuk mengurangi gesekan. Jangan menyeret tubuh dengan menarik lengan atau pakaian. Pastikan lipatan seprai, alat, dan bagian tubuh tidak menimbulkan tekanan yang tidak diinginkan. Bel panggil tetap dalam jangkauan setelah selesai.\n\nAmati kulit ketika kegiatan perawatan memungkinkan, dengan persetujuan. Kemerahan baru, luka, kelembapan berlebih, atau keluhan nyeri perlu dilaporkan kepada perawat. Jangan memijat kemerahan atau mengoleskan produk baru tanpa arahan. Catat lokasi dan waktu temuan secara objektif.",
       "Periksa rencana posisi serta keluhan pengguna.|Siapkan alat penopang dan bantuan yang diperlukan.|Ubah posisi dengan teknik terlatih serta minim gesekan.|Periksa kenyamanan, kulit yang dapat diamati, dan jangkauan bel.",
       "Jangan menentukan jadwal reposisi atau penanganan luka sendiri; ikuti penilaian klinis dan rencana individual.",
       "Saat merapikan pakaian, petugas melihat kemerahan baru di area tumit pengguna yang jarang bergerak.",
       "Laporkan lokasi, waktu, dan keluhan kepada perawat, lalu ikuti instruksi perlindungan tekanan. Jangan langsung memijat area atau mengoleskan krim tanpa arahan.",
       "仰臥位~ぎょうがい~posisi telentang|側臥位~そくがい~posisi miring|体位変換~たいいへんかん~perubahan posisi|褥瘡~じょくそう~luka akibat tekanan",
       "この姿勢は楽ですか。~このしせいはらくですか。~Apakah posisi ini nyaman?|かかとに赤いところがあります。~かかとにあかいところがあります。~Ada bagian kemerahan pada tumit.",
       [q("Kemerahan kulit baru sebaiknya…", "Dipijat kuat|Dilaporkan dan ditangani sesuai arahan|Ditutupi tanpa catatan", 1, "Perubahan kulit memerlukan penilaian serta perlindungan yang sesuai."),
        q("Jadwal perubahan posisi harus…", "Mengikuti rencana individual|Sama untuk semua orang|Ditentukan dari kebiasaan petugas saja", 0, "Risiko, kenyamanan, serta kondisi pengguna berbeda-beda.")])
lesson("kaigo-18", "Kursi roda, berjalan, dan keluar ruangan",
       "Memeriksa kesiapan alat dan jalur|Menjaga aktivitas tanpa mengabaikan perubahan kondisi",
       "Kursi roda dan alat jalan mendukung partisipasi, tetapi perlu ukuran, cara penggunaan, dan rencana yang sesuai. Periksa bagian yang relevan seperti rem, dudukan, pijakan kaki, serta jalur sebelum digunakan. Tidak semua kursi atau alat memiliki cara operasi yang sama; ikuti pelatihan dan instruksi perangkat.\n\nBantuan berjalan mempertimbangkan keseimbangan, kelelahan, nyeri, serta kemampuan mengikuti arahan. Pengguna yang biasanya mampu berjalan dapat membutuhkan bantuan berbeda pada hari tertentu. Aktivitas bukan ujian keberanian; beri kesempatan beristirahat dan minta penilaian bila muncul perubahan.",
       "Sebelum transfer ke atau dari kursi roda, tempatkan alat dan atur rem serta pijakan sesuai prosedur. Jangan membiarkan pengguna berdiri di atas pijakan kaki. Transfer dilakukan dengan cara yang telah ditetapkan, bukan menarik lengan. Setelah duduk, periksa kenyamanan, posisi kaki, dan kebutuhan penopang.\n\nUntuk perjalanan, rencanakan jalur, cuaca, akses toilet, pendampingan, dan titik istirahat. Melewati tanjakan, ambang pintu, atau jalan tidak rata memerlukan teknik alat yang telah dilatih. Jelaskan perubahan gerakan sebelum dilakukan agar pengguna tidak terkejut.",
       "Periksa kondisi pengguna, alat, dan jalur.|Atur transfer serta bagian alat sesuai prosedur.|Jelaskan gerakan dan pantau kenyamanan serta kelelahan.|Berhenti bila muncul pusing, nyeri, atau kesulitan baru dan laporkan.",
       "Jangan menuruni lereng atau melewati rintangan dengan teknik yang belum dilatih.",
       "Petugas melihat pijakan kaki kursi roda masih menghalangi saat pengguna hendak berdiri.",
       "Hentikan dulu, jelaskan, lalu atur pijakan dan rem sesuai prosedur alat. Mulai kembali hanya setelah transfer siap dan bantuan sesuai rencana.",
       "車椅子~くるまいす~kursi roda|歩行~ほこう~berjalan|杖~つえ~tongkat|段差~だんさ~perbedaan tinggi lantai",
       "ブレーキを確認します。~ぶれーきをかくにんします。~Saya akan memeriksa rem.|疲れたら休みましょう。~つかれたらやすみましょう。~Mari beristirahat bila lelah.",
       [q("Saat transfer, pijakan kaki yang menghalangi harus…", "Dibiarkan agar cepat|Dipakai sebagai tempat berdiri|Diatur sesuai prosedur sebelum mulai", 2, "Persiapan alat mencegah hambatan dan risiko saat transfer."),
        q("Pengguna tiba-tiba pusing saat berjalan. Petugas sebaiknya…", "Memaksa menyelesaikan jarak|Berhenti dengan aman dan melaporkan|Menganggap malas", 1, "Perubahan kondisi perlu perhatian sebelum aktivitas dilanjutkan.")])

module("food", "Makan, minum, dan kesehatan mulut",
       "Menghormati pilihan makan, memahami rencana menelan, dan mengamati respons.")
lesson("kaigo-19", "Persiapan makan dan dukungan kemandirian",
       "Memeriksa identitas, diet, dan alergi sesuai sistem|Memberi bantuan makan yang terarah dan menghormati tempo",
       "Makan memenuhi kebutuhan tubuh sekaligus menjadi kegiatan sosial dan sumber kesenangan. Periksa makanan yang diberikan kepada pengguna sesuai sistem identifikasi fasilitas, diet, alergi, serta rencana bantuan. Jangan mengandalkan lokasi kursi atau tampilan piring untuk memastikan penerima. Pilihan budaya, agama, dan kebiasaan perlu dicatat serta disesuaikan oleh tim.\n\nDukungan kemandirian dimulai dari lingkungan: meja, peralatan, pencahayaan, dan jangkauan. Pengguna mungkin dapat makan sendiri jika makanan dipotong sesuai rencana atau alat tertentu disiapkan. Posisi tubuh, tingkat kesiagaan, dan kemampuan menelan harus sesuai penilaian sebelum bantuan dimulai.",
       "Jelaskan makanan dan minta izin membantu. Ikuti tempo pengguna, jangan menyuapi ketika mulut masih berisi makanan. Amati respons seperti kesulitan, batuk, atau kelelahan dan ikuti jalur pelaporan. Jangan memaksa menghabiskan porsi atau mencampur semua makanan hanya demi efisiensi.\n\nCatat jumlah yang dimakan berdasarkan cara fasilitas, jenis bantuan, penolakan, serta keluhan. Angka konsumsi saja tidak menjelaskan pengalaman makan; tambahkan pengamatan yang dapat membantu penyesuaian rencana. Bila makanan tidak sesuai, konfirmasikan sebelum memberikannya.",
       "Pastikan pengguna, makanan, alergi, dan rencana cocok.|Atur lingkungan serta posisi menurut penilaian.|Jelaskan, beri kesempatan makan sendiri, dan bantu sesuai tempo.|Amati serta catat konsumsi, bantuan, dan respons.",
       "Jangan mengubah tekstur atau diet sendiri saat makanan tampak sulit dikonsumsi.",
       "Dua nampan mirip diletakkan berdampingan. Petugas belum memeriksa label dan rencana diet.",
       "Konfirmasikan penerima serta diet melalui sistem fasilitas sebelum makanan diberikan. Kemiripan menu tidak memastikan keamanan atau kesesuaian.",
       "食事~しょくじ~makan|食欲~しょくよく~nafsu makan|主食~しゅしょく~makanan pokok|副食~ふくしょく~lauk atau pendamping",
       "今日の昼ご飯はこちらです。~きょうのひるごはんはこちらです。~Ini makan siang hari ini.|ご自分で食べてみますか。~ごじぶんでたべてみますか。~Apakah ingin mencoba makan sendiri?",
       [q("Nampan yang mirip perlu…", "Diberikan menurut posisi kursi|Diperiksa penerima dan rencana dietnya|Dibagi secara acak", 1, "Identitas, alergi, serta rencana diet harus cocok sebelum pemberian."),
        q("Tempo bantuan makan sebaiknya…", "Mengikuti kemampuan dan respons pengguna|Selalu cepat|Sama pada semua pengguna", 0, "Kenyamanan serta kemampuan pengguna menentukan bantuan, bukan jadwal petugas saja.")])
lesson("kaigo-20", "Menelan, aspirasi, dan rencana hidrasi",
       "Mengenali kesulitan yang perlu dilaporkan|Mengikuti rencana tekstur dan cairan individual",
       "Menelan memindahkan makanan atau cairan menuju saluran pencernaan. Kesulitan menelan disebut enge shōgai; masuknya bahan ke saluran napas disebut goen atau aspirasi. Batuk ketika makan, suara berubah setelah menelan, atau kesulitan berulang perlu dilaporkan. Tidak adanya batuk juga tidak memastikan bahwa menelan aman. Penilaian menelan dilakukan oleh tenaga berwenang.\n\nRencana dapat mengatur posisi, tekstur makanan, konsistensi cairan, alat, serta tingkat bantuan. Petugas mengikuti rencana individual dan pelatihan. Kekentalan atau jumlah tidak ditentukan dari perkiraan. Memotong makanan lebih kecil pun tidak otomatis cocok untuk setiap kesulitan menelan.",
       "Jika muncul kesulitan baru saat makan, hentikan pemberian, pastikan keselamatan, dan panggil petugas sesuai prosedur. Bila pengguna tidak dapat bernapas atau tidak responsif, gunakan respons darurat yang telah dilatih. Jangan memaksa minum untuk 'mendorong' makanan.\n\nHidrasi juga mengikuti kondisi dan rencana. Tidak ada satu target cairan yang cocok untuk semua pengguna, terutama bila ada pembatasan klinis. Catat asupan menurut sistem fasilitas dan laporkan perubahan kebiasaan minum, mulut kering, atau keluhan baru. Pengamatan menjadi dasar komunikasi, bukan diagnosis dehidrasi sendiri.",
       "Baca rencana menelan dan hidrasi terbaru.|Pastikan posisi, tekstur, alat, dan bantuan sesuai rencana.|Pantau respons selama pemberian dan hentikan bila muncul kesulitan.|Laporkan perubahan dan catat asupan secara akurat.",
       "Jangan menawarkan air, mengubah kekentalan, atau meneruskan makanan saat pengguna mengalami kesulitan menelan baru.",
       "Saat makan, suara pengguna menjadi basah dan ia berulang kali batuk. Teman menyarankan memberi banyak air.",
       "Hentikan pemberian makanan atau minuman, panggil petugas berwenang, dan ikuti prosedur. Laporkan makanan yang diberikan serta respons; jangan menggunakan air sebagai solusi spontan.",
       "嚥下~えんげ~menelan|誤嚥~ごえん~aspirasi|水分~すいぶん~cairan|とろみ~とろみ~kekentalan cairan",
       "飲み込みにくいですか。~のみこみにくいですか。~Apakah sulit menelan?|食事を止めて、看護師を呼びます。~しょくじをとめて、かんごしをよびます。~Saya hentikan makan dan memanggil perawat.",
       [q("Kesulitan baru saat menelan harus direspons dengan…", "Air sebanyak mungkin|Hentikan pemberian dan eskalasi sesuai prosedur|Lanjutkan lebih cepat", 1, "Perubahan menelan memerlukan penilaian dan tidak diselesaikan dengan minuman spontan."),
        q("Tekstur dan target cairan ditetapkan berdasarkan…", "Selera petugas|Aturan yang sama bagi semua lansia|Rencana individual dari tim", 2, "Kebutuhan serta pembatasan tiap pengguna dapat berbeda.")])
lesson("kaigo-21", "Kebersihan mulut dan perawatan gigi palsu",
       "Menyiapkan bantuan mulut sesuai rencana|Mengamati keluhan dan menjaga identitas alat pribadi",
       "Kebersihan mulut mendukung kenyamanan, makan, komunikasi, dan hubungan sosial. Pengguna mungkin membutuhkan bantuan menyikat, membersihkan mulut, atau merawat gigi palsu. Pilihan alat, posisi, serta penggunaan cairan harus mengikuti kemampuan pengguna dan rencana, terutama bila ada kesulitan menelan.\n\nAmati keluhan seperti nyeri, sariawan, perdarahan, mulut kering, atau gigi palsu tidak nyaman. Sampaikan temuan kepada tim; jangan menganggap penolakan makan selalu berkaitan dengan tidak suka menu. Caregiver tidak melakukan penyesuaian gigi palsu atau tindakan gigi sendiri.",
       "Minta izin, siapkan alat pribadi yang sesuai, dan pastikan posisi aman. Berikan kesempatan pengguna melakukan bagian yang mampu dilakukan. Bantuan dilakukan dengan lembut sesuai pelatihan, sambil memantau kenyamanan dan toleransi. Jika muncul kesulitan atau perdarahan baru, hentikan bagian yang bermasalah dan laporkan.\n\nGigi palsu milik tiap pengguna perlu identifikasi, penanganan, penyimpanan, dan produk pembersih sesuai petunjuk serta kebijakan. Jangan menukar wadah atau memakai air sangat panas secara sembarangan. Setelah selesai, catat bantuan serta temuan yang perlu ditindaklanjuti.",
       "Cek rencana, alat pribadi, dan kondisi mulut.|Jelaskan serta minta izin sebelum bantuan.|Bantu dengan lembut sesuai kemampuan dan pelatihan.|Bersihkan serta simpan alat dengan identitas tepat; laporkan keluhan.",
       "Jangan memaksa berkumur atau menggunakan volume cairan tanpa menyesuaikan rencana menelan.",
       "Ibu Noda mengurangi makan dan mengatakan gigi palsunya sakit saat dipakai.",
       "Catat ucapan serta perubahan konsumsi, laporkan kepada tim untuk penilaian, dan ikuti rencana bantuan. Jangan mengikir atau memperbaiki gigi palsu sendiri.",
       "口腔~こうくう~rongga mulut|義歯~ぎし~gigi palsu|歯磨き~はみがき~menyikat gigi|うがい~うがい~berkumur",
       "口の中に痛みはありますか。~くちのなかにいたみはありますか。~Apakah ada nyeri di dalam mulut?|入れ歯の具合はどうですか。~いればのぐあいはどうですか。~Bagaimana kenyamanan gigi palsunya?",
       [q("Pengguna mengatakan gigi palsu sakit. Petugas harus…", "Melaporkan untuk penilaian|Memperbaiki sendiri|Memaksa agar tetap dipakai", 0, "Keluhan perlu penilaian dan tindak lanjut sesuai kewenangan."),
        q("Wadah gigi palsu perlu…", "Dipakai bergantian|Diberi identitas dan disimpan sesuai prosedur|Dicampur agar mudah dibersihkan", 1, "Identifikasi mencegah alat pribadi tertukar dan mendukung kebersihan.")])

module("toileting", "Toilet, eliminasi, dan martabat",
       "Mendukung kebiasaan buang air, menjaga kebersihan, dan mencatat perubahan.")
lesson("kaigo-22", "Bantuan toilet dan kebiasaan eliminasi",
       "Mengenali pola dan kebutuhan bantuan|Menjaga privasi tanpa meninggalkan risiko",
       "Eliminasi mencakup buang air kecil dan besar. Dukungan toilet mempertimbangkan kebiasaan waktu, kemampuan bergerak, pakaian, komunikasi, serta rasa malu. Pengguna yang membutuhkan bantuan tetap berhak mendapat penjelasan dan pilihan. Jangan membuat semua orang mengikuti jadwal yang sama tanpa melihat rencana individual.\n\nKemandirian dapat didukung dengan jalur yang aman, pakaian yang mudah digunakan, penanda toilet, dan alat yang sesuai. Kebutuhan mendesak dapat meningkatkan risiko jika bel panggil jauh atau respons lambat. Tujuan bukan sekadar menghindari kebocoran, tetapi membantu pengguna nyaman dan mempertahankan kebiasaan yang mungkin dilakukan.",
       "Sebelum bantuan, cek rencana transfer, siapkan kebersihan serta alat pelindung, dan jelaskan tahapnya. Setelah pengguna berada pada posisi yang aman, jaga privasi sesuai kebutuhan pengawasan. Tidak semua pengguna aman ditinggalkan sendirian; keputusan mengikuti penilaian. Bel panggil harus dapat dijangkau dan pengguna memahami cara memakainya.\n\nBantu pakaian dan kebersihan sesuai kemampuan. Setelah selesai, pastikan posisi nyaman, alur kotor ditangani, serta tangan dibersihkan. Catat waktu dan jenis bantuan. Keluhan nyeri, kesulitan baru, atau perubahan pola perlu disampaikan kepada tim.",
       "Tanyakan kebutuhan dan cek pola serta rencana toilet.|Siapkan jalur, alat, transfer, dan kebersihan.|Jaga privasi dengan tingkat pengawasan sesuai penilaian.|Bantu kebersihan, rapikan dengan izin, dan catat respons.",
       "Jangan meninggalkan pengguna berisiko di toilet hanya untuk memberi privasi.",
       "Pak Hara perlu diawasi saat duduk dan berdiri. Petugas ingin keluar jauh dari toilet demi privasi.",
       "Ikuti tingkat pengawasan yang ditetapkan dan cari cara menjaga privasi tanpa meninggalkan risiko. Jelaskan kepada pengguna serta pastikan bantuan dapat diberikan segera.",
       "排泄~はいせつ~eliminasi|排尿~はいにょう~buang air kecil|排便~はいべん~buang air besar|便座~べんざ~dudukan toilet",
       "お手洗いに行きますか。~おてあらいにいきますか。~Apakah ingin ke toilet?|終わったら呼んでください。~おわったらよんでください。~Panggil saya setelah selesai.",
       [q("Privasi saat toilet harus…", "Selalu membuat pengguna ditinggal|Diseimbangkan dengan kebutuhan pengawasan|Diabaikan bila petugas sibuk", 1, "Privasi penting, tetapi bantuan tetap mengikuti penilaian risiko individual."),
        q("Jadwal dukungan toilet sebaiknya mengikuti…", "Kebiasaan dan rencana pengguna|Satu jadwal untuk semua|Kenyamanan petugas saja", 0, "Pola serta kebutuhan individual menjadi dasar dukungan.")])
lesson("kaigo-23", "Inkontinensia, pad, dan kebersihan kulit",
       "Menghindari stigma pada kebocoran|Mengelola alat penyerap dan alur bersih-kotor sesuai rencana",
       "Inkontinensia adalah keluarnya urine atau feses tanpa kendali yang diinginkan. Kejadian ini dapat membuat pengguna malu dan khawatir berinteraksi. Respons caregiver perlu tenang, privat, dan tanpa menyalahkan. Jangan menyebut pengguna 'jorok' atau membicarakan kebocoran di depan orang lain.\n\nAlat penyerap dapat membantu, tetapi pemilihan jenis, ukuran, frekuensi pemeriksaan, serta penggunaan mengikuti rencana. Alat tidak otomatis menggantikan kesempatan menggunakan toilet. Tim perlu memahami pola kebocoran serta hambatan yang dapat dibantu, bukan hanya mengganti produk setiap kali kejadian.",
       "Siapkan perlengkapan sebelum mulai agar pengguna tidak dibiarkan terbuka. Jaga kenyamanan dan tutupi bagian tubuh yang tidak sedang dibantu. Pisahkan alat bersih serta kotor dan ikuti teknik kebersihan yang telah dilatih. Kulit dikeringkan dengan lembut; penggunaan pelindung kulit sesuai arahan.\n\nPeriksa keluhan, kemerahan, atau perubahan kulit yang terlihat selama bantuan dan laporkan temuan baru. Setelah penggantian, pastikan alat tidak terlalu ketat, pakaian nyaman, dan posisi sesuai rencana. Catat waktu, kebocoran, serta bantuan yang diberikan; jangan mengurangi cairan secara mandiri untuk mencegah kebocoran.",
       "Jelaskan bantuan dengan suara serta tempat yang menjaga privasi.|Siapkan alat sesuai rencana dan pisahkan bersih-kotor.|Bantu kebersihan serta penggantian sesuai pelatihan.|Periksa kenyamanan, kulit, dan pola; laporkan perubahan.",
       "Jangan mengurangi minum, menambah lapisan pad, atau memakai produk kulit baru tanpa rencana tim.",
       "Ibu Kondo mengalami kebocoran dan meminta petugas tidak berbicara keras tentangnya.",
       "Hormati permintaan, beri bantuan privat dengan tenang, dan catat lewat sistem resmi. Tinjau pola bersama tim tanpa mempermalukan atau mengubah asupan cairan sendiri.",
       "失禁~しっきん~inkontinensia|尿~にょう~urine|皮膚~ひふ~kulit|清潔~せいけつ~kebersihan",
       "着替えをお持ちします。~きがえをおもちします。~Saya akan membawakan pakaian ganti.|皮膚がかゆいですか。~ひふがかゆいですか。~Apakah kulit terasa gatal?",
       [q("Mengurangi cairan sendiri agar tidak bocor adalah…", "Selalu benar|Cara utama caregiver|Tidak tepat tanpa rencana dan penilaian tim", 2, "Kebutuhan cairan tidak ditentukan dari kenyamanan penggantian alat."),
        q("Penggunaan pad sebaiknya…", "Mengikuti rencana serta tetap mempertimbangkan toilet|Selalu mengganti semua kesempatan toilet|Ditambah lapisan sebanyak mungkin", 0, "Dukungan perlu sesuai kebutuhan, pola, dan kenyamanan pengguna.")])
lesson("kaigo-24", "Catatan urine, feses, dan perubahan pola",
       "Menulis pengamatan eliminasi secara objektif|Melaporkan nyeri atau perubahan tanpa mendiagnosis",
       "Catatan eliminasi membantu tim memahami pola serta perubahan. Informasi dapat mencakup waktu, frekuensi, jumlah bila diukur menurut prosedur, bentuk atau warna yang diamati, keluhan, dan jenis bantuan. Bedakan pengamatan langsung dari ucapan pengguna. Jika jumlah tidak diukur, jangan menuliskan angka seolah pasti.\n\nSembelit, diare, atau kesulitan berkemih memerlukan penilaian dan rencana; caregiver tidak menentukan obat atau diagnosis. Perubahan baru, darah yang terlihat, nyeri, atau keadaan pengguna memburuk perlu dilaporkan menurut urgensi prosedur. Jangan menunggu catatan beberapa hari jika ada tanda yang memerlukan respons segera.",
       "Gunakan istilah atau skala yang dipakai fasilitas setelah pelatihan. Catatan 'BAB buruk' kurang membantu dibanding keterangan waktu, bentuk yang diamati, serta keluhan. Catat juga apakah pengguna menolak bantuan atau tidak dapat mencapai toilet tepat waktu.\n\nDukungan aktivitas, diet, dan cairan mengikuti rencana individual. Jangan memberikan obat pencahar, mengubah obat, atau melakukan tindakan invasif karena hasil catatan tampak tidak biasa. Bila pengguna malu membicarakan eliminasi, jelaskan tujuan pengamatan serta lakukan percakapan secara privat.",
       "Amati dengan persetujuan dan sesuai kebutuhan pelayanan.|Catat waktu, fakta, ukuran yang benar-benar diukur, serta keluhan.|Bandingkan dengan pola yang diketahui.|Laporkan perubahan dan ikuti instruksi tanpa membuat terapi sendiri.",
       "Darah yang terlihat atau keluhan baru perlu disampaikan; jangan menetapkan penyebabnya sendiri.",
       "Pak Ueda berkata buang air kecil terasa sakit. Catatan sebelumnya tidak mencantumkan keluhan ini.",
       "Catat ucapan, waktu, serta pengamatan yang relevan dan laporkan kepada perawat. Jangan menyimpulkan infeksi atau memberi obat dari persediaan lain.",
       "便秘~べんぴ~sembelit|下痢~げり~diare|記録~きろく~catatan|量~りょう~jumlah",
       "いつ排便がありましたか。~いつはいべんがありましたか。~Kapan terakhir buang air besar?|排尿のときに痛いと話しています。~はいにょうのときにいたいとはなしています。~Beliau mengatakan sakit ketika buang air kecil.",
       [q("Jumlah urine tidak diukur. Catatan sebaiknya…", "Menebak angka|Tidak mencantumkan angka sebagai hasil pasti|Menggunakan angka pengguna lain", 1, "Catatan harus membedakan ukuran nyata dari perkiraan."),
        q("Nyeri baru saat berkemih perlu…", "Diobati sendiri|Diabaikan bila malu|Dicatat dan dilaporkan", 2, "Penilaian serta keputusan terapi berada pada petugas berwenang.")])

module("hygiene", "Mandi, pakaian, dan perawatan diri",
       "Merencanakan bantuan kebersihan yang nyaman, privat, dan sesuai kemampuan.")
lesson("kaigo-25", "Persiapan mandi dan keselamatan kamar mandi",
       "Memeriksa kesiapan pengguna dan lingkungan|Mendukung preferensi serta privasi saat mandi",
       "Mandi mendukung kebersihan, kenyamanan, dan rutinitas. Sebelum mulai, cek rencana, kondisi hari itu, preferensi, serta bantuan yang diperlukan. Pengguna dapat merasa tidak nyaman karena dingin, takut jatuh, malu, atau pengalaman sebelumnya. Penjelasan dan persetujuan membantu menentukan cara dukungan yang sesuai.\n\nKamar mandi memiliki risiko lantai basah, perubahan suhu, ruang terbatas, dan perpindahan posisi. Periksa alat, jalur, pegangan, serta kebutuhan petugas berdasarkan prosedur. Suhu air dan ruangan diperiksa sesuai SOP dan dikonfirmasi kepada pengguna; jangan hanya mengandalkan rasa tangan petugas atau satu angka tanpa konteks.",
       "Siapkan pakaian, handuk, perlengkapan, dan rencana transfer sebelum membuka pakaian. Tutupi tubuh sejauh mungkin dan berikan kesempatan pengguna membersihkan bagian yang mampu dijangkau. Pantau respons selama kegiatan, termasuk keluhan pusing, nyeri, atau kelelahan. Bila kondisi berubah, hentikan dan ikuti jalur pelaporan.\n\nSetelah mandi, keringkan dengan lembut sesuai rencana, bantu berpakaian, dan pastikan pengguna nyaman. Amati kulit ketika sesuai kebutuhan dan persetujuan. Laporkan temuan baru serta catat bantuan dan respons; jadwal mandi yang selesai tidak cukup menggambarkan keamanan kegiatan.",
       "Konfirmasi persetujuan, kondisi, serta rencana mandi.|Siapkan perlengkapan, alat, petugas, dan lingkungan.|Periksa air menurut SOP, lalu bantu dengan menjaga privasi.|Pantau respons, bantu selesai dengan nyaman, dan dokumentasikan.",
       "Pengguna yang memerlukan pengawasan tidak boleh ditinggalkan sendiri di area mandi.",
       "Ibu Hayashi tampak pusing sebelum mandi yang sudah dijadwalkan.",
       "Jangan memaksakan jadwal. Pastikan aman, laporkan perubahan, dan tunggu arahan tentang penyesuaian kegiatan.",
       "入浴~にゅうよく~mandi|浴室~よくしつ~kamar mandi|湯温~ゆおん~suhu air mandi|更衣~こうい~berganti pakaian",
       "お湯の温度はいかがですか。~おゆのおんどはいかがですか。~Bagaimana suhu airnya?|めまいはありませんか。~めまいはありませんか。~Apakah ada rasa pusing?",
       [q("Pusing sebelum mandi perlu…", "Dilaporkan sebelum kegiatan dilanjutkan|Diabaikan karena jadwal|Diatasi dengan air lebih panas", 0, "Kondisi aktual harus diperiksa sebelum melanjutkan aktivitas."),
        q("Persiapan mandi meliputi…", "Handuk saja|Lingkungan, alat, rencana, dan kondisi pengguna|Hanya menyalakan air", 1, "Risiko berasal dari pengguna, alat, dan lingkungan sekaligus.")])
lesson("kaigo-26", "Seishiki, rambut, dan perawatan kebersihan",
       "Mengenali alternatif kebersihan saat mandi tidak sesuai|Mengamati kulit dengan menjaga kenyamanan",
       "Seishiki (清拭) adalah membersihkan tubuh dengan menyeka. Cara ini dapat menjadi bagian rencana ketika mandi penuh tidak sesuai, tetapi keputusan mengikuti kondisi serta arahan tim. Bantuan kebersihan tetap mencakup persetujuan, privasi, kenyamanan, serta kesempatan melakukan bagian yang mampu dikerjakan pengguna.\n\nSiapkan perlengkapan sesuai SOP dan pertahankan pemisahan alat bersih serta kotor. Membuka seluruh tubuh lebih lama tidak diperlukan untuk membersihkan satu area. Perhatikan keluhan dingin, nyeri, serta keadaan kulit. Produk yang digunakan perlu sesuai rencana, terutama bila ada sensitivitas atau masalah kulit.",
       "Kerjakan bertahap dengan teknik yang telah dilatih, lalu keringkan area dengan lembut. Jangan menggosok kuat untuk menghilangkan semua noda. Pada rambut, sesuaikan cara mencuci atau menyisir dengan posisi, toleransi, dan preferensi. Alat pribadi tidak dipakai bergantian tanpa pengelolaan yang disetujui.\n\nPerawatan kuku atau area yang berisiko membutuhkan aturan khusus fasilitas dan penilaian kondisi pengguna. Jangan memotong area yang tidak dapat dinilai dengan baik, menangani luka, atau melakukan tindakan di luar pelatihan. Laporkan perubahan kulit, kuku, atau keluhan kepada petugas terkait.",
       "Cek rencana alternatif kebersihan dan preferensi.|Siapkan alat, suhu lingkungan, serta pelindung sesuai tugas.|Bantu satu area dengan privasi dan teknik lembut.|Keringkan, rapikan, lalu laporkan temuan baru.",
       "Perawatan kuku, luka, dan kulit bermasalah hanya sesuai kewenangan serta arahan; jangan mencoba tindakan baru sendiri.",
       "Pengguna tidak menjalani mandi hari ini menurut rencana, tetapi merasa tidak nyaman karena berkeringat.",
       "Konfirmasikan bentuk bantuan kebersihan yang diizinkan, jelaskan, dan bantu bertahap sambil menjaga tubuh tetap hangat serta privat.",
       "清拭~せいしき~membersihkan tubuh dengan menyeka|洗髪~せんぱつ~mencuci rambut|整容~せいよう~merapikan penampilan|保湿~ほしつ~menjaga kelembapan kulit",
       "体を拭いてもよろしいですか。~からだをふいてもよろしいですか。~Bolehkah saya membantu menyeka tubuh?|寒くないですか。~さむくないですか。~Apakah tidak kedinginan?",
       [q("Saat menyeka tubuh, privasi dijaga dengan…", "Membuka seluruh tubuh lama|Menutup area yang tidak sedang dibantu|Membiarkan pintu terbuka", 1, "Bantuan bertahap menjaga kenyamanan serta martabat."),
        q("Temuan kulit baru sebaiknya…", "Digosok sampai hilang|Ditangani dengan produk apa pun|Dilaporkan untuk penilaian", 2, "Penanganan kulit perlu sesuai kondisi, arahan, dan kewenangan.")])
lesson("kaigo-27", "Berpakaian, pilihan penampilan, dan bantuan sebagian",
       "Memberi pilihan pakaian yang nyata|Menyesuaikan urutan bantuan dengan kemampuan",
       "Pakaian berkaitan dengan identitas, kenyamanan, budaya, suhu, serta kegiatan yang akan dilakukan. Pengguna dapat mempunyai preferensi warna atau model yang penting baginya. Pilihan sederhana yang dapat dihormati membantu partisipasi. Kebutuhan bantuan fisik tidak menghilangkan hak menentukan penampilan.\n\nPeriksa kemampuan tangan, keseimbangan saat duduk, nyeri, serta bagian pakaian yang sulit digunakan. Dukungan dapat berupa menata pakaian, memberi arahan, membantu kancing, atau membantu sebagian gerakan. Pada kelemahan satu sisi, urutan umum memasukkan sisi yang membutuhkan bantuan dahulu disesuaikan dengan rencana dan teknik yang telah dilatih.",
       "Siapkan pakaian sebelum mengganti, tutup pintu atau tirai, dan jelaskan tiap tahap. Jangan menarik anggota tubuh atau memaksakan rentang gerak. Jika pakaian sulit digunakan, diskusikan alternatif dengan pengguna dan tim, bukan mengganti preferensi tanpa penjelasan.\n\nSetelah berpakaian, periksa lipatan, bagian yang terlalu ketat, posisi alat bantu, dan kenyamanan. Sepatu serta alas kaki juga disesuaikan dengan rencana mobilitas. Beri cermin atau kesempatan pengguna merapikan diri bila diinginkan. Catat bantuan yang diperlukan agar petugas berikutnya dapat memberi dukungan konsisten.",
       "Tanyakan pilihan serta siapkan pakaian yang sesuai.|Periksa kemampuan, nyeri, serta posisi bantuan.|Jelaskan dan bantu bagian sulit sesuai urutan rencana.|Periksa kenyamanan serta beri kesempatan merapikan diri.",
       "Jangan menarik lengan yang lemah atau memaksakan sendi demi memasukkan pakaian.",
       "Pak Ishii ingin memakai kemeja favorit, tetapi kesulitan menutup kancing kecil.",
       "Hormati pilihannya dan bantu kancing yang sulit. Diskusikan alat atau alternatif yang sesuai jika dibutuhkan, sambil mempertahankan bagian kegiatan yang masih mampu ia lakukan.",
       "衣服~いふく~pakaian|袖~そで~lengan baju|着脱~ちゃくだつ~memakai dan melepas|身だしなみ~みだしなみ~kerapian penampilan",
       "今日はどの服にしますか。~きょうはどのふくにしますか。~Hari ini ingin memakai pakaian yang mana?|ボタンをお手伝いしましょうか。~ぼたんをおてつだいしましょうか。~Apakah saya bantu mengancingkan?",
       [q("Pengguna kesulitan kancing tetapi memilih bajunya sendiri. Petugas sebaiknya…", "Mengganti tanpa bertanya|Membantu bagian kancing yang sulit|Melarang memilih", 1, "Bantuan sebagian dapat menghormati pilihan serta mempertahankan kemampuan."),
        q("Setelah berpakaian, petugas memeriksa…", "Kenyamanan dan bagian yang menekan|Hanya warna|Hanya jadwal selesai", 0, "Lipatan serta pakaian yang menekan dapat mengganggu kenyamanan.")])

module("daily-life", "Istirahat, aktivitas, dan lingkungan hidup",
       "Menjaga rutinitas, partisipasi, dan kenyamanan ruang pengguna.")
lesson("kaigo-28", "Tidur, istirahat, dan pengamatan malam",
       "Mengidentifikasi kebiasaan tidur dan gangguan lingkungan|Mencatat pola tanpa menetapkan obat tidur",
       "Tidur dan istirahat dipengaruhi rutinitas, aktivitas siang, lingkungan, kenyamanan tubuh, serta kekhawatiran. Pengguna mungkin terbiasa tidur dengan cahaya tertentu atau memerlukan bantuan toilet pada malam hari. Dukungan perlu mempertimbangkan kebiasaan ini dan rencana, bukan menganggap semua orang harus tidur pada waktu yang sama.\n\nPengamatan tidur mencakup waktu tampak tidur, terbangun, keluhan, kegiatan, dan bantuan. Petugas hanya mencatat apa yang dapat diketahui; jangan menuliskan jumlah jam tidur yang pasti bila tidak diamati sepanjang waktu. Mengantuk baru atau sulit dibangunkan memerlukan perhatian dan pelaporan, bukan sekadar dimasukkan ke catatan tidur.",
       "Siapkan lingkungan nyaman dengan pencahayaan dan suara sesuai preferensi serta risiko. Pastikan jalur, bel panggil, dan alat bantu dapat digunakan. Pemeriksaan malam mengikuti rencana dan sebisa mungkin mengurangi gangguan yang tidak perlu. Tetap sapa dan jelaskan sebelum bantuan tubuh.\n\nBila pengguna terus terbangun, tanyakan kebutuhan seperti nyeri, toilet, suhu, atau cemas, lalu laporkan pola kepada tim. Jangan memberi obat tidur tambahan atau menahan pengguna di tempat tidur sebagai keputusan mandiri. Dukungan aktivitas siang dan rutinitas tidur dibahas bersama tim.",
       "Pelajari kebiasaan tidur serta rencana malam.|Atur lingkungan dan jangkauan bel dengan aman.|Amati pola serta bantu kebutuhan sesuai rencana.|Catat fakta dan laporkan perubahan atau keluhan baru.",
       "Perubahan kesadaran atau napas saat malam memerlukan respons sesuai prosedur, bukan menunggu pagi.",
       "Pengguna terbangun tiga kali karena ingin ke toilet. Catatan petugas hanya menyebut 'sulit tidur'.",
       "Catat waktu terbangun, kebutuhan toilet, bantuan, dan respons setelahnya. Informasi tersebut lebih berguna untuk menyesuaikan dukungan daripada label umum.",
       "睡眠~すいみん~tidur|休息~きゅうそく~istirahat|起床~きしょう~bangun tidur|就寝~しゅうしん~pergi tidur",
       "眠れましたか。~ねむれましたか。~Apakah bisa tidur?|何か気になることがありますか。~なにかきになることがありますか。~Apakah ada yang mengganggu pikiran?",
       [q("Catatan tidur yang baik mencakup…", "Waktu dan kebutuhan yang diamati|Label malas|Jumlah jam yang ditebak", 0, "Fakta dan konteks membantu tim menilai pola istirahat."),
        q("Pengguna sangat sulit dibangunkan, berbeda dari biasanya. Petugas perlu…", "Menunggu pagi|Memberi obat tidur lagi|Mengeskalasi segera", 2, "Perubahan kesadaran baru merupakan kebutuhan penilaian segera.")])
lesson("kaigo-29", "Rekreasi, peran sosial, dan aktivitas bermakna",
       "Memilih aktivitas berdasarkan minat pengguna|Menyesuaikan bantuan dan mengevaluasi pengalaman",
       "Kegiatan bermakna berbeda untuk tiap orang: berkebun, musik, membaca, memasak sederhana, berbincang, atau kegiatan spiritual. Tujuan dukungan bukan membuat semua pengguna mengikuti aktivitas yang sama. Minat, riwayat hidup, kemampuan, dan keinginan saat itu perlu diperhatikan. Menolak satu kegiatan tidak otomatis berarti tidak mau bersosialisasi.\n\nPartisipasi dapat berbentuk melakukan sebagian kegiatan, memilih lagu, mengamati, atau membantu menentukan hasil. Aktivitas yang sederhana tetap dapat memberi peran dan rasa dihargai. Hindari kompetisi atau pujian yang merendahkan pengguna sebagai anak kecil.",
       "Cek risiko kegiatan dan rencana individual: posisi, kelelahan, alat, serta kebutuhan pendampingan. Sediakan pilihan yang nyata dan jangan memaksa. Latihan gerak yang bersifat terapi mengikuti arahan terapis; caregiver tidak membuat program latihan klinis sendiri.\n\nEvaluasi pengalaman, bukan hanya jumlah peserta. Tanyakan apakah pengguna menikmati kegiatan, amati kelelahan atau rasa tidak nyaman, dan catat dukungan yang efektif. Kegiatan pendek yang disukai bisa lebih bermakna daripada sesi panjang yang membuat pengguna tertekan.",
       "Tanyakan minat dan pilihan kegiatan.|Periksa kesesuaian dengan kemampuan serta rencana.|Sesuaikan alat, tempo, dan bentuk partisipasi.|Evaluasi kenyamanan, kesenangan, dan keinginan mengulang.",
       "Jangan memaksa kegiatan atau menjalankan latihan terapi di luar arahan dan pelatihan.",
       "Pak Fujita menolak bingo, tetapi antusias ketika ditanya tentang tanaman yang pernah dirawatnya.",
       "Tawarkan kegiatan yang terkait minat tanaman dengan dukungan aman. Gunakan riwayat hidup sebagai panduan, bukan memaksa bingo agar jumlah peserta bertambah.",
       "余暇~よか~waktu luang|役割~やくわり~peran|参加~さんか~partisipasi|趣味~しゅみ~hobi",
       "好きなことは何ですか。~すきなことはなんですか。~Kegiatan apa yang Anda sukai?|またやってみたいですか。~またやってみたいですか。~Apakah ingin mencoba lagi?",
       [q("Keberhasilan aktivitas dapat dilihat dari…", "Semua orang dipaksa hadir|Kesenangan dan partisipasi yang bermakna|Hanya durasi terpanjang", 1, "Pengalaman serta pilihan pengguna merupakan bagian utama evaluasi."),
        q("Pengguna menolak bingo. Petugas sebaiknya…", "Melabeli tidak sosial|Menghukum|Menanyakan minat dan menawarkan pilihan lain", 2, "Penolakan terhadap satu aktivitas tidak menjelaskan semua minat pengguna.")])
lesson("kaigo-30", "Kamar, linen, dan lingkungan yang mendukung",
       "Menata ruang tanpa menghilangkan pilihan|Mengelola linen dan barang pribadi sesuai prosedur",
       "Ruang pengguna adalah lingkungan hidup, bukan hanya tempat petugas bekerja. Susunan barang dapat membantu orientasi dan rasa memiliki. Menata ulang semua benda tanpa izin mungkin membuat pengguna bingung atau kehilangan akses ke barang yang penting. Kebersihan, keselamatan, dan preferensi perlu dipertimbangkan bersama.\n\nPeriksa pencahayaan, jalur, suhu, suara, dan jangkauan barang. Bel panggil, minuman yang diizinkan, kacamata, atau alat komunikasi ditempatkan sesuai rencana dan kemampuan. Jangan menaruh barang di sisi yang tidak dapat dijangkau hanya karena terlihat lebih rapi.",
       "Linen bersih dan kotor ditangani dengan alur berbeda. Gunakan pelindung bila diperlukan dan ikuti penanganan fasilitas tanpa mengguncang linen kotor. Jangan meletakkan linen terpakai di lantai atau membawa alat bersih melalui area kotor. Teknik mengganti seprai saat tempat tidur digunakan harus dipelajari dengan supervisi.\n\nSetelah merapikan, pastikan tidak ada lipatan atau benda yang menekan pengguna, serta alat dan barang penting kembali dapat dijangkau. Konfirmasikan perubahan kepada pengguna. Laporkan kerusakan alat atau lingkungan agar diperbaiki melalui jalur yang benar.",
       "Tanyakan pilihan susunan barang dan cek rencana.|Siapkan linen serta alat dengan alur bersih-kotor.|Rapikan sesuai teknik yang dilatih tanpa memindahkan barang sembarangan.|Periksa kenyamanan, akses bel, serta keamanan jalur.",
       "Jangan mengubah posisi alat, tinggi tempat tidur, atau pengaman tanpa mengikuti rencana dan prosedur.",
       "Setelah kamar dirapikan, bel panggil diletakkan di meja yang jauh dari tangan pengguna.",
       "Kembalikan bel ke lokasi yang dapat dijangkau sesuai rencana dan pastikan pengguna dapat menggunakannya. Kerapian tidak boleh mengurangi akses bantuan.",
       "環境~かんきょう~lingkungan|寝具~しんぐ~perlengkapan tidur|シーツ~しーつ~seprai|整理~せいり~penataan",
       "物の場所を変えてもいいですか。~もののばしょをかえてもいいですか。~Bolehkah saya mengubah letak barang?|呼び出しボタンに手が届きますか。~よびだしぼたんにてがとどきますか。~Apakah tombol panggil dapat dijangkau?",
       [q("Menata ulang barang pengguna sebaiknya…", "Dilakukan tanpa bertanya|Mempertimbangkan izin, kebiasaan, dan jangkauan|Mengutamakan foto yang rapi", 1, "Ruang adalah lingkungan hidup dan harus tetap mendukung kemampuan pengguna."),
        q("Linen kotor perlu…", "Ditangani melalui alur yang ditetapkan|Diguncang di dekat alat bersih|Diletakkan di lantai", 0, "Pemisahan dan penanganan benar mengurangi kontaminasi.")])

module("team", "Proses perawatan, laporan, dan profesionalisme",
       "Mengubah pengamatan menjadi catatan, serah terima, dan evaluasi dukungan.")
lesson("kaigo-31", "Asesmen, rencana, pelaksanaan, dan evaluasi",
       "Menjelaskan siklus proses perawatan|Menghubungkan tujuan dengan dukungan dan respons nyata",
       "Proses perawatan adalah siklus memahami kebutuhan, menyusun tujuan serta rencana, melaksanakan dukungan, dan mengevaluasi hasil. Caregiver menyumbang pengamatan aktivitas harian dan preferensi pengguna kepada tim. Kebutuhan tidak hanya dinilai dari kekurangan; kemampuan yang masih dimiliki dan hal yang ingin dilakukan juga penting.\n\nTujuan sebaiknya dapat dipahami pengguna dan ditinjau melalui pengamatan. Contohnya bukan 'semua lebih baik', melainkan pengguna dapat memilih pakaian dan memakai bagian tertentu dengan bantuan yang sesuai. Rencana menjelaskan bentuk bantuan, alat, tingkat pengawasan, serta hal yang perlu dilaporkan.",
       "Pelaksanaan mengikuti rencana terbaru dan kondisi aktual. Bila pengguna menolak, mengalami keluhan baru, atau cara dukungan tidak efektif, catat dan diskusikan. Mengikuti rencana tidak berarti mengabaikan perubahan. Sebaliknya, perubahan rencana dilakukan melalui tim agar konsisten dan aman.\n\nEvaluasi membandingkan tujuan dengan respons: bagian yang berhasil, bantuan yang masih diperlukan, kenyamanan, dan pilihan pengguna. Dokumentasi yang jelas membuat keputusan penyesuaian memiliki dasar. Jangan menandai tujuan tercapai hanya karena kegiatan sudah dijadwalkan atau petugas selesai membantu.",
       "Kumpulkan kebutuhan, kemampuan, dan preferensi pengguna.|Pahami tujuan serta rencana yang disepakati tim.|Laksanakan sambil mengamati respons aktual.|Laporkan hasil dan bahas penyesuaian bila diperlukan.",
       "Rencana tidak diubah sendiri untuk kegiatan yang memerlukan penilaian tim atau kewenangan klinis.",
       "Rencana meminta bantuan sebagian berpakaian. Pengguna berhasil memakai lengan baju tetapi masih perlu bantuan kancing.",
       "Catat kemampuan tersebut dan jenis bantuan yang masih dibutuhkan. Informasi spesifik membantu tim mengevaluasi tujuan tanpa menyebut 'mandiri penuh' secara keliru.",
       "介護過程~かいごかてい~proses perawatan|計画~けいかく~rencana|実施~じっし~pelaksanaan|評価~ひょうか~evaluasi",
       "計画を確認してから行います。~けいかくをかくにんしてからおこないます。~Saya akan melakukannya setelah memeriksa rencana.|どこまでできたか記録します。~どこまでできたかきろくします。~Saya mencatat bagian yang berhasil dilakukan.",
       [q("Evaluasi rencana menilai…", "Jadwal saja|Respons, tujuan, dan kebutuhan bantuan|Kecepatan petugas saja", 1, "Hasil dukungan harus dihubungkan dengan tujuan serta pengalaman pengguna."),
        q("Pengguna hanya mandiri pada sebagian berpakaian. Catatan sebaiknya…", "Mandiri penuh|Tidak mampu sama sekali|Menjelaskan bagian mandiri dan bagian yang dibantu", 2, "Kemampuan spesifik memberi dasar penyesuaian yang akurat.")])
lesson("kaigo-32", "Ho-ren-so dan catatan objektif",
       "Membedakan laporan, berbagi informasi, dan konsultasi|Memisahkan fakta, ucapan, dan dugaan",
       "Hōkoku adalah melaporkan hasil atau perubahan kepada penanggung jawab. Renraku adalah menyampaikan informasi yang diperlukan pihak terkait. Sōdan adalah berkonsultasi saat membutuhkan keputusan atau bantuan. Ketiganya sering dirangkum sebagai hō-ren-sō. Informasi perlu disampaikan tepat waktu, kepada orang yang tepat, dan melalui saluran yang sesuai.\n\nCatatan objektif menjelaskan hal yang dapat diamati atau diukur. Ucapan pengguna dapat dicatat sebagai ucapan, sementara dugaan tidak ditulis sebagai fakta. 'Pukul 12.10 mengatakan mual dan makan dua sendok' lebih berguna daripada 'sengaja tidak makan'. Hindari label yang menyalahkan atau menilai kepribadian.",
       "Untuk laporan, susun waktu, lokasi, kejadian, kondisi biasanya, tindakan yang dilakukan, serta respons. Jika memerlukan keputusan, nyatakan pertanyaan dengan jelas. Informasi darurat disampaikan segera; catatan lengkap menyusul. Jangan menunggu akhir shift bila perubahan memerlukan respons saat itu.\n\nGunakan sistem resmi dan koreksi kesalahan sesuai kebijakan agar jejak perubahan tetap jelas. Jangan menghapus fakta untuk membuat laporan tampak baik. Pemeriksaan nama serta penerima informasi membantu mencegah data salah masuk ke pengguna lain.",
       "Identifikasi apakah perlu laporan, informasi, atau konsultasi.|Sampaikan fakta, waktu, dan kebutuhan keputusan.|Gunakan saluran resmi serta penerima yang tepat.|Catat respons dan tindak lanjut tanpa mengganti fakta dengan dugaan.",
       "Jangan mencatat diagnosis yang belum ditetapkan atau menjadikan penilaian pribadi sebagai fakta.",
       "Petugas ingin menulis 'pengguna malas makan', padahal pengguna mengatakan mual dan berhenti setelah dua sendok.",
       "Tulis waktu, ucapan mual, konsumsi yang diamati, serta tindakan pelaporan. Informasi ini dapat ditindaklanjuti tanpa label yang menghakimi.",
       "報告~ほうこく~laporan|連絡~れんらく~penyampaian informasi|相談~そうだん~konsultasi|事実~じじつ~fakta",
       "十二時十分に、吐き気がすると話しました。~じゅうにじじゅっぷんに、はきけがするとはなしました。~Pukul 12.10 beliau mengatakan mual.|対応について相談したいです。~たいおうについてそうだんしたいです。~Saya ingin berkonsultasi tentang tindakannya.",
       [q("Catatan objektif adalah…", "Beliau malas|Sengaja merepotkan|Pukul 12.10 mengatakan mual dan makan dua sendok", 2, "Waktu, ucapan, dan pengamatan dapat diverifikasi serta ditindaklanjuti."),
        q("Sōdan dilakukan ketika…", "Perlu keputusan atau bantuan|Ingin menyembunyikan kesalahan|Tidak ingin memberi informasi", 0, "Konsultasi membantu tugas ditangani dengan kewenangan dan dukungan yang sesuai.")])
lesson("kaigo-33", "Serah terima, insiden, dan kesiapan caregiver",
       "Memberi serah terima yang dapat ditindaklanjuti|Menjaga kebugaran kerja dan meminta dukungan",
       "Serah terima membantu kesinambungan dukungan antarpengguna shift. Isinya mencakup kondisi, perubahan, bantuan yang sudah diberikan, respons, serta hal yang belum selesai. Informasi harus cukup spesifik agar petugas berikutnya tahu apa yang perlu dipantau atau dilakukan. Membaca catatan tidak menggantikan klarifikasi bila ada bagian penting yang belum jelas.\n\nLaporan insiden menjelaskan urutan fakta dan tindakan setelah keselamatan ditangani. Kesalahan atau near miss perlu dilaporkan melalui prosedur. Menyembunyikan kejadian untuk melindungi citra petugas dapat menghilangkan kesempatan pencegahan dan membuat tim kehilangan informasi penting.",
       "Saat serah terima, pastikan penerima memahami prioritas serta tugas tertunda. Ulangi poin yang berkaitan dengan keselamatan dan konfirmasikan tanggung jawab. Lindungi privasi dengan memilih tempat atau sistem yang sesuai. Hindari menyertakan gosip atau label yang tidak membantu dukungan.\n\nCaregiver juga perlu menjaga kesiapan diri: istirahat, pelatihan, ergonomi, dan dukungan emosi. Laporkan kelelahan berat, cedera, atau keterbatasan yang dapat memengaruhi keselamatan. Meminta bantuan merupakan bagian tanggung jawab profesional, bukan kegagalan pribadi. Gunakan saluran supervisi ketika menghadapi tekanan atau dugaan perlakuan tidak layak.",
       "Susun perubahan, tindakan, respons, dan tugas tertunda.|Sampaikan kepada penerima yang tepat serta konfirmasi prioritas.|Laporkan insiden secara faktual melalui jalur resmi.|Minta dukungan bila kondisi petugas memengaruhi keselamatan.",
       "Jangan menyembunyikan insiden atau melanjutkan tugas berisiko ketika kondisi diri tidak aman.",
       "Petugas malam menemukan keluhan baru yang sudah dilaporkan ke perawat, tetapi belum ada tindak lanjut saat pergantian shift.",
       "Serahkan keluhan, waktu laporan, respons yang diterima, dan tugas tindak lanjut yang belum selesai. Konfirmasikan siapa yang akan memantau, bukan hanya berkata 'sudah dilaporkan'.",
       "申し送り~もうしおくり~serah terima|引継ぎ~ひきつぎ~pengalihan tugas|勤務~きんむ~dinas atau kerja|疲労~ひろう~kelelahan",
       "まだ確認が必要です。~まだかくにんがひつようです。~Masih perlu dikonfirmasi.|次の担当者に申し送ります。~つぎのたんとうしゃにもうしおくります。~Saya akan menyerahkannya kepada petugas berikutnya.",
       [q("Serah terima keluhan yang belum selesai perlu mencakup…", "Tindak lanjut dan penanggung jawab|Hanya 'sudah dilaporkan'|Pendapat tentang sifat pengguna", 0, "Tugas tertunda harus dapat dipahami dan ditindaklanjuti oleh tim berikutnya."),
        q("Kelelahan berat yang memengaruhi tugas sebaiknya…", "Disembunyikan|Dilaporkan untuk pengaturan dukungan|Diatasi dengan memaksa semua transfer sendiri", 1, "Kesiapan petugas merupakan bagian keselamatan pengguna serta tim.")])

module("japanese", "Bahasa Jepang Kaigo dan integrasi materi",
       "Membaca istilah, menyampaikan bantuan, dan memahami catatan kerja dalam konteks.")
lesson("kaigo-34", "Istilah tubuh, posisi, angka, dan waktu",
       "Membaca istilah penting dalam konteks|Mengonfirmasi waktu, sisi tubuh, dan jumlah secara tepat",
       "Kosakata Kaigo perlu dipelajari bersama konteks. 右 (migi) berarti kanan, 左 (hidari) kiri, 手 (te) tangan, 足 (ashi) kaki, dan 腰 (koshi) pinggang. Istilah 痛み (itami) adalah nyeri, sedangkan かゆみ (kayumi) adalah gatal. Jangan mengganti satu keluhan dengan istilah lain ketika membuat laporan.\n\nWaktu dan urutan sering menentukan tindakan: 前 (mae) sebelum, 後 (ato) sesudah, 今 (ima) sekarang, dan 後で (ato de) nanti. 午前 (gozen) serta 午後 (gogo) membedakan waktu sebelum dan sesudah tengah hari. Bacaan angka berubah dalam beberapa gabungan; 四時 dibaca よじ, 七時 しちじ, dan 九時 くじ. Pastikan isi melalui pengulangan, bukan hanya mengenali kanji.",
       "Latih membaca instruksi singkat: 食事の前に手を洗います berarti mencuci tangan sebelum makan. 右の足が痛いです berarti kaki kanan sakit. Bedakan pengamatan petugas dari ucapan pengguna dengan menambahkan と話しています saat melaporkan ucapan.\n\nPada jumlah cairan atau ukuran lain, sebutkan satuan sesuai catatan dan jangan mengubah angka tanpa dasar. Istilah 少し (sedikit) atau 全部 (seluruhnya) bukan pengukuran pasti. Jika belum memahami angka, sisi tubuh, atau urutan, ulangi kepada pemberi informasi dan minta konfirmasi. Buat kartu belajar pribadi dari kosakata lesson sebelumnya, lalu gunakan dalam kalimat yang relevan.",
       "Baca istilah Jepang, kana, lalu artinya.|Gunakan istilah dalam satu kalimat kerja.|Latih waktu, sisi tubuh, serta urutan sebelum-sesudah.|Ulangi informasi penting dan minta konfirmasi bila ragu.",
       "Jangan bertindak berdasarkan tebakan angka, sisi tubuh, atau satuan yang belum dipahami.",
       "Pengguna berkata '左の足が痛いです'. Petugas baru hanya memahami 足 dan menulis kaki kanan sakit.",
       "左 berarti kiri. Konfirmasikan keluhan dengan menunjuk secara sopan atau mengulang ucapan, lalu laporkan lokasi yang benar. Jangan mengisi sisi tubuh dari tebakan.",
       "右~みぎ~kanan|左~ひだり~kiri|午前~ごぜん~sebelum tengah hari|午後~ごご~sesudah tengah hari",
       "左の足が痛いですか。~ひだりのあしがいたいですか。~Apakah kaki kiri sakit?|午後四時ですね。~ごごよじですね。~Pukul empat sore, benar?|食事の後で行います。~しょくじのあとでおこないます。~Dilakukan setelah makan.|量と時間を確認します。~りょうとじかんをかくにんします。~Saya memastikan jumlah dan waktunya.",
       [q("左の足が痛いです berarti…", "Kaki kanan gatal|Kaki kiri sakit|Tangan kiri sakit", 1, "左 adalah kiri, 足 adalah kaki, dan 痛い menyampaikan nyeri."),
        q("午後四時 adalah…", "Pukul empat sore|Pukul empat pagi|Pukul sembilan malam", 0, "午後 menyatakan sesudah tengah hari; 四時 dibaca よじ.")])
lesson("kaigo-35", "Percakapan bantuan dan konfirmasi keluhan",
       "Menggunakan permintaan izin serta tawaran bantuan|Menjawab penolakan dengan klarifikasi dan empati",
       "Pola ～てもよろしいですか meminta izin secara sopan, misalnya お手伝いしてもよろしいですか (bolehkah saya membantu?). Pola ～ましょうか menawarkan bantuan, misalnya カーテンを閉めましょうか (apakah tirainya saya tutup?). Gunakan kalimat yang jelas sesuai kemampuan pengguna, bukan keigo panjang yang sulit dipahami.\n\nUntuk keluhan, tanyakan どこが痛いですか (bagian mana yang sakit?) dan いつからですか (sejak kapan?). Dengarkan tanpa memberi diagnosis. Jawaban 大丈夫です dapat berarti baik-baik saja atau penolakan sopan terhadap tawaran; konteks dan konfirmasi tetap penting. まだ berarti belum atau masih, sedangkan もう dapat berarti sudah. Salah memahami kata ini dapat mengubah urutan bantuan.",
       "Contoh percakapan: petugas 'お手洗いに行きますか。' Pengguna 'まだ大丈夫です。' Petugas '分かりました。行きたくなったら教えてください。' Artinya: apakah ingin ke toilet; belum perlu; baik, beri tahu jika ingin pergi. Respons petugas menghormati jawaban dan tetap menyediakan akses bantuan.\n\nContoh lain: pengguna '寒いです。' Petugas '寒いのですね。毛布をお持ちしましょうか。' Artinya: dingin; jadi terasa dingin, apakah saya bawakan selimut? Jika keluhan baru atau terkait kondisi klinis, tambahkan pelaporan kepada tim sesuai prosedur. Tujuan latihan bahasa adalah komunikasi yang akurat, bukan sekadar menghafal terjemahan.",
       "Latih sapaan dan izin sebelum setiap bantuan.|Tanyakan satu keluhan atau pilihan dalam satu waktu.|Ulangi inti jawaban untuk konfirmasi.|Sampaikan rencana bantuan dan akses panggilan yang jelas.",
       "Bahasa sopan tidak menggantikan persetujuan; perhatikan jawaban dan respons pengguna.",
       "Petugas menawarkan bantuan. Pengguna menjawab 'まだ大丈夫です'. Petugas mulai membantu tanpa konfirmasi.",
       "Dalam konteks ini pengguna menyatakan belum perlu. Konfirmasikan jawaban, hormati pilihan, dan jelaskan bahwa bantuan tetap tersedia bila dibutuhkan.",
       "痛み~いたみ~nyeri|寒い~さむい~dingin|まだ~まだ~belum atau masih|手伝う~てつだう~membantu",
       "お手伝いしてもよろしいですか。~おてつだいしてもよろしいですか。~Bolehkah saya membantu?|まだ大丈夫です。~まだだいじょうぶです。~Saat ini belum perlu.|行きたくなったら教えてください。~いきたくなったらおしえてください。~Beri tahu jika ingin pergi.|どこが痛いですか。~どこがいたいですか。~Bagian mana yang sakit?|寒いのですね。~さむいのですね。~Jadi terasa dingin, ya.|毛布をお持ちしましょうか。~もうふをおもちしましょうか。~Apakah saya bawakan selimut?",
       [q("～てもよろしいですか digunakan untuk…", "Memberi diagnosis|Meminta izin|Mengancam", 1, "Pola ini meminta persetujuan secara sopan sebelum tindakan."),
        q("'まだ大丈夫です' saat tawaran bantuan dapat berarti…", "Selalu setuju dibantu|Harus dipaksa|Belum perlu, sehingga perlu dihormati dan dikonfirmasi", 2, "Makna 大丈夫 bergantung konteks; jangan otomatis menganggap persetujuan.")])
lesson("kaigo-36", "Membaca catatan dan menyelesaikan kasus terpadu",
       "Mengambil waktu, fakta, dan tindak lanjut dari catatan Jepang|Menggabungkan pilihan pengguna, keselamatan, dan laporan",
       "Catatan kerja sering berisi waktu, kegiatan, bantuan, ucapan, serta tindak lanjut. Contoh buatan untuk latihan: '12:10 昼食を半分食べた。吐き気がすると話した。看護師に報告した。' Artinya: pukul 12.10 makan setengah makan siang; mengatakan mual; telah dilaporkan kepada perawat. Catatan ini tidak menyebut diagnosis atau memastikan semua tindak lanjut selesai.\n\nContoh kedua: '15:00 歩行中に疲れたと話した。椅子で休んだ。' Artinya: pukul 15.00 mengatakan lelah saat berjalan; beristirahat di kursi. Dalam serah terima, petugas masih perlu menjelaskan respons setelah istirahat dan arahan tim bila ada. Kata 報告した menunjukkan telah melapor, bukan otomatis memperoleh keputusan.",
       "Kasus terpadu: pengguna memilih makan sendiri, tetapi hari ini tampak lebih mengantuk dan mengalami kesulitan baru ketika menelan. Dukungan kemandirian tetap penting, namun keselamatan dan perubahan kondisi mengharuskan penghentian pemberian serta eskalasi sesuai prosedur. Jangan melanjutkan hanya karena pengguna sebelumnya mandiri.\n\nGunakan alur belajar yang sudah dipelajari: kenali kondisi dasar, dengarkan keluhan, cek rencana, minta persetujuan, siapkan bantuan, amati respons, dan laporkan perubahan. Setelah kebutuhan segera ditangani, dokumentasikan fakta serta tindak lanjut. Menuntaskan kursus menunjukkan materi telah dipelajari dan cek pemahaman dikerjakan; kompetensi praktik tetap dinilai melalui pelatihan dan supervisi kerja.",
       "Cari waktu serta tindakan utama dalam catatan.|Pisahkan ucapan pengguna dari pengamatan petugas.|Identifikasi informasi atau tindak lanjut yang belum jelas.|Gunakan prinsip martabat, keselamatan, dan koordinasi pada kasus terpadu.",
       "Cek pemahaman kursus bukan ujian resmi SSW dan tidak membuktikan kewenangan praktik klinis.",
       "Catatan menyebut '看護師に報告した', tetapi petugas berikutnya menganggap semua instruksi sudah dilaksanakan.",
       "Konfirmasikan respons perawat, instruksi yang diterima, dan tugas yang masih tertunda. Catatan 'sudah dilaporkan' tidak sama dengan 'tindak lanjut sudah selesai'.",
       "半分~はんぶん~setengah|吐き気~はきけ~mual|歩行中~ほこうちゅう~saat berjalan|対応~たいおう~respons atau penanganan",
       "昼食を半分食べました。~ちゅうしょくをはんぶんたべました。~Makan setengah porsi makan siang.|吐き気がすると話しています。~はきけがするとはなしています。~Beliau mengatakan mual.|看護師に報告しました。~かんごしにほうこくしました。~Saya telah melaporkan kepada perawat.|その後の対応を確認します。~そのごのたいおうをかくにんします。~Saya memastikan penanganan setelahnya.",
       [q("看護師に報告した berarti…", "Telah dilaporkan kepada perawat|Diagnosis sudah pasti|Semua tindak lanjut pasti selesai", 0, "Kalimat ini menyatakan tindakan pelaporan; respons dan tindak lanjut tetap perlu diketahui."),
        q("Pengguna biasanya mandiri makan, tetapi kini ada kesulitan menelan baru. Petugas perlu…", "Tetap lanjut demi kemandirian|Hentikan pemberian dan eskalasi sesuai prosedur|Mengubah tekstur sendiri", 1, "Kondisi aktual dan keselamatan didahulukan sambil tetap menghormati pengguna.")])

ethics = modules[0]["lessons"][2]
ethics["objectives"].append("Mengenali perlakuan tidak layak dan melaporkannya melalui jalur fasilitas")
ethics["sections"].append(dict(title="Mencegah perlakuan tidak layak", body=
    "Perlakuan tidak layak dapat berbentuk kekerasan, penghinaan, ancaman, pengabaian kebutuhan, penyalahgunaan uang, atau pelanggaran batas tubuh dan privasi. Tekanan kerja tidak membenarkan tindakan tersebut. Pengguna yang kesulitan bicara atau membutuhkan bantuan penuh tetap memiliki hak untuk didengar dan dilindungi.\n\n"
    "Bila menyaksikan kejadian atau menerima keluhan, prioritaskan keselamatan dan ikuti jalur pelaporan perlindungan pengguna. Catat ucapan serta fakta yang diamati, jangan menyelidiki atau menyimpulkan sendiri. Hindari membalas, menyalahkan pengguna, atau menjanjikan rahasia yang mencegah perlindungan. Gunakan supervisi dan dukungan tim ketika tekanan kerja mulai memengaruhi perilaku petugas."))
body = modules[1]["lessons"][0]
body["objectives"].append("Melaporkan nyeri dan mendukung kenyamanan sesuai rencana, termasuk akhir hayat")
body["sections"].append(dict(title="Nyeri, kenyamanan, dan dukungan akhir hayat", body=
    "Keluhan nyeri perlu didengar: lokasi, waktu mulai, kegiatan saat muncul, serta kata yang digunakan pengguna. Pada orang yang sulit bicara, amati perubahan ekspresi, gerak, dan respons bantuan lalu laporkan. Gunakan alat penilaian hanya bila telah ditugaskan dan dilatih. Jangan menganggap diam berarti tidak nyeri atau mengubah obat untuk mengatasinya.\n\n"
    "Dalam perawatan akhir hayat, dukungan mengikuti rencana tim dan pilihan yang telah dikomunikasikan oleh pengguna atau pihak berwenang sesuai prosedur. Caregiver membantu kenyamanan, privasi, hubungan keluarga, dan kebutuhan spiritual tanpa memaksakan keyakinan sendiri. Penolakan makan, perubahan napas, atau keluhan baru tetap disampaikan; tidak ditangani dengan pemaksaan maupun keputusan klinis mandiri. Kehadiran yang tenang, mendengarkan, dan serah terima yang jelas merupakan bagian penting dukungan."))

sources = [
    dict(title="MHLW — materi Kaigo dan revisi ke-2 (2026)",
         url="https://www.mhlw.go.jp/stf/newpage_28131.html"),
    dict(title="MHLW — teks resmi Kaigo bahasa Indonesia, revisi ke-2",
         url="https://www.mhlw.go.jp/content/12000000/001472066.pdf"),
    dict(title="MHLW — cakupan evaluasi keterampilan dan bahasa Kaigo",
         url="https://www.mhlw.go.jp/stf/newpage_000117702.html"),
    dict(title="WHO — pemahaman demensia",
         url="https://www.who.int/news-room/fact-sheets/detail/dementia"),
    dict(title="CDC — kebersihan tangan petugas",
         url="https://www.cdc.gov/clean-hands/hcp/clinical-safety/"),
    dict(title="OSHA — penanganan pengguna dan alat transfer",
         url="https://www.osha.gov/healthcare/safe-patient-handling"),
    dict(title="NHS — kesulitan menelan",
         url="https://www.nhs.uk/symptoms/swallowing-problems-dysphagia/"),
]

if __name__ == "__main__":
    destination = Path(__file__).resolve().parents[1] / "assets/content/ssw/kaigo.json"
    destination.parent.mkdir(parents=True, exist_ok=True)
    destination.write_text(json.dumps(dict(version=1, reviewed="2026-10-11",
                                          modules=modules, sources=sources),
                                      ensure_ascii=False, indent=2) + "\n", encoding="utf-8")
    print(f"Created {len(modules)} modules, {sum(len(m['lessons']) for m in modules)} lessons")
    docs = Path(__file__).resolve().parents[1] / "docs/kaigo-curriculum.md"
    lines = ["# Kurikulum Kaigo", "",
             "Kurikulum aktif: **12 modul, 36 lesson, 72 cek pemahaman, 144 entri kosakata, dan 80 ungkapan Jepang**. Estimasi 20 menit per lesson (sekitar 12 jam termasuk latihan serta pengulangan).",
             "",
             "Setiap lesson memiliki tujuan belajar, penjelasan konsep, penerapan kerja, alur bantuan, perhatian keselamatan, kasus dengan respons, kosakata dengan kana, ungkapan dengan terjemahan, serta dua pertanyaan dengan pembahasan.",
             "",
             "Materi dan pertanyaan ditulis mandiri dalam bahasa Indonesia; referensi resmi dipakai untuk memetakan cakupan. Konten tidak menyalin buku atau bank soal resmi dan tidak menjanjikan kelulusan. Keterampilan praktik memerlukan pelatihan langsung, supervisi, SOP serta rencana individual.",
             "", "## Urutan modul", ""]
    for i, m in enumerate(modules, 1):
        lines += [f"### {i}. {m['title']}", "", m["description"], ""]
        for l in m["lessons"]:
            lines += [f"- **{l['id']}: {l['title']}** — " + "; ".join(l["objectives"]) + "."]
        lines += [""]
    lines += ["## Alur aplikasi", "",
              "Buka Belajar → SSW → Kaigo → Materi. Halaman bidang menampilkan kurikulum dan progress; tiap modul dapat dibuka untuk memilih lesson. Tombol Mulai/Lanjutkan memilih lesson pertama yang belum selesai. Route lesson `/ssw/nursing-care/lessons/{lessonId}` menangani ID tidak dikenal tanpa error.",
              "",
              "Jawaban dipilih sebelum pembahasan ditampilkan. Jika belum tepat, pengguna dapat membaca penjelasan dan mencoba kembali. Semua jawaban perlu tepat sebelum tombol Simpan & lanjutkan tersedia. Kelulusan cek ini hanya penanda belajar, bukan pengukuran kompetensi praktik atau skor ujian resmi.",
              "",
              "Materi berupa asset offline `assets/content/ssw/kaigo.json`. Pembuka referensi membutuhkan browser/jaringan. Progress menggunakan SharedPreferences dengan key per akun `ssw_kaigo_completed_v1_{userId}`, tersimpan pada perangkat yang sama, tidak disinkronkan ke server, dan dibersihkan setelah penghapusan akun berhasil. Tamu dapat belajar dengan progress sesi saja. Menyelesaikan lesson tidak memberi XP atau achievement palsu.",
              "",
              "Materi Kaigo dapat dibaca tanpa pembelian. Ini tidak mengubah entitlement JLPT atau ujian yang sudah ada. Latihan dan simulasi Kaigo serta materi, latihan, dan simulasi Pembersihan Gedung kini aktif; lihat [latihan SSW](ssw-assessments.md). Pertanian serta Perikanan dan Budi Daya juga aktif; lihat [bidang baru](ssw-agriculture-fishery.md). Materi 15 bidang lain masih Coming Soon.",
              "", "## Referensi", "", "Ditinjau 11 Oktober 2026.", ""]
    lines += [f"- [{s['title']}]({s['url']})" for s in sources]
    docs.write_text("\n".join(lines) + "\n", encoding="utf-8")
