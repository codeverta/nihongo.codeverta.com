import 'dart:async';

import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';
import 'package:nihongo_master/features/jobs/data/jobs_repository.dart';
import 'package:nihongo_master/features/jobs/domain/job_models.dart';
import 'package:nihongo_master/features/jobs/presentation/job_apply_button.dart';
import 'package:nihongo_master/features/jobs/presentation/jobs_visuals.dart';
import 'package:nihongo_master/shared/presentation/widgets/explore_feature_card.dart';
import 'package:url_launcher/url_launcher.dart';

class JobsEntryCard extends StatelessWidget {
  const JobsEntryCard({super.key});
  @override
  Widget build(BuildContext context) => ExploreFeatureCard(
    title: 'Kerja di Jepang',
    eyebrow: 'LANGKAH BERIKUTNYA UNTUK KARIERMU',
    description:
        'Temukan peluang sesuai bidang, lokasi, gaji, dan kualifikasimu',
    asset: 'assets/images/jobs/job-category-atlas.webp',
    color: const Color(0xFF443060),
    icon: Icons.work_rounded,
    onTap: () => context.push('/jobs'),
  );
}

class JobsScreen extends ConsumerStatefulWidget {
  const JobsScreen({super.key});
  @override
  ConsumerState<JobsScreen> createState() => _JobsScreenState();
}

class _JobsScreenState extends ConsumerState<JobsScreen> {
  final _search = TextEditingController();
  final _scroll = ScrollController();
  Timer? _debounce;
  Map<String, String> _filters = {};
  List<JapanJob> _jobs = [];
  int _total = 0;
  int _generation = 0;
  String? _cursor;
  String? _error;
  bool _loading = true;
  bool _loadingMore = false;
  bool _hasMore = false;

  @override
  void initState() {
    super.initState();
    _scroll.addListener(() {
      if (_scroll.position.extentAfter < 450 &&
          _hasMore &&
          !_loading &&
          !_loadingMore &&
          _error == null) {
        unawaited(_load(more: true));
      }
    });
    unawaited(_load());
  }

  @override
  void dispose() {
    _debounce?.cancel();
    _search.dispose();
    _scroll.dispose();
    super.dispose();
  }

  Future<void> _load({bool more = false}) async {
    if (more && (_loadingMore || !_hasMore)) return;
    final generation = more ? _generation : ++_generation;
    setState(() {
      _error = null;
      if (more) {
        _loadingMore = true;
      } else {
        _loading = true;
        _loadingMore = false;
        _jobs = [];
        _total = 0;
        _cursor = null;
        _hasMore = false;
      }
    });
    try {
      final page = await ref.read(jobsRepositoryProvider).list({
        ..._filters,
        if (_search.text.trim().isNotEmpty) 'q': _search.text.trim(),
      }, cursor: more ? _cursor : null);
      if (!mounted || generation != _generation) return;
      setState(() {
        _jobs = more
            ? {
                for (final j in [..._jobs, ...page.items]) j.id: j,
              }.values.toList()
            : page.items;
        _total = page.total;
        _hasMore =
            page.hasMore &&
            page.nextCursor != null &&
            page.nextCursor != _cursor;
        _cursor = page.nextCursor;
        _loading = false;
        _loadingMore = false;
      });
    } on Object catch (error) {
      if (!mounted || generation != _generation) return;
      setState(() {
        _error = error is JobsFailure
            ? error.message
            : 'Data belum dapat dimuat. Coba kembali.';
        _loading = false;
        _loadingMore = false;
      });
    }
  }

  void _changed(String value) {
    _debounce?.cancel();
    // Invalidate an older response immediately, before the next request starts.
    _generation++;
    setState(() => _loading = true);
    _debounce = Timer(
      const Duration(milliseconds: 450),
      () => unawaited(_load()),
    );
  }

