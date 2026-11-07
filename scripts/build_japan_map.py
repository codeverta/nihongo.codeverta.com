"""Build the simplified Japan prefecture geometry used by the map feature.

Input: prefectures.geojson from the npm package `open-data-jp-prefectures-geojson`
(derived from MLIT 国土数値情報 行政区域データ N03), e.g.

    npm pack open-data-jp-prefectures-geojson && tar -xzf open-data-jp-*.tgz
    python scripts/build_japan_map.py package/output/prefectures.geojson

Output (same JSON for the app and the backend):
    assets/map/japan_prefectures.json
    backend/internal/places/japan_prefectures.json

Coordinates stay in [lon, lat]. The app projects them; Okinawa (47) is drawn in
an inset by translating it with the `inset` offsets below.
"""

import json
import sys
from pathlib import Path

from shapely.geometry import Point, shape
from shapely.geometry.polygon import orient

ROOT = Path(__file__).resolve().parent.parent
TOLERANCE = 0.008  # degrees, about 800 m
MIN_AREA = 0.0004  # square degrees, about 4 km²; smaller islets are skipped
DECIMALS = 3

# (code, English/Indonesian name, Japanese name, capital, lat, lng)
PREFECTURES = [
    ("01", "Hokkaido", "北海道", "Sapporo", 43.064, 141.347),
    ("02", "Aomori", "青森県", "Aomori", 40.824, 140.740),
    ("03", "Iwate", "岩手県", "Morioka", 39.704, 141.153),
    ("04", "Miyagi", "宮城県", "Sendai", 38.269, 140.872),
    ("05", "Akita", "秋田県", "Akita", 39.719, 140.102),
    ("06", "Yamagata", "山形県", "Yamagata", 38.240, 140.364),
    ("07", "Fukushima", "福島県", "Fukushima", 37.750, 140.468),
    ("08", "Ibaraki", "茨城県", "Mito", 36.342, 140.447),
    ("09", "Tochigi", "栃木県", "Utsunomiya", 36.566, 139.884),
    ("10", "Gunma", "群馬県", "Maebashi", 36.391, 139.061),
    ("11", "Saitama", "埼玉県", "Saitama", 35.857, 139.649),
    ("12", "Chiba", "千葉県", "Chiba", 35.605, 140.123),
    ("13", "Tokyo", "東京都", "Tokyo", 35.690, 139.692),
    ("14", "Kanagawa", "神奈川県", "Yokohama", 35.448, 139.642),
    ("15", "Niigata", "新潟県", "Niigata", 37.902, 139.023),
    ("16", "Toyama", "富山県", "Toyama", 36.695, 137.211),
    ("17", "Ishikawa", "石川県", "Kanazawa", 36.594, 136.626),
    ("18", "Fukui", "福井県", "Fukui", 36.065, 136.222),
    ("19", "Yamanashi", "山梨県", "Kofu", 35.664, 138.568),
    ("20", "Nagano", "長野県", "Nagano", 36.651, 138.181),
    ("21", "Gifu", "岐阜県", "Gifu", 35.391, 136.722),
    ("22", "Shizuoka", "静岡県", "Shizuoka", 34.977, 138.383),
    ("23", "Aichi", "愛知県", "Nagoya", 35.180, 136.907),
    ("24", "Mie", "三重県", "Tsu", 34.730, 136.509),
    ("25", "Shiga", "滋賀県", "Otsu", 35.004, 135.869),
    ("26", "Kyoto", "京都府", "Kyoto", 35.021, 135.756),
    ("27", "Osaka", "大阪府", "Osaka", 34.686, 135.520),
    ("28", "Hyogo", "兵庫県", "Kobe", 34.691, 135.183),
    ("29", "Nara", "奈良県", "Nara", 34.685, 135.833),
    ("30", "Wakayama", "和歌山県", "Wakayama", 34.226, 135.168),
    ("31", "Tottori", "鳥取県", "Tottori", 35.504, 134.238),
    ("32", "Shimane", "島根県", "Matsue", 35.472, 133.051),
    ("33", "Okayama", "岡山県", "Okayama", 34.662, 133.935),
    ("34", "Hiroshima", "広島県", "Hiroshima", 34.397, 132.460),
    ("35", "Yamaguchi", "山口県", "Yamaguchi", 34.186, 131.471),
    ("36", "Tokushima", "徳島県", "Tokushima", 34.066, 134.559),
    ("37", "Kagawa", "香川県", "Takamatsu", 34.340, 134.043),
    ("38", "Ehime", "愛媛県", "Matsuyama", 33.842, 132.766),
    ("39", "Kochi", "高知県", "Kochi", 33.560, 133.531),
    ("40", "Fukuoka", "福岡県", "Fukuoka", 33.607, 130.418),
    ("41", "Saga", "佐賀県", "Saga", 33.249, 130.299),
    ("42", "Nagasaki", "長崎県", "Nagasaki", 32.745, 129.874),
    ("43", "Kumamoto", "熊本県", "Kumamoto", 32.790, 130.742),
    ("44", "Oita", "大分県", "Oita", 33.238, 131.613),
    ("45", "Miyazaki", "宮崎県", "Miyazaki", 31.911, 131.424),
    ("46", "Kagoshima", "鹿児島県", "Kagoshima", 31.560, 130.558),
    ("47", "Okinawa", "沖縄県", "Naha", 26.212, 127.681),
]

