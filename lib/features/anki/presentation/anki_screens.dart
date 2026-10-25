import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';
import 'package:nihongo_master/core/localization/app_localizations.dart';
import 'package:nihongo_master/core/localization/region_formatters.dart';
import 'package:nihongo_master/features/anki/data/anki_repository.dart';
import 'package:nihongo_master/features/anki/presentation/anki_controller.dart';
import 'package:nihongo_master/features/anki/presentation/anki_deck_media.dart';
import 'package:nihongo_master/features/anki/presentation/anki_offline_preparation.dart';
import 'package:nihongo_master/features/anki/presentation/anki_transfer_actions.dart';
import 'package:nihongo_master/features/auth/presentation/controllers/auth_controller.dart';
import 'package:nihongo_master/shared/presentation/widgets/explore_feature_card.dart';
import 'package:url_launcher/url_launcher.dart';

String ankiDeckPath(String id) => '/anki/decks/${Uri.encodeComponent(id)}';
String ankiStudyPath(String id) => '${ankiDeckPath(id)}/study';
String ankiCount(int value) => value.toString().replaceAllMapped(
  RegExp(r'(\d)(?=(\d{3})+(?!\d))'),
  (m) => '${m[1]}.',
);

class AnkiEntryCard extends StatelessWidget {
  const AnkiEntryCard({super.key});
  @override
  Widget build(BuildContext context) => ExploreFeatureCard(
    title: 'Anki · Flashcards',
    eyebrow: 'REVIEW SEDIKIT, INGAT LEBIH LAMA'.localized(context),
    description: 'Pilih deck favorit dan bangun kebiasaan review kosakata'
        .localized(context),
    asset: 'assets/images/explore/anki-flashcards.webp',
    color: const Color(0xFF145B60),
    icon: Icons.style_rounded,
    onTap: () => context.push('/anki'),
  );
}

class AnkiDeckListScreen extends ConsumerStatefulWidget {
  const AnkiDeckListScreen({super.key});
  @override
  ConsumerState<AnkiDeckListScreen> createState() => _DeckListState();
}

