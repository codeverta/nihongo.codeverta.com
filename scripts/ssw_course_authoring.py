"""Shared schema for original Indonesian SSW lessons, cases and explanations."""
import json
from pathlib import Path

ROOT = Path(__file__).resolve().parents[1]

class Course:
    def __init__(self, sector, prefix, sources):
        self.sector, self.prefix, self.sources = sector, prefix, sources
        self.modules = []

    def module(self, id, title, track, description):
        self.modules.append(dict(id=id, title=title, track=track, description=description, lessons=[]))

    def lesson(self, title, body, steps, case, responses, knowledge, terms):
        """Four distinct checks: concept, case, vocabulary, and workflow priority."""
        n = sum(len(m['lessons']) for m in self.modules)+1
        steps = steps.split('|')
        options = responses.split('|')
        prompt, answers, explanation = knowledge.split('~')
        words = []
        for term in terms.split('|'):
            jp, kana, meaning = term.split('~')
            words.append(dict(japanese=jp, reading=kana, meaning=meaning))
        checks = [
            dict(prompt=prompt, options=answers.split('|'), answer=0, explanation=explanation, kind='knowledge'),
            dict(prompt=f'Kasus: {case} Tindakan yang paling tepat adalah…', options=options, answer=0, explanation=f'{options[0]}. '+explanation, kind='case'),
            dict(prompt=f'Dalam konteks lesson ini, apa arti {words[n%4]["japanese"]}（{words[n%4]["reading"]}）?', options=[w['meaning'] for w in words], answer=n%4, explanation=f'{words[n%4]["japanese"]} dibaca {words[n%4]["reading"]} dan berarti {words[n%4]["meaning"]}. Kenali istilah ini sebelum mengikuti instruksi kerja.', kind='vocabulary'),
            dict(prompt=f'Pada alur kerja «{title}» dalam materi, langkah pertama sebelum tahap berikutnya adalah…', options=steps, answer=0, explanation='Mulai dengan '+steps[0].lower()+'. Setelah itu: '+' → '.join(steps[1:])+'. Urutan ini berlaku pada kasus dan prosedur yang dijelaskan di lesson.', kind='workflow'),
        ]
        assert len(steps)==4 and len(options)==4 and len(words)==4
        for i,q in enumerate(checks):
            assert len(q['options'])==4 and len(set(q['options']))==4
            shift=(n+i)%4
            q['options']=q['options'][shift:]+q['options'][:shift]
            q['answer']=(q['answer']-shift)%4
        application = (f'Bayangkan kondisi kerja berikut: {case} Keputusan yang dianjurkan adalah: {options[0].lower()}. '
            f'Lakukan alur secara teratur: {"; ".join(steps)}. Catat kondisi sebelum dan sesudah tindakan, lokasi atau kelompok yang terdampak, waktu pengamatan, serta siapa yang diberi laporan. '
            'Jika hasil berbeda dari rencana, hentikan bagian yang berisiko dan minta arahan penanggung jawab sebelum mengubah metode. '
            'Angka target, dosis produk, setelan alat, dan batas cuaca mengikuti instruksi lokasi, label, serta jenis tanaman atau hewan; contoh belajar bukan angka universal.')
        self.modules[-1]['lessons'].append(dict(id=f'{self.prefix}-{n:02}',title=title,minutes=25,
            objectives=[f'Menjelaskan prinsip {title.lower()}.','Menentukan urutan kerja dan respons kasus yang sesuai.','Membaca istilah Jepang dan mencatat hasil pengamatan.'],
            sections=[dict(title='Prinsip dan pengetahuan bidang',body=body),dict(title='Penerapan, kasus, dan pemeriksaan hasil',body=application)],
            steps=steps,caution='Ikuti SOP dan batas kewenangan. Mesin bergerak, bahan kimia, obat hewan, pekerjaan di air, serta keadaan darurat memerlukan pelatihan dan petugas berwenang. Jangan menebak dosis atau melakukan tindakan berisiko dari materi tertulis saja.',
            scenario=case,response=options[0],vocabulary=words,
            dialogue=[dict(japanese='指示をもう一度お願いします。',reading='しじをもういちどおねがいします。',meaning='Tolong ulangi instruksinya.'),dict(japanese=words[0]['japanese']+'を確認します。',reading=words[0]['reading']+'をかくにんします。',meaning='Saya akan memeriksa '+words[0]['meaning']+'.')],checks=checks))

    def write(self):
        (ROOT/f'assets/content/ssw/{self.sector}.json').write_text(json.dumps(dict(modules=self.modules,sources=self.sources),ensure_ascii=False,indent=2)+'\n',encoding='utf-8')
        print(self.sector,len(self.modules),'modules',sum(len(m['lessons']) for m in self.modules),'lessons')
