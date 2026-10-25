package api

// englishSSWSourceTitleCatalogB contains hand-written English for every
// source-list title in the assigned SSW sectors. Organization names and
// official Japanese titles are retained where an authoritative English name
// is not established.
var englishSSWSourceTitleCatalogB = map[sswSourceTitleEnglishKey]string{
	{sector: "industrial-manufacturing", source: `METI: kategori tempat usaha dan pekerjaan yang memenuhi syarat`}: `METI: Categories of Eligible Workplaces and Jobs`,
	{sector: "industrial-manufacturing", source: `JAIM: gambaran ujian SSW No. 1 dan No. 2`}:                       `JAIM: Overview of the SSW No. 1 and No. 2 Exams`,
	{sector: "industrial-manufacturing", source: `JAIM: standar rinci ujian kategori industri`}:                    `JAIM: Detailed Standards for the Industrial Category Exam`,
	{sector: "industrial-manufacturing", source: `METI: kebijakan penerimaan tenaga kerja manufaktur`}:             `METI: Policy on Accepting Workers in Manufacturing`,

	{sector: "kaigo", source: `MHLW — materi Kaigo dan revisi ke-2 (2026)`}:            `MHLW — Kaigo materials and second revision (2026)`,
	{sector: "kaigo", source: `MHLW — teks resmi Kaigo bahasa Indonesia, revisi ke-2`}: `MHLW — Official Indonesian-language Kaigo text, second revision`,
	{sector: "kaigo", source: `MHLW — cakupan evaluasi keterampilan dan bahasa Kaigo`}: `MHLW — Kaigo skills and language assessment scope`,
	{sector: "kaigo", source: `WHO — pemahaman demensia`}:                              `WHO — Understanding dementia`,
	{sector: "kaigo", source: `CDC — kebersihan tangan petugas`}:                       `CDC — Hand hygiene for care staff`,
	{sector: "kaigo", source: `OSHA — penanganan pengguna dan alat transfer`}:          `OSHA — Assisting people and using transfer equipment`,
	{sector: "kaigo", source: `NHS — kesulitan menelan`}:                               `NHS — Swallowing difficulties`,

	{sector: "linen-supply", source: `ISA Jepang — ruang lingkup pekerjaan linen supply untuk SSW`}:  `Japan Immigration Services Agency — SSW linen-supply job scope`,
	{sector: "linen-supply", source: `MHLW — informasi resmi SSW sektor linen supply`}:               `MHLW — Official information on the SSW linen-supply sector`,
	{sector: "linen-supply", source: `MHLW — alur kerja dan pengembangan keterampilan linen supply`}: `MHLW — Linen-supply workflows and skills development`,
	{sector: "linen-supply", source: `MHLW — standar resmi kerja terkait linen supply`}:              `MHLW — Official standards for linen-supply work`,

	{sector: "logistics-warehouse", source: `国土交通省 — 物流倉庫分野における特定技能外国人・育成就労外国人の受入れ`}:     `国土交通省 — Accepting Specified Skilled Worker and Employment-for-Skill-Development Workers in the Logistics Warehouse Sector`,
	{sector: "logistics-warehouse", source: `出入国在留管理庁 — 特定技能制度（物流倉庫分野を含む制度全般）`}:          `出入国在留管理庁 — Specified Skilled Worker System (general information, including the logistics warehouse sector)`,
	{sector: "logistics-warehouse", source: `国土交通省 — 物流倉庫分野別運用方針（2026年1月23日閣議決定、掲載ページ）`}: `国土交通省 — Operational Policy for the Logistics Warehouse Sector (Cabinet decision of January 23, 2026; publication page)`,

	{sector: "railway", source: `国土交通省 — 鉄道分野における外国人材の受入れ（特定技能・育成就労）`}:        `国土交通省 — Accepting Foreign Workers in the Railway Sector (Specified Skilled Worker and Employment for Skill Development)`,
	{sector: "railway", source: `国土交通省 — 鉄道分野の業務区分（業務内容の詳細）`}:                 `国土交通省 — Railway Sector Job Categories (Detailed Duties)`,
	{sector: "railway", source: `国土交通省 — 鉄道分野における特定技能・育成就労の運用方針（2026年1月23日）`}: `国土交通省 — Operational Policy for Specified Skilled Worker and Employment for Skill Development in the Railway Sector (January 23, 2026)`,
	{sector: "railway", source: `一般社団法人日本鉄道運転協会 — 運輸係員区分の試験案内`}:               `一般社団法人日本鉄道運転協会 — Exam Information for the Train Operations Staff Category`,
	{sector: "railway", source: `一般社団法人日本鉄道施設協会 — 軌道整備区分の特定技能`}:               `一般社団法人日本鉄道施設協会 — Specified Skilled Worker Category for Track Maintenance`,

	{sector: "resource-recycling", source: `ISA Jepang — pekerjaan SSW(i) sektor resource recycling`}:                                  `Japan Immigration Services Agency — SSW (i) work in the resource-recycling sector`,
	{sector: "resource-recycling", source: `Ministry of the Environment — sektor resource recycling, scope and official test status`}:  `Ministry of the Environment — Resource-recycling sector scope and official test status`,
	{sector: "resource-recycling", source: `Ministry of the Environment — official overview of resource-recycling intake preparation`}: `Ministry of the Environment — Official overview of preparations for accepting workers in resource recycling`,
	{sector: "resource-recycling", source: `ISA — sector-specific policy pages, including Resource Recycling`}:                         `ISA — Sector-specific policy pages, including resource recycling`,

	{sector: "road-transport", source: `出入国在留管理庁 — 特定技能１号の各分野の仕事内容（Job Description）`}: `出入国在留管理庁 — Job descriptions for each Specified Skilled Worker (i) sector`,
	{sector: "road-transport", source: `出入国在留管理庁 — 自動車運送業分野`}:                         `出入国在留管理庁 — Motor Vehicle Transportation Sector`,
	{sector: "road-transport", source: `国土交通省 — 自動車運送業分野における特定技能外国人の受入れについて`}:         `国土交通省 — Accepting Specified Skilled Workers in the Motor Vehicle Transportation Sector`,
	{sector: "road-transport", source: `国土交通省 — 自動車運送業分野特定技能１号評価試験実施要領`}:              `国土交通省 — Implementation Guidelines for the Specified Skilled Worker (i) Evaluation Test in the Motor Vehicle Transportation Sector`,

	{sector: "shipbuilding", source: `MLIT: lingkup bidang dan ujian shipbuilding/marine equipment`}: `MLIT: Shipbuilding and Marine Equipment Sector Scope and Exams`,
	{sector: "shipbuilding", source: `ISA: uraian kerja SSW No. 2 untuk tiga divisi`}:                `ISA: SSW No. 2 Job Descriptions for the Three Divisions`,
	{sector: "shipbuilding", source: `MLIT: standar ujian keterampilan No. 1`}:                       `MLIT: No. 1 Skills Test Standards`,
	{sector: "shipbuilding", source: `ClassNK: operator ujian dan informasi trade`}:                  `ClassNK: Test Operator and Trade Information`,

	{sector: "wood-industry", source: `Forestry Agency — penerimaan tenaga asing di sektor industri kayu`}:             `Forestry Agency — Accepting Foreign Workers in the Wood Industry Sector`,
	{sector: "wood-industry", source: `Forestry Agency — teks belajar Wood Industry SSW(i), termasuk versi Indonesia`}: `Forestry Agency — Wood Industry SSW (i) study text, including the Indonesian version`,
	{sector: "wood-industry", source: `全国木材組合連合会 — situs resmi ujian sektor industri kayu`}:                            `全国木材組合連合会 — Official website for the wood industry sector exam`,
	{sector: "wood-industry", source: `Prometric — informasi ujian Wood Industry SSW(i)`}:                              `Prometric — Wood Industry SSW (i) exam information`,
	{sector: "wood-industry", source: `ISA Jepang — ringkasan pekerjaan SSW(i)`}:                                       `Japan Immigration Services Agency — Summary of SSW (i) work`,
}
