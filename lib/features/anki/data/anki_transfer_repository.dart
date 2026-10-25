import 'dart:convert';
import 'dart:typed_data';

import 'package:dio/dio.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:nihongo_master/core/network/dio_client.dart';
import 'package:nihongo_master/features/anki/domain/anki_models.dart';
import 'package:nihongo_master/features/auth/presentation/controllers/auth_controller.dart';

final ankiTransferRepositoryProvider = Provider(
  (ref) => AnkiTransferRepository(
    ref.watch(dioProvider),
    () => ref.read(authControllerProvider).user?.id,
  ),
);

class AnkiTransferRepository {
  AnkiTransferRepository(this.dio, this.currentUserId);
  final Dio dio;
  final String? Function() currentUserId;
  static const maxFileBytes = 20 * 1024 * 1024;

  Options _options(String owner, {ResponseType? responseType}) => Options(
    extra: {'offlineAccountId': owner, 'offlineCacheBypass': true},
    responseType: responseType,
    sendTimeout: const Duration(seconds: 90),
    receiveTimeout: const Duration(minutes: 2),
  );
  String _owner() =>
      currentUserId() ??
      (throw const AnkiTransferFailure(
        'Masuk untuk mengimpor atau mengekspor deck.',
      ));
  void _checkOwner(String owner) {
    if (currentUserId() != owner) {
      throw const AnkiTransferFailure(
        'Akun sudah berubah. Buka kembali halaman ini.',
      );
    }
  }

  Future<Map<String, dynamic>> importFile({
    required String name,
    required Uint8List bytes,
    required bool preview,
    String? title,
    String? digest,
    String? requestId,
    CancelToken? cancelToken,
  }) async {
    final owner = _owner();
    if (bytes.isEmpty || bytes.length > maxFileBytes) {
      throw const AnkiTransferFailure(
        'Pilih file dengan ukuran maksimal 20 MB.',
      );
    }
    try {
      final result = await dio.post<Map<String, dynamic>>(
        preview ? '/anki/import/preview' : '/anki/import',
        data: FormData.fromMap({
          'file': MultipartFile.fromBytes(bytes, filename: name),
          if (title != null) 'title': title,
          if (digest != null) 'sha256': digest,
          if (requestId != null) 'requestId': requestId,
        }),
        options: _options(owner),
        cancelToken: cancelToken,
      );
      _checkOwner(owner);
      return Map<String, dynamic>.from(result.data!['data'] as Map);
    } on DioException catch (error) {
      throw _failure(error);
    }
  }

  Future<Uint8List> export(
    String deckId,
    String format, {
    CancelToken? cancelToken,
  }) async {
    final owner = _owner();
    final transferCancel = cancelToken ?? CancelToken();
    try {
      final result = await dio.get<List<int>>(
        '/anki/decks/${Uri.encodeComponent(deckId)}/export',
        queryParameters: {'format': format},
        options: _options(owner, responseType: ResponseType.bytes),
        cancelToken: transferCancel,
        onReceiveProgress: (received, total) {
          if (received > maxFileBytes || total > maxFileBytes) {
            transferCancel.cancel('Export exceeds 20 MB.');
          }
        },
      );
      _checkOwner(owner);
      return Uint8List.fromList(result.data!);
    } on DioException catch (error) {
      throw _failure(error);
    }
  }

  Future<void> delete(String deckId) async {
    final owner = _owner();
    try {
      await dio.delete<void>(
        '/anki/decks/${Uri.encodeComponent(deckId)}',
        options: _options(owner),
      );
      _checkOwner(owner);
    } on DioException catch (error) {
      throw _failure(error);
    }
  }

  AnkiDeck importedDeck(Map<String, dynamic> value) => AnkiDeck.fromJson(value);
}

class AnkiTransferFailure implements Exception {
  const AnkiTransferFailure(this.message);
  final String message;
  @override
  String toString() => message;
}

AnkiTransferFailure _failure(DioException error) {
  Object? body = error.response?.data;
  if (body is List<int> && body.length < 65536) {
    try {
      body = jsonDecode(utf8.decode(body));
    } on Object {
      body = null;
    }
  }
  final errorObject = body is Map ? body['error'] : null;
  final serverMessage = errorObject is Map ? errorObject['message'] : null;
  if (serverMessage is String && serverMessage.isNotEmpty) {
    return AnkiTransferFailure(serverMessage);
  }
  return AnkiTransferFailure(switch (error.response?.statusCode) {
    401 => 'Masuk untuk mengelola deck pribadi.',
    402 => 'Deck ini memerlukan Lifetime Access.',
    404 =>
      'Fitur transfer belum tersedia di server, atau deck tidak ditemukan.',
    413 => 'File terlalu besar. Batas unggahan 20 MB.',
    422 => 'Format file deck tidak didukung atau isinya tidak valid.',
    429 => 'Pemrosesan deck sedang penuh. Coba lagi beberapa detik.',
    _ => 'Transfer belum selesai. Periksa koneksi dan coba lagi.',
  });
}
