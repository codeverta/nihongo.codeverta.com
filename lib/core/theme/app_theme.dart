import 'package:flutter/material.dart';
import 'package:google_fonts/google_fonts.dart';

/// Class utilitas untuk gaya tipografi khusus Bahasa Jepang (Noto Sans JP).
class AppTypography {
  const AppTypography._();

  /// Font family dasar untuk Bahasa Jepang.
  static String get japaneseFontFamily => GoogleFonts.notoSansJp().fontFamily!;

  /// Utility method untuk menghasilkan [TextStyle] bahasa Jepang (Kana & Kanji).
  static TextStyle japanese({
    double? fontSize,
    FontWeight? fontWeight,
    Color? color,
    double? height,
    TextDecoration? decoration,
  }) {
    return GoogleFonts.notoSansJp(
      fontSize: fontSize,
      fontWeight: fontWeight,
      color: color,
      height: height,
      decoration: decoration,
    );
  }

  /// Teks Hiragana / Katakana / Kanji ukuran besar (Hero Display).
  static TextStyle jpKanjiHero({Color? color}) {
    return japanese(
      fontSize: 48,
      fontWeight: FontWeight.bold,
      color: color,
    );
  }

  /// Teks Kanji / Vocab ukuran kartu/flashcard.
  static TextStyle jpCardTitle({Color? color}) {
    return japanese(
      fontSize: 28,
      fontWeight: FontWeight.w700,
      color: color,
    );
  }

  /// Teks Furigana / Sub-teks Kanji.
  static TextStyle jpFurigana({Color? color}) {
    return japanese(
      fontSize: 12,
      fontWeight: FontWeight.w500,
      color: color ?? const Color(0xFF666666),
    );
  }
}

/// Extension pada [TextStyle] agar mudah mengubah font menjadi Noto Sans JP.
extension JapaneseTextStyleX on TextStyle {
  TextStyle toJapanese() {
    return GoogleFonts.notoSansJp(textStyle: this);
  }
}

/// Warna tambahan tema Sakura Pagi yang tidak tercakup [ColorScheme].
///
/// Ambil lewat `SakuraColors.of(context)` agar otomatis mengikuti mode
/// terang/gelap.
@immutable
class SakuraColors extends ThemeExtension<SakuraColors> {
  const SakuraColors({
    required this.heroStart,
    required this.heroEnd,
    required this.gold,
    required this.onGold,
    required this.goldSoft,
    required this.goldText,
    required this.success,
    required this.successSoft,
    required this.vocabulary,
    required this.grammar,
    required this.kanji,
    required this.reading,
    required this.mapSea,
    required this.mapLand,
    required this.mapLandActive,
    required this.mapStroke,
    required this.glass,
    required this.shadow,
  });

  final Color heroStart;
  final Color heroEnd;
  final Color gold;
  final Color onGold;
  final Color goldSoft;
  final Color goldText;
  final Color success;
  final Color successSoft;
  final Color vocabulary;
  final Color grammar;
  final Color kanji;
  final Color reading;
  final Color mapSea;
  final Color mapLand;
  final Color mapLandActive;
  final Color mapStroke;
  final Color glass;
  final Color shadow;

  static const light = SakuraColors(
    heroStart: Color(0xFF24285F),
    heroEnd: Color(0xFF4B3F8F),
    gold: Color(0xFFF4C66B),
    onGold: Color(0xFF1B1B1B),
    goldSoft: Color(0xFFFFF1D6),
    goldText: Color(0xFF8A5A00),
    success: Color(0xFF1E9E6A),
    successSoft: Color(0xFFE2F5EC),
    vocabulary: Color(0xFFD7263D),
    grammar: Color(0xFF3B42A8),
    kanji: Color(0xFF7C3AED),
    reading: Color(0xFFB45309),
    mapSea: Color(0xFFEEF0F6),
    mapLand: Color(0xFFFFFFFF),
    mapLandActive: Color(0xFFFCE8EA),
    mapStroke: Color(0xFFD8CEC2),
    glass: Color(0xD6FFFFFF),
    shadow: Color(0x241F1B2E),
  );

