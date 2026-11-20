# AI Lesson Chat

## Problem Statement

Learners studying a grammar, vocabulary, kanji, or reading lesson often have follow-up questions the lesson text does not answer: why a pattern is used, how it differs from a similar one, more example sentences, or whether a sentence they wrote is correct. Today they must leave the app to ask elsewhere, losing the lesson context. They also have no quick way to get a concise recap of a lesson before moving on.

The feature must stay affordable at scale, must not let the AI be abused as a free general-purpose chatbot, and must resist prompt injection.

## Solution

Add an AI chat, powered by DeepSeek, to every grammar, vocabulary, kanji, and reading lesson. The AI is scoped to Japanese learning and receives the current lesson as context. Quick-action chips cover the most common requests (summary, grammar explanation, example sentences, a 5-question quiz, and sentence correction), and answers stream in as they are generated.

All model calls go through the Go backend, which owns the API key, the system prompt, the lesson context, the conversation history, and quota enforcement. Free users get 3 messages per day; after that, on Android, each extra message requires watching an opt-in rewarded ad, up to a daily cap. Lifetime users get 15 messages per day; after that, on Android, they may choose to watch a rewarded ad that adds 5 more messages, up to 4 ads per day (all three numbers are admin settings). Lifetime users never see banner or full-screen ads. Token usage is minimized through DeepSeek prefix caching, compact lesson context, a short history window, and a shared cache for deterministic chip answers.

## User Stories

1. As a learner on any grammar, vocabulary, kanji, or reading lesson, I want to ask the AI a question about that lesson, so that I can clear up confusion without leaving the app.
2. As a learner, I want the AI to already know which lesson I am on, so that I do not have to paste or explain the material.
3. As a learner, I want a "Ringkas" (summarize) chip, so that I can get a concise recap of the lesson.
4. As a learner, I want a "Jelaskan grammar" chip, so that I can get a clearer explanation of the lesson's patterns.
5. As a learner, I want a "Beri contoh kalimat" chip, so that I can see more example sentences.
6. As a learner, I want a "Kuis aku (5 soal)" chip, so that I can test myself with fresh questions each time.
7. As a learner, I want a "Koreksi kalimatku" chip where I write a Japanese sentence and the AI corrects it and explains the mistakes, so that I can practice producing Japanese.
8. As a learner, I want answers to appear progressively while they are generated, so that I am not staring at a spinner.
9. As a learner, I want Japanese examples to always show kanji/kana, romaji, and meaning, so that I can read them at my level.
10. As a learner using the app in Indonesian, I want explanations in Indonesian; as a learner using English, I want them in English.
11. As a learner, I want my conversation for a lesson to still be there when I reopen the lesson, on any of my devices.
12. As a learner, I want to delete the conversation for a lesson.
13. As a learner, I want reopening a summary I already received to be free.
14. As a free user, I want 3 free AI messages per day across the whole app.
15. As a free Android user who has used the free messages, I want the option to watch a short ad to send another message, so that I can keep studying without paying.
16. As a free user, I want to be asked before an ad plays and to be able to cancel, so that ads never surprise me.
17. As a free user who watched an ad but whose answer failed, I do not want to watch another ad to retry.
18. As a free user, I do not want a message to count against my quota if the answer failed.
19. As a free iOS user, or when ads are unavailable, I want to be told the daily limit is reached and be offered Lifetime access.
20. As a Lifetime user, I want 15 AI messages per day without ads, and the option to watch an ad for 5 more.
21. As a Lifetime user who reached 100 messages, I want a clear message that the limit resets tomorrow.
22. As a user without internet, I want the AI button disabled with an explanation, rather than a failing request.
23. As a user considering Lifetime, I want the premium page to state the daily AI messages and the optional ad.
24. As a user, I want the privacy policy to state that my questions are processed by DeepSeek.
25. As an admin, I want to change the free quota, ad cap, Lifetime limit, rate limit, and turn the feature off without a deploy, so that I can control cost.
26. As an admin, I want to configure the rewarded ad unit remotely like the existing placements.
27. As an admin, I want to clear cached chip answers when a lesson's content changes.
28. As an admin, I want to be alerted when an account's usage is abnormal.
29. As the product owner, I want analytics on messages, quota exhaustion, ad offers/views/declines, errors, and paywall views, so that I can tune quotas and measure Lifetime conversion.
30. As the product owner, I want users to be unable to turn the AI into a general chatbot or to override its instructions, so that cost and brand risk stay low.

