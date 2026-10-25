# Public English content

This directory accompanies the English public site and store listing. Its copy and the HTML under backend/internal/api/site/index.en.html and backend/internal/api/legal/en/ are drafts for internal QA.

The intended public routes are:

| Content | English route or asset |
| --- | --- |
| Marketing page | /en/ → backend/internal/api/site/index.en.html |
| Legal and policy index | /en/legal → backend/internal/api/legal/en/index.html |
| Privacy policy | /en/privacy-policy → backend/internal/api/legal/en/privacy.html |
| Terms of service | /en/terms-of-service → backend/internal/api/legal/en/terms.html |
| Purchase and refund policy | /en/refund-policy → backend/internal/api/legal/en/refund.html |
| Account deletion | /en/delete-account → backend/internal/api/legal/en/deletion.html |
| Support | /en/support → backend/internal/api/legal/en/support.html |
| Data safety summary | /en/data-safety → backend/internal/api/legal/en/safety.html |
| Store listing | Google Play and App Store copy in google-play-listing.md |

The current question, option, shared-text, and learning-material inventory has hand-authored English mappings, including bundled SSW curriculum copies. These mappings have not yet received a full human language review. The app, API, and public-page gates default to off. Before enabling public English, have fluent reviewers check the complete user-facing experience and product claims, and have a qualified reviewer check privacy, terms, refund, deletion, and data-safety text against the actual distributed build and applicable requirements. Recheck factual statements whenever app behavior, providers, permissions, purchase products, or retention practices change.

Translate learning content by hand into exact source mappings. Do not use a machine-translation pass or heuristic parser for questions and lesson explanations. Preserve Japanese study strings. The source-of-truth JSON for bundled SSW lessons stays in `assets/content/ssw/`; reviewed English copies belong in `assets/content/ssw/en/` with the same sector filename and structure. The app selects the English copy for English QA and reports that a sector is not ready when its copy is missing; it must never silently show Indonesian lesson content in English mode.

For private QA only, build Flutter with `--dart-define=ENGLISH_QA_PREVIEW=true` and set `ENGLISH_QA_PREVIEW_ENABLED=true` on the private backend. These preview switches do not enable English in public builds or expose English public pages. After content review is complete, a release requires `--dart-define=ENGLISH_PUBLIC_RELEASE=true` in Flutter and `PUBLIC_ENGLISH_CONTENT_ENABLED=true` on the backend. Do not set those public flags as part of this implementation.

Keep the Indonesian pages available at their existing routes. The legal English templates use the same facts and limitations as the current Indonesian pages and preserve the Japanese source text in the marketing page.
