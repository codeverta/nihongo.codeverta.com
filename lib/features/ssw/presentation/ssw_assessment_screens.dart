import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';
import 'package:nihongo_master/core/network/dio_client.dart';
import 'package:nihongo_master/features/auth/presentation/controllers/auth_controller.dart';
import 'package:nihongo_master/features/exam/data/diagnostic_repository.dart';
import 'package:nihongo_master/features/exam/data/simulation_repository.dart';
import 'package:nihongo_master/features/ssw/domain/ssw_sector.dart';

bool sswAvailable(String sector) => sswActiveSectorIds.contains(sector);

String sswExamTitle(String level) => switch (level) {
  'SSW-KAIGO' => 'Kaigo · Keterampilan',
  'SSW-KAIGO-JP' => 'Kaigo · Bahasa Jepang',
  'SSW-CLEANING' => 'Pembersihan Gedung',
  'SSW-AGRICULTURE-CROP' => 'Pertanian · Tanaman',
  'SSW-AGRICULTURE-LIVESTOCK' => 'Pertanian · Peternakan',
  'SSW-FISHERY-CAPTURE' => 'Perikanan Tangkap',
  'SSW-FISHERY-AQUACULTURE' => 'Budi Daya',
  'SSW-LINEN-SUPPLY' => 'Pengolahan Linen',
  'SSW-INDUSTRIAL-MANUFACTURING' => 'Manufaktur Produk Industri',
  'SSW-CONSTRUCTION' => 'Konstruksi',
  'SSW-SHIPBUILDING' => 'Pembuatan Kapal dan Mesin Kapal',
  'SSW-AUTOMOBILE-MAINTENANCE' => 'Perawatan dan Perbaikan Mobil',
  'SSW-AVIATION' => 'Penerbangan',
  'SSW-ACCOMMODATION' => 'Perhotelan',
  'SSW-ROAD-TRANSPORT' => 'Transportasi Kendaraan Bermotor',
  'SSW-RAILWAY' => 'Perkeretaapian',
  'SSW-LOGISTICS-WAREHOUSE' => 'Gudang Logistik',
  'SSW-FOOD-MANUFACTURING' => 'Produksi Makanan dan Minuman',
  'SSW-FOOD-SERVICE' => 'Layanan Makanan / Restoran',
  'SSW-FORESTRY' => 'Kehutanan',
  'SSW-WOOD-INDUSTRY' => 'Industri Kayu',
  'SSW-RESOURCE-RECYCLING' => 'Daur Ulang Sumber Daya',
  _ => 'SSW',
};

String sswExamSummary(String level) => switch (level) {
  'SSW-KAIGO' => '45 soal · 60 menit · 5 ranah keterampilan',
  'SSW-KAIGO-JP' => '15 soal · 30 menit · soal bahasa Jepang',
  'SSW-CLEANING' => '50 soal · 50 menit · teori + penilaian praktik',
  'SSW-AGRICULTURE-CROP' || 'SSW-AGRICULTURE-LIVESTOCK' =>
    '70 soal · 60 menit · adaptasi tertulis tanpa audio',
  'SSW-FISHERY-CAPTURE' => '55 soal · 70 menit · teori + keputusan praktik',
  'SSW-LINEN-SUPPLY' ||
  'SSW-INDUSTRIAL-MANUFACTURING' ||
  'SSW-CONSTRUCTION' ||
  'SSW-SHIPBUILDING' => '24 soal · 35 menit · latihan belajar orisinal',
  'SSW-AUTOMOBILE-MAINTENANCE' ||
  'SSW-AVIATION' ||
  'SSW-ACCOMMODATION' ||
  'SSW-ROAD-TRANSPORT' ||
  'SSW-RAILWAY' ||
  'SSW-LOGISTICS-WAREHOUSE' ||
  'SSW-FOOD-MANUFACTURING' ||
  'SSW-FOOD-SERVICE' ||
  'SSW-FORESTRY' ||
  'SSW-WOOD-INDUSTRY' ||
  'SSW-RESOURCE-RECYCLING' => '24 soal · 35 menit · latihan belajar orisinal',
  _ => '50 soal · 70 menit · teori + keputusan praktik',
};