class _DeckListState extends ConsumerState<AnkiDeckListScreen> {
  String query = '';
  String category = 'Semua';
  @override
  Widget build(BuildContext context) {
    final colors = Theme.of(context).colorScheme;
    return Scaffold(
      appBar: AppBar(
        title: const Text('Anki'),
        actions: [
          IconButton(
            key: const Key('anki-import'),
            tooltip: 'Import deck'.localized(context),
            onPressed: () async {
              if (!ref.read(authControllerProvider).isAuthenticated) {
                await context.push('/login');
                return;
              }
              final id = await context.push<String>('/anki/import');
              if (id != null && context.mounted) {
                ref.invalidate(ankiDecksProvider);
                await context.push(ankiDeckPath(id));
              }
            },
            icon: const Icon(Icons.file_upload_outlined),
          ),
          IconButton(
            tooltip: 'Sumber data & lisensi'.localized(context),
            onPressed: () => _credits(context),
            icon: const Icon(Icons.info_outline_rounded),
          ),
        ],
      ),
      body: ref
          .watch(ankiDecksProvider)
          .when(
            loading: () => const Center(child: CircularProgressIndicator()),
            error: (error, _) => AnkiErrorView(
              error: error,
              onRetry: () => ref.invalidate(ankiDecksProvider),
            ),
            data: (decks) {
              final categories = [
                'Semua',
                ...decks.map((d) => d.category).toSet().toList()..sort(),
              ];
              final filtered = decks
                  .where(
                    (d) =>
                        (category == 'Semua' || d.category == category) &&
                        '${d.title} ${d.id} ${d.description} ${d.originalName}'
                            .toLowerCase()
                            .contains(query.trim().toLowerCase()),
                  )
                  .toList();
              const preferred = [
                'n5',
                'hiragana',
                'katakana',
                'n4',
                'wakayama',
              ];
              filtered.sort((a, b) {
                final ai = preferred.indexOf(a.id);
                final bi = preferred.indexOf(b.id);
                if (ai >= 0 || bi >= 0) {
                  return (ai < 0 ? 999 : ai).compareTo(bi < 0 ? 999 : bi);
                }
                return a.title.compareTo(b.title);
              });
              final premium = ref.watch(
                authControllerProvider.select(
                  (s) => s.user?.isPremium ?? false,
                ),
              );
              return CustomScrollView(
                slivers: [
                  SliverPadding(
                    padding: const EdgeInsets.fromLTRB(20, 12, 20, 0),
                    sliver: SliverList.list(
                      children: [
                        Container(
                          padding: const EdgeInsets.all(22),
                          decoration: BoxDecoration(
                            color: colors.secondary.withValues(alpha: .08),
                            borderRadius: BorderRadius.circular(24),
                          ),
                          child: Column(
                            crossAxisAlignment: CrossAxisAlignment.start,
                            children: [
                              Icon(
                                Icons.style_rounded,
                                color: colors.secondary,
                                size: 38,
                              ),
                              const SizedBox(height: 16),
                              Text(
                                'Sedikit hari ini.\nLebih ingat besok.'
                                    .localized(context),
                                style: Theme.of(context).textTheme.headlineSmall
                                    ?.copyWith(fontWeight: FontWeight.w900),
                              ),
                              const SizedBox(height: 12),
                              Text(
                                'Kenali topik deck, balik kartu untuk melihat jawaban, lalu nilai seberapa mudah kamu mengingatnya.'
                                    .localized(context),
                                style: const TextStyle(height: 1.6),
                              ),
                              const SizedBox(height: 16),
                              Wrap(
                                spacing: 8,
                                runSpacing: 8,
                                children: [
                                  _Tag(
                                    '${decks.length} deck'.localized(context),
                                  ),
                                  _Tag(
                                    '${ankiCount(decks.fold<int>(0, (n, d) => n + d.cardCount))} kartu'
                                        .localized(context),
                                  ),
                                ],
                              ),
                            ],
                          ),
                        ),
                        const SizedBox(height: 22),
                        TextField(
                          key: const Key('anki-search'),
                          onChanged: (value) => setState(() => query = value),
                          decoration: InputDecoration(
                            prefixIcon: const Icon(Icons.search_rounded),
                            labelText: 'Cari deck'.localized(context),
                            hintText: 'N5, Wakayama, hiragana...'.localized(
                              context,
                            ),
                          ),
                        ),
                        const SizedBox(height: 14),
                        SingleChildScrollView(
                          scrollDirection: Axis.horizontal,
                          child: Row(
                            children: [
                              for (final value in categories)
                                Padding(
                                  padding: const EdgeInsets.only(right: 8),
                                  child: ChoiceChip(
                                    selected: category == value,
                                    label: Text(value.localized(context)),
                                    onSelected: (_) =>
                                        setState(() => category = value),
                                  ),
                                ),
                            ],
                          ),
                        ),
                        const SizedBox(height: 20),
                        Text(
                          '${filtered.length} deck'.localized(context),
                          style: const TextStyle(fontWeight: FontWeight.w800),
                        ),
                        const SizedBox(height: 10),
                        if (filtered.isEmpty)
                          Padding(
                            padding: const EdgeInsets.symmetric(vertical: 32),
                            child: Text(
                              'Deck tidak ditemukan. Coba nama atau kategori lain.'
                                  .localized(context),
                            ),
                          ),
                      ],
                    ),
                  ),
                  SliverPadding(
                    padding: const EdgeInsets.fromLTRB(20, 0, 20, 28),
                    sliver: SliverList.builder(
                      itemCount: filtered.length,
                      itemBuilder: (context, index) {
                        final deck = filtered[index];
                        return Padding(
                          padding: const EdgeInsets.only(bottom: 12),
                          child: Card(
                            child: ListTile(
                              key: Key('anki-deck-${deck.id}'),
                              contentPadding: const EdgeInsets.all(16),
                              leading: _DeckIcon(category: deck.category),
                              title: Text(
                                deck.title.localized(context),
                                style: const TextStyle(
                                  fontWeight: FontWeight.w800,
                                ),
                              ),
                              subtitle: Padding(
                                padding: const EdgeInsets.only(top: 8),
                                child: Column(
                                  crossAxisAlignment: CrossAxisAlignment.start,
                                  children: [
                                    Text(
                                      '${deck.category} · ${ankiCount(deck.available ? deck.cardCount : deck.sourceCount)} kartu'
                                          .localized(context),
                                    ),
                                    const SizedBox(height: 8),
                                    if (!deck.available)
                                      _Tag(
                                        'Data tersimpan · Belum tersedia'
                                            .localized(context),
                                      )
                                    else if (deck.premium && !premium)
                                      const _Tag('Lifetime Access')
                                    else
                                      _Tag(
                                        'Siap dipelajari'.localized(context),
                                      ),
                                  ],
                                ),
                              ),
                              trailing: const Icon(Icons.chevron_right_rounded),
                              onTap: () => context.push(ankiDeckPath(deck.id)),
                            ),
                          ),
                        );
                      },
                    ),
                  ),
                ],
              );
            },
          ),
    );
  }
}

