package api

// sswCurriculumFoodManufacturingEnglish contains exact, hand-reviewed English
// translations for food-manufacturing SSW learning material.
var sswCurriculumFoodManufacturingEnglish = map[examSSWEnglishKey]string{
	{field: "ssw.title", source: `Catatan produksi dan serah terima lot`}:         `Production Records and Lot Handover`,
	{field: "ssw.title", source: `Cuci tangan dan penggantian sarung tangan`}:     `Handwashing and Glove Changes`,
	{field: "ssw.title", source: `Kendali proses waktu dan suhu`}:                 `Time and Temperature Process Controls`,
	{field: "ssw.title", source: `Pembersihan alat dan pemisahan zona`}:           `Equipment Cleaning and Zone Separation`,
	{field: "ssw.title", source: `Pemeriksaan awal mesin dan pelindung`}:          `Pre-Operation Machine and Guard Checks`,
	{field: "ssw.title", source: `Pemeriksaan bahan yang datang`}:                 `Incoming Material Checks`,
	{field: "ssw.title", source: `Pemeriksaan cacat dan benda asing`}:             `Checking for Defects and Foreign Objects`,
	{field: "ssw.title", source: `Pemeriksaan kemasan dan label produk`}:          `Product Packaging and Label Checks`,
	{field: "ssw.title", source: `Pemeriksaan kesehatan dan masuk area produksi`}: `Health Checks and Entry to the Production Area`,
	{field: "ssw.title", source: `Penimbangan resep dan urutan bahan`}:            `Recipe Weighing and Ingredient Sequence`,
	{field: "ssw.title", source: `Penyimpanan dan pemantauan suhu`}:               `Storage and Temperature Monitoring`,
	{field: "ssw.title", source: `Pergantian produk dan pengendalian alergen`}:    `Product Changeovers and Allergen Control`,

	{field: "ssw.description", source: `Membangun kebiasaan higienis sejak masuk area produksi sampai pergantian pekerjaan, serta mencegah kontaminasi silang melalui pemisahan zona, alat, dan alur bahan.`}: `Build hygienic habits from entry into the production area through task changes, and prevent cross-contamination by separating zones, equipment, and material flows.`,
	{field: "ssw.description", source: `Memeriksa identitas, kondisi, jumlah, penyimpanan, dan alur bahan agar lot serta resep tetap cocok sepanjang proses.`}:                                                `Check material identity, condition, quantity, storage, and flow so lots and recipes remain matched throughout processing.`,
	{field: "ssw.description", source: `Memeriksa produk sebelum dikemas dan dikirim, menjaga kecocokan informasi, serta mencatat ketidaksesuaian agar dapat ditelusuri.`}:                                    `Inspect products before packaging and dispatch, keep information consistent, and record discrepancies for traceability.`,
	{field: "ssw.description", source: `Mengikuti urutan proses, batas kendali, perlindungan mesin, serta pergantian produk dengan fokus pada penyimpangan dan alergen.`}:                                     `Follow process sequences, control limits, machine safeguards, and product changeovers, with attention to deviations and allergens.`,

	{field: "ssw.objective", source: `Melakukan cuci tangan pada saat yang ditetapkan dan mengganti sarung tangan yang terkontaminasi.`}:                    `Wash hands at designated times and change contaminated gloves.`,
	{field: "ssw.objective", source: `Membaca batas proses yang berlaku, memantau waktu atau suhu, dan menahan produk ketika kendali menyimpang.`}:          `Read the applicable process limits, monitor time or temperature, and hold product when controls deviate.`,
	{field: "ssw.objective", source: `Memeriksa kondisi mesin, tombol berhenti, pelindung, dan area kerja sebelum menjalankan alat yang menjadi tugasnya.`}: `Check machine condition, stop controls, guards, and the work area before operating assigned equipment.`,
	{field: "ssw.objective", source: `Memeriksa kondisi tubuh, pakaian kerja, rambut, dan perlengkapan sebelum menangani pangan.`}:                          `Check your health, work clothing, hair covering, and equipment before handling food.`,
	{field: "ssw.objective", source: `Mencatat status proses, jumlah, penyimpangan, dan tindak lanjut agar giliran berikutnya dapat bekerja aman.`}:         `Record process status, quantities, deviations, and follow-up so the next shift can work safely.`,
	{field: "ssw.objective", source: `Mencegah perpindahan alergen saat pergantian resep dengan memeriksa bahan, pembersihan, alat, dan label.`}:            `Prevent allergen transfer during recipe changeovers by checking ingredients, cleaning, equipment, and labels.`,
	{field: "ssw.objective", source: `Mencocokkan kode produk, lot, tanggal, label alergen, jumlah, dan keadaan kemasan sebelum pelepasan.`}:                `Verify the product code, lot, date, allergen label, quantity, and package condition before release.`,
	{field: "ssw.objective", source: `Mencocokkan nama bahan, nomor lot, jumlah, tanggal, kemasan, dan kondisi penerimaan dengan dokumen kerja.`}:           `Match ingredient names, lot numbers, quantities, dates, packaging, and receiving conditions to the work documents.`,
	{field: "ssw.objective", source: `Menempatkan bahan pada lokasi dan kondisi simpan yang ditetapkan serta mencatat suhu sesuai rencana.`}:                `Store ingredients in the designated location and conditions, and record temperatures according to plan.`,
	{field: "ssw.objective", source: `Mengenali cacat atau benda asing, memisahkan produk terdampak, dan melaporkan sesuai titik kendali fasilitas.`}:       `Recognize defects or foreign objects, separate affected products, and report them at the facility's control points.`,
	{field: "ssw.objective", source: `Menimbang bahan dengan satuan resep yang benar dan menjaga urutan penambahan serta identitas kelompok campuran.`}:     `Weigh ingredients using the recipe's units and maintain the addition sequence and batch identity.`,
	{field: "ssw.objective", source: `Menjaga meja, pisau, wadah, dan alat tetap sesuai status bersih atau belum dibersihkan.`}:                             `Keep tables, knives, containers, and equipment in their designated clean or unclean status.`,

	{field: "ssw.caution", source: `Jangan masuk ke area penanganan pangan dengan gejala atau luka yang belum dilaporkan.`}:                    `Do not enter a food-handling area with symptoms or an unreported wound.`,
	{field: "ssw.caution", source: `Jangan melanjutkan lot dengan asumsi hasilnya aman setelah suhu kembali normal.`}:                          `Do not continue a lot on the assumption that it is safe after the temperature returns to normal.`,
	{field: "ssw.caution", source: `Jangan melepas produk dengan label yang belum dicocokkan pada awal lot atau setelah pergantian gulungan.`}: `Do not release product with labels that have not been verified at the start of a lot or after a roll change.`,
	{field: "ssw.caution", source: `Jangan memakai angka suhu hafalan sebagai pengganti batas yang ditetapkan untuk produk dan proses.`}:       `Do not rely on a remembered temperature instead of the limit specified for the product and process.`,
	{field: "ssw.caution", source: `Jangan memakai sarung tangan yang sama setelah menyentuh sampah, telepon, atau bahan mentah.`}:             `Do not keep using the same gloves after touching waste, a phone, or raw ingredients.`,
	{field: "ssw.caution", source: `Jangan memasukkan bahan dengan identitas hilang atau kondisi di luar kriteria tanpa persetujuan.`}:         `Do not use ingredients with missing identity or conditions outside the criteria without approval.`,
	{field: "ssw.caution", source: `Jangan membuka pelindung atau memasukkan tangan ke mesin yang belum diisolasi sesuai prosedur.`}:           `Do not open a guard or put your hand into a machine that has not been isolated according to procedure.`,
	{field: "ssw.caution", source: `Jangan menandai lot selesai atau menghapus status tahan sebelum ada konfirmasi berwenang.`}:                `Do not mark a lot complete or remove its hold status before receiving authorization.`,
	{field: "ssw.caution", source: `Jangan mengandalkan ingatan atau hanya mengganti sarung tangan untuk menyatakan lini bebas alergen.`}:      `Do not rely on memory or glove changes alone to declare a line free of allergens.`,
	{field: "ssw.caution", source: `Jangan menganggap alat siap pakai hanya karena permukaannya terlihat kering.`}:                             `Do not assume equipment is ready to use just because its surface looks dry.`,
	{field: "ssw.caution", source: `Jangan mengira jumlah atau mengganti bahan tanpa persetujuan formula.`}:                                    `Do not estimate quantities or substitute ingredients without recipe approval.`,
	{field: "ssw.caution", source: `Jangan menyingkirkan benda asing lalu mengirim produk tanpa penilaian penanggung jawab.`}:                  `Do not remove a foreign object and ship the product without an assessment by the person in charge.`,

	{field: "ssw.scenario", source: `Catatan pemanasan menunjukkan pembacaan belum mencapai batas pada instruksi produk, tetapi baki berikutnya sudah menunggu.`}:       `Heating records show a reading below the limit in the product instructions, but the next trays are waiting.`,
	{field: "ssw.scenario", source: `Catatan ruang pendingin menunjukkan suhu di atas batas kerja pada formulir, tetapi satu lot akan segera dipakai.`}:                 `Cold-room records show a temperature above the form's operating limit, but a lot is about to be used.`,
	{field: "ssw.scenario", source: `Dokumen mencatat 12 kotak bahan, tetapi satu kotak bocor dan nomor lotnya berbeda dari daftar penerimaan.`}:                        `The documents list 12 ingredient boxes, but one box is leaking and its lot number differs from the receiving list.`,
	{field: "ssw.scenario", source: `Kode label pada gulungan baru berbeda satu angka dari lembar produksi; beberapa bungkus sudah dibuat.`}:                            `The label code on a new roll differs by one digit from the production sheet, and several packages have already been made.`,
	{field: "ssw.scenario", source: `Lembar formula meminta bubuk bahan A, tetapi wadah di meja berlabel bahan B dan pekerja belum mulai menimbang.`}:                   `The recipe calls for ingredient A powder, but the container on the table is labeled ingredient B, and weighing has not started.`,
	{field: "ssw.scenario", source: `Pada akhir giliran, satu lot masih menunggu pemeriksaan label, tetapi pekerja berikutnya hanya menerima catatan jumlah produksi.`}: `At shift end, a lot is still awaiting a label check, but the next worker receives only the production quantity record.`,
	{field: "ssw.scenario", source: `Pekerja menyentuh gagang tempat sampah dengan sarung tangan, lalu hendak mengisi wadah produk siap makan.`}:                        `A worker touches a waste-bin handle with gloves and then prepares to fill a container with ready-to-eat product.`,
	{field: "ssw.scenario", source: `Pelindung pada pengaduk terasa longgar, tetapi lini sedang mengejar jumlah produksi.`}:                                             `A mixer guard feels loose, but the line is trying to meet its production target.`,
	{field: "ssw.scenario", source: `Sebelum giliran dimulai, seorang pekerja mengalami diare dan akan memakai sarung tangan untuk mengemas produk siap makan.`}:        `Before the shift begins, a worker has diarrhea and plans to wear gloves while packaging ready-to-eat product.`,
	{field: "ssw.scenario", source: `Seorang pekerja melihat serpihan plastik di dekat produk di konveyor dan tidak tahu apakah serpihan berasal dari kemasan.`}:        `A worker sees a plastic fragment near a product on the conveyor and does not know whether it came from the packaging.`,
	{field: "ssw.scenario", source: `Setelah produk yang memakai susu, gulungan label produk tanpa susu dipasang sebelum catatan pembersihan pergantian selesai.`}:      `After a product containing milk, a label roll for a milk-free product is installed before the changeover cleaning record is complete.`,
	{field: "ssw.scenario", source: `Wadah berlabel bersih jatuh ke lantai dekat area bahan mentah; pekerja ingin segera menggunakannya kembali.`}:                      `A container marked clean falls on the floor near the raw-ingredient area, and a worker wants to reuse it immediately.`,

	{field: "ssw.response", source: `Hentikan atau tahan aliran sesuai instruksi, pisahkan produk terkait, dan laporkan lokasi serta waktu temuan untuk pemeriksaan.`}:           `Stop or hold the line as instructed, separate the affected products, and report where and when the item was found for inspection.`,
	{field: "ssw.response", source: `Hentikan lini dan tahan produk sejak pergantian label; laporkan urutan kejadian dan tunggu pemeriksaan pembersihan serta label.`}:           `Stop the line and hold product from the label change onward; report the sequence of events and wait for cleaning and label checks.`,
	{field: "ssw.response", source: `Hentikan pelepasan lot, tahan produk yang terdampak, catat pembacaan dan waktu, lalu minta keputusan penanggung jawab.`}:                    `Stop lot release, hold affected products, record the reading and time, and ask the person in charge to decide what to do.`,
	{field: "ssw.response", source: `Hentikan pengemasan, tahan bungkus sejak titik pemeriksaan terakhir, catat jumlah dan kode, lalu laporkan untuk pemeriksaan ulang.`}:        `Stop packaging, hold packages made since the last check, record the quantity and code, and report them for reinspection.`,
	{field: "ssw.response", source: `Jangan lanjut mengisi. Buang sarung tangan, cuci tangan sesuai prosedur, dan kenakan sarung tangan baru sebelum kembali ke produk.`}:        `Do not continue filling. Remove the gloves, wash your hands as required, and put on new gloves before handling the product again.`,
	{field: "ssw.response", source: `Jangan menyalakan pengaduk. Jauhkan orang dari area berbahaya dan beri tahu pemimpin agar mesin diperiksa.`}:                                `Do not start the mixer. Keep people away from the danger area and tell the leader so the machine can be inspected.`,
	{field: "ssw.response", source: `Jangan mulai pencampuran. Pisahkan wadah, cocokkan identitas dengan daftar bahan, dan minta konfirmasi penanggung jawab.`}:                  `Do not start mixing. Set the container aside, match its identity to the ingredient list, and ask the person in charge to confirm.`,
	{field: "ssw.response", source: `Laporkan gejala kepada pemimpin dan jangan menangani pangan; tunggu penetapan tugas aman sesuai prosedur fasilitas.`}:                       `Report the symptoms to the leader and do not handle food; wait to be assigned safe duties under facility procedures.`,
	{field: "ssw.response", source: `Pisahkan kotak bermasalah, catat jumlah dan nomor lot yang terlihat, lalu minta penanggung jawab memutuskan penerimaan.`}:                   `Set the problem box aside, record the quantity and visible lot number, and ask the person in charge to decide whether to accept it.`,
	{field: "ssw.response", source: `Pisahkan wadah itu sebagai belum siap pakai dan minta pencucian serta sanitasi ulang sesuai aturan fasilitas.`}:                             `Set the container aside as not ready for use and request washing and sanitizing again according to facility rules.`,
	{field: "ssw.response", source: `Tahan lot yang terdampak, catat waktu dan pembacaan, lalu laporkan agar penanggung jawab menentukan penanganannya.`}:                        `Hold the affected lot, record the time and reading, and report it so the person in charge can determine how to handle it.`,
	{field: "ssw.response", source: `Tambahkan status lot, lokasi tahan, alasan dan pemeriksaan terbuka sesuai aturan koreksi, lalu serahkan langsung kepada penerima giliran.`}: `Record the lot status, hold location, reason, and outstanding check according to correction rules, then hand it over directly to the next shift.`,

	{field: "ssw.body", source: `Alergen dapat berpindah melalui sisa bahan di mesin, meja, sarung tangan, alat ukur, atau kemasan. Daftar alergen dan informasi label dapat berubah mengikuti ketentuan terkini; pekerja memakai spesifikasi produk dan matriks yang berlaku di fasilitas. Pergantian produksi memerlukan persetujuan, pembersihan sesuai prosedur, serta pemeriksaan ulang label dan identitas lot.`}: `Allergens can transfer through residue on machines, tables, gloves, measuring tools, or packaging. Allergen lists and label information may change under current requirements; use the product specifications and matrix in force at your facility. Production changeovers require approval, cleaning according to procedure, and another check of labels and lot identity.`,
	{field: "ssw.body", source: `Baca kode pada lembar kerja dan kemasan, lalu cocokkan dengan sampel label yang disetujui. Periksa tanggal cetak, keterbacaan, segel, dan arah label. Jika ada selisih, hentikan pengemasan, tahan produk sejak pemeriksaan terakhir yang sesuai, catat kode dan jumlah, kemudian minta penanggung jawab menentukan pemeriksaan ulang.

Kasus latihan: Kode label pada gulungan baru berbeda satu angka dari lembar produksi; beberapa bungkus sudah dibuat.
Respons yang dianjurkan: Hentikan pengemasan, tahan bungkus sejak titik pemeriksaan terakhir, catat jumlah dan kode, lalu laporkan untuk pemeriksaan ulang.`}: `Read the code on the work sheet and package, then compare it with the approved label sample. Check the print date, legibility, seal, and label orientation. If anything differs, stop packaging, hold product made since the last satisfactory check, record the code and quantity, and ask the person in charge to arrange a recheck.

Practice scenario: The label code on a new roll differs by one digit from the production sheet, and several packages have already been made.
Suggested response: Stop packaging, hold packages made since the last check, record the quantity and code, and report them for reinspection.`,
	{field: "ssw.body", source: `Bandingkan kondisi mesin dengan daftar pemeriksaan awal. Bila pelindung longgar, bunyi atau getaran tidak biasa, atau tombol tidak dapat diuji sesuai instruksi, jangan jalankan mesin. Pasang tanda bila diwajibkan dan panggil penanggung jawab atau teknisi. Jangan memasukkan tangan ke bagian bergerak untuk mengatasi sumbatan.

Kasus latihan: Pelindung pada pengaduk terasa longgar, tetapi lini sedang mengejar jumlah produksi.
Respons yang dianjurkan: Jangan menyalakan pengaduk. Jauhkan orang dari area berbahaya dan beri tahu pemimpin agar mesin diperiksa.`}: `Compare the machine with the startup checklist. Do not run it if a guard is loose, there is unusual noise or vibration, or a control cannot be tested as instructed. Post a sign if required and call the person in charge or a technician. Never put your hand near moving parts to clear a blockage.

Practice scenario: A mixer guard feels loose, but the line is trying to meet its production target.
Suggested response: Do not start the mixer. Keep people away from the danger area and tell the leader so the machine can be inspected.`,
	{field: "ssw.body", source: `Catatan produksi menghubungkan bahan, parameter, pemeriksaan, kemasan, dan hasil lot. Catatan harus berisi fakta aktual, waktu, identitas produk, serta tindakan yang dilakukan. Pada pergantian giliran, status barang tahan, mesin bermasalah, dan pemeriksaan yang belum selesai perlu disampaikan lisan dan tertulis menurut sistem fasilitas.`}: `Production records link ingredients, parameters, inspections, packaging, and lot results. Record actual facts, times, product identity, and actions taken. At shift change, communicate the status of held items, faulty machines, and unfinished checks verbally and in writing according to the facility's system.`,
	{field: "ssw.body", source: `Cocokkan nama bahan dan lot dengan lembar formula sebelum membuka kemasan. Timbang satu bahan pada satu waktu, catat nilai aktual, lalu tandai langkah yang selesai. Bila angka, label, atau jumlah tidak sesuai, hentikan pencampuran dan minta pemeriksaan kedua; jangan menambahkan bahan pengganti dari stok terdekat.

Kasus latihan: Lembar formula meminta bubuk bahan A, tetapi wadah di meja berlabel bahan B dan pekerja belum mulai menimbang.
Respons yang dianjurkan: Jangan mulai pencampuran. Pisahkan wadah, cocokkan identitas dengan daftar bahan, dan minta konfirmasi penanggung jawab.`}: `Before opening a package, match the ingredient name and lot to the recipe sheet. Weigh one ingredient at a time, record the actual value, and mark each completed step. If a number, label, or quantity does not match, stop mixing and request a second check; do not substitute a nearby stock item.

Practice scenario: The recipe calls for ingredient A powder, but the container on the table is labeled ingredient B, and weighing has not started.
Suggested response: Do not start mixing. Set the container aside, match its identity to the ingredient list, and ask the person in charge to confirm.`,
	{field: "ssw.body", source: `Cuci tangan sebelum mulai, setelah dari toilet, setelah menyentuh bahan mentah atau sampah, dan ketika berpindah dari tugas kotor ke tugas bersih. Lepas sarung tangan tanpa menyentuh pangan, buang sesuai aturan, cuci tangan, lalu gunakan pasangan baru. Ikuti pencatatan yang diminta.

Kasus latihan: Pekerja menyentuh gagang tempat sampah dengan sarung tangan, lalu hendak mengisi wadah produk siap makan.
Respons yang dianjurkan: Jangan lanjut mengisi. Buang sarung tangan, cuci tangan sesuai prosedur, dan kenakan sarung tangan baru sebelum kembali ke produk.`}: `Wash your hands before starting, after using the toilet, after touching raw ingredients or waste, and when moving from a dirty task to a clean one. Remove gloves without touching food, discard them as required, wash your hands, and put on a new pair. Complete any required records.

Practice scenario: A worker touches a waste-bin handle with gloves and then prepares to fill a container with ready-to-eat product.
Suggested response: Do not continue filling. Remove the gloves, wash your hands as required, and put on new gloves before handling the product again.`,
	{field: "ssw.body", source: `Gunakan alat ukur yang ditetapkan dan catat hasil aktual, waktu, serta nomor lot. Pisahkan kemasan rusak atau bahan dengan data tidak cocok di area tahan, tanpa menebak kelayakannya. Jangan mengubah angka dokumen atau memindahkan label dari lot lain.

Kasus latihan: Dokumen mencatat 12 kotak bahan, tetapi satu kotak bocor dan nomor lotnya berbeda dari daftar penerimaan.
Respons yang dianjurkan: Pisahkan kotak bermasalah, catat jumlah dan nomor lot yang terlihat, lalu minta penanggung jawab memutuskan penerimaan.`}: `Use the designated measuring tools and record actual results, time, and lot number. Separate damaged packages or ingredients with mismatched records in the hold area without guessing whether they are acceptable. Do not change document figures or move a label from another lot.

Practice scenario: The documents list 12 ingredient boxes, but one box is leaking and its lot number differs from the receiving list.
Suggested response: Set the problem box aside, record the quantity and visible lot number, and ask the person in charge to decide whether to accept it.`,
	{field: "ssw.body", source: `Isi formulir saat pekerjaan dilakukan, bukan dari ingatan di akhir giliran. Jangan menghapus nilai yang salah; ikuti cara koreksi resmi agar perubahan terlihat. Saat serah terima, sebutkan nomor lot, lokasi barang tahan, alasan, hasil yang sudah dicek, dan siapa yang perlu menindaklanjuti. Minta penerima mengulang poin penting bila belum jelas.

Kasus latihan: Pada akhir giliran, satu lot masih menunggu pemeriksaan label, tetapi pekerja berikutnya hanya menerima catatan jumlah produksi.
Respons yang dianjurkan: Tambahkan status lot, lokasi tahan, alasan dan pemeriksaan terbuka sesuai aturan koreksi, lalu serahkan langsung kepada penerima giliran.`}: `Complete forms as work is performed rather than from memory at shift end. Do not erase an incorrect value; use the official correction method so the change remains visible. During handover, state the lot number, hold location and reason, checks already completed, and who must follow up. Ask the recipient to repeat key points if anything is unclear.

Practice scenario: At shift end, a lot is still awaiting a label check, but the next worker receives only the production quantity record.
Suggested response: Record the lot status, hold location, reason, and outstanding check according to correction rules, then hand it over directly to the next shift.`,
	{field: "ssw.body", source: `Kemasan melindungi pangan; label menyampaikan identitas dan informasi penting kepada pengguna. Karena itu, gulungan label, kemasan, dan produk harus sesuai dengan rencana produksi. Salah label dapat membuat informasi produk dan alergen keliru walaupun isi kemasan benar. Pemeriksaan dilakukan saat mulai, saat bahan kemas berganti, dan pada titik lain yang ditetapkan fasilitas.`}: `Packaging protects food, and labels provide consumers with product identity and important information. Label rolls, packages, and products must therefore match the production plan. A wrong label can misstate product or allergen information even when the contents are correct. Check labels at startup, when packaging materials change, and at other points set by the facility.`,
	{field: "ssw.body", source: `Mesin pangan seperti pengaduk, pemotong, dan konveyor dapat menjepit atau melukai. Pemeriksaan sebelum kerja memastikan pelindung, tombol darurat, kabel, dan area bebas dari alat lepas. Pengoperasian dan pembersihan hanya dilakukan oleh pekerja yang telah mendapat instruksi; isolasi energi dan perbaikan mengikuti prosedur serta kewenangan setempat.`}:                             `Food-processing machines such as mixers, cutters, and conveyors can trap or injure workers. Before work, check guards, emergency stops, cables, and the area for loose tools. Only instructed workers may operate or clean equipment; energy isolation and repairs follow local procedures and authorization.`,
	{field: "ssw.body", source: `Pekerja dapat membawa cemaran lewat tangan, rambut, pakaian, atau luka. Pemeriksaan awal membantu menemukan demam, diare, muntah, luka, atau pakaian kotor sebelum produk terpapar. Ikuti rencana higiene fasilitas: pekerja melapor, lalu penanggung jawab menentukan tugas yang aman.`}:                                                                                                    `Workers can carry contaminants on their hands, hair, clothing, or wounds. An entry check can identify fever, diarrhea, vomiting, wounds, or dirty clothing before products are exposed. Follow the facility's hygiene plan: report the issue, and let the person in charge assign safe duties.`,
	{field: "ssw.body", source: `Pemanasan, pendinginan, pencampuran, dan proses lain memiliki parameter yang ditetapkan untuk produk tertentu. Pekerja memeriksa instruksi lot, memakai alat yang disyaratkan, dan mencatat hasil aktual pada titik pengawasan. Jika pembacaan di luar batas atau alat gagal, produk sejak pemeriksaan terakhir mungkin perlu ditahan untuk penilaian.`}:                                     `Heating, cooling, mixing, and other processes have parameters set for specific products. Check the lot instructions, use required instruments, and record actual results at monitoring points. If a reading is outside limits or an instrument fails, product made since the last check may need to be held for assessment.`,
	{field: "ssw.body", source: `Pembersihan mengangkat sisa pangan dan kotoran; sanitasi dilakukan menurut prosedur setelah tahap pembersihan yang diwajibkan. Alat yang terlihat bersih belum tentu siap dipakai. Penanda, lokasi simpan, dan jadwal fasilitas menunjukkan status alat serta batas antara area bahan mentah, proses, dan produk siap makan.`}:                                                               `Cleaning removes food residue and soil; sanitize according to procedure after the required cleaning steps. Equipment that looks clean is not necessarily ready to use. Facility markers, storage locations, and schedules show equipment status and boundaries between raw-material, processing, and ready-to-eat areas.`,
	{field: "ssw.body", source: `Pemeriksaan visual dan alat pemeriksa yang ditetapkan membantu menemukan kemasan rusak, perubahan warna, bagian yang hilang, atau benda asing. Kriteria penerimaan berbeda menurut jenis produk. Pekerja mengikuti sampel standar dan batas fasilitas, bukan menilai sendiri bahwa cacat aman. Perangkat pemeriksa tidak boleh dilewati atau disetel tanpa otorisasi.`}:                      `Visual checks and designated inspection devices help detect damaged packaging, discoloration, missing parts, or foreign objects. Acceptance criteria vary by product. Follow standard samples and facility limits; do not decide on your own that a defect is safe. Do not bypass or adjust inspection devices without authorization.`,
	{field: "ssw.body", source: `Penerimaan menghubungkan pemasok, bahan, dan lot dengan produk akhir. Periksa dokumen bersama tanda pada kemasan, keadaan segel, kebocoran, kerusakan, dan suhu bila bahan mensyaratkannya. Kriteria penerimaan berasal dari spesifikasi bahan dan rencana fasilitas; barang yang tidak cocok perlu ditahan untuk keputusan penanggung jawab.`}:                                              `Receiving links suppliers, ingredients, and lots to the finished product. Check documents against package markings, seal condition, leaks, damage, and temperature when required for the material. Acceptance criteria come from ingredient specifications and facility plans; hold nonconforming items for a decision by the person in charge.`,
	{field: "ssw.body", source: `Periksa label penyimpanan, gunakan rak dan wadah bertanda, lalu catat pembacaan alat ukur pada waktu yang ditentukan. Jika suhu melewati batas, pintu terbuka lama, atau alat ukur meragukan, tahan lot dan laporkan. Jangan mengubah setelan atau menetapkan sendiri bahwa bahan masih aman.

Kasus latihan: Catatan ruang pendingin menunjukkan suhu di atas batas kerja pada formulir, tetapi satu lot akan segera dipakai.
Respons yang dianjurkan: Tahan lot yang terdampak, catat waktu dan pembacaan, lalu laporkan agar penanggung jawab menentukan penanganannya.`}: `Check storage labels, use marked shelves and containers, and record instrument readings at designated times. If temperature exceeds a limit, a door has been left open, or an instrument is questionable, hold the lot and report it. Do not change settings or decide on your own that the ingredients are still safe.

Practice scenario: Cold-room records show a temperature above the form's operating limit, but a lot is about to be used.
Suggested response: Hold the affected lot, record the time and reading, and report it so the person in charge can determine how to handle it.`,
	{field: "ssw.body", source: `Resep produksi mengendalikan komposisi, hasil, mutu, dan informasi alergen. Timbangan harus sesuai satuan kerja, wadah tara diperhitungkan, dan bahan ditambahkan pada urutan yang ditetapkan. Pergantian formula atau substitusi bukan keputusan pekerja lini kecuali disetujui dan dicatat oleh pihak berwenang.`}: `The production recipe controls composition, yield, quality, and allergen information. Use scales suited to the required units, account for container tare, and add ingredients in the specified order. Line workers may not change a recipe or substitute ingredients unless an authorized person approves and records the change.`,
	{field: "ssw.body", source: `Saat berganti ke produk dengan profil alergen berbeda, tahan awal produksi sampai prosedur pergantian selesai. Periksa alat yang ditetapkan, bahan dan label baru, lalu minta pemeriksaan silang bila diwajibkan. Bila label tertukar atau residu terlihat, hentikan lini, pisahkan produk sejak titik terakhir yang diyakini benar, dan laporkan; jangan menjanjikan produk bebas alergen tanpa data resmi.

Kasus latihan: Setelah produk yang memakai susu, gulungan label produk tanpa susu dipasang sebelum catatan pembersihan pergantian selesai.
Respons yang dianjurkan: Hentikan lini dan tahan produk sejak pergantian label; laporkan urutan kejadian dan tunggu pemeriksaan pembersihan serta label.`}: `When changing to a product with a different allergen profile, hold startup until the changeover procedure is complete. Check designated equipment, new ingredients, and labels, and request a second check if required. If labels are mixed up or residue is visible, stop the line, separate product made since the last known-correct point, and report it. Do not promise an allergen-free product without official data.

Practice scenario: After a product containing milk, a label roll for a milk-free product is installed before the changeover cleaning record is complete.
Suggested response: Stop the line and hold product from the label change onward; report the sequence of events and wait for cleaning and label checks.`,
	{field: "ssw.body", source: `Saat menemukan benda asing atau alarm alat pemeriksa, hentikan aliran sesuai instruksi dan gunakan wadah tahan yang beridentitas. Catat waktu, mesin, produk, lot, serta jumlah sejak cek terakhir yang diterima. Jangan buang barang bukti atau memasukkan produk yang sama kembali ke aliran sebelum penilaian dilakukan.

Kasus latihan: Seorang pekerja melihat serpihan plastik di dekat produk di konveyor dan tidak tahu apakah serpihan berasal dari kemasan.
Respons yang dianjurkan: Hentikan atau tahan aliran sesuai instruksi, pisahkan produk terkait, dan laporkan lokasi serta waktu temuan untuk pemeriksaan.`}: `If you find a foreign object or an inspection-device alarm occurs, stop the line as instructed and use an identified hold container. Record the time, machine, product, lot, and quantity since the last accepted check. Do not discard evidence or put the same product back into the line before assessment.

Practice scenario: A worker sees a plastic fragment near a product on the conveyor and does not know whether it came from the packaging.
Suggested response: Stop or hold the line as instructed, separate the affected products, and report where and when the item was found for inspection.`,
	{field: "ssw.body", source: `Sebelum masuk, baca papan instruksi, simpan barang pribadi, ganti pakaian sesuai zona, tutup rambut, dan periksa perlengkapan. Bila merasa tidak sehat atau penutup luka lepas, hentikan tugas yang bersentuhan dengan pangan dan beri tahu pemimpin. Jangan menutupi masalah kesehatan demi target produksi.

Kasus latihan: Sebelum giliran dimulai, seorang pekerja mengalami diare dan akan memakai sarung tangan untuk mengemas produk siap makan.
Respons yang dianjurkan: Laporkan gejala kepada pemimpin dan jangan menangani pangan; tunggu penetapan tugas aman sesuai prosedur fasilitas.`}: `Before entering, read the posted instructions, store personal items, change clothing according to the zone, cover your hair, and check your equipment. If you feel unwell or a wound dressing comes loose, stop tasks involving food and tell the leader. Do not hide a health problem to meet production targets.

Practice scenario: Before the shift begins, a worker has diarrhea and plans to wear gloves while packaging ready-to-eat product.
Suggested response: Report the symptoms to the leader and do not handle food; wait to be assigned safe duties under facility procedures.`,
	{field: "ssw.body", source: `Sebelum mulai, periksa label status dan kondisi alat. Setelah tugas kotor, kirim alat ke tempat pencucian yang benar; jangan taruh kembali di rak alat bersih. Jika wadah bersih jatuh atau terkena percikan dari area kotor, pisahkan dan minta penilaian ulang. Simpan bahan kimia pembersih terpisah dari pangan.

Kasus latihan: Wadah berlabel bersih jatuh ke lantai dekat area bahan mentah; pekerja ingin segera menggunakannya kembali.
Respons yang dianjurkan: Pisahkan wadah itu sebagai belum siap pakai dan minta pencucian serta sanitasi ulang sesuai aturan fasilitas.`}: `Before starting, check equipment status labels and condition. After a dirty task, send tools to the correct wash area; do not put them back on the clean-equipment shelf. If a clean container falls or is splashed from a dirty area, set it aside and request a reassessment. Keep cleaning chemicals separate from food.

Practice scenario: A container marked clean falls on the floor near the raw-ingredient area, and a worker wants to reuse it immediately.
Suggested response: Set the container aside as not ready for use and request washing and sanitizing again according to facility rules.`,
	{field: "ssw.body", source: `Sebelum proses, pastikan kode produk, setelan, dan alat ukur cocok dengan instruksi. Catat waktu mulai, pembacaan, dan identitas lot tanpa membulatkan angka untuk menutupi selisih. Saat alarm berbunyi atau batas terlampaui, hentikan alur sesuai instruksi, tandai produk terdampak, dan laporkan segera.

Kasus latihan: Catatan pemanasan menunjukkan pembacaan belum mencapai batas pada instruksi produk, tetapi baki berikutnya sudah menunggu.
Respons yang dianjurkan: Hentikan pelepasan lot, tahan produk yang terdampak, catat pembacaan dan waktu, lalu minta keputusan penanggung jawab.`}: `Before processing, verify that the product code, settings, and measuring instruments match the instructions. Record the start time, readings, and lot identity without rounding to hide a discrepancy. When an alarm sounds or a limit is exceeded, stop the line as instructed, mark affected product, and report it immediately.

Practice scenario: Heating records show a reading below the limit in the product instructions, but the next trays are waiting.
Suggested response: Stop lot release, hold affected products, record the reading and time, and ask the person in charge to decide what to do.`,
	{field: "ssw.body", source: `Setiap bahan memiliki syarat penyimpanan yang ditentukan oleh produk, spesifikasi, dan sistem kendali fasilitas. Suhu ruang pendingin atau pembeku dipantau dengan alat dan batas tertulis; tidak ada satu angka umum yang cocok untuk semua produk. Bahan mentah dan produk siap makan juga dipisahkan untuk mencegah tetesan atau sentuhan silang.`}: `Each ingredient has storage requirements set by the product, specification, and facility control system. Monitor refrigerator and freezer temperatures with the specified instruments and written limits; there is no single value that applies to every product. Separate raw ingredients from ready-to-eat products to prevent drips or contact.`,
	{field: "ssw.body", source: `Tangan dapat memindahkan kotoran atau mikroorganisme dari toilet, sampah, uang, bahan mentah, telepon, dan permukaan ke pangan. Sarung tangan sekali pakai juga dapat tercemar dan bukan pengganti cuci tangan. Rencana fasilitas menentukan urutan cuci, pengeringan, sanitasi, serta kapan sarung tangan harus diganti.`}:                            `Hands can transfer dirt or microorganisms from toilets, waste, money, raw ingredients, phones, and surfaces to food. Disposable gloves can also become contaminated and do not replace handwashing. The facility's plan specifies the washing, drying, and sanitizing steps and when gloves must be changed.`,

	{field: "ssw.step", source: `Amati produk dan perhatikan alarm alat pemeriksa`}:                 `Observe the product and watch for inspection-device alarms`,
	{field: "ssw.step", source: `Ambil alat bersih dari tempat yang ditetapkan`}:                    `Take clean tools from the designated location`,
	{field: "ssw.step", source: `Baca aturan masuk untuk zona produksi`}:                            `Read the entry rules for the production zone`,
	{field: "ssw.step", source: `Baca instruksi dan batas tugas operator`}:                          `Read the operator instructions and task limits`,
	{field: "ssw.step", source: `Baca spesifikasi dan matriks alergen produk berikutnya`}:           `Read the next product's specifications and allergen matrix`,
	{field: "ssw.step", source: `Baca syarat simpan pada spesifikasi atau label`}:                   `Read storage requirements on the specification or label`,
	{field: "ssw.step", source: `Bersihkan atau ganti alat menurut prosedur yang berlaku`}:          `Clean or replace tools according to current procedure`,
	{field: "ssw.step", source: `Catat hasil pemeriksaan dan selisih aktual`}:                       `Record inspection results and actual discrepancies`,
	{field: "ssw.step", source: `Catat temuan dan serahkan untuk pemeriksaan`}:                      `Record the finding and submit it for inspection`,
	{field: "ssw.step", source: `Catat waktu dan pembacaan aktual`}:                                 `Record the time and actual reading`,
	{field: "ssw.step", source: `Catat waktu, lot, proses, dan hasil aktual`}:                       `Record the time, lot, process, and actual result`,
	{field: "ssw.step", source: `Cocokkan dokumen dan identitas lot`}:                               `Match the documents and lot identity`,
	{field: "ssw.step", source: `Cocokkan formula, bahan, dan lot`}:                                 `Match the recipe, ingredients, and lot`,
	{field: "ssw.step", source: `Cocokkan kode produk dan lot dengan instruksi`}:                    `Match the product code and lot to the instructions`,
	{field: "ssw.step", source: `Cocokkan produk dengan instruksi kendali proses`}:                  `Match the product to the process-control instructions`,
	{field: "ssw.step", source: `Cuci tangan sesuai urutan yang dipasang`}:                          `Wash your hands according to the posted sequence`,
	{field: "ssw.step", source: `Hentikan atau alihkan aliran sesuai instruksi`}:                    `Stop or divert the line as instructed`,
	{field: "ssw.step", source: `Hentikan kontak dengan produk dan buang sarung tangan yang kotor`}: `Stop handling product and discard contaminated gloves`,
	{field: "ssw.step", source: `Jalankan hanya setelah dinyatakan siap`}:                           `Start only after the equipment is declared ready`,
	{field: "ssw.step", source: `Jelaskan tindakan serta pekerjaan yang belum selesai`}:             `Explain the actions taken and unfinished work`,
	{field: "ssw.step", source: `Kenakan pakaian dan penutup rambut sesuai aturan`}:                 `Wear clothing and hair covering as required`,
	{field: "ssw.step", source: `Kenali batas zona dan label status alat`}:                          `Identify zone boundaries and equipment-status labels`,
	{field: "ssw.step", source: `Keringkan dan lakukan sanitasi sesuai prosedur`}:                   `Dry and sanitize according to procedure`,
	{field: "ssw.step", source: `Laporkan alat atau permukaan yang terkena cemaran`}:                `Report contaminated equipment or surfaces`,
	{field: "ssw.step", source: `Laporkan kerusakan atau keadaan tidak normal`}:                     `Report damage or abnormal conditions`,
	{field: "ssw.step", source: `Minta pemeriksaan ulang sebelum mulai menangani pangan`}:           `Request another check before handling food`,
	{field: "ssw.step", source: `Pakai sarung tangan baru bila tugas memerlukannya`}:                `Put on new gloves if the task requires them`,
	{field: "ssw.step", source: `Periksa alat ukur atau tampilan yang ditentukan`}:                  `Check the designated measuring instrument or display`,
	{field: "ssw.step", source: `Periksa jumlah, kemasan, tanggal, dan syarat kondisi`}:             `Check the quantity, packaging, date, and condition requirements`,
	{field: "ssw.step", source: `Periksa kondisi tubuh dan laporkan gejala atau luka`}:              `Check your health and report symptoms or wounds`,
	{field: "ssw.step", source: `Periksa label, informasi alergen, dan tanggal`}:                    `Check the label, allergen information, and date`,
	{field: "ssw.step", source: `Periksa pelindung, tombol, kabel, dan area sekitar`}:               `Check guards, controls, cables, and the surrounding area`,
	{field: "ssw.step", source: `Periksa satuan serta kondisi timbangan`}:                           `Check the units and scale condition`,
	{field: "ssw.step", source: `Periksa segel serta keadaan kemasan`}:                              `Check the seal and package condition`,
	{field: "ssw.step", source: `Pisahkan alat bekas ke tempat pencucian`}:                          `Put used tools in the wash area`,
	{field: "ssw.step", source: `Pisahkan produk lama dan hentikan alur sesuai instruksi`}:          `Separate the previous product and stop the line as instructed`,
	{field: "ssw.step", source: `Pisahkan produk terkait dengan identitas lot`}:                     `Separate affected products and keep their lot identity`,
	{field: "ssw.step", source: `Serahkan informasi dan pastikan penerima memahaminya`}:             `Hand over the information and confirm the recipient understands`,
	{field: "ssw.step", source: `Tahan bahan bermasalah serta minta keputusan`}:                     `Hold problem ingredients and ask for a decision`,
	{field: "ssw.step", source: `Tahan dan laporkan setiap ketidakcocokan`}:                         `Hold and report any mismatch`,
	{field: "ssw.step", source: `Tahan produk terdampak dan laporkan penyimpangan`}:                 `Hold affected product and report the deviation`,
	{field: "ssw.step", source: `Tahan serta laporkan penyimpangan sebelum bahan dipakai`}:          `Hold and report deviations before using the ingredients`,
	{field: "ssw.step", source: `Tambahkan menurut urutan dan tandai hasil kerja`}:                  `Add ingredients in sequence and mark completed work`,
	{field: "ssw.step", source: `Tandai produk atau proses yang masih ditahan`}:                     `Mark products or processes that remain on hold`,
	{field: "ssw.step", source: `Tempatkan lot di lokasi beridentitas yang benar`}:                  `Place the lot in the correctly identified location`,
	{field: "ssw.step", source: `Timbang dan catat tiap bahan secara terpisah`}:                     `Weigh and record each ingredient separately`,
	{field: "ssw.step", source: `Ukur dan catat suhu sesuai jadwal fasilitas`}:                      `Measure and record temperature on the facility schedule`,
	{field: "ssw.step", source: `Verifikasi bahan dan label sebelum melepas produk baru`}:           `Verify ingredients and labels before releasing a new product`,

	{field: "ssw.meaning", source: `Alat bersih diletakkan di rak ini.`}:                           `Clean tools are kept on this shelf.`,
	{field: "ssw.meaning", source: `batas atau standar`}:                                           `limit or standard`,
	{field: "ssw.meaning", source: `berhenti darurat`}:                                             `emergency stop`,
	{field: "ssw.meaning", source: `bersih dan higienis`}:                                          `clean and hygienic`,
	{field: "ssw.meaning", source: `catatan suhu`}:                                                 `temperature record`,
	{field: "ssw.meaning", source: `Hentikan pengemasan dan tahan produknya.`}:                     `Stop packaging and hold the product.`,
	{field: "ssw.meaning", source: `Jangan sentuh pangan dengan sarung tangan itu.`}:               `Do not touch food with those gloves.`,
	{field: "ssw.meaning", source: `keadaan atau status`}:                                          `condition or status`,
	{field: "ssw.meaning", source: `komposisi atau formula`}:                                       `composition or recipe`,
	{field: "ssw.meaning", source: `kontak silang`}:                                                `cross-contact`,
	{field: "ssw.meaning", source: `label atau keterangan`}:                                        `label or information`,
	{field: "ssw.meaning", source: `Label mungkin berbeda. Saya menahan produk ini.`}:              `The label may be different. I am holding this product.`,
	{field: "ssw.meaning", source: `label produk`}:                                                 `product label`,
	{field: "ssw.meaning", source: `Labelnya berbeda, jadi saya hentikan penimbangan.`}:            `The label is different, so I am stopping the weighing.`,
	{field: "ssw.meaning", source: `Mohon catat waktu dan nomor lot.`}:                             `Please record the time and lot number.`,
	{field: "ssw.meaning", source: `Mohon periksa catatan pembersihan pergantian.`}:                `Please check the changeover cleaning record.`,
	{field: "ssw.meaning", source: `Nomor label berbeda dari instruksi produksi.`}:                 `The label number differs from the production instructions.`,
	{field: "ssw.meaning", source: `nomor lot`}:                                                    `lot number`,
	{field: "ssw.meaning", source: `Nomor lot berbeda dari dokumen. Mohon periksa.`}:               `The lot number differs from the documents. Please check it.`,
	{field: "ssw.meaning", source: `pakaian kerja`}:                                                `work clothing`,
	{field: "ssw.meaning", source: `pelindung keselamatan`}:                                        `safety guard`,
	{field: "ssw.meaning", source: `Pelindung keselamatan longgar. Saya menghentikan mesin.`}:      `The safety guard is loose. I am stopping the machine.`,
	{field: "ssw.meaning", source: `pembekuan`}:                                                    `freezing`,
	{field: "ssw.meaning", source: `pemeriksaan kesehatan`}:                                        `health check`,
	{field: "ssw.meaning", source: `Pemeriksaan label lot ini belum selesai.`}:                     `The label check for this lot is not complete.`,
	{field: "ssw.meaning", source: `pemicu alergi`}:                                                `allergen`,
	{field: "ssw.meaning", source: `pemisahan atau isolasi`}:                                       `separation or isolation`,
	{field: "ssw.meaning", source: `penahanan sementara`}:                                          `temporary hold`,
	{field: "ssw.meaning", source: `pergantian produk`}:                                            `product changeover`,
	{field: "ssw.meaning", source: `produk tidak sesuai`}:                                          `nonconforming product`,
	{field: "ssw.meaning", source: `sanitasi atau disinfeksi`}:                                     `sanitation or disinfection`,
	{field: "ssw.meaning", source: `sarung tangan sekali pakai`}:                                   `disposable gloves`,
	{field: "ssw.meaning", source: `Saya agak demam. Saya akan melapor kepada penanggung jawab.`}:  `I have a slight fever. I will report it to the person in charge.`,
	{field: "ssw.meaning", source: `Saya akan memeriksa kondisi penyimpanan pada label.`}:          `I will check the storage conditions on the label.`,
	{field: "ssw.meaning", source: `Saya akan mencocokkan nama bahan dan nomor lot.`}:              `I will match the ingredient name and lot number.`,
	{field: "ssw.meaning", source: `Saya akan mengganti sarung tangan. Saya cuci tangan dahulu.`}:  `I will change my gloves. I will wash my hands first.`,
	{field: "ssw.meaning", source: `Saya menahan produk ini dan melapor kepada penanggung jawab.`}: `I am holding this product and reporting it to the person in charge.`,
	{field: "ssw.meaning", source: `Saya menemukan benda asing. Saya menghentikan lini.`}:          `I found a foreign object. I am stopping the line.`,
	{field: "ssw.meaning", source: `Saya minta petugas perbaikan memeriksanya.`}:                   `I will ask maintenance staff to inspect it.`,
	{field: "ssw.meaning", source: `Saya pisahkan kotak ini dan melapor kepada penanggung jawab.`}: `I am setting this box aside and reporting it to the person in charge.`,
	{field: "ssw.meaning", source: `Saya serahkan lokasi penahanan dan pemeriksaan berikutnya.`}:   `I am handing over the hold location and the next inspection.`,
	{field: "ssw.meaning", source: `Sebelum mulai bekerja, periksa kondisi tubuh Anda.`}:           `Before starting work, check your health.`,
	{field: "ssw.meaning", source: `Suhu belum mencapai batas pada instruksi.`}:                    `The temperature has not reached the limit in the instructions.`,
	{field: "ssw.meaning", source: `Suhu melewati batas. Saya menahan lot ini.`}:                   `The temperature exceeded the limit. I am holding this lot.`,
	{field: "ssw.meaning", source: `tanggal baik digunakan sebelum`}:                               `best-before date`,
	{field: "ssw.meaning", source: `timbangan`}:                                                    `scale`,
	{field: "ssw.meaning", source: `Wadah ini jatuh ke lantai. Mohon dicuci.`}:                     `This container fell on the floor. Please wash it.`,
	{field: "ssw.meaning", source: `zona atau area`}:                                               `zone or area`,
}
