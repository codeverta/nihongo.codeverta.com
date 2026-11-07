/// Syarat kapan aplikasi boleh menampilkan pop-up rating Play Store/App Store.
///
/// Pop-up hanya diminta lewat API resmi (In-App Review API / StoreKit) tanpa
/// pertanyaan pembuka, sesuai kebijakan Google Play dan pedoman App Store.
class AppReviewPolicy {
  const AppReviewPolicy({
    this.minimumLessons = 3,
    this.minimumActiveDays = 3,
    this.cooldown = const Duration(days: 30),
  });

  /// Jumlah pelajaran berbeda yang selesai untuk pertama kali.
  final int minimumLessons;

  /// Jumlah hari kalender berbeda saat pengguna aktif.
  final int minimumActiveDays;

  /// Jeda minimum sejak permintaan rating terakhir.
  final Duration cooldown;

  bool isEligible({
    required int completedLessons,
    required int activeDays,
    required DateTime now,
    DateTime? lastRequestedAt,
  }) =>
      completedLessons >= minimumLessons &&
      activeDays >= minimumActiveDays &&
      (lastRequestedAt == null || !now.isBefore(lastRequestedAt.add(cooldown)));

  /// Kunci hari kalender lokal, misalnya `2026-11-05`.
  static String dayKey(DateTime time) {
    final local = time.toLocal();
    String two(int value) => value.toString().padLeft(2, '0');
    return '${local.year}-${two(local.month)}-${two(local.day)}';
  }
}
