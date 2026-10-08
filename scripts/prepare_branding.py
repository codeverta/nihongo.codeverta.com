"""Export launcher/native splash assets. Requires Pillow.

python scripts/prepare_branding.py
python scripts/prepare_branding.py --backend https://api.example.com/v1
Remote branding is exported for the next build; runtime settings are independent.
"""
import argparse
import io
import json
import shutil
from pathlib import Path
from urllib.request import urlopen
from PIL import Image

ROOT = Path(__file__).resolve().parents[1]


def load_url(url):
    with urlopen(url, timeout=15) as response:
        data = response.read(10 * 1024 * 1024 + 1)
    if len(data) > 10 * 1024 * 1024:
        raise ValueError('Branding image exceeds 10 MB')
    image = Image.open(io.BytesIO(data))
    image.load()
    return image


def export_icon(image):
    image = image.convert('RGBA')
    side = max(image.size)
    square = Image.new('RGBA', (side, side), '#FAFAFA')
    square.alpha_composite(image, ((side-image.width)//2, (side-image.height)//2))
    for density, size in [('mdpi',48), ('hdpi',72), ('xhdpi',96), ('xxhdpi',144), ('xxxhdpi',192)]:
        folder = ROOT / f'android/app/src/main/res/mipmap-{density}'
        folder.mkdir(parents=True, exist_ok=True)
        square.resize((size,size), Image.Resampling.LANCZOS).save(folder / 'ic_launcher.png')
        # Android adaptive mask needs extra inset beyond legacy artwork.
        foreground = Image.new('RGBA', (size*3, size*3), '#FAFAFA')
        mark = square.resize((size*2,size*2), Image.Resampling.LANCZOS)
        foreground.alpha_composite(mark, (size//2,size//2))
        foreground.save(folder / 'ic_launcher_foreground.png')
    folder = ROOT / 'android/app/src/main/res/drawable-nodpi'
    folder.mkdir(parents=True, exist_ok=True)
    square.resize((288,288), Image.Resampling.LANCZOS).save(folder / 'brand_splash.png')
    shutil.copyfile(folder / 'brand_splash.png', ROOT / 'assets/images/app-icon-small.png')
    for scale in [1,2,3]:
        name = 'LaunchImage.png' if scale == 1 else f'LaunchImage@{scale}x.png'
        square.convert('RGB').resize((168*scale,168*scale), Image.Resampling.LANCZOS).save(
            ROOT/'ios/Runner/Assets.xcassets/LaunchImage.imageset'/name)
    square.resize((512,512), Image.Resampling.LANCZOS).save(ROOT / 'output/play-store/app-icon-512.png')
    for platform in ['ios/Runner', 'macos/Runner']:
        asset = ROOT / platform / 'Assets.xcassets/AppIcon.appiconset'
        for item in json.loads((asset / 'Contents.json').read_text())['images']:
            size = round(float(item['size'].split('x')[0]) * float(item['scale'].rstrip('x')))
            square.convert('RGB').resize((size,size), Image.Resampling.LANCZOS).save(asset/item['filename'])
    for size in [192,512]:
        for prefix in ['Icon-', 'Icon-maskable-']:
            square.convert('RGB').resize((size,size),Image.Resampling.LANCZOS).save(ROOT/f'web/icons/{prefix}{size}.png')
    square.resize((32,32),Image.Resampling.LANCZOS).save(ROOT/'web/favicon.png')


def main():
    parser = argparse.ArgumentParser()
    parser.add_argument('--backend', help='Public API base URL, ending /v1')
    args = parser.parse_args()
    icon = Image.open(ROOT / 'assets/images/app-icon-default.png')
    if args.backend:
        with urlopen(args.backend.rstrip('/')+'/branding', timeout=15) as response:
            branding = json.load(response)['data']
        if branding['iconUrl']:
            icon = load_url(branding['iconUrl'])
    export_icon(icon)
    print('Exported Android, iOS, macOS, web and Play Store icons.')


if __name__ == '__main__':
    main()