  Future<void> _filter(JobMeta meta) async {
    final result = await showModalBottomSheet<Map<String, String>>(
      context: context,
      isScrollControlled: true,
      useSafeArea: true,
      builder: (context) => _JobFilters(meta: meta, initial: _filters),
    );
    if (result == null || !mounted) return;
    _debounce?.cancel();
    _filters = result;
    unawaited(_load());
  }

  @override
  Widget build(BuildContext context) {
    final meta = ref.watch(jobsMetaProvider);
    return Scaffold(
      appBar: AppBar(
        title: const Text('Kerja di Jepang'),
        actions: [
          IconButton(
            tooltip: 'Sumber lowongan',
            onPressed: () => context.push('/jobs/sources'),
            icon: const Icon(Icons.info_outline),
          ),
        ],
      ),
      body: RefreshIndicator(
        onRefresh: _load,
        child: ListView.builder(
          controller: _scroll,
          physics: const AlwaysScrollableScrollPhysics(),
          padding: const EdgeInsets.fromLTRB(20, 12, 20, 24),
          itemCount: _loading && _jobs.isEmpty ? 2 : _jobs.length + 2,
          itemBuilder: (context, index) {
            if (index == 0) {
              final categories =
                  meta.valueOrNull?.groups['categories'] ?? const [];
              return Padding(
                padding: const EdgeInsets.only(bottom: 14),
                child: Column(
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: [
                    const JobBoardHero(),
                    const SizedBox(height: 16),
                    Row(
                      children: [
                        Expanded(
                          child: TextField(
                            controller: _search,
                            onChanged: _changed,
                            textInputAction: TextInputAction.search,
                            onSubmitted: (_) {
                              _debounce?.cancel();
                              unawaited(_load());
                            },
                            decoration: InputDecoration(
                              hintText: 'Posisi, perusahaan, kata kunci…',
                              prefixIcon: const Icon(Icons.search),
                              suffixIcon: IconButton(
                                tooltip: 'Hapus pencarian',
                                onPressed: () {
                                  _search.clear();
                                  _changed('');
                                },
                                icon: const Icon(Icons.close, size: 18),
                              ),
                            ),
                          ),
                        ),
                        const SizedBox(width: 8),
                        IconButton.filledTonal(
                          tooltip: 'Filter lowongan',
                          onPressed: meta.valueOrNull == null
                              ? null
                              : () => _filter(meta.requireValue),
                          icon: Badge(
                            isLabelVisible: _filters.isNotEmpty,
                            label: Text('${_filters.length}'),
                            child: const Icon(Icons.tune),
                          ),
                        ),
                      ],
                    ),
                    if (meta.hasError)
                      TextButton(
                        onPressed: () => ref.invalidate(jobsMetaProvider),
                        child: const Text('Muat ulang pilihan filter'),
                      ),
                    const SizedBox(height: 8),
                    Row(
                      children: [
                        Expanded(
                          child: Text(
                            _loading ? 'Mencari peluang…' : '$_total lowongan',
                            style: const TextStyle(fontWeight: FontWeight.w700),
                          ),
                        ),
                        if (_filters.isNotEmpty)
                          TextButton(
                            onPressed: () {
                              _filters = {};
                              unawaited(_load());
                            },
                            child: const Text('Reset filter'),
                          ),
                      ],
                    ),
                    const SizedBox(height: 10),
                    if (_error != null && _jobs.isEmpty) ...[
                      _Message(
                        message: _error!,
                        action: 'Coba lagi',
                        onTap: _load,
                      ),
                      const SizedBox(height: 10),
                    ],
                    JobCategoryStrip(
                      categories: categories,
                      selectedId: _filters['category'],
                      onSelected: (category) {
                        _filters = {..._filters};
                        if (category == null) {
                          _filters.remove('category');
                        } else {
                          _filters['category'] = category;
                        }
                        unawaited(_load());
                      },
                    ),
                  ],
                ),
              );
            }
            if (_loading && _jobs.isEmpty) {
              return const Padding(
                padding: EdgeInsets.all(32),
                child: Center(child: CircularProgressIndicator()),
              );
            }
            final jobIndex = index - 1;
            if (jobIndex < _jobs.length) {
              return _JobCard(job: _jobs[jobIndex], meta: meta.valueOrNull);
            }
            if (_error != null) {
              if (_jobs.isEmpty) return const SizedBox.shrink();
              return _Message(
                message: _error!,
                action: 'Coba lagi',
                onTap: () => _load(more: _jobs.isNotEmpty),
              );
            }
            if (_loadingMore) {
              return const Padding(
                padding: EdgeInsets.all(24),
                child: Center(child: CircularProgressIndicator()),
              );
            }
            if (_jobs.isEmpty) {
              return const _Message(
                message:
                    'Belum ada lowongan yang cocok. Ubah pencarian atau filter. Sumber baru juga dapat ditambahkan secara bertahap.',
              );
            }
            if (_hasMore) {
              return TextButton(
                onPressed: () => _load(more: true),
                child: const Text('Muat lebih banyak'),
              );
            }
            return const Padding(
              padding: EdgeInsets.all(16),
              child: Text(
                'Seluruh hasil telah ditampilkan. Data diperbarui berkala dari sumber.',
                textAlign: TextAlign.center,
              ),
            );
          },
        ),
      ),
    );
  }
}

