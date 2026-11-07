package api

import "strings"

// translateAccountPayload localizes the app-authored text in the successful
// JSON responses of the account, billing, jobs, Anki and branding endpoints.
// The payload must be decoded with json.Decoder.UseNumber (maps are
// map[string]any, arrays []any). It is mutated in place and the result reports
// whether anything changed.
//
// Only display fields are translated, and only through exact lookups: ids,
// kinds, statuses, tiers and categories that the Flutter client uses as logic
// keys are never touched, and unknown strings (user content, Japanese source
// text, external job postings, manga and Anki card content) stay unchanged.
// The function never panics on unexpected payload shapes.
func translateAccountPayload(path string, payload any) bool {
	path = strings.TrimRight(path, "/")
	switch {
	case accountPathIs(path, "/v1/me"):
		return translateAccountMe(path, payload)
	case accountPathIs(path, "/v1/plans"), accountPathIs(path, "/v1/pricing"), accountPathIs(path, "/v1/billing"):
		return translatePlanFeatures(payload)
	case accountPathIs(path, "/v1/jobs"):
		return translateJobsPayload(payload)
	case accountPathIs(path, "/v1/anki"):
		return translateAnkiPayload(payload)
	case accountPathIs(path, "/v1/branding"):
		return translateBrandingPayload(payload)
	}
	// /v1/auth/*, /v1/manga* and /v1/app-update-policy carry no app-authored
	// Indonesian text in successful responses (ids, tokens, profile data,
	// source-site content and administrator-entered messages only).
	return false
}

func accountPathIs(path, base string) bool {
	return path == base || strings.HasPrefix(path, base+"/")
}

// walkAccountMaps calls visit for every JSON object in value, parents before
// children. Children stored under any of skipKeys are not visited.
func walkAccountMaps(value any, visit func(map[string]any) bool, skipKeys ...string) bool {
	return walkAccountMapsDepth(value, 0, visit, skipKeys)
}

func walkAccountMapsDepth(value any, depth int, visit func(map[string]any) bool, skipKeys []string) bool {
	if depth > 64 {
		return false
	}
	changed := false
	switch value := value.(type) {
	case map[string]any:
		if visit(value) {
			changed = true
		}
	children:
		for key, child := range value {
			for _, skip := range skipKeys {
				if key == skip {
					continue children
				}
			}
			if walkAccountMapsDepth(child, depth+1, visit, skipKeys) {
				changed = true
			}
		}
	case []any:
		for _, child := range value {
			if walkAccountMapsDepth(child, depth+1, visit, skipKeys) {
				changed = true
			}
		}
	}
	return changed
}

// translateAccountField replaces m[key] when it is a string with an exact
// entry in table.
func translateAccountField(m map[string]any, key string, table map[string]string) bool {
	current, ok := m[key].(string)
	if !ok {
		return false
	}
	translated, found := table[current]
	if !found || translated == current {
		return false
	}
	m[key] = translated
	return true
}

func setAccountField(m map[string]any, key, value string) bool {
	current, ok := m[key].(string)
	if !ok || current == value {
		return false
	}
	m[key] = value
	return true
}

func isAccountDigits(text string) bool {
	if text == "" {
		return false
	}
	for _, r := range text {
		if r < '0' || r > '9' {
			return false
		}
	}
	return true
}

// ---------------------------------------------------------------------------
// /v1/me/*
// ---------------------------------------------------------------------------

func translateAccountMe(path string, payload any) bool {
	changed := false
	if path == "/v1/me/gamification" || path == "/v1/me/achievements" {
		changed = translateAchievements(payload) || changed
	}
	if accountPathIs(path, "/v1/me/learning-history") {
		changed = translateLearningHistory(payload) || changed
	}
	// /v1/me/profile, /v1/me/timezone, /v1/me/progress and /v1/me/account
	// contain only ids, numbers, timestamps and user-entered values.
	return changed
}

type accountAchievementText struct {
	Name        string
	Description string
}

