import 'package:flutter/foundation.dart';
import 'package:flutter/material.dart';
import 'package:flutter_localizations/flutter_localizations.dart';
import 'package:nihongo_master/core/localization/anki_english_part_a.dart';
import 'package:nihongo_master/core/localization/anki_english_part_b.dart';
import 'package:nihongo_master/core/localization/anki_english_part_c.dart';

/// Lightweight app localization catalog. Japanese source text remains as-is;
/// only app-owned Indonesian copy is translated for English users.
class AppLocalizations {
  const AppLocalizations(this.locale);
  AppLocalizations.of(BuildContext context)
    : locale =
          Localizations.of<AppLocalizations>(
            context,
            AppLocalizations,
          )?.locale ??
          const Locale('id');

  final Locale locale;

  static const supportedLocales = [Locale('id'), Locale('en')];

  static const localizationsDelegates = <LocalizationsDelegate<dynamic>>[
    delegate,
    GlobalMaterialLocalizations.delegate,
    GlobalWidgetsLocalizations.delegate,
    GlobalCupertinoLocalizations.delegate,
  ];

  static const LocalizationsDelegate<AppLocalizations> delegate =
      _AppLocalizationsDelegate();

  String text(String source) {
    if (locale.languageCode != 'en') return source;
    final curriculum = RegExp(r'^Kurikulum (.+)$').firstMatch(source);
    if (curriculum != null) {
      return 'The ${text(curriculum.group(1)!)} curriculum';
    }
    final achievementCount = RegExp(
      r'^(?:Lihat semua achievement \(([0-9]+)\/([0-9]+)\)|([0-9]+)\/([0-9]+) achievement terbuka)$',
    ).firstMatch(source);
    if (achievementCount != null) {
      final unlocked = achievementCount.group(1) ?? achievementCount.group(3)!;
      final total = achievementCount.group(2) ?? achievementCount.group(4)!;
      return source.startsWith('Lihat semua')
          ? 'View all achievements ($unlocked/$total)'
          : '$unlocked/$total achievements unlocked';
    }
    final bannerFeaturePrice = RegExp(
      r'^(.+) • Rp([\d.]+)$',
    ).firstMatch(source);
    if (bannerFeaturePrice != null) {
      final feature = text(bannerFeaturePrice.group(1)!);
      final price = bannerFeaturePrice.group(2)!.replaceAll('.', ',');
      return '$feature · IDR $price';
    }
    final sswLessonSummary = RegExp(
      r'^(\d+) modul · (\d+) lesson · ±(\d+) jam belajar$',
    ).firstMatch(source);
    if (sswLessonSummary != null) {
      return '${sswLessonSummary.group(1)} modules · ${sswLessonSummary.group(2)} lessons · about ${sswLessonSummary.group(3)} hours of study';
    }
    final lessonProgress = RegExp(
      r'^(\d+)\/(\d+) lesson selesai$',
    ).firstMatch(source);
    if (lessonProgress != null) {
      return '${lessonProgress.group(1)}/${lessonProgress.group(2)} lessons completed';
    }
    final completedProgress = RegExp(
      r'^(\d+)\/(\d+) selesai$',
    ).firstMatch(source);
    if (completedProgress != null) {
      return '${completedProgress.group(1)}/${completedProgress.group(2)} completed';
    }
    final lessonChecks = RegExp(
      r'^±(\d+) menit · (\d+) cek pemahaman$',
    ).firstMatch(source);
    if (lessonChecks != null) {
      return 'About ${lessonChecks.group(1)} min · ${lessonChecks.group(2)} comprehension checks';
    }
    final lessonHeader = RegExp(
      r'^LESSON (\d+) \/ (\d+) · ±(\d+) MENIT$',
    ).firstMatch(source);
    if (lessonHeader != null) {
      return 'LESSON ${lessonHeader.group(1)} / ${lessonHeader.group(2)} · ABOUT ${lessonHeader.group(3)} MIN';
    }
    final answer = RegExp(r'^Jawaban: (.+)$').firstMatch(source);
    if (answer != null) return 'Answer: ${answer.group(1)}';
    final sswFieldCount = RegExp(
      r'^Latihan soal dan simulasi untuk (\d+) bidang SSW$',
    ).firstMatch(source);
    if (sswFieldCount != null) {
      return 'Practice questions and simulations for ${sswFieldCount.group(1)} SSW fields';
    }
    final sswPracticeCount = RegExp(
      r'^Latihan per topik dan simulasi belajar tersedia untuk (\d+) bidang SSW\. Paket aplikasi memakai soal orisinal dan dapat berbeda dari format tes resmi\.$',
    ).firstMatch(source);
    if (sswPracticeCount != null) {
      return 'Topic practice and study simulations are available for ${sswPracticeCount.group(1)} SSW fields. The app uses original questions that may differ from the official exam format.';
    }
    final sswCurriculumCount = RegExp(
      r'^Kurikulum, kasus kerja, latihan per topik, dan simulasi belajar tersedia untuk (\d+) bidang SSW\. Mulai dari materi, lalu lanjutkan ke latihan dan simulasi\.$',
    ).firstMatch(source);
    if (sswCurriculumCount != null) {
      return 'Curricula, workplace cases, topic practice, and study simulations are available for ${sswCurriculumCount.group(1)} SSW fields. Start with the lessons, then continue to practice and simulations.';
    }
    final fieldCount = RegExp(r'^(\d+) bidang$').firstMatch(source);
    if (fieldCount != null) return '${fieldCount.group(1)} fields';
    final examCount = RegExp(r'^(\d+) soal · (\d+) menit$').firstMatch(source);
    if (examCount != null) {
      return '${examCount.group(1)} questions · ${examCount.group(2)} min';
    }
    final questionCount = RegExp(r'^(\d+) soal$').firstMatch(source);
    if (questionCount != null) return '${questionCount.group(1)} questions';
    final firstTry = RegExp(
      r'^(\d+)\/(\d+) benar pada jawaban pertama$',
    ).firstMatch(source);
    if (firstTry != null) {
      return '${firstTry.group(1)}/${firstTry.group(2)} correct on the first try';
    }
    final questionProgress = RegExp(
      r'^Soal (\d+)\/(\d+) · (\d+) benar$',
    ).firstMatch(source);
    if (questionProgress != null) {
      return 'Question ${questionProgress.group(1)}/${questionProgress.group(2)} · ${questionProgress.group(3)} correct';
    }
    final sectionHeading = RegExp(
      r'^(\d+)\. (文字・語彙|文法・読解|文法) — (\d+) soal \/ (\d+) menit$',
    ).firstMatch(source);
    if (sectionHeading != null) {
      return '${sectionHeading.group(1)}. ${sectionHeading.group(2)} — ${sectionHeading.group(3)} questions / ${sectionHeading.group(4)} min';
    }
    final jftSectionHeading = RegExp(
      r'^Bagian (\d+) · (文字と語彙|会話と表現|読解)$',
    ).firstMatch(source);
    if (jftSectionHeading != null) {
      return 'Part ${jftSectionHeading.group(1)} · ${jftSectionHeading.group(2)}';
    }
    final jlptSimulationTitle = RegExp(
      r'^Simulasi (N[1-5])$',
    ).firstMatch(source);
    if (jlptSimulationTitle != null) {
      return '${jlptSimulationTitle.group(1)} simulation';
    }
    final simulationTitle = RegExp(r'^Simulasi( ·)? (.+)$').firstMatch(source);
    if (simulationTitle != null) {
      return 'Simulation${simulationTitle.group(1) ?? ''} ${text(simulationTitle.group(2)!)}';
    }
    final natExamIntro = RegExp(
      r'^Exam (\d+) · (\d+) soal · (\d+) menit\nKosakata, kanji, grammar, dan reading NAT 5Q\. Listening belum termasuk\.$',
    ).firstMatch(source);
    if (natExamIntro != null) {
      return 'NAT 5Q Exam ${natExamIntro.group(1)} · ${natExamIntro.group(2)} questions · ${natExamIntro.group(3)} minutes\nVocabulary, kanji, grammar, and reading for NAT 5Q. Listening is not included.';
    }
    final jftExamIntro = RegExp(
      r'^Exam (\d+) · 36 soal · 45 menit\nLatihan tertulis orisinal JFT-Basic\. Listening belum termasuk\.$',
    ).firstMatch(source);
    if (jftExamIntro != null) {
      return 'JFT-Basic Exam ${jftExamIntro.group(1)} · 36 questions · 45 minutes\nOriginal written practice for JFT-Basic. Listening is not included.';
    }
    final simulationResultTitle = RegExp(
      r'^Hasil simulasi (.+)$',
    ).firstMatch(source);
    if (simulationResultTitle != null) {
      return '${text(simulationResultTitle.group(1)!)} simulation results';
    }
    final jftResultTitle = RegExp(r'^Hasil JFT Exam (.+)$').firstMatch(source);
    if (jftResultTitle != null) {
      return 'JFT Exam ${jftResultTitle.group(1)} Results';
    }
    final natResultTitle = RegExp(
      r'^Hasil NAT 5Q Exam (.+)$',
    ).firstMatch(source);
    if (natResultTitle != null) {
      return 'NAT 5Q Exam ${natResultTitle.group(1)} Results';
    }
    final topicLevel = RegExp(r'^Jelajahi tema (.+)$').firstMatch(source);
    if (topicLevel != null) {
      return 'Explore ${text(topicLevel.group(1)!)} topics';
    }
    final sectionSubmit = RegExp(r'^Kirim (.+)\?$').firstMatch(source);
    if (sectionSubmit != null) {
      return 'Submit ${text(sectionSubmit.group(1)!)}?';
    }
    final unitPosition = RegExp(r'^(.+) (\d+) dari (\d+)$').firstMatch(source);
    if (unitPosition != null) {
      return '${text(unitPosition.group(1)!)} ${unitPosition.group(2)} of ${unitPosition.group(3)}';
    }
    final answeredSummary = RegExp(
      r'^(\d+) dijawab\n(\d+) soal kosong\n(\d+) menit (\d+) detik$',
    ).firstMatch(source);
    if (answeredSummary != null) {
      return '${answeredSummary.group(1)} answered\n${answeredSummary.group(2)} unanswered\n${answeredSummary.group(3)} min ${answeredSummary.group(4)} sec';
    }
    final simulationSummary = RegExp(
      r'^(\d+) dijawab · (\d+) kosong\nDurasi (\d+) menit (\d+) detik$',
    ).firstMatch(source);
    if (simulationSummary != null) {
      return '${simulationSummary.group(1)} answered · ${simulationSummary.group(2)} unanswered\nDuration ${simulationSummary.group(3)} min ${simulationSummary.group(4)} sec';
    }
    final correctSummary = RegExp(
      r'^(\d+) dari (\d+) soal benar\. Gunakan hasil ini untuk menentukan fokus belajar\.$',
    ).firstMatch(source);
    if (correctSummary != null) {
      return '${correctSummary.group(1)} of ${correctSummary.group(2)} questions correct. Use these results to choose what to focus on.';
    }
    final natExamHistory = RegExp(
      r'^Exam (\d+) · (\d+)\/(\d+) benar$',
    ).firstMatch(source);
    if (natExamHistory != null) {
      return 'Exam ${natExamHistory.group(1)} · ${natExamHistory.group(2)}/${natExamHistory.group(3)} correct';
    }
    final jlptExamHistory = RegExp(
      r'^Paket 01 · (\d+)\/(\d+) benar$',
    ).firstMatch(source);
    if (jlptExamHistory != null) {
      return 'Set 01 · ${jlptExamHistory.group(1)}/${jlptExamHistory.group(2)} correct';
    }
    final jftResultSubtitle = RegExp(
      r'^JFT Exam (\d+) · Kosakata, Percakapan & Reading\n(\d+)\/(\d+) benar · (\d+)%$',
    ).firstMatch(source);
    if (jftResultSubtitle != null) {
      return 'JFT Exam ${jftResultSubtitle.group(1)} · Vocabulary, Conversation & Reading\n${jftResultSubtitle.group(2)}/${jftResultSubtitle.group(3)} correct · ${jftResultSubtitle.group(4)}%';
    }
    final natResultSubtitle = RegExp(
      r'^NAT 5Q Exam (\d+) · Kosakata, Grammar & Reading\n(\d+)\/(\d+) benar · (\d+)%$',
    ).firstMatch(source);
    if (natResultSubtitle != null) {
      return 'NAT 5Q Exam ${natResultSubtitle.group(1)} · Vocabulary, Grammar & Reading\n${natResultSubtitle.group(2)}/${natResultSubtitle.group(3)} correct · ${natResultSubtitle.group(4)}%';
    }
    final jlptResultSubtitle = RegExp(
      r'^Paket 01 · (N[1-5]) · Kosakata & Grammar( & Reading)?\n(\d+)\/(\d+) benar · (\d+)%$',
    ).firstMatch(source);
    if (jlptResultSubtitle != null) {
      return 'Set 01 · ${jlptResultSubtitle.group(1)} · Vocabulary & Grammar${jlptResultSubtitle.group(2) ?? ''}\n${jlptResultSubtitle.group(3)}/${jlptResultSubtitle.group(4)} correct · ${jlptResultSubtitle.group(5)}%';
    }
    final unansweredNotice = RegExp(
      r'^Masih ada (\d+) soal kosong; soal kosong dihitung salah\.\n\n$',
    ).firstMatch(source);
    if (unansweredNotice != null) {
      return 'There are ${unansweredNotice.group(1)} unanswered questions; unanswered questions count as incorrect.\n\n';
    }
    final unansweredDiagnostic = RegExp(
      r'^Masih ada (\d+) soal kosong\. Soal kosong dihitung salah\. Kamu bisa memeriksanya sebelum mengirim\.$',
    ).firstMatch(source);
    if (unansweredDiagnostic != null) {
      return 'There are ${unansweredDiagnostic.group(1)} unanswered questions. They count as incorrect, and you can review them before submitting.';
    }
    final sswSectionChange = RegExp(
      r'^Bagian sebelumnya dikunci\. (.+) dimulai dengan timer bagian baru\.$',
    ).firstMatch(source);
    if (sswSectionChange != null) {
      return 'The previous section is locked. ${text(sswSectionChange.group(1)!)} starts with a new section timer.';
    }
    final jftSectionChange = RegExp(
      r'^Bagian sebelumnya dikunci\. (.+) dimulai; timer total terus berjalan\.$',
    ).firstMatch(source);
    if (jftSectionChange != null) {
      return 'The previous section is locked. ${text(jftSectionChange.group(1)!)} starts; the total timer keeps running.';
    }
    final jlptLegacy = RegExp(r'^JLPT lama (.+)$').firstMatch(source);
    if (jlptLegacy != null) return 'Legacy JLPT ${jlptLegacy.group(1)}';
    final transferCounts = RegExp(
      r'^(\d+) kartu · (\d+) media$',
    ).firstMatch(source);
    if (transferCounts != null) {
      return '${transferCounts.group(1)} cards · ${transferCounts.group(2)} media';
    }
    final ankiCategoryCount = RegExp(
      r'^(.+) · (\d+) kartu$',
    ).firstMatch(source);
    if (ankiCategoryCount != null) {
      return '${text(ankiCategoryCount.group(1)!)} · ${ankiCategoryCount.group(2)} cards';
    }
    final exact =
        _englishCopy[source] ??
        ankiEnglishCopyA[source] ??
        ankiEnglishCopyB[source] ??
        ankiEnglishCopyC[source];
    if (exact != null) return exact;

    final learningCourseStats = RegExp(
      r'^(\d+) lesson • (\d+) pola • (\d+) soal$',
    ).firstMatch(source);
    if (learningCourseStats != null) {
      return '${learningCourseStats.group(1)} lessons · ${learningCourseStats.group(2)} patterns · ${learningCourseStats.group(3)} questions';
    }
    final kanjiCourseStats = RegExp(
      r'^(\d+) lesson • (\d+) kanji • (\d+) soal$',
    ).firstMatch(source);
    if (kanjiCourseStats != null) {
      return '${kanjiCourseStats.group(1)} lessons · ${kanjiCourseStats.group(2)} kanji · ${kanjiCourseStats.group(3)} questions';
    }
    final readingCourseStats = RegExp(
      r'^(\d+) lesson • (\d+) adegan bergambar • (\d+) soal$',
    ).firstMatch(source);
    if (readingCourseStats != null) {
      return '${readingCourseStats.group(1)} lessons · ${readingCourseStats.group(2)} illustrated scenes · ${readingCourseStats.group(3)} questions';
    }
    final readingLessonSubtitle = RegExp(
      r'^(.+)\n(\d+) menit • (\d+) adegan • Terbaik (\d+)%$',
    ).firstMatch(source);
    if (readingLessonSubtitle != null) {
      return '${readingLessonSubtitle.group(1)}\n${readingLessonSubtitle.group(2)} min · ${readingLessonSubtitle.group(3)} scenes · Best ${readingLessonSubtitle.group(4)}%';
    }
    final courseProgress = RegExp(
      r'^(\d+)/(\d+) (lesson|cerita) selesai • (\d+)%$',
    ).firstMatch(source);
    if (courseProgress != null) {
      final noun = courseProgress.group(3) == 'cerita' ? 'stories' : 'lessons';
      return '${courseProgress.group(1)}/${courseProgress.group(2)} $noun completed · ${courseProgress.group(4)}%';
    }
    final questionMapProgress = RegExp(
      r'^(\d+)/(\d+) dijawab [•·] (\d+) ditandai$',
    ).firstMatch(source);
    if (questionMapProgress != null) {
      return '${questionMapProgress.group(1)}/${questionMapProgress.group(2)} answered · ${questionMapProgress.group(3)} marked';
    }
    final correctAnswers = RegExp(
      r'^(\d+)/(\d+) jawaban benar$',
    ).firstMatch(source);
    if (correctAnswers != null) {
      return '${correctAnswers.group(1)}/${correctAnswers.group(2)} correct answers';
    }
    final correctAnswersOutOf = RegExp(
      r'^(\d+) dari (\d+) jawaban benar$',
    ).firstMatch(source);
    if (correctAnswersOutOf != null) {
      return '${correctAnswersOutOf.group(1)} of ${correctAnswersOutOf.group(2)} answers correct';
    }
    final bestScore = RegExp(r'^Nilai terbaik:? (\d+)%$').firstMatch(source);
    if (bestScore != null) return 'Best score: ${bestScore.group(1)}%';
    final courseProgressPercent = RegExp(
      r'^Progress kursus (\d+)%$',
    ).firstMatch(source);
    if (courseProgressPercent != null) {
      return 'Course progress ${courseProgressPercent.group(1)}%';
    }
    final totalProgressPercent = RegExp(
      r'^Progress total: (\d+)%$',
    ).firstMatch(source);
    if (totalProgressPercent != null) {
      return 'Overall progress: ${totalProgressPercent.group(1)}%';
    }
    final lessonSubtitle = RegExp(
      r'^(.+)\n(\d+) pola • (\d+) soal • Terbaik (\d+)%$',
    ).firstMatch(source);
    if (lessonSubtitle != null) {
      return '${lessonSubtitle.group(1)}\n${lessonSubtitle.group(2)} patterns · ${lessonSubtitle.group(3)} questions · Best ${lessonSubtitle.group(4)}%';
    }
    final kanjiLessonSubtitle = RegExp(
      r'^(.+)\n(\d+) kanji • (\d+) soal • Terbaik (\d+)%$',
    ).firstMatch(source);
    if (kanjiLessonSubtitle != null) {
      return '${kanjiLessonSubtitle.group(1)}\n${kanjiLessonSubtitle.group(2)} kanji · ${kanjiLessonSubtitle.group(3)} questions · Best ${kanjiLessonSubtitle.group(4)}%';
    }
    final vocabularyQuestionCount = RegExp(
      r'^(\d+) kata • (\d+) soal$',
    ).firstMatch(source);
    if (vocabularyQuestionCount != null) {
      return '${vocabularyQuestionCount.group(1)} words · ${vocabularyQuestionCount.group(2)} questions';
    }
    final vocabularyMastered = RegExp(
      r'^Dikuasai • (\d+)%$',
    ).firstMatch(source);
    if (vocabularyMastered != null) {
      return 'Mastered · ${vocabularyMastered.group(1)}%';
    }
    final vocabularyBest = RegExp(r'^Terbaik (\d+)%$').firstMatch(source);
    if (vocabularyBest != null) return 'Best ${vocabularyBest.group(1)}%';
    final vocabularyIllustration = RegExp(
      r'^Ilustrasi tema (.+)$',
    ).firstMatch(source);
    if (vocabularyIllustration != null) {
      return 'Illustration for ${vocabularyIllustration.group(1)}';
    }
    final kanjiCopyArea = RegExp(
      r'^Area menyalin kanji (.+)$',
    ).firstMatch(source);
    if (kanjiCopyArea != null) {
      return 'Kanji copy area: ${kanjiCopyArea.group(1)}';
    }
    final exploreLevelTopics = RegExp(
      r'^Jelajahi tema (N[1-5])$',
    ).firstMatch(source);
    if (exploreLevelTopics != null) {
      return 'Explore ${exploreLevelTopics.group(1)} topics';
    }
    final diagnosticSummary = RegExp(
      r'^(\d+) dari (\d+) soal benar\. Gunakan hasil ini untuk menentukan fokus belajar\.$',
    ).firstMatch(source);
    if (diagnosticSummary != null) {
      return '${diagnosticSummary.group(1)} of ${diagnosticSummary.group(2)} questions correct. Use these results to choose what to focus on.';
    }
    final diagnosticStats = RegExp(
      r'^(\d+) dijawab\n(\d+) soal kosong\n(\d+) menit (\d+) detik$',
    ).firstMatch(source);
    if (diagnosticStats != null) {
      return '${diagnosticStats.group(1)} answered\n${diagnosticStats.group(2)} unanswered\n${diagnosticStats.group(3)} min ${diagnosticStats.group(4)} sec';
    }
    final simulationStats = RegExp(
      r'^(\d+) dijawab · (\d+) kosong\nDurasi (\d+) menit (\d+) detik$',
    ).firstMatch(source);
    if (simulationStats != null) {
      return '${simulationStats.group(1)} answered · ${simulationStats.group(2)} unanswered\nDuration ${simulationStats.group(3)} min ${simulationStats.group(4)} sec';
    }
    final missingQuestionsConfirmation = RegExp(
      r'^Masih ada (\d+) soal kosong; soal kosong dihitung salah\.(\n\n)?$',
    ).firstMatch(source);
    if (missingQuestionsConfirmation != null) {
      final lines = missingQuestionsConfirmation.group(2) ?? '';
      return 'There are ${missingQuestionsConfirmation.group(1)} unanswered questions; unanswered questions count as incorrect.$lines';
    }

    // Keep content supplied by the API in its selected language. These small
    // templates cover app-owned messages that interpolate user values.
    if (source.startsWith('Halo, ')) {
      return source.replaceFirst('Halo, ', 'Hello, ');
    }
    if (source.startsWith('おはよう, ')) {
      return source.replaceFirst('おはよう, ', 'Good morning, ');
    }
    if (source.startsWith('Selamat datang, ')) {
      return source.replaceFirst('Selamat datang, ', 'Welcome, ');
    }
    if (source.startsWith('Halaman ') && source.contains(' dari ')) {
      return source.replaceFirst(' dari ', ' of ');
    }
    final filterCount = RegExp(r'^Filter · (\d+)$').firstMatch(source);
    if (filterCount != null) return 'Filters · ${filterCount.group(1)}';
    final favorites = RegExp(
      r'^Favorit tersimpan di perangkat • (.+)$',
    ).firstMatch(source);
    if (favorites != null) {
      return 'Favorites saved on this device • ${favorites.group(1)}';
    }
    final loadedWords = RegExp(
      r'^(\d+) kata dimuat untuk “(.+)”$',
    ).firstMatch(source);
    if (loadedWords != null) {
      return '${loadedWords.group(1)} words loaded for “${loadedWords.group(2)}”';
    }
    final articleCount = RegExp(r'^(\d+) artikel$').firstMatch(source);
    if (articleCount != null) return '${articleCount.group(1)} articles';
    final daysAgo = RegExp(r'^(\d+) hari lalu$').firstMatch(source);
    if (daysAgo != null) return '${daysAgo.group(1)} days ago';
    final jobsCount = RegExp(r'^(\d+) lowongan$').firstMatch(source);
    if (jobsCount != null) {
      return '${jobsCount.group(1)} ${jobsCount.group(1) == '1' ? 'job' : 'jobs'}';
    }
    final checkedJob = RegExp(r'^(.+) · Diperiksa (.+)$').firstMatch(source);
    if (checkedJob != null) {
      return '${checkedJob.group(1)} · Checked ${checkedJob.group(2)}';
    }
    final updateInterval = RegExp(
      r'^Interval: (\d+) (hari|jam)$',
    ).firstMatch(source);
    if (updateInterval != null) {
      final unit = updateInterval.group(2) == 'hari' ? 'days' : 'hours';
      return 'Update interval: ${updateInterval.group(1)} $unit';
    }
    final lastUpdated = RegExp(
      r'^Berhasil diperbarui: (.+)$',
    ).firstMatch(source);
    if (lastUpdated != null) return 'Last updated: ${lastUpdated.group(1)}';
    final nextUpdate = RegExp(r'^Jadwal berikutnya: (.+)$').firstMatch(source);
    if (nextUpdate != null) {
      return 'Next scheduled update: ${nextUpdate.group(1)}';
    }
    final retryAt = RegExp(r'^Coba kembali: (.+)$').firstMatch(source);
    if (retryAt != null) return 'Retry at: ${retryAt.group(1)}';
    final jobCategory = RegExp(r'^Kategori (.+)$').firstMatch(source);
    if (jobCategory != null) return 'Category ${text(jobCategory.group(1)!)}';
    final mangaPage = RegExp(r'^Halaman (\d+)$').firstMatch(source);
    if (mangaPage != null) return 'Page ${mangaPage.group(1)}';
    final mangaTitles = RegExp(
      r'^(\d+) judul untuk dijelajahi$',
    ).firstMatch(source);
    if (mangaTitles != null) {
      final noun = mangaTitles.group(1) == '1' ? 'title' : 'titles';
      return '${mangaTitles.group(1)} $noun to explore';
    }
    final mangaChapterCount = RegExp(
      r'^(\d+) chapter · (.+)$',
    ).firstMatch(source);
    if (mangaChapterCount != null) {
      return '${mangaChapterCount.group(1)} chapters · ${text(mangaChapterCount.group(2)!)}';
    }
    final ankiDeckCount = RegExp(r'^(\d+) deck$').firstMatch(source);
    if (ankiDeckCount != null) {
      final noun = ankiDeckCount.group(1) == '1' ? 'deck' : 'decks';
      return '${ankiDeckCount.group(1)} $noun';
    }
    final ankiCardCount = RegExp(r'^([\d.]+) kartu$').firstMatch(source);
    if (ankiCardCount != null) {
      final noun = ankiCardCount.group(1) == '1' ? 'card' : 'cards';
      return '${ankiCardCount.group(1)} $noun';
    }
    final ankiReadyCount = RegExp(r'^([\d.]+) kartu siap$').firstMatch(source);
    if (ankiReadyCount != null) {
      return '${ankiReadyCount.group(1)} cards ready';
    }
    final deckCards = RegExp(r'^(.+) · ([\d.]+) kartu$').firstMatch(source);
    if (deckCards != null) {
      return '${text(deckCards.group(1)!)} · ${deckCards.group(2)} cards';
    }
    final skippedAnkiEntries = RegExp(
      r'^([\d.]+) entri sumber disimpan\. ([\d.]+) entri belum dapat digunakan sebagai flashcard karena data kosong atau memerlukan mesin tambahan\.$',
    ).firstMatch(source);
    if (skippedAnkiEntries != null) {
      return '${skippedAnkiEntries.group(1)} source entries are saved. ${skippedAnkiEntries.group(2)} entries cannot be used as flashcards because data is missing or requires an additional engine.';
    }
    final nextReview = RegExp(r'^Review berikutnya: (.+)$').firstMatch(source);
    if (nextReview != null) return 'Next review: ${nextReview.group(1)}';
    final studyCard = RegExp(r'^KARTU (\d+) / (\d+)$').firstMatch(source);
    if (studyCard != null) {
      return 'CARD ${studyCard.group(1)} / ${studyCard.group(2)}';
    }
    final moreAnkiAnswers = RegExp(
      r'^Lihat ([\d.]+) jawaban lainnya$',
    ).firstMatch(source);
    if (moreAnkiAnswers != null) {
      return 'See ${moreAnkiAnswers.group(1)} more answers';
    }
    final sessionPending = RegExp(
      r'^(\d+) kartu telah kamu nilai\. Penilaian tersimpan di perangkat dan menunggu sinkronisasi\. Jadwal resmi diperbarui setelah terhubung\.$',
    ).firstMatch(source);
    if (sessionPending != null) {
      return '${sessionPending.group(1)} cards rated. Ratings are saved on this device and waiting to sync. The official schedule will update when you reconnect.';
    }
    final sessionComplete = RegExp(
      r'^(\d+) kartu telah kamu nilai\. Jadwal berikutnya tersimpan di akunmu\. Kartu yang belum hafal dijadwalkan lagi sekitar 1 menit\.$',
    ).firstMatch(source);
    if (sessionComplete != null) {
      return '${sessionComplete.group(1)} cards rated. Your next schedule is saved to your account. Cards you did not remember will be scheduled again in about 1 minute.';
    }
    final synchronizedAt = RegExp(r'^Disinkronkan (.+)$').firstMatch(source);
    if (synchronizedAt != null) {
      return 'Synced ${synchronizedAt.group(1)}';
    }
    final articleDate = RegExp(
      r'^(.+) · (Tanggal sumber|Terbit) (.+)$',
    ).firstMatch(source);
    if (articleDate != null) {
      final label = articleDate.group(2) == 'Tanggal sumber'
          ? 'Source date'
          : 'Published';
      return '${articleDate.group(1)} · $label ${articleDate.group(3)}';
    }
    final imageCredit = RegExp(r'^Gambar: (.+) · (.+)$').firstMatch(source);
    if (imageCredit != null) {
      return 'Image: ${imageCredit.group(1)} · ${imageCredit.group(2)}';
    }
    final credit = RegExp(
      r'^(Penerjemah|Penerbit|Penulis): (.+)$',
    ).firstMatch(source);
    if (credit != null) {
      final role = switch (credit.group(1)) {
        'Penerjemah' => 'Translator',
        'Penerbit' => 'Publisher',
        _ => 'Author',
      };
      return '$role: ${credit.group(2)}';
    }
    if (source.startsWith('Lisensi gambar · ')) {
      return source.replaceFirst('Lisensi gambar · ', 'Image license · ');
    }
    if (source.startsWith('Lisensi gambar ')) {
      return source.replaceFirst('Lisensi gambar ', 'Image license ');
    }
    if (source.startsWith('Sumber gambar')) {
      return source.replaceFirst('Sumber gambar', 'Image source');
    }
    if (source.startsWith('Artikel asli')) {
      return source.replaceFirst('Artikel asli', 'Original article');
    }
    final achievementProgress = RegExp(
      r'^(\d+) lesson lulus · (\d+) jenis ujian dikerjakan$',
    ).firstMatch(source);
    if (achievementProgress != null) {
      return '${achievementProgress.group(1)} lessons passed · ${achievementProgress.group(2)} exam types completed';
    }
    final streakSummary = RegExp(
      r'^Streak terbaik (\d+) hari · (\d+) hari aktif$',
    ).firstMatch(source);
    if (streakSummary != null) {
      return 'Best streak: ${streakSummary.group(1)} days · ${streakSummary.group(2)} active days';
    }
    final earnedOn = RegExp(r'^Diperoleh (.+)$').firstMatch(source);
    if (earnedOn != null) return 'Earned ${earnedOn.group(1)}';
    final lessons = RegExp(r'^(\d+) pelajaran$').firstMatch(source);
    if (lessons != null) return '${lessons.group(1)} lessons';
    final progress = RegExp(r'^(.+) (\d+) dari (\d+)$').firstMatch(source);
    if (progress != null) {
      return '${text(progress.group(1)!)} ${progress.group(2)} of ${progress.group(3)}';
    }
    final kanjiDetails = RegExp(r'^Detail kanji (.+)$').firstMatch(source);
    if (kanjiDetails != null) return 'Kanji details: ${kanjiDetails.group(1)}';
    final otherMeanings = RegExp(r'^\+(\d+) makna lainnya$').firstMatch(source);
    if (otherMeanings != null) {
      return '+${otherMeanings.group(1)} more meanings';
    }
    if (source.startsWith('Hapus ') && source.endsWith(' dari favorit')) {
      return source
          .replaceFirst('Hapus ', 'Remove ')
          .replaceFirst(' dari favorit', ' from favorites');
    }
    if (source.startsWith('Simpan ')) {
      return source.replaceFirst('Simpan ', 'Save ');
    }
    if (source.startsWith('Berlaku untuk: ')) {
      return source.replaceFirst('Berlaku untuk: ', 'Applies to: ');
    }
    if (source.startsWith('Gunakan zona waktu perangkat (') &&
        source.endsWith(')')) {
      return source.replaceFirst(
        'Gunakan zona waktu perangkat (',
        'Use device time zone (',
      );
    }
    final strokes = RegExp(r'^(\d+) goresan$').firstMatch(source);
    if (strokes != null) return '${strokes.group(1)} strokes';
    final frequency = RegExp(
      r'^Peringkat frekuensi koran: (.+)$',
    ).firstMatch(source);
    if (frequency != null) {
      return 'Newspaper frequency rank: ${frequency.group(1)}';
    }
    return source;
  }
}

