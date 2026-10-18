# Japan job sources: research and ingestion decisions

Checked on 18 October 2026. This document separates a directory of useful career services from a feed that can actually supply public vacancies. A directory entry is not a job offer and does not establish eligibility, visa sponsorship, language requirements, or salary for any individual job.

## Sources suitable for bounded ingestion

### Lever public Postings API

The [official Lever Postings API documentation](https://github.com/lever/postings-api) describes the public JSON API for building job sites and explicitly notes that published jobs may be scraped by third parties. The private Lever Data API is a different product and is not used. Applications stay on the employer's hosted application page; the app does not collect resumes or submit candidates through an API.

Verified public endpoints:

| Employer | Endpoint | Observed public posts | Japan-location posts observed |
| --- | --- | ---: | ---: |
| Mujin | `https://api.lever.co/v0/postings/mujininc?mode=json` | 89 | 85 |
| Woven by Toyota | `https://api.lever.co/v0/postings/woven-by-toyota?mode=json` | 147 | 54 |

Counts are a research snapshot and will change. Location must be checked from `country`, `categories.location`, and `categories.allLocations`, not merely a mention of Japan somewhere in the description. Global jobs must not appear in the Japan catalogue just because the employer has a Japanese office. A multi-location job needs to identify Japan as an actual work location.

The list response includes the full public posting: stable `id`, `text`, `categories` (location, commitment, department, team, all locations), `country`, `createdAt`, `descriptionPlain`, `openingPlain`, `additionalPlain`, `lists`, `hostedUrl`, `applyUrl`, and workplace type. Salary fields are optional. Absence of pay, visa, or language information must remain unknown; employer-level recruiting claims must not become individual-job guarantees. Lists may contain HTML and must be converted to plain text before display.

One board list request can fetch all currently published posts without visiting every detail page. This makes these two feeds particularly well suited to the user's weekly-per-site limit. Honor pagination if explicitly requested by the adapter, cap response bytes and records, and do not silently archive a partially retrieved board as if it were complete.

The `smartnews` and `moneytree` Lever identifiers returned HTTP 404 during research and are not enabled sources.

### SmartRecruiters public Accor hospitality postings

Verified endpoint:

`https://api.smartrecruiters.com/v1/companies/AccorHotel/postings?limit=100&country=jp&destination=PUBLIC`

It returned 22 public Japan postings, including Osaka hotel and sales roles. The [official list endpoint reference](https://developers.smartrecruiters.com/reference/v1listpostings) documents `country`, `destination=PUBLIC`, `limit` (maximum 100), and `offset`. The [Posting API guide](https://developers.smartrecruiters.com/docs/posting-api) describes exporting public job information to custom career sites. This endpoint was readable without credentials; internal destinations and authenticated APIs are excluded.

Important limitation: the list includes title, location, employment type, experience level, department, industry, custom fields, release date, and an API detail reference, but does **not** contain full qualifications or descriptions. Fetching every detail would require additional per-job requests. A list-only adapter can provide a source link and explicitly show qualifications/salary as unpublished; it must not fabricate those fields. The canonical hosted page pattern is `https://jobs.smartrecruiters.com/AccorHotel/{id}`. Recheck completeness using `totalFound` before treating missing records as closed. More than 100 results requires a deliberately budgeted pagination policy.

## Named boards: directory only

`backend/internal/jobs/resources/sources.json` contains 53 source-directory records with verified official domains or official search results. No automatic request is scheduled for these records. `mode=directory` and `status=directory_only` are deliberate, not a failed ingestion state.

| Source | Evidence or limitation | Decision |
| --- | --- | --- |
| TokyoDev | [robots.txt](https://www.tokyodev.com/robots.txt) allows ordinary public paths but disallows `/c/` and frame paths. A robots allow rule alone does not grant a republication licence. No documented jobs feed/partnership was verified. | Directory link |
| Japan Dev | [robots.txt](https://japan-dev.com/robots.txt) has narrow exclusions. No documented jobs syndication feed or republication agreement was verified. Blog feeds are not job feeds. | Directory link |
| Jobs in Japan | [robots.txt](https://jobsinjapan.com/robots.txt) explicitly disallows `/api/`, candidate, employer, and admin paths. `/feed` is a blog feed, not confirmed job data. | Directory link; do not probe private API paths |
| GaijinPot Jobs | Automated request to `https://jobs.gaijinpot.com/robots.txt` returned HTTP 403. | Directory link; no bypass |
| YOLO JAPAN | The `/en/recruit/` route redirects to login. | Directory link; no account-only collection |
| WeXpats | Old `/en/job/as/jp/` URL returned 404. [Current official site](https://we-xpats.com/ja/) describes recruitment support. [Terms](https://we-xpats.com/ja/legal/) require review before adding any integration. | Current official directory link |
| Indeed / LinkedIn / major recruiting boards | No approved free syndication agreement configured. Indeed research request returned HTTP 403. | Directory links |
| Hello Work | [Official public portal](https://www.hellowork.mhlw.go.jp/) is an employment service. API/data partnership and access conditions have not been arranged. | Government service directory |

Research used one-off reads of public landing pages/robots and official documentation. No authentication, CAPTCHA bypass, browser fingerprint spoofing, rotating proxies, internal endpoints, or bulk detail-page harvesting were used.

## SSW and other industries

These remain discoverable in the source directory even though no verified public SSW job feed was found:

- [TOKUTEI JOB MATCHING](https://tokutei-job-matching.co.jp/): matching service; no approved feed configured.
- [CSI Job Matching](https://tokuteiginouvisa.com/job_matching.html): SSW, training and professional matching; no approved feed configured.
- [Nakayoshi Matching](https://form.nakayoshigc.com/matching-job): the official page explicitly requires profile/certificate verification before access to its jobs group. Private group content is excluded.
- [Mandiri Tokutei Ginou](https://mandiritokuteiginou.com/id): training and matching service, not a confirmed public inventory of live offers.
- [TSK Keihinco](https://tsk.keihinco.com/): information gateway linking to Maha Job, not itself an approved feed.
- [NIPPON Shigoto](https://nipponshigoto.com/): foreign-worker job portal; no approved feed configured.
- [SSW Official](https://www.ssw.go.jp/id/): government guidance, examinations and recruitment events; do not present informational pages as vacancies.
- [Tokyo SSW Job Matching](https://tokyo-tokuteigino.metro.tokyo.lg.jp/): local-government matching programme; no approved feed configured.

Industry-specific directories include nursing/care, hotel, food service, factories, construction, agriculture, logistics, education and creative careers. A nursing licence job is distinct from SSW caregiving. A hotel or restaurant vacancy is not automatically an SSW vacancy. Immigration or recruiting eligibility is recorded only when the actual posting explicitly states it.

The pasted directory named some regional employment centres and Gourmet Career without URLs. Unverified regional centre URLs and an unverified Gourmet Career endpoint were omitted rather than guessed. The official Cookbiz domain redirects from `cookbiz.jp` to `cookbiz.co.jp`; the current destination is stored.

## Collection policy

The user requested one weekly refresh per site, optionally twice weekly. Runtime scheduling belongs to the backend and persisted MariaDB state, not to the mobile screen, client refresh, or a desktop automation. The app reads the saved catalogue. Source URLs, original posting URLs, last successful fetch time, known metadata and unknown fields should remain visible. Temporary errors must retain the last successful data; explicit complete snapshots can mark removed postings closed. Conditional HTTP requests and bounded timeouts reduce external load. Respect server throttling, Retry-After, and failed-source cooldowns; do not immediately loop on failures.

Adding a new adapter requires checking an actual public endpoint and its intended use. A new directory record does not enable crawling. Recheck source policies when enabling a feed and disable collection if the provider blocks access or withdraws it.