class AnkiDeckDetailScreen extends ConsumerWidget {
  const AnkiDeckDetailScreen({required this.deckId, super.key});
  final String deckId;
  @override
  Widget build(BuildContext context, WidgetRef ref) => Scaffold(
    appBar: AppBar(title: Text('Tentang deck'.localized(context))),
    body: ref
        .watch(ankiDeckProvider(deckId))
        .when(
          loading: () => const Center(child: CircularProgressIndicator()),
          error: (error, _) => AnkiErrorView(
            error: error,
            onRetry: () => ref.invalidate(ankiDeckProvider(deckId)),
          ),
          data: (deck) {
            final signedIn = ref.watch(
              authControllerProvider.select((s) => s.isAuthenticated),
            );
            return ListView(
              padding: const EdgeInsets.all(20),
              children: [
                if (signedIn && !deck.locked)
                  AnkiOfflinePreparation(deckId: deckId, level: deck.level),
                Align(
                  alignment: Alignment.centerLeft,
                  child: _DeckIcon(category: deck.category, size: 72),
                ),
                const SizedBox(height: 20),
                Text(
                  deck.title.localized(context),
                  style: Theme.of(context).textTheme.headlineSmall?.copyWith(
                    fontWeight: FontWeight.w900,
                  ),
                ),
                const SizedBox(height: 10),
                Wrap(
                  spacing: 8,
                  runSpacing: 8,
                  children: [
                    _Tag(deck.category.localized(context)),
                    if (deck.personal) _Tag('Deck pribadi'.localized(context)),
                    _Tag(
                      '${ankiCount(deck.cardCount)} kartu siap'.localized(
                        context,
                      ),
                    ),
                    if (deck.premium) const _Tag('Lifetime Access'),
                  ],
                ),
                const SizedBox(height: 24),
                if (deck.available && !deck.locked)
                  AnkiExportButton(deckId: deck.id),
                if (deck.personal)
                  TextButton.icon(
                    onPressed: () => deleteAnkiDeck(context, ref, deck.id),
                    icon: const Icon(Icons.delete_outline_rounded),
                    label: Text('Hapus deck pribadi'.localized(context)),
                  ),
                const SizedBox(height: 16),
                _Section(
                  title: 'Deck ini tentang apa?'.localized(context),
                  children: [
                    Text(
                      deck.description.localized(context),
                      style: const TextStyle(height: 1.7),
                    ),
                  ],
                ),
                const SizedBox(height: 16),
                _Section(
                  title: 'Cara belajar'.localized(context),
                  children: [
                    Text(
                      deck.instructions.localized(context),
                      style: const TextStyle(height: 1.7),
                    ),
                    const SizedBox(height: 10),
                    Text(
                      'Satu sesi berisi hingga 20 kartu. Review yang sudah jatuh tempo didahulukan, kemudian kartu baru. Pilihanmu menentukan jadwal berikutnya.'
                          .localized(context),
                      style: const TextStyle(height: 1.7),
                    ),
                  ],
                ),
                const SizedBox(height: 16),
                if (deck.available)
                  _Section(
                    title: 'Progress belajarmu'.localized(context),
                    children: [
                      Wrap(
                        spacing: 20,
                        runSpacing: 16,
                        children: [
                          _Stat(
                            ankiCount(deck.seen),
                            'Pernah dipelajari'.localized(context),
                          ),
                          _Stat(
                            ankiCount(deck.due),
                            'Perlu review'.localized(context),
                          ),
                          _Stat(
                            ankiCount(deck.newCards),
                            'Kartu baru'.localized(context),
                          ),
                        ],
                      ),
                      const SizedBox(height: 14),
                      LinearProgressIndicator(
                        backgroundColor: Theme.of(
                          context,
                        ).colorScheme.primary.withValues(alpha: .1),
                        color: Theme.of(context).colorScheme.primary,
                        value: deck.cardCount == 0
                            ? 0
                            : deck.seen / deck.cardCount,
                        minHeight: 6,
                        borderRadius: BorderRadius.circular(8),
                      ),
                      const SizedBox(height: 12),
                      Text(
                        signedIn
                            ? 'Jadwal dan progress tersimpan di akunmu.'
                                  .localized(context)
                            : 'Masuk untuk menyimpan progress dan jadwal review.'
                                  .localized(context),
                      ),
                      if (deck.nextDue != null && deck.due == 0) ...[
                        const SizedBox(height: 8),
                        Text(
                          'Review berikutnya: ${_date(context, deck.nextDue!)}'
                              .localized(context),
                        ),
                      ],
                    ],
                  ),
                const SizedBox(height: 20),
                if (!deck.available)
                  _Section(
                    title: 'Ketersediaan deck'.localized(context),
                    children: [
                      Text(
                        deck.unavailableReason.localized(context),
                        style: const TextStyle(height: 1.7),
                      ),
                    ],
                  )
                else
                  FilledButton.icon(
                    key: const Key('anki-start'),
                    onPressed: () async {
                      if (!signedIn) {
                        await context.push<bool>('/login');
                        return;
                      }
                      if (deck.locked) {
                        await context.push('/premium');
                        return;
                      }
                      await context.push(ankiStudyPath(deck.id));
                      if (context.mounted) {
                        ref.invalidate(ankiDeckProvider(deckId));
                      }
                    },
                    icon: Icon(
                      !signedIn
                          ? Icons.login_rounded
                          : deck.locked
                          ? Icons.lock_outline_rounded
                          : Icons.play_arrow_rounded,
                    ),
                    label: Text(
                      !signedIn
                          ? 'Masuk untuk belajar'.localized(context)
                          : deck.locked
                          ? 'Buka dengan Lifetime Access'.localized(context)
                          : deck.due > 0
                          ? 'Mulai review'.localized(context)
                          : 'Mulai belajar'.localized(context),
                    ),
                  ),
                if (deck.preview.isNotEmpty) ...[
                  const SizedBox(height: 26),
                  Text(
                    'Contoh isi deck'.localized(context),
                    style: Theme.of(context).textTheme.titleLarge?.copyWith(
                      fontWeight: FontWeight.w800,
                    ),
                  ),
                  const SizedBox(height: 8),
                  Text(
                    'Preview menampilkan pertanyaan dan jawabannya.'.localized(
                      context,
                    ),
                  ),
                  const SizedBox(height: 12),
                  for (final card in deck.preview)
                    Padding(
                      padding: const EdgeInsets.only(bottom: 12),
                      child: _Section(
                        title: card.front,
                        children: [
                          if (card.imageUrl.isNotEmpty)
                            AnkiDeckImage(card.imageUrl),
                          if (card.audioUrl.isNotEmpty)
                            AnkiDeckAudio(card.audioUrl),
                          _Answers(answers: card.answers),
                          if (card.meaning.isNotEmpty) ...[
                            const SizedBox(height: 8),
                            Text(
                              card.meaning,
                              style: const TextStyle(height: 1.6),
                            ),
                          ],
                        ],
                      ),
                    ),
                ],
                const SizedBox(height: 18),
                _Section(
                  title: 'Asal data'.localized(context),
                  children: [
                    Text(
                      'Kotoba · ${deck.originalName}',
                      style: const TextStyle(fontWeight: FontWeight.w700),
                    ),
                    const SizedBox(height: 10),
                    Text(
                      deck.meaningLanguage.localized(context),
                      style: const TextStyle(height: 1.6),
                    ),
                    if (deck.originalDescription.isNotEmpty) ...[
                      const SizedBox(height: 10),
                      Text(
                        deck.originalDescription,
                        style: const TextStyle(height: 1.6),
                      ),
                    ],
                    if (deck.skippedCount > 0) ...[
                      const SizedBox(height: 10),
                      Text(
                        '${ankiCount(deck.sourceCount)} entri sumber disimpan. ${ankiCount(deck.skippedCount)} entri belum dapat digunakan sebagai flashcard karena data kosong atau memerlukan mesin tambahan.'
                            .localized(context),
                        style: const TextStyle(height: 1.6),
                      ),
                    ],
                    if (const [
                      'http',
                      'https',
                    ].contains(Uri.tryParse(deck.sourceUrl)?.scheme))
                      TextButton.icon(
                        onPressed: () async {
                          var opened = false;
                          try {
                            opened = await launchUrl(
                              Uri.parse(deck.sourceUrl),
                              mode: LaunchMode.externalApplication,
                            );
                          } on Object {
                            opened = false;
                          }
                          if (!opened && context.mounted) {
                            ScaffoldMessenger.of(context).showSnackBar(
                              SnackBar(
                                content: Text(
                                  'Tautan sumber belum dapat dibuka.'.localized(
                                    context,
                                  ),
                                ),
                              ),
                            );
                          }
                        },
                        icon: const Icon(Icons.open_in_new_rounded),
                        label: Text(
                          'Lihat data sumber deck'.localized(context),
                        ),
                      ),
                    TextButton.icon(
                      onPressed: () => _credits(context),
                      icon: const Icon(Icons.info_outline_rounded),
                      label: Text('Sumber data & lisensi'.localized(context)),
                    ),
                  ],
                ),
                const SizedBox(height: 24),
              ],
            );
          },
        ),
  );
}

