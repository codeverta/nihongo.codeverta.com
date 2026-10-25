package api

// EnglishSSWSourceTitleTranslation returns a source title reviewed for the
// sector whose English curriculum displays it.
func EnglishSSWSourceTitleTranslation(sector, source string) (string, bool) {
	key := sswSourceTitleEnglishKey{sector: sector, source: source}
	if translated, ok := englishSSWSourceTitleCatalogA[key]; ok {
		return translated, true
	}
	if translated, ok := englishSSWSourceTitleCatalogB[key]; ok {
		return translated, true
	}
	return "", false
}