## Implementation Decisions

### Scope and UX
- An "Ask AI" entry point on all four lesson types: grammar, vocabulary, kanji, reading.
- v1 features: quick-action chips ("Ringkas", "Jelaskan grammar", "Beri contoh kalimat", "Kuis aku (5 soal)", "Koreksi kalimatku"), free-text questions, sentence correction, and streamed answers.
- No remaining-quota indicator is shown.
- Answers follow the app locale (`id` or `en`). Japanese examples are always written as kanji/kana + romaji + meaning.
- Answers render as restricted markdown: no raw HTML, no images, links are not tappable.
- Offline: the entry point is disabled with "Butuh koneksi internet". History is not viewable offline in v1.
- The premium page lists "Tanya AI 15 pesan/hari — Bisa ditambah dengan iklan opsional" as a Lifetime benefit.

### Quota and ads
- Quota is counted server-side per account ID per day and resets at 00:00 Asia/Jakarta (WIB). It is global across all lessons.
- Free users: 3 messages/day. Every chip and free-text message counts as one message, including "Ringkas". Reopening an earlier answer from history does not count.
- Free Android users beyond the free quota: before sending, show an opt-in bottom sheet ("Tonton iklan singkat untuk mengirim pertanyaan ini" with [Tonton] / [Batal]). One completed rewarded ad grants one message credit. At most 20 ad-granted messages per day.
- If the rewarded ad fails to load, the message is still sent (fail-open), but it counts toward the 20/day ad cap.
- Free users on iOS, or when the `aiChatRewarded` placement is disabled: once the free quota is used, chat is locked until the next reset and a Lifetime upsell is shown.
- Lifetime users (`isPremium` / `isLifetime` / offline premium): `premiumPerDay` messages/day (default 15). Then the server answers 429 `ai_premium_ad_offer` with `messagesPerAd`; on Android the app offers an opt-in sheet ("Tonton iklan singkat untuk menambah 5 pesan"). A watched ad (`POST /v1/ai/ad-credits` with `watched`; `unavailable` is refused with 422 `ad_not_watched`) adds `lifetimeAdBonus` messages (default 5), up to `lifetimeAdPerDay` ads (default 4). Without the placement (iOS, ads off, no consent) or past the cap, the answer is `ai_premium_limit`: "Batas harian tercapai, coba lagi besok". The app uses `optInAdsReadyProvider` for this ad, so Lifetime accounts still never get banners or interstitials. On the first start of this version, a saved `premiumPerDay` of 100 (the old default) becomes 15 once; any other saved value is kept.
- All users: a hidden rate limit of 6 messages/minute. Admin alert when an account's daily usage is abnormal.
- Quota and ad credits are consumed only when an answer completes successfully. If DeepSeek errors or times out mid-stream, nothing is consumed. Unused ad credits remain valid until the end of the day.
- v1 trusts the client's "ad watched" claim, bounded by the server-side daily cap. AdMob Server-Side Verification is a later hardening step.

### Architecture
- New authenticated Go route `POST /v1/ai/chat` (behind `requireAuth`) streaming the answer via SSE. A companion endpoint deletes a lesson's conversation.
- The client sends only `lessonType`, `lessonID`, an optional `chip`, the `message` (max 500 characters), and the locale. It never sends history, system text, or lesson content.
- Model: `deepseek-chat`, `max_tokens` 800. `DEEPSEEK_API_KEY` lives in `backend/.env`; the app never holds it.
- New MariaDB tables for: conversations and messages (90-day retention, auto-deleted), daily usage and ad credits per account, the chip answer cache, and compact lesson context.
- Admin:
  - A new `aiChatRewarded` placement in `ads.json` and `/admin/ads`, following the banner/native pattern (enable flag and unit ID served by `/v1/ads-config`).
  - A new `/admin/ai` page holding: free quota (3), ad cap (20), Lifetime limit (100), rate limit (6/min), a kill switch, and clearing the chip cache.
