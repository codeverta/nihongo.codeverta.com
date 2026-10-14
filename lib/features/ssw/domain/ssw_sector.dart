class SswSector {
  const SswSector(this.id, this.name, this.japanese, this.keywords);

  final String id;
  final String name;
  final String japanese;
  final String keywords;
}

// SSW (i) fields, checked against ssw.go.jp/id/about/visa/ on 2026-10-11.
const sswSectors = [
  SswSector(
    'nursing-care',
    'Perawatan lansia / Kaigo',
    '介護',
    'keperawatan caregiver',
  ),
  SswSector(
    'building-cleaning',
    'Pembersihan gedung',
    'ビルクリーニング',
    'cleaning kebersihan',
  ),
  SswSector(
    'linen-supply',
    'Pasokan linen',
    'リネンサプライ',
    'laundry seprai handuk',
  ),
  SswSector(
    'industrial-manufacturing',
    'Manufaktur produk industri',
    '工業製品製造業',
    'pabrik manufacturing',
  ),
  SswSector('construction', 'Konstruksi', '建設', 'bangunan kensetsu'),
  SswSector(
    'shipbuilding',
    'Pembuatan kapal dan mesin kapal',
    '造船・舶用工業',
    'shipbuilding perkapalan',
  ),
  SswSector(
    'automobile-maintenance',
    'Perawatan dan perbaikan mobil',
    '自動車整備',
    'otomotif mekanik jidousha seibi',
  ),
  SswSector('aviation', 'Penerbangan', '航空', 'bandara aviation koukuu'),
  SswSector('accommodation', 'Perhotelan', '宿泊', 'hotel akomodasi shukuhaku'),
  SswSector(
    'road-transport',
    'Transportasi kendaraan bermotor',
    '自動車運送業',
    'sopir truk bus taksi',
  ),
  SswSector('railway', 'Perkeretaapian', '鉄道', 'kereta railway tetsudou'),
  SswSector(
    'logistics-warehouse',
    'Gudang logistik',
    '物流倉庫',
    'warehouse penyimpanan',
  ),
  SswSector('agriculture', 'Pertanian', '農業', 'agriculture peternakan nougyou'),
  SswSector(
    'fishery',
    'Perikanan dan budi daya',
    '漁業',
    'aquaculture budidaya gyogyou',
  ),
  SswSector(
    'food-manufacturing',
    'Produksi makanan dan minuman',
    '飲食料品製造業',
    'pengolahan pabrik makanan',
  ),
  SswSector(
    'food-service',
    'Layanan makanan / Restoran',
    '外食業',
    'restoran gaishoku food service',
  ),
  SswSector('forestry', 'Kehutanan', '林業', 'hutan forestry ringyou'),
  SswSector('wood-industry', 'Industri kayu', '木材産業', 'wood mokuzai'),
  SswSector(
    'resource-recycling',
    'Daur ulang sumber daya',
    '資源循環',
    'recycling limbah sampah',
  ),
];

SswSector? sswSectorById(String id) {
  for (final sector in sswSectors) {
    if (sector.id == id) return sector;
  }
  return null;
}
