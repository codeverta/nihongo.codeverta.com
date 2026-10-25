# Onboarding preview and artwork

Open **Profil → Panduan aplikasi**, available to signed-in users and guests.
The route `/onboarding?preview=1` explicitly reopens the guide after first-run
completion. Close or finish returns to the previous screen (Home for direct
entry). Preview does not reset `is_first_time`, sign out, or modify progress.
On a fresh install, onboarding remains the normal first-run flow.

The three generated illustrations are bundled locally, with no remote request:

- `assets/images/onboarding/journey.webp`
- `assets/images/onboarding/exams.webp`
- `assets/images/onboarding/anywhere.webp`

Created with the built-in image generation tool. Originals are retained under
Codex generated_images; project copies are optimized to 960px WebP, quality 84.
The combined asset size is approximately 393 KiB.

## Generation prompts

### 1

Create a polished square onboarding illustration for Nihongo Master, a Japanese learning app. Illustration-story, premium editorial 3D paper and soft clay diorama, warm ivory background, indigo blue and vermilion red accents, subtle paper grain, soft daylight. Scene: an open Japanese study notebook and pencil in the foreground, a winding path leading to a tiny Kyoto pagoda and Mount Fuji, cherry blossom branches framing the top, a small red rising sun. Inviting, calm, sophisticated, generous breathing room, objects entirely inside frame. No text, no letters, no numbers, no logos, no watermark. One complete illustration, no panels. Intended for a mobile onboarding hero.

### 2

Create a polished square onboarding illustration for Nihongo Master, a Japanese learning app. Illustration-story, premium editorial 3D paper and soft clay diorama, warm ivory background, indigo blue and vermilion red accents, subtle paper grain, soft daylight. Scene: a beautiful organized study desk with a cream exam sheet containing only abstract blank lines and circles, indigo notebook, red pencil, a small elegant gold achievement medallion and a calendar made of simple blank geometric cells, little sakura petals. Convey confident exam preparation. Inviting, calm, sophisticated, generous breathing room, objects entirely inside frame. No text, no letters, no numbers, no logos, no watermark. One complete illustration, no panels. Intended for a mobile onboarding hero, consistent art direction with Japanese travel study diorama.

### 3

Create a polished square onboarding illustration for Nihongo Master, a Japanese learning app. Illustration-story, premium editorial 3D paper and soft clay diorama, warm ivory background, indigo blue and vermilion red accents, subtle paper grain, soft daylight. Scene: a cozy Japanese train window with Mount Fuji and greenery outside, a tiny indigo bag with an open book and headphones on the seat, a phone showing only simple blank flashcard shapes, warm sunrise. Convey learning anywhere and a daily study habit, offline on the train. Inviting, calm, sophisticated, generous breathing room, objects entirely inside frame. No text, no letters, no numbers, no logos, no watermark. One complete illustration, no panels. Intended for a mobile onboarding hero, consistent with a Japanese study diorama series.

\n