class AnkiStudyScreen extends ConsumerWidget {
  const AnkiStudyScreen({required this.deckId, super.key});
  final String deckId;
  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final state = ref.watch(ankiStudyProvider(deckId));
    final controller = ref.read(ankiStudyProvider(deckId).notifier);
    final pendingSync = ref.watch(ankiRepositoryProvider).hasPendingReviews;
    final card = state.current;
    final colors = Theme.of(context).colorScheme;
    return Scaffold(
      appBar: AppBar(
        title: Text(
          state.session?.title ?? 'Belajar flashcard'.localized(context),
          maxLines: 1,
          overflow: TextOverflow.ellipsis,
        ),
      ),
      body: state.loading
          ? const Center(child: CircularProgressIndicator())
          : state.session == null
          ? AnkiErrorView(
              error: state.error ?? 'Sesi belum tersedia.'.localized(context),
              onRetry: controller.load,
            )
          : card == null
          ? ListView(
              padding: const EdgeInsets.all(24),
              children: [
                Icon(Icons.task_alt_rounded, size: 72, color: colors.primary),
                const SizedBox(height: 24),
                Text(
                  state.position == 0
                      ? 'Belum ada kartu yang perlu dipelajari.'.localized(
                          context,
                        )
                      : 'Sesi selesai!'.localized(context),
                  textAlign: TextAlign.center,
                  style: Theme.of(context).textTheme.headlineSmall?.copyWith(
                    fontWeight: FontWeight.w800,
                  ),
                ),
                const SizedBox(height: 16),
                Text(
                  (pendingSync
                          ? '${state.position} kartu telah kamu nilai. Penilaian tersimpan di perangkat dan menunggu sinkronisasi. Jadwal resmi diperbarui setelah terhubung.'
                          : '${state.position} kartu telah kamu nilai. Jadwal berikutnya tersimpan di akunmu. Kartu yang belum hafal dijadwalkan lagi sekitar 1 menit.')
                      .localized(context),
                  textAlign: TextAlign.center,
                  style: const TextStyle(height: 1.7),
                ),
                const SizedBox(height: 24),
                FilledButton(
                  onPressed: controller.load,
                  child: Text('Periksa kartu berikutnya'.localized(context)),
                ),
                const SizedBox(height: 10),
                OutlinedButton(
                  onPressed: () => context.go(ankiDeckPath(deckId)),
                  child: Text('Kembali ke detail deck'.localized(context)),
                ),
              ],
            )
          : ListView(
              padding: const EdgeInsets.all(20),
              children: [
                if (pendingSync) ...[
                  Text(
                    'Penilaian tersimpan di perangkat · menunggu sinkronisasi'
                        .localized(context),
                  ),
                  const SizedBox(height: 12),
                ],
                Text(
                  'KARTU ${state.position + 1} / ${state.session!.cards.length}'
                      .localized(context),
                  style: TextStyle(
                    color: colors.primary,
                    fontWeight: FontWeight.w800,
                    letterSpacing: 1,
                  ),
                ),
                const SizedBox(height: 12),
                LinearProgressIndicator(
                  backgroundColor: colors.primary.withValues(alpha: .1),
                  color: colors.primary,
                  value: state.position / state.session!.cards.length,
                  minHeight: 6,
                  borderRadius: BorderRadius.circular(8),
                ),
                if (state.session!.instructions.isNotEmpty) ...[
                  const SizedBox(height: 16),
                  Text(
                    state.session!.instructions,
                    style: const TextStyle(height: 1.6),
                  ),
                ],
                const SizedBox(height: 24),
                Card(
                  child: Padding(
                    padding: const EdgeInsets.symmetric(
                      horizontal: 22,
                      vertical: 32,
                    ),
                    child: Column(
                      crossAxisAlignment: CrossAxisAlignment.stretch,
                      children: [
                        Text(
                          state.revealed
                              ? 'PERTANYAAN & JAWABAN'.localized(context)
                              : 'INGAT DULU JAWABANNYA'.localized(context),
                          textAlign: TextAlign.center,
                          style: TextStyle(
                            color: colors.onSurfaceVariant,
                            fontWeight: FontWeight.w600,
                            fontSize: 12,
                            letterSpacing: .8,
                          ),
                        ),
                        const SizedBox(height: 24),
                        if (card.imageUrl.isNotEmpty)
                          AnkiDeckImage(card.imageUrl),
                        if (card.audioUrl.isNotEmpty)
                          AnkiDeckAudio(card.audioUrl),
                        SelectableText(
                          card.front,
                          key: const Key('anki-card-front'),
                          textAlign: TextAlign.center,
                          scrollPhysics: const NeverScrollableScrollPhysics(),
                          style: TextStyle(
                            fontSize: card.front.runes.length > 40 ? 22 : 36,
                            fontWeight: FontWeight.w800,
                            height: 1.6,
                          ),
                        ),
                        if (state.revealed) ...[
                          const Padding(
                            padding: EdgeInsets.symmetric(vertical: 24),
                            child: Divider(),
                          ),
                          _Answers(
                            answers: card.answers,
                            key: const Key('anki-card-back'),
                          ),
                          if (card.meaning.isNotEmpty) ...[
                            const SizedBox(height: 20),
                            Text(
                              card.meaning,
                              textAlign: TextAlign.center,
                              style: const TextStyle(height: 1.7, fontSize: 16),
                            ),
                          ],
                        ],
                      ],
                    ),
                  ),
                ),
                const SizedBox(height: 24),
                if (!state.revealed)
                  FilledButton.icon(
                    key: const Key('anki-reveal'),
                    onPressed: controller.reveal,
                    icon: const Icon(Icons.flip_rounded),
                    label: Text('Tampilkan jawaban'.localized(context)),
                  ),
                if (state.revealed) ...[
                  Text(
                    'Seberapa mudah kamu mengingatnya?'.localized(context),
                    style: const TextStyle(fontWeight: FontWeight.w800),
                  ),
                  const SizedBox(height: 12),
                  if (state.error != null) ...[
                    Text(
                      state.error!.message.localized(context),
                      style: TextStyle(color: colors.error, height: 1.6),
                    ),
                    const SizedBox(height: 12),
                    if (state.error!.status == 409 ||
                        state.error!.status == 401 ||
                        state.error!.status == 402)
                      AnkiErrorView(
                        error: state.error!,
                        onRetry: controller.load,
                      )
                    else
                      FilledButton(
                        onPressed: state.saving
                            ? null
                            : () => controller.rate(state.pendingRating!),
                        child: Text('Coba simpan kembali'.localized(context)),
                      ),
                  ] else
                    LayoutBuilder(
                      builder: (context, constraints) => Wrap(
                        spacing: 10,
                        runSpacing: 10,
                        children: [
                          for (final grade in [
                            ('again', 'Belum hafal'.localized(context)),
                            ('hard', 'Sulit'.localized(context)),
                            ('good', 'Ingat'.localized(context)),
                            ('easy', 'Mudah'.localized(context)),
                          ])
                            SizedBox(
                              width: (constraints.maxWidth - 10) / 2,
                              child: OutlinedButton(
                                key: Key('anki-rate-${grade.$1}'),
                                onPressed: state.saving
                                    ? null
                                    : () => controller.rate(grade.$1),
                                style: OutlinedButton.styleFrom(
                                  padding: const EdgeInsets.symmetric(
                                    horizontal: 10,
                                    vertical: 16,
                                  ),
                                ),
                                child: Column(
                                  children: [
                                    Text(
                                      grade.$2,
                                      style: const TextStyle(
                                        fontWeight: FontWeight.w800,
                                      ),
                                    ),
                                    const SizedBox(height: 5),
                                    Text(
                                      card.intervals[grade.$1] ?? '',
                                      style: const TextStyle(fontSize: 12),
                                    ),
                                  ],
                                ),
                              ),
                            ),
                        ],
                      ),
                    ),
                  if (state.saving)
                    const Padding(
                      padding: EdgeInsets.all(16),
                      child: Center(child: CircularProgressIndicator()),
                    ),
                ],
                const SizedBox(height: 24),
                Text(
                  'Nilai berdasarkan ingatanmu sebelum membuka jawaban. Kamu bisa keluar kapan saja; setiap penilaian yang berhasil sudah tersimpan.'
                      .localized(context),
                  style: TextStyle(
                    height: 1.6,
                    fontSize: 13,
                    color: colors.onSurfaceVariant,
                  ),
                ),
              ],
            ),
    );
  }
}