String sswExamDescription(String level) => switch (level) {
  'SSW-KAIGO' =>
    'Dasar Kaigo 10 soal · tubuh dan mental 6 · komunikasi 4 · dukungan kehidupan 20 · keputusan praktik 5. Bahasa soal Indonesia; istilah Jepang sesuai materi.',
  'SSW-KAIGO-JP' =>
    'Kosakata 5 soal · percakapan 5 · dokumen 5. Soal bahasa Jepang dengan bacaan kana dalam tanda kurung dan pembahasan Indonesia.',
  'SSW-CLEANING' =>
    'Teori 20 soal/20 menit, lalu penilaian praktik 30 soal/30 menit. Bagian pertama dikunci saat berpindah. Soal Indonesia dengan istilah Jepang; ujian resmi menggunakan bahasa Jepang dengan furigana.',
  'SSW-AGRICULTURE-CROP' =>
    'Jalur tanaman: tanah, pupuk, benih, fasilitas, pemeliharaan, perlindungan, panen, dan hasil. Satu timer 60 menit untuk 70 soal tertulis Indonesia dengan istilah Jepang. Ujian resmi juga memuat mendengar instruksi Jepang; audio belum tersedia di simulasi ini.',
  'SSW-AGRICULTURE-LIVESTOCK' =>
    'Jalur peternakan: biosekuriti, pakan, lingkungan, reproduksi, sapi, babi, unggas, dan produk. Satu timer 60 menit untuk 70 soal tertulis Indonesia dengan istilah Jepang. Ujian resmi juga memuat mendengar instruksi Jepang; audio belum tersedia di simulasi ini.',
  'SSW-FISHERY-CAPTURE' =>
    'Perikanan tangkap: dasar, keselamatan, alat jaring, pancing, dan mutu hasil. Teori 30 soal/50 menit, lalu 25 keputusan praktik/20 menit dengan timer baru. Soal Indonesia dengan istilah Jepang; ujian resmi menggunakan bahasa Jepang.',
  'SSW-LINEN-SUPPLY' =>
    'Alur pengolahan linen hotel dan rumah sakit: penerimaan, pemilahan, pencucian, finishing, pemeriksaan, pengemasan, dan pengiriman. Paket ini latihan belajar orisinal; sumber resmi yang ditinjau belum memuat blueprint atau format ujian rinci. Penilaian praktik langsung tidak direplikasi.',
  'SSW-INDUSTRIAL-MANUFACTURING' =>
    'Prinsip lintas kategori untuk keselamatan, mutu, proses, pemeriksaan, dan pelaporan. Sektor ini mencakup 17 kategori kerja resmi; setiap kategori memiliki standar teknis tersendiri. Paket ini tidak menggantikan standar kategori, demonstrasi alat, gambar kerja, atau format CBT resmi.',
  'SSW-CONSTRUCTION' =>
    'Dasar kerja lintas divisi sipil, bangunan, serta lifeline dan peralatan: membaca instruksi kerja, keselamatan, koordinasi, mutu, dan serah-terima. Latihan belajar orisinal ini tidak mereplikasi format ujian resmi, gambar proyek khusus, atau kerja praktik lapangan.',
  'SSW-SHIPBUILDING' =>
    'Pengetahuan dan skenario untuk alur pembangunan kapal, mesin kapal, dan peralatan listrik kapal. Divisi kerja resmi berbeda dari kategori trade ujian; simulasi teks ini tidak mencakup uji keterampilan langsung di lokasi atau demonstrasi trade tertentu.',
  'SSW-AUTOMOBILE-MAINTENANCE' =>
    'Latihan dasar pemeriksaan kendaraan, perawatan berkala, penggunaan alat secara aman, pencatatan, dan komunikasi temuan. Ikuti manual kendaraan, instruksi bengkel, serta arahan teknisi yang berwenang; simulasi teks ini tidak menggantikan praktik servis atau sertifikasi mekanik.',
  'SSW-AVIATION' =>
    'Latihan untuk tugas penerbangan yang berbeda: penanganan darat dan bagasi, atau pemeriksaan/perawatan pesawat sesuai penugasan. Keselamatan apron, komunikasi, identifikasi alat, dan eskalasi jadi prioritas. Tidak menggantikan pelatihan berizin, praktik lapangan, atau prosedur operator.',
  'SSW-ACCOMMODATION' =>
    'Situasi hotel: reservasi dan front desk, menyampaikan informasi, menangani permintaan/tanggapan tamu, serta koordinasi kebersihan dan layanan makan sesuai peran. Dialog memakai bahasa Jepang sederhana. Ikuti kebijakan hotel untuk privasi, uang, kehilangan barang, dan keadaan darurat.',
  'SSW-ROAD-TRANSPORT' =>
    'Latihan berbasis skenario untuk pengemudi truk, bus, dan taksi: pemeriksaan sebelum berangkat, komunikasi dengan penumpang/penerima barang, pemilihan tindakan aman, serta laporan gangguan. Jenis SIM dan syarat berbeda menurut kendaraan; materi ini bukan panduan memperoleh SIM atau menggantikan pelatihan mengemudi.',
  'SSW-RAILWAY' =>
    'Latihan pengenalan tugas operasi stasiun/kereta, perawatan sarana, jalur, dan fasilitas sesuai penempatan. Fokus pada aturan keselamatan, komunikasi yang terkonfirmasi, pengamanan area kerja, pelaporan kondisi tidak normal, dan serah terima. Tidak memberikan kewenangan mengoperasikan sistem perkeretaapian.',
  'SSW-LOGISTICS-WAREHOUSE' =>
    'Latihan penerimaan, pemeriksaan, penyimpanan, picking, packing, dan pengiriman barang di gudang logistik, termasuk identitas lot dan keselamatan pergerakan barang. Pemerintah Jepang menyatakan rincian pelaksanaan ujian masih disiapkan; paket ini latihan orisinal dan tidak mengklaim format resmi.',
  'SSW-FOOD-MANUFACTURING' =>
    'Latihan higiene produksi, pencegahan kontaminasi silang, penanganan alergen, kontrol proses, catatan lot, kebersihan alat, dan pelaporan penyimpangan. Batas suhu dan prosedur harus mengikuti standar produk serta SOP fasilitas; contoh tidak menggantikan HACCP atau instruksi perusahaan.',
  'SSW-FOOD-SERVICE' =>
    'Latihan layanan restoran: sambutan dan pesanan, konfirmasi alergi, penyajian, kebersihan, penanganan keluhan, dan kerja sama dapur. Status situs resmi SSW yang ditinjau menyatakan penerimaan baru pada bidang ini sedang tidak dibuka; periksa pengumuman resmi terbaru sebelum merencanakan pengajuan.',
  'SSW-FORESTRY' =>
    'Latihan tahapan pembibitan/penanaman, pemeliharaan tegakan, pekerjaan hutan, panen dan pemindahan hasil dengan perhatian pada cuaca, lereng, alat, komunikasi, serta konservasi. Operasi gergaji dan mesin hanya boleh dilakukan setelah pelatihan serta izin kerja setempat.',
  'SSW-WOOD-INDUSTRY' =>
    'Latihan alur pengolahan kayu dan bahan berbasis kayu: identifikasi bahan, pengukuran, pemotongan sesuai instruksi, pengendalian debu dan mesin, pemeriksaan mutu, serta pencatatan. Batas toleransi dan pengaturan mesin selalu mengikuti gambar kerja, spesifikasi, dan operator berwenang.',
  'SSW-RESOURCE-RECYCLING' =>
    'Latihan penerimaan dan pemilahan material, identifikasi kontaminan, penelusuran aliran barang, penggunaan area dan mesin secara aman, serta penanganan temuan yang tidak dikenal. Rincian ujian resmi belum diumumkan pada halaman Kementerian Lingkungan Jepang yang ditinjau. Jangan membuka atau mencampur material berbahaya; isolasi dan laporkan sesuai prosedur fasilitas.',
  _ =>
    'Budi daya: umum, keselamatan, kualitas air, sistem berpakan, serta kerang dan rumput laut tanpa pakan tambahan. Teori 40 soal/50 menit, lalu 10 keputusan praktik/20 menit dengan timer baru. Soal Indonesia dengan istilah Jepang; ujian resmi menggunakan bahasa Jepang.',
};

