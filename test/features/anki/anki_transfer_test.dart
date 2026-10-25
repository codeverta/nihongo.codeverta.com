import 'dart:convert';
import 'dart:typed_data';

import 'package:dio/dio.dart';
import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:nihongo_master/features/anki/data/anki_repository.dart';
import 'package:nihongo_master/features/anki/data/anki_transfer_repository.dart';
import 'package:nihongo_master/features/anki/domain/anki_models.dart';
import 'package:nihongo_master/features/anki/presentation/anki_screens.dart';
import 'package:nihongo_master/features/auth/presentation/controllers/auth_controller.dart';

void main() {
  testWidgets(
    'expired premium still allows deleting a personal deck without export access',
    (tester) async {
      final deck = AnkiDeck.fromJson({
        'id': 'user-private',
        'title': 'Private',
        'category': 'Personal',
        'description': 'Personal cards',
        'instructions': 'Review',
        'originalName': 'Private',
        'sourceUrl': '',
        'cardCount': 1,
        'sourceCount': 1,
        'skippedCount': 0,
        'available': true,
        'locked': true,
        'personal': true,
        'level': 'N3',
      });
      await tester.pumpWidget(
        ProviderScope(
          overrides: [
            ankiDeckProvider('user-private').overrideWith((ref) async => deck),
            authControllerProvider.overrideWith(
              (ref) => AuthController.test(
                const AuthenticationState.authenticated(
                  AuthenticatedUser(
                    id: 'owner',
                    displayName: 'Learner',
                    email: 'learner@example.test',
                  ),
                ),
              ),
            ),
          ],
          child: const MaterialApp(
            home: AnkiDeckDetailScreen(deckId: 'user-private'),
          ),
        ),
      );
      await tester.pumpAndSettle();
      await tester.scrollUntilVisible(find.text('Hapus deck pribadi'), 200);
      expect(find.text('Hapus deck pribadi'), findsOneWidget);
      expect(find.byKey(const Key('anki-export')), findsNothing);
      expect(tester.takeException(), isNull);
    },
  );
  test(
    'import retry preserves request ID, original digest and account stamp',
    () async {
      final dio = Dio();
      final seen = <Map<String, String>>[];
      dio.interceptors.add(
        InterceptorsWrapper(
          onRequest: (options, handler) {
            expect(options.extra['offlineAccountId'], 'owner');
            expect(options.extra['offlineCacheBypass'], true);
            final form = options.data as FormData;
            expect(form.files.single.value.filename, '日本語.apkg');
            seen.add(Map.fromEntries(form.fields));
            handler.resolve(
              Response(
                requestOptions: options,
                data: {
                  'data': {'id': 'user-deck'},
                },
              ),
            );
          },
        ),
      );
      final repository = AnkiTransferRepository(dio, () => 'owner');
      for (var i = 0; i < 2; i++) {
        await repository.importFile(
          name: '日本語.apkg',
          bytes: Uint8List.fromList([1, 2]),
          preview: false,
          title: '日本語',
          digest: 'digest',
          requestId: 'stable-id',
        );
      }
      expect(seen[0], seen[1]);
      expect(seen[0]['requestId'], 'stable-id');
      expect(seen[0]['sha256'], 'digest');
    },
  );

  test('response from a previous account cannot become an export', () async {
    String? user = 'first';
    final dio = Dio();
    dio.interceptors.add(
      InterceptorsWrapper(
        onRequest: (options, handler) {
          user = 'second';
          handler.resolve(Response(requestOptions: options, data: [1, 2, 3]));
        },
      ),
    );
    await expectLater(
      AnkiTransferRepository(dio, () => user).export('deck', 'apkg'),
      throwsA(isA<AnkiTransferFailure>()),
    );
  });

  test('binary export errors retain the server explanation', () async {
    final dio = Dio();
    dio.interceptors.add(
      InterceptorsWrapper(
        onRequest: (options, handler) {
          handler.reject(
            DioException(
              requestOptions: options,
              response: Response(
                requestOptions: options,
                statusCode: 422,
                data: utf8.encode(
                  '{"error":{"message":"Deck exceeds export limit"}}',
                ),
              ),
            ),
          );
        },
      ),
    );
    await expectLater(
      AnkiTransferRepository(dio, () => 'owner').export('deck', 'apkg'),
      throwsA(
        isA<AnkiTransferFailure>().having(
          (error) => error.message,
          'message',
          'Deck exceeds export limit',
        ),
      ),
    );
  });

  test('guest and oversized files are rejected before any upload', () async {
    final dio = Dio();
    var requests = 0;
    dio.interceptors.add(
      InterceptorsWrapper(
        onRequest: (_, handler) {
          requests++;
          handler.reject(DioException(requestOptions: RequestOptions()));
        },
      ),
    );
    await expectLater(
      AnkiTransferRepository(dio, () => null).export('deck', 'json'),
      throwsA(isA<AnkiTransferFailure>()),
    );
    await expectLater(
      AnkiTransferRepository(dio, () => 'owner').importFile(
        name: 'large.json',
        bytes: Uint8List(AnkiTransferRepository.maxFileBytes + 1),
        preview: true,
      ),
      throwsA(isA<AnkiTransferFailure>()),
    );
    expect(requests, 0);
  });
}
