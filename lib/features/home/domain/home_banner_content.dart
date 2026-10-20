import 'package:nihongo_master/features/home/domain/home_dashboard.dart';

/// Five distinct destinations, even when an older API has no Manga banner yet.
/// Subscribers see Anki in place of the purchase offer.
List<HomeBanner> homeBannersFor(HomeDashboard data, {required bool lifetime}) {
  const defaults = [
    HomeBanner(
      id: 'learn',
      title: 'Belajar Jepang, setiap hari',
      description:
          'Kosakata, kanji, grammar, dan reading untuk target JLPT-mu.',
      imageUrl: '',
      action: 'learn',
      buttonLabel: 'Mulai belajar',
    ),
    HomeBanner(
      id: 'ssw',
      title: 'Siap bekerja di Jepang',
      description:
          'Kenali bidang SSW, pelajari materi, dan coba simulasi ujiannya.',
      imageUrl: '',
      action: 'ssw',
      buttonLabel: 'Jelajahi SSW',
    ),
    HomeBanner(
      id: 'jobs',
      title: 'Temukan peluang kariermu',
      description:
          'Jelajahi lowongan Jepang sesuai lokasi, bidang, dan kualifikasimu.',
      imageUrl: '',
      action: 'jobs',
      buttonLabel: 'Jelajahi lowongan',
    ),
  ];
  const manga = HomeBanner(
    id: 'manga-reading',
    title: 'Baca Manga Jepang',
    description: 'Nikmati cerita dalam bahasa aslinya. Khusus Lifetime Access.',
    imageUrl: '',
    action: 'manga',
    buttonLabel: 'Jelajahi manga',
  );
  const anki = HomeBanner(
    id: 'anki',
    title: 'Ingat lebih lama dengan Anki',
    description: 'Pilih deck favorit dan bangun kebiasaan review kosakata.',
    imageUrl: '',
    action: 'anki',
    buttonLabel: 'Jelajahi deck',
  );
  final plan =
      data.plans.where((p) => p.billingType == 'lifetime').firstOrNull ??
      data.plans.where((p) => p.recommended).firstOrNull;
  final offer = !lifetime && plan != null
      ? HomeBanner(
          id: 'lifetime-offer',
          title: plan.name,
          description:
              '${plan.features.firstOrNull ?? plan.name} • Rp${plan.priceIdr.toString().replaceAllMapped(RegExp(r"(\d)(?=(\d{3})+(?!\d))"), (m) => '${m[1]}.')}',
          imageUrl: '',
          action: 'premium',
          buttonLabel: 'Lihat paket',
        )
      : null;
  HomeBanner configured(HomeBanner fallback) =>
      data.banners.where((b) => b.action == fallback.action).firstOrNull ??
      fallback;
  return [
    if (offer != null) offer,
    ...defaults.map(configured),
    if (offer == null) configured(anki),
    configured(manga),
  ];
}
