package api

// sswCurriculumResourceRecyclingEnglish contains exact, hand-reviewed English
// translations for resource-recycling SSW learning material.
var sswCurriculumResourceRecyclingEnglish = map[examSSWEnglishKey]string{
	{field: "ssw.title", source: `Bongkar muatan dan jalur kendaraan`}:                     `Unloading and Vehicle Routes`,
	{field: "ssw.title", source: `Keselamatan ban berjalan dan jalur pemilahan`}:           `Conveyor and Sorting-Line Safety`,
	{field: "ssw.title", source: `Kondisi tidak normal dan perawatan peralatan`}:           `Abnormal Conditions and Equipment Maintenance`,
	{field: "ssw.title", source: `Memilah berdasarkan kriteria dan tujuan`}:                `Sorting by Criteria and Destination`,
	{field: "ssw.title", source: `Mengikuti urutan proses dan mencatat hasil`}:             `Following Process Order and Recording Results`,
	{field: "ssw.title", source: `Pemeriksaan akhir dan serah terima giliran`}:             `Final Checks and Shift Handover`,
	{field: "ssw.title", source: `Pemeriksaan awal peralatan pengolahan`}:                  `Pre-Operation Checks for Processing Equipment`,
	{field: "ssw.title", source: `Pemeriksaan bahan tidak dikenal dan bahaya tersembunyi`}: `Checking Unidentified Materials and Hidden Hazards`,
	{field: "ssw.title", source: `Pencatatan keluaran dan tujuan berikutnya`}:              `Recording Output and Its Next Destination`,
	{field: "ssw.title", source: `Penerimaan limbah dan pencatatan lot`}:                   `Waste Receipt and Lot Records`,
	{field: "ssw.title", source: `Penyiapan fraksi, wadah, dan tanda tujuan`}:              `Preparing Fractions, Containers, and Destination Labels`,
	{field: "ssw.title", source: `Penyimpanan aman dan pengendalian debu atau tumpahan`}:   `Safe Storage and Dust or Spill Control`,

	{field: "ssw.description", source: `Memilah sesuai kriteria penerimaan dan tujuan proses, menjaga fraksi tetap terpisah, serta melaporkan benda yang salah atau tercampur.`}:                                                                                                                                                                                                                            `Sort according to acceptance criteria and process destination, keep fractions separate, and report misplaced or mixed items.`,
	{field: "ssw.description", source: `Mempelajari pekerjaan antara: menerima limbah, mencocokkan dokumen dan lot, memeriksa kondisi, serta menjaga keselamatan saat bongkar. Per 22 Oktober 2026, Kementerian Lingkungan Hidup Jepang menargetkan penerimaan sektor dimulai pada 2027 dan menyatakan rincian ujian keterampilan belum diputuskan; materi ini tetap tersedia sebagai pembelajaran dasar.`}: `Learn about intermediate processing work: receiving waste, matching documents and lots, checking conditions, and working safely during unloading. As of October 22, 2026, Japan's Ministry of the Environment aims to begin accepting workers in this sector in 2027 and says the skills-test details have not yet been decided; this material remains available as foundational learning.`,
	{field: "ssw.description", source: `Mengikuti instruksi pengoperasian proses dan menjaga peralatan sesuai batas tugas, termasuk menghentikan serta melaporkan kondisi tidak normal.`}:                                                                                                                                                                                                                   `Follow process-operation instructions and maintain equipment within your assigned role, including stopping and reporting abnormal conditions.`,
	{field: "ssw.description", source: `Menjaga identitas hasil, area penyimpanan, keselamatan pemindahan, serta catatan dan serah terima antarproses.`}:                                                                                                                                                                                                                                                    `Maintain product identity, storage areas, safe transfers, records, and handovers between processes.`,

	{field: "ssw.objective", source: `Memeriksa kondisi mesin, pelindung, jalur, dan tombol keselamatan sebelum menjalankan proses yang ditugaskan.`}:            `Check the machine, guards, routes, and safety controls before operating your assigned process.`,
	{field: "ssw.objective", source: `Memisahkan material sesuai label aliran dan kriteria fasilitas, serta menjaga bahan yang belum dikenali tetap tertahan.`}:  `Separate materials according to stream labels and facility criteria, and keep unidentified materials on hold.`,
	{field: "ssw.objective", source: `Mencocokkan area, alat, dan lot pada akhir giliran serta menyerahkan status pekerjaan yang masih terbuka.`}:                `Check the area, tools, and lots at the end of a shift and hand over the status of any open work.`,
	{field: "ssw.objective", source: `Mencocokkan asal, jenis, berat atau jumlah, waktu, dan identitas kelompok dengan catatan penerimaan fasilitas.`}:           `Match the source, type, weight or quantity, time, and group identity against the facility's receiving records.`,
	{field: "ssw.objective", source: `Mencocokkan hasil olahan, berat atau jumlah, nomor lot, dan tujuan proses sebelum pemindahan.`}:                            `Verify processed output, weight or quantity, lot number, and process destination before transfer.`,
	{field: "ssw.objective", source: `Mengenali kebocoran, getaran, bunyi, panas, atau kerusakan dan melaporkannya tanpa melakukan perbaikan di luar wewenang.`}: `Recognize and report leaks, vibration, unusual sounds, heat, or damage without making unauthorized repairs.`,
	{field: "ssw.objective", source: `Mengenali muatan yang mencurigakan atau tidak sesuai dan menjaga jarak sambil menghubungi petugas berwenang.`}:             `Recognize suspicious or nonconforming loads, keep your distance, and contact authorized staff.`,
	{field: "ssw.objective", source: `Mengikuti urutan pengolahan yang ditetapkan serta mencatat lot, waktu, dan hasil sesuai format fasilitas.`}:                `Follow the specified processing sequence and record lots, times, and results in the facility's format.`,
	{field: "ssw.objective", source: `Menjaga area simpan tertata dan melaporkan debu, tumpahan, bau, atau bahan yang dapat mencemari lingkungan.`}:              `Keep storage areas orderly and report dust, spills, odors, or materials that could contaminate the environment.`,
	{field: "ssw.objective", source: `Menjaga fraksi yang telah dipilah tetap terpisah serta memberi tanda kelompok dan tujuan sebelum tahap berikutnya.`}:       `Keep sorted fractions separate and label the group and destination before the next stage.`,
	{field: "ssw.objective", source: `Menjaga jarak dari kendaraan serta memastikan zona bongkar dan komunikasi dengan operator aman.`}:                          `Keep clear of vehicles and confirm that the unloading zone and communication with the operator are safe.`,
	{field: "ssw.objective", source: `Menjaga posisi tubuh serta alat jauh dari titik jepit dan menghentikan pekerjaan saat kondisi lini tidak aman.`}:           `Keep your body and tools clear of pinch points and stop work whenever line conditions are unsafe.`,

	{field: "ssw.caution", source: `Jangan berdiri di belakang truk atau di bawah bak yang sedang terangkat.`}:                    `Do not stand behind a truck or under a raised dump bed.`,
	{field: "ssw.caution", source: `Jangan melepas pelindung atau memperbaiki mesin tanpa pelatihan serta prosedur isolasi.`}:     `Do not remove guards or repair a machine without training and the required isolation procedure.`,
	{field: "ssw.caution", source: `Jangan melepas tanda tahan atau meninggalkan area berbahaya tanpa serah terima.`}:             `Do not remove a hold marker or leave a hazardous area without a handover.`,
	{field: "ssw.caution", source: `Jangan membuka panel listrik atau membersihkan sumbatan tanpa isolasi dan kewenangan.`}:       `Do not open an electrical panel or clear a blockage without isolation and authorization.`,
	{field: "ssw.caution", source: `Jangan membuka wadah, mencium dekat, atau memilah benda yang tidak dikenal dengan tangan.`}:   `Do not open a container, smell an unidentified item closely, or sort it by hand.`,
	{field: "ssw.caution", source: `Jangan memilah berdasarkan tebakan atau mencampur fraksi untuk mengosongkan wadah.`}:          `Do not sort by guessing or mix fractions just to empty a container.`,
	{field: "ssw.caution", source: `Jangan memindahkan atau mengganti label wadah tanpa pembaruan catatan.`}:                      `Do not move a container or change its label without updating the records.`,
	{field: "ssw.caution", source: `Jangan menerima atau mencampur lot yang identitasnya tidak cocok dengan kriteria fasilitas.`}: `Do not accept or mix a lot whose identity does not match the facility's criteria.`,
	{field: "ssw.caution", source: `Jangan mengirim keluaran tanpa identitas dan tujuan yang dapat diverifikasi.`}:                `Do not send output without a verifiable identity and destination.`,
	{field: "ssw.caution", source: `Jangan mengubah urutan atau mencampur lot untuk mengejar jumlah produksi.`}:                   `Do not change the sequence or mix lots to meet a production target.`,
	{field: "ssw.caution", source: `Jangan menyiram atau menyapu bahan tidak dikenal ke saluran air.`}:                            `Do not wash or sweep unidentified material into a drain.`,
	{field: "ssw.caution", source: `Jangan meraih ke dalam ban berjalan atau rol yang masih bergerak.`}:                           `Do not reach into a conveyor or roller while it is moving.`,

	{field: "ssw.scenario", source: `Berat keluaran pada alat timbang berbeda dari catatan, dan label tujuan pada wadah terlepas.`}:                    `The output weight on the scale differs from the record, and the container's destination label has come off.`,
	{field: "ssw.scenario", source: `Cairan keluar dari drum penyimpanan di dekat saluran lantai dan pekerja tidak mengetahui isi drum.`}:              `Liquid is leaking from a storage drum near a floor drain, and the worker does not know what the drum contains.`,
	{field: "ssw.scenario", source: `Dokumen menyebut logam bersih, tetapi muatan tampak bercampur dengan wadah tertutup tanpa label.`}:                `The documents say clean metal, but the load appears mixed with a sealed, unlabeled container.`,
	{field: "ssw.scenario", source: `Keluaran dari satu proses memiliki tanda lot yang tidak sama dengan catatan bahan masuk.`}:                        `Output from a process has a lot mark that does not match the incoming-material record.`,
	{field: "ssw.scenario", source: `Lampu peringatan mesin pemadat menyala dan pintu pelindung belum terkunci penuh.`}:                                `The compactor's warning light is on, and the guard door is not fully locked.`,
	{field: "ssw.scenario", source: `Pada akhir giliran, satu wadah hasil masih menunggu verifikasi lot dan jalur di dekat mesin diberi tanda batas.`}: `At the end of the shift, one output container is still awaiting lot verification, and the route near the machine has been marked off.`,
	{field: "ssw.scenario", source: `Pada aliran plastik terlihat satu wadah aerosol tertutup yang tidak tercantum di kriteria pemilahan.`}:            `A sealed aerosol container appears in the plastic stream, but it is not listed in the sorting criteria.`,
	{field: "ssw.scenario", source: `Pada ban berjalan terlihat baterai menggembung dan mengeluarkan bau, sementara pekerja lain akan mengambilnya.`}:  `A swollen, odorous battery is visible on the conveyor while another worker is about to pick it up.`,
	{field: "ssw.scenario", source: `Pompa pada lini mengeluarkan bau terbakar dan getaran bertambah saat berjalan.`}:                                  `A line pump smells burnt and is vibrating more than usual while running.`,
	{field: "ssw.scenario", source: `Potongan kain tersangkut di antara rol dan pekerja hendak menariknya dengan tangan saat mesin berjalan lambat.`}:  `A piece of cloth is caught between rollers, and a worker is about to pull it out by hand while the machine runs slowly.`,
	{field: "ssw.scenario", source: `Seorang pengemudi mulai membuka bak, sementara pekerja masih berada di belakang truk untuk membaca label.`}:       `A driver starts raising the dump bed while a worker is still behind the truck reading a label.`,
	{field: "ssw.scenario", source: `Wadah logam berisi sejumlah kecil kaca yang masuk dari aliran sebelah; label masih menunjukkan logam bersih.`}:    `A metal container has a small amount of glass from the adjacent stream in it; the label still says clean metal.`,

	{field: "ssw.response", source: `Berikan tanda berhenti sesuai aturan, jaga jarak, peringatkan kru, dan laporkan segera agar petugas terlatih menentukan penanganan.`}:              `Give the stop signal as required, keep your distance, warn the crew, and report it immediately so trained staff can decide how to handle it.`,
	{field: "ssw.response", source: `Hentikan bongkar sesuai prosedur, minta pekerja mundur ke zona aman, dan lanjutkan hanya setelah operator melihat area kosong.`}:                   `Stop unloading according to procedure, ask the worker to move back to the safe zone, and resume only after the operator confirms the area is clear.`,
	{field: "ssw.response", source: `Hentikan pengisian dan tahan wadah; catat campuran serta minta keputusan apakah perlu sortir ulang atau penanganan lain.`}:                         `Stop filling and hold the container. Record the contamination and ask whether it needs to be resorted or handled another way.`,
	{field: "ssw.response", source: `Hentikan peralatan sesuai prosedur darurat, menjauh dari area, dan laporkan nomor mesin serta gejalanya kepada pemimpin.`}:                         `Stop the equipment according to the emergency procedure, move away, and report the machine number and symptoms to the leader.`,
	{field: "ssw.response", source: `Jangan jalankan mesin. Jauhkan orang dari area, beri tahu operator atau pemimpin, dan tunggu pemeriksaan sebelum proses dimulai.`}:                 `Do not start the machine. Keep people away, notify the operator or leader, and wait for an inspection before starting the process.`,
	{field: "ssw.response", source: `Jangan lanjutkan bongkar biasa. Tahan kendaraan di lokasi aman, catat perbedaan, dan minta pemeriksaan penanggung jawab.`}:                         `Do not continue normal unloading. Hold the vehicle in a safe location, record the discrepancy, and request an inspection by the person in charge.`,
	{field: "ssw.response", source: `Jangan masukkan ke fraksi plastik. Hentikan kontak, pisahkan area sesuai aturan, dan laporkan benda tersebut kepada penanggung jawab.`}:            `Do not put it in the plastic fraction. Stop handling it, isolate the area as required, and report the item to the person in charge.`,
	{field: "ssw.response", source: `Jangan menyentuh atau menyiram cairan. Jauhkan orang, tandai area sesuai aturan, dan hubungi petugas berwenang.`}:                                  `Do not touch or wash away the liquid. Keep people back, mark the area as required, and contact authorized staff.`,
	{field: "ssw.response", source: `Jangan meraih. Hentikan lini dengan cara yang diajarkan, jaga orang menjauh, dan minta operator menangani setelah mesin aman.`}:                    `Do not reach in. Stop the line using the method you were taught, keep people back, and ask the operator to handle it once the machine is safe.`,
	{field: "ssw.response", source: `Pertahankan tanda tahan dan batas jalur, catat nomor wadah serta alasan, lalu serahkan langsung kepada penerima giliran dan pemimpin.`}:            `Keep the hold marker and route boundary in place, record the container number and reason, and hand over directly to the next shift and leader.`,
	{field: "ssw.response", source: `Tahan keluaran, catat tanda yang terlihat, dan minta penanggung jawab menelusuri lot sebelum dipindahkan ke proses berikutnya.`}:                   `Hold the output, record the visible markings, and ask the person in charge to trace the lot before moving it to the next process.`,
	{field: "ssw.response", source: `Tahan wadah, catat hasil timbang aktual dan informasi yang masih ada, lalu minta penanggung jawab memulihkan tujuan sebelum pengiriman internal.`}: `Hold the container, record the actual weight and any remaining information, and ask the person in charge to restore its destination before internal transfer.`,

	{field: "ssw.body", source: `Amati dari posisi aman tanpa mengguncang atau membuka wadah. Beri tanda atau batas sesuai prosedur, sampaikan lokasi, bentuk, tulisan, dan bau yang teramati, lalu tunggu instruksi. Jangan menyentuh cairan, baterai rusak, tabung, atau bahan berdebu tanpa penugasan dan perlindungan yang sesuai.

Kasus latihan: Pada ban berjalan terlihat baterai menggembung dan mengeluarkan bau, sementara pekerja lain akan mengambilnya.
Respons yang dianjurkan: Berikan tanda berhenti sesuai aturan, jaga jarak, peringatkan kru, dan laporkan segera agar petugas terlatih menentukan penanganan.`}: `Observe from a safe position without shaking or opening the container. Mark or cordon off the area as required, report the location, shape, markings, and any odor you notice, then wait for instructions. Do not touch liquid, damaged batteries, cylinders, or dusty materials unless assigned and properly protected.

Practice scenario: A swollen, odorous battery is visible on the conveyor while another worker is about to pick it up.
Suggested response: Give the stop signal as required, keep your distance, warn the crew, and report it immediately so trained staff can decide how to handle it.`,
	{field: "ssw.body", source: `Area bongkar memiliki kendaraan bergerak, bak yang naik, muatan yang dapat jatuh, dan titik buta. Pemisahan jalur pejalan kaki, pemberi isyarat, dan operator mengurangi risiko tertabrak atau tertimpa. Pekerja tidak mendekati bak atau muatan sampai operator memberi tanda aman sesuai tata tertib fasilitas.`}:                                           `The unloading area has moving vehicles, raised dump beds, loads that may fall, and blind spots. Separate pedestrian routes and use a signaler to reduce the risk of being struck or crushed. Do not approach the dump bed or load until the operator gives the all-clear required by facility rules.`,
	{field: "ssw.body", source: `Area simpan dapat memiliki material berbeda dengan batas tumpukan, penutup, dan rute kendaraan tersendiri. Debu, cairan, atau serpihan dapat menyebabkan gangguan pernapasan, lantai licin, atau pencemaran. Pengendalian mengikuti perlengkapan dan prosedur fasilitas; pekerja tidak menyapu, menyiram, atau memindahkan tumpahan berbahaya tanpa arahan.`}: `Storage areas may contain different materials with specific stacking limits, covers, and vehicle routes. Dust, liquid, or fragments can cause breathing problems, slippery floors, or contamination. Follow the facility's controls and equipment requirements; do not sweep, wash, or move a hazardous spill without instructions.`,
	{field: "ssw.body", source: `Baca tanda wadah sebelum memasukkan material. Periksa barang yang tampak berbeda dan jangan meletakkan baterai, wadah tertutup, atau bahan asing ke aliran biasa. Jika label tidak terbaca atau kategori tidak jelas, hentikan dan minta konfirmasi; catat lokasi dan jenis yang terlihat.

Kasus latihan: Pada aliran plastik terlihat satu wadah aerosol tertutup yang tidak tercantum di kriteria pemilahan.
Respons yang dianjurkan: Jangan masukkan ke fraksi plastik. Hentikan kontak, pisahkan area sesuai aturan, dan laporkan benda tersebut kepada penanggung jawab.`}: `Read the container label before adding material. Check items that look different; do not put batteries, sealed containers, or foreign materials into a regular stream. If the label is unreadable or the category is unclear, stop and ask for confirmation; record the location and visible type.

Practice scenario: A sealed aerosol container appears in the plastic stream, but it is not listed in the sorting criteria.
Suggested response: Do not put it in the plastic fraction. Stop handling it, isolate the area as required, and report the item to the person in charge.`,
	{field: "ssw.body", source: `Ban berjalan, rol, pemotong, dan alat pemilah dapat menarik pakaian atau menjepit tangan. Area kerja harus bebas orang yang tidak ditugaskan, dan alat bantu hanya digunakan sesuai instruksi. Pekerja tidak meraih benda di antara rol saat mesin bergerak; penghentian dan isolasi mengikuti prosedur fasilitas.`}: `Conveyors, rollers, cutters, and sorting equipment can catch clothing or crush hands. Keep unauthorized people out of the work area and use aids only as instructed. Never reach between rollers while the machine is moving; follow facility procedures for stopping and isolating equipment.`,
	{field: "ssw.body", source: `Berdiri di posisi yang ditandai, gunakan alat bantu resmi, dan jaga lengan serta pakaian dari bagian bergerak. Bila material tersangkut, pelindung terbuka, atau tangan pekerja masuk ke zona bahaya, hentikan sesuai prosedur darurat lalu panggil operator. Jangan menarik sumbatan sebelum mesin aman.

Kasus latihan: Potongan kain tersangkut di antara rol dan pekerja hendak menariknya dengan tangan saat mesin berjalan lambat.
Respons yang dianjurkan: Jangan meraih. Hentikan lini dengan cara yang diajarkan, jaga orang menjauh, dan minta operator menangani setelah mesin aman.`}: `Stand in the marked position, use approved aids, and keep your arms and clothing away from moving parts. If material becomes stuck, a guard opens, or someone's hand enters a danger zone, stop the equipment according to the emergency procedure and call the operator. Do not pull out a blockage until the machine is safe.

Practice scenario: A piece of cloth is caught between rollers, and a worker is about to pull it out by hand while the machine runs slowly.
Suggested response: Do not reach in. Stop the line using the method you were taught, keep people back, and ask the operator to handle it once the machine is safe.`,
	{field: "ssw.body", source: `Catat nomor mesin, waktu, lokasi masalah, bunyi atau indikator yang terlihat, dan langkah sementara yang diambil. Jauhkan rekan dari area bila perlu dan pasang tanda sesuai prosedur. Jangan membuka panel, memasukkan alat, atau mencoba menyalakan ulang beberapa kali untuk menghilangkan alarm.

Kasus latihan: Pompa pada lini mengeluarkan bau terbakar dan getaran bertambah saat berjalan.
Respons yang dianjurkan: Hentikan peralatan sesuai prosedur darurat, menjauh dari area, dan laporkan nomor mesin serta gejalanya kepada pemimpin.`}: `Record the machine number, time, problem location, sounds or indicators observed, and temporary action taken. Keep coworkers away if needed and post signs as required. Do not open a panel, insert a tool, or repeatedly restart the machine to clear an alarm.

Practice scenario: A line pump smells burnt and is vibrating more than usual while running.
Suggested response: Stop the equipment according to the emergency procedure, move away, and report the machine number and symptoms to the leader.`,
	{field: "ssw.body", source: `Cocokkan identitas bahan dengan instruksi sebelum memasukkan material, periksa indikator yang diminta, dan catat perubahan lot atau hasil timbang. Bila keluaran terlihat berbeda, mesin berhenti, atau identitas hilang, tahan produk dan laporkan. Jangan mencampur keluaran dua proses untuk menutup selisih berat.

Kasus latihan: Keluaran dari satu proses memiliki tanda lot yang tidak sama dengan catatan bahan masuk.
Respons yang dianjurkan: Tahan keluaran, catat tanda yang terlihat, dan minta penanggung jawab menelusuri lot sebelum dipindahkan ke proses berikutnya.`}: `Verify material identity against the instructions before feeding it into the process, check required indicators, and record lot changes or weigh-ins. If output looks different, the machine stops, or identity is lost, hold the product and report it. Do not mix output from two processes to hide a weight difference.

Practice scenario: Output from a process has a lot mark that does not match the incoming-material record.
Suggested response: Hold the output, record the visible markings, and ask the person in charge to trace the lot before moving it to the next process.`,
	{field: "ssw.body", source: `Gunakan daftar pemeriksaan sebelum mulai dan laporkan kebocoran, bunyi, getaran, atau lampu peringatan yang tidak normal. Jangan menghidupkan mesin jika pelindung terbuka. Jangan memasukkan tangan untuk membersihkan sumbatan atau membuka panel listrik.

Kasus latihan: Lampu peringatan mesin pemadat menyala dan pintu pelindung belum terkunci penuh.
Respons yang dianjurkan: Jangan jalankan mesin. Jauhkan orang dari area, beri tahu operator atau pemimpin, dan tunggu pemeriksaan sebelum proses dimulai.`}: `Use the startup checklist and report leaks, unusual sounds or vibration, and warning lights. Do not start a machine if a guard is open. Never put your hand in to clear a blockage or open an electrical panel.

Practice scenario: The compactor's warning light is on, and the guard door is not fully locked.
Suggested response: Do not start the machine. Keep people away, notify the operator or leader, and wait for an inspection before starting the process.`,
	{field: "ssw.body", source: `Gunakan jalur pejalan kaki dan zona tunggu yang ditetapkan. Pastikan operator melihat pemberi isyarat; jangan melintas di belakang kendaraan atau berdiri di bawah bagian yang terangkat. Jika orang masuk zona bongkar, hentikan pekerjaan dengan cara yang diajarkan dan beri tahu operator.

Kasus latihan: Seorang pengemudi mulai membuka bak, sementara pekerja masih berada di belakang truk untuk membaca label.
Respons yang dianjurkan: Hentikan bongkar sesuai prosedur, minta pekerja mundur ke zona aman, dan lanjutkan hanya setelah operator melihat area kosong.`}: `Use the designated pedestrian routes and waiting zones. Make sure the operator can see the signaler; do not walk behind a vehicle or stand beneath a raised part. If someone enters the unloading zone, stop work using the method you were taught and alert the operator.

Practice scenario: A driver starts raising the dump bed while a worker is still behind the truck reading a label.
Suggested response: Stop unloading according to procedure, ask the worker to move back to the safe zone, and resume only after the operator confirms the area is clear.`,
	{field: "ssw.body", source: `Hasil proses dapat menjadi bahan baku untuk tahap lain, produk yang dapat digunakan kembali, atau residu yang memerlukan jalur berbeda. Catatan fasilitas menghubungkan berat dan identitas bahan masuk dengan hasil serta tujuan berikutnya. Format dan data yang dicatat mengikuti prosedur setempat; pekerja tidak menebak kategori atau tujuan.`}: `Processed material may become feedstock for another stage, a reusable product, or residue that needs a different route. Facility records link incoming material weight and identity to the output and its next destination. Follow local procedures for record formats and data; do not guess the category or destination.`,
	{field: "ssw.body", source: `Letakkan wadah pada area bertanda, jaga jalur keluar tetap kosong, dan periksa penutup serta kebocoran. Bila debu berlebihan atau cairan keluar dari wadah, beri peringatan, jaga jarak, dan laporkan lokasi serta jenis yang diketahui. Gunakan alat pelindung dan prosedur hanya jika tugas itu diberikan kepada Anda.

Kasus latihan: Cairan keluar dari drum penyimpanan di dekat saluran lantai dan pekerja tidak mengetahui isi drum.
Respons yang dianjurkan: Jangan menyentuh atau menyiram cairan. Jauhkan orang, tandai area sesuai aturan, dan hubungi petugas berwenang.`}: `Place containers in marked areas, keep exit routes clear, and check lids and for leaks. If there is excessive dust or liquid escaping, warn others, keep your distance, and report the location and any known material type. Use protective equipment and procedures only if assigned to the task.

Practice scenario: Liquid is leaking from a storage drum near a floor drain, and the worker does not know what the drum contains.
Suggested response: Do not touch or wash away the liquid. Keep people back, mark the area as required, and contact authorized staff.`,
	{field: "ssw.body", source: `Material hasil pemilahan dapat dipadatkan, dibungkus, atau disiapkan untuk tahap pengolahan berikutnya. Wadah dan label menjaga identitas fraksi, jumlah, dan proses tujuan. Bila fraksi bercampur atau tanda hilang, kelompok perlu ditahan agar tidak salah masuk ke mesin atau tertukar saat pengiriman internal.`}:                                                        `Sorted material may be compacted, wrapped, or prepared for the next processing stage. Containers and labels preserve the fraction identity, quantity, and destination process. If fractions are mixed or a marking is missing, hold the group so it is not fed into the wrong machine or switched during internal transfer.`,
	{field: "ssw.body", source: `Muatan dapat mengandung benda tajam, baterai, tabung, cairan, bahan kimia, atau benda tertutup yang tidak sesuai dengan keterangan awal. Bahan tidak dikenal tidak aman untuk dibuka atau dipilah dengan tangan. Kriteria dan jalur penanganan ditetapkan fasilitas; pekerja menghentikan pekerjaan, menjaga orang lain menjauh, dan meminta petugas yang terlatih menilai.`}: `A load may contain sharp objects, batteries, cylinders, liquids, chemicals, or sealed items that do not match the initial description. It is unsafe to open or hand-sort unidentified material. The facility sets handling criteria and routes; workers must stop, keep others back, and request an assessment by trained staff.`,
	{field: "ssw.body", source: `Pekerjaan fasilitas berlangsung lintas proses dan giliran. Serah terima yang jelas menyebut lokasi, nomor lot, status mesin, jumlah yang dicatat, serta tindakan berikutnya. Catatan faktual membantu melanjutkan pekerjaan tanpa mencampur keluaran atau mengulang langkah yang sudah selesai. Bahaya yang masih aktif harus diserahkan langsung.`}:                          `Facility work spans processes and shifts. A clear handover states the location, lot number, machine status, recorded quantity, and next action. Factual records help the next team continue without mixing output or repeating completed steps. Hand over active hazards directly.`,
	{field: "ssw.body", source: `Pemilahan mengarahkan bahan ke proses yang sesuai dan mencegah kontaminasi antarfraksi. Kategori tidak ditentukan hanya dari warna atau bentuk; label, daftar penerimaan, contoh, dan instruksi lini menjadi rujukan. Bahan campuran yang belum dapat dipilah dengan aman tetap diberi identitas dan ditangani menurut aturan tempat kerja.`}:                                 `Sorting directs material to the appropriate process and prevents cross-contamination between fractions. Do not classify material by color or shape alone; use labels, acceptance lists, examples, and line instructions. Identify mixed material that cannot yet be sorted safely and handle it according to workplace rules.`,
	{field: "ssw.body", source: `Pengolahan antara mengubah bentuk atau kondisi bahan agar dapat dipulihkan atau diproses selanjutnya. Setiap fasilitas menetapkan jalur dan parameter untuk mesin tertentu, termasuk urutan material masuk dan keluaran. Catatan aktual membantu menghubungkan lot penerimaan dengan fraksi hasil dan menemukan penyimpangan.`}:                                               `Intermediate processing changes material's form or condition so it can be recovered or processed further. Each facility sets routes and parameters for its machines, including the sequence of incoming material and output. Actual records link incoming lots to output fractions and help identify deviations.`,
	{field: "ssw.body", source: `Peralatan dapat mengalami keausan, sumbatan, kebocoran, atau kerusakan listrik. Pekerja mengamati tanda yang ditetapkan dan menghentikan penggunaan bila ada bahaya, lalu memberi tahu petugas perawatan. Pembersihan, pelumasan, penggantian bagian, dan isolasi energi hanya dilakukan oleh orang yang berwenang mengikuti aturan fasilitas.`}:                              `Equipment can wear out, clog, leak, or develop electrical faults. Watch for designated warning signs, stop using equipment if there is a hazard, and notify maintenance staff. Cleaning, lubrication, part replacement, and energy isolation are restricted to authorized people following facility rules.`,
	{field: "ssw.body", source: `Peralatan pengolahan dapat memotong, menghancurkan, memadatkan, atau memisahkan material. Pemeriksaan awal memastikan pelindung terpasang, area tidak terhalang, dan alat ukur atau indikator berada dalam kondisi kerja. Pekerja menjalankan hanya peralatan yang telah diajarkan; pembersihan, perbaikan, dan isolasi energi mengikuti prosedur dan kewenangan fasilitas.`}: `Processing equipment may cut, crush, compact, or separate material. Pre-operation checks confirm guards are fitted, the area is clear, and measuring devices or indicators are working. Operate only equipment you have been trained to use; cleaning, repairs, and energy isolation must follow facility procedures and authorization.`,
	{field: "ssw.body", source: `Periksa daftar akhir, kembalikan alat pada lokasi, dan tandai lot yang masih menunggu hasil. Beri tahu penerima tentang kerusakan, tanda yang hilang, atau area yang dibatasi. Jangan menulis pekerjaan selesai jika belum dilakukan; minta penerima mengulang informasi penting untuk memastikan pemahaman.

Kasus latihan: Pada akhir giliran, satu wadah hasil masih menunggu verifikasi lot dan jalur di dekat mesin diberi tanda batas.
Respons yang dianjurkan: Pertahankan tanda tahan dan batas jalur, catat nomor wadah serta alasan, lalu serahkan langsung kepada penerima giliran dan pemimpin.`}: `Complete the closing checklist, return tools to their locations, and mark lots still awaiting results. Tell the next person about damage, missing labels, or restricted areas. Do not mark unfinished work as complete; ask the recipient to repeat key information to confirm understanding.

Practice scenario: At the end of the shift, one output container is still awaiting lot verification, and the route near the machine has been marked off.
Suggested response: Keep the hold marker and route boundary in place, record the container number and reason, and hand over directly to the next shift and leader.`,
	{field: "ssw.body", source: `Periksa label dan kartu proses sebelum pemindahan, timbang atau hitung sesuai instruksi, lalu catat hasil aktual. Bila jumlah tidak cocok atau tanda tujuan hilang, tahan hasil di lokasi yang ditetapkan. Jangan mengisi dokumen dengan angka target atau memindahkan bahan ke area lain tanpa izin.

Kasus latihan: Berat keluaran pada alat timbang berbeda dari catatan, dan label tujuan pada wadah terlepas.
Respons yang dianjurkan: Tahan wadah, catat hasil timbang aktual dan informasi yang masih ada, lalu minta penanggung jawab memulihkan tujuan sebelum pengiriman internal.`}: `Before transfer, check the label and process card, weigh or count as instructed, and record the actual result. If the quantity does not match or the destination marking is missing, hold the output in the designated location. Do not enter target figures in the records or move material elsewhere without permission.

Practice scenario: The output weight on the scale differs from the record, and the container's destination label has come off.
Suggested response: Hold the container, record the actual weight and any remaining information, and ask the person in charge to restore its destination before internal transfer.`,
	{field: "ssw.body", source: `Periksa wadah kosong, label, dan tujuan sebelum mengisinya. Catat waktu dan lot sesuai format fasilitas, jaga tutup atau ikatan yang ditentukan, dan pindahkan menggunakan jalur aman. Jika material salah masuk, jangan menuang balik tanpa persetujuan; tandai wadah dan laporkan jumlah perkiraan serta jenis campuran.

Kasus latihan: Wadah logam berisi sejumlah kecil kaca yang masuk dari aliran sebelah; label masih menunjukkan logam bersih.
Respons yang dianjurkan: Hentikan pengisian dan tahan wadah; catat campuran serta minta keputusan apakah perlu sortir ulang atau penanganan lain.`}: `Check the empty container, label, and destination before filling it. Record the time and lot in the facility's format, secure the specified lid or tie, and use a safe route for transfer. If the wrong material is added, do not pour it back without approval; mark the container and report the estimated amount and type of mixture.

Practice scenario: A metal container has a small amount of glass from the adjacent stream in it; the label still says clean metal.
Suggested response: Stop filling and hold the container. Record the contamination and ask whether it needs to be resorted or handled another way.`,
	{field: "ssw.body", source: `Sebelum bongkar, cocokkan dokumen, nomor kendaraan atau wadah, jenis yang dinyatakan, dan berat sesuai alat timbang. Catat hasil aktual serta waktu. Jika jenis atau jumlah berbeda, tahan penerimaan atau pindahkan ke area tunggu yang ditetapkan dan minta penanggung jawab memutuskan.

Kasus latihan: Dokumen menyebut logam bersih, tetapi muatan tampak bercampur dengan wadah tertutup tanpa label.
Respons yang dianjurkan: Jangan lanjutkan bongkar biasa. Tahan kendaraan di lokasi aman, catat perbedaan, dan minta pemeriksaan penanggung jawab.`}: `Before unloading, match the documents, vehicle or container number, declared material type, and weight from the scale. Record the actual result and time. If the type or quantity differs, hold acceptance or move it to the designated waiting area and ask the person in charge to decide.

Practice scenario: The documents say clean metal, but the load appears mixed with a sealed, unlabeled container.
Suggested response: Do not continue normal unloading. Hold the vehicle in a safe location, record the discrepancy, and request an inspection by the person in charge.`,
	{field: "ssw.body", source: `Sektor ini berfokus pada pemrosesan antara limbah untuk daur ulang atau pengolahan yang tepat, bukan semua jenis pekerjaan pengumpulan. Tiap fasilitas memiliki kriteria penerimaan dan format catatan sendiri. Identitas lot membantu menghubungkan bahan yang datang dengan pemilahan, proses, dan keluaran berikutnya.`}: `This sector focuses on intermediate waste processing for recycling or appropriate treatment, not every type of collection work. Each facility has its own acceptance criteria and record format. Lot identity links incoming material to sorting, processing, and subsequent output.`,

	{field: "ssw.step", source: `Amati material sebelum memasukkannya`}:                             `Check the material before feeding it into the process`,
	{field: "ssw.step", source: `Amati penutup, kebocoran, debu, dan tumpahan`}:                     `Check lids, leaks, dust, and spills`,
	{field: "ssw.step", source: `Amati tanda kerusakan dari posisi aman`}:                           `Observe signs of damage from a safe position`,
	{field: "ssw.step", source: `Baca dokumen serta kriteria penerimaan fasilitas`}:                 `Read the documents and facility acceptance criteria`,
	{field: "ssw.step", source: `Baca instruksi dan tugas operator hari ini`}:                       `Review today's instructions and operator assignments`,
	{field: "ssw.step", source: `Baca label aliran dan kriteria yang berlaku`}:                      `Check stream labels and applicable criteria`,
	{field: "ssw.step", source: `Batasi area serta laporkan kondisi yang tidak aman`}:               `Restrict the area and report unsafe conditions`,
	{field: "ssw.step", source: `Catat berat, waktu, lot, dan tujuan aktual`}:                       `Record the actual weight, time, lot, and destination`,
	{field: "ssw.step", source: `Catat ciri yang terlihat tanpa membuka atau memindahkan`}:          `Record visible features without opening or moving the item`,
	{field: "ssw.step", source: `Catat fakta, tindakan, dan pekerjaan yang belum selesai`}:          `Record facts, actions, and unfinished work`,
	{field: "ssw.step", source: `Catat lot, jumlah, dan lokasi wadah`}:                              `Record the lot, quantity, and container location`,
	{field: "ssw.step", source: `Catat mesin, waktu, dan gejala yang terlihat`}:                     `Record the machine, time, and visible symptoms`,
	{field: "ssw.step", source: `Catat waktu, hasil, dan identitas keluaran`}:                       `Record the time, result, and output identity`,
	{field: "ssw.step", source: `Cocokkan identitas kendaraan, wadah, dan lot`}:                     `Verify the vehicle, container, and lot identities`,
	{field: "ssw.step", source: `Cocokkan label hasil dengan kartu proses`}:                         `Match the output label to the process card`,
	{field: "ssw.step", source: `Cocokkan lot bahan dengan instruksi proses`}:                       `Match the material lot to the process instructions`,
	{field: "ssw.step", source: `Gunakan alat bantu yang disediakan`}:                               `Use the provided aids`,
	{field: "ssw.step", source: `Hentikan alat sesuai instruksi bila kondisi berbahaya`}:            `Stop the equipment as instructed if conditions are hazardous`,
	{field: "ssw.step", source: `Hentikan kontak dengan benda yang tidak dikenal`}:                  `Stop handling the unidentified item`,
	{field: "ssw.step", source: `Hentikan proses dan laporkan orang atau muatan di zona berbahaya`}: `Stop the process and report anyone or any load in a hazardous zone`,
	{field: "ssw.step", source: `Hentikan serta laporkan sumbatan atau bahaya`}:                     `Stop work and report blockages or hazards`,
	{field: "ssw.step", source: `Hubungi penanggung jawab dan tunggu arahan`}:                       `Contact the person in charge and wait for instructions`,
	{field: "ssw.step", source: `Ikuti urutan dan batas tugas operator`}:                            `Follow the sequence and limits of the operator's task`,
	{field: "ssw.step", source: `Isi hanya dengan fraksi yang telah dikonfirmasi`}:                  `Fill it only with the confirmed fraction`,
	{field: "ssw.step", source: `Jaga jalur kendaraan dan evakuasi tetap terbuka`}:                  `Keep vehicle routes and evacuation routes clear`,
	{field: "ssw.step", source: `Jaga jarak dari kendaraan dan muatan saat bergerak`}:               `Keep clear of moving vehicles and loads`,
	{field: "ssw.step", source: `Jaga jarak dari rol dan titik jepit`}:                              `Keep clear of rollers and pinch points`,
	{field: "ssw.step", source: `Jalankan hanya setelah pemeriksaan dinyatakan siap`}:               `Start only after the inspection confirms the equipment is ready`,
	{field: "ssw.step", source: `Jauhkan orang dari area sesuai aturan fasilitas`}:                  `Keep people away from the area as required by facility rules`,
	{field: "ssw.step", source: `Laporkan kepada petugas perawatan atau pemimpin`}:                  `Report it to maintenance staff or the leader`,
	{field: "ssw.step", source: `Laporkan kerusakan atau tanda tidak normal`}:                       `Report damage or abnormal signs`,
	{field: "ssw.step", source: `Pastikan lot serta mesin memiliki status yang terbaca`}:            `Confirm the lot and machine have a readable status`,
	{field: "ssw.step", source: `Pastikan material berada pada zona simpan yang benar`}:             `Confirm the material is in the correct storage zone`,
	{field: "ssw.step", source: `Pastikan pemberi isyarat serta operator siap`}:                     `Confirm the signaler and operator are ready`,
	{field: "ssw.step", source: `Periksa daftar penutupan dan kelengkapan alat`}:                    `Check the closeout list and account for all tools`,
	{field: "ssw.step", source: `Periksa pelindung, lantai, dan indikator yang ditetapkan`}:         `Check guards, the floor, and required indicators`,
	{field: "ssw.step", source: `Periksa posisi kerja dan pelindung sebelum lini berjalan`}:         `Check the work position and guards before starting the line`,
	{field: "ssw.step", source: `Periksa tanda jalur dan zona tunggu`}:                              `Check route markers and waiting zones`,
	{field: "ssw.step", source: `Siapkan wadah dan label tujuan yang benar`}:                        `Prepare the container and correct destination label`,
	{field: "ssw.step", source: `Tahan dan laporkan bahan yang tidak dikenali`}:                     `Hold and report unidentified material`,
	{field: "ssw.step", source: `Tahan hasil yang menyimpang dan minta pemeriksaan`}:                `Hold nonconforming output and request an inspection`,
	{field: "ssw.step", source: `Tahan selisih dan minta keputusan penanggung jawab`}:               `Hold the discrepancy and ask the person in charge to decide`,
	{field: "ssw.step", source: `Tahan selisih serta minta konfirmasi sebelum pindah`}:              `Hold the discrepancy and request confirmation before moving it`,
	{field: "ssw.step", source: `Tahan serta laporkan fraksi yang tercampur`}:                       `Hold and report mixed fractions`,
	{field: "ssw.step", source: `Tempatkan pada wadah atau jalur yang ditetapkan`}:                  `Place it in the designated container or stream`,
	{field: "ssw.step", source: `Timbang atau hitung sesuai prosedur dan catat hasil aktual`}:       `Weigh or count according to procedure and record the actual result`,
	{field: "ssw.step", source: `Ukur atau hitung hasil sesuai satuan fasilitas`}:                   `Measure or count output using the facility's units`,

	{field: "ssw.meaning", source: `Ada kaca yang tercampur. Saya menahan wadah ini.`}:                      `Glass is mixed in. I am holding this container.`,
	{field: "ssw.meaning", source: `Ada orang di belakang. Mohon hentikan kendaraan.`}:                      `There is someone behind you. Please stop the vehicle.`,
	{field: "ssw.meaning", source: `alarm`}:                                                                 `alarm`,
	{field: "ssw.meaning", source: `Alarm muncul. Saya tidak menjalankan mesin.`}:                           `An alarm has appeared. I will not start the machine.`,
	{field: "ssw.meaning", source: `ban berjalan`}:                                                          `conveyor`,
	{field: "ssw.meaning", source: `Baterai ini menggembung. Jangan sentuh.`}:                               `This battery is swollen. Do not touch it.`,
	{field: "ssw.meaning", source: `Berat berbeda dari catatan.`}:                                           `The weight differs from the record.`,
	{field: "ssw.meaning", source: `bongkar muatan`}:                                                        `unloading`,
	{field: "ssw.meaning", source: `Cairan bocor dari drum.`}:                                               `Liquid is leaking from the drum.`,
	{field: "ssw.meaning", source: `getaran`}:                                                               `vibration`,
	{field: "ssw.meaning", source: `Jangan masukkan tangan. Saya menghentikan mesin.`}:                      `Do not put your hand in. I am stopping the machine.`,
	{field: "ssw.meaning", source: `Jangan sentuh, hubungi petugas.`}:                                       `Do not touch it; contact staff.`,
	{field: "ssw.meaning", source: `jenis`}:                                                                 `type`,
	{field: "ssw.meaning", source: `Jenis muatan berbeda dari dokumen.`}:                                    `The load type differs from the documents.`,
	{field: "ssw.meaning", source: `Kain tersangkut di rol.`}:                                               `Cloth is caught in the rollers.`,
	{field: "ssw.meaning", source: `kebocoran atau pelepasan`}:                                              `leak or release`,
	{field: "ssw.meaning", source: `kendaraan`}:                                                             `vehicle`,
	{field: "ssw.meaning", source: `label atau tanda`}:                                                      `label or marking`,
	{field: "ssw.meaning", source: `lot atau kelompok`}:                                                     `lot or group`,
	{field: "ssw.meaning", source: `Mohon tunggu sampai pemeriksaan selesai.`}:                              `Please wait until the inspection is complete.`,
	{field: "ssw.meaning", source: `Nomor lot berbeda dari catatan.`}:                                       `The lot number differs from the record.`,
	{field: "ssw.meaning", source: `pemindahan keluar`}:                                                     `outbound transfer`,
	{field: "ssw.meaning", source: `pemindahan material`}:                                                   `material transfer`,
	{field: "ssw.meaning", source: `pemisahan kategori`}:                                                    `category separation`,
	{field: "ssw.meaning", source: `pencampuran yang tidak diinginkan`}:                                     `unwanted mixing`,
	{field: "ssw.meaning", source: `pencocokan`}:                                                            `matching`,
	{field: "ssw.meaning", source: `pengeluaran atau pemindahan keluar`}:                                    `dispatch or outbound transfer`,
	{field: "ssw.meaning", source: `pengolahan antara`}:                                                     `intermediate processing`,
	{field: "ssw.meaning", source: `pengoperasian`}:                                                         `operation`,
	{field: "ssw.meaning", source: `perawatan`}:                                                             `maintenance`,
	{field: "ssw.meaning", source: `Pompa mengeluarkan bau terbakar.`}:                                      `The pump smells burnt.`,
	{field: "ssw.meaning", source: `saluran pembuangan`}:                                                    `drain`,
	{field: "ssw.meaning", source: `Saya akan bertanya kepada penanggung jawab sebelum memperbaiki label.`}: `I will ask the person in charge before correcting the label.`,
	{field: "ssw.meaning", source: `Saya berhenti di tempat aman dan melapor kepada penanggung jawab.`}:     `I have stopped in a safe place and will report it to the person in charge.`,
	{field: "ssw.meaning", source: `Saya juga menyerahkan informasi area terlarang.`}:                       `I will also hand over the information about the restricted area.`,
	{field: "ssw.meaning", source: `Saya menahan keluaran dan meminta konfirmasi petugas.`}:                 `I am holding the output and requesting staff confirmation.`,
	{field: "ssw.meaning", source: `Saya menghentikan operasi dan menghubungi petugas perawatan.`}:          `I am stopping the operation and contacting maintenance staff.`,
	{field: "ssw.meaning", source: `Saya menghentikan pekerjaan dan memanggil penanggung jawab.`}:           `I am stopping work and calling the person in charge.`,
	{field: "ssw.meaning", source: `Saya mundur ke zona aman.`}:                                             `I am moving back to the safe zone.`,
	{field: "ssw.meaning", source: `Saya pisahkan dan meminta konfirmasi penanggung jawab.`}:                `I will set it aside and ask the person in charge to confirm.`,
	{field: "ssw.meaning", source: `Saya tidak memindahkannya sampai tujuannya diketahui.`}:                 `I will not move it until its destination is known.`,
	{field: "ssw.meaning", source: `status`}:                                                                `status`,
	{field: "ssw.meaning", source: `tahap proses`}:                                                          `process stage`,
	{field: "ssw.meaning", source: `tertarik masuk mesin`}:                                                  `caught in machinery`,
	{field: "ssw.meaning", source: `Wadah ini sedang menunggu pemeriksaan lot.`}:                            `This container is awaiting a lot inspection.`,
	{field: "ssw.meaning", source: `Wadah ini tidak ada dalam kriteria pemilahan.`}:                         `This container is not included in the sorting criteria.`,
	{field: "ssw.meaning", source: `zona aman`}:                                                             `safe zone`,
}
