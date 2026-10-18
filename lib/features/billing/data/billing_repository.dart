import 'package:dio/dio.dart';

class BillingAccount {
  const BillingAccount({
    required this.productId,
    required this.accountToken,
    required this.platforms,
    required this.premium,
    required this.priceIdr,
  });
  factory BillingAccount.fromJson(Map<String, dynamic> data) => BillingAccount(
    productId: data['productId'] as String,
    accountToken: data['accountToken'] as String,
    platforms: (data['platforms'] as List<dynamic>).cast<String>(),
    premium: (data['entitlement'] as Map<String, dynamic>)['tier'] == 'premium',
    priceIdr: data['priceIdr'] as int?,
  );
  final String productId;
  final String accountToken;
  final List<String> platforms;
  final bool premium;
  final int? priceIdr;
}

class BillingRepository {
  BillingRepository(this.dio);
  final Dio dio;

  Future<int> displayPrice() async {
    final response = await dio.get<Map<String, dynamic>>('/pricing');
    final data = response.data?['data'];
    if (data is! Map<String, dynamic> || data['lifetimePriceIdr'] is! int) {
      throw const FormatException('Format harga Lifetime tidak valid.');
    }
    return data['lifetimePriceIdr'] as int;
  }

  Future<BillingAccount> status() async {
    final response = await dio.get<Map<String, dynamic>>(
      '/billing',
      options: Options(receiveTimeout: const Duration(seconds: 30)),
    );
    return BillingAccount.fromJson(
      response.data!['data'] as Map<String, dynamic>,
    );
  }

  Future<String> verify({
    required String platform,
    required String productId,
    required String proof,
    String? transactionId,
  }) async {
    final response = await dio.post<Map<String, dynamic>>(
      '/billing/verify',
      options: Options(receiveTimeout: const Duration(seconds: 30)),
      data: {
        'platform': platform,
        'productId': productId,
        'verificationData': proof,
        'transactionId': transactionId ?? '',
      },
    );
    return (response.data!['data'] as Map<String, dynamic>)['state'] as String;
  }
}