  static const dark = SakuraColors(
    heroStart: Color(0xFF262A66),
    heroEnd: Color(0xFF46398A),
    gold: Color(0xFFF4C66B),
    onGold: Color(0xFF1B1B1B),
    goldSoft: Color(0x24F4C66B),
    goldText: Color(0xFFF4C66B),
    success: Color(0xFF4ADE80),
    successSoft: Color(0x244ADE80),
    vocabulary: Color(0xFFFF8593),
    grammar: Color(0xFFA9B0FF),
    kanji: Color(0xFFC4A5FF),
    reading: Color(0xFFF4C66B),
    mapSea: Color(0xFF100E17),
    mapLand: Color(0xFF221E2F),
    mapLandActive: Color(0x33FF6B7D),
    mapStroke: Color(0xFF3D3652),
    glass: Color(0xD61E1B29),
    shadow: Color(0x99000000),
  );

  /// Warna tema aktif; aman dipakai di widget yang tidak memakai
  /// [AppTheme] (misalnya tes yang hanya memakai `MaterialApp()` polos).
  static SakuraColors of(BuildContext context) {
    final theme = Theme.of(context);
    return theme.extension<SakuraColors>() ??
        (theme.brightness == Brightness.dark ? dark : light);
  }

  /// Warna kategori materi (vocabulary, grammar, kanji, reading, dst).
  Color category(String id) => switch (id) {
    'vocabulary' => vocabulary,
    'grammar' => grammar,
    'kanji' => kanji,
    'reading' => reading,
    'listening' => success,
    _ => grammar,
  };

  @override
  SakuraColors copyWith({
    Color? heroStart,
    Color? heroEnd,
    Color? gold,
    Color? onGold,
    Color? goldSoft,
    Color? goldText,
    Color? success,
    Color? successSoft,
    Color? vocabulary,
    Color? grammar,
    Color? kanji,
    Color? reading,
    Color? mapSea,
    Color? mapLand,
    Color? mapLandActive,
    Color? mapStroke,
    Color? glass,
    Color? shadow,
  }) {
    return SakuraColors(
      heroStart: heroStart ?? this.heroStart,
      heroEnd: heroEnd ?? this.heroEnd,
      gold: gold ?? this.gold,
      onGold: onGold ?? this.onGold,
      goldSoft: goldSoft ?? this.goldSoft,
      goldText: goldText ?? this.goldText,
      success: success ?? this.success,
      successSoft: successSoft ?? this.successSoft,
      vocabulary: vocabulary ?? this.vocabulary,
      grammar: grammar ?? this.grammar,
      kanji: kanji ?? this.kanji,
      reading: reading ?? this.reading,
      mapSea: mapSea ?? this.mapSea,
      mapLand: mapLand ?? this.mapLand,
      mapLandActive: mapLandActive ?? this.mapLandActive,
      mapStroke: mapStroke ?? this.mapStroke,
      glass: glass ?? this.glass,
      shadow: shadow ?? this.shadow,
    );
  }

  @override
  SakuraColors lerp(ThemeExtension<SakuraColors>? other, double t) {
    if (other is! SakuraColors) return this;
    Color mix(Color a, Color b) => Color.lerp(a, b, t)!;
    return SakuraColors(
      heroStart: mix(heroStart, other.heroStart),
      heroEnd: mix(heroEnd, other.heroEnd),
      gold: mix(gold, other.gold),
      onGold: mix(onGold, other.onGold),
      goldSoft: mix(goldSoft, other.goldSoft),
      goldText: mix(goldText, other.goldText),
      success: mix(success, other.success),
      successSoft: mix(successSoft, other.successSoft),
      vocabulary: mix(vocabulary, other.vocabulary),
      grammar: mix(grammar, other.grammar),
      kanji: mix(kanji, other.kanji),
      reading: mix(reading, other.reading),
      mapSea: mix(mapSea, other.mapSea),
      mapLand: mix(mapLand, other.mapLand),
      mapLandActive: mix(mapLandActive, other.mapLandActive),
      mapStroke: mix(mapStroke, other.mapStroke),
      glass: mix(glass, other.glass),
      shadow: mix(shadow, other.shadow),
    );
  }
}