class AnkiErrorView extends StatelessWidget {
  const AnkiErrorView({required this.error, required this.onRetry, super.key});
  final Object error;
  final VoidCallback onRetry;
  @override
  Widget build(BuildContext context) {
    final status = error is AnkiFailure ? (error as AnkiFailure).status : null;
    return Center(
      child: Padding(
        padding: const EdgeInsets.all(24),
        child: Column(
          mainAxisSize: MainAxisSize.min,
          children: [
            const Icon(Icons.cloud_off_rounded, size: 36),
            const SizedBox(height: 14),
            Text(
              error is AnkiFailure
                  ? error.toString().localized(context)
                  : 'Deck belum dapat dimuat. Coba kembali.'.localized(context),
              textAlign: TextAlign.center,
              style: const TextStyle(height: 1.6),
            ),
            const SizedBox(height: 16),
            if (status == 401)
              FilledButton(
                onPressed: () async {
                  await context.push<bool>('/login');
                  if (context.mounted) onRetry();
                },
                child: Text('Masuk'.localized(context)),
              )
            else if (status == 402)
              FilledButton(
                onPressed: () async {
                  await context.push('/premium');
                  if (context.mounted) onRetry();
                },
                child: Text('Lifetime Access'.localized(context)),
              )
            else
              OutlinedButton(
                onPressed: onRetry,
                child: Text(
                  status == 409
                      ? 'Muat ulang sesi'.localized(context)
                      : 'Muat ulang'.localized(context),
                ),
              ),
          ],
        ),
      ),
    );
  }
}