- Android: integrate the rewarded ad format in the existing ads controller. Debug builds use Google's test rewarded unit.

### Prompt injection defense
1. The system prompt exists only on the server and contains no secrets.
2. Lesson context and conversation history are assembled by the server from its own data. Clients cannot inject context or forge assistant turns.
3. User input is length-limited, wrapped in delimiters, and the system prompt instructs the model to treat it as data, not instructions.
4. No tool or function calling: a successful injection can at worst make the model say something off-topic; it cannot act or reach data.
5. Output is rendered as restricted markdown (no HTML, images, or clickable links).
6. The model declines questions unrelated to learning Japanese.
7. Requests and responses are logged for audit.

### Token efficiency
- No vector RAG in v1: a single lesson is about 1–4k tokens and fits whole. (DeepSeek also offers no embeddings API.)
- Prompts are assembled in a fixed order (system prompt → lesson context → history → new message) so that users on the same lesson hit DeepSeek's automatic prefix cache.
- Lesson context uses a compact text form generated once per lesson. It drops `image`, `id`, `order`, quiz questions, and JSON keys, keeping patterns, meanings, explanations, and examples.
- Only the last 4 turns of history are included.
- Answers to "Ringkas", "Jelaskan grammar", and "Beri contoh kalimat" with no extra user text are cached per (lesson, chip, locale) and shared across all users. Cache hits still consume quota and are shown with the same streaming animation. "Kuis aku" and "Koreksi kalimatku" are never cached.

### Analytics
- Firebase events: `ai_chat_sent` (chip, lesson type, lesson ID), `ai_quota_exhausted`, `ai_ad_offered`, `ai_ad_watched`, `ai_ad_declined`, `ai_error`, `ai_paywall_view`.

## Testing Decisions

- Good tests assert observable behavior at the HTTP and UI boundaries: which requests are allowed, what is streamed, what quota state results. They do not assert internal struct names or prompt wording.
- Go API handler tests, using HTTP test requests with a fake DeepSeek client, cover:
  - free quota of 3, then a required ad credit
  - the ad cap of 20
  - the Lifetime limit of 100
  - the per-minute rate limit
  - WIB day rollover
  - no consumption on upstream failure
  - ad credit surviving a failed answer
  - chip cache hits and misses, including no caching for quiz and correction chips
  - client-supplied history or system text being ignored
  - the 500-character limit
  - the kill switch
- Prompt assembly tests assert ordering (stable prefix) and that user text is placed inside delimiters, separately from server-owned content.
- Flutter widget tests with overridden providers cover:
  - the opt-in ad sheet appearing after the free quota
  - the cancel path
  - the iOS / ads-disabled path showing the Lifetime upsell
  - the Lifetime limit message
  - the offline disabled state
  - restricted markdown rendering (no HTML, no clickable links)
- Prior art: the existing Go handler tests and Flutter feature widget tests.

## Out of Scope

- AdMob on iOS.
- AdMob Server-Side Verification (planned hardening).
- Vector RAG or semantic caching of free-text questions.
- Cross-lesson or general Japanese chat outside a lesson.
- Saving AI answers as notes or flashcards, TTS audio, and copy buttons on examples.
- Viewing chat history offline.
- A remaining-quota indicator.
- A separate moderation/classifier model.

## Further Notes

- The quota day is fixed to WIB, while gamification stores a per-account study timezone. Using WIB keeps quota resets uniform and prevents gaming the quota by changing timezone. For English users outside Indonesia, the reset time will not be local midnight.
- `AndroidManifest.xml` still uses Google's test AdMob app ID. A real app ID is required before rewarded ads can earn revenue in production.
- The privacy policy and Play Store data-safety form must be updated to disclose that questions are sent to DeepSeek.
- DeepSeek pricing and caching behavior should be re-checked against the current DeepSeek documentation at implementation time.