class _JobCard extends StatelessWidget {
  const _JobCard({required this.job, this.meta});
  final JapanJob job;
  final JobMeta? meta;
  @override
  Widget build(BuildContext context) => Card(
    margin: const EdgeInsets.only(bottom: 12),
    child: InkWell(
      borderRadius: BorderRadius.circular(16),
      onTap: () => context.push('/jobs/${Uri.encodeComponent(job.id)}'),
      child: Padding(
        padding: const EdgeInsets.all(18),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            Text(
              job.company.isEmpty ? job.sourceName : job.company,
              style: TextStyle(
                color: Theme.of(context).colorScheme.primary,
                fontWeight: FontWeight.w700,
              ),
            ),
            const SizedBox(height: 7),
            Text(
              job.title,
              style: Theme.of(
                context,
              ).textTheme.titleMedium?.copyWith(fontWeight: FontWeight.w800),
            ),
            const SizedBox(height: 12),
            _IconLine(
              Icons.location_on_outlined,
              job.location.isEmpty
                  ? (meta?.label('prefectures', job.prefecture) ??
                        job.prefecture)
                  : job.location,
            ),
            const SizedBox(height: 6),
            _IconLine(Icons.payments_outlined, job.salary),
            const SizedBox(height: 12),
            Wrap(
              spacing: 6,
              runSpacing: 6,
              children: [
                _Tag(meta?.label('categories', job.category) ?? job.category),
                if (job.employmentType != 'unknown')
                  _Tag(
                    meta?.label('employmentTypes', job.employmentType) ??
                        job.employmentType,
                  ),
                if (!job.active) const _Tag('Tidak aktif'),
              ],
            ),
            const SizedBox(height: 12),
            Text(
              '${job.sourceName} · Diperiksa ${jobDate(job.lastSeenAt)}',
              style: Theme.of(context).textTheme.bodySmall,
            ),
          ],
        ),
      ),
    ),
  );
}

class _JobFilters extends StatefulWidget {
  const _JobFilters({required this.meta, required this.initial});
  final JobMeta meta;
  final Map<String, String> initial;
  @override
  State<_JobFilters> createState() => _JobFiltersState();
}