class _Section extends StatelessWidget {
  const _Section({required this.title, required this.children});
  final String title;
  final List<Widget> children;
  @override
  Widget build(BuildContext context) => Card(
    child: Padding(
      padding: const EdgeInsets.all(18),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.stretch,
        children: [
          Text(
            title,
            style: const TextStyle(
              fontSize: 17,
              fontWeight: FontWeight.w800,
              height: 1.5,
            ),
          ),
          const SizedBox(height: 12),
          ...children,
        ],
      ),
    ),
  );
}

class _Tag extends StatelessWidget {
  const _Tag(this.text);
  final String text;
  @override
  Widget build(BuildContext context) => Container(
    padding: const EdgeInsets.symmetric(horizontal: 10, vertical: 6),
    decoration: BoxDecoration(
      color: Theme.of(context).colorScheme.secondary.withValues(alpha: .08),
      borderRadius: BorderRadius.circular(10),
    ),
    child: Text(
      text,
      style: const TextStyle(fontSize: 12, fontWeight: FontWeight.w600),
    ),
  );
}

class _Stat extends StatelessWidget {
  const _Stat(this.value, this.label);
  final String value;
  final String label;
  @override
  Widget build(BuildContext context) => Column(
    crossAxisAlignment: CrossAxisAlignment.start,
    children: [
      Text(
        value,
        style: const TextStyle(fontSize: 24, fontWeight: FontWeight.w900),
      ),
      Text(label, style: const TextStyle(fontSize: 12)),
    ],
  );
}

