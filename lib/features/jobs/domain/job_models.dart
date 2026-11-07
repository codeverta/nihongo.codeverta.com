import 'package:intl/intl.dart';

class JapanJob {
  JapanJob.fromJson(Map<String, dynamic> json)
    : id = '${json['id'] ?? ''}',
      title = '${json['title'] ?? ''}',
      company = '${json['company'] ?? ''}',
      category = '${json['category'] ?? ''}',
      sector = '${json['sector'] ?? ''}',
      prefecture = '${json['prefecture'] ?? ''}',
      location = '${json['location'] ?? ''}',
      employmentType = '${json['employmentType'] ?? 'unknown'}',
      remote = '${json['remote'] ?? 'unknown'}',
      visa = '${json['visa'] ?? 'unknown'}',
      japaneseLevel = '${json['japaneseLevel'] ?? 'unknown'}',
      salaryMin = (json['salaryMin'] as num?)?.toDouble(),
      salaryMax = (json['salaryMax'] as num?)?.toDouble(),
      salaryCurrency = '${json['salaryCurrency'] ?? 'JPY'}',
      salaryPeriod = '${json['salaryPeriod'] ?? 'unknown'}',
      salaryText = '${json['salaryText'] ?? ''}',
      description = '${json['description'] ?? ''}',
      qualifications = '${json['qualifications'] ?? ''}',
      benefits = _jobBenefits(json),
      sourceName = '${json['sourceName'] ?? ''}',
      sourceUrl = '${json['sourceUrl'] ?? ''}',
      applyUrl = '${json['applyUrl'] ?? ''}',
      postedAt = DateTime.tryParse('${json['postedAt'] ?? ''}'),
      expiresAt = DateTime.tryParse('${json['expiresAt'] ?? ''}'),
      lastSeenAt = DateTime.tryParse('${json['lastSeenAt'] ?? ''}'),
      status = '${json['status'] ?? 'unknown'}';

  final String id;
  final String title;
  final String company;
  final String category;
  final String sector;
  final String prefecture;
  final String location;
  final String employmentType;
  final String remote;
  final String visa;
  final String japaneseLevel;
  final double? salaryMin;
  final double? salaryMax;
  final String salaryCurrency;
  final String salaryPeriod;
  final String salaryText;
  final String description;
  final String qualifications;
  final String benefits;
  final String sourceName;
  final String sourceUrl;
  final String applyUrl;
  final DateTime? postedAt;
  final DateTime? expiresAt;
  final DateTime? lastSeenAt;
  final String status;

  bool get active =>
      status == 'active' &&
      (expiresAt == null || expiresAt!.isAfter(DateTime.now()));
  String get salary => salaryFor('en_US');

  /// [translate] localizes the fixed Indonesian fragments (not [salaryText],
  /// which is server-provided). Defaults to the Indonesian source text.
  String salaryFor(String localeName, {String Function(String)? translate}) {
    String t(String source) => translate == null ? source : translate(source);
    if (salaryText.trim().isNotEmpty) return salaryText;
    if (salaryMin == null && salaryMax == null) {
      return t('Gaji belum dicantumkan');
    }
    String amount(double value) =>
        NumberFormat.decimalPattern(localeName).format(value);
    final range = salaryMin != null && salaryMax != null
        ? '${amount(salaryMin!)}–${amount(salaryMax!)}'
        : salaryMin != null
        ? '${t('mulai')} ${amount(salaryMin!)}'
        : '${t('hingga')} ${amount(salaryMax!)}';
    final period = switch (salaryPeriod) {
      'hour' || 'hourly' => t('jam'),
      'month' || 'monthly' => t('bulan'),
      'year' || 'yearly' || 'annual' => t('tahun'),
      'day' || 'daily' => t('hari'),
      'week' || 'weekly' => t('minggu'),
      _ => '',
    };
    return '$salaryCurrency $range${period.isEmpty ? '' : ' / $period'}';
  }
}

class JobOption {
  const JobOption(this.id, this.label);
  factory JobOption.fromJson(Map<String, dynamic> json) =>
      JobOption('${json['id']}', '${json['label']}');
  final String id;
  final String label;
}

class JobMeta {
  JobMeta.fromJson(Map<String, dynamic> json)
    : groups = {
        for (final key in const [
          'categories',
          'sectors',
          'prefectures',
          'employmentTypes',
          'japaneseLevels',
          'visaOptions',
        ])
          key: ((json[key] as List?) ?? [])
              .map((v) => JobOption.fromJson(v as Map<String, dynamic>))
              .toList(),
      };
  final Map<String, List<JobOption>> groups;
  String label(String group, String id) =>
      groups[group]?.where((v) => v.id == id).firstOrNull?.label ??
      (id.isEmpty || id == 'unknown' ? 'Belum dicantumkan' : id);
}

class JobPage {
  JobPage.fromJson(Map<String, dynamic> json)
    : items = ((json['items'] as List?) ?? [])
          .map((v) => JapanJob.fromJson(v as Map<String, dynamic>))
          .toList(),
      total = (json['total'] as num?)?.toInt() ?? 0,
      nextCursor = json['nextCursor'] as String?,
      hasMore = json['hasMore'] == true;
  final List<JapanJob> items;
  final int total;
  final String? nextCursor;
  final bool hasMore;
}

Uri? jobExternalUri(String value) {
  final uri = Uri.tryParse(value);
  if (uri == null ||
      uri.scheme != 'https' ||
      uri.host.isEmpty ||
      uri.userInfo.isNotEmpty) {
    return null;
  }
  return uri;
}

String jobDate(DateTime? value, {String localeName = 'id_ID'}) => value == null
    ? 'Belum tersedia'
    : DateFormat.yMd(localeName).format(value.toLocal());

// Older cached listings may contain a clearly labeled benefits section in their
// description. Copy only that section; never infer perks from keywords.
String _jobBenefits(Map<String, dynamic> json) {
  final raw = json['benefits'];
  if (raw is String && raw.trim().isNotEmpty) return raw.trim();
  if (raw is List) {
    return raw.whereType<String>().where((v) => v.trim().isNotEmpty).join('\n');
  }
  final description = json['description'] as String? ?? '';
  final blocks = description.split(RegExp(r'\n\s*\n'));
  final heading = RegExp(
    r'^(?:benefits(?:\s*(?:&|and)\s*perks)?|perks(?:\s*(?:&|and)\s*benefits)?|福利厚生|待遇[・/／]福利厚生|benefit(?: dan tunjangan)?)[:：]?$',
    caseSensitive: false,
  );
  for (var i = 0; i < blocks.length; i++) {
    final lines = blocks[i].trim().split('\n');
    if (lines.isNotEmpty && heading.hasMatch(lines.first.trim())) {
      if (lines.length > 1) return lines.skip(1).join('\n').trim();
      // A detached heading is ambiguous; show the honest empty state instead.
    }
  }
  return '';
}
