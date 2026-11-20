import 'dart:async';
import 'dart:convert';

import 'package:dio/dio.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:nihongo_master/core/network/dio_client.dart';

/// Quick actions offered above the input. Wire names match the backend.
enum AiChip {
  summary('summary', 'Ringkas'),
  explain('explain', 'Jelaskan grammar'),
  examples('examples', 'Beri contoh kalimat'),
  quiz('quiz', 'Kuis aku (5 soal)'),
  correct('correct', 'Koreksi kalimatku');

  const AiChip(this.wire, this.label);

  final String wire;

  /// Indonesian source label; localized with `.localized(context)`.
  final String label;

  /// Correction needs the student's sentence; other chips send no text.
  bool get needsText => this == AiChip.correct;

  static AiChip? fromWire(String? value) {
    for (final chip in values) {
      if (chip.wire == value) return chip;
    }
    return null;
  }
}

/// The lesson types that offer Tanya AI.
enum AiLessonType { grammar, vocabulary, kanji, reading }

class AiChatMessage {
  const AiChatMessage({
    required this.role,
    required this.content,
    this.chip,
    this.cached = false,
  });

  factory AiChatMessage.fromJson(Map<String, dynamic> json) => AiChatMessage(
    role: json['role'] == 'assistant' ? AiRole.assistant : AiRole.user,
    content: json['content'] as String? ?? '',
    chip: AiChip.fromWire(json['chip'] as String?),
    cached: json['cached'] == true,
  );

  final AiRole role;
  final String content;
  final AiChip? chip;
  final bool cached;
}

enum AiRole { user, assistant }

/// A server-side refusal. [code] drives the UI (ads, upsell, retry).
class AiChatException implements Exception {
  const AiChatException(this.code, this.message, {this.messagesPerAd = 0});

  final String code;
  final String message;

  /// For [premiumAdOffer]: how many messages one ad adds.
  final int messagesPerAd;

  static const adRequired = 'ai_ad_required';
  static const dailyLimit = 'ai_daily_limit';
  static const premiumLimit = 'ai_premium_limit';

  /// A Lifetime account used its daily messages; an optional ad adds more.
  static const premiumAdOffer = 'ai_premium_ad_offer';
  static const rateLimited = 'ai_rate_limited';
  static const disabled = 'ai_disabled';
  static const upstream = 'ai_upstream_error';
  static const network = 'network';

  @override
  String toString() => message;
}

/// One streamed event from `POST /v1/ai/chat`.
sealed class AiStreamEvent {
  const AiStreamEvent();
}

class AiDelta extends AiStreamEvent {
  const AiDelta(this.text);
  final String text;
}

class AiDone extends AiStreamEvent {
  const AiDone({required this.cached, required this.truncated});
  final bool cached;
  final bool truncated;
}

const _genericError = 'Tanya AI sedang bermasalah. Coba lagi sebentar lagi.';
const _networkError = 'Butuh koneksi internet';

class AiChatRepository {
  AiChatRepository(this._dio);

  final Dio _dio;

  static const _streamTimeout = Duration(minutes: 3);

  /// Sends one message and yields the streamed answer. Throws
  /// [AiChatException] for quota, validation, upstream and network failures.
  Stream<AiStreamEvent> send({
    required AiLessonType lessonType,
    required String lessonId,
    AiChip? chip,
    String message = '',
  }) async* {
    final Response<ResponseBody> response;
    try {
      response = await _dio.post<ResponseBody>(
        '/ai/chat',
        data: {
          'lessonType': lessonType.name,
          'lessonId': lessonId,
          'chip': chip?.wire ?? '',
          'message': message,
        },
        options: Options(
          responseType: ResponseType.stream,
          receiveTimeout: _streamTimeout,
          headers: {'Accept': 'text/event-stream'},
        ),
      );
    } on DioException catch (error) {
      throw await _toException(error);
    }
    final body = response.data;
    if (body == null) {
      throw const AiChatException(AiChatException.upstream, _genericError);
    }
    var finished = false;
    try {
      await for (final event in parseServerSentEvents(body.stream)) {
        switch (event.name) {
          case 'delta':
            final text = event.data['text'];
            if (text is String && text.isNotEmpty) yield AiDelta(text);
          case 'done':
            finished = true;
            yield AiDone(
              cached: event.data['cached'] == true,
              truncated: event.data['truncated'] == true,
            );
          case 'error':
            throw AiChatException(
              event.data['code'] as String? ?? AiChatException.upstream,
              event.data['message'] as String? ?? _genericError,
            );
        }
      }
    } on AiChatException {
      rethrow;
    } on Object {
      throw const AiChatException(AiChatException.network, _networkError);
    }
    if (!finished) {
      // The connection dropped mid-answer; the server released the quota.
      throw const AiChatException(AiChatException.network, _networkError);
    }
  }