class _DeckIcon extends StatelessWidget {
  const _DeckIcon({required this.category, this.size = 48});
  final String category;
  final double size;
  @override
  Widget build(BuildContext context) => Container(
    width: size,
    height: size,
    decoration: BoxDecoration(
      color: Theme.of(context).colorScheme.secondary.withValues(alpha: .08),
      borderRadius: BorderRadius.circular(16),
    ),
    child: Icon(
      switch (category) {
        'JLPT' => Icons.translate_rounded,
        'Nama & Geografi' => Icons.explore_rounded,
        'Stasiun' => Icons.train_rounded,
        'Kana & Kanji' => Icons.draw_rounded,
        'Inggris & Trivia' => Icons.lightbulb_outline_rounded,
        _ => Icons.style_rounded,
      },
      color: Theme.of(context).colorScheme.secondary,
      size: size * .48,
    ),
  );
}

class _Answers extends StatelessWidget {
  const _Answers({required this.answers, super.key});
  final List<String> answers;
  @override
  Widget build(BuildContext context) => Column(
    crossAxisAlignment: CrossAxisAlignment.stretch,
    children: [
      SelectableText(
        answers.take(8).join(' · '),
        textAlign: TextAlign.center,
        scrollPhysics: const NeverScrollableScrollPhysics(),
        style: TextStyle(
          fontSize: 22,
          fontWeight: FontWeight.w700,
          height: 1.7,
          color: Theme.of(context).colorScheme.secondary,
        ),
      ),
      if (answers.length > 8)
        ExpansionTile(
          title: Text(
            'Lihat ${ankiCount(answers.length - 8)} jawaban lainnya'.localized(
              context,
            ),
          ),
          children: [
            SelectableText(
              answers.skip(8).join(' · '),
              scrollPhysics: const NeverScrollableScrollPhysics(),
              style: const TextStyle(fontSize: 18, height: 1.7),
            ),
          ],
        ),
    ],
  );
}

