# JLPT vocabulary lists and N4 data research

## Findings

The JLPT does not publish an official fixed vocabulary list for the current test. The official JLPT guide FAQ says the post-2010 "Test Content Specifications" containing vocabulary, kanji, and grammar lists will not be published. It recommends level-competence summaries, question composition, sample questions, and past materials as preparation references. Therefore, third-party lists should be described as community estimates or study selections, not as authoritative complete N4 coverage.

- Official guide FAQ (PDF, Q7–Q9): <https://www.jlpt.jp/tw/reference/pdf/guidebook_s_e.pdf>
- Current official FAQ / question composition entry point: <https://www.jlpt.jp/e/faq/>
- Official sample question/reference downloads: <https://ns.jlpt.jp/e/reference/materials.html>

## Usable N4 vocabulary data

[OpenJLPT](https://github.com/evanclan/OpenJLPT) is the strongest ready-to-use candidate found. Its repository provides N5–N1 vocabulary in JSON/CSV/SQLite; the current README reports 630 N4 words, each assigned to its easiest listed level. The N4 JSON is directly available at <https://github.com/evanclan/OpenJLPT/blob/main/data/json/vocab/n4.json>. Its README says levels follow Jonathan Waller's community lists (Tanos/Jisho convention), not an official JLPT list, and labels them approximations. This gives a practical source for curriculum selection, subject to review for this app's scope and Indonesian glosses.

OpenJLPT states its dataset and code are CC BY-SA 4.0, with upstream attribution for JMdict, KANJIDIC2, Waller's lists, and Tatoeba. Its notice file gives the required credit: <https://github.com/evanclan/OpenJLPT/blob/main/NOTICE.md>. The ShareAlike condition applies to redistributed adapted data, so check compatibility before copying the dataset or modified entries into a separately licensed app dataset. The README also says a one-line About-screen credit is sufficient for its package; this does not eliminate ShareAlike obligations.

- OpenJLPT README and level caveat: <https://github.com/evanclan/OpenJLPT>
- N4 JSON: <https://github.com/evanclan/OpenJLPT/blob/main/data/json/vocab/n4.json>
- OpenJLPT license: <https://github.com/evanclan/OpenJLPT/blob/main/LICENSE>
- OpenJLPT upstream credits: <https://github.com/evanclan/OpenJLPT/blob/main/NOTICE.md>

## License and attribution notes

- JMdict/EDRDG data is under CC BY-SA 4.0. The EDRDG license says derived data files are covered and smartphone/tablet apps must acknowledge the source, for example on an About or Sources screen. Its suggested JMdict project URL is <https://www.edrdg.org/wiki/index.php/JMdict-EDICT_Dictionary_Project>. License details: <https://www.edrdg.org/edrdg/licence.html>.
- Tatoeba text downloads are CC BY 2.0 FR by default, with some sentences separately marked CC0; the download page carries per-sentence author/license information and warns audio has separate contributor-selected licensing. If importing examples, retain attribution and inspect the actual row/license. Sources: <https://tatoeba.org/en/downloads> and <https://en.wiki.tatoeba.org/articles/show/using-the-tatoeba-corpus>.
- Tanos/Jonathan Waller lists are commonly used as a level reference, but this research did not find an explicit reuse license on the list itself. Treat it as a selection/reference source, not a dataset to copy, unless its terms or permission are established. List entry point: <https://www.tanos.co.uk/jlpt/skills/vocab/>. OpenJLPT documents its own sourcing of Waller lists and its own redistribution license; that license is not evidence that upstream source material is independently unrestricted.

## Recommendation for this repository

Use the OpenJLPT N4 JSON as a candidate for review and cross-checking, while labeling N4 as an approximate community level. If incorporating its data, follow OpenJLPT's NOTICE and CC BY-SA 4.0 terms, including ShareAlike for redistributed adaptations. Continue writing original Indonesian translations, explanations, and teaching sentences where feasible. For example sentences sourced from Tatoeba, preserve sentence-level IDs/authorship and license attribution; avoid reusing audio without checking its separate license. Do not describe any third-party list as the official JLPT list or as exhaustive coverage.
