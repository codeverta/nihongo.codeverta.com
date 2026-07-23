import 'package:flutter/material.dart';
import 'package:nihongo_master/features/home/presentation/screens/dashboard_screen.dart';

/// Compatibility wrapper untuk pemanggil lama.
///
/// Navigasi utama baru menggunakan [DashboardScreen] melalui MainScaffold.
class HomeScreen extends StatelessWidget {
  const HomeScreen({
    required this.openLearn,
    required this.openExam,
    required this.openProfile,
    super.key,
  });

  final VoidCallback openLearn;
  final VoidCallback openExam;
  final VoidCallback openProfile;

  @override
  Widget build(BuildContext context) => DashboardScreen(
        openLearn: openLearn,
        openExam: openExam,
        openProfile: openProfile,
      );
}
