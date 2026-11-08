#!/usr/bin/env bash
# Upload the learning/onboarding images in backend/assets/images to Tencent COS.
#
# The production backend points image URLs at COS (COS_BUCKET_URL), so every
# change to backend/assets/images must be uploaded, otherwise the app shows
# empty placeholders (this is what made the reading images disappear after the
# PNG -> WebP conversion in commit 177f67e).
#
# Requirements: coscli (https://www.tencentcloud.com/document/product/436/43249)
# configured once with `coscli config init` (SecretId/SecretKey). Credentials
# are never stored in this repository.
#
# Usage:  backend/scripts/upload-cos-images.sh            # upload
#         DRY_RUN=1 backend/scripts/upload-cos-images.sh  # only list files
set -euo pipefail

BUCKET="${COS_BUCKET:-nihongo-1354822798}"
ENDPOINT="${COS_ENDPOINT:-cos.ap-jakarta.myqcloud.com}"
ROOT="$(cd "$(dirname "$0")/.." && pwd)"
SRC="$ROOT/assets/images"

count=$(find "$SRC" -type f -name '*.webp' | wc -l)
echo "Uploading $count WebP images from $SRC to cos://$BUCKET/assets/images/"
if [ "${DRY_RUN:-0}" = "1" ]; then
  find "$SRC" -type f -name '*.webp' | sed "s#^$SRC/#assets/images/#"
  exit 0
fi

coscli cp -r "$SRC/" "cos://$BUCKET/assets/images/" \
  -e "$ENDPOINT" \
  --include '.*\.webp$' \
  --meta "Content-Type:image/webp#Cache-Control:public, max-age=86400"

echo "Done. Check one file:"
echo "  curl -I https://$BUCKET.$ENDPOINT/assets/images/reading/story-01.webp"
