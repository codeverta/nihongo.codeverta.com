import 'dart:async';
import 'package:dio/dio.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:in_app_purchase/in_app_purchase.dart';
import 'package:nihongo_master/core/routing/app_router.dart';
import 'package:nihongo_master/features/billing/data/billing_repository.dart';
import 'package:nihongo_master/features/billing/presentation/billing_controller.dart';

void main() {
  TestWidgetsFlutterBinding.ensureInitialized();
  late FakeBilling repository;
  late FakeGateway gateway;
  late BillingController controller;
  setUp(() async {
    repository = FakeBilling();
    gateway = FakeGateway();
    controller = BillingController(
      repository: repository,
      gateway: gateway,
      platform: 'android',
      onChanged: () async {},
    );
    await controller.setUser('buyer');
  });
  tearDown(() async {
    controller.dispose();
    await gateway.stream.close();
  });

  test('premium route policy keeps N4 lessons and N5 exams free', () {
    for (final route in [
      '/learn/grammar?level=N3',
      '/learn/reading?level=N2',
      '/learn/vocabulary/lessons/n1-daily',
      '/exams/n4/simulation',
    ]) {
      expect(requiresPremium(Uri.parse(route)), isTrue);
    }
    for (final route in [
      '/learn/grammar?level=N4',
      '/exams/n5/simulation',
      '/home',
    ]) {
      expect(requiresPremium(Uri.parse(route)), isFalse);
    }
  });
  test(
    'buy uses store product and server account binding; pending/cancel do not grant',
    () async {
      expect(controller.state.product!.price, 'Rp99.000');
      expect(controller.state.priceIdr, 149000);
      await controller.buy();
      expect(gateway.param!.applicationUserName, 'account-token');
      await controller.handle([purchase(PurchaseStatus.pending)]);
      expect(controller.state.pending, isTrue);
      expect(controller.state.premium, isFalse);
      await controller.handle([purchase(PurchaseStatus.canceled)]);
      expect(controller.state.busy, isFalse);
      expect(controller.state.pending, isFalse);
      expect(repository.verifications, 0);
    },
  );
  test(
    'successful duplicate notifications deliver entitlement and finish once',
    () async {
      final paid = purchase(PurchaseStatus.purchased)
        ..pendingCompletePurchase = true;
      await controller.handle([paid, paid]);
      expect(controller.state.premium, isTrue);
      expect(repository.verifications, 1);
      expect(gateway.finishes, 1);
      await controller.buy();
      expect(gateway.buys, 0);
      repository.premium = false;
      await controller.refresh();
      expect(controller.state.premium, isFalse);
    },
  );
  test(
    'offline verification leaves transaction unfinished and retries safely',
    () async {
      repository.fail = true;
      final paid = purchase(PurchaseStatus.purchased)
        ..pendingCompletePurchase = true;
      await controller.handle([paid]);
      expect(gateway.finishes, 0);
      expect(controller.state.premium, isFalse);
      repository.fail = false;
      await controller.retry();
      expect(gateway.finishes, 1);
      expect(controller.state.premium, isTrue);
    },
  );
  test(
    'finish failure retries already verified purchase without losing access',
    () async {
      gateway.failFinish = true;
      final paid = purchase(PurchaseStatus.purchased)
        ..pendingCompletePurchase = true;
      await controller.handle([paid]);
      expect(controller.retryPurchases, isNotEmpty);
      gateway.failFinish = false;
      await controller.retry();
      expect(controller.state.premium, isTrue);
      expect(controller.retryPurchases, isEmpty);
    },
  );
  test('account conflict does not finish or transfer a purchase', () async {
    repository.conflict = true;
    final paid = purchase(PurchaseStatus.restored)
      ..pendingCompletePurchase = true;
    await controller.handle([paid]);
    expect(controller.state.message, contains('akun aplikasi lain'));
    expect(controller.state.premium, isFalse);
    expect(gateway.finishes, 0);
  });
  test('logout during verification does not unlock the next account', () async {
    final delayed = Completer<String>();
    repository.delayed = delayed;
    final paid = purchase(PurchaseStatus.purchased)
      ..pendingCompletePurchase = true;
    final processing = controller.handle([paid]);
    await Future<void>.delayed(Duration.zero);
    await controller.setUser(null);
    delayed.complete('purchased');
    await processing;
    expect(controller.state.premium, isFalse);
    expect(gateway.finishes, 0);
  });
  test('empty restore exits busy state without granting premium', () async {
    await controller.restore();
    expect(controller.state.busy, isFalse);
    expect(controller.state.premium, isFalse);
    expect(controller.state.message, contains('Belum ada pembelian'));
  });
  test('transaction delivered before login is verified after login', () async {
    await controller.setUser(null);
    final paid = purchase(PurchaseStatus.purchased)
      ..pendingCompletePurchase = true;
    await controller.handle([paid]);
    expect(repository.verifications, 0);
    expect(gateway.finishes, 0);
    await controller.setUser('buyer');
    expect(controller.state.premium, isTrue);
    expect(gateway.finishes, 1);
  });
}

PurchaseDetails purchase(PurchaseStatus status) => PurchaseDetails(
  productID: 'nihongo_master_lifetime',
  purchaseID: 'transaction',
  verificationData: PurchaseVerificationData(
    localVerificationData: '',
    serverVerificationData: 'store-proof',
    source: 'google_play',
  ),
  transactionDate: '1',
  status: status,
);

class FakeBilling extends BillingRepository {
  FakeBilling() : super(Dio());
  bool premium = false;
  bool fail = false;
  bool conflict = false;
  int verifications = 0;
  Completer<String>? delayed;
  @override
  Future<BillingAccount> status() async => BillingAccount(
    productId: 'nihongo_master_lifetime',
    accountToken: 'account-token',
    platforms: ['android', 'ios'],
    premium: premium,
    priceIdr: 149000,
  );
  @override
  Future<String> verify({
    required String platform,
    required String productId,
    required String proof,
    String? transactionId,
  }) async {
    verifications++;
    if (fail || conflict) {
      throw DioException(
        requestOptions: RequestOptions(),
        response: conflict
            ? Response<dynamic>(
                requestOptions: RequestOptions(),
                statusCode: 402,
              )
            : null,
      );
    }
    if (delayed != null) return delayed!.future;
    premium = true;
    return 'purchased';
  }
}

class FakeGateway implements PurchaseGateway {
  final stream = StreamController<List<PurchaseDetails>>();
  bool failFinish = false;
  int finishes = 0;
  int buys = 0;
  PurchaseParam? param;
  @override
  Stream<List<PurchaseDetails>> get purchases => stream.stream;
  @override
  Future<bool> available() async => true;
  @override
  Future<ProductDetailsResponse> products(Set<String> ids) async =>
      ProductDetailsResponse(
        productDetails: [
          ProductDetails(
            id: ids.first,
            title: 'Lifetime',
            description: 'Lifetime access',
            price: 'Rp99.000',
            rawPrice: 99000,
            currencyCode: 'IDR',
          ),
        ],
        notFoundIDs: [],
      );
  @override
  Future<bool> buy(PurchaseParam purchaseParam) async {
    buys++;
    param = purchaseParam;
    return true;
  }

  @override
  Future<void> restore(String token) async {}
  @override
  Future<void> complete(PurchaseDetails purchase) async {
    if (failFinish) throw StateError('store unavailable');
    finishes++;
  }
}
