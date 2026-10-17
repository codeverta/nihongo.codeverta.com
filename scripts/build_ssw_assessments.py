"""Build original SSW question resources from authored curricula and Japanese cases."""
import json
from pathlib import Path
ROOT = Path(__file__).resolve().parents[1]

def load(name):
    return json.loads((ROOT / f'assets/content/ssw/{name}.json').read_text(encoding='utf-8'))

def bank(course, sector):
    result = []
    for module in course['modules']:
        for lesson in module['lessons']:
            for i, q in enumerate(lesson['checks']):
                result.append(dict(id=f'{lesson["id"]}-{i+1}', moduleId=module['id'], group=module['title'], lessonId=lesson['id'], prompt=q['prompt'], options=q['options'], answer=q['answer'], explanation=q['explanation'], scenario=lesson['scenario']))
    return result

kaigo = load('kaigo')
cleaning = load('building-cleaning')
banks = {'nursing-care':bank(kaigo,'nursing-care'), 'building-cleaning':bank(cleaning,'building-cleaning')}
japanese = [
('利用者（りようしゃ）とは、だれですか。','介護（かいご）のサービスを使（つか）う人（ひと）|建物（たてもの）を修理（しゅうり）する人|荷物（にもつ）を売（う）る人|試験（しけん）を作（つく）る人','利用者 adalah orang yang menggunakan layanan; konteksnya penerima layanan perawatan.'),
('「車（くるま）いす」の読（よ）み方（かた）はどれですか。','くるまいす|くるまいし|しゃいす|くるまいず','車いす dibaca くるまいす, kursi roda.'),
('「食後（しょくご）」は、いつですか。','ご飯（はん）を食（た）べた後（あと）|ご飯を食べる前（まえ）|寝（ね）ている間（あいだ）|起（お）きる前','食後 = setelah makan; 食前 = sebelum makan.'),
('「排泄（はいせつ）」に関係（かんけい）があるのはどれですか。','トイレ|散歩（さんぽ）|買（か）い物（もの）|洗濯（せんたく）','排泄 berkaitan dengan eliminasi, misalnya buang air di toilet.'),
('「むせる」とは、どのような様子（ようす）ですか。','飲食中（いんしょくちゅう）などに、せきが出（で）る|気持（きも）ちよく眠（ねむ）る|服（ふく）を着替（きが）える|ゆっくり歩（ある）く','むせる adalah tersedak/batuk saat menelan; perlu menghentikan bantuan makan dan menilai sesuai prosedur.'),
('ベッドから車（くるま）いすに移（うつ）る前（まえ）に、何（なん）と言（い）いますか。','これから車いすに移りましょう。|もう食事（しょくじ）は終（お）わりました。|窓（まど）を閉（し）めました。|今日（きょう）は休（やす）みです。','Sampaikan apa yang akan dilakukan sebelum transfer agar pengguna memahami dan dapat berpartisipasi.'),
('利用者（りようしゃ）が「寒（さむ）いです」と言（い）いました。適切（てきせつ）な返事（へんじ）はどれですか。','毛布（もうふ）をお持（も）ちしましょうか。|もっと冷（つめ）たい水（みず）を飲（の）んでください。|寒くありません。|何（なに）も言わないでください。','Tawarkan selimut dan periksa kenyamanan; jangan menyangkal sensasi pengguna.'),
('利用者（りようしゃ）が「右（みぎ）の足（あし）が痛（いた）いです」と言（い）いました。どこが痛いですか。','右の足|左（ひだり）の手（て）|おなか|頭（あたま）','右の足 = kaki kanan. Lokasi keluhan harus dikonfirmasi dan dilaporkan.'),
('「立（た）つ前（まえ）に、ブレーキを確認（かくにん）してください。」最初（さいしょ）に何（なに）をしますか。','ブレーキを確認する|すぐに立つ|車（くるま）いすを動（うご）かす|靴（くつ）を捨（す）てる','Instruksi 前に menunjukkan pemeriksaan rem sebelum berdiri.'),
('利用者（りようしゃ）に聞（き）こえなかったとき、何（なん）と言（い）いますか。','もう一度（いちど）お話（はな）しいただけますか。|話さないでください。|絶対（ぜったい）に違（ちが）います。|何でもいいです。','Minta pengulangan dengan sopan agar tidak menebak kebutuhan pengguna.'),
('記録（きろく）：「午前（ごぜん）十時（じ）、水（みず）を百（ひゃく）ミリリットル飲（の）んだ。」何（なに）を飲みましたか。','水を100ミリリットル|薬（くすり）を100個（こ）|お茶（ちゃ）を1リットル|牛乳（ぎゅうにゅう）を10ミリリットル','Catatan menyebut pukul 10 pagi minum air 100 ml; satuan dan jenis minuman perlu dibaca tepat.'),
('連絡（れんらく）：「田中（たなか）さんは、今日（きょう）は入浴（にゅうよく）しません。」正（ただ）しいのはどれですか。','田中さんは今日はお風呂（ふろ）に入りません。|田中さんは毎日（まいにち）入浴しません。|全員（ぜんいん）が入浴しません。|田中さんは今（いま）入浴しています。','今日は hanya hari ini; 入浴しません = tidak mandi. Jangan memperluas instruksi ke semua orang atau semua hari.'),
('申し送（もうしおく）り：「昼食後（ちゅうしょくご）、せきが続（つづ）いています。看護師（かんごし）に報告済（ほうこくず）みです。」報告はどうなっていますか。','すでに看護師に報告しました。|まだ誰（だれ）にも報告していません。|明日（あした）報告します。|利用者（りようしゃ）が報告を断（ことわ）りました。','報告済み berarti sudah dilaporkan. Catatan menyebut batuk terus setelah makan siang dan sudah melapor ke perawat.'),
('予定（よてい）：「午後（ごご）二時（じ）に体操（たいそう）。その前（まえ）にトイレを確認（かくにん）。」順番（じゅんばん）はどれですか。','トイレを確認してから、二時に体操をする|体操をしてから、トイレを確認する|午前二時に体操をする|トイレの確認はしない','その前に menunjukkan pemeriksaan kebutuhan toilet sebelum senam pukul 14.00.'),
('利用者（りようしゃ）：「自分（じぶん）で顔（かお）を洗（あら）いたいです。」返事（へんじ）はどれがよいですか。','分（わ）かりました。必要（ひつよう）なときはお手伝（てつだ）いします。|時間（じかん）がないので全部（ぜんぶ）私（わたし）がします。|自分では何（なに）もしないでください。|顔を洗ってはいけません。','Hormati pilihan dan kemampuan pengguna; tawarkan bantuan sesuai kebutuhan dan rencana perawatan.'),
]
for i,(prompt,options,explanation) in enumerate(japanese):
    banks['nursing-care'].append(dict(id=f'kaigo-jp-{i+1:02}',moduleId='japanese-assessment',group=['Kosakata Kaigo','Percakapan Kaigo','Dokumen Kaigo'][i//5],lessonId='kaigo-34',prompt=prompt,options=options.split('|'),answer=0,explanation=explanation,scenario=''))

def select(questions,n,used):
    out = [q for q in questions if q['id'] not in used][:n]
    assert len(out)==n
    used.update(q['id'] for q in out)
    return out

used=set()
k=banks['nursing-care'][:72]
# Categories follow the skill assessment outline, with text cases for practical judgment.
categories = [
 ('Dasar Kaigo',10,[q for q in k if q['lessonId'] in [f'kaigo-{i:02}' for i in [1,2,3,31,32,33,36]]]),
 ('Tubuh dan mental',6,[q for q in k if q['lessonId'] in [f'kaigo-{i:02}' for i in range(4,10)]]),
 ('Komunikasi',4,[q for q in k if q['lessonId'] in ['kaigo-10','kaigo-11','kaigo-12']]),
 ('Dukungan kehidupan',20,[q for q in k if q['lessonId'] in [f'kaigo-{i:02}' for i in range(16,31)]]),
 ('Keputusan praktik · skenario teks',5,k),
]
papers={}
items=[]
for group,n,pool in categories:
    for q in select(pool,n,used):
        item=dict(q, section='vocabulary',group=group)
        if group.startswith('Keputusan'): item['passage']=q['scenario']
        items.append(item)
papers['SSW-KAIGO']=dict(title='Kaigo · Simulasi Keterampilan 1',minutes=60,sections=[dict(id='vocabulary',name='Keterampilan Kaigo · 45 soal',durationSeconds=3600,questionCount=45)],items=items)
papers['SSW-KAIGO-JP']=dict(title='Kaigo · Simulasi Bahasa Jepang 1',minutes=30,sections=[dict(id='vocabulary',name='Bahasa Jepang Kaigo · 15 soal',durationSeconds=1800,questionCount=15)],items=[dict(q,section='vocabulary') for q in banks['nursing-care'][72:]])
# Twenty theory questions and thirty situational practical questions; sections have independent clocks.
pool=banks['building-cleaning']
theory_lessons=[1,2,3,4,5,6,7,8,9,10,11,12,13,14,15,16,19,22,23,24]
theory=[next(q for q in pool if q['id']==f'cleaning-{n:02}-1') for n in theory_lessons]
used={q['id'] for q in theory}
practical=select(pool,30,used)
papers['SSW-CLEANING']=dict(title='Pembersihan Gedung · Simulasi 1',minutes=50,sections=[dict(id='vocabulary',name='Teori · 学科',durationSeconds=1200,questionCount=20),dict(id='grammar',name='Penilaian praktik · 実技',durationSeconds=1800,questionCount=30)],items=[dict(q,section='vocabulary') for q in theory]+[dict(q,section='grammar',passage=q['scenario']) for q in practical])

# Randomize the stored option positions deterministically; server re-randomizes simulations.
for questions in banks.values():
    for i,q in enumerate(questions):
        shift=i%len(q['options'])
        q['options']=q['options'][shift:]+q['options'][:shift]
        q['answer']=(q['answer']-shift)%len(q['options'])

if __name__ == '__main__':
    out=ROOT/'backend/internal/store/ssw_resources'
    out.mkdir(exist_ok=True)
    (out/'assessments.json').write_text(json.dumps(dict(banks=banks,papers=papers),ensure_ascii=False,indent=2)+'\n',encoding='utf-8')
    print({k:len(v) for k,v in banks.items()}, {k:len(v['items']) for k,v in papers.items()})