class _JobFiltersState extends State<_JobFilters> {
  late final Map<String, String> _values = {...widget.initial};
  late final _salary = TextEditingController(text: _values['salaryMin']);
  @override
  void dispose() {
    _salary.dispose();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) => Padding(
    padding: EdgeInsets.only(bottom: MediaQuery.viewInsetsOf(context).bottom),
    child: ListView(
      shrinkWrap: true,
      padding: const EdgeInsets.all(24),
      children: [
        Text(
          'Temukan yang sesuai',
          style: Theme.of(context).textTheme.titleLarge,
        ),
        const SizedBox(height: 16),
        for (final field in const [
          ('category', 'categories', 'Kategori pekerjaan'),
          ('sector', 'sectors', 'Bidang SSW'),
          ('prefecture', 'prefectures', 'Prefektur'),
          ('employmentType', 'employmentTypes', 'Jenis pekerjaan'),
          ('japaneseLevel', 'japaneseLevels', 'Bahasa Jepang'),
          ('visa', 'visaOptions', 'Dukungan visa'),
        ]) ...[
          DropdownButtonFormField<String>(
            initialValue: _values[field.$1],
            isExpanded: true,
            decoration: InputDecoration(labelText: field.$3),
            items: [
              const DropdownMenuItem<String>(child: Text('Semua')),
              ...widget.meta.groups[field.$2]!.map(
                (v) => DropdownMenuItem(
                  value: v.id,
                  child: Text(v.label, overflow: TextOverflow.ellipsis),
                ),
              ),
            ],
            onChanged: (v) => setState(() {
              v == null ? _values.remove(field.$1) : _values[field.$1] = v;
            }),
          ),
          const SizedBox(height: 12),
        ],
        DropdownButtonFormField<String>(
          initialValue: _values['remote'],
          isExpanded: true,
          decoration: const InputDecoration(labelText: 'Cara bekerja'),
          items: const [
            DropdownMenuItem<String>(child: Text('Semua')),
            DropdownMenuItem(value: 'remote', child: Text('Remote')),
            DropdownMenuItem(value: 'hybrid', child: Text('Hybrid')),
            DropdownMenuItem(value: 'onsite', child: Text('Di tempat kerja')),
          ],
          onChanged: (v) {
            v == null ? _values.remove('remote') : _values['remote'] = v;
          },
        ),
        const SizedBox(height: 12),
        TextField(
          controller: _salary,
          keyboardType: TextInputType.number,
          decoration: const InputDecoration(
            labelText: 'Gaji minimum (JPY)',
            hintText: 'Contoh: 250000',
          ),
        ),
        const SizedBox(height: 12),
        DropdownButtonFormField<String>(
          initialValue: _values['salaryPeriod'],
          isExpanded: true,
          decoration: const InputDecoration(labelText: 'Periode gaji'),
          items: const [
            DropdownMenuItem<String>(child: Text('Pilih periode')),
            DropdownMenuItem(value: 'month', child: Text('Per bulan')),
            DropdownMenuItem(value: 'year', child: Text('Per tahun')),
            DropdownMenuItem(value: 'hour', child: Text('Per jam')),
            DropdownMenuItem(value: 'day', child: Text('Per hari')),
            DropdownMenuItem(value: 'week', child: Text('Per minggu')),
          ],
          onChanged: (v) {
            v == null
                ? _values.remove('salaryPeriod')
                : _values['salaryPeriod'] = v;
          },
        ),
        const SizedBox(height: 8),
        const Text(
          'Filter gaji hanya berlaku pada angka dan periode yang dicantumkan sumber.',
        ),
        const SizedBox(height: 20),
        FilledButton(
          onPressed: () {
            final amount = int.tryParse(_salary.text.trim());
            if (_salary.text.trim().isNotEmpty &&
                (amount == null ||
                    amount <= 0 ||
                    !_values.containsKey('salaryPeriod'))) {
              ScaffoldMessenger.of(context).showSnackBar(
                const SnackBar(
                  content: Text(
                    'Masukkan gaji minimum yang valid dan pilih periodenya.',
                  ),
                ),
              );
              return;
            }
            if (amount != null) {
              _values['salaryMin'] = '$amount';
            } else {
              _values
                ..remove('salaryMin')
                ..remove('salaryPeriod');
            }
            Navigator.pop(context, _values);
          },
          child: const Text('Tampilkan lowongan'),
        ),
        TextButton(
          onPressed: () => Navigator.pop(context, <String, String>{}),
          child: const Text('Reset semua filter'),
        ),
      ],
    ),
  );
}

