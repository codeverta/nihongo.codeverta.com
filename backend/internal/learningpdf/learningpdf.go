// Package learningpdf renders level-wide grammar and kanji cheat sheets.
// Noto Sans JP (SIL OFL 1.1, see fonts/NotoSansJP-OFL.txt) is embedded and
// subset into every PDF. The Regular and Bold files are static instances of
// the Google Fonts variable font (wght 400 and 700).
package learningpdf

import (
	"bytes"
	_ "embed"
	"fmt"
	"strings"

	"github.com/go-pdf/fpdf"

	"nihongo-master-backend/internal/domain"
)

var (
	//go:embed fonts/NotoSansJP-Regular.ttf
	notoSansJPRegular []byte
	//go:embed fonts/NotoSansJP-Bold.ttf
	notoSansJPBold []byte
)

const (
	fontName   = "NotoSansJP"
	pageWidth  = 210.0
	margin     = 14.0
	contentW   = pageWidth - 2*margin
	footerSite = "Nihongo Master · nihongo.codeverta.com"
)

type labels struct {
	grammarTitle, kanjiTitle, contents, page, formation, meaning, on, kun, strokes, examples, lesson string
}

var copyByLocale = map[string]labels{
	"id": {
		grammarTitle: "Daftar Grammar JLPT %s", kanjiTitle: "Daftar Kanji JLPT %s", contents: "Daftar isi",
		page: "Halaman", formation: "Pembentukan", meaning: "Arti", on: "On", kun: "Kun",
		strokes: "Goresan", examples: "Contoh", lesson: "Lesson",
	},
	"en": {
		grammarTitle: "JLPT %s Grammar List", kanjiTitle: "JLPT %s Kanji List", contents: "Contents",
		page: "Page", formation: "Formation", meaning: "Meaning", on: "On", kun: "Kun",
		strokes: "Strokes", examples: "Examples", lesson: "Lesson",
	},
}

func labelsFor(locale string) labels {
	if l, ok := copyByLocale[locale]; ok {
		return l
	}
	return copyByLocale["id"]
}

// tocEntry is one lesson line in the table of contents.
type tocEntry struct {
	title string
	page  int
	link  int
}

// renderer draws the content of one document type into a PDF.
type renderer interface {
	title(l labels) string
	lessonCount() int
	lessonTitle(i int, l labels) string
	drawLesson(pdf *fpdf.Fpdf, i int, l labels)
}

const tocLinesPerPage = 38

func tocPages(n int) int {
	// The first page also holds the document title (about 6 lines).
	return (n+6)/tocLinesPerPage + 1
}

func newPDF(l labels, title string) (*fpdf.Fpdf, error) {
	pdf := fpdf.New("P", "mm", "A4", "")
	pdf.SetMargins(margin, margin, margin)
	pdf.SetAutoPageBreak(true, 16)
	pdf.AddUTF8FontFromBytes(fontName, "", notoSansJPRegular)
	pdf.AddUTF8FontFromBytes(fontName, "B", notoSansJPBold)
	if err := pdf.Error(); err != nil {
		return nil, err
	}
	pdf.SetTitle(title, true)
	pdf.SetAuthor("Nihongo Master", true)
	pdf.SetCreator("Nihongo Master", true)
	pdf.AliasNbPages("{nb}")
	pdf.SetFooterFunc(func() {
		pdf.SetY(-11)
		pdf.SetFont(fontName, "", 7.5)
		pdf.SetTextColor(120, 120, 130)
		pdf.CellFormat(0, 5, fmt.Sprintf("%s · %s %d/{nb}", footerSite, l.page, pdf.PageNo()), "", 0, "C", false, 0, "")
	})
	return pdf, nil
}

// render lays the document out twice: the first pass finds the page where
// each lesson starts so the table of contents can show page numbers.
func render(r renderer, locale string) ([]byte, error) {
	l := labelsFor(locale)
	pages, err := layout(r, l, nil)
	if err != nil {
		return nil, err
	}
	var out bytes.Buffer
	pdf, _, err := build(r, l, pages)
	if err != nil {
		return nil, err
	}
	if err := pdf.Output(&out); err != nil {
		return nil, err
	}
	return out.Bytes(), nil
}

