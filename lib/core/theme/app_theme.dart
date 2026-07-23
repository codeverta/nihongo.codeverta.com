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

/// System Design Theme untuk Aplikasi 'Nihongo Master'.
/// Mengusung konsep **Minimalis Jepang** (Washi Paper aesthetic, Crimson Red, Indigo Blue).
class AppTheme {
  const AppTheme._();

  // --- PALET WARNA UTAMA ---
  /// Merah Crimson Jepang (Sun / Aka)
  static const Color primaryColor = Color(0xFFE60012);

  /// Biru Indigo Tradisional Jepang (Aizome)
  static const Color secondaryColor = Color(0xFF1D2088);

  /// Warna Kertas Jepang Off-White (Washi Paper)
  static const Color backgroundColor = Color(0xFFFAFAFA);

  /// Warna Putih Murni untuk Surface & Card
  static const Color surfaceColor = Color(0xFFFFFFFF);

  /// Warna Teks Primer (Graphene / Dark Charcoal)
  static const Color textPrimaryColor = Color(0xFF1A1A1A);

  /// Warna Teks Sekunder (Slate Grey)
  static const Color textSecondaryColor = Color(0xFF666666);

  /// Warna Border Neutral
  static const Color borderColor = Color(0xFFE5E7EB);

  /// Theme Data untuk Mode Terang (Light Mode)
  static ThemeData get lightTheme {
    final baseTextTheme = GoogleFonts.poppinsTextTheme();

    return ThemeData(
      useMaterial3: true,
      scaffoldBackgroundColor: backgroundColor,
      colorScheme: const ColorScheme(
        brightness: Brightness.light,
        primary: primaryColor,
        onPrimary: Colors.white,
        secondary: secondaryColor,
        onSecondary: Colors.white,
        error: Color(0xFFD32F2F),
        onError: Colors.white,
        surface: surfaceColor,
        onSurface: textPrimaryColor,
      ),

      // --- TIPOGRAFI UTAMA (POPPINS) ---
      textTheme: baseTextTheme.copyWith(
        displayLarge: baseTextTheme.displayLarge?.copyWith(
          color: textPrimaryColor,
          fontWeight: FontWeight.bold,
        ),
        titleLarge: baseTextTheme.titleLarge?.copyWith(
          color: textPrimaryColor,
          fontWeight: FontWeight.w600,
          fontSize: 20,
        ),
        titleMedium: baseTextTheme.titleMedium?.copyWith(
          color: textPrimaryColor,
          fontWeight: FontWeight.w600,
          fontSize: 16,
        ),
        bodyLarge: baseTextTheme.bodyLarge?.copyWith(
          color: textPrimaryColor,
          fontSize: 16,
        ),
        bodyMedium: baseTextTheme.bodyMedium?.copyWith(
          color: textSecondaryColor,
          fontSize: 14,
        ),
        labelLarge: baseTextTheme.labelLarge?.copyWith(
          fontWeight: FontWeight.w600,
          fontSize: 14,
        ),
      ),

      // --- APPBAR THEME ---
      appBarTheme: AppBarTheme(
        elevation: 0,
        backgroundColor: backgroundColor,
        surfaceTintColor: Colors.transparent,
        centerTitle: true,
        iconTheme: const IconThemeData(color: textPrimaryColor),
        actionsIconTheme: const IconThemeData(color: textPrimaryColor),
        titleTextStyle: GoogleFonts.poppins(
          fontSize: 18,
          fontWeight: FontWeight.w600,
          color: textPrimaryColor,
        ),
      ),

      // --- ELEVATED BUTTON THEME ---
      elevatedButtonTheme: ElevatedButtonThemeData(
        style: ElevatedButton.styleFrom(
          backgroundColor: primaryColor,
          foregroundColor: Colors.white,
          elevation: 0,
          padding: const EdgeInsets.symmetric(horizontal: 24, vertical: 16),
          minimumSize: const Size.fromHeight(52),
          shape: RoundedRectangleBorder(
            borderRadius: BorderRadius.circular(12),
          ),
          textStyle: GoogleFonts.poppins(
            fontSize: 16,
            fontWeight: FontWeight.w600,
            letterSpacing: 0.5,
          ),
        ),
      ),

      // --- OUTLINED BUTTON THEME ---
      outlinedButtonTheme: OutlinedButtonThemeData(
        style: OutlinedButton.styleFrom(
          foregroundColor: secondaryColor,
          side: const BorderSide(color: secondaryColor, width: 1.5),
          padding: const EdgeInsets.symmetric(horizontal: 24, vertical: 16),
          minimumSize: const Size.fromHeight(52),
          shape: RoundedRectangleBorder(
            borderRadius: BorderRadius.circular(12),
          ),
          textStyle: GoogleFonts.poppins(
            fontSize: 16,
            fontWeight: FontWeight.w600,
          ),
        ),
      ),

      // --- INPUT DECORATION THEME ---
      inputDecorationTheme: InputDecorationTheme(
        filled: true,
        fillColor: surfaceColor,
        contentPadding:
            const EdgeInsets.symmetric(horizontal: 20, vertical: 16),
        hintStyle: GoogleFonts.poppins(
          color: textSecondaryColor.withValues(alpha: 0.7),
          fontSize: 14,
        ),
        labelStyle: GoogleFonts.poppins(
          color: textSecondaryColor,
          fontSize: 14,
        ),
        border: OutlineInputBorder(
          borderRadius: BorderRadius.circular(12),
          borderSide: const BorderSide(color: borderColor),
        ),
        enabledBorder: OutlineInputBorder(
          borderRadius: BorderRadius.circular(12),
          borderSide: const BorderSide(color: borderColor),
        ),
        focusedBorder: OutlineInputBorder(
          borderRadius: BorderRadius.circular(12),
          borderSide: const BorderSide(color: secondaryColor, width: 2),
        ),
        errorBorder: OutlineInputBorder(
          borderRadius: BorderRadius.circular(12),
          borderSide: const BorderSide(color: Color(0xFFD32F2F)),
        ),
        focusedErrorBorder: OutlineInputBorder(
          borderRadius: BorderRadius.circular(12),
          borderSide: const BorderSide(color: Color(0xFFD32F2F), width: 2),
        ),
      ),

      // --- CARD THEME ---
      cardTheme: CardThemeData(
        color: surfaceColor,
        elevation: 1,
        shadowColor: Colors.black.withValues(alpha: 0.05),
        margin: EdgeInsets.zero,
        shape: RoundedRectangleBorder(
          borderRadius: BorderRadius.circular(16),
          side: const BorderSide(color: borderColor, width: 0.8),
        ),
      ),
    );
  }