class _AppLocalizationsDelegate
    extends LocalizationsDelegate<AppLocalizations> {
  const _AppLocalizationsDelegate();

  @override
  bool isSupported(Locale locale) =>
      locale.languageCode == 'id' || locale.languageCode == 'en';

  @override
  Future<AppLocalizations> load(Locale locale) =>
      SynchronousFuture<AppLocalizations>(AppLocalizations(locale));

  @override
  bool shouldReload(_AppLocalizationsDelegate old) => false;
}

extension LocalizedAppCopy on String {
  String localized(BuildContext context) =>
      AppLocalizations.of(context).text(this);
}

const _englishCopy = <String, String>{
  'Import deck': 'Import deck',
  'Export deck': 'Export deck',
  'Personal': 'Personal',
  'Deck pribadi': 'Personal deck',
  'Deck pribadi hasil impor': 'Imported personal deck',
  'Bawa kartu belajarmu ke sini': 'Bring your flashcards here',
  'Pilih file deck': 'Choose a deck file',
  'Menyimpan deck…': 'Saving deck…',
  'Membaca dan memeriksa file…': 'Reading and checking file…',
  'Nama deck': 'Deck name',
  'Import sebagai deck pribadi': 'Import as a personal deck',
  'Format file & media': 'File formats & media',
  'Ukuran file maksimal 20 MB.': 'Maximum file size is 20 MB.',
  'Isi nama deck terlebih dahulu.': 'Enter a deck name first.',
  'File belum dapat dibuka. Coba pilih ulang.':
      'Could not open this file. Select it again.',
  'Impor belum selesai. Coba lagi dengan file yang sama.':
      'Import has not finished. Retry with the same file.',
  'APKG · untuk Anki, dengan media': 'APKG · for Anki, including media',
  'CSV · teks dan tautan eksternal': 'CSV · text and external links',
  'JSON · backup kartu dan media': 'JSON · card and media backup',
  'Menyiapkan file…': 'Preparing file…',
  'Hapus deck pribadi?': 'Delete personal deck?',
  'Hapus deck pribadi': 'Delete personal deck',
  'Deck belum dapat dihapus. Coba lagi.':
      'Could not delete this deck. Try again.',
  'File belum dapat disimpan atau dibagikan. Coba lagi.':
      'Could not save or share the file. Try again.',
  'Media belum dapat dimuat.': 'Could not load media.',
  'Hentikan audio': 'Stop audio',
  'Memuat audio…': 'Loading audio…',
  'Putar audio': 'Play audio',
  'Audio belum dapat diputar. Periksa koneksi lalu coba lagi.':
      'Could not play audio. Check your connection and try again.',
  'Masuk untuk mengimpor atau mengekspor deck.':
      'Sign in to import or export a deck.',
  'Akun sudah berubah. Buka kembali halaman ini.':
      'Your account has changed. Reopen this screen.',
  'Pilih file dengan ukuran maksimal 20 MB.':
      'Choose a file no larger than 20 MB.',
  'Masuk untuk mengelola deck pribadi.': 'Sign in to manage personal decks.',
  'Fitur transfer belum tersedia di server, atau deck tidak ditemukan.':
      'Transfer is unavailable on the server, or the deck was not found.',
  'File terlalu besar. Batas unggahan 20 MB.':
      'File too large. The upload limit is 20 MB.',
  'Format file deck tidak didukung atau isinya tidak valid.':
      'Unsupported deck format or invalid contents.',
  'Pemrosesan deck sedang penuh. Coba lagi beberapa detik.':
      'Deck processing is busy. Try again shortly.',
  'Transfer belum selesai. Periksa koneksi dan coba lagi.':
      'Transfer has not finished. Check your connection and try again.',
  'Pilih APKG dari Anki, CSV, atau JSON Nihongo Master. Maksimal 20 MB dan 10.000 kartu per deck. File dipratinjau dahulu sebelum disimpan.':
      'Choose an Anki APKG, CSV, or Nihongo Master JSON file. Maximum 20 MB and 10,000 cards per deck. Preview the file before saving.',
  'Deck pribadi disimpan pada akunmu. Deck bawaan tidak ditimpa. Jadwal review dimulai baru; progres dari Anki tidak dipindahkan.':
      'Personal decks are saved to your account without replacing built-in decks. Review schedules start fresh; Anki progress is not transferred.',
  'CSV memakai kolom front, answers, meaning, imageUrl, audio. Answers dapat berupa teks atau array JSON. APKG/JSON membawa media yang dibundel; URL media eksternal tetap berupa tautan. Template HTML diubah menjadi teks. Deck cloze atau template yang tidak didukung akan ditolak dengan penjelasan.':
      'CSV columns: front, answers, meaning, imageUrl, audio. Answers accepts text or a JSON array. APKG/JSON includes bundled media; external media URLs remain links. HTML is converted to text. Cloze or unsupported card templates are rejected with an explanation.',
  'Export berisi semua kartu, bukan hanya sesi 20 kartu. Jadwal review tidak disertakan. Gambar/audio yang dibundel ikut APKG/JSON; media eksternal tetap berupa tautan. Maksimal 10.000 kartu.':
      'Export includes all cards, up to 10,000 per deck, without review schedules. APKG/JSON includes bundled images/audio; external media remains linked.',
  'Kartu, media, dan jadwal review deck ini akan dihapus dari akunmu. Simpan export terlebih dahulu jika masih diperlukan.':
      'This deck’s cards, media, and review schedules will be deleted from your account. Export a backup first if needed.',

  'Akun': 'Account',
  'Akun sudah berubah.': 'The account has changed.',
  'Bahasa aplikasi': 'App language',
  'Atur bantuan membaca di seluruh aplikasi.':
      'Adjust reading aids throughout the app.',
  'Akses materi offline dan raih pencapaian belajarmu.':
      'Access lessons offline and earn learning achievements.',
  'Achievement': 'Achievements',
  'Achievement berhasil dibuka': 'Achievement unlocked',
  'Aktifkan notifikasi': 'Turn on notifications',
  'Anki · Flashcards': 'Anki · Flashcards',
  'Apa yang ingin kamu pelajari?': 'What would you like to study?',
  'Bahasa otomatis': 'Auto-detect',
  'Bahasa Inggris': 'English',
  'Bahasa Jepang': 'Japanese',
  'Bahasa pencarian': 'Search language',
  'Bahasa soal Indonesia': 'Questions in Indonesian',
  'Bahasa Indonesia': 'Indonesian',
  'Belajar': 'Learn',
  'Latihan ujian': 'Exam practice',
  'Lihat semua': 'See all',
  'Simulasi JLPT, JFT, dan NAT': 'JLPT, JFT, and NAT simulations',
  'Jelajahi Jepang': 'Explore Japan',
  'Selamat datang!': 'Welcome!',
  'Jelajahi materi gratis, lalu masuk untuk menyimpan progress.':
      'Explore free lessons, then sign in to save your progress.',
  'Offline · data terakhir tersimpan. Materi yang sudah disiapkan tetap bisa dibuka.':
      'Offline · showing saved data. Prepared lessons are still available.',
  'Kosakata': 'Vocabulary',
  'Tata bahasa': 'Grammar',
  'Kanji': 'Kanji',
  'Membaca': 'Reading',
  'Mendengarkan': 'Listening',
  'Berbicara': 'Speaking',
  'Belajar Kapan Saja': 'Learn Anytime',
  'Belajar dari N5 hingga N1 dengan metode interaktif.':
      'Study from N5 to N1 with interactive lessons.',
  'Belajar flashcard': 'Flashcard study',
  'Belajar offline & penyimpanan': 'Offline learning & storage',
  'Belajar, baca, latihan, dan temukan peluang dalam satu akses.':
      'Study, read, practise, and find opportunities in one place.',
  'Belum ada data': 'No data yet',
  'Bersihkan pencarian': 'Clear search',
  'Beranda': 'Home',
  'Beranda belum dapat dimuat': 'Home could not be loaded',
  'Kemarin': 'Yesterday',
  'Kembali ke materi yang terakhir kamu buka.':
      'Return to the lesson you last opened.',
  'Materi ini sudah tidak tersedia. Pilih materi lain di kursusnya.':
      'This lesson is no longer available. Choose another lesson in its course.',
  'Buka kursus': 'Open course',
  'Buka akses belajar': 'Unlock learning access',
  'Pelajari lagi': 'Study again',
  'Sebelumnya kamu belajar': 'Previously studied',
  'Materi tidak tersedia': 'Lesson unavailable',
  'Akses Lifetime diperlukan': 'Lifetime Access required',
  'Selesai': 'Completed',
  'Cari di kamus': 'Search dictionary',
  'Cari kata. Pahami maknanya.': 'Find words. Understand their meaning.',
  'Cari dengan Jepang, romaji, atau Inggris':
      'Search in Japanese, romaji, or English',
  'Cari kata, kanji, atau romaji': 'Search words, kanji, or romaji',
  'Cerita seru.\nBahasa aslinya.': 'Great stories.\nIn the original language.',
  'Coba lagi': 'Try again',
  'Data Jisho.org · Arti dalam bahasa Inggris · Gratis':
      'Jisho.org data · English definitions · Free',
  'Data Jisho.org • Arti dalam bahasa Inggris • Gratis':
      'Jisho.org data • English definitions • Free',
  'Dengarkan instruksi': 'Listen to the instructions',
  'Daftar berita': 'Article list',
  'Berita lainnya': 'More articles',
  'Berita Jepang': 'Japanese news',
  '日本のニュース': 'Japanese News',
  'Berita · Reguler': 'News · Regular',
  'Info Jepang · Easy': 'Japan brief · Easy',
  'Baca Jepang, temukan kata baru.': 'Read Japanese and discover new words.',
  'Baca': 'Read',
  'Baca berita': 'Read news',
  'Terbaru': 'Latest',
  'Latih membaca dengan berita asli. Pilih kata untuk melihat artinya.':
      'Practise reading with authentic news. Select a word to see its meaning.',
  'menunggu sinkronisasi': 'Waiting to sync',
  'Berita dunia': 'World news',
  'Tanggal mengikuti terbitan sumber. Berita reguler memakai teks Jepang asli; info Easy memakai bahasa yang lebih sederhana.':
      'Dates follow the source publication. Regular articles use original Japanese; Easy articles use simpler language.',
  'Sebelumnya': 'Previous',
  'Berikutnya': 'Next',
  'Halaman': 'Page',
  'dari': 'of',
  'Terbit': 'Published',
  'Tanggal sumber': 'Source date',
  'Disinkronkan': 'Synced',
  'Matikan furigana': 'Hide furigana',
  'Tampilkan furigana': 'Show furigana',
  'Pilih kata atau kalimat, lalu tekan “Cari arti”.':
      'Select a word or sentence, then tap “Look up meaning.”',
  'Furigana dapat dibuat otomatis dan bisa keliru, terutama pada nama.':
      'Furigana may be generated automatically and can be incorrect, especially for names.',
  'Sumber & kredit': 'Sources & credits',
  'Artikel asli': 'Original article',
  'Sumber gambar': 'Image source',
  'Lisensi gambar': 'Image license',
  'Referensi dalam artikel': 'References in this article',
  'Pilih kata untuk melihat arti. Bentuk dasar dipakai untuk mencari kata yang berubah bentuk.':
      'Select a word to see its meaning. Search uses the base form for inflected words.',
  'Arti belum dapat dimuat. Periksa koneksi dan coba kembali.':
      'Meanings could not be loaded. Check your connection and try again.',
  'Berita belum dapat dimuat. Periksa koneksi dan coba kembali.':
      'News could not be loaded. Check your connection and try again.',
  'Belum ada artikel tersedia. Berita akan muncul setelah sinkronisasi sumber berhasil.':
      'No articles are available yet. News will appear after a source sync succeeds.',
  'Detail kanji': 'Kanji details',
  'Dukungan kehidupan': 'Daily living support',
  'Lihat': 'View',
  'Materi N3–N1 yang tersedia': 'Available N3–N1 materials',
  'Seluruh tes yang tersedia': 'All available tests',
  'Sekali bayar tanpa perpanjangan': 'One-time payment with no renewal',
  'JLPT N5–N4 dasar': 'Basic JLPT N5–N4',
  'Progress dasar': 'Basic progress',
  'Streak hari': 'Day streak',
  'Badge': 'Badges',
  '50 XP saat pertama lulus lesson. 100 XP saat pertama mendapat ≥80% pada tiap jenis ujian. Ulangan tidak menggandakan XP. Streak dihitung dari kuis atau ujian yang selesai, zona waktu belajar yang dipilih.':
      'Earn 50 XP the first time you pass a lesson. Earn 100 XP the first time you score at least 80% on each exam type. Retakes do not award XP again. Streaks count completed quizzes and exams using your selected study timezone.',
  'Progres belum bisa dimuat. Periksa koneksi atau masuk kembali jika sesi berakhir.':
      'Progress could not be loaded. Check your connection or sign in again if your session has expired.',
  'First Step': 'First Step',
  'Pelajar Tekun': 'Dedicated Learner',
  '50 Lesson': '50 Lessons',
  'Word Collector': 'Word Collector',
  'Grammar Builder': 'Grammar Builder',
  'Kanji Explorer': 'Kanji Explorer',
  'Bookworm': 'Bookworm',
  'Serba Bisa': 'All-Rounder',
  '7-Day Streak': '7-Day Streak',
  '30-Day Streak': '30-Day Streak',
  'Level Up': 'Level Up',
  'Exam Ready': 'Exam Ready',
  'Lulus satu kuis lesson dengan nilai minimal 80%':
      'Pass one lesson quiz with a score of at least 80%.',
  'Lulus 10 lesson berbeda': 'Pass 10 different lessons.',
  'Lulus 50 lesson berbeda': 'Pass 50 different lessons.',
  'Lulus 10 lesson kosakata': 'Pass 10 vocabulary lessons.',
  'Lulus 10 lesson grammar': 'Pass 10 grammar lessons.',
  'Lulus 10 lesson kanji': 'Pass 10 kanji lessons.',
  'Lulus 5 lesson reading': 'Pass 5 reading lessons.',
  'Lulus lesson di keempat bidang': 'Pass lessons in all four subjects.',
  'Selesaikan kuis atau ujian selama 7 hari berturut-turut':
      'Complete a quiz or exam for 7 consecutive days.',
  'Selesaikan kuis atau ujian selama 30 hari berturut-turut':
      'Complete a quiz or exam for 30 consecutive days.',
  'Kumpulkan 500 XP': 'Earn 500 XP.',
  'Dapatkan minimal 80% pada diagnostic atau simulasi':
      'Score at least 80% on a diagnostic or simulation.',
  'Exam': 'Exam',
  'Furigana': 'Furigana',
  'Gagal memuat data': 'Could not load data',
  'Gamification': 'Gamification',
  'Gunakan bahasa perangkat': 'Use device language',
  'Hapus': 'Clear',
  'Hari ini': 'Today',
  'Hasil': 'Results',
  'Home': 'Home',
  'Inggris': 'English',
  'Jepang': 'Japanese',
  'Jepang, romaji, atau Inggris': 'Japanese, romaji, or English',
  'Kamus': 'Dictionary',
  'Kamus Jepang': 'Japanese Dictionary',
  'Mau cari kata apa hari ini?': 'What would you like to look up today?',
  'Coba salah satu kata untuk mulai menjelajah.':
      'Try one of these words to get started.',
  'Filter pencarian': 'Search filters',
  'Filter': 'Filters',
  'Jepang / romaji': 'Japanese / romaji',
  'Favorit tersimpan di perangkat': 'Favorites saved on this device',
  'Tutup filter': 'Close filters',
  'Kata': 'Word',
  'Kata dan arti disalin.': 'Word and meaning copied.',
  'Kata umum': 'Common words',
  'Kata umum saja': 'Common words only',
  'Kata untuk diingat': 'Words to remember',
  'Kata belum ditemukan di kamus. Nama orang atau tempat tidak selalu tersedia.':
      'Word not found in the dictionary. Names of people and places may not be available.',
  'Kata belum ditemukan. Coba bentuk dasar, kana, romaji, atau kata Inggris; ubah juga filter JLPT.':
      'Word not found. Try its base form, kana, romaji, or English spelling, or change the JLPT filter.',
  'Kuasai Bahasa Jepang': 'Master Japanese',
  'Panduan aplikasi': 'App guide',
  'Lewati': 'Skip',
  'Tutup panduan': 'Close guide',
  'Kembali ke aplikasi': 'Back to app',
  'PERJALANAN BARU': 'A NEW JOURNEY',
  'LANGKAH YANG TERARAH': 'A CLEAR WAY FORWARD',
  'SEDIKIT, SETIAP HARI': 'A LITTLE, EVERY DAY',
  'Dari kata pertama sampai percakapan yang lebih percaya diri. Temukan ritme belajarmu, dari N5 hingga N1.':
      'From your first word to more confident conversations. Find your learning rhythm, from N5 to N1.',
  'Siap untuk\nujian berikutnya.': 'Ready for\nyour next exam.',
  'Pahami materinya, coba latihan soalnya, lalu ukur kemampuanmu lewat simulasi ujian.':
      'Learn the material, practise with questions, then measure your skills with mock exams.',
  'Belajarmu ikut\nke mana pun.': 'Your learning,\nwherever you go.',
  'Buka lagi materi yang sudah tersimpan, bahkan saat offline. Lanjutkan perjalananmu, satu langkah setiap hari.':
      'Revisit saved lessons, even offline. Keep moving forward, one step every day.',
  'Kosakata & kanji': 'Vocabulary & kanji',
  'Materi tersimpan': 'Saved lessons',
  'Progres belajar': 'Learning progress',
  'Panduan belum dapat diselesaikan. Coba lagi.':
      'Could not finish the guide. Try again.',
  'Ilustrasi buku belajar, sakura, dan perjalanan menuju Gunung Fuji':
      'Study book, cherry blossoms, and a path towards Mount Fuji',
  'Ilustrasi meja belajar, lembar ujian, dan medali pencapaian':
      'Study desk, exam paper, and achievement medal',
  'Ilustrasi buku dan kartu belajar di perjalanan kereta Jepang':
      'Book and study cards on a Japanese train journey',
  'Latihan Ujian': 'Exam Practice',
  'Lihat paket': 'View plans',
  'Lanjutkan': 'Continue',
  'Lanjutkan belajar': 'Continue learning',
  'Lanjutkan membaca': 'Continue reading',
  'Lanjutkan simulasi': 'Continue simulation',
  'Lanjutkan sesi': 'Continue session',
  'Lanjutkan versi ini': 'Continue with this version',
  'Level': 'Level',
  'Loading...': 'Loading...',
  'Masuk': 'Sign in',
  'Masuk untuk mulai belajar': 'Sign in to start learning',
  'Materi belajar': 'Learning materials',
  'Materi khusus untuk JLPT, NAT, dan JFT-Basic.':
      'Focused preparation for JLPT, NAT, and JFT-Basic.',
  'Mulai Belajar': 'Start Learning',
  'Mulai belajar': 'Start learning',
  'Mulai kuis': 'Start quiz',
  'Mulai membaca': 'Start reading',
  'Mulai review': 'Start review',
  'Mulai simulasi': 'Start simulation',
  'Mulai dari kata yang ingin kamu pahami.':
      'Start with a word you want to understand.',
  'Mulai Exam': 'Start Exam',
  'Nihongo Master': 'Nihongo Master',
  'N5–N4 gratis. Materi N3–N1 memerlukan Lifetime Access.':
      'N5–N4 are free. N3–N1 materials require Lifetime Access.',
  'Notifikasi': 'Notifications',
  'Offline': 'Offline',
  'Pengaturan': 'Settings',
  'Pengaturan tampilan': 'Display settings',
  'Pengaturan notifikasi': 'Notification settings',
  'Pencarian terakhir': 'Recent searches',
  'Pilih Jepang untuk romaji, atau Inggris untuk mencari arti.':
      'Choose Japanese to search with romaji, or English to search by meaning.',
  'Level JLPT': 'JLPT level',
  'Semua level': 'All levels',
  'Tampilkan kata yang sering digunakan.': 'Show commonly used words.',
  'Reset filter': 'Reset filters',
  'Terapkan filter': 'Apply filters',
  'Buka Jisho': 'Open Jisho',
  'Belum ada kata tersimpan. Ketuk ikon bookmark pada hasil pencarian.':
      'No saved words yet. Tap the bookmark icon on a search result to save it.',
  'Muat kata berikutnya': 'Load more words',
  'Memuat…': 'Loading…',
  'Riwayat belum dapat dihapus.': 'Search history could not be cleared.',
  'Sumber dan bantuan': 'Sources and help',
  'Pilih bahasa aplikasi': 'Choose app language',
  'Pilih level JLPT': 'Choose a JLPT level',
  'Pilih bahasa yang ingin digunakan untuk menu dan petunjuk aplikasi.':
      'Choose the language for app menus and instructions.',
  'Pilih level': 'Choose a level',
  'Profil': 'Profile',
  'Progress lesson sudah tersimpan. Lanjutkan ke lesson berikutnya.':
      'Lesson progress is saved. Continue to the next lesson.',
  'Progres cerita sudah tersimpan. Lanjutkan membaca cerita berikutnya.':
      'Story progress is saved. Continue reading the next story.',
  'Pulihkan pembelian': 'Restore purchases',
  'Pencapaian': 'Achievements',
  'Semua': 'All',
  'Segera': 'Coming soon',
  'Selanjutnya': 'Next',
  'Selesaikan pelajaran': 'Complete a lesson',
  'Sumber data & lisensi': 'Data sources & licenses',
  'Terjemahan': 'Translation',
  'Tampilkan arti contoh dan bacaan.':
      'Show meanings for examples and reading passages.',
  'Tampilkan cara baca di atas kanji.': 'Show readings above kanji.',
  'Target harian': 'Daily goal',
  'TARGET YANG JELAS': 'CLEAR GOALS',
  'TEMAN BELAJARMU': 'YOUR STUDY COMPANION',
  'Ujian': 'Exams',
  'Uji pemahaman': 'Check understanding',
  'Vocabulary API': 'Vocabulary API',
  'Kosakata, contoh, kuis & progress':
      'Vocabulary, examples, quizzes & progress',
  'Pola, contoh kalimat, kuis & progress':
      'Patterns, example sentences, quizzes & progress',
  'Arti, bacaan, contoh kata & kuis':
      'Meanings, readings, example words & quizzes',
  'Cerita bergambar, kosakata & kuis':
      'Illustrated stories, vocabulary & quizzes',
  'Pelajari kata, contoh, lalu uji pemahamanmu':
      'Study words and examples, then check your understanding',
  'Waktu belajar': 'Study time',
  'Kurikulum, kasus kerja, latihan per topik, dan simulasi belajar tersedia untuk':
      'Curriculum, workplace scenarios, topic practice, and study simulations are available for',
  'Pilih bahasa': 'Choose language',
  'Bahasa yang dipilih hanya disimpan di perangkat ini.':
      'Your language choice is saved on this device only.',
  'Ilustrasi perjalanan belajar bahasa Jepang':
      'Illustration of a Japanese learning journey',
  'Ilustrasi persiapan ujian bahasa Jepang':
      'Illustration of Japanese exam preparation',
  'Ilustrasi belajar secara offline kapan saja':
      'Illustration of learning offline anytime',
  'Lewati (Skip)': 'Skip',
  'Persiapan Ujian\nTepat Sasaran': 'Focused Exam\nPreparation',
  'Cerita seru. Bahasa aslinya.': 'Great stories in their original language.',
  'Satu timer': 'One timer',
  'Kamu yakin ingin keluar?': 'Are you sure you want to exit?',
  'Ubah bahasa aplikasi': 'Change app language',
  'Kata tersimpan': 'Saved words',
  'Kembali ke hasil': 'Back to results',
  'Hasil pencarian': 'Search results',
  'Tidak ada hasil': 'No results',
  'Coba kata lain': 'Try another word',
  'Koneksi internet diperlukan.': 'An internet connection is required.',
  'Periksa koneksi atau coba lagi sebentar.':
      'Check your connection or try again shortly.',
  'Kamus belum dapat diakses. Periksa koneksi atau coba lagi sebentar.':
      'The dictionary is unavailable. Check your connection or try again shortly.',
  'Belajar di mana saja': 'Learn anywhere',
  'Ulangi': 'Try again',
  'Tutup': 'Close',
  'Batal': 'Cancel',
  'Simpan': 'Save',
  'Hapus semua': 'Clear all',
  'Berhasil': 'Done',
  'Gagal': 'Failed',
  'Penuh': 'Full',
  'Gratis': 'Free',
  'Premium': 'Premium',
  'Terkunci': 'Locked',
  'Buka materi': 'Open lessons',
  'Kerja di Jepang': 'Work in Japan',
  'Berita': 'News',
  'Pekerjaan': 'Jobs',
  'Baca manga': 'Read manga',
  'Kamus bacaan': 'Reading dictionary',
  'Kuis': 'Quiz',
  'Pola berikutnya': 'Next pattern',
  'Kanji berikutnya': 'Next kanji',
  'Cerita berikutnya': 'Next story',
  'Mulai diagnostic': 'Start diagnostic',
  'Lanjutkan diagnostic N5': 'Continue N5 diagnostic',
  'Lanjutkan Exam': 'Continue Exam',
  'Selesaikan pelajaran untuk mendapatkan XP.': 'Complete lessons to earn XP.',
  'Atribusi & lisensi data': 'Data attribution & licenses',
  'Tautan belum dapat dibuka.': 'This link could not be opened.',
  'Bentuk dan bacaan lain': 'Other forms and readings',
  'Kata disalin.': 'Word copied.',
  'Gagal menyimpan atau favorit sudah mencapai batas 60 kata.':
      'Could not save this word, or your 60-word favorites limit has been reached.',
  'Tutup detail': 'Close details',
  'Tutup kanji': 'Close kanji details',
  'Tutup bantuan': 'Close help',
  'Salin': 'Copy',
  'Lihat di Jisho': 'View on Jisho',
  'Contoh kalimat di Jisho': 'Example sentences on Jisho',
  'Sumber: Jisho.org. Makna dan urutan mengikuti sumber; bukan terjemahan otomatis.':
      'Source: Jisho.org. Meanings and ordering follow the source; they are not machine-translated.',
  'Lawan kata': 'Antonyms',
  'Lihat juga': 'See also',
  'Detail kanji belum dapat dimuat dari sumber kamus.':
      'Kanji details could not be loaded from the dictionary source.',
  'Tidak tersedia': 'Not available',
  'Label JLPT kanji berasal dari klasifikasi lama KANJIDIC dan berbeda dengan label N1–N5 pada kosakata.':
      'Kanji JLPT labels come from an older KANJIDIC classification and differ from the N1–N5 labels used for vocabulary.',
  'Radikal & urutan goresan di Jisho': 'Radicals & stroke order on Jisho',
  'Sumber: kanjiapi.dev / KANJIDIC2 (EDRDG).':
      'Source: kanjiapi.dev / KANJIDIC2 (EDRDG).',
  'Tentang kamus': 'About the dictionary',
  'Cari dengan kanji, hiragana, katakana, romaji, atau kata Inggris. Gunakan filter JLPT dan kata umum. Arti disajikan dalam bahasa Inggris sesuai sumber.':
      'Search with kanji, hiragana, katakana, romaji, or English words. Filter by JLPT level and common words. Definitions are shown in English as provided by the source.',
  'Kata: Jisho.org, menggunakan JMdict/JMnedict dari Electronic Dictionary Research and Development Group (EDRDG), serta DBpedia untuk makna ensiklopedia. Detail kanji: kanjiapi.dev, menggunakan KANJIDIC2. Data EDRDG digunakan sesuai lisensinya (CC BY-SA). DBpedia memiliki ketentuan atribusi dan berbagi serupa.':
      'Words: Jisho.org, using JMdict/JMnedict from the Electronic Dictionary Research and Development Group (EDRDG), and DBpedia for encyclopedia definitions. Kanji details: kanjiapi.dev, using KANJIDIC2. EDRDG data is used under its license (CC BY-SA). DBpedia has similar attribution and share-alike requirements.',
  'Favorit menyimpan salinan entri di perangkat, terpisah untuk setiap akun. Favorit dapat dibaca tanpa internet; pencarian dan detail kanji memerlukan koneksi. Riwayat bisa dihapus. Maksimal 60 favorit dan 20 pencarian terakhir.':
      'Favorites save a copy of each entry on this device, separately for each account. Saved words are available offline; search and kanji details require a connection. Search history can be cleared. You can save up to 60 favorites and keep 20 recent searches.',
  'Kamus ini memakai sumber gratis tanpa API key. Ketersediaannya mengikuti layanan sumber. Contoh kalimat lengkap, pencarian radikal, dan diagram goresan dibuka di Jisho. Nihongo Master tidak berafiliasi dengan Jisho.org.':
      'This dictionary uses free sources without an API key. Availability depends on those services. Full example sentences, radical search, and stroke diagrams open on Jisho. Nihongo Master is not affiliated with Jisho.org.',
  'Jisho.org & sumber data': 'Jisho.org & data sources',
  'EDRDG & lisensi kamus': 'EDRDG & dictionary license',
  'Zona waktu belajar': 'Study time zone',
  'Zona waktu akun': 'Account time zone',
  'Pilih zona waktu': 'Choose time zone',
  'Zona waktu ini menentukan pergantian hari untuk streak dan progress belajar.':
      'This time zone sets the day boundary for streaks and study progress.',
  'Perubahan zona waktu hanya berlaku setelah kamu menyimpannya.':
      'A time zone change takes effect only after you save it.',
  'Masuk untuk mengatur zona waktu belajar.':
      'Sign in to manage your study time zone.',
  'Zona waktu belajar belum dapat dimuat.':
      'Study time zone could not be loaded.',
  'Simpan zona waktu': 'Save time zone',
  'Zona waktu belajar berhasil diperbarui.': 'Study time zone updated.',
  'Zona waktu belajar gagal diperbarui.': 'Could not update study time zone.',
  'Harga belum tersedia': 'Price unavailable',
  'Memeriksa harga store…': 'Checking store price…',
  'Harga store belum tersedia': 'Store price unavailable',
  'Lebih jauh dari sekadar belajar': 'Go beyond learning Japanese',
  'Dari persiapan JLPT sampai langkah kariermu di Jepang.':
      'From JLPT preparation to your career in Japan.',
  'Sekali bayar. Tanpa biaya bulanan.': 'One payment. No monthly fees.',
  'Akses dengan akun aplikasi yang sama di Android dan iOS':
      'Use access with the same app account on Android and iOS',
  'Tidak ada perpanjangan otomatis': 'No automatic renewal',
  'Harga akhir mengikuti store dan wilayah akun Anda':
      'The final price is set by your store and billing region',
  'Masuk untuk melanjutkan': 'Sign in to continue',
  'Jelajahi akses saya': 'Explore my access',
  'Akses aktif': 'Access active',
  'Menunggu pembayaran': 'Payment pending',
  'Buka Lifetime Access': 'Get Lifetime Access',
  'Restore pembelian': 'Restore purchases',
  'Coba verifikasi lagi': 'Try verification again',
  'Materi N5–N4 dan tes N5 tetap gratis. Kuis di dalam lesson mengikuti akses materi.':
      'N5–N4 materials and N5 tests remain free. Quizzes inside lessons follow material access.',
  'Tentang pembelianmu': 'About your purchase',
  'Pembayaran yang dibatalkan atau tertunda tidak mengaktifkan akses. Refund atau pencabutan oleh store mengunci premium kembali. Menghapus akun tidak mengembalikan uang; pembelian dapat dipulihkan dengan akun store pembeli setelah masuk ke akun aplikasi baru.':
      'Canceled or pending payments do not activate access. A refund or revocation by the store locks premium access again. Deleting your account does not issue a refund; you can restore a purchase with the purchasing store account after signing into a new app account.',
  'Bantuan refund Google Play': 'Google Play refund help',
  'Bantuan refund App Store': 'App Store refund help',
  'Tautan bantuan tidak dapat dibuka.': 'Could not open the help link.',
  'Perjalanan belajar bahasa dan meraih peluang di Jepang':
      'Study Japanese and find opportunities in Japan',
  'LIFETIME AKTIF': 'LIFETIME ACTIVE',
  'SEKALI BAYAR · AKSES SELAMANYA': 'ONE PAYMENT · LIFETIME ACCESS',
  'Lifetime Access aktif': 'Lifetime Access is active',
  'Buka lebih banyak.\nMelangkah lebih jauh.': 'Unlock more.\nGo further.',
  'Materi JLPT lanjutan': 'Advanced JLPT materials',
  'Buka materi N3, N2, dan N1 yang tersedia beserta kuis pembelajarannya.':
      'Unlock available N3, N2, and N1 materials with their learning quizzes.',
  'Latihan sampai siap ujian': 'Practise until you are exam-ready',
  'Akses seluruh tes yang tersedia, termasuk simulasi JLPT, JFT, NAT, dan SSW.':
      'Access all available tests, including JLPT, JFT, NAT, and SSW simulations.',
  'Baca Manga Jepang': 'Read Japanese manga',
  'Nikmati koleksi manga berbahasa Jepang dan latih membaca lewat cerita.':
      'Explore a collection of Japanese manga and practise reading through stories.',
  'Lanjut ke peluang kerja': 'Take the next step toward a job',
  'Buka tautan lamaran lowongan aktif dan lanjutkan prosesnya di situs sumber.':
      'Open applications for active listings and continue on the source website.',
  'Lifetime Access': 'Lifetime Access',
  'Koneksi store terputus. Coba lagi.': 'Store connection lost. Try again.',
  'Tidak dapat memeriksa pembelian lama. Gunakan Restore.':
      'Could not check previous purchases. Use Restore.',
  'Masuk ke akun untuk membeli atau restore.':
      'Sign in to buy or restore purchases.',
  'Memeriksa akses dan store…': 'Checking access and store…',
  'Pembelian store belum tersedia di perangkat ini.':
      'Store purchases are not available on this device.',
  'Produk belum tersedia. Coba lagi nanti.':
      'Product unavailable. Try again later.',
  'Masuk dengan akun Google pribadi untuk membeli atau Restore. Akun demo digunakan bersama.':
      'Sign in with a personal Google account to buy or restore. Demo accounts are shared.',
  'Sesi atau koneksi bermasalah. Masuk lagi jika sesi sudah berakhir.':
      'There is a session or connection issue. Sign in again if your session has expired.',
  'Tidak dapat memeriksa akses. Periksa koneksi lalu coba lagi.':
      'Could not check access. Check your connection and try again.',
  'Selesaikan pembayaran di store.': 'Complete your payment in the store.',
  'Store belum memberikan hasil. Gunakan Restore atau coba verifikasi lagi sebelum membeli ulang.':
      'The store has not returned a result. Restore or verify again before trying to buy once more.',
  'Google Play tidak melanjutkan pembayaran. Jika dialog tadi ditutup, tidak ada biaya yang dikenakan dan akses tetap sama.':
      'Google Play did not continue the payment. If you closed the dialog, you were not charged and your access is unchanged.',
  'Pembayaran tidak dapat dimulai. Coba Restore jika sudah membeli.':
      'Could not start the payment. Use Restore if you have already purchased.',
  'Memulihkan pembelian…': 'Restoring purchases…',
  'Lifetime Access aktif.': 'Lifetime Access is active.',
  'Belum ada pembelian yang berhasil dipulihkan. Pastikan akun store benar.':
      'No purchase could be restored. Check that you are using the correct store account.',
  'Restore gagal. Periksa akun store dan koneksi, lalu coba lagi.':
      'Restore failed. Check your store account and connection, then try again.',
  'Pembelian belum selesai. Coba verifikasi lagi atau restore.':
      'The purchase is not complete. Try verifying or restoring again.',
  'Pembayaran dibatalkan. Tidak ada biaya yang dikenakan dan akses tidak berubah.':
      'Payment was canceled. You were not charged and your access is unchanged.',
  'Google Play tidak dapat memulai pembayaran. Akses tidak berubah; silakan coba lagi.':
      'Google Play could not start the payment. Your access is unchanged; please try again.',
  'Pembayaran menunggu persetujuan. Akses aktif setelah terverifikasi.':
      'Payment is awaiting approval. Access will activate after verification.',
  'Pembayaran gagal. Coba lagi atau Restore jika sudah membeli.':
      'Payment failed. Try again or use Restore if you have already purchased.',
  'Memverifikasi pembelian…': 'Verifying purchase…',
  'Pembayaran masih tertunda di store.':
      'Payment is still pending in the store.',
  'Lifetime Access berhasil diaktifkan.': 'Lifetime Access activated.',
  'Pembelian telah dicabut atau dikembalikan oleh store.':
      'The store revoked or refunded the purchase.',
  'Pembelian ini terkait akun aplikasi lain. Masuk ke akun pembeli lalu Restore.':
      'This purchase is linked to another app account. Sign in to the purchasing account and restore it.',
  'Verifikasi belum selesai. Jangan membeli lagi; coba verifikasi atau Restore.':
      'Verification is not complete. Do not buy again; try verifying or restoring.',

  // Jobs: app-owned navigation, filters, and availability messages.
  'LANGKAH BERIKUTNYA UNTUK KARIERMU': 'YOUR NEXT CAREER STEP',
  'Temukan peluang sesuai bidang, lokasi, gaji, dan kualifikasimu':
      'Find opportunities by field, location, salary, and qualifications',
  'Sumber lowongan': 'Job sources',
  'Posisi, perusahaan, kata kunci…': 'Role, company, or keyword…',
  'Hapus pencarian': 'Clear search',
  'Filter lowongan': 'Filter jobs',
  'Muat ulang pilihan filter': 'Reload filter options',
  'Mencari peluang…': 'Searching opportunities…',
  'Belum ada lowongan yang cocok. Ubah pencarian atau filter. Sumber baru juga dapat ditambahkan secara bertahap.':
      'No matching jobs. Try changing your search or filters. More sources may be added over time.',
  'Muat lebih banyak': 'Load more',
  'Seluruh hasil telah ditampilkan. Data diperbarui berkala dari sumber.':
      'You have reached the end of the results. Listings are refreshed from their sources periodically.',
  'Tidak aktif': 'Inactive',
  'Temukan yang sesuai': 'Find the right fit',
  'Kategori pekerjaan': 'Job category',
  'Bidang SSW': 'SSW sector',
  'Prefektur': 'Prefecture',
  'Jenis pekerjaan': 'Employment type',
  'Dukungan visa': 'Visa support',
  'Cara bekerja': 'Work arrangement',
  'Di tempat kerja': 'On-site',
  'Gaji minimum (JPY)': 'Minimum salary (JPY)',
  'Contoh: 250000': 'Example: 250000',
  'Periode gaji': 'Salary period',
  'Pilih periode': 'Choose a period',
  'Per bulan': 'Per month',
  'Per tahun': 'Per year',
  'Per jam': 'Per hour',
  'Per hari': 'Per day',
  'Per minggu': 'Per week',
  'Filter gaji hanya berlaku pada angka dan periode yang dicantumkan sumber.':
      'Salary filters only apply to amounts and periods listed by the source.',
  'Masukkan gaji minimum yang valid dan pilih periodenya.':
      'Enter a valid minimum salary and choose its period.',
  'Tampilkan lowongan': 'Show jobs',
  'Reset semua filter': 'Reset all filters',
  'Detail lowongan': 'Job details',
  'Lowongan ini sudah tidak aktif atau kedaluwarsa. Lamaran tidak tersedia.':
      'This listing is inactive or has expired. Applications are unavailable.',
  'Lokasi belum dicantumkan': 'Location not listed',
  'Bidang': 'Field',
  'Sektor': 'Sector',
  'Visa': 'Visa',
  'Tentang pekerjaan': 'About the job',
  'Kualifikasi': 'Qualifications',
  'Benefit & tunjangan': 'Benefits and allowances',
  'Benefit belum dicantumkan oleh sumber lowongan. Konfirmasikan tunjangan, fasilitas, dan ketentuannya kepada perekrut sebelum melamar.':
      'The job source has not listed benefits. Confirm allowances, facilities, and terms with the recruiter before applying.',
  'Tanggal publikasi': 'Posted',
  'Diperiksa oleh sistem': 'Last checked',
  'Batas lowongan': 'Listing expires',
  'Informasi mengikuti sumber aslinya dan dapat berubah. Baca persyaratan lengkap di situs penyedia sebelum melamar.':
      'Listing details may change. Read the full requirements on the provider’s website before applying.',
  'Lamar dengan Lifetime Access': 'Apply with Lifetime Access',
  'Akses tautan lamaran untuk lowongan yang aktif, sekaligus buka materi lanjutan, simulasi ujian, dan Manga Jepang.':
      'Open application links for active jobs, along with advanced lessons, exam simulations, and Japanese manga.',
  'Lamaran diproses oleh situs sumber. Premium tidak menjamin diterima kerja atau persetujuan visa.':
      'Applications are handled by the source website. Premium does not guarantee a job offer or visa approval.',
  'Sumber & pembaruan': 'Sources & updates',
  'Lowongan disimpan di server kami. Membuka atau mencari di aplikasi tidak mengambil ulang data dari situs penyedia.':
      'Job listings are stored on our server. Viewing or searching in the app does not fetch data again from provider websites.',
  'Direktori · Kunjungi situs untuk lowongan lengkap':
      'Directory · Visit the website for full listings',
  'Pembaruan otomatis': 'Automatic updates',
  'Interval:': 'Update interval:',
  'Berhasil diperbarui:': 'Last updated:',
  'Jadwal berikutnya:': 'Next scheduled update:',
  'Kunjungi situs': 'Visit website',
  'Ketentuan sumber': 'Source terms',
  'Tautan belum dapat dibuka. Coba kembali.':
      'Could not open the link. Please try again.',
  'Belum dicantumkan': 'Not listed',
  'Lihat informasi lengkap di situs sumber.':
      'See the full details on the source website.',
  'Data publik belum dapat dibaca': 'Public data could not be read',
  'Pembaruan tertunda': 'Update delayed',
  'Pengambilan dihentikan': 'Fetching is paused',
  'Sampel lowongan diperbarui': 'Sample listings updated',
  'Menunggu pembaruan': 'Waiting for an update',
  'PELUANG KERJA': 'JOB OPPORTUNITIES',
  'Langkah berikutnya\nmenuju Jepang': 'Your next step\ntoward Japan',
  'Temukan peluang dari berbagai sumber terpercaya.':
      'Explore opportunities from trusted sources.',
  'Jelajahi bidang': 'Explore fields',
  'Ilustrasi pekerja dari berbagai bidang di Jepang':
      'Illustration of workers in different fields in Japan',
  'Lihat & lamar di situs sumber': 'View and apply on source website',
  'Memeriksa akses…': 'Checking access…',
  'Tautan lamaran belum dapat dibuka. Coba lagi.':
      'Could not open the application link. Try again.',
  'Akses premium belum dapat diverifikasi. Periksa koneksi lalu coba lagi.':
      'Could not verify premium access. Check your connection and try again.',
  'Belum dijadwalkan': 'Not scheduled',
  'Data belum dapat dimuat. Coba kembali.':
      'Could not load data. Please try again.',
  'Lowongan belum dapat dimuat. Periksa koneksi dan coba kembali.':
      'Could not load jobs. Check your connection and try again.',
  'Lowongan tidak ditemukan atau sudah dihapus.':
      'Job not found or has been removed.',

  // Manga: translate app controls while leaving Japanese source material intact.
  'Manga Jepang': 'Japanese Manga',
  'Nikmati cerita dan latih membaca Jepang lewat panel aslinya':
      'Enjoy original stories and practise reading Japanese panel by panel',
  'Latih membaca Jepang lewat panel demi panel.':
      'Practise reading Japanese one panel at a time.',
  'Sumber & atribusi ↗': 'Sources & attribution ↗',
  'Cari judul manga…': 'Search manga titles…',
  'Cari': 'Search',
  'Manga belum ditemukan': 'No manga found',
  'Koleksi sedang disiapkan': 'Collection in progress',
  'Judul terkurasi akan tersedia di sini. Coba muat ulang beberapa saat lagi.':
      'Curated titles will appear here. Try refreshing in a little while.',
  'judul untuk dijelajahi': 'titles to explore',
  'Halaman sebelumnya': 'Previous page',
  'Halaman berikutnya': 'Next page',
  'Pilihan tanpa konten dewasa/ecchi. Katalog diperiksa mingguan.':
      'Titles are selected to exclude adult/ecchi content. The catalog is reviewed weekly.',
  'Detail manga': 'Manga details',
  'chapter · Jepang asli': 'chapters · Original Japanese',
  'Jepang asli': 'Original Japanese',
  'Tentang cerita': 'About this story',
  'Daftar chapter': 'Chapter list',
  'Urutan terlama dahulu': 'Oldest first',
  'Urutan terbaru dahulu': 'Newest first',
  'Cari nomor chapter': 'Search chapter number',
  'Chapter tidak ditemukan.': 'No chapters found.',
  'Siap dibaca': 'Ready to read',
  'Sedang disiapkan': 'Preparing',
  'Belum tersedia': 'Unavailable',
  'Siapkan untuk dibaca': 'Prepare to read',
  'Pembaca manga': 'Manga reader',
  'Baca per halaman': 'Read one page at a time',
  'Baca dengan scroll': 'Read with scrolling',
  'Baca dari kanan ke kiri': 'Read from right to left',
  'Chapter sedang disiapkan': 'Chapter is being prepared',
  'Siap menjelajahi cerita?': 'Ready to explore the story?',
  'Halaman akan terbuka setelah semua gambar siap. Kamu juga bisa kembali dan membukanya lagi nanti.':
      'Pages will open when all images are ready. You can also come back later.',
  'Beberapa halaman belum siap. Cek kembali setelah waktu berikutnya.':
      'Some pages are not ready yet. Check back later.',
  'Siapkan chapter ini agar tersedia untuk dibaca.':
      'Prepare this chapter to make it available for reading.',
  'Siapkan chapter': 'Prepare chapter',
  'Muat ulang': 'Refresh',
  'Sumber: Rawkuma': 'Source: Rawkuma',
  'Sumber: Rawkuma ↗': 'Source: Rawkuma ↗',
  'Halaman manga bahasa Jepang': 'Japanese manga page',
  'Gambar belum dapat dimuat': 'Could not load image',
  'Tautan sumber belum dapat dibuka.': 'Could not open the source link.',
  'Manga Jepang khusus pengguna Lifetime Access.':
      'Japanese manga is available to Lifetime Access members.',
  'Lihat Lifetime Access': 'View Lifetime Access',
  'Manga belum dapat dimuat. Periksa koneksi lalu coba kembali.':
      'Could not load manga. Check your connection and try again.',
  'Chapter belum dapat dimuat. Silakan coba lagi.':
      'Could not load the chapter. Please try again.',

  // Anki: translate app controls and study guidance, not deck or card content.
  'REVIEW SEDIKIT, INGAT LEBIH LAMA': 'REVIEW A LITTLE, REMEMBER MORE',
  'Pilih deck favorit dan bangun kebiasaan review kosakata':
      'Choose a deck and build a vocabulary review habit',
  'Sedikit hari ini.\nLebih ingat besok.':
      'A little today.\nRemember more tomorrow.',
  'Kenali topik deck, balik kartu untuk melihat jawaban, lalu nilai seberapa mudah kamu mengingatnya.':
      'Choose a deck topic, flip each card to reveal the answer, then rate how easily you remembered it.',
  'Cari deck': 'Search decks',
  'Deck tidak ditemukan. Coba nama atau kategori lain.':
      'No decks found. Try another name or category.',
  'Data tersimpan · Belum tersedia': 'Saved data · Not available yet',
  'Siap dipelajari': 'Ready to study',
  'Tentang deck': 'About this deck',
  'kartu siap': 'cards ready',
  'Deck ini tentang apa?': 'About this deck',
  'Cara belajar': 'How to study',
  'Satu sesi berisi hingga 20 kartu. Review yang sudah jatuh tempo didahulukan, kemudian kartu baru. Pilihanmu menentukan jadwal berikutnya.':
      'Each session has up to 20 cards. Due reviews come first, followed by new cards. Your ratings determine the next review schedule.',
  'Progress belajarmu': 'Your study progress',
  'Pernah dipelajari': 'Studied',
  'Perlu review': 'Due for review',
  'Kartu baru': 'New cards',
  'Jadwal dan progress tersimpan di akunmu.':
      'Your schedule and progress are saved to your account.',
  'Masuk untuk menyimpan progress dan jadwal review.':
      'Sign in to save your progress and review schedule.',
  'Review berikutnya:': 'Next review:',
  'Ketersediaan deck': 'Deck availability',
  'Masuk untuk belajar': 'Sign in to study',
  'Buka dengan Lifetime Access': 'Unlock with Lifetime Access',
  'Contoh isi deck': 'Deck preview',
  'Preview menampilkan pertanyaan dan jawabannya.':
      'The preview shows sample questions and answers.',
  'Asal data': 'Data source',
  'Lihat data sumber deck': 'View deck source data',
  'Sesi belum tersedia.': 'Study session is unavailable.',
  'Belum ada kartu yang perlu dipelajari.': 'No cards are due to study.',
  'Sesi selesai!': 'Session complete!',
  'Penilaian tersimpan di perangkat · menunggu sinkronisasi':
      'Ratings saved on this device · waiting to sync',
  'PERTANYAAN & JAWABAN': 'QUESTION & ANSWER',
  'INGAT DULU JAWABANNYA': 'TRY TO REMEMBER FIRST',
  'Tampilkan jawaban': 'Show answer',
  'Seberapa mudah kamu mengingatnya?': 'How easily did you remember it?',
  'Coba simpan kembali': 'Try saving again',
  'Belum hafal': 'Again',
  'Sulit': 'Hard',
  'Ingat': 'Good',
  'Mudah': 'Easy',
  'Nilai berdasarkan ingatanmu sebelum membuka jawaban. Kamu bisa keluar kapan saja; setiap penilaian yang berhasil sudah tersimpan.':
      'Rate each card based on what you remembered before revealing the answer. You can leave at any time; each successful rating is saved.',
  'Deck belum dapat dimuat. Coba kembali.':
      'Could not load the deck. Try again.',
  'Muat ulang sesi': 'Reload session',
  'Masuk untuk belajar dan menyimpan jadwal review.':
      'Sign in to study and save your review schedule.',
  'Deck ini memerlukan Lifetime Access.': 'This deck requires Lifetime Access.',
  'Deck tidak ditemukan.': 'Deck not found.',
  'Jadwal kartu sudah berubah. Muat ulang sesi untuk melanjutkan.':
      'The card schedule has changed. Reload the session to continue.',
  'Belum dapat terhubung ke server. Coba kembali.':
      'Could not connect to the server. Please try again.',
  'Jadwal kartu berubah di perangkat lain. Muat ulang sesi untuk memakai jadwal terbaru.':
      'The card schedule changed on another device. Reload the session to use the latest schedule.',
  'Verifikasi Lifetime Access diperlukan untuk deck ini.':
      'Lifetime Access verification is required for this deck.',
  'Penilaian belum tersimpan. Coba kembali.':
      'Your rating was not saved. Please try again.',
  'Review belum terkonfirmasi. Coba simpan kembali.':
      'Your review was not confirmed. Try saving it again.',
  'Lisensi repository Kotoba': 'Kotoba repository license',
  'Materi SSW belum dapat dimuat.': 'Could not load SSW materials.',
  'Materi SSW belum tersedia dalam bahasa Inggris.':
      'English SSW materials for this sector are not available.',
  'Materi': 'Materials',
  'Tes': 'Tests',
  'Lifetime Access diperlukan': 'Lifetime Access is required',
  'Periksa kartu berikutnya': 'Check the next card',
  'Kembali ke detail deck': 'Back to deck details',
  'Detail kata': 'Word details',
  'Kanji dalam kata ini': 'Kanji in this word',
  'Caregiver dan lansia bersama pekerja berbagai bidang di Jepang':
      'A caregiver and an older adult with workers from different fields in Japan',
  'Materi & tes tersedia': 'Lessons & tests available',
  'KETERAMPILAN & PERSIAPAN UJIAN': 'SKILLS & EXAM PREPARATION',
  'BELAJAR UNTUK DUNIA KERJA': 'LEARN FOR THE WORLD OF WORK',
  'Temukan bidangmu.\nBangun keterampilanmu.':
      'Find your field.\nBuild your skills.',
  'Pelajari Kaigo': 'Learn Kaigo',
  'Cari bidang SSW': 'Search SSW fields',
  'bidang': 'fields',
  'Bidang tidak ditemukan. Coba kata lain.':
      'No fields found. Try another search term.',
  'Bidang tidak ditemukan.': 'Field not found.',
  'Latihan soal': 'Practice questions',
  'Latihan pemahaman keterampilan sesuai bidang.':
      'Practice the skills required for this field.',
  'Simulasi ujian': 'Exam simulation',
  'Paket ujian dengan timer, hasil, dan pembahasan.':
      'An exam set with a timer, results, and explanations.',
  'Materi bidang': 'Field lessons',
  'Materi keterampilan untuk bidang pilihanmu.':
      'Skill content for your chosen field.',
  'Kosakata bidang': 'Field vocabulary',
  'Istilah bahasa Jepang yang digunakan di bidang ini.':
      'Japanese terms used in this field.',
  'Konten belum tersedia. Kamu belum dapat memulai materi atau tes pada bidang ini.':
      'Content is not available yet. You cannot start lessons or tests in this field.',
  'Latih pemahaman. Ukur kesiapan.':
      'Practice your understanding. Assess your readiness.',
  'Latihan per topik memberikan pembahasan langsung. Simulasi memakai timer server dan pembahasan setelah selesai.':
      'Topic practice provides immediate explanations. Simulations use a server timer and show explanations when complete.',
  'Latihan dan simulasi termasuk Lifetime Access. Materi bidang tetap dapat dibaca gratis.':
      'Practice and simulations are included with Lifetime Access. Field lessons remain free to read.',
  'Latihan per topik': 'Practice by topic',
  'Soal orisinal persiapan SSW 1号. Bagian praktik memakai skenario teks; format foto, ilustrasi, dan demonstrasi alat tidak seluruhnya direplikasi. Hasil aplikasi bukan kelulusan resmi.':
      'Original SSW 1 exam preparation questions. Practical sections use text scenarios; photo, illustration, and tool-demonstration formats are not fully reproduced. App results are not official exam results.',
  'Timer tetap berjalan ketika aplikasi ditutup. Jawaban tersimpan setelah dikonfirmasi server. Gunakan peta soal dan tanda ragu. Setelah bagian dikirim atau waktu habis, jawaban dikunci. Kunci serta pembahasan tersedia setelah seluruh bagian selesai. Sesi dan riwayat tersimpan pada akun di server.':
      'The timer keeps running while the app is closed. Answers are saved after server confirmation. Use the question map and flag uncertain answers. Answers are locked when a section is submitted or time runs out. The answer key and explanations are available after all sections are complete. Sessions and history are saved to your server account.',
  'Simulasi belajar orisinal, bukan soal resmi. Keputusan praktik menggunakan skenario teks; tidak menggantikan latihan lapangan terawasi atau seluruh variasi visual CBT resmi.':
      'An original study simulation, not an official exam. Practical decisions use text scenarios; this does not replace supervised field practice or reproduce every visual format used in the official CBT.',
  'Riwayat simulasi': 'Simulation history',
  'Belum ada simulasi yang selesai.': 'No completed simulations yet.',
  'Latihan SSW · per topik': 'SSW practice · by topic',
  'Pilih topik latihan': 'Choose a practice topic',
  'Tanpa timer. Periksa jawaban untuk melihat pembahasan, lalu lanjutkan. Jika offline, buka halaman ini saat online terlebih dahulu agar paket latihan tersimpan. Hasil latihan tidak memberi XP.':
      'No timer. Check your answer to see the explanation, then continue. If you are offline, open this page while online first to save the practice set. Practice results do not award XP.',
  'Latihan selesai': 'Practice complete',
  'Jawaban tersimpan di perangkat dan menunggu sinkronisasi. Latihan ini tidak memberi XP.':
      'Answers are saved on this device and waiting to sync. This practice does not award XP.',
  'Jawaban sudah diverifikasi server. Latihan ini tidak memberi XP.':
      'Answers have been verified by the server. This practice does not award XP.',
  'Ulangi topik': 'Retry topic',
  'Pilih topik lain': 'Choose another topic',
  'Memeriksa…': 'Checking…',
  'Periksa jawaban': 'Check answer',
  'Tepat': 'Correct',
  'Belum tepat': 'Not quite',
  'Pelajari materi terkait': 'Study related material',
  'Menyimpan…': 'Saving…',
  'Lihat hasil latihan': 'View practice results',
  'Soal berikutnya': 'Next question',
  'Kembali ke pilihan topik': 'Back to topic selection',
  'Dua jalur: Tanaman (20 lesson) dan Peternakan (20 lesson). Pilih modul sesuai pekerjaan yang ingin kamu persiapkan. Simulasi tiap jalur tersedia terpisah di tab Tes.':
      'Two tracks: Crops (20 lessons) and Livestock (20 lessons). Choose modules for the work you are preparing for. Each track has a separate simulation in the Tests tab.',
  'Dua jalur: Perikanan Tangkap (15 lesson) dan Budi Daya (15 lesson). Pilih modul sesuai jalurmu; paket simulasi Tangkap dan Budi Daya berbeda.':
      'Two tracks: Capture Fisheries (15 lessons) and Aquaculture (15 lessons). Choose your track; the capture fisheries and aquaculture simulation sets are different.',
  'Pelajari berurutan dari dasar bidang hingga bahasa Jepang di tempat kerja. Tiap lesson berisi tujuan, penjelasan, langkah kerja, kasus, kosakata, dan cek pemahaman.':
      'Study in order, from field basics to Japanese for the workplace. Each lesson includes objectives, explanations, work steps, a case, vocabulary, and a comprehension check.',
  'Ritme belajar: 1–2 lesson per hari. Baca materi, ucapkan contoh Jepang, jelaskan kembali kasus dengan kata sendiri, lalu kerjakan cek pemahaman. Ulangi bagian yang belum jelas sebelum lanjut.':
      'Study rhythm: 1–2 lessons per day. Read the material, say the Japanese examples aloud, explain the case in your own words, then complete the comprehension check. Review anything unclear before moving on.',
  'Progress tersimpan untuk akunmu di perangkat ini.':
      'Progress is saved for your account on this device.',
  'Masuk untuk menyimpan progress setelah aplikasi ditutup.':
      'Sign in to keep your progress after closing the app.',
  'Pelajari kembali': 'Review',
  'Lesson tidak ditemukan.': 'Lesson not found.',
  'Akun berubah. Buka kembali materi untuk melanjutkan.':
      'The account changed. Reopen the lesson to continue.',
  'Progress belum tersimpan. Silakan coba lagi.':
      'Progress was not saved. Please try again.',
  '✓ Sudah dipelajari': '✓ Completed',
  'Tujuan belajar': 'Learning objectives',
  'Alur yang perlu diingat': 'Workflow to remember',
  'Perhatikan': 'Keep in mind',
  'Contoh kasus': 'Case example',
  'Respons yang tepat': 'Appropriate response',
  'Ungkapan di tempat kerja': 'Workplace expressions',
  'Cek pemahaman': 'Comprehension check',
  'Latihan dalam lesson, bukan simulasi ujian. Jawab semua pertanyaan, baca pembahasannya, dan coba kembali jika belum tepat.':
      'This is a lesson exercise, not an exam simulation. Answer every question, read the explanations, and try again if needed.',
  'Semua jawaban tepat. Tinjau kembali materi saat diperlukan, lalu lanjutkan ke lesson berikutnya.':
      'All answers are correct. Review the material as needed, then continue to the next lesson.',
  'Progress lesson sudah tersimpan. Lanjutkan ke pola berikutnya.':
      'Lesson progress has been saved. Continue to the next pattern.',
  'Semua jawaban benar. Kamu bisa membuka seluruh pembahasan.':
      'All answers are correct. You can view all explanations.',
  'Selesaikan kurikulum': 'Complete curriculum',
  'Simpan & lanjutkan': 'Save & continue',
  'Materi persiapan SSW ditulis untuk Nihongo Master. Keterampilan lapangan memerlukan pelatihan dan supervisi. Ikuti SOP, spesifikasi produk, serta arahan petugas berwenang.':
      'SSW preparation materials were written for Nihongo Master. Field skills require training and supervision. Follow SOPs, product specifications, and instructions from authorized staff.',
  'Referensi dan cakupan materi': 'References and topic coverage',
  'Cakupan disusun berdasarkan referensi di bawah. Materi dan soal aplikasi ditulis mandiri, bukan salinan buku atau bank soal resmi. Pembaruan SSW: 17 Oktober 2026.':
      'Coverage is based on the references below. App lessons and questions are original and are not copied from books or official question banks. SSW updated: October 17, 2026.',
  // Shared navigation and profile copy.
  'Simpan progress, ikuti ujian, dan buka achievement-mu.':
      'Save your progress, take exams, and unlock achievements.',
  'Lanjutkan dengan Google': 'Continue with Google',
  'atau dengan email': 'or with email',
  'Lanjut sebagai tamu': 'Continue as guest',
  'Masuk untuk melihat profil.': 'Sign in to view your profile.',
  'Foto dikompres sebelum diunggah. Tautan foto dapat dibuka oleh siapa pun yang memilikinya.':
      'Photos are compressed before upload. Anyone with the photo link can view them.',
  'Pilih dari galeri': 'Choose from gallery',
  'Ambil foto': 'Take a photo',
  'Keluar dari akun?': 'Sign out of your account?',
  'Data pribadi di perangkat akan dihapus. Progres dan penilaian offline yang belum tersinkron juga akan hilang. Hubungkan internet sebelum logout untuk menyinkronkannya.':
      'Personal data on this device will be deleted. Unsynced progress and offline ratings will also be lost. Connect to the internet before signing out to sync them.',
  'Hapus akun permanen?': 'Permanently delete your account?',
  'Profil, progress, achievement, dan sesi login akan dihapus. Pembayaran store tidak otomatis direfund. Bukti transaksi minimum disimpan untuk restore dan mencegah penyalahgunaan. Setelah membuat akun baru, gunakan Restore dengan akun store pembeli.':
      'Your profile, progress, achievements, and login session will be deleted. Store payments are not automatically refunded. A minimum transaction record is retained to support restores and prevent abuse. After creating a new account, use Restore with the purchasing store account.',
  'Ubah foto profil': 'Change profile photo',
  'Memuat profil...': 'Loading profile…',
  'Sekali bayar, belajar selamanya.': 'One payment, lifetime learning.',
  'Statistik pembelajaran': 'Learning statistics',
  'Paket & pembayaran': 'Plans & payments',
  'Masuk untuk melihat achievements': 'Sign in to view achievements',
  'Belum ada achievement dalam kategori ini.':
      'There are no achievements in this category yet.',

  // Data source and attribution screens.
  'Sumber data & atribusi': 'Data sources & attributions',
  'Dibangun bersama banyak sumber': 'Built with many sources',
  'Terima kasih kepada penerbit, pembuat dataset, dan komunitas yang menjadi sumber konten Nihongo Master. Hak cipta tetap milik masing-masing pemilik.':
      'Thank you to the publishers, dataset creators, and communities that provide content for Nihongo Master. Copyright remains with each respective owner.',
  'Sumber katalog, metadata, dan panel manga berbahasa Jepang. Karya tetap milik penulis dan penerbit masing-masing.':
      'Source of the Japanese-language manga catalog, metadata, and panels. Works remain the property of their respective authors and publishers.',
  'Berita & informasi Jepang': 'Japanese news & information',
  'Berita Jepang · CC BY 3.0. Furigana ditambahkan oleh Nihongo Master. Kredit penulis, penerjemah, dan gambar tersedia pada setiap artikel.':
      'Japanese news · CC BY 3.0. Furigana added by Nihongo Master. Author, translator, and image credits are available on each article.',
  'Informasi publik Jepang. Ketentuan penggunaan Osaka City kompatibel dengan CC BY 4.0; furigana ditambahkan oleh Nihongo Master.':
      'Japanese public information. Osaka City’s terms of use are compatible with CC BY 4.0; furigana added by Nihongo Master.',
  '大阪市 · Pemerintah Kota Osaka': 'Osaka City · Osaka City Government',
  'Hasil pencarian kata dan bacaan Jepang melalui API Jisho. Atribusi dataset yang mendasari entri mengikuti informasi dari Jisho.':
      'Japanese word and reading search results via the Jisho API. Attribution for the datasets behind entries follows Jisho’s information.',
  'Informasi kanji dan bacaannya melalui kanjiapi.dev.':
      'Kanji information and readings via kanjiapi.dev.',
  'Kosakata JLPT': 'JLPT vocabulary',
  'Dataset kosakata · CC BY-SA 4.0. Data diadaptasi menjadi lesson, dikelompokkan menurut tema, dan diterjemahkan ke bahasa Indonesia. Tingkat JLPT merupakan perkiraan komunitas.':
      'Vocabulary dataset · CC BY-SA 4.0. Data adapted into lessons, grouped by theme, and translated into Indonesian. JLPT levels are community estimates.',
  'Anki & flashcards': 'Anki & flashcards',
  'Resource quiz_data untuk deck flashcard. Lisensi repository dan catatan sumber individual dapat dilihat di halaman Anki → Sumber data & lisensi.':
      'The quiz_data resource for flashcard decks. The repository license and individual source notes are listed on Anki → Data sources & licenses.',
  'Sumber lowongan kerja': 'Job listing sources',
  'Daftar portal dan penyedia lowongan. Sumber asli juga dicantumkan pada setiap lowongan.':
      'A directory of job portals and providers. The original source is also listed for each job.',
  'Lisensi perangkat lunak open source': 'Open-source software licenses',
  'Tautan sumber belum dapat dibuka. Coba lagi.':
      'Could not open the source link. Try again.',
  'Cari arti': 'Look up meaning',
  'DBpedia & lisensi': 'DBpedia & license',
  'KETERAMPILAN UNTUK JEPANG': 'SKILLS FOR JAPAN',
  'PELUANG KARIER': 'CAREER OPPORTUNITIES',
  'BELAJAR BAHASA JEPANG': 'LEARN JAPANESE',
  'Belajar Jepang, setiap hari': 'Learn Japanese every day',
  'Kosakata, kanji, grammar, dan reading untuk target JLPT-mu.':
      'Vocabulary, kanji, grammar, and reading for your JLPT goal.',
  'Siap bekerja di Jepang': 'Get ready to work in Japan',
  'Kenali bidang SSW, pelajari materi, dan coba simulasi ujiannya.':
      'Explore SSW fields, study the material, and try an exam simulation.',
  'Jelajahi SSW': 'Explore SSW',
  'Temukan peluang kariermu': 'Find your career opportunity',
  'Jelajahi lowongan Jepang sesuai lokasi, bidang, dan kualifikasimu.':
      'Explore jobs in Japan by location, field, and qualifications.',
  'Jelajahi lowongan': 'Explore jobs',
  'Nikmati cerita dalam bahasa aslinya. Khusus Lifetime Access.':
      'Enjoy stories in their original language. For Lifetime Access members.',
  'Jelajahi manga': 'Explore manga',
  'Ingat lebih lama dengan Anki': 'Remember more with Anki',
  'Pilih deck favorit dan bangun kebiasaan review kosakata.':
      'Choose a deck and build a vocabulary review habit.',
  'Jelajahi deck': 'Explore decks',
  'NIHONGO PREMIUM': 'NIHONGO PREMIUM',
  'PREMIUM · LIFETIME ACCESS': 'PREMIUM · LIFETIME ACCESS',

  // Learn and exam navigation, instructions, and results.
  'Arti cerita': 'Story meaning',
  'Aturan latihan': 'Practice rules',
  'Aturan simulasi': 'Simulation rules',
  'BACAAN': 'PASSAGE',
  'Baca cerita lagi': 'Read the story again',
  'Baca cerita pendek, lalu cari informasi waktu, tempat, alasan, dan tujuan dalam teks.':
      'Read a short story, then find information about time, place, reasons, and purpose in the text.',
  'Bagian 1 · 文字・語彙': 'Part 1 · 文字・語彙',
  'Bagian 2 · 文法': 'Part 2 · 文法',
  'Bagian 2 · 文法・読解': 'Part 2 · 文法・読解',
  'Bagian tertulis · Tanpa listening': 'Written sections · No listening',
  'Belum dijawab': 'Unanswered',
  'Benar!': 'Correct!',
  'Buka cerita pertama': 'Open the first story',
  'Buka materi belajar': 'Open study material',
  'Cara membentuk': 'How to form it',
  'Cari judul lesson': 'Search lessons',
  'Cari tema, misalnya makanan': 'Search topics, such as food',
  'Catatan membaca': 'Reading notes',
  'Catatan pemakaian': 'Usage notes',
  'Cek kemampuan awal': 'Check your starting level',
  'Cerita kecil, dunia baru': 'Small stories, new worlds',
  'Coba tulis ulang kanji': 'Try writing the kanji again',
  'Contoh kalimat': 'Example sentences',
  'Contoh kata dan pemakaian': 'Word examples and usage',
  'Dasarmu sudah kuat.': 'You have a strong foundation.',
  'Diagnostic N5': 'N5 Diagnostic',
  'Dua bagian, dua timer': 'Two sections, two timers',
  'Fokus belajar berikutnya': 'Your next study focus',
  'Hapus pilihan': 'Clear selection',
  'Hapus tulisan': 'Clear writing',
  'Hasil diagnostic': 'Diagnostic results',
  'Hasil per jenis soal': 'Results by question type',
  'Hasil sementara': 'Provisional result',
  'Hindari kesalahan ini': 'Avoid this mistake',
  'Indigo: dijawab  ·  Bendera: ditandai': 'Indigo: answered · Flag: marked',
  'Indigo: terjawab · Bendera: ragu': 'Indigo: answered · Flag: unsure',
  'Irodori · panduan persiapan': 'Irodori · preparation guide',
  'Jawaban benar': 'Correct answer',
  'Jawaban tersimpan': 'Answer saved',
  'Jawabanmu sudah tersimpan.': 'Your answers have been saved.',
  'Kata berikutnya': 'Next word',
  'Ke soal belum dijawab': 'Go to an unanswered question',
  'Keluar': 'Exit',
  'Keluar dari sesi?': 'Leave this session?',
  'Keluar dari simulasi?': 'Leave the simulation?',
  'Kembali ke daftar cerita': 'Back to story list',
  'Kembali ke daftar lesson': 'Back to lesson list',
  'Kenali dan baca kanji dasar': 'Recognize and read basic kanji',
  'Kenali titik\nawal belajarmu.': 'Discover your starting point.',
  'Kerjakan dengan kemampuanmu saat ini. Hasilnya membantu memilih materi belajar berikutnya.':
      'Answer using your current ability. Your results will help you choose what to study next.',
  'Kerjakan ulang diagnostic': 'Retake the diagnostic',
  'Kirim bagian': 'Submit section',
  'Kirim bagian grammar?': 'Submit the grammar section?',
  'Kirim grammar dan reading?': 'Submit grammar and reading?',
  'Kirim jawaban': 'Submit answers',
  'Kosakata, kanji & grammar': 'Vocabulary, kanji & grammar',
  'Kunci bagian kosakata?': 'Lock the vocabulary section?',
  'Langkah berikutnya\nlebih jelas sekarang.':
      'Your next steps\nare clearer now.',
  'Lanjutkan cerita': 'Continue story',
  'Latih arti dan cara baca kanji dasar, lalu gunakan contoh kata untuk mengingatnya.':
      'Practice the meanings and readings of basic kanji, then use example words to help remember them.',
  'Latihan buatan Nihongo Master. Hasil berupa persentase jawaban benar, bukan skor atau prediksi kelulusan resmi JLPT.':
      'A Nihongo Master practice test. Results show the percentage of correct answers, not an official JLPT score or pass prediction.',
  'Latihan pemahaman': 'Comprehension practice',
  'Latihan tertulis · Tanpa listening': 'Written practice · No listening',
  'Lesson dikuasai!': 'Lesson mastered!',
  'Lihat hasil': 'View results',
  'Masuk kembali': 'Sign in again',
  'Materi belum tersedia.': 'No material is available yet.',
  'Menunggu sinkronisasi': 'Waiting to sync',
  'Menyiapkan kuis…': 'Preparing quiz…',
  'Menyiapkan ujian...': 'Preparing exam...',
  'Pakai dalam kalimat': 'Use it in a sentence',
  'Pelajari materi lagi': 'Review the material',
  'Pembahasan': 'Review',
  'Pembahasan jawaban': 'Answer review',
  'Penjelasan': 'Explanation',
  'Periksa lagi': 'Review again',
  'Perlu ditinjau': 'Needs review',
  'Perluas kemampuan kanji N4': 'Expand your N4 kanji knowledge',
  'Persentase dihitung dari jawaban benar. Diagnostic ini bukan penilaian kelulusan resmi JLPT.':
      'The percentage is based on correct answers. This diagnostic is not an official JLPT pass assessment.',
  'Pertanyaan berikutnya': 'Next question',
  'Peta kemampuanmu': 'Your skills overview',
  'Peta soal': 'Question map',
  'Peta soal bagian aktif': 'Active section question map',
  'Petunjuk mengingat': 'Memory tip',
  'Pilih satu jawaban': 'Choose one answer',
  'Ragu ditandai': 'Marked as unsure',
  'Riwayat diagnostic': 'Diagnostic history',
  'Salah & belum dijawab': 'Incorrect & unanswered',
  'Sebelum mulai': 'Before you start',
  'Sedikit lagi!': 'Almost there!',
  'Selesaikan diagnostic?': 'Finish the diagnostic?',
  'Selesaikan ujian': 'Finish exam',
  'Simulasi JLPT': 'JLPT simulations',
  'Simulasi selesai.': 'Simulation complete.',
  'Soal orisinal · Tanpa listening': 'Original questions · No listening',
  'Soal sebelumnya': 'Previous question',
  'Struktur resmi JFT-Basic': 'Official JFT-Basic format',
  'Sumber rujukan': 'References',
  'Tautan belum dapat dibuka. Coba lagi.':
      'Could not open the link. Try again.',
  'Bacaan pada kartu merupakan pilihan bacaan kamus.':
      'Readings on the cards are dictionary readings.',
  'Petunjuk mengingat dibuat sebagai bantuan belajar.':
      'Memory tips are provided as study aids.',
  'Jumlah goresan dan data bacaan menggunakan KANJIDIC oleh James William Breen dan EDRDG,':
      'Stroke counts and reading data use KANJIDIC by James William Breen and EDRDG,',
  'melalui kanji-data oleh David Gouveia.': 'via kanji-data by David Gouveia.',
  'Sumber:': 'Source:',
  'Lisensi kamus:': 'Dictionary license:',
  'JLPT tidak menerbitkan daftar kanji resmi yang tertutup:':
      'The JLPT does not publish a definitive official kanji list:',
  '367 kanji dari daftar rujukan N3, dengan contoh kosakata dan kalimat.':
      '367 kanji from the N3 reference list, with vocabulary and sentence examples.',
  '367 kanji dari pemetaan JLPT N2 yang dirujuk; kartu memakai contoh kosakata dari korpus latihan.':
      '367 kanji based on referenced JLPT N2 mappings; cards use vocabulary examples from the practice corpus.',
  '1.232 kanji dari pemetaan JLPT N1 yang dirujuk, dengan contoh kosakata kamus.':
      '1,232 kanji based on referenced JLPT N1 mappings, with dictionary vocabulary examples.',
  '166 kanji dari daftar rujukan N4, termasuk 21 kanji pengulangan N5.':
      '166 kanji from the N4 reference list, including 21 N5 review kanji.',
  'Gloss kamus dan contoh N2 mengikuti sumber berbahasa Inggris; belum diterjemahkan ke Bahasa Indonesia.':
      'Dictionary glosses and N2 examples follow English-language sources and have not been translated into Indonesian.',
  'Gloss kamus dan contoh N1 mengikuti sumber berbahasa Inggris; belum diterjemahkan ke Bahasa Indonesia.':
      'Dictionary glosses and N1 examples follow English-language sources and have not been translated into Indonesian.',
  'Materi data N3 diadaptasi dengan arti, contoh, dan kuis Indonesia; CC BY-SA 4.0.':
      'N3 data content was adapted with Indonesian meanings, examples, and quizzes; CC BY-SA 4.0.',
  'Materi data N4 diadaptasi dengan arti, contoh, dan kuis Indonesia; CC BY-SA 4.0.':
      'N4 data content was adapted with Indonesian meanings, examples, and quizzes; CC BY-SA 4.0.',
  'Tandai ragu': 'Mark as unsure',
  'Tandai soal': 'Mark question',
  'Tanpa listening': 'No listening',
  'Tanpa reading & listening': 'No reading or listening',
  'Tentang materi dan sumber': 'About the material and sources',
  'Terus berlatih!': 'Keep practicing!',
  'Tetap di sini': 'Stay here',
  'Tetap di ujian': 'Stay in exam',
  'Tidak dijawab': 'Unanswered',
  'Tiga bagian, satu timer': 'Three sections, one timer',
  'Tinjau partikel, perubahan bentuk kata kerja, dan pola kalimat dasar.':
      'Review particles, verb conjugations, and basic sentence patterns.',
  'Tutup pemberitahuan': 'Dismiss notification',
  'Ulangi kata sehari-hari dan pasangan kata yang sering tertukar. Perhatikan konteks kalimat.':
      'Review everyday words and commonly confused pairs. Pay attention to sentence context.',
  'Ulangi kuis': 'Retry quiz',
  'Ulangi lesson': 'Retry lesson',
  'Ulangi paket 01': 'Retry set 01',
  'Yang akan kamu kerjakan': 'What you’ll work on',
  'akses': 'access',
  'menit': 'minutes',
  'soal': 'questions',

  // Fixed simulation summaries retain the Japanese exam section names.
  'Paket 01 · 74 soal · 100 menit\nKosakata, kanji, grammar, dan reading N3. Listening dilewati.':
      'Set 01 · 74 questions · 100 minutes\nVocabulary, kanji, grammar, and reading for N3. Listening is omitted.',
  'Paket 01 · 57 soal · 80 menit\nKosakata, kanji, grammar, dan reading N4. Listening dilewati.':
      'Set 01 · 57 questions · 80 minutes\nVocabulary, kanji, grammar, and reading for N4. Listening is omitted.',
  'Paket 01 · 38 soal · 40 menit\nKosakata, kanji, dan grammar N5. Dokkai dan listening dilewati.':
      'Set 01 · 38 questions · 40 minutes\nVocabulary, kanji, and grammar for N5. Dokkai and listening are omitted.',
  'Soal orisinal untuk latihan NAT-TEST 5Q. Dua sesi tertulis: kosakata/kanji 21 soal selama 20 menit, grammar 17 soal dan reading 5 soal selama 40 menit. Listening belum termasuk. Nilai berupa persentase benar, bukan skor berskala atau kelulusan resmi NAT-TEST.':
      'Original NAT-TEST 5Q practice questions. Two written sessions: 21 vocabulary/kanji questions in 20 minutes, then 17 grammar and 5 reading questions in 40 minutes. Listening is not included. Results are a percentage of correct answers, not a scaled score or official NAT-TEST pass result.',
  'Latihan orisinal dengan jenis soal N3. Waktu kosakata 30 menit dan grammar + reading 70 menit. Listening dilewati. Nilai berupa persentase benar, bukan skor berskala atau kelulusan resmi JLPT.':
      'Original practice using N3 question types. Vocabulary time is 30 minutes, and grammar + reading time is 70 minutes. Listening is omitted. Results are a percentage of correct answers, not a scaled score or official JLPT pass result.',
  'Struktur mengikuti jenis soal N4 dan perkiraan jumlah setelah revisi 2020. Waktu kosakata 25 menit dan grammar + reading 55 menit. Listening dilewati. Soal orisinal Nihongo Master. Nilai berupa persentase benar, bukan skor berskala atau kelulusan resmi JLPT.':
      'The structure follows N4 question types and estimated counts after the 2020 revision. Vocabulary time is 25 minutes, and grammar + reading time is 55 minutes. Listening is omitted. Questions are original to Nihongo Master. Results are a percentage of correct answers, not a scaled score or official JLPT pass result.',
  'Struktur mengikuti jenis soal N5 dan perkiraan jumlah resmi setelah revisi 2020. Grammar teks termasuk bunpou, bukan tes pemahaman dokkai. Waktu grammar disesuaikan menjadi 20 menit karena dokkai dilewati; waktu resmi grammar + dokkai adalah 40 menit. Soal orisinal Nihongo Master. Nilai berupa persentase benar, bukan skor berskala atau kelulusan resmi JLPT.':
      'The structure follows N5 question types and estimated official counts after the 2020 revision. Text grammar is bunpou, not a dokkai comprehension test. Grammar time is set to 20 minutes because dokkai is omitted; the official grammar + dokkai time is 40 minutes. Questions are original to Nihongo Master. Results are a percentage of correct answers, not a scaled score or official JLPT pass result.',
  '7 bacaan kanji · 5 penulisan · 8 konteks · 4 padanan makna · 4 penggunaan kata':
      '7 kanji readings · 5 writing questions · 8 context questions · 4 meaning matches · 4 word usage questions',
  '7 bacaan kanji · 5 penulisan · 6 konteks · 3 padanan makna':
      '7 kanji readings · 5 writing questions · 6 context questions · 3 meaning matches',
  '9 pilihan grammar · 4 susunan kalimat ★ · 4 grammar teks':
      '9 grammar choices · 4 sentence-ordering questions ★ · 4 text grammar questions',
  '13 pilihan grammar · 4 susunan kalimat ★ · 4 grammar teks · 8 reading (3 pendek, 3 sedang, 2 pencarian informasi)':
      '13 grammar choices · 4 sentence-ordering questions ★ · 4 text grammar questions · 8 reading questions (3 short, 3 medium, 2 information searches)',
  '13 pilihan grammar · 5 susunan kalimat ★ · 5 grammar teks · 16 reading (4 pendek, 6 sedang, 4 panjang, 2 pencarian informasi)':
      '13 grammar choices · 5 sentence-ordering questions ★ · 5 text grammar questions · 16 reading questions (4 short, 6 medium, 4 long, 2 information searches)',
  '9 pilihan grammar · 4 susunan kalimat ★ · 4 grammar teks · 5 reading (2 pendek, 2 sedang, 1 pencarian informasi)':
      '9 grammar choices · 4 sentence-ordering questions ★ · 4 text grammar questions · 5 reading questions (2 short, 2 medium, 1 information search)',
  'Kosakata, kanji, dan ungkapan dalam konteks':
      'Vocabulary, kanji, and expressions in context',
  'Grammar, pemahaman teks, dan pencarian informasi':
      'Grammar, reading comprehension, and information search',
  '8 bacaan kanji · 6 penulisan · 11 konteks · 5 padanan makna · 5 penggunaan kata':
      '8 kanji readings · 6 writing questions · 11 context questions · 5 meaning matches · 5 word usage questions',
  '• Instruksi dan soal dalam bahasa Jepang.\n• Pilih satu dari empat jawaban; boleh mengubahnya selama bagian aktif.\n• Gunakan peta soal dan tanda ragu untuk pemeriksaan.\n• Setelah dikirim atau waktu habis, bagian dikunci. Bagian berikutnya langsung dimulai.\n• Timer tetap berjalan saat aplikasi ditutup.\n• Jawaban tersimpan di server setelah koneksi berhasil. Jawaban kosong dihitung salah.\n• Kunci dan pembahasan terbuka setelah kedua bagian selesai.':
      '• Instructions and questions are in Japanese.\n• Choose one of four answers; you can change it while the section is active.\n• Use the question map and flag uncertain answers for review.\n• After you submit or time runs out, the section locks. The next section starts immediately.\n• The timer keeps running while the app is closed.\n• Answers are saved on the server once connected. Unanswered questions count as incorrect.\n• The answer key and explanations unlock after both sections are complete.',
  '1. 文字と語彙 — 12 soal kosakata dan kanji\n2. 会話と表現 — 12 soal grammar dan ungkapan\n3. 読解 — 12 soal pesan, pengumuman, dan pencarian informasi':
      '1. 文字と語彙 — 12 vocabulary and kanji questions\n2. 会話と表現 — 12 grammar and expression questions\n3. 読解 — 12 messages, notices, and information-search questions',
  '• Waktu total 45 menit, dibagi sendiri untuk ketiga bagian.\n• Jawaban dapat diubah selama bagian masih aktif. Gunakan tanda ragu dan peta soal.\n• Setelah mengirim bagian, kamu tidak dapat kembali. Timer terus berjalan.\n• Timer tetap berjalan saat aplikasi ditutup. Jawaban tersimpan setelah dikonfirmasi server.\n• Soal kosong dihitung salah. Pembahasan tersedia setelah selesai.':
      '• The total time is 45 minutes; manage it across the three sections.\n• You can change answers while a section is active. Use the unsure flag and question map.\n• You cannot return to a section after submitting it. The timer keeps running.\n• The timer keeps running while the app is closed. Answers are saved after server confirmation.\n• Unanswered questions count as incorrect. Explanations are available after you finish.',
  'Soal orisinal Nihongo Master untuk latihan kehidupan sehari-hari. JFT-Basic resmi memiliki empat bagian, sekitar 50 soal, dan waktu 60 menit. Paket ini hanya mencakup tiga bagian tertulis dengan waktu latihan 45 menit. Skor adalah persentase benar, bukan skor berskala atau kelulusan resmi JFT-Basic.':
      'Original Nihongo Master questions for everyday-life practice. The official JFT-Basic has four sections, around 50 questions, and a 60-minute time limit. This set covers only three written sections with 45 minutes of practice time. The score is a percentage of correct answers, not a scaled score or official JFT-Basic pass result.',
  'Format dan cakupan kemampuan mengacu pada panduan publik Japan Foundation. Tema komunikasi sehari-hari merujuk secara umum pada Irodori; semua soal di aplikasi ditulis orisinal.':
      'The format and skill coverage refer to public Japan Foundation guidance. Everyday communication themes are broadly based on Irodori; all questions in the app are original.',

  // Learn and exam repository/controller error messages.
  'Respons ujian tidak valid. Coba lagi.': 'Invalid exam response. Try again.',
  'Sesi berakhir. Silakan masuk kembali.':
      'Your session has expired. Please sign in again.',
  'Lifetime Access diperlukan. Periksa status akses di halaman Premium.':
      'Lifetime Access is required. Check your access status on the Premium page.',
  'Sesi ujian tidak ditemukan. Kembali ke halaman Ujian untuk mulai lagi.':
      'Exam session not found. Return to the Exams page to start again.',
  'Jawaban tidak dapat disimpan. Sesi sudah selesai atau waktunya habis.':
      'Answers could not be saved. The session has ended or time has run out.',
  'Ujian memerlukan koneksi internet. Hubungkan kembali internet lalu coba lagi; akun tetap masuk.':
      'Exams require an internet connection. Reconnect to the internet and try again; you will stay signed in.',
  'Respons vocabulary dari server tidak valid.':
      'Invalid vocabulary response from the server.',
  'Respons grammar tidak valid. Coba muat ulang.':
      'Invalid grammar response. Reload and try again.',
  'Respons kanji tidak valid. Coba muat ulang.':
      'Invalid kanji response. Reload and try again.',
  'Respons reading tidak valid. Coba muat ulang.':
      'Invalid reading response. Reload and try again.',
  'Materi vocabulary tidak dapat dimuat. Periksa koneksi Anda.':
      'Could not load vocabulary materials. Check your connection.',
  'Materi grammar tidak dapat dimuat. Periksa koneksi dan backend.':
      'Could not load grammar materials. Check your connection and the backend.',
  'Materi kanji tidak dapat dimuat. Periksa koneksi dan backend.':
      'Could not load kanji materials. Check your connection and the backend.',
  'Materi reading tidak dapat dimuat. Periksa koneksi dan backend.':
      'Could not load reading materials. Check your connection and the backend.',
  'Lesson ini belum memiliki kosakata.': 'This lesson has no vocabulary yet.',
  'Lesson ini belum memiliki materi.': 'This lesson has no material yet.',
  'Kuis lesson belum tersedia.': 'The lesson quiz is not available yet.',
  'Kuis belum disiapkan untuk offline. Hubungkan internet dan buka kuis saat online untuk menyimpannya.':
      'This quiz has not been prepared for offline use. Connect to the internet and open it while online to save it.',
  'Ada terlalu banyak latihan yang belum tersinkron.':
      'There are too many unsynced practice sessions.',
  'Paket kuis offline tidak valid.': 'The offline quiz package is invalid.',
  'Latihan tidak lagi menerima jawaban.':
      'This practice session is no longer accepting answers.',
  'Jawaban untuk soal ini sudah disimpan.':
      'An answer for this question has already been saved.',
  'Latihan tidak ditemukan.': 'Practice session not found.',
  'Jawab semua soal sebelum menyelesaikan latihan.':
      'Answer all questions before finishing the practice session.',
  'Hasil latihan server tidak valid.': 'The server practice result is invalid.',
  'Akses offline ke latihan ini sudah tidak tersedia.':
      'Offline access to this practice is no longer available.',
  'Antrean latihan terlalu besar. Hubungkan internet untuk menyinkronkan latihan yang tertunda.':
      'The practice queue is too large. Connect to the internet to sync pending practice.',
  'Respons latihan offline tidak valid.': 'Invalid offline practice response.',
  'Masuk untuk menyimpan latihan.': 'Sign in to save practice.',
  'Akun tidak tersedia.': 'Account unavailable.',
  'Akun berubah.': 'Account changed.',
  'Akun berubah selama latihan berlangsung.':
      'The account changed while practice was in progress.',
  'Semua soal sudah dijawab. Setelah dikirim, jawaban tidak dapat diubah.':
      'All questions have been answered. Answers cannot be changed after submission.',
  'Jawaban yang sudah tersimpan bisa dilanjutkan dari halaman Ujian. Timer tetap berjalan. Tanda soal tidak disimpan saat keluar.':
      'You can resume saved answers from the Exams page. The timer keeps running. Question flags are not saved when you leave.',
  'Waktu habis. Jawaban tersimpan akan dinilai.':
      'Time is up. Saved answers will be scored.',
  '• Pilih satu jawaban untuk setiap soal.\n• Kamu boleh kembali dan mengubah jawaban.\n• Tandai soal yang ingin diperiksa lagi.\n• Jawaban disimpan saat koneksi tersedia.\n• Waktu tetap berjalan saat keluar dari sesi.\n• Soal kosong dihitung salah.':
      '• Choose one answer for each question.\n• You can go back and change your answers.\n• Flag questions you want to review.\n• Answers are saved when a connection is available.\n• Time keeps running when you leave the session.\n• Unanswered questions count as incorrect.',
  'Kosakata dikunci. Grammar dan reading dimulai; timer 40 menit berjalan.':
      'Vocabulary is locked. Grammar and reading have started; the 40-minute timer is running.',
  'Kosakata dikunci. Bagian grammar dan reading sudah dimulai; timer 70 menit berjalan.':
      'Vocabulary is locked. The grammar and reading section has started; the 70-minute timer is running.',
  'Kosakata dikunci. Bagian grammar dan reading sudah dimulai; timer 55 menit berjalan.':
      'Vocabulary is locked. The grammar and reading section has started; the 55-minute timer is running.',
  'Kosakata dikunci. Bagian grammar sudah dimulai; timer 20 menit berjalan.':
      'Vocabulary is locked. The grammar section has started; the 20-minute timer is running.',
  'Seluruh bagian selesai. Jawaban tersimpan telah dinilai.':
      'All sections are complete. Saved answers have been scored.',
  'Jawaban dikunci dan hasil ditampilkan.':
      'Answers are locked and the results are displayed.',
  'Bagian ini dikunci. Timer bagian berikutnya langsung dimulai.':
      'This section is locked. The next section timer starts immediately.',
  'Setelah dikirim, jawaban dikunci dan hasil ditampilkan.':
      'After submission, answers are locked and results are displayed.',
  'Setelah dikirim, kamu tidak bisa kembali. Bagian berikutnya dimulai dengan sisa waktu yang sama.':
      'After submission, you cannot return. The next section starts with the same remaining time.',
  'Setelah dikirim, kamu tidak bisa kembali. Timer bagian kedua langsung dimulai.':
      'After submission, you cannot return. The second section timer starts immediately.',
  'Timer tetap berjalan. Kamu dapat melanjutkan dari jawaban yang sudah tersimpan melalui halaman simulasi.':
      'The timer keeps running. You can resume your saved answers from the simulation page.',
  'Pahami pola, baca contoh, lalu kerjakan kuis. Lulus dengan nilai minimal 80% untuk menyelesaikan lesson.':
      'Study the patterns, read the examples, then take the quiz. Score at least 80% to complete the lesson.',
  'Pahami kanji, baca contoh, lalu kerjakan kuis. Lulus dengan nilai minimal 80% untuk menyelesaikan lesson.':
      'Study the kanji, read the examples, then take the quiz. Score at least 80% to complete the lesson.',
  'Ikuti Hana dalam cerita bergambar. Buka bantuan kana dan terjemahan saat diperlukan. Lulus dengan nilai minimal 80% untuk menyelesaikan lesson.':
      'Follow Hana through illustrated stories. Open kana and translation help when needed. Score at least 80% to complete the lesson.',
  'Semua lesson selesai! Anda tetap dapat mengulang materi dan kuis.':
      'All lessons are complete! You can still review the material and retake the quizzes.',
  'Semua cerita selesai! Anda tetap dapat mengulang materi dan kuis.':
      'All stories are complete! You can still review the material and retake the quizzes.',
  'Nilai minimum adalah 80%. Pelajari penjelasan dan coba kuis kembali.':
      'The passing score is 80%. Review the explanations and try the quiz again.',
  'Jawaban tersimpan di perangkat. Kelulusan, XP, dan achievement menunggu verifikasi saat terhubung kembali.':
      'Answers are saved on this device. Completion, XP, and achievements will be verified when you reconnect.',
  'Perhatikan cara kata ini digunakan dalam situasi berbeda.':
      'Notice how this word is used in different situations.',
  'jawaban benar': 'correct answers',
  'dijawab': 'answered',
  'goresan': 'strokes',
  'Terbaik': 'Best',
  'Simulasi': 'Simulation',
  'Simulasi ujian memerlukan koneksi internet. Materi dan latihan lesson yang sudah disiapkan tetap bisa dipakai offline.':
      'Exam simulations require an internet connection. Prepared lesson materials and exercises can still be used offline.',
  'Duduk. Fokus.\nKerjakan seperti ujian.':
      'Sit down. Focus.\nTreat it like an exam.',
  'Kosakata, grammar & reading tingkat lanjut':
      'Advanced vocabulary, grammar & reading',
  'Kosakata, grammar & reading tingkat mahir':
      'Proficient-level vocabulary, grammar & reading',
  'Temukan bagian yang perlu kamu latih.':
      'Find the areas you need to practise.',
  'Logout': 'Sign out',
  'Hapus akun': 'Delete account',
  'Bantuan': 'Help',
  'Coba kembali': 'Try again',
  'Masukkan email yang valid.': 'Enter a valid email address.',
  'Password minimal 8 karakter.': 'Password must be at least 8 characters.',
  'Sesi Anda berakhir. Silakan masuk kembali atau lanjut sebagai tamu.':
      'Your session has expired. Sign in again or continue as a guest.',
  'Login Google dibatalkan.': 'Google sign-in was canceled.',
  'Google Sign-In gagal. Periksa konfigurasi OAuth.':
      'Google sign-in failed. Check the OAuth configuration.',
  'Response autentikasi tidak valid.': 'Invalid authentication response.',
  'Tidak dapat terhubung ke server. Pastikan backend aktif.':
      'Could not connect to the server. Make sure the backend is running.',
  'GOOGLE_SERVER_CLIENT_ID belum dikonfigurasi.':
      'GOOGLE_SERVER_CLIENT_ID is not configured.',
  'Google Sign-In interaktif belum didukung pada platform ini.':
      'Interactive Google sign-in is not supported on this platform.',
  'Google tidak mengembalikan ID token.': 'Google did not return an ID token.',
  'Masuk ke akun untuk mengubah foto profil.':
      'Sign in to change your profile photo.',
  'Foto tidak dapat diproses. Pilih foto lain.':
      'The photo could not be processed. Choose another photo.',
  'Foto masih terlalu besar. Pilih foto lain.':
      'The photo is still too large. Choose another photo.',
  'Foto profil berhasil diperbarui.': 'Profile photo updated.',
  'Foto gagal diproses atau diunggah. Periksa koneksi lalu coba lagi.':
      'The photo could not be processed or uploaded. Check your connection and try again.',
  'Beranda belum tersimpan di perangkat ini. Hubungkan internet untuk menyiapkannya.':
      'Home content is not saved on this device yet. Connect to the internet to set it up.',
  'Pencarian maksimal 100 karakter.': 'Search is limited to 100 characters.',
  'Pencarian tidak valid. Coba gunakan kana jika ejaan romaji belum dikenali pada mode Jepang.':
      'Invalid search. In Japanese mode, try kana if the romaji spelling is not recognized.',
  'Halaman berikutnya gagal dimuat. Coba lagi.':
      'Could not load the next page. Try again.',
  'Bacaan nama': 'Name readings',
  'Gaji belum dicantumkan': 'Salary not listed',
  'Pengambilan terbatas setiap minggu; lihat situs sumber untuk seluruh lowongan.':
      'Fetching is limited each week; visit the source website for all listings.',
  'Halaman publik belum menyediakan data lowongan yang dapat dibaca. Buka situs sumber untuk melihat seluruh lowongan.':
      'The public page does not provide readable job data. Visit the source website to see all listings.',
  'Situs membatasi akses otomatis. Pengambilan dihentikan sampai jadwal berikutnya.':
      'The site limits automated access. Fetching is paused until the next scheduled run.',
  'Pembaruan terakhir belum berhasil. Data tersimpan tetap tersedia selama masih berlaku.':
      'The latest update failed. Saved data remains available while it is still valid.',
  'Nama & Geografi': 'Names & Geography',
  'Inggris & Trivia': 'English & Trivia',
  'Budaya & Trivia': 'Culture & Trivia',
  'Stasiun': 'Stations',
  'Mengikuti sumber; kebanyakan arti menggunakan bahasa Inggris.':
      'Follows the source; most meanings are in English.',
  'Data sumber tidak menyertakan jawaban sinonim. Kotoba membentuk jawabannya melalui layanan thesaurus eksternal; sesi belajar belum tersedia.':
      'The source data does not include synonym answers. Kotoba generates them through an external thesaurus service; study sessions are not available yet.',
  'Deck ini memerlukan diagram goresan serta jawaban dinamis dari mesin Kotoba. Data sumber tersimpan; sesi belajar belum tersedia.':
      'This deck requires stroke diagrams and dynamic answers from the Kotoba engine. Source data is saved; study sessions are not available yet.',

  'Kaigo · Keterampilan': 'Kaigo · Skills',
  'Kaigo · Bahasa Jepang': 'Kaigo · Japanese',
  'Pembersihan Gedung': 'Building Cleaning',
  'Pertanian · Tanaman': 'Agriculture · Crops',
  'Pertanian · Peternakan': 'Agriculture · Livestock',
  'Perikanan Tangkap': 'Capture Fisheries',
  'Budi Daya': 'Aquaculture',
  'Pengolahan Linen': 'Linen Processing',
  'Manufaktur Produk Industri': 'Industrial Product Manufacturing',
  'Konstruksi': 'Construction',
  'Pembuatan Kapal dan Mesin Kapal': 'Shipbuilding and Ship Machinery',
  'Perawatan dan Perbaikan Mobil': 'Automobile Maintenance and Repair',
  'Penerbangan': 'Aviation',
  'Perhotelan': 'Accommodation',
  'Transportasi Kendaraan Bermotor': 'Road Transportation',
  'Perkeretaapian': 'Railway',
  'Gudang Logistik': 'Logistics Warehousing',
  'Produksi Makanan dan Minuman': 'Food and Beverage Manufacturing',
  'Layanan Makanan / Restoran': 'Food Service / Restaurants',
  'Kehutanan': 'Forestry',
  'Industri Kayu': 'Wood Industry',
  'Daur Ulang Sumber Daya': 'Resource Recycling',
  '87 soal · 12 modul + bahasa Jepang · tanpa timer':
      '87 questions · 12 modules + Japanese · no timer',
  '160 soal · 10 modul · tanaman & peternakan':
      '160 questions · 10 modules · crops & livestock',
  '120 soal · 10 modul · tangkap & budi daya':
      '120 questions · 10 modules · capture fisheries & aquaculture',
  '48 soal · 4 modul · pengolahan linen':
      '48 questions · 4 modules · linen processing',
  '36 soal · 4 modul · lintas kategori industri':
      '36 questions · 4 modules · cross-category industrial skills',
  '36 soal · 4 modul · konstruksi': '36 questions · 4 modules · construction',
  '36 soal · 4 modul · kapal & mesin kapal':
      '36 questions · 4 modules · ships & ship machinery',
  '36 soal · 4 modul · perawatan dan pemeriksaan kendaraan':
      '36 questions · 4 modules · vehicle maintenance and inspection',
  '36 soal · 4 modul · layanan darat dan perawatan pesawat':
      '36 questions · 4 modules · ground services and aircraft maintenance',
  '36 soal · 4 modul · komunikasi dan layanan hotel':
      '36 questions · 4 modules · hotel communication and services',
  '36 soal · 4 modul · truk, bus, dan taksi':
      '36 questions · 4 modules · trucks, buses & taxis',
  '36 soal · 4 modul · operasi dan keselamatan kereta':
      '36 questions · 4 modules · railway operations and safety',
  '36 soal · 4 modul · alur kerja gudang logistik':
      '36 questions · 4 modules · logistics warehouse workflows',
  '36 soal · 4 modul · higiene dan kendali produksi pangan':
      '36 questions · 4 modules · food production hygiene and process control',
  '36 soal · 4 modul · layanan restoran dan higiene pangan':
      '36 questions · 4 modules · restaurant service and food hygiene',
  '36 soal · 4 modul · keselamatan kerja kehutanan':
      '36 questions · 4 modules · forestry work safety',
  '36 soal · 4 modul · pengolahan dan mutu kayu':
      '36 questions · 4 modules · wood processing and quality',
  '36 soal · 4 modul · pemilahan dan keselamatan fasilitas daur ulang':
      '36 questions · 4 modules · sorting and safety at recycling facilities',
  '60 soal · 10 modul · tanpa timer': '60 questions · 10 modules · no timer',
  '45 soal · 60 menit · 5 ranah keterampilan':
      '45 questions · 60 min · 5 skill areas',
  '15 soal · 30 menit · soal bahasa Jepang':
      '15 questions · 30 min · Japanese-language questions',
  '50 soal · 50 menit · teori + penilaian praktik':
      '50 questions · 50 min · theory + practical assessment',
  '70 soal · 60 menit · adaptasi tertulis tanpa audio':
      '70 questions · 60 min · adapted written format, no audio',
  '55 soal · 70 menit · teori + keputusan praktik':
      '55 questions · 70 min · theory + practical judgment',
  '24 soal · 35 menit · latihan belajar orisinal':
      '24 questions · 35 min · original study practice',
  '50 soal · 70 menit · teori + keputusan praktik':
      '50 questions · 70 min · theory + practical judgment',
  'Dasar Kaigo 10 soal · tubuh dan mental 6 · komunikasi 4 · dukungan kehidupan 20 · keputusan praktik 5. Bahasa soal Indonesia; istilah Jepang sesuai materi.':
      'Kaigo basics: 10 questions · body and mind: 6 · communication: 4 · daily-life support: 20 · practical judgment: 5. Questions are in Indonesian; Japanese terms follow the study material.',
  'Kosakata 5 soal · percakapan 5 · dokumen 5. Soal bahasa Jepang dengan bacaan kana dalam tanda kurung dan pembahasan Indonesia.':
      'Vocabulary: 5 questions · conversations: 5 · documents: 5. Questions are in Japanese with kana readings in parentheses and explanations in Indonesian.',
  'Teori 20 soal/20 menit, lalu penilaian praktik 30 soal/30 menit. Bagian pertama dikunci saat berpindah. Soal Indonesia dengan istilah Jepang; ujian resmi menggunakan bahasa Jepang dengan furigana.':
      'Theory: 20 questions/20 min, followed by a practical assessment: 30 questions/30 min. The first section locks when you move on. Questions are in Indonesian with Japanese terms; the official exam is in Japanese with furigana.',
  'Jalur tanaman: tanah, pupuk, benih, fasilitas, pemeliharaan, perlindungan, panen, dan hasil. Satu timer 60 menit untuk 70 soal tertulis Indonesia dengan istilah Jepang. Ujian resmi juga memuat mendengar instruksi Jepang; audio belum tersedia di simulasi ini.':
      'Crop track: soil, fertilizer, seeds, facilities, maintenance, crop protection, harvesting, and produce. One 60-minute timer for 70 written questions in Indonesian with Japanese terms. The official exam also includes listening to Japanese instructions; audio is not available in this simulation.',
  'Jalur peternakan: biosekuriti, pakan, lingkungan, reproduksi, sapi, babi, unggas, dan produk. Satu timer 60 menit untuk 70 soal tertulis Indonesia dengan istilah Jepang. Ujian resmi juga memuat mendengar instruksi Jepang; audio belum tersedia di simulasi ini.':
      'Livestock track: biosecurity, feed, environment, reproduction, cattle, pigs, poultry, and products. One 60-minute timer for 70 written questions in Indonesian with Japanese terms. The official exam also includes listening to Japanese instructions; audio is not available in this simulation.',
  'Perikanan tangkap: dasar, keselamatan, alat jaring, pancing, dan mutu hasil. Teori 30 soal/50 menit, lalu 25 keputusan praktik/20 menit dengan timer baru. Soal Indonesia dengan istilah Jepang; ujian resmi menggunakan bahasa Jepang.':
      'Capture fisheries: fundamentals, safety, nets, fishing gear, and catch quality. Theory: 30 questions/50 min, followed by 25 practical-judgment questions/20 min with a new timer. Questions are in Indonesian with Japanese terms; the official exam is in Japanese.',
  'Alur pengolahan linen hotel dan rumah sakit: penerimaan, pemilahan, pencucian, finishing, pemeriksaan, pengemasan, dan pengiriman. Paket ini latihan belajar orisinal; sumber resmi yang ditinjau belum memuat blueprint atau format ujian rinci. Penilaian praktik langsung tidak direplikasi.':
      'Hotel and hospital linen-processing workflow: receiving, sorting, washing, finishing, inspection, packing, and delivery. This is original study practice; the official sources reviewed do not yet provide a detailed blueprint or exam format. The hands-on practical assessment is not reproduced.',
  'Prinsip lintas kategori untuk keselamatan, mutu, proses, pemeriksaan, dan pelaporan. Sektor ini mencakup 17 kategori kerja resmi; setiap kategori memiliki standar teknis tersendiri. Paket ini tidak menggantikan standar kategori, demonstrasi alat, gambar kerja, atau format CBT resmi.':
      'Cross-category principles for safety, quality, processes, inspection, and reporting. This sector covers 17 official work categories, each with its own technical standards. This set does not replace category-specific standards, equipment demonstrations, working drawings, or the official CBT format.',
  'Dasar kerja lintas divisi sipil, bangunan, serta lifeline dan peralatan: membaca instruksi kerja, keselamatan, koordinasi, mutu, dan serah-terima. Latihan belajar orisinal ini tidak mereplikasi format ujian resmi, gambar proyek khusus, atau kerja praktik lapangan.':
      'Cross-division fundamentals for civil engineering, building construction, and lifeline/equipment work: reading work instructions, safety, coordination, quality, and handover. This original study practice does not reproduce the official exam format, project-specific drawings, or hands-on fieldwork.',
  'Pengetahuan dan skenario untuk alur pembangunan kapal, mesin kapal, dan peralatan listrik kapal. Divisi kerja resmi berbeda dari kategori trade ujian; simulasi teks ini tidak mencakup uji keterampilan langsung di lokasi atau demonstrasi trade tertentu.':
      'Knowledge and scenarios covering workflows for shipbuilding, ship machinery, and ship electrical equipment. Official work divisions differ from the exam trade categories; this text simulation does not include on-site skill tests or demonstrations for specific trades.',
  'Latihan dasar pemeriksaan kendaraan, perawatan berkala, penggunaan alat secara aman, pencatatan, dan komunikasi temuan. Ikuti manual kendaraan, instruksi bengkel, serta arahan teknisi yang berwenang; simulasi teks ini tidak menggantikan praktik servis atau sertifikasi mekanik.':
      'Practice the basics of vehicle inspection, routine maintenance, safe tool use, recordkeeping, and communicating findings. Follow the vehicle manual, workshop instructions, and directions from authorized technicians; this text simulation does not replace service work or mechanic certification.',
  'Latihan untuk tugas penerbangan yang berbeda: penanganan darat dan bagasi, atau pemeriksaan/perawatan pesawat sesuai penugasan. Keselamatan apron, komunikasi, identifikasi alat, dan eskalasi jadi prioritas. Tidak menggantikan pelatihan berizin, praktik lapangan, atau prosedur operator.':
      'Practice for different aviation duties: ground and baggage handling, or aircraft inspection/maintenance as assigned. Apron safety, communication, tool identification, and escalation are priorities. This does not replace licensed training, field practice, or operator procedures.',
  'Situasi hotel: reservasi dan front desk, menyampaikan informasi, menangani permintaan/tanggapan tamu, serta koordinasi kebersihan dan layanan makan sesuai peran. Dialog memakai bahasa Jepang sederhana. Ikuti kebijakan hotel untuk privasi, uang, kehilangan barang, dan keadaan darurat.':
      'Hotel scenarios: reservations and front desk, sharing information, handling guest requests and responses, and coordinating housekeeping and food service according to your role. Dialogues use simple Japanese. Follow hotel policies for privacy, money, lost property, and emergencies.',
  'Latihan berbasis skenario untuk pengemudi truk, bus, dan taksi: pemeriksaan sebelum berangkat, komunikasi dengan penumpang/penerima barang, pemilihan tindakan aman, serta laporan gangguan. Jenis SIM dan syarat berbeda menurut kendaraan; materi ini bukan panduan memperoleh SIM atau menggantikan pelatihan mengemudi.':
      'Scenario-based practice for truck, bus, and taxi drivers: pre-departure checks, communication with passengers or goods recipients, choosing safe actions, and reporting problems. License types and requirements vary by vehicle; this material is not a guide to obtaining a license and does not replace driver training.',
  'Latihan pengenalan tugas operasi stasiun/kereta, perawatan sarana, jalur, dan fasilitas sesuai penempatan. Fokus pada aturan keselamatan, komunikasi yang terkonfirmasi, pengamanan area kerja, pelaporan kondisi tidak normal, dan serah terima. Tidak memberikan kewenangan mengoperasikan sistem perkeretaapian.':
      'Introductory practice for station/train operations and maintenance of rolling stock, tracks, and facilities according to assignment. Focus on safety rules, confirmed communication, securing work areas, reporting abnormal conditions, and handover. This does not authorize you to operate railway systems.',
  'Latihan penerimaan, pemeriksaan, penyimpanan, picking, packing, dan pengiriman barang di gudang logistik, termasuk identitas lot dan keselamatan pergerakan barang. Pemerintah Jepang menyatakan rincian pelaksanaan ujian masih disiapkan; paket ini latihan orisinal dan tidak mengklaim format resmi.':
      'Practice receiving, inspecting, storing, picking, packing, and shipping goods in a logistics warehouse, including lot identification and safe material movement. The Japanese government says exam implementation details are still being prepared; this is original practice and does not claim to follow an official format.',
  'Latihan higiene produksi, pencegahan kontaminasi silang, penanganan alergen, kontrol proses, catatan lot, kebersihan alat, dan pelaporan penyimpangan. Batas suhu dan prosedur harus mengikuti standar produk serta SOP fasilitas; contoh tidak menggantikan HACCP atau instruksi perusahaan.':
      'Practice production hygiene, cross-contamination prevention, allergen handling, process control, lot records, equipment sanitation, and reporting deviations. Temperature limits and procedures must follow product standards and facility SOPs; examples do not replace HACCP or company instructions.',
  'Latihan layanan restoran: sambutan dan pesanan, konfirmasi alergi, penyajian, kebersihan, penanganan keluhan, dan kerja sama dapur. Status situs resmi SSW yang ditinjau menyatakan penerimaan baru pada bidang ini sedang tidak dibuka; periksa pengumuman resmi terbaru sebelum merencanakan pengajuan.':
      'Practice restaurant service: greetings and orders, allergy confirmation, serving, hygiene, handling complaints, and working with the kitchen. The official SSW site status reviewed says new applications in this field are currently not being accepted; check the latest official announcements before planning an application.',
  'Latihan tahapan pembibitan/penanaman, pemeliharaan tegakan, pekerjaan hutan, panen dan pemindahan hasil dengan perhatian pada cuaca, lereng, alat, komunikasi, serta konservasi. Operasi gergaji dan mesin hanya boleh dilakukan setelah pelatihan serta izin kerja setempat.':
      'Practice nursery/planting stages, stand maintenance, forestry work, harvesting, and moving forest products, with attention to weather, slopes, tools, communication, and conservation. Saws and machinery may only be operated after training and local work authorization.',
  'Latihan alur pengolahan kayu dan bahan berbasis kayu: identifikasi bahan, pengukuran, pemotongan sesuai instruksi, pengendalian debu dan mesin, pemeriksaan mutu, serta pencatatan. Batas toleransi dan pengaturan mesin selalu mengikuti gambar kerja, spesifikasi, dan operator berwenang.':
      'Practice wood and wood-based material processing workflows: identifying materials, measuring, cutting to instructions, controlling dust and machinery, inspecting quality, and keeping records. Tolerances and machine settings must always follow working drawings, specifications, and directions from authorized operators.',
  'Latihan penerimaan dan pemilahan material, identifikasi kontaminan, penelusuran aliran barang, penggunaan area dan mesin secara aman, serta penanganan temuan yang tidak dikenal. Rincian ujian resmi belum diumumkan pada halaman Kementerian Lingkungan Jepang yang ditinjau. Jangan membuka atau mencampur material berbahaya; isolasi dan laporkan sesuai prosedur fasilitas.':
      'Practice receiving and sorting materials, identifying contaminants, tracing material flows, using work areas and machinery safely, and handling unfamiliar findings. The reviewed page from Japan’s Ministry of the Environment has not announced official exam details. Do not open or mix hazardous materials; isolate them and report according to facility procedures.',
  'Budi daya: umum, keselamatan, kualitas air, sistem berpakan, serta kerang dan rumput laut tanpa pakan tambahan. Teori 40 soal/50 menit, lalu 10 keputusan praktik/20 menit dengan timer baru. Soal Indonesia dengan istilah Jepang; ujian resmi menggunakan bahasa Jepang.':
      'Aquaculture: general knowledge, safety, water quality, feeding systems, and shellfish and seaweed that require no supplemental feed. Theory: 40 questions/50 min, followed by 10 practical-judgment questions/20 min with a new timer. Questions are in Indonesian with Japanese terms; the official exam is in Japanese.',
};