func layout(r renderer, l labels, pages []int) ([]int, error) {
	_, starts, err := build(r, l, pages)
	return starts, err
}

func build(r renderer, l labels, pages []int) (*fpdf.Fpdf, []int, error) {
	title := r.title(l)
	pdf, err := newPDF(l, title)
	if err != nil {
		return nil, nil, err
	}
	n := r.lessonCount()
	toc := tocPages(n)
	links := make([]int, n)
	for i := range links {
		links[i] = pdf.AddLink()
	}

	// Title and table of contents. Exactly `toc` pages are reserved so the
	// page numbers found in the first pass stay valid in the second.
	pdf.AddPage()
	pdf.SetFont(fontName, "B", 20)
	pdf.SetTextColor(30, 32, 64)
	pdf.MultiCell(0, 10, title, "", "L", false)
	pdf.Ln(2)
	pdf.SetFont(fontName, "B", 12)
	pdf.SetTextColor(200, 70, 90)
	pdf.CellFormat(0, 8, l.contents, "", 1, "L", false, 0, "")
	pdf.SetFont(fontName, "", 9.5)
	pdf.SetTextColor(40, 40, 50)
	line := 6
	tocPage := 1
	for i := 0; i < n; i++ {
		if line >= tocLinesPerPage {
			pdf.AddPage()
			tocPage++
			line = 0
		}
		pageText := ""
		if pages != nil && i < len(pages) {
			pageText = fmt.Sprint(pages[i])
		}
		pdf.CellFormat(contentW-14, 6, truncate(r.lessonTitle(i, l), 90), "", 0, "L", false, links[i], "")
		pdf.CellFormat(14, 6, pageText, "", 1, "R", false, links[i], "")
		line++
	}
	for tocPage < toc {
		pdf.AddPage()
		tocPage++
	}

	starts := make([]int, n)
	for i := 0; i < n; i++ {
		// Each lesson starts on a new line; a page break is added only when
		// the lesson header would otherwise be stranded at the bottom.
		if i == 0 || pdf.GetY() > 250 {
			pdf.AddPage()
		} else {
			pdf.Ln(4)
		}
		starts[i] = pdf.PageNo()
		pdf.SetLink(links[i], pdf.GetY(), -1)
		lessonHeader(pdf, r.lessonTitle(i, l))
		r.drawLesson(pdf, i, l)
	}
	return pdf, starts, pdf.Error()
}

func lessonHeader(pdf *fpdf.Fpdf, title string) {
	pdf.SetFillColor(246, 232, 236)
	pdf.SetTextColor(150, 40, 70)
	pdf.SetFont(fontName, "B", 11.5)
	pdf.CellFormat(0, 8, " "+title, "", 1, "L", true, 0, "")
	pdf.Ln(1.5)
}

func truncate(s string, max int) string {
	r := []rune(s)
	if len(r) <= max {
		return s
	}
	return string(r[:max-1]) + "…"
}

func join(parts ...string) string {
	kept := parts[:0]
	for _, p := range parts {
		if p = strings.TrimSpace(p); p != "" {
			kept = append(kept, p)
		}
	}
	return strings.Join(kept, " · ")
}

func exampleLine(japanese, reading, meaning string) string {
	text := strings.TrimSpace(japanese)
	if r := strings.TrimSpace(reading); r != "" && r != text {
		text += " (" + r + ")"
	}
	if m := strings.TrimSpace(meaning); m != "" {
		text += " — " + m
	}
	return text
}

// ---------- grammar ----------

type grammarDoc struct {
	level   string
	lessons []domain.GrammarLesson
}

func (d grammarDoc) title(l labels) string { return fmt.Sprintf(l.grammarTitle, d.level) }
func (d grammarDoc) lessonCount() int      { return len(d.lessons) }
func (d grammarDoc) lessonTitle(i int, l labels) string {
	lesson := d.lessons[i]
	return fmt.Sprintf("%s %d · %s", l.lesson, i+1, lesson.Title)
}

