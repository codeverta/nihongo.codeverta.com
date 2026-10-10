import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';
import 'package:nihongo_master/features/auth/presentation/controllers/auth_controller.dart';
import 'package:nihongo_master/features/dictionary/presentation/dictionary_screen.dart';
import 'package:nihongo_master/features/exam/presentation/screens/exam_hub_screen.dart';
import 'package:nihongo_master/features/home/presentation/screens/dashboard_screen.dart';
import 'package:nihongo_master/features/learn/presentation/screens/learn_hub_screen.dart';
import 'package:nihongo_master/features/profile/presentation/screens/profile_screen.dart';
import 'package:nihongo_master/shared/presentation/widgets/auth_gate.dart';

class MainScaffold extends ConsumerStatefulWidget {
  const MainScaffold({super.key});

  @override
  ConsumerState<MainScaffold> createState() => _MainScaffoldState();
}

class _MainScaffoldState extends ConsumerState<MainScaffold> {
  int _selectedIndex = 0;
  bool _dictionaryOpened = false;

  Future<bool> _ensureAuthenticated() async {
    if (ref.read(authControllerProvider).isAuthenticated) return true;
    final result = await context.push<bool>('/login');
    return result ?? false;
  }

  Future<void> _selectDestination(int index) async {
    if (index == 0 || index == 3 || await _ensureAuthenticated()) {
      if (mounted) {
        setState(() {
          _selectedIndex = index;
          if (index == 3) _dictionaryOpened = true;
        });
      }
    }
  }

  @override
  Widget build(BuildContext context) {
    final colorScheme = Theme.of(context).colorScheme;
    final screens = [
      DashboardScreen(
        openLearn: () => _selectDestination(1),
        openExam: () => _selectDestination(2),
        openProfile: () => _selectDestination(4),
      ),
      const LearnHubScreen(),
      const ExamHubScreen(),
      if (_dictionaryOpened)
        const DictionaryScreen(embedded: true)
      else
        const SizedBox.shrink(),
      const ProfileScreen(),
    ];
    ref.listen<AuthenticationState>(authControllerProvider, (previous, next) {
      if ((previous?.isAuthenticated ?? false) &&
          !next.isAuthenticated &&
          mounted) {
        setState(() => _selectedIndex = 0);
      }
    });

    return Scaffold(
      body: SafeArea(
        bottom: false,
        child: AuthGate(
          authorize: _ensureAuthenticated,
          child: IndexedStack(index: _selectedIndex, children: screens),
        ),
      ),
      bottomNavigationBar: NavigationBarTheme(
        data: NavigationBarThemeData(
          height: 72,
          backgroundColor: colorScheme.surface,
          indicatorColor: colorScheme.primary.withValues(alpha: 0.12),
          iconTheme: WidgetStateProperty.resolveWith((states) {
            final isSelected = states.contains(WidgetState.selected);
            return IconThemeData(
              color: isSelected
                  ? colorScheme.primary
                  : colorScheme.onSurfaceVariant,
              size: 25,
            );
          }),
          labelTextStyle: WidgetStateProperty.resolveWith((states) {
            final isSelected = states.contains(WidgetState.selected);
            return Theme.of(context).textTheme.labelMedium?.copyWith(
                  color: isSelected
                      ? colorScheme.primary
                      : colorScheme.onSurfaceVariant,
                  fontWeight: isSelected ? FontWeight.w700 : FontWeight.w500,
                );
          }),
        ),
        child: NavigationBar(
          selectedIndex: _selectedIndex,
          onDestinationSelected: _selectDestination,
          destinations: const [
            NavigationDestination(
              icon: Icon(Icons.home_outlined),
              selectedIcon: Icon(Icons.home_rounded),
              label: 'Beranda',
            ),
            NavigationDestination(
              icon: Icon(Icons.local_library_outlined),
              selectedIcon: Icon(Icons.local_library_rounded),
              label: 'Belajar',
            ),
            NavigationDestination(
              icon: Icon(Icons.quiz_outlined),
              selectedIcon: Icon(Icons.quiz_rounded),
              label: 'Ujian',
            ),
            NavigationDestination(
              icon: Icon(Icons.search_outlined),
              selectedIcon: Icon(Icons.manage_search_rounded),
              label: 'Kamus',
            ),
            NavigationDestination(
              icon: Icon(Icons.person_outline_rounded),
              selectedIcon: Icon(Icons.person_rounded),
              label: 'Profil',
            ),
          ],
        ),
      ),
    );
  }
}