class JobDetailScreen extends ConsumerWidget {
  const JobDetailScreen({required this.id, super.key});
  final String id;
  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final detail = ref.watch(jobDetailProvider(id));
    final meta = ref.watch(jobsMetaProvider).valueOrNull;
    return Scaffold(
      appBar: AppBar(title: const Text('Detail lowongan')),
      body: detail.when(
        loading: () => const Center(child: CircularProgressIndicator()),
        error: (e, _) => _Message(
          message: '$e',
          action: 'Coba lagi',
          onTap: () => ref.invalidate(jobDetailProvider(id)),
        ),
        data: (job) => ListView(
          padding: const EdgeInsets.all(24),
          children: [
            Text(
              job.company,
              style: TextStyle(
                color: Theme.of(context).colorScheme.primary,
                fontWeight: FontWeight.w700,
              ),
            ),
            const SizedBox(height: 8),
            Text(
              job.title,
              style: Theme.of(
                context,
              ).textTheme.headlineSmall?.copyWith(fontWeight: FontWeight.w800),
            ),
            const SizedBox(height: 16),
            if (!job.active)
              const Card(
                child: Padding(
                  padding: EdgeInsets.all(16),
                  child: Text(
                    'Lowongan ini sudah tidak aktif atau kedaluwarsa. Lamaran tidak tersedia.',
                  ),
                ),
              ),
            _IconLine(
              Icons.location_on_outlined,
              job.location.isEmpty ? 'Lokasi belum dicantumkan' : job.location,
            ),
            const SizedBox(height: 10),
            _IconLine(Icons.payments_outlined, job.salary),
            const SizedBox(height: 20),
            _DetailRow(
              'Bidang',
              meta?.label('categories', job.category) ?? job.category,
            ),
            _DetailRow(
              'Sektor',
              meta?.label('sectors', job.sector) ?? job.sector,
            ),
            _DetailRow(
              'Jenis pekerjaan',
              meta?.label('employmentTypes', job.employmentType) ??
                  job.employmentType,
            ),
            _DetailRow(
              'Bahasa Jepang',
              meta?.label('japaneseLevels', job.japaneseLevel) ??
                  job.japaneseLevel,
            ),
            _DetailRow(
              'Visa',
              meta?.label('visaOptions', job.visa) ?? job.visa,
            ),
            _DetailRow('Cara bekerja', switch (job.remote) {
              'remote' => 'Remote',
              'onsite' => 'Di tempat kerja',
              'hybrid' => 'Hybrid',
              _ => 'Belum dicantumkan',
            }),
            _Section('Tentang pekerjaan', job.description),
            _Section('Kualifikasi', job.qualifications),
            _Section(
              'Benefit & tunjangan',
              job.benefits.isEmpty
                  ? 'Benefit belum dicantumkan oleh sumber lowongan. Konfirmasikan tunjangan, fasilitas, dan ketentuannya kepada perekrut sebelum melamar.'
                  : job.benefits,
            ),
            const Divider(height: 32),
            _DetailRow('Sumber', job.sourceName),
            _DetailRow('Tanggal publikasi', jobDate(job.postedAt)),
            _DetailRow('Diperiksa oleh sistem', jobDate(job.lastSeenAt)),
            _DetailRow('Batas lowongan', jobDate(job.expiresAt)),
            const SizedBox(height: 16),
            const Text(
              'Informasi mengikuti sumber aslinya dan dapat berubah. Baca persyaratan lengkap di situs penyedia sebelum melamar.',
            ),
            const SizedBox(height: 20),
            Container(
              padding: const EdgeInsets.all(16),
              decoration: BoxDecoration(
                color: const Color(0xFFF1F3FC),
                border: Border.all(color: const Color(0xFFDDE2F4)),
                borderRadius: BorderRadius.circular(18),
              ),
              child: const Row(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  Icon(
                    Icons.workspace_premium_rounded,
                    color: Color(0xFF232667),
                  ),
                  SizedBox(width: 12),
                  Expanded(
                    child: Column(
                      crossAxisAlignment: CrossAxisAlignment.start,
                      children: [
                        Text(
                          'Lamar dengan Lifetime Access',
                          style: TextStyle(
                            fontWeight: FontWeight.w800,
                            color: Color(0xFF232667),
                          ),
                        ),
                        SizedBox(height: 5),
                        Text(
                          'Akses tautan lamaran untuk lowongan yang aktif, sekaligus buka materi lanjutan, simulasi ujian, dan Manga Jepang.',
                          style: TextStyle(
                            color: Color(0xFF44496B),
                            height: 1.5,
                          ),
                        ),
                      ],
                    ),
                  ),
                ],
              ),
            ),
            const SizedBox(height: 12),
            JobApplyButton(job: job),
            const SizedBox(height: 8),
            Text(
              'Lamaran diproses oleh situs sumber. Premium tidak menjamin diterima kerja atau persetujuan visa.',
              style: Theme.of(context).textTheme.bodySmall,
            ),
          ],
        ),
      ),
    );
  }
}