func (d grammarDoc) drawLesson(pdf *fpdf.Fpdf, i int, l labels) {
	for _, point := range d.lessons[i].Points {
		if pdf.GetY() > 262 {
			pdf.AddPage()
		}
		pdf.SetFont(fontName, "B", 11)
		pdf.SetTextColor(30, 32, 64)
		pdf.MultiCell(0, 6, join(point.Pattern, point.Meaning), "", "L", false)
		pdf.SetFont(fontName, "", 8.8)
		pdf.SetTextColor(70, 70, 85)
		if f := strings.TrimSpace(point.Formation); f != "" {
			pdf.MultiCell(0, 4.6, l.formation+": "+f, "", "L", false)
		}
		for j, e := range point.Examples {
			if j == 2 {
				break
			}
			pdf.SetX(margin + 3)
			pdf.MultiCell(contentW-3, 4.6, "・"+exampleLine(e.Japanese, e.Reading, e.Meaning), "", "L", false)
		}
		pdf.Ln(2.2)
	}
}

// Grammar renders the grammar list of one level.
func Grammar(level, locale string, lessons []domain.GrammarLesson) ([]byte, error) {
	return render(grammarDoc{level: level, lessons: lessons}, locale)
}

// ---------- kanji ----------

type kanjiDoc struct {
	level   string
	lessons []domain.KanjiLesson
}

func (d kanjiDoc) title(l labels) string { return fmt.Sprintf(l.kanjiTitle, d.level) }
func (d kanjiDoc) lessonCount() int      { return len(d.lessons) }
func (d kanjiDoc) lessonTitle(i int, l labels) string {
	return fmt.Sprintf("%s %d · %s", l.lesson, i+1, d.lessons[i].Title)
}

var kanjiColumns = []float64{16, 36, 30, 30, 17, contentW - 16 - 36 - 30 - 30 - 17}

// style makes the table header row bold.
func style(header bool) string {
	if header {
		return "B"
	}
	return ""
}

func (d kanjiDoc) drawLesson(pdf *fpdf.Fpdf, i int, l labels) {
	header := []string{"", l.meaning, l.on, l.kun, l.strokes, l.examples}
	drawRow := func(cells []string, sizes []float64, lineH float64, fill bool) {
		lines := 1
		for c, text := range cells {
			pdf.SetFont(fontName, style(fill), sizes[c])
			if n := len(pdf.SplitText(text, kanjiColumns[c]-2)); n > lines {
				lines = n
			}
		}
		height := float64(lines)*lineH + 2
		if pdf.GetY()+height > 281 {
			pdf.AddPage()
		}
		x, y := margin, pdf.GetY()
		for c, text := range cells {
			pdf.SetXY(x, y)
			if fill {
				pdf.SetFillColor(240, 240, 246)
				pdf.Rect(x, y, kanjiColumns[c], height, "F")
			}
			pdf.SetFont(fontName, style(fill), sizes[c])
			pdf.SetXY(x+1, y+1)
			cellLineH := lineH
			if c == 0 && !fill {
				cellLineH = height - 2
			}
			pdf.MultiCell(kanjiColumns[c]-2, cellLineH, text, "", map[bool]string{true: "C", false: "L"}[c == 0], false)
			x += kanjiColumns[c]
		}
		pdf.SetDrawColor(225, 225, 232)
		pdf.Line(margin, y+height, margin+contentW, y+height)
		pdf.SetXY(margin, y+height)
	}
	pdf.SetTextColor(90, 90, 105)
	drawRow(header, []float64{8, 8, 8, 8, 8, 8}, 4.5, true)
	pdf.SetTextColor(30, 32, 64)
	for _, c := range d.lessons[i].Characters {
		examples := make([]string, 0, 2)
		for j, e := range c.Examples {
			if j == 2 {
				break
			}
			examples = append(examples, exampleLine(e.Japanese, e.Reading, e.Meaning))
		}
		strokes := ""
		if c.Strokes > 0 {
			strokes = fmt.Sprint(c.Strokes)
		}
		drawRow(
			[]string{c.Character, c.Meaning, c.OnReading, c.KunReading, strokes, strings.Join(examples, "\n")},
			[]float64{18, 8.5, 8.5, 8.5, 8.5, 8},
			4.4, false,
		)
	}
}

// Kanji renders the kanji list of one level.
func Kanji(level, locale string, lessons []domain.KanjiLesson) ([]byte, error) {
	return render(kanjiDoc{level: level, lessons: lessons}, locale)
}