/// Token gerak (durasi & kurva) tema Sakura Pagi.
class AppMotion {
  const AppMotion._();

  static const fast = Duration(milliseconds: 120);
  static const normal = Duration(milliseconds: 240);
  static const slow = Duration(milliseconds: 360);
  static const emphasized = Duration(milliseconds: 520);

  /// Spring lembut dengan sedikit overshoot.
  static const spring = Cubic(0.34, 1.56, 0.64, 1);
  static const standard = Cubic(0.2, 0, 0, 1);
  static const enter = Cubic(0, 0, 0.2, 1);
  static const exit = Cubic(0.4, 0, 1, 1);

  /// `true` bila pengguna meminta animasi dikurangi.
  static bool reduced(BuildContext context) =>
      MediaQuery.maybeDisableAnimationsOf(context) ?? false;

  static Duration of(BuildContext context, Duration duration) =>
      reduced(context) ? Duration.zero : duration;
}

/// System Design Theme untuk Aplikasi 'Nihongo Master'.
/// Tema **Sakura Pagi**: kertas washi hangat, merah vermilion, indigo ai-iro,
/// dengan mode gelap ungu-tinta.
class AppTheme {
  const AppTheme._();

  // --- PALET WARNA UTAMA (MODE TERANG) ---
  /// Merah vermilion (Shu / Aka)
  static const Color primaryColor = Color(0xFFD7263D);

  /// Biru indigo tradisional (Ai-iro)
  static const Color secondaryColor = Color(0xFF2A2F6E);

  /// Kertas washi hangat
  static const Color backgroundColor = Color(0xFFFBF8F4);

  /// Permukaan kartu
  static const Color surfaceColor = Color(0xFFFFFFFF);

  /// Teks primer (tinta ungu gelap)
  static const Color textPrimaryColor = Color(0xFF1F1B2E);

  /// Teks sekunder
  static const Color textSecondaryColor = Color(0xFF6B6577);

  /// Border netral
  static const Color borderColor = Color(0xFFECE5DC);

  // --- PALET MODE GELAP ---
  static const Color darkBackground = Color(0xFF14121C);
  static const Color darkSurface = Color(0xFF1E1B29);
  static const Color darkSurface2 = Color(0xFF272334);
  static const Color darkSurface3 = Color(0xFF302B40);
  static const Color darkTextPrimary = Color(0xFFF3EFF7);
  static const Color darkTextSecondary = Color(0xFFA8A1B8);
  static const Color darkBorder = Color(0xFF322D43);
  static const Color darkPrimary = Color(0xFFCF2C42);
  static const Color darkPrimaryText = Color(0xFFFF8593);

  static ColorScheme get _lightScheme =>
      ColorScheme.fromSeed(
        seedColor: primaryColor,
      ).copyWith(
        primary: primaryColor,
        onPrimary: Colors.white,
        primaryContainer: const Color(0xFFFCE8EA),
        onPrimaryContainer: const Color(0xFF8E1426),
        secondary: secondaryColor,
        onSecondary: Colors.white,
        secondaryContainer: const Color(0xFFE8E9F6),
        onSecondaryContainer: secondaryColor,
        tertiary: const Color(0xFFB45309),
        onTertiary: Colors.white,
        tertiaryContainer: const Color(0xFFFFF1D6),
        onTertiaryContainer: const Color(0xFF8A5A00),
        error: const Color(0xFFC62828),
        onError: Colors.white,
        errorContainer: const Color(0xFFFDE7E7),
        onErrorContainer: const Color(0xFF8C1D18),
        surface: backgroundColor,
        onSurface: textPrimaryColor,
        onSurfaceVariant: textSecondaryColor,
        surfaceContainerLowest: surfaceColor,
        surfaceContainerLow: surfaceColor,
        surfaceContainer: const Color(0xFFF8F4EE),
        surfaceContainerHigh: const Color(0xFFF4EFE8),
        surfaceContainerHighest: const Color(0xFFEDE6DC),
        outline: const Color(0xFFCFC6BA),
        outlineVariant: borderColor,
        shadow: const Color(0xFF1F1B2E),
        inverseSurface: textPrimaryColor,
        onInverseSurface: backgroundColor,
        inversePrimary: darkPrimaryText,
        surfaceTint: Colors.transparent,
      );