class SswAssessmentPanel extends ConsumerWidget {
  const SswAssessmentPanel({required this.sector, super.key});
  final String sector;

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final premium = ref.watch(
      authControllerProvider.select((s) => s.user?.isPremium ?? false),
    );
    return Column(
      crossAxisAlignment: CrossAxisAlignment.stretch,
      children: [
        Text(
          'Latih pemahaman. Ukur kesiapan.',
          style: Theme.of(context).textTheme.titleLarge,
        ),
        const SizedBox(height: 12),
        Text(
          premium
              ? 'Latihan per topik memberikan pembahasan langsung. Simulasi memakai timer server dan pembahasan setelah selesai.'
              : 'Latihan dan simulasi termasuk Lifetime Access. Materi bidang tetap dapat dibaca gratis.',
        ),
        const SizedBox(height: 16),
        Card(
          child: ListTile(
            contentPadding: const EdgeInsets.all(16),
            leading: const Icon(Icons.psychology_alt_outlined),
            title: const Text('Latihan per topik'),
            subtitle: Text(switch (sector) {
              'nursing-care' =>
                '87 soal · 12 modul + bahasa Jepang · tanpa timer',
              'agriculture' => '160 soal · 10 modul · tanaman & peternakan',
              'fishery' => '120 soal · 10 modul · tangkap & budi daya',
              'linen-supply' => '48 soal · 4 modul · pengolahan linen',
              'industrial-manufacturing' =>
                '36 soal · 4 modul · lintas kategori industri',
              'construction' => '36 soal · 4 modul · konstruksi',
              'shipbuilding' => '36 soal · 4 modul · kapal & mesin kapal',
              'automobile-maintenance' =>
                '36 soal · 4 modul · perawatan dan pemeriksaan kendaraan',
              'aviation' =>
                '36 soal · 4 modul · layanan darat dan perawatan pesawat',
              'accommodation' =>
                '36 soal · 4 modul · komunikasi dan layanan hotel',
              'road-transport' => '36 soal · 4 modul · truk, bus, dan taksi',
              'railway' => '36 soal · 4 modul · operasi dan keselamatan kereta',
              'logistics-warehouse' =>
                '36 soal · 4 modul · alur kerja gudang logistik',
              'food-manufacturing' =>
                '36 soal · 4 modul · higiene dan kendali produksi pangan',
              'food-service' =>
                '36 soal · 4 modul · layanan restoran dan higiene pangan',
              'forestry' => '36 soal · 4 modul · keselamatan kerja kehutanan',
              'wood-industry' => '36 soal · 4 modul · pengolahan dan mutu kayu',
              'resource-recycling' =>
                '36 soal · 4 modul · pemilahan dan keselamatan fasilitas daur ulang',
              _ => '60 soal · 10 modul · tanpa timer',
            }),
            trailing: Icon(premium ? Icons.chevron_right : Icons.lock_outline),
            onTap: () => context.push('/ssw/$sector/practice'),
          ),
        ),
        for (final level in sswAssessmentLevels[sector] ?? <String>[])
          Card(
            child: ListTile(
              contentPadding: const EdgeInsets.all(16),
              leading: const Icon(Icons.timer_outlined),
              title: Text('Simulasi · ${sswExamTitle(level)}'),
              subtitle: Text(sswExamSummary(level)),
              trailing: Icon(
                premium ? Icons.chevron_right : Icons.lock_outline,
              ),
              onTap: () =>
                  context.push('/exams/${level.toLowerCase()}/simulation'),
            ),
          ),
        const SizedBox(height: 12),
        const Text(
          'Soal orisinal persiapan SSW 1号. Bagian praktik memakai skenario teks; format foto, ilustrasi, dan demonstrasi alat tidak seluruhnya direplikasi. Hasil aplikasi bukan kelulusan resmi.',
          style: TextStyle(height: 1.6, fontSize: 13),
        ),
      ],
    );
  }
}

