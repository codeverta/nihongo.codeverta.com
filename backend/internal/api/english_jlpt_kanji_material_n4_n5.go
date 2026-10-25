package api

// jlptKanjiMaterialEnglishKey keeps kanji-material meanings scoped to the
// lesson that supplies their sense. The same Indonesian word can mean
// different things in unrelated kanji lessons.
type jlptKanjiMaterialEnglishKey struct {
	contextID string
	field     string
	source    string
}

// jlptKanjiMaterialEnglishCatalog contains reviewed, context-specific
// translations for the current N4 and N5 Kanji material gaps.
var jlptKanjiMaterialEnglishCatalog = map[jlptKanjiMaterialEnglishKey]string{
	{contextID: `kanji:N4:kanji-n4-03`, field: `meaning`, source: `tujuan`}:      `purpose`,
	{contextID: `kanji:N4:kanji-n4-04`, field: `meaning`, source: `berjalan`}:    `to walk`,
	{contextID: `kanji:N4:kanji-n4-04`, field: `meaning`, source: `cukup`}:       `enough; sufficient`,
	{contextID: `kanji:N4:kanji-n4-04`, field: `meaning`, source: `jalan-jalan`}: `to take a walk; stroll`,
	{contextID: `kanji:N4:kanji-n4-10`, field: `meaning`, source: `pekerjaan`}:   `work; job`,
	{contextID: `kanji:N4:kanji-n4-21`, field: `meaning`, source: `kekurangan`}:  `shortage; lack`,
	{contextID: `kanji:N4:kanji-n4-23`, field: `meaning`, source: `berat`}:       `heavy`,
	{contextID: `kanji:N4:kanji-n4-25`, field: `meaning`, source: `kematian`}:    `death`,
	{contextID: `kanji:N4:kanji-n4-28`, field: `meaning`, source: `perasaan`}:    `feeling`,
	{contextID: `kanji:N4:kanji-n4-29`, field: `meaning`, source: `pakaian`}:     `clothing; clothes`,
	{contextID: `kanji:N4:kanji-n4-30`, field: `meaning`, source: `luas`}:        `wide; spacious`,
	{contextID: `kanji:N4:kanji-n4-31`, field: `meaning`, source: `selesai`}:     `to finish; be over`,
	{contextID: `kanji:N5:kanji-n5-04`, field: `meaning`, source: `tanah`}:       `soil; earth`,
	{contextID: `kanji:N5:kanji-n5-09`, field: `meaning`, source: `cukup`}:       `enough; sufficient`,
	{contextID: `kanji:N5:kanji-n5-17`, field: `meaning`, source: `makan`}:       `to eat`,
}