  static ColorScheme get _darkScheme =>
      ColorScheme.fromSeed(
        seedColor: primaryColor,
        brightness: Brightness.dark,
      ).copyWith(
        primary: darkPrimaryText,
        onPrimary: const Color(0xFF3A0710),
        primaryContainer: darkPrimary,
        onPrimaryContainer: Colors.white,
        secondary: const Color(0xFF7A80E6),
        onSecondary: Colors.white,
        secondaryContainer: const Color(0x24A9B0FF),
        onSecondaryContainer: const Color(0xFFB9BEFF),
        tertiary: const Color(0xFFF4C66B),
        onTertiary: const Color(0xFF1B1B1B),
        tertiaryContainer: const Color(0x24F4C66B),
        onTertiaryContainer: const Color(0xFFF4C66B),
        error: const Color(0xFFFF8A80),
        onError: const Color(0xFF3A0710),
        errorContainer: const Color(0x33FF8A80),
        onErrorContainer: const Color(0xFFFFDAD6),
        surface: darkBackground,
        onSurface: darkTextPrimary,
        onSurfaceVariant: darkTextSecondary,
        surfaceContainerLowest: const Color(0xFF100E17),
        surfaceContainerLow: darkSurface,
        surfaceContainer: darkSurface,
        surfaceContainerHigh: darkSurface2,
        surfaceContainerHighest: darkSurface3,
        outline: const Color(0xFF4A4361),
        outlineVariant: darkBorder,
        shadow: Colors.black,
        inverseSurface: darkTextPrimary,
        onInverseSurface: darkBackground,
        inversePrimary: primaryColor,
        surfaceTint: Colors.transparent,
      );

  /// Theme Data untuk Mode Terang (Light Mode)
  static ThemeData get lightTheme => _build(_lightScheme, SakuraColors.light);

  /// Theme Data untuk Mode Gelap (Dark Mode)
  static ThemeData get darkTheme => _build(_darkScheme, SakuraColors.dark);