class SswSimulationIntroScreen extends ConsumerStatefulWidget {
  const SswSimulationIntroScreen({required this.level, super.key});
  final String level;
  @override
  ConsumerState<SswSimulationIntroScreen> createState() => _SswIntroState();
}

class _SswIntroState extends ConsumerState<SswSimulationIntroScreen> {
  bool busy = false;
  String? error;
  Future<void> start() async {
    setState(() {
      busy = true;
      error = null;
    });
    try {
      final id = await ref
          .read(simulationLevelRepositoryProvider(widget.level))
          .start();
      if (!mounted) return;
      await context.push('/exams/${widget.level.toLowerCase()}/simulation/$id');
      if (!mounted) return;
      ref.invalidate(simulationLevelOverviewProvider(widget.level));
    } on Object catch (e) {
      if (mounted) setState(() => error = diagnosticError(e));
    } finally {
      if (mounted) setState(() => busy = false);
    }
  }

  @override
  Widget build(BuildContext context) => Scaffold(
    appBar: AppBar(title: Text('Simulasi ${sswExamTitle(widget.level)}')),
    body: ref
        .watch(simulationLevelOverviewProvider(widget.level))
        .when(
          loading: () => const Center(child: CircularProgressIndicator()),
          error: (e, _) => Center(
            child: Padding(
              padding: const EdgeInsets.all(24),
              child: Column(
                mainAxisSize: MainAxisSize.min,
                children: [
                  Text(diagnosticError(e)),
                  OutlinedButton(
                    onPressed: () => ref.invalidate(
                      simulationLevelOverviewProvider(widget.level),
                    ),
                    child: const Text('Coba lagi'),
                  ),
                ],
              ),
            ),
          ),
          data: (data) => ListView(
            padding: const EdgeInsets.all(20),
            children: [
              Container(
                padding: const EdgeInsets.all(24),
                decoration: BoxDecoration(
                  color: Theme.of(context).colorScheme.primaryContainer,
                  borderRadius: BorderRadius.circular(24),
                ),
                child: Column(
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: [
                    const Icon(Icons.assignment_outlined, size: 38),
                    const SizedBox(height: 16),
                    Text(
                      data['title'] as String,
                      style: Theme.of(context).textTheme.headlineSmall,
                    ),
                    const SizedBox(height: 12),
                    Text(
                      '${data['questionCount']} soal · ${data['durationMinutes']} menit',
                    ),
                  ],
                ),
              ),
              const SizedBox(height: 20),
              for (final section in diagnosticList(data['sections']))
                ListTile(
                  contentPadding: EdgeInsets.zero,
                  leading: const Icon(Icons.schedule),
                  title: Text(section['name'] as String),
                  subtitle: Text(
                    '${section['questionCount']} soal · ${(section['durationSeconds'] as int) ~/ 60} menit',
                  ),
                ),
              const SizedBox(height: 12),
              Text(sswExamDescription(widget.level)),
              const SizedBox(height: 20),
              const Text(
                'Timer tetap berjalan ketika aplikasi ditutup. Jawaban tersimpan setelah dikonfirmasi server. Gunakan peta soal dan tanda ragu. Setelah bagian dikirim atau waktu habis, jawaban dikunci. Kunci serta pembahasan tersedia setelah seluruh bagian selesai. Sesi dan riwayat tersimpan pada akun di server.',
                style: TextStyle(height: 1.7),
              ),
              const SizedBox(height: 16),
              const Text(
                'Simulasi belajar orisinal, bukan soal resmi. Keputusan praktik menggunakan skenario teks; tidak menggantikan latihan lapangan terawasi atau seluruh variasi visual CBT resmi.',
                style: TextStyle(height: 1.6, fontSize: 13),
              ),
              if (error != null)
                Padding(
                  padding: const EdgeInsets.symmetric(vertical: 12),
                  child: Text(
                    error!,
                    style: TextStyle(
                      color: Theme.of(context).colorScheme.error,
                    ),
                  ),
                ),
              const SizedBox(height: 20),
              FilledButton.icon(
                onPressed: busy ? null : start,
                icon: busy
                    ? const SizedBox(
                        width: 18,
                        height: 18,
                        child: CircularProgressIndicator(strokeWidth: 2),
                      )
                    : const Icon(Icons.play_arrow),
                label: Text(
                  (data['activeAttemptId'] as String).isEmpty
                      ? 'Mulai simulasi'
                      : 'Lanjutkan simulasi',
                ),
              ),
              const SizedBox(height: 24),
              Text(
                'Riwayat simulasi',
                style: Theme.of(context).textTheme.titleLarge,
              ),
              if (diagnosticList(data['history']).isEmpty)
                const Padding(
                  padding: EdgeInsets.symmetric(vertical: 12),
                  child: Text('Belum ada simulasi yang selesai.'),
                ),
              for (final row in diagnosticList(data['history']))
                ListTile(
                  contentPadding: EdgeInsets.zero,
                  title: Text(
                    '${row['correct']}/${row['total']} benar · ${row['score']}%',
                  ),
                  subtitle: Text(
                    (row['completedAt'] as String).split('T').first,
                  ),
                  trailing: const Icon(Icons.chevron_right),
                  onTap: () => context.push(
                    '/exams/${widget.level.toLowerCase()}/simulation/${row['attemptId']}',
                  ),
                ),
            ],
          ),
        ),
  );
}

