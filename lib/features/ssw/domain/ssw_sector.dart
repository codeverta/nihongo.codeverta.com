class SswSector {
  const SswSector(
    this.id,
    this.name,
    this.japanese,
    this.romaji,
    this.keywords,
  );

  final String id;
  final String name;
  final String japanese;
  final String romaji;
  final String keywords;
}

// SSW (i) fields, checked against ssw.go.jp/id/about/visa/ on 2026-10-22.
const sswSectors = [
  SswSector(
    'nursing-care',
    'Perawatan lansia / Kaigo',
    '介護',
    'kaigo',
    'keperawatan caregiver',
  ),
  SswSector(
    'building-cleaning',
    'Pembersihan gedung',
    'ビルクリーニング',
    'biru kuriiningu',
    'cleaning kebersihan',
  ),
  SswSector(
    'linen-supply',
    'Pengolahan linen',
    'リネンサプライ',
    'rinen sapurai',
    'laundry seprai handuk',
  ),
  SswSector(
    'industrial-manufacturing',
    'Manufaktur produk industri',
    '工業製品製造業',
    'kougyou seihin seizougyou',
    'pabrik manufacturing',
  ),
  SswSector(
    'construction',
    'Konstruksi',
    '建設',
    'kensetsu',
    'bangunan kensetsu',
  ),
  SswSector(
    'shipbuilding',
    'Pembuatan kapal dan mesin kapal',
    '造船・舶用工業',
    'zousen / hakuyou kougyou',
    'shipbuilding perkapalan',
  ),
  SswSector(
    'automobile-maintenance',
    'Perawatan dan perbaikan mobil',
    '自動車整備',
    'jidousha seibi',
    'otomotif mekanik jidousha seibi',
  ),
  SswSector(
    'aviation',
    'Penerbangan',
    '航空',
    'koukuu',
    'bandara aviation koukuu',
  ),
  SswSector(
    'accommodation',
    'Perhotelan',
    '宿泊',
    'shukuhaku',
    'hotel akomodasi shukuhaku',
  ),
  SswSector(
    'road-transport',
    'Transportasi kendaraan bermotor',
    '自動車運送業',
    'jidousha unsougyou',
    'sopir truk bus taksi',
  ),
  SswSector(
    'railway',
    'Perkeretaapian',
    '鉄道',
    'tetsudou',
    'kereta railway tetsudou',
  ),
  SswSector(
    'logistics-warehouse',
    'Gudang logistik',
    '物流倉庫',
    'butsuryuu souko',
    'warehouse penyimpanan',
  ),
  SswSector(
    'agriculture',
    'Pertanian',
    '農業',
    'nougyou',
    'agriculture peternakan nougyou',
  ),
  SswSector(
    'fishery',
    'Perikanan dan budi daya',
    '漁業',
    'gyogyou',
    'aquaculture budidaya gyogyou',
  ),
  SswSector(
    'food-manufacturing',
    'Produksi makanan dan minuman',
    '飲食料品製造業',
    'inshokuryouhin seizougyou',
    'pengolahan pabrik makanan',
  ),
  SswSector(
    'food-service',
    'Layanan makanan / Restoran',
    '外食業',
    'gaishokugyou',
    'restoran gaishoku food service',
  ),
  SswSector('forestry', 'Kehutanan', '林業', 'ringyou', 'hutan forestry ringyou'),
  SswSector(
    'wood-industry',
    'Industri kayu',
    '木材産業',
    'mokuzai sangyou',
    'wood mokuzai',
  ),
  SswSector(
    'resource-recycling',
    'Daur ulang sumber daya',
    '資源循環',
    'shigen junkan',
    'recycling limbah sampah',
  ),
];

SswSector? sswSectorById(String id) {
  for (final sector in sswSectors) {
    if (sector.id == id) return sector;
  }
  return null;
}

const sswActiveSectorIds = {
  'nursing-care',
  'building-cleaning',
  'agriculture',
  'fishery',
  'linen-supply',
  'industrial-manufacturing',
  'construction',
  'shipbuilding',
  'automobile-maintenance',
  'aviation',
  'accommodation',
  'road-transport',
  'railway',
  'logistics-warehouse',
  'food-manufacturing',
  'food-service',
  'forestry',
  'wood-industry',
  'resource-recycling',
};

const sswAssessmentLevels = <String, List<String>>{
  'nursing-care': ['SSW-KAIGO', 'SSW-KAIGO-JP'],
  'building-cleaning': ['SSW-CLEANING'],
  'agriculture': ['SSW-AGRICULTURE-CROP', 'SSW-AGRICULTURE-LIVESTOCK'],
  'fishery': ['SSW-FISHERY-CAPTURE', 'SSW-FISHERY-AQUACULTURE'],
  'linen-supply': ['SSW-LINEN-SUPPLY'],
  'industrial-manufacturing': ['SSW-INDUSTRIAL-MANUFACTURING'],
  'construction': ['SSW-CONSTRUCTION'],
  'shipbuilding': ['SSW-SHIPBUILDING'],
  'automobile-maintenance': ['SSW-AUTOMOBILE-MAINTENANCE'],
  'aviation': ['SSW-AVIATION'],
  'accommodation': ['SSW-ACCOMMODATION'],
  'road-transport': ['SSW-ROAD-TRANSPORT'],
  'railway': ['SSW-RAILWAY'],
  'logistics-warehouse': ['SSW-LOGISTICS-WAREHOUSE'],
  'food-manufacturing': ['SSW-FOOD-MANUFACTURING'],
  'food-service': ['SSW-FOOD-SERVICE'],
  'forestry': ['SSW-FORESTRY'],
  'wood-industry': ['SSW-WOOD-INDUSTRY'],
  'resource-recycling': ['SSW-RESOURCE-RECYCLING'],
};