// accountAchievementEnglish is keyed by the stable achievement id from
// internal/store/gamification.go. The Flutter client reads `id`, `icon`,
// `progress`, `target` and `unlockedAt` for logic and only displays `name` and
// `description`, so those two are the only fields translated.
var accountAchievementEnglish = map[string]accountAchievementText{
	"first-lesson":  {"First Step", "Pass one lesson quiz with a score of at least 80%."},
	"lessons-10":    {"Dedicated Learner", "Pass 10 different lessons."},
	"lessons-50":    {"50 Lessons", "Pass 50 different lessons."},
	"vocabulary-10": {"Word Collector", "Pass 10 vocabulary lessons."},
	"grammar-10":    {"Grammar Builder", "Pass 10 grammar lessons."},
	"kanji-10":      {"Kanji Explorer", "Pass 10 kanji lessons."},
	"reading-5":     {"Bookworm", "Pass 5 reading lessons."},
	"all-subjects":  {"All-Rounder", "Pass lessons in all four subjects."},
	"streak-7":      {"7-Day Streak", "Complete a quiz or exam for 7 consecutive days."},
	"streak-30":     {"30-Day Streak", "Complete a quiz or exam for 30 consecutive days."},
	"xp-500":        {"Level Up", "Earn 500 XP."},
	"exam-ready":    {"Exam Ready", "Score at least 80% on a diagnostic or simulation."},
}

// translateAchievements localizes every achievement object (identified by its
// stable id) found anywhere in payload: the bare array of /v1/me/achievements,
// the `achievements` list of /v1/me/gamification, and the same list when it is
// embedded in other responses.
func translateAchievements(payload any) bool {
	return walkAccountMaps(payload, func(m map[string]any) bool {
		id, _ := m["id"].(string)
		text, ok := accountAchievementEnglish[id]
		if !ok {
			return false
		}
		changed := setAccountField(m, "name", text.Name)
		return setAccountField(m, "description", text.Description) || changed
	})
}

// translateLearningHistory localizes the lesson `title` of learning-history
// items with the same reviewed lesson-title catalogs used by the learning
// endpoints. `kind`, `lessonId` and `level` are logic keys and stay as is.
func translateLearningHistory(payload any) bool {
	return walkAccountMaps(payload, func(m map[string]any) bool {
		kind, _ := m["kind"].(string)
		title, ok := m["title"].(string)
		if !ok || title == "" {
			return false
		}
		switch kind {
		case "vocabulary", "grammar", "kanji", "reading":
		default:
			return false
		}
		lessonID, _ := m["lessonId"].(string)
		materialContextID := ""
		if kind == "kanji" {
			parts := strings.Split(lessonID, "-")
			if len(parts) >= 3 && strings.HasPrefix(parts[1], "n") {
				materialContextID = "kanji:" + strings.ToUpper(parts[1]) + ":" + lessonID
			}
		}
		translated := translateLearningTextInQuestionContext("title", title, nil, kind, materialContextID)
		return setAccountField(m, "title", translated)
	})
}

// ---------------------------------------------------------------------------
// /v1/plans, /v1/pricing, /v1/billing
// ---------------------------------------------------------------------------

// accountPlanFeatureEnglish covers the feature bullets of the seeded plans in
// internal/store/memory.go. Plan ids, `billingType` and `recommended` drive
// client logic and are untouched; plan names are already English.
var accountPlanFeatureEnglish = map[string]string{
	"JLPT N5–N4 dasar":                "Basic JLPT N5–N4",
	"Daily practice":                  "Daily practice",
	"Progress dasar":                  "Basic progress",
	"Materi N3–N1 yang tersedia":      "Available N3–N1 materials",
	"Seluruh tes yang tersedia":       "All available tests",
	"Restore pembelian":               "Restore purchases",
	"Sekali bayar tanpa perpanjangan": "One-time payment with no renewal",
}

// translatePlanFeatures localizes `features` string arrays of plan objects.
func translatePlanFeatures(payload any) bool {
	return walkAccountMaps(payload, func(m map[string]any) bool {
		features, ok := m["features"].([]any)
		if !ok {
			return false
		}
		changed := false
		for index, item := range features {
			text, ok := item.(string)
			if !ok {
				continue
			}
			if translated, found := accountPlanFeatureEnglish[text]; found && translated != text {
				features[index] = translated
				changed = true
			}
		}
		return changed
	})
}

// ---------------------------------------------------------------------------
// /v1/jobs*
// ---------------------------------------------------------------------------