  static ThemeData _build(ColorScheme scheme, SakuraColors sakura) {
    final dark = scheme.brightness == Brightness.dark;
    final base = dark ? ThemeData.dark() : ThemeData.light();
    final cardColor = dark ? darkSurface : surfaceColor;
    final fieldColor = dark ? darkSurface : surfaceColor;
    final chipColor = dark ? darkSurface2 : surfaceColor;
    final baseTextTheme = GoogleFonts.plusJakartaSansTextTheme(
      base.textTheme,
    ).apply(bodyColor: scheme.onSurface, displayColor: scheme.onSurface);
    final textTheme = baseTextTheme.copyWith(
      displayLarge: baseTextTheme.displayLarge?.copyWith(
        fontWeight: FontWeight.w800,
        letterSpacing: -1,
      ),
      headlineMedium: baseTextTheme.headlineMedium?.copyWith(
        fontWeight: FontWeight.w800,
        letterSpacing: -0.6,
      ),
      headlineSmall: baseTextTheme.headlineSmall?.copyWith(
        fontWeight: FontWeight.w800,
        letterSpacing: -0.5,
      ),
      titleLarge: baseTextTheme.titleLarge?.copyWith(
        fontWeight: FontWeight.w700,
        fontSize: 20,
        letterSpacing: -0.3,
      ),
      titleMedium: baseTextTheme.titleMedium?.copyWith(
        fontWeight: FontWeight.w700,
        fontSize: 16,
      ),
      titleSmall: baseTextTheme.titleSmall?.copyWith(
        fontWeight: FontWeight.w700,
      ),
      bodyLarge: baseTextTheme.bodyLarge?.copyWith(fontSize: 16, height: 1.5),
      bodyMedium: baseTextTheme.bodyMedium?.copyWith(
        color: scheme.onSurfaceVariant,
        fontSize: 14,
        height: 1.45,
      ),
      bodySmall: baseTextTheme.bodySmall?.copyWith(
        color: scheme.onSurfaceVariant,
      ),
      labelLarge: baseTextTheme.labelLarge?.copyWith(
        fontWeight: FontWeight.w700,
        fontSize: 14,
      ),
    );
    final buttonText = GoogleFonts.plusJakartaSans(
      fontSize: 15,
      fontWeight: FontWeight.w700,
    );
    final pill = RoundedRectangleBorder(
      borderRadius: BorderRadius.circular(999),
    );
    final filledBg = dark ? darkPrimary : primaryColor;

    return ThemeData(
      useMaterial3: true,
      brightness: scheme.brightness,
      colorScheme: scheme,
      scaffoldBackgroundColor: scheme.surface,
      canvasColor: scheme.surface,
      textTheme: textTheme,
      extensions: [sakura],
      dividerTheme: DividerThemeData(color: scheme.outlineVariant, space: 1),

      appBarTheme: AppBarTheme(
        elevation: 0,
        scrolledUnderElevation: 0,
        backgroundColor: scheme.surface,
        foregroundColor: scheme.onSurface,
        surfaceTintColor: Colors.transparent,
        centerTitle: true,
        iconTheme: IconThemeData(color: scheme.onSurface),
        actionsIconTheme: IconThemeData(color: scheme.onSurface),
        titleTextStyle: GoogleFonts.plusJakartaSans(
          fontSize: 18,
          fontWeight: FontWeight.w800,
          color: scheme.onSurface,
        ),
      ),

      elevatedButtonTheme: ElevatedButtonThemeData(
        style: ElevatedButton.styleFrom(
          backgroundColor: filledBg,
          foregroundColor: Colors.white,
          elevation: 0,
          padding: const EdgeInsets.symmetric(horizontal: 24, vertical: 16),
          minimumSize: const Size.fromHeight(52),
          shape: pill,
          textStyle: buttonText,
        ),
      ),
      filledButtonTheme: FilledButtonThemeData(
        style: FilledButton.styleFrom(
          backgroundColor: filledBg,
          foregroundColor: Colors.white,
          shape: pill,
          textStyle: buttonText,
        ),
      ),
      outlinedButtonTheme: OutlinedButtonThemeData(
        style: OutlinedButton.styleFrom(
          foregroundColor: dark ? const Color(0xFFB9BEFF) : secondaryColor,
          side: BorderSide(color: scheme.outline, width: 1.2),
          padding: const EdgeInsets.symmetric(horizontal: 24, vertical: 16),
          minimumSize: const Size.fromHeight(52),
          shape: pill,
          textStyle: buttonText,
        ),
      ),
      textButtonTheme: TextButtonThemeData(
        style: TextButton.styleFrom(
          foregroundColor: scheme.primary,
          textStyle: GoogleFonts.plusJakartaSans(
            fontSize: 13,
            fontWeight: FontWeight.w700,
          ),
          shape: pill,
        ),
      ),

      inputDecorationTheme: InputDecorationTheme(
        filled: true,
        fillColor: fieldColor,
        contentPadding: const EdgeInsets.symmetric(
          horizontal: 20,
          vertical: 16,
        ),
        hintStyle: GoogleFonts.plusJakartaSans(
          color: scheme.onSurfaceVariant.withValues(alpha: 0.8),
          fontSize: 14,
        ),
        labelStyle: GoogleFonts.plusJakartaSans(
          color: scheme.onSurfaceVariant,
          fontSize: 14,
        ),
        border: OutlineInputBorder(
          borderRadius: BorderRadius.circular(16),
          borderSide: BorderSide(color: scheme.outlineVariant),
        ),
        enabledBorder: OutlineInputBorder(
          borderRadius: BorderRadius.circular(16),
          borderSide: BorderSide(color: scheme.outlineVariant),
        ),
        focusedBorder: OutlineInputBorder(
          borderRadius: BorderRadius.circular(16),
          borderSide: BorderSide(color: scheme.primary, width: 1.6),
        ),
        errorBorder: OutlineInputBorder(
          borderRadius: BorderRadius.circular(16),
          borderSide: BorderSide(color: scheme.error),
        ),
        focusedErrorBorder: OutlineInputBorder(
          borderRadius: BorderRadius.circular(16),
          borderSide: BorderSide(color: scheme.error, width: 1.6),
        ),
      ),

      cardTheme: CardThemeData(
        color: cardColor,
        elevation: 0,
        shadowColor: Colors.transparent,
        surfaceTintColor: Colors.transparent,
        margin: EdgeInsets.zero,
        shape: RoundedRectangleBorder(
          borderRadius: BorderRadius.circular(18),
          side: BorderSide(color: scheme.outlineVariant, width: 0.8),
        ),
      ),

      chipTheme: ChipThemeData(
        backgroundColor: chipColor,
        selectedColor: dark ? const Color(0xFFA9B0FF) : secondaryColor,
        secondarySelectedColor: dark
            ? const Color(0xFFA9B0FF)
            : secondaryColor,
        checkmarkColor: dark ? darkBackground : Colors.white,
        side: BorderSide(color: scheme.outlineVariant),
        shape: pill,
        // Warna label mengikuti status terpilih agar tetap terbaca di atas
        // latar indigo chip aktif.
        labelStyle: GoogleFonts.plusJakartaSans(
          fontSize: 13,
          fontWeight: FontWeight.w700,
          color: WidgetStateColor.resolveWith(
            (states) => states.contains(WidgetState.selected)
                ? (dark ? darkBackground : Colors.white)
                : scheme.onSurface,
          ),
        ),
        iconTheme: IconThemeData(size: 18, color: scheme.onSurfaceVariant),
        secondaryLabelStyle: GoogleFonts.plusJakartaSans(
          fontSize: 13,
          fontWeight: FontWeight.w700,
          color: dark ? darkBackground : Colors.white,
        ),
        padding: const EdgeInsets.symmetric(horizontal: 6, vertical: 2),
      ),

      segmentedButtonTheme: SegmentedButtonThemeData(
        style: ButtonStyle(
          shape: WidgetStatePropertyAll(pill),
          textStyle: WidgetStatePropertyAll(
            GoogleFonts.plusJakartaSans(
              fontSize: 13,
              fontWeight: FontWeight.w700,
            ),
          ),
        ),
      ),

      listTileTheme: ListTileThemeData(
        iconColor: scheme.onSurfaceVariant,
        shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(16)),
      ),

      bottomSheetTheme: BottomSheetThemeData(
        backgroundColor: cardColor,
        surfaceTintColor: Colors.transparent,
        showDragHandle: true,
        dragHandleColor: scheme.outlineVariant,
        shape: const RoundedRectangleBorder(
          borderRadius: BorderRadius.vertical(top: Radius.circular(24)),
        ),
      ),

      dialogTheme: DialogThemeData(
        backgroundColor: cardColor,
        surfaceTintColor: Colors.transparent,
        shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(24)),
      ),