  /// Records a watched (or unavailable) rewarded ad as one message credit.
  Future<void> grantAdCredit({required bool watched}) async {
    try {
      await _dio.post<void>(
        '/ai/ad-credits',
        data: {'status': watched ? 'watched' : 'unavailable'},
      );
    } on DioException catch (error) {
      throw await _toException(error);
    }
  }

  Future<List<AiChatMessage>> history(
    AiLessonType lessonType,
    String lessonId,
  ) async {
    try {
      final response = await _dio.get<Map<String, dynamic>>(
        '/ai/conversations/${lessonType.name}/${Uri.encodeComponent(lessonId)}',
      );
      final data = response.data?['data'];
      final messages = data is Map<String, dynamic> ? data['messages'] : null;
      if (messages is! List) return const [];
      return [
        for (final item in messages)
          if (item is Map<String, dynamic>) AiChatMessage.fromJson(item),
      ];
    } on DioException catch (error) {
      throw await _toException(error);
    }
  }

  Future<void> clear(AiLessonType lessonType, String lessonId) async {
    try {
      await _dio.delete<void>(
        '/ai/conversations/${lessonType.name}/${Uri.encodeComponent(lessonId)}',
      );
    } on DioException catch (error) {
      throw await _toException(error);
    }
  }

  static Future<AiChatException> _toException(DioException error) async {
    final response = error.response;
    if (response == null) {
      return const AiChatException(AiChatException.network, _networkError);
    }
    Object? data = response.data;
    if (data is ResponseBody) {
      try {
        final bytes = await data.stream
            .fold<List<int>>(<int>[], (all, chunk) => all..addAll(chunk))
            .timeout(const Duration(seconds: 10));
        data = jsonDecode(utf8.decode(bytes));
      } on Object {
        data = null;
      }
    }
    if (data is Map && data['error'] is Map) {
      final payload = data['error'] as Map;
      final code = payload['code'];
      final message = payload['message'];
      if (code is String) {
        final perAd = payload['messagesPerAd'];
        return AiChatException(
          code,
          message is String && message.isNotEmpty ? message : _genericError,
          messagesPerAd: perAd is num ? perAd.toInt() : 0,
        );
      }
    }
    return const AiChatException(AiChatException.upstream, _genericError);
  }
}

class ServerSentEvent {
  const ServerSentEvent(this.name, this.data);
  final String name;
  final Map<String, dynamic> data;
}

/// Parses a `text/event-stream` byte stream into events with JSON data.
Stream<ServerSentEvent> parseServerSentEvents(Stream<List<int>> bytes) async* {
  var name = '';
  final data = StringBuffer();
  await for (final line
      in bytes
          .cast<List<int>>()
          .transform(utf8.decoder)
          .transform(const LineSplitter())) {
    if (line.isEmpty) {
      if (name.isNotEmpty && data.isNotEmpty) {
        final decoded = jsonDecode(data.toString());
        yield ServerSentEvent(
          name,
          decoded is Map<String, dynamic> ? decoded : const {},
        );
      }
      name = '';
      data.clear();
    } else if (line.startsWith('event:')) {
      name = line.substring(6).trim();
    } else if (line.startsWith('data:')) {
      data.write(line.substring(5).trim());
    }
  }
}

final aiChatRepositoryProvider = Provider<AiChatRepository>(
  (ref) => AiChatRepository(ref.watch(dioProvider)),
);
