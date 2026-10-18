# Kanji N2

Kurikulum tersedia dari **Belajar → N2 → Kanji** setelah login. Materi memuat
**367 kanji dalam 74 lesson**, satu contoh kosakata untuk tiap kanji, dan dua
soal per kanji: arti kamus serta bacaan kata contoh. Progres N2 disimpan terpisah
dari level lain.

## Cakupan dan batas materi

Karakter dipilih dari entri `jlpt_new == 2` pada snapshot `kanji-data`.
Pengelompokan lesson mengikuti frekuensi sumber sebagai urutan belajar, bukan
kelompok tema atau urutan resmi JLPT. Jumlah karakter yang disebut sumber lain
bisa berbeda. Ini daftar persiapan aplikasi; JLPT tidak menerbitkan daftar
resmi kanji per level setelah revisi 2010 ([FAQ resmi JLPT](https://www.jlpt.jp/e/faq/)).

Kartu menampilkan jumlah goresan dan bacaan on/kun dari KANJIDIC. Contoh kata
dan bacaan mengikuti korpus latihan kosakata yang telah disimpan dalam aplikasi;
gloss kamus dan contoh pada N2 masih berbahasa Inggris sesuai sumbernya.
Petunjuk kartu mengarahkan pelajar mengingat karakter melalui kata contoh,
bukan mengklaim penjelasan asal-usul bentuk kanji. Titik dalam bacaan kun’yomi
menandai okurigana pada format KANJIDIC.

## Sumber

- Data jumlah goresan, bacaan, arti ringkas, dan level `jlpt_new`: [kanji-data](https://github.com/davidluzgouveia/kanji-data), yang mengolah data KANJIDIC. Penetapan level diatribusikan dataset itu kepada [Jonathan Waller JLPT Resources](https://www.tanos.co.uk/jlpt/).
- Lisensi data KANJIDIC: [EDRDG KANJIDIC Project](https://www.edrdg.org/wiki/KANJIDIC_Project.html), CC BY-SA 4.0. Metadata konversi dataset tercatat pada [lisensi kanji-data](licenses/kanji-data-MIT.txt).
- Contoh kata, bacaan, dan gloss: kosakata JLPT yang tersimpan di `backend/data/jlpt/`; sumber upstream dan revisi tercatat dalam [katalog deck](../backend/internal/anki/resources/catalog.json), termasuk [deck JLPT N2](https://github.com/mistval/kotoba/blob/a0f8f984bf5dd8014527c6608d36b479ab622b07/resources/quiz_data/n2.json).
- Tujuh contoh kata tambahan untuk karakter yang tidak muncul di deck lokal diperiksa di [Jisho.org](https://jisho.org/), yang menandai sumber JMdict; lisensi dan syarat atribusi kamus tercantum di [EDRDG](https://www.edrdg.org/edrdg/newlic.html). URL tiap kata ada di snapshot rujukan.
- Korpus kosakata berbahasa Indonesia untuk pembanding belajar: [Daftar kosakata N2 Jepang.org](https://legacy.jepang.org/jlpt/n2/kosa-kata/). Materi aplikasi tidak mengimpor tabel situs itu.

Penyajian level bukan klaim resmi JLPT. Bacaan kartu adalah pilihan kamus,
bukan seluruh kemungkinan bacaan. Gloss Inggris dipertahankan agar maknanya
tidak berubah karena terjemahan otomatis.

## Integrasi dan pembaruan

- Materi tersemat: `backend/data/kanji/n2.json`.
- Snapshot rujukan: `kanji-n2-reference.json`, diambil 18 Oktober 2026. Snapshot
  mencatat URL sumber, SHA-256 dataset penuh yang digunakan, dan rujukan untuk
  tujuh kata pelengkap.
- Regenerator offline: `python scripts/build_kanji_n2.py`. Generator membaca
  snapshot rujukan dan korpus kosakata lokal N1–N5; tidak menghubungi jaringan.
- API katalog: `GET /v1/kanji/course?level=N2`. Endpoint materi dan kuis sama
  dengan level kanji lain dan memerlukan login.
- ID lesson: `kanji-n2-01` sampai `kanji-n2-74`.

| 01 | Kosakata dan kanji N2 01 · 党協総区領 | 党 協 総 区 領 |
| 02 | Kosakata dan kanji N2 02 · 県設改府査 | 県 設 改 府 査 |
| 03 | Kosakata dan kanji N2 03 · 委軍団各島 | 委 軍 団 各 島 |
| 04 | Kosakata dan kanji N2 04 · 革村勢減再 | 革 村 勢 減 再 |
| 05 | Kosakata dan kanji N2 05 · 税営比防補 | 税 営 比 防 補 |
| 06 | Kosakata dan kanji N2 06 · 境導副算輸 | 境 導 副 算 輸 |
| 07 | Kosakata dan kanji N2 07 · 述線農州武 | 述 線 農 州 武 |
| 08 | Kosakata dan kanji N2 08 · 象域額欧担 | 象 域 額 欧 担 |
| 09 | Kosakata dan kanji N2 09 · 準賞辺造被 | 準 賞 辺 造 被 |
| 10 | Kosakata dan kanji N2 10 · 技低復移個 | 技 低 復 移 個 |
| 11 | Kosakata dan kanji N2 11 · 門課脳極含 | 門 課 脳 極 含 |
| 12 | Kosakata dan kanji N2 12 · 蔵量型況針 | 蔵 量 型 況 針 |
| 13 | Kosakata dan kanji N2 13 · 専谷史階管 | 専 谷 史 階 管 |
| 14 | Kosakata dan kanji N2 14 · 兵接細効丸 | 兵 接 細 効 丸 |
| 15 | Kosakata dan kanji N2 15 · 湾録省旧橋 | 湾 録 省 旧 橋 |
| 16 | Kosakata dan kanji N2 16 · 岸周材戸央 | 岸 周 材 戸 央 |
| 17 | Kosakata dan kanji N2 17 · 券編捜竹超 | 券 編 捜 竹 超 |
| 18 | Kosakata dan kanji N2 18 · 並療採森競 | 並 療 採 森 競 |
| 19 | Kosakata dan kanji N2 19 · 介根販歴将 | 介 根 販 歴 将 |
| 20 | Kosakata dan kanji N2 20 · 幅般貿講林 | 幅 般 貿 講 林 |
| 21 | Kosakata dan kanji N2 21 · 装諸劇河航 | 装 諸 劇 河 航 |
| 22 | Kosakata dan kanji N2 22 · 鉄児禁印逆 | 鉄 児 禁 印 逆 |
| 23 | Kosakata dan kanji N2 23 · 換久短油暴 | 換 久 短 油 暴 |
| 24 | Kosakata dan kanji N2 24 · 輪占植清倍 | 輪 占 植 清 倍 |
| 25 | Kosakata dan kanji N2 25 · 均億圧芸署 | 均 億 圧 芸 署 |
| 26 | Kosakata dan kanji N2 26 · 伸停爆陸玉 | 伸 停 爆 陸 玉 |
| 27 | Kosakata dan kanji N2 27 · 波帯延羽固 | 波 帯 延 羽 固 |
| 28 | Kosakata dan kanji N2 28 · 則乱普測豊 | 則 乱 普 測 豊 |
| 29 | Kosakata dan kanji N2 29 · 厚齢囲卒略 | 厚 齢 囲 卒 略 |
| 30 | Kosakata dan kanji N2 30 · 承順岩練軽 | 承 順 岩 練 軽 |
| 31 | Kosakata dan kanji N2 31 · 了庁城患層 | 了 庁 城 患 層 |
| 32 | Kosakata dan kanji N2 32 · 版令角絡損 | 版 令 角 絡 損 |
| 33 | Kosakata dan kanji N2 33 · 募裏仏績築 | 募 裏 仏 績 築 |
| 34 | Kosakata dan kanji N2 34 · 貨混昇池血 | 貨 混 昇 池 血 |
| 35 | Kosakata dan kanji N2 35 · 温季星永著 | 温 季 星 永 著 |
| 36 | Kosakata dan kanji N2 36 · 誌庫刊像香 | 誌 庫 刊 像 香 |
| 37 | Kosakata dan kanji N2 37 · 坂底布寺宇 | 坂 底 布 寺 宇 |
| 38 | Kosakata dan kanji N2 38 · 巨震希触依 | 巨 震 希 触 依 |
| 39 | Kosakata dan kanji N2 39 · 籍汚枚複郵 | 籍 汚 枚 複 郵 |
| 40 | Kosakata dan kanji N2 40 · 仲栄札板骨 | 仲 栄 札 板 骨 |
| 41 | Kosakata dan kanji N2 41 · 傾届巻燃跡 | 傾 届 巻 燃 跡 |
| 42 | Kosakata dan kanji N2 42 · 包駐弱紹雇 | 包 駐 弱 紹 雇 |
| 43 | Kosakata dan kanji N2 43 · 替預焼簡章 | 替 預 焼 簡 章 |
| 44 | Kosakata dan kanji N2 44 · 臓律贈照薄 | 臓 律 贈 照 薄 |
| 45 | Kosakata dan kanji N2 45 · 群秒奥詰双 | 群 秒 奥 詰 双 |
| 46 | Kosakata dan kanji N2 46 · 刺純翌快片 | 刺 純 翌 快 片 |
| 47 | Kosakata dan kanji N2 47 · 敬悩泉皮漁 | 敬 悩 泉 皮 漁 |
| 48 | Kosakata dan kanji N2 48 · 荒貯硬埋柱 | 荒 貯 硬 埋 柱 |
| 49 | Kosakata dan kanji N2 49 · 祭袋筆訓浴 | 祭 袋 筆 訓 浴 |
| 50 | Kosakata dan kanji N2 50 · 童宝封胸砂 | 童 宝 封 胸 砂 |
| 51 | Kosakata dan kanji N2 51 · 塩賢腕兆床 | 塩 賢 腕 兆 床 |
| 52 | Kosakata dan kanji N2 52 · 毛緑尊祝柔 | 毛 緑 尊 祝 柔 |
| 53 | Kosakata dan kanji N2 53 · 殿濃液衣肩 | 殿 濃 液 衣 肩 |
| 54 | Kosakata dan kanji N2 54 · 零幼荷泊黄 | 零 幼 荷 泊 黄 |
| 55 | Kosakata dan kanji N2 55 · 甘臣浅掃雲 | 甘 臣 浅 掃 雲 |
| 56 | Kosakata dan kanji N2 56 · 掘捨軟沈凍 | 掘 捨 軟 沈 凍 |
| 57 | Kosakata dan kanji N2 57 · 乳恋紅郊腰 | 乳 恋 紅 郊 腰 |
| 58 | Kosakata dan kanji N2 58 · 炭踊冊勇械 | 炭 踊 冊 勇 械 |
| 59 | Kosakata dan kanji N2 59 · 菜珍卵湖喫 | 菜 珍 卵 湖 喫 |
| 60 | Kosakata dan kanji N2 60 · 干虫刷湯溶 | 干 虫 刷 湯 溶 |
| 61 | Kosakata dan kanji N2 61 · 鉱涙匹孫鋭 | 鉱 涙 匹 孫 鋭 |
| 62 | Kosakata dan kanji N2 62 · 枝塗軒毒叫 | 枝 塗 軒 毒 叫 |
| 63 | Kosakata dan kanji N2 63 · 拝氷乾棒祈 | 拝 氷 乾 棒 祈 |
| 64 | Kosakata dan kanji N2 64 · 拾粉糸綿汗 | 拾 粉 糸 綿 汗 |
| 65 | Kosakata dan kanji N2 65 · 銅湿瓶咲召 | 銅 湿 瓶 咲 召 |
| 66 | Kosakata dan kanji N2 66 · 缶隻脂蒸肌 | 缶 隻 脂 蒸 肌 |
| 67 | Kosakata dan kanji N2 67 · 耕鈍泥隅灯 | 耕 鈍 泥 隅 灯 |
| 68 | Kosakata dan kanji N2 68 · 辛磨麦姓筒 | 辛 磨 麦 姓 筒 |
| 69 | Kosakata dan kanji N2 69 · 鼻粒詞胃畳 | 鼻 粒 詞 胃 畳 |
| 70 | Kosakata dan kanji N2 70 · 机膚濯塔沸 | 机 膚 濯 塔 沸 |
| 71 | Kosakata dan kanji N2 71 · 灰菓帽枯涼 | 灰 菓 帽 枯 涼 |
| 72 | Kosakata dan kanji N2 72 · 舟貝符憎 | 舟 貝 符 憎 |
| 73 | Kosakata dan kanji N2 73 · 皿肯燥畜 | 皿 肯 燥 畜 |
| 74 | Kosakata dan kanji N2 74 · 挟曇滴伺 | 挟 曇 滴 伺 |
| Lesson | Judul | Kanji |
|---|---|---|
| 01 | Kosakata dan kanji N2 01 · 党協総区領 | 党 協 総 区 領 |
| 02 | Kosakata dan kanji N2 02 · 県設改府査 | 県 設 改 府 査 |
| 03 | Kosakata dan kanji N2 03 · 委軍団各島 | 委 軍 団 各 島 |
| 04 | Kosakata dan kanji N2 04 · 革村勢減再 | 革 村 勢 減 再 |
| 05 | Kosakata dan kanji N2 05 · 税営比防補 | 税 営 比 防 補 |
| 06 | Kosakata dan kanji N2 06 · 境導副算輸 | 境 導 副 算 輸 |
| 07 | Kosakata dan kanji N2 07 · 述線農州武 | 述 線 農 州 武 |
| 08 | Kosakata dan kanji N2 08 · 象域額欧担 | 象 域 額 欧 担 |
| 09 | Kosakata dan kanji N2 09 · 準賞辺造被 | 準 賞 辺 造 被 |
| 10 | Kosakata dan kanji N2 10 · 技低復移個 | 技 低 復 移 個 |
| 11 | Kosakata dan kanji N2 11 · 門課脳極含 | 門 課 脳 極 含 |
| 12 | Kosakata dan kanji N2 12 · 蔵量型況針 | 蔵 量 型 況 針 |
| 13 | Kosakata dan kanji N2 13 · 専谷史階管 | 専 谷 史 階 管 |
| 14 | Kosakata dan kanji N2 14 · 兵接細効丸 | 兵 接 細 効 丸 |
| 15 | Kosakata dan kanji N2 15 · 湾録省旧橋 | 湾 録 省 旧 橋 |
| 16 | Kosakata dan kanji N2 16 · 岸周材戸央 | 岸 周 材 戸 央 |
| 17 | Kosakata dan kanji N2 17 · 券編捜竹超 | 券 編 捜 竹 超 |
| 18 | Kosakata dan kanji N2 18 · 並療採森競 | 並 療 採 森 競 |
| 19 | Kosakata dan kanji N2 19 · 介根販歴将 | 介 根 販 歴 将 |
| 20 | Kosakata dan kanji N2 20 · 幅般貿講林 | 幅 般 貿 講 林 |
| 21 | Kosakata dan kanji N2 21 · 装諸劇河航 | 装 諸 劇 河 航 |
| 22 | Kosakata dan kanji N2 22 · 鉄児禁印逆 | 鉄 児 禁 印 逆 |
| 23 | Kosakata dan kanji N2 23 · 換久短油暴 | 換 久 短 油 暴 |
| 24 | Kosakata dan kanji N2 24 · 輪占植清倍 | 輪 占 植 清 倍 |
| 25 | Kosakata dan kanji N2 25 · 均億圧芸署 | 均 億 圧 芸 署 |
| 26 | Kosakata dan kanji N2 26 · 伸停爆陸玉 | 伸 停 爆 陸 玉 |
| 27 | Kosakata dan kanji N2 27 · 波帯延羽固 | 波 帯 延 羽 固 |
| 28 | Kosakata dan kanji N2 28 · 則乱普測豊 | 則 乱 普 測 豊 |
| 29 | Kosakata dan kanji N2 29 · 厚齢囲卒略 | 厚 齢 囲 卒 略 |
| 30 | Kosakata dan kanji N2 30 · 承順岩練軽 | 承 順 岩 練 軽 |
| 31 | Kosakata dan kanji N2 31 · 了庁城患層 | 了 庁 城 患 層 |
| 32 | Kosakata dan kanji N2 32 · 版令角絡損 | 版 令 角 絡 損 |
| 33 | Kosakata dan kanji N2 33 · 募裏仏績築 | 募 裏 仏 績 築 |
| 34 | Kosakata dan kanji N2 34 · 貨混昇池血 | 貨 混 昇 池 血 |
| 35 | Kosakata dan kanji N2 35 · 温季星永著 | 温 季 星 永 著 |
| 36 | Kosakata dan kanji N2 36 · 誌庫刊像香 | 誌 庫 刊 像 香 |
| 37 | Kosakata dan kanji N2 37 · 坂底布寺宇 | 坂 底 布 寺 宇 |
| 38 | Kosakata dan kanji N2 38 · 巨震希触依 | 巨 震 希 触 依 |
| 39 | Kosakata dan kanji N2 39 · 籍汚枚複郵 | 籍 汚 枚 複 郵 |
| 40 | Kosakata dan kanji N2 40 · 仲栄札板骨 | 仲 栄 札 板 骨 |
| 41 | Kosakata dan kanji N2 41 · 傾届巻燃跡 | 傾 届 巻 燃 跡 |
| 42 | Kosakata dan kanji N2 42 · 包駐弱紹雇 | 包 駐 弱 紹 雇 |
| 43 | Kosakata dan kanji N2 43 · 替預焼簡章 | 替 預 焼 簡 章 |
| 44 | Kosakata dan kanji N2 44 · 臓律贈照薄 | 臓 律 贈 照 薄 |
| 45 | Kosakata dan kanji N2 45 · 群秒奥詰双 | 群 秒 奥 詰 双 |
| 46 | Kosakata dan kanji N2 46 · 刺純翌快片 | 刺 純 翌 快 片 |
| 47 | Kosakata dan kanji N2 47 · 敬悩泉皮漁 | 敬 悩 泉 皮 漁 |
| 48 | Kosakata dan kanji N2 48 · 荒貯硬埋柱 | 荒 貯 硬 埋 柱 |
| 49 | Kosakata dan kanji N2 49 · 祭袋筆訓浴 | 祭 袋 筆 訓 浴 |
| 50 | Kosakata dan kanji N2 50 · 童宝封胸砂 | 童 宝 封 胸 砂 |
| 51 | Kosakata dan kanji N2 51 · 塩賢腕兆床 | 塩 賢 腕 兆 床 |
| 52 | Kosakata dan kanji N2 52 · 毛緑尊祝柔 | 毛 緑 尊 祝 柔 |
| 53 | Kosakata dan kanji N2 53 · 殿濃液衣肩 | 殿 濃 液 衣 肩 |
| 54 | Kosakata dan kanji N2 54 · 零幼荷泊黄 | 零 幼 荷 泊 黄 |
| 55 | Kosakata dan kanji N2 55 · 甘臣浅掃雲 | 甘 臣 浅 掃 雲 |
| 56 | Kosakata dan kanji N2 56 · 掘捨軟沈凍 | 掘 捨 軟 沈 凍 |
| 57 | Kosakata dan kanji N2 57 · 乳恋紅郊腰 | 乳 恋 紅 郊 腰 |
| 58 | Kosakata dan kanji N2 58 · 炭踊冊勇械 | 炭 踊 冊 勇 械 |
| 59 | Kosakata dan kanji N2 59 · 菜珍卵湖喫 | 菜 珍 卵 湖 喫 |
| 60 | Kosakata dan kanji N2 60 · 干虫刷湯溶 | 干 虫 刷 湯 溶 |
| 61 | Kosakata dan kanji N2 61 · 鉱涙匹孫鋭 | 鉱 涙 匹 孫 鋭 |
| 62 | Kosakata dan kanji N2 62 · 枝塗軒毒叫 | 枝 塗 軒 毒 叫 |
| 63 | Kosakata dan kanji N2 63 · 拝氷乾棒祈 | 拝 氷 乾 棒 祈 |
| 64 | Kosakata dan kanji N2 64 · 拾粉糸綿汗 | 拾 粉 糸 綿 汗 |
| 65 | Kosakata dan kanji N2 65 · 銅湿瓶咲召 | 銅 湿 瓶 咲 召 |
| 66 | Kosakata dan kanji N2 66 · 缶隻脂蒸肌 | 缶 隻 脂 蒸 肌 |
| 67 | Kosakata dan kanji N2 67 · 耕鈍泥隅灯 | 耕 鈍 泥 隅 灯 |
| 68 | Kosakata dan kanji N2 68 · 辛磨麦姓筒 | 辛 磨 麦 姓 筒 |
| 69 | Kosakata dan kanji N2 69 · 鼻粒詞胃畳 | 鼻 粒 詞 胃 畳 |
| 70 | Kosakata dan kanji N2 70 · 机膚濯塔沸 | 机 膚 濯 塔 沸 |
| 71 | Kosakata dan kanji N2 71 · 灰菓帽枯涼 | 灰 菓 帽 枯 涼 |
| 72 | Kosakata dan kanji N2 72 · 舟貝符憎 | 舟 貝 符 憎 |
| 73 | Kosakata dan kanji N2 73 · 皿肯燥畜 | 皿 肯 燥 畜 |
| 74 | Kosakata dan kanji N2 74 · 挟曇滴伺 | 挟 曇 滴 伺 |
