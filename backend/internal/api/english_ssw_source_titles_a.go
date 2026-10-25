package api

type sswSourceTitleEnglishKey struct {
	sector string
	source string
}

var englishSSWSourceTitleCatalogA = map[sswSourceTitleEnglishKey]string{
	{sector: "accommodation", source: `ISA — sektor akomodasi: lingkup pekerjaan, tes keterampilan, dan bahasa`}: `ISA — Accommodation Sector: Job Scope, Skills Tests, and Japanese Language`,
	{sector: "accommodation", source: `ISA — deskripsi pekerjaan SSW No. 1 menurut sektor`}:                      `ISA — SSW No. 1 Job Descriptions by Sector`,
	{sector: "accommodation", source: `Japan Tourism Agency — kebijakan dan pedoman ujian sektor akomodasi`}:     `Japan Tourism Agency — Accommodation Sector Policies and Examination Guidelines`,
	{sector: "accommodation", source: `CAIPT — ruang lingkup, syarat, dan informasi ujian SSW sektor akomodasi`}: `CAIPT — Scope, Requirements, and SSW Accommodation Examination Information`,

	{sector: "agriculture", source: `NCA · Cakupan pertanian tanaman dan peternakan`}: `NCA · Scope of Crop and Livestock Farming`,
	{sector: "agriculture", source: `NCA · Pedoman ASAT 1, 70 soal sekitar 60 menit`}: `NCA · ASAT 1 Guidelines: 70 Questions, About 60 Minutes`,
	{sector: "agriculture", source: `MAFF · Keselamatan kerja pertanian`}:             `MAFF · Agricultural Workplace Safety`,

	{sector: "automobile-maintenance", source: `ISA — sektor perawatan kendaraan bermotor: pekerjaan, tes keterampilan dan bahasa`}: `ISA — Motor Vehicle Maintenance Sector: Jobs, Skills Tests, and Japanese Language`,
	{sector: "automobile-maintenance", source: `ISA — deskripsi pekerjaan SSW No. 1 menurut sektor`}:                                `ISA — SSW No. 1 Job Descriptions by Sector`,
	{sector: "automobile-maintenance", source: `MLIT — penerimaan pekerja SSW di sektor perawatan kendaraan bermotor`}:              `MLIT — Accepting SSW Workers in the Motor Vehicle Maintenance Sector`,
	{sector: "automobile-maintenance", source: `JASPA — ruang lingkup dan informasi ujian evaluasi perawatan kendaraan bermotor`}:   `JASPA — Scope and Information for the Motor Vehicle Maintenance Skills Assessment Test`,

	{sector: "aviation", source: `ISA — sektor penerbangan: ruang lingkup pekerjaan, keterampilan, dan bahasa`}: `ISA — Aviation Sector: Job Scope, Skills, and Japanese Language`,
	{sector: "aviation", source: `ISA — deskripsi pekerjaan SSW No. 2, ground handling dan perawatan pesawat`}:  `ISA — SSW No. 2 Job Descriptions: Ground Handling and Aircraft Maintenance`,
	{sector: "aviation", source: `MLIT — kebijakan, persyaratan, dan informasi ujian sektor penerbangan`}:       `MLIT — Aviation Sector Policies, Requirements, and Examination Information`,
	{sector: "aviation", source: `MLIT — pedoman ujian ground handling berlaku 1 Agustus 2026`}:                 `MLIT — Ground Handling Examination Guidelines, Effective August 1, 2026`,
	{sector: "aviation", source: `JAEA — revisi 2026 dan bahan belajar ground handling`}:                        `JAEA — 2026 Revision and Ground Handling Study Materials`,
	{sector: "aviation", source: `JAEA — bahan belajar dan materi ujian perawatan pesawat`}:                     `JAEA — Aircraft Maintenance Study and Examination Materials`,
	{sector: "aviation", source: `MLIT — pedoman ujian SSW No. 2 untuk perawatan pesawat`}:                      `MLIT — SSW No. 2 Examination Guidelines for Aircraft Maintenance`,

	{sector: "building-cleaning", source: `JBMA · SSW 1号: cakupan, CBT dan bahan belajar (ditinjau 17 Oktober 2026)`}: `JBMA · SSW No. 1: Scope, CBT, and Study Materials (Reviewed October 17, 2026)`,
	{sector: "building-cleaning", source: `JBMA · Bidang pembersihan gedung`}:                                         `JBMA · Building Cleaning Sector`,
	{sector: "building-cleaning", source: `MHLW · Keselamatan kerja dan informasi bahan kimia`}:                       `MHLW · Workplace Safety and Chemical Information`,
	{sector: "building-cleaning", source: `CDC · Cleaning and disinfecting`}:                                          `CDC · Cleaning and disinfecting`,

	{sector: "construction", source: `ISA: bidang SSW konstruksi dan jenis pekerjaan`}:           `ISA: SSW Construction Sector and Job Types`,
	{sector: "construction", source: `JAC: ujian dan lingkup bahan belajar per bidang`}:          `JAC: Examinations and Study Material Scope by Field`,
	{sector: "construction", source: `JAC: bahan resmi dan lingkup ujian`}:                       `JAC: Official Materials and Examination Scope`,
	{sector: "construction", source: `MLIT: standar evaluasi keterampilan konstruksi SSW No. 1`}: `MLIT: SSW No. 1 Construction Skills Evaluation Standards`,

	{sector: "fishery", source: `Japan Fisheries Agency · Keselamatan kerja perikanan`}: `Japan Fisheries Agency · Fisheries Workplace Safety`,
	{sector: "fishery", source: `Dainippon Suisankai · Cakupan dan teks SSW 1`}:         `Dainippon Suisankai · SSW No. 1 Scope and Study Text`,
	{sector: "fishery", source: `Prometric · SSW 1 perikanan tangkap`}:                  `Prometric · SSW No. 1 Capture Fisheries`,
	{sector: "fishery", source: `Prometric · SSW 1 budi daya`}:                          `Prometric · SSW No. 1 Aquaculture`,

	{sector: "food-manufacturing", source: `ISA Jepang — informasi sektor produksi makanan dan minuman untuk SSW`}: `ISA Japan — Information on the SSW Food and Beverage Manufacturing Sector`,
	{sector: "food-manufacturing", source: `MAFF — penerimaan tenaga asing dan ujian sektor produksi makanan`}:     `MAFF — Foreign Worker Intake and Examinations for the Food Manufacturing Sector`,
	{sector: "food-manufacturing", source: `OTAFF — panduan ujian SSW(i) produksi makanan dan minuman`}:            `OTAFF — SSW(i) Food and Beverage Manufacturing Examination Guide`,
	{sector: "food-manufacturing", source: `OTAFF — teks belajar ujian produksi makanan dan minuman`}:              `OTAFF — Food and Beverage Manufacturing Examination Study Text`,
	{sector: "food-manufacturing", source: `MHLW — panduan HACCP dan kebersihan pangan`}:                           `MHLW — HACCP and Food Hygiene Guide`,
	{sector: "food-manufacturing", source: `Consumer Affairs Agency — informasi label alergen pangan`}:             `Consumer Affairs Agency — Food Allergen Labeling Information`,

	{sector: "food-service", source: `ISA Jepang — informasi dan status sektor Food Service untuk SSW`}:        `ISA Japan — SSW Food Service Sector Information and Status`,
	{sector: "food-service", source: `ISA Jepang — ringkasan pekerjaan SSW(i)`}:                                `ISA Japan — SSW(i) Job Overview`,
	{sector: "food-service", source: `MAFF — penerimaan tenaga asing di sektor Food Service`}:                  `MAFF — Foreign Worker Intake in the Food Service Sector`,
	{sector: "food-service", source: `OTAFF — ujian SSW(i) Food Service dan panduan ujian`}:                    `OTAFF — SSW(i) Food Service Examination and Guide`,
	{sector: "food-service", source: `OTAFF — pemberitahuan penghentian pemesanan ujian Food Service SSW(i)`}:  `OTAFF — Notice on the Suspension of SSW(i) Food Service Examination Reservations`,
	{sector: "food-service", source: `Japan Food Service Association — materi belajar ujian SSW Food Service`}: `Japan Food Service Association — SSW Food Service Examination Study Materials`,
	{sector: "food-service", source: `MHLW — panduan HACCP dan kebersihan pangan`}:                             `MHLW — HACCP and Food Hygiene Guide`,
	{sector: "food-service", source: `Consumer Affairs Agency — informasi label alergen pangan`}:               `Consumer Affairs Agency — Food Allergen Labeling Information`,

	{sector: "forestry", source: `ISA Jepang — ringkasan pekerjaan SSW(i)`}:                              `ISA Japan — SSW(i) Job Overview`,
	{sector: "forestry", source: `Forestry Agency — penerimaan tenaga asing sektor kehutanan dan ujian`}: `Forestry Agency — Foreign Worker Intake and Examinations in the Forestry Sector`,
	{sector: "forestry", source: `Forestry Agency — informasi 林業技能測定試験`}:                                 `Forestry Agency — Information on 林業技能測定試験`,
	{sector: "forestry", source: `林業技能向上センター — badan pelaksana dan informasi ujian`}:                     `林業技能向上センター — Examination Administrator and Information`,
	{sector: "forestry", source: `MHLW — keselamatan kerja penebangan dan kehutanan`}:                    `MHLW — Logging and Forestry Workplace Safety`,
	{sector: "forestry", source: `Forestry Agency — materi pelatihan kehutanan bagi pekerja asing`}:      `Forestry Agency — Forestry Training Materials for Foreign Workers`,
}
