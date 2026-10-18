package dictionary

import (
	"net/url"
	"strconv"
	"strings"
	"unicode"
	"unicode/utf8"

	"golang.org/x/text/unicode/norm"
	"golang.org/x/text/width"
)

// Auto deliberately recognizes a reviewed vocabulary only. A Latin word may
// also be English; use Japanese mode to transliterate an arbitrary romaji word.
var autoJapanese = map[string]string{
	"konnichiwa": "こんにちは", "konnichiha": "こんにちは",
	"konbanwa": "こんばんは", "konbanha": "こんばんは",
	"arigatou": "ありがとう", "arigato": "ありがと", "ohayou": "おはよう",
	"sayounara": "さようなら", "sayonara": "さよなら", "sumimasen": "すみません",
	"neko": "ねこ", "inu": "いぬ", "gakkou": "がっこう", "nihongo": "にほんご",
	"taberu": "たべる", "watashi": "わたし", "sensei": "せんせい",
	"tōkyō": "とうきょう", "toukyou": "とうきょう",
	"ōsaka": "おおさか", "oosaka": "おおさか",
}

func canonicalQuery(query, mode string) (string, string, error) {
	if mode == "" {
		mode = "auto"
	}
	if mode != "auto" && mode != "japanese" && mode != "english" {
		return "", "", ErrInvalid
	}
	q := norm.NFKC.String(width.Fold.String(strings.TrimSpace(query)))
	// Advanced Jisho expressions remain raw and have their own cache identity.
	advanced := strings.ContainsAny(q, "#*?\"[]:|")
	if mode == "english" {
		if strings.ContainsAny(q, "\"\\") {
			return "", "", ErrInvalid
		}
		return strings.ToLower(q), "english", nil
	}
	if advanced {
		return q, "raw", nil
	}
	lower := strings.ToLower(q)
	if alias, ok := autoJapanese[lower]; ok {
		return alias, "japanese", nil
	}
	hasJapanese := false
	for _, r := range q {
		if unicode.Is(unicode.Han, r) || unicode.Is(unicode.Hiragana, r) || unicode.Is(unicode.Katakana, r) {
			hasJapanese = true
		}
	}
	if mode == "japanese" && !hasJapanese {
		kana, ok := romajiToKana(lower)
		if !ok {
			return "", "", ErrInvalid
		}
		q = kana
	} else if mode == "auto" && !hasJapanese {
		return lower, "raw", nil
	}
	q = strings.Map(func(r rune) rune {
		if r >= 'ァ' && r <= 'ヶ' {
			return r - 0x60
		}
		return r
	}, q)
	// These are spelling aliases for greetings, not a general wa -> ha rule.
	if q == "こんにちわ" {
		q = "こんにちは"
	}
	if q == "こんばんわ" {
		q = "こんばんは"
	}
	return norm.NFC.String(q), "japanese", nil
}

var syllables = func() map[string]string {
	m := map[string]string{}
	groups := []string{
		"a あ i い u う e え o お",
		"ka か ki き ku く ke け ko こ ga が gi ぎ gu ぐ ge げ go ご",
		"sa さ shi し si し su す se せ so そ za ざ ji じ zi じ zu ず ze ぜ zo ぞ",
		"ta た chi ち ti ち tsu つ tu つ te て to と da だ di ぢ du づ de で do ど",
		"na な ni に nu ぬ ne ね no の ha は hi ひ fu ふ hu ふ he へ ho ほ",
		"ba ば bi び bu ぶ be べ bo ぼ pa ぱ pi ぴ pu ぷ pe ぺ po ぽ",
		"ma ま mi み mu む me め mo も ya や yu ゆ yo よ",
		"ra ら ri り ru る re れ ro ろ wa わ wo を vu ゔ",
		"kya きゃ kyu きゅ kyo きょ gya ぎゃ gyu ぎゅ gyo ぎょ",
		"sha しゃ shu しゅ sho しょ sya しゃ syu しゅ syo しょ",
		"cha ちゃ chu ちゅ cho ちょ tya ちゃ tyu ちゅ tyo ちょ",
		"ja じゃ ju じゅ jo じょ jya じゃ jyu じゅ jyo じょ zya じゃ zyu じゅ zyo じょ",
		"nya にゃ nyu にゅ nyo にょ hya ひゃ hyu ひゅ hyo ひょ",
		"bya びゃ byu びゅ byo びょ pya ぴゃ pyu ぴゅ pyo ぴょ",
		"mya みゃ myu みゅ myo みょ rya りゃ ryu りゅ ryo りょ",
		"fa ふぁ fi ふぃ fe ふぇ fo ふぉ va ゔぁ vi ゔぃ ve ゔぇ vo ゔぉ",
		"xa ぁ xi ぃ xu ぅ xe ぇ xo ぉ xya ゃ xyu ゅ xyo ょ xtsu っ xtu っ",
		"la ぁ li ぃ lu ぅ le ぇ lo ぉ lya ゃ lyu ゅ lyo ょ ltsu っ ltu っ",
	}
	for _, g := range groups {
		f := strings.Fields(g)
		for i := 0; i < len(f); i += 2 {
			m[f[i]] = f[i+1]
		}
	}
	return m
}()

func romajiToKana(s string) (string, bool) {
	// A macron marks a long vowel. Preserve its length as ー rather than
	// guessing whether a dictionary word spells it with う or お.
	s = strings.NewReplacer("ā", "a-", "ī", "i-", "ū", "u-", "ē", "e-", "ō", "o-").Replace(s)
	var out strings.Builder
	for len(s) > 0 {
		if strings.HasPrefix(s, "tch") {
			out.WriteRune('っ')
			s = s[1:]
			continue
		}
		if s[0] == ' ' || s[0] == '-' {
			if s[0] == '-' {
				out.WriteRune('ー')
			} else {
				out.WriteByte(' ')
			}
			s = s[1:]
			continue
		}
		if s[0] == 'n' && (len(s) == 1 || (len(s) > 1 && !strings.ContainsRune("aiueoy", rune(s[1])))) {
			out.WriteRune('ん')
			if len(s) > 1 && s[1] == '\'' {
				s = s[2:]
			} else if s == "nn" {
				s = ""
			} else {
				s = s[1:]
			}
			continue
		}
		if len(s) > 1 && s[0] == s[1] && strings.ContainsRune("bcdfghjklmpqrstvwxyz", rune(s[0])) {
			out.WriteRune('っ')
			s = s[1:]
			continue
		}
		matched := false
		for n := 4; n >= 1; n-- {
			if len(s) >= n {
				if kana, ok := syllables[s[:n]]; ok {
					out.WriteString(kana)
					s = s[n:]
					matched = true
					break
				}
			}
		}
		if !matched {
			return "", false
		}
	}
	return out.String(), out.Len() > 0
}

func searchEndpoint(query string, page int, jlpt string, common bool, mode string) (string, error) {
	q, kind, err := canonicalQuery(query, mode)
	if err != nil {
		return "", err
	}
	if utf8.RuneCountInString(q) > 100 {
		return "", ErrInvalid
	}
	keyword := q
	if kind == "english" {
		keyword = "\"" + q + "\""
	}
	if jlpt != "" {
		keyword += " #jlpt-" + strings.ToLower(jlpt)
	}
	if common {
		keyword += " #common"
	}
	u := url.URL{Scheme: "https", Host: "jisho.org", Path: "/api/v1/search/words", RawQuery: url.Values{"keyword": {keyword}, "page": {strconv.Itoa(page)}}.Encode(), Fragment: "v2|" + kind + "|" + jlpt + "|" + strconv.FormatBool(common)}
	return u.String(), nil
}