String _date(BuildContext context, DateTime value) =>
    regionalDateTime(value, localeName: appRegionLocaleName(context));

void _credits(BuildContext context) => showModalBottomSheet<void>(
  context: context,
  isScrollControlled: true,
  showDragHandle: true,
  builder: (context) => FractionallySizedBox(
    heightFactor: .8,
    child: Consumer(
      builder: (context, ref, _) => ref
          .watch(ankiCreditsProvider)
          .when(
            loading: () => const Center(child: CircularProgressIndicator()),
            error: (error, _) => AnkiErrorView(
              error: error,
              onRetry: () => ref.invalidate(ankiCreditsProvider),
            ),
            data: (data) => ListView(
              padding: const EdgeInsets.all(24),
              children: [
                Text(
                  'Kotoba / mistval',
                  style: Theme.of(context).textTheme.titleLarge,
                ),
                const SizedBox(height: 12),
                Text(data['repository'] as String),
                const SizedBox(height: 12),
                Text(
                  data['notice'] as String,
                  style: const TextStyle(height: 1.6),
                ),
                const SizedBox(height: 20),
                Text(
                  'Lisensi repository Kotoba'.localized(context),
                  style: const TextStyle(fontWeight: FontWeight.w800),
                ),
                const SizedBox(height: 12),
                SelectableText(
                  data['license'] as String,
                  scrollPhysics: const NeverScrollableScrollPhysics(),
                  style: const TextStyle(height: 1.6),
                ),
              ],
            ),
          ),
    ),
  ),
);
