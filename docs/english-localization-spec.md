# English App Experience and Language Detection

## Problem Statement

Nihongo Master currently presents its user-facing experience primarily in Indonesian and is deployed for an Indonesian audience. People who use English or another device language cannot comfortably navigate the app, understand lessons and questions, or evaluate the product before installing it. Translating interface labels alone would leave substantial parts of the experience unusable because learning materials, exercises, public pages, store information, notifications, and legal documents also need an English version.

The app also needs a predictable language choice: new users should get a language that makes sense for their device, while people already using the Indonesian app should not have their experience unexpectedly changed by an update.

## Solution

Provide a complete, human-reviewed English experience alongside Indonesian. On a fresh install, use Indonesian when the device language is Indonesian and English otherwise. Keep existing installations in Indonesian until the user changes the language. Let users choose Indonesian or English in Settings and save that choice on the current device. Preserve Japanese learning text in both experiences.

English is a single neutral `en` locale. Format dates and numbers using the device's regional settings, and show mobile purchase prices supplied by the relevant app store. Translate and review the complete in-scope experience before making English selectable by the public. Roll out on Android and iOS in the countries where the app is already available, initially keeping the backend in Indonesia and monitoring international latency before deciding whether to add regions.

## User Stories

1. As a new user whose device language is Indonesian, I want the app to start in Indonesian, so that it feels familiar without setup.
2. As a new user whose device language is English, I want the app to start in English, so that I can navigate and learn immediately.
3. As a new user whose device uses an unsupported language, I want the app to start in English, so that I still receive a complete supported experience.
4. As an existing Indonesian user, I want an app update to keep my language in Indonesian, so that an update does not unexpectedly change how I use the app.
5. As a user, I want to choose English or Indonesian in Settings, so that I can use the language I understand best.
6. As a user who selected a language manually, I want that choice to survive app restarts and updates, so that I do not need to choose it again.
7. As a user, I want my language preference to remain on my device, so that it does not silently change on my other devices through account synchronization.
8. As a user who changes language, I want the change to apply consistently throughout the app, so that screens do not mix English and Indonesian interface text.
9. As an English-speaking learner, I want onboarding and help text in English, so that I can understand how to start studying.
10. As an English-speaking learner, I want navigation, home, profile, settings, authentication, and account messages in English, so that I can operate the whole app without relying on Indonesian.
11. As a JLPT learner, I want vocabulary, kanji, grammar, reading, lesson instructions, quizzes, explanations, and feedback in English, so that I can study through every available JLPT level.
12. As a learner preparing for the Specified Skilled Worker (SSW) exams, I want available sector material, assessments, questions, explanations, and progress information in English, so that the English experience includes the same active study paths.
13. As a learner taking JLPT, NAT, JFT, or other available simulations, I want exam instructions, questions, answer choices, results, and review explanations in English, so that I can complete a realistic practice session.
14. As a Japanese learner, I want Japanese words, kanji, example sentences, and ruby/furigana text to remain Japanese in either app language, so that the material I am learning is not translated away.
15. As a learner, I want English translations to appear alongside Japanese study content where translations are part of the lesson, so that I can understand the meaning without losing the original Japanese.
16. As a dictionary user, I want search controls, filters, meanings, empty states, and source information to be understandable in English, so that I can use the dictionary without encountering Indonesian-only UI.
17. As a user, I want loading, empty, offline, error, permission, and confirmation messages translated, so that the app remains understandable outside the happy path.
18. As a user evaluating premium access, I want plan descriptions, purchase explanations, restore-purchase states, and entitlement messages in English, so that I know what I am buying and what access I have.
19. As a user, I want news, job listings, manga features, notifications, and other user-facing modules localized where the app provides their text, so that English support is consistent beyond the learning screens.
20. As a prospective user, I want the public website and store listing to explain the app in English, so that I can assess it before installing it.
21. As a user, I want user-facing legal and support documents translated into English, so that I can understand the terms and help information presented to me.
22. As a user, I want dates and number formatting to respect my device region, so that the English interface remains natural where I live.
23. As a user considering an in-app purchase, I want the displayed price to come from the Google Play or App Store product for my billing region, so that I see the price I will actually be charged.
24. As a user on Android or iOS, I want the same language choices and equivalent translated content, so that English support does not depend on my platform.
25. As a user in a country where the app is currently available, I want to be able to use the English app there, so that English support is not restricted to Indonesia.
26. As an English user outside Indonesia, I want lessons and app screens to remain responsive during the initial rollout, so that studying is practical while the backend is hosted in Indonesia.
27. As a new account, I want my study-day timezone to initialize from my device timezone, so that streaks and daily progress match my local day.
28. As a user who travels, I want my saved study timezone not to change automatically, so that travel does not unexpectedly shift my streak or daily progress boundary.
29. As a user, I want to be able to change my saved study timezone manually, so that I can correct it when needed.
30. As an existing account, I want the current Asia/Jakarta study timezone to remain in place until I change it, so that the language rollout does not alter my historical progress boundary.
31. As an internal administrator, I want the internal admin tools to remain in Indonesian, so that this user-facing localization effort does not unexpectedly change internal workflows.
32. As an English-speaking user, I want the English option to appear only after all agreed user-facing content has been translated and reviewed, so that I do not encounter incomplete or misleading language support.
33. As an internal reviewer, I want English translations to be available for internal quality assurance before public release, so that the complete experience can be checked before users see it.

