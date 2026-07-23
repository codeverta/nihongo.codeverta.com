import 'package:flutter/material.dart';

class ExamHubScreen extends StatelessWidget {
  const ExamHubScreen({super.key});

  @override
  Widget build(BuildContext context) {
    return CustomScrollView(
      key: const PageStorageKey('exam-scroll'),
      slivers: [
        SliverPadding(
          padding: const EdgeInsets.fromLTRB(20, 20, 20, 32),
          sliver: SliverList.list(
            children: const [
              _ExamHeader(),
              SizedBox(height: 22),
              _ReadinessCard(),
              SizedBox(height: 28),
              _Title('Pilih jenis ujian'),
              SizedBox(height: 14),
              _ExamTypeCard(
                code: 'JLPT',
                title: 'Japanese Language Proficiency Test',
                description: 'Simulasi N5–N1 dengan format dan durasi asli.',
                icon: Icons.school_rounded,
                color: Color(0xFFE60012),
                badge: 'N5–N1',
              ),
              SizedBox(height: 12),
              _ExamTypeCard(
                code: 'NAT',
                title: 'NAT-TEST',
                description: 'Latihan level 5Q–1Q dan analisis kemampuan.',
                icon: Icons.assignment_turned_in_rounded,
                color: Color(0xFF1D2088),
                badge: '5Q–1Q',
              ),
              SizedBox(height: 12),
              _ExamTypeCard(
                code: 'JFT',
                title: 'JFT-Basic',
                description: 'Persiapan A2 untuk kebutuhan kerja di Jepang.',
                icon: Icons.flight_takeoff_rounded,
                color: Color(0xFF168C72),
                badge: 'A2',
              ),
              SizedBox(height: 28),
              _Title('Riwayat terbaru'),
              SizedBox(height: 14),
              _ExamHistory(),
            ],
          ),
        ),
      ],
    );
  }
}

class _ExamHeader extends StatelessWidget {
  const _ExamHeader();

  @override
  Widget build(BuildContext context) {
    return Row(
      children: [
        Expanded(
          child: Column(
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              Text(
                'Pusat Ujian',
                style: Theme.of(context).textTheme.headlineMedium?.copyWith(
                      fontWeight: FontWeight.w900,
                    ),
              ),
              const SizedBox(height: 4),
              const Text('Ukur kemampuan dan raih targetmu.'),
            ],
          ),
        ),
        IconButton.filledTonal(
          onPressed: () {},
          icon: const Icon(Icons.calendar_month_rounded),
          tooltip: 'Jadwal ujian',
        ),
      ],
    );
  }
}

class _ReadinessCard extends StatelessWidget {
  const _ReadinessCard();

  @override
  Widget build(BuildContext context) {
    final colorScheme = Theme.of(context).colorScheme;

    return Container(
      padding: const EdgeInsets.all(20),
      decoration: BoxDecoration(
        color: colorScheme.primary,
        borderRadius: BorderRadius.circular(24),
        gradient: LinearGradient(
          colors: [colorScheme.primary, const Color(0xFFB80018)],
        ),
      ),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          const Row(
            mainAxisAlignment: MainAxisAlignment.spaceBetween,
            children: [
              Text(
                'Kesiapan JLPT N5',
                style: TextStyle(
                  color: Colors.white,
                  fontWeight: FontWeight.w800,
                  fontSize: 17,
                ),
              ),
              Text(
                '68%',
                style: TextStyle(
                  color: Colors.white,
                  fontWeight: FontWeight.w900,
                  fontSize: 22,
                ),
              ),
            ],
          ),
          const SizedBox(height: 14),
          LinearProgressIndicator(
            value: 0.68,
            minHeight: 9,
            borderRadius: BorderRadius.circular(99),
            backgroundColor: Colors.white.withValues(alpha: 0.18),
            color: Colors.white,
          ),
          const SizedBox(height: 12),
          Text(
            'Fokus berikutnya: Listening dan Grammar',
            style: TextStyle(color: Colors.white.withValues(alpha: 0.78)),
          ),
          const SizedBox(height: 18),
          FilledButton.icon(
            onPressed: () {},
            style: FilledButton.styleFrom(
              backgroundColor: Colors.white,
              foregroundColor: colorScheme.primary,
            ),
            icon: const Icon(Icons.play_arrow_rounded),
            label: const Text('Mulai diagnostic test'),
          ),
        ],
      ),
    );
  }
}