# Okinawa is moved into the empty Sea of Japan area (upper left), as on
# printed maps of Japan. Applies to its outline, capital, and spot pins.
INSET = {"code": "47", "dLon": 3.6, "dLat": 15.5}

# Tokyo's Izu and Ogasawara islands south of this latitude would stretch the map
# far into the Pacific; they are left out of the drawing.
TOKYO_MIN_LAT = 33.0


def rings(geometry, min_lat=None):
    polygons = geometry.geoms if geometry.geom_type == "MultiPolygon" else [geometry]
    out = []
    for polygon in polygons:
        if min_lat is not None and polygon.centroid.y < min_lat:
            continue
        polygon = orient(polygon)
        simple = polygon.simplify(TOLERANCE, preserve_topology=True)
        if simple.is_empty or simple.area < MIN_AREA:
            continue
        coords = [[round(x, DECIMALS), round(y, DECIMALS)] for x, y in simple.exterior.coords]
        deduped = [c for i, c in enumerate(coords) if i == 0 or c != coords[i - 1]]
        if len(deduped) >= 4:
            out.append(deduped)
    return out


def main(source: str) -> None:
    data = json.loads(Path(source).read_text(encoding="utf-8"))
    by_name = {f["properties"]["P"]: shape(f["geometry"]) for f in data["features"]}
    result = []
    for code, name, name_ja, capital, lat, lng in PREFECTURES:
        geometry = by_name[name_ja]
        if not geometry.buffer(0).contains(Point(lng, lat)):
            raise SystemExit(f"capital of {name} is outside its prefecture")
        polygons = rings(geometry, TOKYO_MIN_LAT if code == "13" else None)
        if not polygons:
            raise SystemExit(f"{name} has no polygon after simplification")
        result.append({
            "code": code,
            "name": name,
            "nameJa": name_ja,
            "capital": {"name": capital, "lat": lat, "lng": lng},
            "polygons": polygons,
        })
    output = {
        "source": "国土数値情報（行政区域データ）国土交通省, via open-data-jp-prefectures-geojson; simplified",
        "inset": INSET,
        "prefectures": result,
    }
    text = json.dumps(output, ensure_ascii=False, separators=(",", ":"))
    for target in ("assets/map/japan_prefectures.json", "backend/internal/places/japan_prefectures.json"):
        path = ROOT / target
        path.parent.mkdir(parents=True, exist_ok=True)
        path.write_text(text, encoding="utf-8")
        print(f"wrote {target}: {len(text) // 1024} KiB, {sum(len(p['polygons']) for p in result)} polygons")


if __name__ == "__main__":
    main(sys.argv[1])