// accountJobOptionEnglish is keyed by metadata group and option id (see
// internal/jobs/metadata.go). The client resolves labels through ids, so only
// `label` is translated. Prefectures are romanized names and stay unchanged.
var accountJobOptionEnglish = map[string]map[string]string{
	"categories": {
		"engineering":        "IT & Engineering",
		"office":             "Administration & Corporate",
		"hospitality":        "Hotels & Tourism",
		"nursing-care":       "Nursing Care (Kaigo)",
		"building-cleaning":  "Building Cleaning",
		"manufacturing":      "Manufacturing",
		"construction":       "Construction",
		"agriculture":        "Agriculture & Livestock",
		"fishery":            "Fishery & Aquaculture",
		"food-service":       "Restaurants",
		"food-manufacturing": "Food Production",
		"logistics":          "Logistics & Warehousing",
		"transportation":     "Transportation",
		"education":          "Education",
		"creative":           "Design & Creative",
		"sales":              "Sales & Marketing",
		"finance":            "Finance",
		"healthcare":         "Healthcare",
		"general":            "Other",
	},
	"employmentTypes": {
		"full-time":  "Full-time",
		"part-time":  "Part-time",
		"contract":   "Contract",
		"temporary":  "Temporary",
		"internship": "Internship",
		"unknown":    "Not specified",
	},
	"japaneseLevels": {
		"N1":           "JLPT N1",
		"N2":           "JLPT N2",
		"N3":           "JLPT N3",
		"N4":           "JLPT N4",
		"N5":           "JLPT N5",
		"not-required": "Japanese not required",
		"unknown":      "Not specified",
	},
	"visaOptions": {
		"sponsored":            "Sponsorship mentioned",
		"ssw":                  "SSW mentioned",
		"work-permit-required": "Work permit required",
		"unknown":              "Not specified",
	},
	"sectors": {
		"nursing-care":             "Nursing Care (Kaigo)",
		"building-cleaning":        "Building Cleaning",
		"linen-supply":             "Linen Processing",
		"industrial-manufacturing": "Industrial Product Manufacturing",
		"construction":             "Construction",
		"shipbuilding":             "Shipbuilding & Ship Machinery",
		"automobile-maintenance":   "Automobile Maintenance",
		"aviation":                 "Aviation",
		"accommodation":            "Accommodation",
		"road-transport":           "Road Transportation",
		"railway":                  "Railway",
		"logistics-warehouse":      "Logistics Warehousing",
		"agriculture":              "Agriculture",
		"fishery":                  "Fishery",
		"food-manufacturing":       "Food Production",
		"food-service":             "Food Service",
		"forestry":                 "Forestry",
		"wood-industry":            "Wood Industry",
		"resource-recycling":       "Recycling",
	},
	"remoteOptions": {
		"onsite":  "On-site",
		"hybrid":  "Hybrid",
		"remote":  "Remote",
		"unknown": "Not specified",
	},
}

// accountJobSourceSuffixEnglish translates the app-authored suffix after
// " · " in source names such as "TokyoDev · lowongan teknologi". The company or
// site name before the separator is a proper name and stays unchanged.
var accountJobSourceSuffixEnglish = map[string]string{
	"lowongan teknologi":  "tech jobs",
	"lowongan agrikultur": "agriculture jobs",
	"lowongan umum":       "general jobs",
	"Karier resmi":        "Official careers",
	"Karier Jepang resmi": "Official Japan careers",
}

var accountJobSourceScopeEnglish = map[string]string{
	"Seluruh Jepang":                      "All of Japan",
	"Jepang":                              "Japan",
	"Tokyo":                               "Tokyo",
	"Lowongan Jepang pada perusahaan ini": "Job listings in Japan from this company",
	"Sampel lowongan publik Jepang":       "Sample of public job listings in Japan",
}

