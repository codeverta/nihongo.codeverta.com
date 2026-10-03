"""Compose Play Store artwork using unmodified screenshots from the real app.

Requires Pillow. No UI is invented: source screenshots remain in screenshots/.
"""
from pathlib import Path
from PIL import Image, ImageDraw, ImageFont

ROOT = Path(__file__).resolve().parents[1]
OUT = ROOT / 'output/play-store'
INDIGO = '#1D2088'
PAPER = '#FAFAFA'
RED = '#E60012'


def font(size, bold=False):
    return ImageFont.truetype('C:/Windows/Fonts/seguisb.ttf' if bold else 'C:/Windows/Fonts/segoeui.ttf', size)


def main():
    previews = OUT/'previews'
    previews.mkdir(exist_ok=True)
    specs = [
        ('01-beranda', 'Mulai kebiasaan\nbelajar Jepang.', 'Target harian dan jalur belajar dalam satu aplikasi.'),
        ('02-belajar', 'Pilih materi.\nMulai langkahmu.', 'Vocabulary, grammar, kanji, dan reading.'),
        ('03-vocabulary', 'Lihat gambarnya.\nIngat katanya.', 'Kosakata N5 dengan ilustrasi dan bantuan kana.'),
        ('04-grammar', 'Pahami pola.\nBangun kalimat.', 'Grammar N5, penjelasan dan contoh bahasa Indonesia.'),
        ('05-kanji', 'Kenali kanji,\nsatu per satu.', 'Arti, bacaan, petunjuk mengingat dan contoh kata.'),
        ('06-reading', 'Baca cerita.\nBuka dunia baru.', 'Reading bergambar dengan kana dan terjemahan.'),
        ('07-ujian', 'Temukan titik\nawal belajarmu.', 'Diagnostic JLPT N5: 25 soal dalam 30 menit.'),
    ]
    thumbs=[]
    for i,(name,title,subtitle) in enumerate(specs,1):
        canvas=Image.new('RGB',(1080,1920),PAPER)
        d=ImageDraw.Draw(canvas)
        d.ellipse((900,-100,1200,200),fill=RED)
        d.text((72,64),'NIHONGO MASTER',font=font(27,True),fill=INDIGO)
        d.multiline_text((72,145),title,font=font(72,True),fill=INDIGO,spacing=4)
        d.text((72,345),subtitle,font=font(30),fill='#606070')
        source=Image.open(OUT/'screenshots'/f'{name}.png').convert('RGB')
        if source.size!=(1080,1920):
            raise ValueError(f'{name}: expected a 1080x1920 capture')
        # Uniform resize preserves every screen pixel; no cropping or retouching.
        source=source.resize((756,1344),Image.Resampling.LANCZOS)
        d.rounded_rectangle((148,462,932,1834),radius=26,fill=INDIGO)
        canvas.paste(source,(162,476))
        d.text((72,1860),f'{i:02d} / 07',font=font(23),fill='#606070')
        d.text((750,1860),'Belajar setiap hari',font=font(23),fill=INDIGO)
        canvas.save(previews/f'{name}.png',optimize=True)
        thumbs.append(canvas.resize((270,480),Image.Resampling.LANCZOS))
    sheet=Image.new('RGB',(270*4,480*2),'#EDEDF3')
    for i,thumb in enumerate(thumbs):
        sheet.paste(thumb,((i%4)*270,(i//4)*480))
    sheet.save(OUT/'preview-overview.png')
    feature=Image.new('RGB',(1024,500),PAPER)
    d=ImageDraw.Draw(feature)
    d.ellipse((870,-75,1090,145),fill=RED)
    icon=Image.open(OUT/'app-icon-512.png').convert('RGB').resize((94,94),Image.Resampling.LANCZOS)
    feature.paste(icon,(50,42))
    d.text((158,70),'NIHONGO MASTER',font=font(25,True),fill=INDIGO)
    d.multiline_text((56,172),'Langkah kecil.\nDunia baru.',font=font(55,True),fill=INDIGO,spacing=3)
    d.multiline_text((58,340),'Belajar bahasa Jepang\ndengan gambar, cerita, dan latihan.',font=font(23),fill='#606070',spacing=6)
    for name,x,y in [('03-vocabulary',585,60),('06-reading',796,108)]:
        shot=Image.open(OUT/'screenshots'/f'{name}.png').convert('RGB').resize((180,320),Image.Resampling.LANCZOS)
        d.rounded_rectangle((x-7,y-7,x+187,y+327),radius=12,fill=INDIGO)
        feature.paste(shot,(x,y))
    feature.save(OUT/'feature-graphic-1024x500.png',optimize=True)
    print('Created seven 1080x1920 previews and 1024x500 feature graphic.')


if __name__=='__main__':
    main()