final sswPracticeProvider = FutureProvider.autoDispose
    .family<List<Json>, String>((ref, sector) async {
      ref.watch(authControllerProvider.select((s) => s.user?.id));
      final repo = DiagnosticRepository(ref.watch(dioProvider));
      final data = await repo.request('/ssw/$sector/practice');
      return diagnosticList(data['questions']);
    });

class SswPracticeScreen extends ConsumerStatefulWidget {
  const SswPracticeScreen({required this.sector, super.key});
  final String sector;
  @override
  ConsumerState<SswPracticeScreen> createState() => _PracticeState();
}

class _PracticeState extends ConsumerState<SswPracticeScreen> {
  String? module;
  int index = 0;
  int? selected;
  int correct = 0;
  bool busy = false;
  bool finished = false;
  Json? feedback;
  String? error;
  void reset(String? value) => setState(() {
    module = value;
    index = 0;
    selected = null;
    correct = 0;
    finished = false;
    feedback = null;
    error = null;
  });

  Future<void> check(Json question) async {
    if (selected == null || busy || feedback != null) return;
    setState(() {
      busy = true;
      error = null;
    });
    try {
      final result = await DiagnosticRepository(ref.read(dioProvider)).request(
        '/ssw/${widget.sector}/practice/check',
        method: 'POST',
        body: {'questionId': question['id'], 'answer': selected},
      );
      if (!mounted) return;
      setState(() {
        feedback = result;
        if (result['correct'] == true) correct++;
      });
    } on Object catch (e) {
      if (mounted) setState(() => error = diagnosticError(e));
    } finally {
      if (mounted) setState(() => busy = false);
    }
  }

