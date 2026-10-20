# Home explore artwork

Generated with the built-in image_gen tool. Final images are bundled Flutter
assets, so the home cards and Manga banner do not depend on an external image
server. The Manga feature remains gated by Lifetime Access.

- `assets/images/explore/manga-reading.png`: original 1536 × 1024 illustration;
  used by the Manga home card and carousel fallback.
- `assets/images/explore/anki-flashcards.png`: original 1536 × 1024 illustration;
  used by the Anki home card and subscriber carousel.

## Manga prompt

Use case: illustration-story. Asset type: wide landscape illustration for a Japanese learning mobile app feature card and home carousel, no UI screenshot. Create a polished editorial anime illustration of a cozy Japanese manga reading nook: an open black-and-white comic book in the foreground, neatly stacked colorful manga volumes, a small reading lamp, a young adult reader in casual modest clothes in the right half, Japanese city rooftops visible through a window, soft lavender and deep indigo palette with warm cream and tiny golden highlights. Composition: landscape 3:2, recognizable subject centered-right and important objects kept inside central 70 percent so cropping into a wide banner works. Left third has calm simple deep indigo space for app text overlay. Rich clean original illustrated artwork, crisp linework, subtle grain, gentle dimensional shading, inviting premium feel. No logos, no lettering, no watermark, no copyrighted characters, no mature or suggestive content.

## Anki prompt

Use case: illustration-story. Asset type: wide landscape illustration for a Japanese learning mobile app feature card, no UI screenshot. Create a polished editorial anime-style illustration about studying Japanese with flashcards: a tidy wooden study desk with a beautiful fan of blank flashcards, a small card organizer with colorful dividers, an open notebook with simple abstract ink marks, pencil, potted plant, a warm tea cup, a young adult student studying in the right half. Soft teal and mint palette with warm cream and coral accents, daylight through a Japanese room window. Landscape 3:2 composition, recognizable flashcard subject centered-right and important objects kept inside central 70 percent for wide cropping. Left third is visually quiet teal for overlay. Original detailed but clean artwork, crisp linework, subtle grain and gentle dimensional shading, inviting learning app aesthetic. No logos, no legible lettering, no watermark, no mature content.

## Carousel behavior

The home dashboard presents five distinct destinations. Accounts with a purchase
offer see Lifetime, Learn, SSW, Jobs, and Manga. Subscribers see Learn, SSW, Jobs,
Anki, and Manga. Missing older API banners use bundled defaults; configured copy
and image URLs are preserved once per destination. Numbered active capsules and
dots provide direct selection without a play/pause button. Automatic rotation
restarts after a manual swipe; inactive tabs, background apps, reduced-motion
preferences, and mouse hover suspend automatic rotation.
