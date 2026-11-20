import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:nihongo_master/core/widgets/system_navigation_inset.dart';

const _screen = Size(390, 800);
const _navBar = 48.0;

Widget _app({required double navBar, double keyboard = 0, Widget? body}) {
  return MediaQuery(
    data: MediaQueryData(
      size: _screen,
      // Flutter reports the bottom padding as zero while the keyboard covers it.
      padding: EdgeInsets.only(bottom: keyboard > 0 ? 0 : navBar),
      viewPadding: const EdgeInsets.only(bottom: _navBar),
      viewInsets: EdgeInsets.only(bottom: keyboard),
    ),
    child: MaterialApp(
      builder: (context, child) => SystemNavigationInset(child: child!),
      home: Scaffold(
        body:
            body ??
            const Align(
              alignment: Alignment.bottomCenter,
              child: SizedBox(
                key: Key('bottom-button'),
                height: 40,
                width: 200,
              ),
            ),
      ),
    ),
  );
}

void main() {
  testWidgets('bottom content stays above the system navigation bar', (
    tester,
  ) async {
    tester.view
      ..physicalSize = _screen
      ..devicePixelRatio = 1;
    addTearDown(tester.view.reset);
    await tester.pumpWidget(_app(navBar: _navBar));
    final rect = tester.getRect(find.byKey(const Key('bottom-button')));
    expect(rect.bottom, _screen.height - _navBar);
  });

  testWidgets('bottom sheets are lifted above the navigation bar too', (
    tester,
  ) async {
    tester.view
      ..physicalSize = _screen
      ..devicePixelRatio = 1;
    addTearDown(tester.view.reset);
    await tester.pumpWidget(
      _app(
        navBar: _navBar,
        body: Builder(
          builder: (context) => TextButton(
            onPressed: () => showModalBottomSheet<void>(
              context: context,
              builder: (_) => const SizedBox(
                key: Key('sheet-footer'),
                height: 60,
                width: double.infinity,
              ),
            ),
            child: const Text('open'),
          ),
        ),
      ),
    );
    await tester.tap(find.text('open'));
    await tester.pumpAndSettle();
    final rect = tester.getRect(find.byKey(const Key('sheet-footer')));
    expect(rect.bottom, lessThanOrEqualTo(_screen.height - _navBar));
  });

  testWidgets('opening the keyboard keeps focus and adds no extra gap', (
    tester,
  ) async {
    tester.view
      ..physicalSize = _screen
      ..devicePixelRatio = 1;
    addTearDown(tester.view.reset);
    const body = Column(
      children: [
        Spacer(),
        TextField(key: Key('field'), autofocus: true),
      ],
    );
    await tester.pumpWidget(_app(navBar: _navBar, body: body));
    await tester.pump();
    final state = tester.state<EditableTextState>(find.byType(EditableText));
    expect(state.widget.focusNode.hasFocus, isTrue);

    // Keyboard opens: padding drops to zero, the tree must not be rebuilt.
    await tester.pumpWidget(_app(navBar: _navBar, keyboard: 300, body: body));
    await tester.pump();
    expect(
      tester.state<EditableTextState>(find.byType(EditableText)),
      same(state),
    );
    expect(state.widget.focusNode.hasFocus, isTrue);
  });
}