class _ExamTypeCard extends StatelessWidget {
  const _ExamTypeCard({
    required this.code,
    required this.title,
    required this.description,
    required this.icon,
    required this.color,
    required this.badge,
  });

  final String code;
  final String title;
  final String description;
  final IconData icon;
  final Color color;
  final String badge;

  @override
  Widget build(BuildContext context) {
    return Card(
      child: InkWell(
        onTap: () {},
        borderRadius: BorderRadius.circular(16),
        child: Padding(
          padding: const EdgeInsets.all(16),
          child: Row(
            children: [
              Container(
                width: 60,
                height: 60,
                decoration: BoxDecoration(
                  color: color.withValues(alpha: 0.10),
                  borderRadius: BorderRadius.circular(18),
                ),
                child: Icon(icon, color: color, size: 30),
              ),
              const SizedBox(width: 15),
              Expanded(
                child: Column(
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: [
                    Row(
                      children: [
                        Text(
                          code,
                          style: TextStyle(
                            color: color,
                            fontWeight: FontWeight.w900,
                            fontSize: 17,
                          ),
                        ),
                        const SizedBox(width: 8),
                        Container(
                          padding: const EdgeInsets.symmetric(
                            horizontal: 8,
                            vertical: 3,
                          ),
                          decoration: BoxDecoration(
                            color: color.withValues(alpha: 0.10),
                            borderRadius: BorderRadius.circular(99),
                          ),
                          child: Text(
                            badge,
                            style: TextStyle(
                              color: color,
                              fontWeight: FontWeight.w700,
                              fontSize: 10,
                            ),
                          ),
                        ),
                      ],
                    ),
                    const SizedBox(height: 3),
                    Text(
                      title,
                      maxLines: 1,
                      overflow: TextOverflow.ellipsis,
                      style: const TextStyle(fontWeight: FontWeight.w700),
                    ),
                    const SizedBox(height: 4),
                    Text(
                      description,
                      maxLines: 2,
                      overflow: TextOverflow.ellipsis,
                      style: Theme.of(context).textTheme.bodySmall,
                    ),
                  ],
                ),
              ),
              const Icon(Icons.chevron_right_rounded),
            ],
          ),
        ),
      ),
    );
  }
}

class _ExamHistory extends StatelessWidget {
  const _ExamHistory();

  @override
  Widget build(BuildContext context) {
    return Card(
      child: Column(
        children: [
          const ListTile(
            leading: CircleAvatar(child: Text('N5')),
            title: Text(
              'JLPT N5 • Mini Test',
              style: TextStyle(fontWeight: FontWeight.w700),
            ),
            subtitle: Text('Kemarin • 25 soal'),
            trailing: Text(
              '84%',
              style: TextStyle(
                color: Color(0xFF168C72),
                fontWeight: FontWeight.w900,
              ),
            ),
          ),
          const Divider(height: 1),
          ListTile(
            leading: const CircleAvatar(child: Icon(Icons.lock_rounded)),
            title: const Text(
              'Analisis lengkap jawaban',
              style: TextStyle(fontWeight: FontWeight.w700),
            ),
            subtitle: const Text('Tersedia untuk pengguna Premium'),
            trailing: const Icon(
              Icons.workspace_premium_rounded,
              color: Color(0xFFE59E24),
            ),
            onTap: () {},
          ),
        ],
      ),
    );
  }
}

class _Title extends StatelessWidget {
  const _Title(this.value);

  final String value;

  @override
  Widget build(BuildContext context) {
    return Text(
      value,
      style: Theme.of(context).textTheme.titleLarge?.copyWith(
            fontWeight: FontWeight.w800,
          ),
    );
  }
}