var accountJobSourceReasonEnglish = map[string]string{
	"Tersedia sebagai direktori. Pengambilan lowongan otomatis dari situs ini belum tersedia.":                            "Available as a directory. Automatic job fetching from this site is not available yet.",
	"Direktori pencarian dan dukungan kerja. Buka situs sumber untuk informasi terkini.":                                  "Job search and employment support directory. Visit the source website for current information.",
	"Koneksi ke situs belum berhasil diverifikasi. Silakan buka situs sumber.":                                            "The connection to the site could not be verified yet. Please visit the source website.",
	"Situs melarang crawler umum melalui robots.txt. Lowongan dapat dibuka langsung di situs sumber.":                     "The site disallows general crawlers through robots.txt. Listings can be opened directly on the source website.",
	"Akses otomatis mendapat pembatasan HTTP 403. Gunakan situs sumber untuk melihat lowongan.":                           "Automated access is restricted (HTTP 403). Use the source website to view listings.",
	"Akses otomatis dibatasi. Lowongan tersedia melalui situs sumber.":                                                    "Automated access is restricted. Listings are available through the source website.",
	"Akses otomatis dibatasi oleh situs. Silakan buka lowongan melalui LinkedIn.":                                         "Automated access is restricted by the site. Please open listings through LinkedIn.",
	"Alamat pencarian lowongan lama tidak tersedia. Gunakan situs sumber untuk informasi terbaru.":                        "The old job search address is no longer available. Use the source website for the latest information.",
	"Akses lowongan memerlukan akun pada situs sumber.":                                                                   "Accessing listings requires an account on the source website.",
	"Layanan ketenagakerjaan pemerintah. Gunakan pencarian pada situs sumber.":                                            "Government employment service. Use the search on the source website.",
	"Detail lowongan yang diperiksa mengalihkan akses. Gunakan situs sumber dan ikuti proses registrasinya.":              "The listing details we checked redirect access. Use the source website and follow its registration process.",
	"Layanan pencocokan kandidat dan perusahaan. Informasi tersedia pada situs sumber.":                                   "Candidate and employer matching service. Information is available on the source website.",
	"Akses lowongan memerlukan registrasi dan verifikasi profil atau sertifikat.":                                         "Accessing listings requires registration and verification of your profile or certificate.",
	"Layanan pelatihan dan pendampingan kerja. Hubungi sumber untuk informasi lowongan.":                                  "Job training and support service. Contact the source for listing information.",
	"Portal informasi dan pendampingan kerja dengan rujukan ke layanan lowongan.":                                         "Employment information and support portal with referrals to job listing services.",
	"Informasi resmi SSW, ujian, prosedur, dan acara rekrutmen.":                                                          "Official information on SSW, exams, procedures, and recruitment events.",
	"Lowongan perawat berlisensi. Persyaratan berbeda dari pekerjaan kaigo SSW.":                                          "Licensed nurse positions. Requirements differ from SSW kaigo work.",
	"Lowongan pengemudi; SIM dan kelayakan kerja mengikuti persyaratan tiap posisi.":                                      "Driver positions; driver's license and work eligibility follow each position's requirements.",
	"Pengambilan terbatas setiap minggu; lihat situs sumber untuk seluruh lowongan.":                                      "Fetching is limited each week; visit the source website for all listings.",
	"Halaman publik belum menyediakan data lowongan yang dapat dibaca. Buka situs sumber untuk melihat seluruh lowongan.": "The public page does not provide readable job data. Visit the source website to see all listings.",
	"Situs membatasi akses otomatis. Pengambilan dihentikan sampai jadwal berikutnya.":                                    "The site limits automated access. Fetching is paused until the next scheduled run.",
	"Pembaruan terakhir belum berhasil. Data tersimpan tetap tersedia selama masih berlaku.":                              "The latest update failed. Saved data remains available while it is still valid.",
}

const accountSourceNameSeparator = " · "

func translateJobSourceName(name string) (string, bool) {
	index := strings.LastIndex(name, accountSourceNameSeparator)
	if index < 0 {
		return name, false
	}
	translated, ok := accountJobSourceSuffixEnglish[name[index+len(accountSourceNameSeparator):]]
	if !ok {
		return name, false
	}
	return name[:index] + accountSourceNameSeparator + translated, true
}

// translateJobsPayload localizes /v1/jobs/meta option labels, the directory
// source list (`name` suffix, `scope`, `reason`) and `sourceName` on listings.
// Job titles, companies, descriptions, qualifications, benefits, salary text
// and locations come from external postings and are never translated; `status`
// and `mode` are client logic keys.
func translateJobsPayload(payload any) bool {
	return walkAccountMaps(payload, func(m map[string]any) bool {
		changed := false
		for group, labels := range accountJobOptionEnglish {
			options, ok := m[group].([]any)
			if !ok {
				continue
			}
			for _, item := range options {
				option, ok := item.(map[string]any)
				if !ok {
					continue
				}
				id, ok := option["id"].(string)
				if !ok {
					continue
				}
				if label, found := labels[id]; found {
					changed = setAccountField(option, "label", label) || changed
				}
			}
		}
		if name, ok := m["sourceName"].(string); ok {
			if translated, found := translateJobSourceName(name); found {
				m["sourceName"] = translated
				changed = true
			}
		}
		if _, isSource := m["mode"].(string); isSource {
			if name, ok := m["name"].(string); ok {
				if translated, found := translateJobSourceName(name); found {
					m["name"] = translated
					changed = true
				}
			}
			changed = translateAccountField(m, "scope", accountJobSourceScopeEnglish) || changed
			changed = translateAccountField(m, "reason", accountJobSourceReasonEnglish) || changed
		}
		return changed
	})
}