class JobSourcesScreen extends ConsumerWidget {
  const JobSourcesScreen({super.key});
  @override
  Widget build(BuildContext context, WidgetRef ref) => Scaffold(
    appBar: AppBar(title: const Text('Sumber & pembaruan')),
    body: ref
        .watch(jobSourcesProvider)
        .when(
          loading: () => const Center(child: CircularProgressIndicator()),
          error: (e, _) => _Message(
            message: '$e',
            action: 'Coba lagi',
            onTap: () => ref.invalidate(jobSourcesProvider),
          ),
          data: (sources) => ListView(
            padding: const EdgeInsets.all(20),
            children: [
              const Text(
                'Lowongan disimpan di server kami. Membuka atau mencari di aplikasi tidak mengambil ulang data dari situs penyedia.',
              ),
              const SizedBox(height: 16),
              for (final s in sources)
                Card(
                  child: Padding(
                    padding: const EdgeInsets.all(18),
                    child: Column(
                      crossAxisAlignment: CrossAxisAlignment.start,
                      children: [
                        Text(
                          '${s['name']}',
                          style: Theme.of(context).textTheme.titleMedium
                              ?.copyWith(fontWeight: FontWeight.w800),
                        ),
                        const SizedBox(height: 8),
                        Text(
                          s['mode'] == 'directory'
                              ? 'Direktori · Kunjungi situs untuk lowongan lengkap'
                              : 'Pembaruan otomatis · ${_sourceStatus(s['status'])}',
                        ),
                        if (s['reason'] != null) Text('${s['reason']}'),
                        if (s['mode'] != 'directory') ...[
                          Text(
                            'Interval: ${_sourceInterval(s['intervalHours'])}',
                          ),
                          Text(
                            'Berhasil diperbarui: ${jobDate(DateTime.tryParse('${s['lastSuccessAt'] ?? ''}'))}',
                          ),
                          Text(
                            'Jadwal berikutnya: ${jobDate(DateTime.tryParse('${s['nextFetchAt'] ?? ''}'))}',
                          ),
                        ],
                        Wrap(
                          spacing: 8,
                          children: [
                            if (jobExternalUri('${s['url']}') != null)
                              TextButton(
                                onPressed: () => _open(context, '${s['url']}'),
                                child: const Text('Kunjungi situs'),
                              ),
                            if (jobExternalUri('${s['termsUrl'] ?? ''}') !=
                                null)
                              TextButton(
                                onPressed: () =>
                                    _open(context, '${s['termsUrl']}'),
                                child: const Text('Ketentuan sumber'),
                              ),
                          ],
                        ),
                      ],
                    ),
                  ),
                ),
            ],
          ),
        ),
  );
}

