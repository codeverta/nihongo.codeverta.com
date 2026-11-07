import 'package:flutter_test/flutter_test.dart';
import 'package:nihongo_master/core/storage/recent_access.dart';
import 'package:shared_preferences/shared_preferences.dart';

void main() {
  test('records, persists and orders recent items', () async {
    SharedPreferences.setMockInitialValues({});
    final prefs = await SharedPreferences.getInstance();
    final store = RecentAccess(prefs, 'test');
    store.record('a', now: DateTime.fromMillisecondsSinceEpoch(1000));
    store.record('b', now: DateTime.fromMillisecondsSinceEpoch(2000));
    store.record('a', now: DateTime.fromMillisecondsSinceEpoch(3000));

    final ids = ['b', 'c', 'a', 'd'];
    ids.sort((x, y) => compareRecent(store.state, x, y));
    expect(ids.sublist(0, 2), ['a', 'b']);
    expect(store.state['a']!.count, 2);
    expect(compareRecent(store.state, 'c', 'd'), 0);

    final reloaded = RecentAccess(prefs, 'test');
    expect(reloaded.state['a']!.lastMs, 3000);
  });
}