  @override
  Widget build(BuildContext context) {
    // An account switch resets ephemeral practice answers, including in-flight work.
    ref.listen(authControllerProvider.select((s) => s.user?.id), (_, __) {
      if (mounted) context.go('/ssw/${widget.sector}?tab=test');
    });
    return Scaffold(
      appBar: AppBar(title: const Text('Latihan SSW · per topik')),
      body: ref
          .watch(sswPracticeProvider(widget.sector))
          .when(
            loading: () => const Center(child: CircularProgressIndicator()),
            error: (e, _) => Center(
              child: Padding(
                padding: const EdgeInsets.all(24),
                child: Column(
                  mainAxisSize: MainAxisSize.min,
                  children: [
                    Text(diagnosticError(e)),
                    OutlinedButton(
                      onPressed: () =>
                          ref.invalidate(sswPracticeProvider(widget.sector)),
                      child: const Text('Coba lagi'),
                    ),
                  ],
                ),
              ),
            ),
            data: (all) {
              final groups = <String, String>{
                for (final q in all)
                  q['moduleId'] as String: q['group'] as String,
              };
              final questions = all
                  .where((q) => module == null || q['moduleId'] == module)
                  .toList();
              if (module == null) {
                return ListView(
                  padding: const EdgeInsets.all(20),
                  children: [
                    Text(
                      'Pilih topik latihan',
                      style: Theme.of(context).textTheme.headlineSmall,
                    ),
                    const SizedBox(height: 12),
                    const Text(
                      'Tanpa timer. Periksa jawaban untuk melihat pembahasan, lalu lanjutkan. Hasil latihan berlaku selama sesi ini dan tidak memberi XP.',
                    ),
                    const SizedBox(height: 16),
                    for (final entry in groups.entries)
                      Card(
                        child: ListTile(
                          leading: const Icon(Icons.quiz_outlined),
                          title: Text(entry.value),
                          subtitle: Text(
                            '${all.where((q) => q['moduleId'] == entry.key).length} soal',
                          ),
                          trailing: const Icon(Icons.chevron_right),
                          onTap: () => reset(entry.key),
                        ),
                      ),
                  ],
                );
              }
              if (finished) {
                return ListView(
                  padding: const EdgeInsets.all(24),
                  children: [
                    const Icon(Icons.task_alt, size: 64),
                    const SizedBox(height: 20),
                    Text(
                      'Latihan selesai',
                      style: Theme.of(context).textTheme.headlineSmall,
                    ),
                    const SizedBox(height: 12),
                    Text(
                      '$correct/${questions.length} benar pada jawaban pertama',
                    ),
                    const SizedBox(height: 20),
                    FilledButton(
                      onPressed: () => reset(module),
                      child: const Text('Ulangi topik'),
                    ),
                    OutlinedButton(
                      onPressed: () => reset(null),
                      child: const Text('Pilih topik lain'),
                    ),
                  ],
                );
              }
              final q = questions[index];
              final options = (q['options'] as List).cast<String>();
              return ListView(
                key: ValueKey(q['id']),
                padding: const EdgeInsets.all(20),
                children: [
                  Text(
                    '${q['group']}',
                    style: Theme.of(context).textTheme.titleMedium,
                  ),
                  const SizedBox(height: 12),
                  LinearProgressIndicator(value: index / questions.length),
                  const SizedBox(height: 12),
                  Text(
                    'Soal ${index + 1}/${questions.length} · $correct benar',
                  ),
                  const SizedBox(height: 20),
                  Text(
                    q['prompt'] as String,
                    style: const TextStyle(
                      fontSize: 20,
                      fontWeight: FontWeight.w700,
                      height: 1.6,
                    ),
                  ),
                  const SizedBox(height: 20),
                  for (var i = 0; i < options.length; i++)
                    Padding(
                      padding: const EdgeInsets.only(bottom: 10),
                      child: OutlinedButton(
                        key: Key('ssw-practice-answer-$i'),
                        onPressed: busy || feedback != null
                            ? null
                            : () => setState(() => selected = i),
                        style: OutlinedButton.styleFrom(
                          padding: const EdgeInsets.all(16),
                          alignment: Alignment.centerLeft,
                          backgroundColor: selected == i
                              ? Theme.of(context).colorScheme.primaryContainer
                              : null,
                          disabledForegroundColor: Theme.of(
                            context,
                          ).colorScheme.onSurface,
                        ),
                        child: Text(
                          '${selected == i ? '●' : '○'}  ${options[i]}',
                        ),
                      ),
                    ),
                  if (error != null)
                    Text(
                      error!,
                      style: TextStyle(
                        color: Theme.of(context).colorScheme.error,
                      ),
                    ),
                  if (feedback == null)
                    FilledButton(
                      onPressed: selected == null || busy
                          ? null
                          : () => check(q),
                      child: Text(busy ? 'Memeriksa…' : 'Periksa jawaban'),
                    ),
                  if (feedback != null) ...[
                    Card(
                      child: Padding(
                        padding: const EdgeInsets.all(16),
                        child: Column(
                          crossAxisAlignment: CrossAxisAlignment.start,
                          children: [
                            Text(
                              feedback!['correct'] == true
                                  ? 'Tepat'
                                  : 'Belum tepat',
                              style: const TextStyle(
                                fontWeight: FontWeight.w800,
                              ),
                            ),
                            const SizedBox(height: 8),
                            Text(
                              'Jawaban: ${options[feedback!['answer'] as int]}',
                            ),
                            const SizedBox(height: 12),
                            Text(
                              feedback!['explanation'] as String,
                              style: const TextStyle(height: 1.7),
                            ),
                          ],
                        ),
                      ),
                    ),
                    TextButton.icon(
                      onPressed: () => context.push(
                        '/ssw/${widget.sector}/lessons/${q['lessonId']}',
                      ),
                      icon: const Icon(Icons.menu_book_outlined),
                      label: const Text('Pelajari materi terkait'),
                    ),
                    FilledButton(
                      onPressed: () => setState(() {
                        if (index + 1 == questions.length) {
                          finished = true;
                        } else {
                          index++;
                          selected = null;
                          feedback = null;
                          error = null;
                        }
                      }),
                      child: Text(
                        index + 1 == questions.length
                            ? 'Lihat hasil latihan'
                            : 'Soal berikutnya',
                      ),
                    ),
                  ],
                  TextButton(
                    onPressed: busy ? null : () => reset(null),
                    child: const Text('Kembali ke pilihan topik'),
                  ),
                ],
              );
            },
          ),
    );
  }
}
