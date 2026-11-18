package learningpdf

import (
	"bytes"
	"fmt"
	"regexp"
	"strings"

	"github.com/go-pdf/fpdf"

	"nihongo-master-backend/internal/domain"
)

// ExamPaper is a printable simulation paper. Reviews carry each question,
// its options in print order, the correct option and the explanation.
type ExamPaper struct {
	Title    string
	Sections []domain.SimulationSection
	Omitted  []string
	Reviews  []domain.SimulationReview
}

type examLabels struct {
	minutes, questions, omitted, answerKey, answer, intro string
}

var examCopy = map[string]examLabels{
	"id": {
		minutes: "menit", questions: "soal",
		omitted:   "Bagian listening tidak disertakan dalam simulasi tertulis ini.",
		answerKey: "Kunci jawaban & pembahasan", answer: "Jawaban",
		intro: "Pilih satu jawaban terbaik (1–4) untuk setiap soal. Kunci jawaban dan pembahasan ada di bagian akhir.",
	},
	"en": {
		minutes: "min", questions: "questions",
		omitted:   "The listening section is not included in this written simulation.",
		answerKey: "Answer key & explanations", answer: "Answer",
		intro: "Choose the best answer (1–4) for each question. The answer key and explanations are at the end.",
	},
}

var (
	furiganaMarkup  = regexp.MustCompile(`\{([^{}|]+)\|([^{}|]+)\}`)
	underlineMarkup = regexp.MustCompile(`\[\[(.+?)\]\]`)
)

// plainExamText turns the app's furigana markup {漢字|かんじ} into
// 漢字（かんじ）. Underline markers are handled by writeExamText.
func plainExamText(text string) string {
	return furiganaMarkup.ReplaceAllString(strings.TrimSpace(text), "$1（$2）")
}

// stripExamMarkup removes all markup, for places that cannot underline.
func stripExamMarkup(text string) string {
	return underlineMarkup.ReplaceAllString(plainExamText(text), "$1")
}

// writeExamText writes text whose [[parts]] are underlined (as in the app),
// wrapping inside the given left indent.
func writeExamText(pdf *fpdf.Fpdf, indent, size, lineH float64, style, text string) {
	text = plainExamText(text)
	pdf.SetLeftMargin(margin + indent)
	pdf.SetX(margin + indent)
	last := 0
	for _, m := range underlineMarkup.FindAllStringSubmatchIndex(text, -1) {
		if m[0] > last {
			pdf.SetFont(fontName, style, size)
			pdf.Write(lineH, text[last:m[0]])
		}
		pdf.SetFont(fontName, style+"U", size)
		pdf.Write(lineH, text[m[2]:m[3]])
		last = m[1]
	}
	pdf.SetFont(fontName, style, size)
	if last < len(text) {
		pdf.Write(lineH, text[last:])
	}
	pdf.Ln(lineH)
	pdf.SetLeftMargin(margin)
}

// Exam renders a simulation paper: questions first, answer key at the end.
func Exam(paper ExamPaper, locale string) ([]byte, error) {
	l := labelsFor(locale)
	el, ok := examCopy[locale]
	if !ok {
		el = examCopy["id"]
	}
	pdf, err := newPDF(l, paper.Title)
	if err != nil {
		return nil, err
	}
	pdf.AddPage()
	pdf.SetFont(fontName, "B", 18)
	pdf.SetTextColor(30, 32, 64)
	pdf.MultiCell(0, 9, paper.Title, "", "L", false)
	pdf.Ln(1)
	pdf.SetFont(fontName, "", 9.5)
	pdf.SetTextColor(70, 70, 85)
	for _, section := range paper.Sections {
		pdf.MultiCell(0, 5.2, fmt.Sprintf("・%s — %d %s, %d %s", section.Name, section.QuestionCount, el.questions, section.DurationSeconds/60, el.minutes), "", "L", false)
	}
	if len(paper.Omitted) > 0 {
		pdf.MultiCell(0, 5.2, el.omitted, "", "L", false)
	}
	pdf.Ln(1)
	pdf.MultiCell(0, 5.2, el.intro, "", "L", false)
	pdf.Ln(3)

	names := map[string]string{}
	for _, section := range paper.Sections {
		names[section.ID] = section.Name
	}
	section, instructions, passage := "", "", ""
	for _, review := range paper.Reviews {
		q := review.Question
		if q.Section != section {
			section = q.Section
			if pdf.GetY() > 240 {
				pdf.AddPage()
			}
			name := names[section]
			if name == "" {
				name = section
			}
			lessonHeader(pdf, name)
			instructions, passage = "", ""
		}
		if text := strings.TrimSpace(q.Instructions); text != "" && text != instructions {
			instructions = text
			pdf.SetTextColor(150, 40, 70)
			writeExamText(pdf, 0, 9.5, 5, "B", text)
			pdf.Ln(1)
		}
		if text := strings.TrimSpace(q.Passage); text != "" && text != passage {
			passage = text
			pdf.SetTextColor(30, 32, 64)
			page, top := pdf.PageNo(), pdf.GetY()
			writeExamText(pdf, 4, 9.8, 5.4, "", text)
			// A rule on the left marks the reading passage (same page only).
			if bottom := pdf.GetY(); pdf.PageNo() == page && bottom > top {
				pdf.SetDrawColor(200, 200, 215)
				pdf.SetLineWidth(0.8)
				pdf.Line(margin+1, top, margin+1, bottom)
				pdf.SetLineWidth(0.2)
			}
			pdf.Ln(2)
		}
		if pdf.GetY() > 255 {
			pdf.AddPage()
		}
		pdf.SetTextColor(30, 32, 64)
		writeExamText(pdf, 0, 10.5, 5.6, "B", fmt.Sprintf("%d. %s", q.Number, q.Prompt))
		for i, option := range q.Options {
			writeExamText(pdf, 5, 10, 5.4, "", fmt.Sprintf("%d  %s", i+1, option.Text))
		}
		pdf.Ln(2.5)
	}

	pdf.AddPage()
	lessonHeader(pdf, el.answerKey)
	for _, review := range paper.Reviews {
		q := review.Question
		answer, answerText := 0, ""
		for i, option := range q.Options {
			if option.ID == review.CorrectOptionID {
				answer, answerText = i+1, stripExamMarkup(option.Text)
			}
		}
		if pdf.GetY() > 268 {
			pdf.AddPage()
		}
		pdf.SetFont(fontName, "B", 9.8)
		pdf.SetTextColor(30, 32, 64)
		pdf.MultiCell(0, 5.2, fmt.Sprintf("%d. %s: %d — %s", q.Number, el.answer, answer, answerText), "", "L", false)
		if text := stripExamMarkup(review.Explanation); text != "" {
			pdf.SetFont(fontName, "", 8.8)
			pdf.SetTextColor(70, 70, 85)
			pdf.SetX(margin + 5)
			pdf.MultiCell(contentW-5, 4.6, text, "", "L", false)
		}
		pdf.Ln(1.5)
	}
	if err := pdf.Error(); err != nil {
		return nil, err
	}
	var out bytes.Buffer
	if err := pdf.Output(&out); err != nil {
		return nil, err
	}
	return out.Bytes(), nil
}
