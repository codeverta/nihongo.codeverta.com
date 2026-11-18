# Grammar & Kanji Level Lists with PDF Export

## Problem Statement

Learners often want to see every grammar pattern or every kanji of a JLPT level at a glance: to review, to check what is left, or to keep a cheat sheet. Today the app only shows lessons; pattern names appear inside lesson cards and kanji are visible only after opening each lesson (N1 alone has 247 kanji lessons).

## Solution

Add a **Daftar Pola** tab to the Grammar course screen and a **Daftar Kanji** tab to the Kanji course screen for N5–N1, available to every signed-in account. Each list is searchable and opens a detail sheet with a "Buka lesson" action. A PDF of the whole level can be downloaded by Lifetime Access users only; free users see a locked button that leads to the Lifetime page.

## User Stories

1. As a learner, I want to see all grammar patterns of a level in one list, so that I can review them quickly.
2. As a learner, I want to search patterns by pattern or meaning.
3. As a learner, I want to tap a pattern to see its formation, explanation and examples, and jump to its lesson.
4. As a learner, I want to see all kanji of a level as a grid with meanings.
5. As a learner, I want to search kanji by character, meaning or reading.
6. As a learner, I want to tap a kanji to see readings, strokes, mnemonic and example words, and jump to its lesson.
7. As a learner, I want the lists to work offline after I have opened them once.
8. As a Lifetime user, I want to download a PDF cheat sheet of a level's grammar or kanji.
9. As a free user, I want to understand that PDF download is a Lifetime benefit.
10. As an English user, I want explanations in English while Japanese text stays Japanese.

## Implementation Decisions

- Tabs: Grammar course screen gets [Lesson] [Daftar Pola]; Kanji course screen gets [Lesson] [Daftar Kanji]. Levels N5–N1. Vocabulary and Reading are out of scope.
- Viewing lists is open to all signed-in accounts; only PDF download requires Lifetime Access.
- New endpoints (requireAuth, localized like lesson content, cached by the app's learning cache):
  - `GET /v1/grammar/levels/{level}/points` — every point with lesson ID/title, pattern, meaning, formation, explanation, examples.
  - `GET /v1/kanji/levels/{level}/characters` — every character with lesson ID/title, meaning, on, kun, strokes, mnemonic, examples.
- PDF endpoints: `GET /v1/grammar/levels/{level}/pdf` and `GET /v1/kanji/levels/{level}/pdf`.
  - Lifetime is checked on the server with the same rule as manga (premium tier and Lifetime); otherwise 402 `premium_required`.
  - Generated in Go with an embedded, subset Noto Sans JP font (SIL OFL); cached per (type, level, locale, content hash).
  - Grammar PDF: grouped by lesson; pattern, meaning, formation and 1–2 examples (Japanese, reading, meaning). No long explanations or quiz.
  - Kanji PDF: table per lesson; kanji, meaning, on, kun, strokes, 1–2 example words. No mnemonics.
  - One file per level with a table of contents; footer "Nihongo Master · nihongo.codeverta.com · Halaman X/Y"; no personal data.
- App delivery reuses the Anki export helper: share sheet on mobile, save dialog on desktop. File names `NihongoMaster-Grammar-N5.pdf`, `NihongoMaster-Kanji-N1.pdf`.
- Free users: the PDF button shows a lock and opens a Lifetime prompt (no ad option). Offline: the button is disabled with "Butuh koneksi internet".
- Font license attribution is added to the app's licenses page.

## Testing Decisions

- Go handler tests: list endpoints return every point/character of a level in lesson order, localized for `en`; PDF returns 401 without auth, 402 for free accounts, `application/pdf` for Lifetime; invalid level returns 400.
- Flutter widget tests: tabs render lists, search filters, detail sheet opens and links to the lesson, PDF button shows the Lifetime prompt for free users and shares a file for Lifetime users, offline disables the button.

## Out of Scope

- Vocabulary and Reading lists.
- Rewarded-ad unlock for PDF.
- Splitting large PDFs.
