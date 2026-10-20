# Lifetime Access and job applications

The job-detail source attribution remains visible. The duplicate “Buka halaman
asli” action has been removed. The application button sends guest/free accounts
to `/premium?from=/jobs/<id>` and verifies signed-in premium accounts through the
existing `/v1/billing` endpoint before opening the external application URL.
Refunded/revoked access and unavailable verification never open that URL.
Inactive listings remain disabled. After purchase/restore, the premium page can
return to the job detail.

Employer benefits are stored in `job_listings.benefits` and imported from explicit
JobPosting `jobBenefits` or labeled Lever benefits sections. Weekly collection
and source cooldowns are unchanged. Older cached descriptions can display an
explicitly labeled benefits section immediately. Missing data is clearly marked;
housing, visa support, insurance, or bonuses are never invented. The new backend
field is migrated automatically when the updated API/collector starts.

The Lifetime page keeps purchase, pending, restore, verification retry, and
refund-help flows. Store-localized prices take precedence over the backend's
display estimate when a store product is available.

## Generated illustration

Built-in image_gen tool, saved as
`assets/images/explore/lifetime-access.png` (1536 × 1024), bundled with Flutter.

Final prompt:

Use case: illustration-story. Asset type: premium membership hero illustration for a Japanese language learning mobile app, wide landscape 3:2. Original polished editorial anime illustration of a warm aspirational Japanese study and career journey: a beautiful navy-blue notebook and Japanese language books, a fan of flashcards and an open manga book on a tidy desk in foreground, a small tasteful gold crown-shaped bookmark, soft glowing golden path leading into a peaceful Tokyo city skyline and distant Mount Fuji beyond a large window. Deep indigo, midnight navy, warm cream and restrained metallic gold palette. Detailed painterly light and crisp clean linework, elegant and welcoming, believable depth. Keep main recognizable objects and skyline near center so cropping into a wide mobile banner works. No text, no logos, no currency, no UI, no watermark, no copyrighted characters, no adult content.
