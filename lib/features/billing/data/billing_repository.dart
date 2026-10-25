import 'package:dio/dio.dart';
import 'package:nihongo_master/core/auth/offline_access.dart';

class BillingAccount {
  const BillingAccount({
    required this.productId,
    required this.accountToken,
    required this.platforms,
    required this.premium,
    required this.priceIdr,
    this.lifetime = false,
    this.offlineLease,
  });
  factory BillingAccount.fromJson(Map<String, dynamic> data) => BillingAccount(
    productId: data['productId'] as String,
    accountToken: data['accountToken'] as String,
    platforms: (data['platforms'] as List<dynamic>).cast<String>(),
    premium: (data['entitlement'] as Map<String, dynamic>)['tier'] == 'premium',
    lifetime:
        (data['entitlement'] as Map<String, dynamic>)['isLifetime'] == true,
    priceIdr: data['priceIdr'] as int?,
    offlineLease: data['offlineLease'] as String?,
  );
  final String productId;
  final String accountToken;
  final List<String> platforms;
  final bool premium;
  final bool lifetime;
  final int? priceIdr;
  final String? offlineLease;
}

class BillingRepository {
  BillingRepository(this.dio, {this.offlineAccess});
  final Dio dio;
  final OfflineAccessStore? offlineAccess;

  Future<int> displayPrice() async {
    final response = await dio.get<Map<String, dynamic>>('/pricing');
    final data = response.data?['data'];
    if (data is! Map<String, dynamic> || data['lifetimePriceIdr'] is! int) {
      throw const FormatException('Format harga Lifetime tidak valid.');
    }
    return data['lifetimePriceIdr'] as int;
  }

  Future<BillingAccount> status({String? userId}) async {
    final response = await dio.get<Map<String, dynamic>>(
      '/billing',
      options: Options(receiveTimeout: const Duration(seconds: 30)),
    );
    final account = BillingAccount.fromJson(
      response.data!['data'] as Map<String, dynamic>,
    );
    await _saveSnapshot(account, userId);
    return account;
  }

  Future<String> verify({
    required String platform,
    required String productId,
    required String proof,
    String? transactionId,
    String? userId,
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
    final data = response.data!['data'] as Map<String, dynamic>;
    final entitlement = data['entitlement'];
    if (entitlement is Map<String, dynamic>) {
      final account = BillingAccount(
        productId: productId,
        accountToken: '',
        platforms: const [],
        premium: entitlement['tier'] == 'premium',
        lifetime: entitlement['isLifetime'] == true,
        priceIdr: null,
        offlineLease: data['offlineLease'] as String?,
      );
      await _saveSnapshot(account, userId);
    }
    return data['state'] as String;
  }

  Future<void> _saveSnapshot(BillingAccount account, String? userId) async {
    if (userId == null || userId.isEmpty) return;
    await offlineAccess?.applyServerSnapshot(
      userId: userId,
      premium: account.premium,
      lifetime: account.lifetime,
      signedLease: account.offlineLease,
    );
  }
}