## Implementation Decisions

- Support two app languages: Indonesian (`id`) and one neutral English locale (`en`). Do not add regional English variants in this release.
- On fresh installs, select Indonesian for an Indonesian device locale; select English for English and unsupported device locales.
- Preserve Indonesian for installations that predate the English rollout until the user makes a manual selection. The migration must distinguish an existing installation from a fresh install using reliable existing installation state; if that distinction is ambiguous, preserve Indonesian rather than changing an existing user's language unexpectedly.
- A manual language selection overrides device detection and is stored locally on the device. It is not synchronized to the user's account.
- Translate the complete agreed user-facing scope: application UI and onboarding; available JLPT and SSW learning materials and assessments; vocabulary and dictionary interfaces; premium and support flows; public website and store listing; notifications; and user-facing legal documents. Internal admin tools remain Indonesian.
- Keep Japanese learning text, including source vocabulary, kanji, example sentences, and ruby/furigana, intact in both locales. Translate explanatory and interface content around it.
- Localize both app-owned strings and backend-provided content. API requests for localized content must carry the selected app language, and backend responses must provide the matching translation without changing Japanese source fields. Cached content must not be shown in the wrong language after the user switches locales.
- English is hidden from public users until all agreed user-facing content is translated and has passed human review. Internal QA may preview it before that release gate is met.
- Ship the experience on Android and iOS in all countries where the app is currently available. Do not change app-store territory availability as part of this work.
- Keep the backend hosted in Indonesia for the initial international rollout. Monitor latency and reliability for international users, then use that evidence to decide whether to add backend regions in a later change.
- In-app purchase prices come from the active Google Play or App Store product for the user's billing region. Do not present a hard-coded IDR amount as the localized store price. Public web pricing must not imply that an IDR-only amount is the price in every market.
- Format dates and numbers with the device's regional settings while keeping the app language preference independent of the device region.
- For gamification, initialize a new account's timezone from the device timezone and persist it with the account. Do not change it automatically when the device timezone changes during travel; allow a manual change. Existing accounts retain Asia/Jakarta until the user changes it.

## Testing Decisions

- Good tests assert what a user observes at supported boundaries: the selected language, translated text and content, persisted preference, and the language response returned by the API. They should not assert the names of internal localization classes, storage keys, or implementation-specific widget structure.
- Use the highest practical Flutter seam at the app root to verify locale selection and representative user-visible behavior. Cover Indonesian device detection, English device detection, unsupported-language fallback, manual selection surviving restart, and a pre-existing installation retaining Indonesian. Include representative learning and purchase flows to verify that Japanese source text remains intact and the price shown follows the localized store product.
- Prior art is the repository's Flutter root and feature widget tests, which build app or feature widgets and override providers for deterministic dependencies. Add the new behavior to that existing style of widget testing rather than creating a separate localization test framework.
- At the Go HTTP API boundary, request localized content in Indonesian and English and assert that each response contains the expected translated fields while preserving Japanese source fields.
- Prior art is the repository's Go API handler tests using HTTP test requests and response assertions. Keep the API contract tests at this handler boundary.
- The user approved these two test seams: Flutter app-root behavior and Go API handler behavior.

## Out of Scope

- Localizing internal admin interfaces.
- Adding languages other than Indonesian and English or adding en-US/en-GB variants.
- Rewriting or translating the Japanese text that is itself the subject of study.
- Changing which countries can install the app through Google Play or the App Store.
- Moving the backend out of Indonesia as part of the initial rollout.
- Changing the account synchronization model for language preference; the choice remains device-local.
- Adding new courses, exam types, or learning features beyond translating the agreed content that already exists.

## Further Notes

- The current repository has no established Flutter localization setup, while many screens and learning datasets contain Indonesian text. This work therefore includes both app-owned UI strings and backend/content localization; replacing only visible labels would not meet the agreed scope.
- Existing purchase APIs expose IDR-specific prices today. The implementation must ensure the mobile purchase UI uses store-provided regional pricing; web pricing behavior also needs to avoid presenting the IDR value as universally applicable.
- No domain glossary or ADR for localization was found during repository review.
- The project issue tracker and its `ready-for-agent` label are not configured in the available workspace. This document is the reviewable local spec; publish it to the project tracker with `ready-for-agent` after running `/setup-matt-pocock-skills` and selecting the project's tracker.