Future<void> _open(BuildContext context, String value) async {
  final uri = jobExternalUri(value);
  if (uri == null) return;
  var opened = false;
  try {
    opened = await launchUrl(uri, mode: LaunchMode.externalApplication);
  } on Object {
    opened = false;
  }
  if (!opened && context.mounted) {
    ScaffoldMessenger.of(context).showSnackBar(
      const SnackBar(content: Text('Tautan belum dapat dibuka. Coba kembali.')),
    );
  }
}

String _sourceStatus(dynamic status) => switch (status) {
  'unsupported' => 'Data publik belum dapat dibaca',
  'ok' => 'Terbaru',
  'error' => 'Pembaruan tertunda',
  'blocked' => 'Pengambilan dihentikan',
  'partial' => 'Sampel lowongan diperbarui',
  _ => 'Menunggu pembaruan',
};

String _sourceInterval(dynamic value) {
  final hours = value is num ? value.toInt() : int.tryParse('$value');
  if (hours == null) return 'Belum dijadwalkan';
  return hours % 24 == 0 ? '${hours ~/ 24} hari' : '$hours jam';
}

class _IconLine extends StatelessWidget {
  const _IconLine(this.icon, this.text);
  final IconData icon;
  final String text;
  @override
  Widget build(BuildContext context) => Row(
    crossAxisAlignment: CrossAxisAlignment.start,
    children: [
      Icon(icon, size: 18),
      const SizedBox(width: 7),
      Expanded(child: Text(text)),
    ],
  );
}

class _Tag extends StatelessWidget {
  const _Tag(this.text);
  final String text;
  @override
  Widget build(BuildContext context) => Container(
    padding: const EdgeInsets.symmetric(horizontal: 9, vertical: 5),
    decoration: BoxDecoration(
      color: Theme.of(context).colorScheme.surfaceContainerHighest,
      borderRadius: BorderRadius.circular(8),
    ),
    child: Text(text, style: Theme.of(context).textTheme.bodySmall),
  );
}

class _DetailRow extends StatelessWidget {
  const _DetailRow(this.label, this.value);
  final String label;
  final String value;
  @override
  Widget build(BuildContext context) => Padding(
    padding: const EdgeInsets.symmetric(vertical: 6),
    child: Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        Text(label, style: Theme.of(context).textTheme.bodySmall),
        Text(
          value.isEmpty || value == 'unknown' ? 'Belum dicantumkan' : value,
          style: const TextStyle(fontWeight: FontWeight.w600),
        ),
      ],
    ),
  );
}

class _Section extends StatelessWidget {
  const _Section(this.title, this.text);
  final String title;
  final String text;
  @override
  Widget build(BuildContext context) => Padding(
    padding: const EdgeInsets.only(top: 24),
    child: Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        Text(
          title,
          style: Theme.of(
            context,
          ).textTheme.titleMedium?.copyWith(fontWeight: FontWeight.w800),
        ),
        const SizedBox(height: 10),
        SelectableText(
          text.trim().isEmpty
              ? 'Lihat informasi lengkap di situs sumber.'
              : text,
          scrollPhysics: const NeverScrollableScrollPhysics(),
        ),
      ],
    ),
  );
}

class _Message extends StatelessWidget {
  const _Message({required this.message, this.action, this.onTap});
  final String message;
  final String? action;
  final VoidCallback? onTap;
  @override
  Widget build(BuildContext context) => Padding(
    padding: const EdgeInsets.all(24),
    child: Column(
      mainAxisSize: MainAxisSize.min,
      children: [
        Text(message, textAlign: TextAlign.center),
        if (onTap != null) TextButton(onPressed: onTap, child: Text(action!)),
      ],
    ),
  );
}