      snackBarTheme: SnackBarThemeData(
        behavior: SnackBarBehavior.floating,
        backgroundColor: scheme.inverseSurface,
        contentTextStyle: GoogleFonts.plusJakartaSans(
          color: scheme.onInverseSurface,
          fontWeight: FontWeight.w600,
          fontSize: 13,
        ),
        actionTextColor: dark ? primaryColor : const Color(0xFFF7B6BF),
        shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(14)),
      ),

      progressIndicatorTheme: ProgressIndicatorThemeData(
        color: scheme.primary,
        linearTrackColor: scheme.surfaceContainerHigh,
        circularTrackColor: Colors.transparent,
      ),

      navigationBarTheme: NavigationBarThemeData(
        backgroundColor: cardColor,
        surfaceTintColor: Colors.transparent,
        indicatorColor: dark ? const Color(0x26FF6B7D) : const Color(0xFFFCE8EA),
        indicatorShape: pill,
      ),

      switchTheme: SwitchThemeData(
        thumbColor: WidgetStateProperty.resolveWith(
          (states) => states.contains(WidgetState.selected)
              ? Colors.white
              : scheme.outline,
        ),
        trackColor: WidgetStateProperty.resolveWith(
          (states) => states.contains(WidgetState.selected)
              ? filledBg
              : scheme.surfaceContainerHighest,
        ),
      ),
    );
  }
}