// ---------------------------------------------------------------------------
// /v1/anki*
// ---------------------------------------------------------------------------

var accountAnkiDeckFields = []string{"title", "description", "instructions", "unavailableReason", "meaningLanguage"}

// accountAnkiPersonalEnglish covers the only app-authored text of a personal
// (imported) deck; its title and description belong to the user.
var accountAnkiPersonalEnglish = map[string]string{
	"Ingat jawaban sebelum membalik kartu; cocokkan dengan semua jawaban yang tersedia.": "Recall the answer before flipping the card; compare it with all available answers.",
}

var accountAnkiNoticeEnglish = map[string]string{
	"Data dan arti mengikuti sumber Kotoba. Sebagian deck memuat atribusi sumber pihak ketiga pada detailnya. Tautan media eksternal tetap mengikuti ketersediaan penyedianya.": "Data and meanings follow the Kotoba source. Some decks include third-party source attribution in their details. External media links remain subject to their providers' availability.",
}

// translateAnkiPayload localizes built-in deck metadata, the credits notice and
// the study-session review intervals. `category` is deliberately not translated:
// the client picks deck icons and filters decks by the Indonesian category
// value. Card content (`front`, `answers`, `meaning`), `originalName`, source
// attribution and the deck `preview` are never touched, and personal decks keep
// their user-entered title and description.
func translateAnkiPayload(payload any) bool {
	return walkAccountMaps(payload, func(m map[string]any) bool {
		changed := false
		personal, _ := m["personal"].(bool)
		for _, key := range []string{"id", "deckId"} {
			if id, ok := m[key].(string); ok && strings.HasPrefix(id, "user-") {
				personal = true
			}
		}
		if personal {
			changed = translateAccountField(m, "instructions", accountAnkiPersonalEnglish) || changed
		} else {
			for _, key := range accountAnkiDeckFields {
				changed = translateAccountField(m, key, accountAnkiCatalogEnglish) || changed
			}
			changed = translateAccountField(m, "instructions", accountAnkiPersonalEnglish) || changed
		}
		changed = translateAccountField(m, "notice", accountAnkiNoticeEnglish) || changed
		if intervals, ok := m["intervals"].(map[string]any); ok {
			for rating, value := range intervals {
				if text, ok := value.(string); ok {
					if translated, found := translateAnkiInterval(text); found {
						intervals[rating] = translated
						changed = true
					}
				}
			}
		}
		return changed
	}, "preview")
}

// translateAnkiInterval converts the study-session interval labels built in
// internal/store/anki.go ("1 menit", "±3 jam", "4 hari"). The client displays
// them verbatim.
func translateAnkiInterval(text string) (string, bool) {
	if number, ok := strings.CutSuffix(text, " menit"); ok && isAccountDigits(number) {
		return number + " min", true
	}
	if rest, ok := strings.CutPrefix(text, "±"); ok {
		if number, ok := strings.CutSuffix(rest, " jam"); ok && isAccountDigits(number) {
			return "±" + number + " hr", true
		}
	}
	if number, ok := strings.CutSuffix(text, " hari"); ok && isAccountDigits(number) {
		if number == "1" {
			return "1 day", true
		}
		return number + " days", true
	}
	return text, false
}

// ---------------------------------------------------------------------------
// /v1/branding
// ---------------------------------------------------------------------------

var accountBrandingTaglineEnglish = map[string]string{
	"Selangkah lebih dekat dengan Jepang": "One step closer to Japan",
}

// translateBrandingPayload localizes the default tagline. `appName`, image
// URLs and the background color are not text; an administrator-entered
// tagline is unknown and stays unchanged.
func translateBrandingPayload(payload any) bool {
	return walkAccountMaps(payload, func(m map[string]any) bool {
		return translateAccountField(m, "tagline", accountBrandingTaglineEnglish)
	})
}