  /// Theme Data untuk Mode Gelap (Dark Mode Minimalis)
  static ThemeData get darkTheme {
    const darkBackground = Color(0xFF121212);
    const darkSurface = Color(0xFF1E1E1E);
    const darkTextPrimary = Color(0xFFF3F4F6);
    const darkTextSecondary = Color(0xFF9CA3AF);
    const darkBorder = Color(0xFF2D2D2D);

    final baseTextTheme =
        GoogleFonts.poppinsTextTheme(ThemeData.dark().textTheme);

    return ThemeData(
      useMaterial3: true,
      scaffoldBackgroundColor: darkBackground,
      colorScheme: const ColorScheme(
        brightness: Brightness.dark,
        primary: primaryColor,
        onPrimary: Colors.white,
        secondary: Color(0xFF6B7280),
        onSecondary: Colors.white,
        error: Color(0xFFCF6679),
        onError: Colors.black,
        surface: darkSurface,
        onSurface: darkTextPrimary,
      ),
      textTheme: baseTextTheme.copyWith(
        displayLarge: baseTextTheme.displayLarge?.copyWith(
          color: darkTextPrimary,
          fontWeight: FontWeight.bold,
        ),
        titleLarge: baseTextTheme.titleLarge?.copyWith(
          color: darkTextPrimary,
          fontWeight: FontWeight.w600,
        ),
        bodyLarge: baseTextTheme.bodyLarge?.copyWith(
          color: darkTextPrimary,
        ),
        bodyMedium: baseTextTheme.bodyMedium?.copyWith(
          color: darkTextSecondary,
        ),
      ),
      appBarTheme: AppBarTheme(
        elevation: 0,
        backgroundColor: darkBackground,
        surfaceTintColor: Colors.transparent,
        centerTitle: true,
        iconTheme: const IconThemeData(color: darkTextPrimary),
        titleTextStyle: GoogleFonts.poppins(
          fontSize: 18,
          fontWeight: FontWeight.w600,
          color: darkTextPrimary,
        ),
      ),
      elevatedButtonTheme: ElevatedButtonThemeData(
        style: ElevatedButton.styleFrom(
          backgroundColor: primaryColor,
          foregroundColor: Colors.white,
          elevation: 0,
          padding: const EdgeInsets.symmetric(horizontal: 24, vertical: 16),
          minimumSize: const Size.fromHeight(52),
          shape: RoundedRectangleBorder(
            borderRadius: BorderRadius.circular(12),
          ),
          textStyle: GoogleFonts.poppins(
            fontSize: 16,
            fontWeight: FontWeight.w600,
          ),
        ),
      ),
      inputDecorationTheme: InputDecorationTheme(
        filled: true,
        fillColor: darkSurface,
        contentPadding:
            const EdgeInsets.symmetric(horizontal: 20, vertical: 16),
        border: OutlineInputBorder(
          borderRadius: BorderRadius.circular(12),
          borderSide: const BorderSide(color: darkBorder),
        ),
        enabledBorder: OutlineInputBorder(
          borderRadius: BorderRadius.circular(12),
          borderSide: const BorderSide(color: darkBorder),
        ),
        focusedBorder: OutlineInputBorder(
          borderRadius: BorderRadius.circular(12),
          borderSide: const BorderSide(color: primaryColor, width: 2),
        ),
      ),
      cardTheme: CardThemeData(
        color: darkSurface,
        elevation: 0,
        shape: RoundedRectangleBorder(
          borderRadius: BorderRadius.circular(16),
          side: const BorderSide(color: darkBorder),
        ),
      ),
    );
  }
}
