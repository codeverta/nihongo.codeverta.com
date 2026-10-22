import 'dart:async';
import 'package:dio/dio.dart';
import 'package:flutter/foundation.dart';
import 'package:flutter/widgets.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:in_app_purchase/in_app_purchase.dart';
import 'package:nihongo_master/core/network/dio_client.dart';
import 'package:nihongo_master/features/auth/presentation/controllers/auth_controller.dart';
import 'package:nihongo_master/features/billing/data/billing_repository.dart';
import 'package:nihongo_master/features/home/presentation/controllers/home_controller.dart';

abstract class PurchaseGateway {
  Stream<List<PurchaseDetails>> get purchases;
  Future<bool> available();
  Future<ProductDetailsResponse> products(Set<String> ids);
  Future<bool> buy(PurchaseParam param);
  Future<void> restore(String token);
  Future<void> complete(PurchaseDetails purchase);
}

class NativePurchaseGateway implements PurchaseGateway {
  final InAppPurchase store = InAppPurchase.instance;
  @override
  Stream<List<PurchaseDetails>> get purchases => store.purchaseStream;
  @override
  Future<bool> available() => store.isAvailable();
  @override
  Future<ProductDetailsResponse> products(Set<String> ids) =>
      store.queryProductDetails(ids);
  @override
  Future<bool> buy(PurchaseParam param) =>
      store.buyNonConsumable(purchaseParam: param);
  @override
  Future<void> restore(String token) =>
      store.restorePurchases(applicationUserName: token);
  @override
  Future<void> complete(PurchaseDetails purchase) =>
      store.completePurchase(purchase);
}

class BillingState {
  const BillingState({
    this.product,
    this.priceIdr,
    this.premium = false,
    this.busy = false,
    this.pending = false,
    this.message,
    this.ready = false,
  });
  final ProductDetails? product;
  final int? priceIdr;
  final bool premium;
  final bool busy;
  final bool pending;
  final bool ready;
  final String? message;
}

final billingControllerProvider =
    StateNotifierProvider<BillingController, BillingState>((ref) {
      final platform =
          !kIsWeb && defaultTargetPlatform == TargetPlatform.android
          ? 'android'
          : !kIsWeb && defaultTargetPlatform == TargetPlatform.iOS
          ? 'ios'
          : '';
      final controller = BillingController(
        repository: BillingRepository(ref.watch(dioProvider)),
        gateway: platform.isEmpty ? null : NativePurchaseGateway(),
        platform: platform,
        onChanged: () async {
          await ref.read(authControllerProvider.notifier).refreshUser();
          ref.invalidate(homeDashboardProvider);
        },
      );
      ref.listen(authControllerProvider.select((s) => s.user?.id), (_, id) {
        unawaited(controller.setUser(id));
      });
      unawaited(controller.setUser(ref.read(authControllerProvider).user?.id));
      return controller;
    });

final lifetimeDisplayPriceProvider = FutureProvider.autoDispose<int>((ref) {
  return BillingRepository(ref.watch(dioProvider)).displayPrice();
});

class BillingController extends StateNotifier<BillingState>
    with WidgetsBindingObserver {
  BillingController({
    required this.repository,
    required this.gateway,
    required this.platform,
    required this.onChanged,
  }) : super(const BillingState()) {
    WidgetsBinding.instance.addObserver(this);
    subscription = gateway?.purchases.listen(
      (items) {
        unawaited(handle(items));
      },
      onError: (Object error) =>
          showMessage('Koneksi store terputus. Coba lagi.'),
    );
  }
  final BillingRepository repository;
  final PurchaseGateway? gateway;
  final String platform;
  final Future<void> Function() onChanged;
  StreamSubscription<List<PurchaseDetails>>? subscription;
  Future<void> queue = Future<void>.value();
  BillingAccount? account;
  String? user;
  final Map<String, PurchaseDetails> retryPurchases = {};
  final Set<String> completed = {};
  final Map<String, PurchaseDetails> waitingForLogin = {};
  bool restoring = false;
  Timer? checkoutTimeout;

  void showMessage(String message, {bool busy = false, bool pending = false}) {
    if (!mounted) return;
    if (!busy) checkoutTimeout?.cancel();
    state = BillingState(
      product: state.product,
      priceIdr: state.priceIdr,
      premium: state.premium,
      ready: state.ready,
      busy: busy,
      pending: pending,
      message: message,
    );
  }

  Future<void> setUser(String? id) async {
    if (id == user) return;
    user = id;
    account = null;
    retryPurchases.clear();
    completed.clear();
    state = const BillingState();
    if (id != null) await refresh();
    if (id != null &&
        user == id &&
        account != null &&
        waitingForLogin.isNotEmpty) {
      final waiting = waitingForLogin.values.toList();
      waitingForLogin.clear();
      await handle(waiting);
    }
    if (id != null &&
        user == id &&
        platform == 'android' &&
        account != null &&
        account!.platforms.contains(platform)) {
      try {
        await gateway?.restore(account!.accountToken);
      } on Object {
        showMessage('Tidak dapat memeriksa pembelian lama. Gunakan Restore.');
      }
    }
  }

  Future<void> refresh() async {
    final id = user;
    if (id == null) {
      showMessage('Masuk ke akun untuk membeli atau restore.');
      return;
    }
    showMessage('Memeriksa akses dan store…', busy: true);
    try {
      final info = await repository.status();
      if (!mounted || user != id) return;
      account = info;
      state = BillingState(
        premium: info.premium,
        priceIdr: info.priceIdr,
        busy: true,
      );
      if (gateway == null ||
          !info.platforms.contains(platform) ||
          !await gateway!.available()) {
        if (mounted && user == id) {
          state = BillingState(
            premium: info.premium,
            priceIdr: info.priceIdr,
            message: 'Pembelian store belum tersedia di perangkat ini.',
          );
        }
        await onChanged();
        return;
      }
      final result = await gateway!.products({info.productId});
      if (!mounted || user != id) return;
      final found = result.productDetails
          .where((p) => p.id == info.productId)
          .firstOrNull;
      state = BillingState(
        premium: info.premium,
        product: found,
        priceIdr: info.priceIdr,
        ready: found != null,
        message: found == null
            ? 'Produk belum tersedia. Coba lagi nanti.'
            : null,
      );
      await onChanged();
    } on DioException catch (error) {
      if (error.response?.statusCode == 401) await onChanged();
      if (mounted && user == id) {
        showMessage(
          error.response?.statusCode == 403
              ? 'Masuk dengan akun Google pribadi untuk membeli atau Restore. Akun demo digunakan bersama.'
              : 'Sesi atau koneksi bermasalah. Masuk lagi jika sesi sudah berakhir.',
        );
      }
    } on Object {
      if (mounted && user == id) {
        showMessage(
          'Tidak dapat memeriksa akses. Periksa koneksi lalu coba lagi.',
        );
      }
    }
  }

  Future<void> buy() async {
    if (state.busy ||
        state.pending ||
        state.premium ||
        !state.ready ||
        account == null ||
        user == null) {
      return;
    }
    showMessage('Selesaikan pembayaran di store.', busy: true);
    checkoutTimeout?.cancel();
    checkoutTimeout = Timer(const Duration(seconds: 90), () {
      showMessage(
        'Store belum memberikan hasil. Gunakan Restore atau coba verifikasi lagi sebelum membeli ulang.',
      );
    });
    try {
      final started = await gateway!.buy(
        PurchaseParam(
          productDetails: state.product!,
          applicationUserName: account!.accountToken,
        ),
      );
      if (!started) {
        showMessage(
          'Google Play tidak melanjutkan pembayaran. Jika dialog tadi ditutup, tidak ada biaya yang dikenakan dan akses tetap sama.',
        );
      }
    } on Object {
      showMessage(
        'Pembayaran tidak dapat dimulai. Coba Restore jika sudah membeli.',
      );
    }
  }

  Future<void> restore() async {
    if (state.busy || gateway == null || account == null || user == null) {
      return;
    }
    restoring = true;
    showMessage('Memulihkan pembelian…', busy: true);
    try {
      await gateway!
          .restore(account!.accountToken)
          .timeout(const Duration(seconds: 25));
      await queue;
      if (restoring) {
        await refresh();
        showMessage(
          state.premium
              ? 'Lifetime Access aktif.'
              : 'Belum ada pembelian yang berhasil dipulihkan. Pastikan akun store benar.',
        );
      }
    } on Object {
      showMessage(
        'Restore gagal. Periksa akun store dan koneksi, lalu coba lagi.',
      );
    } finally {
      restoring = false;
    }
  }

  Future<void> retry() async {
    await refresh();
    await handle(retryPurchases.values.toList());
  }

  Future<void> handle(List<PurchaseDetails> items) async {
    queue = queue.then((_) => _handle(items)).catchError((Object error) {
      showMessage(
        'Pembelian belum selesai. Coba verifikasi lagi atau restore.',
      );
    });
    await queue;
  }

  Future<void> _handle(List<PurchaseDetails> items) async {
    for (final purchase in items) {
      // Google Play can report a canceled or failed billing launch without a
      // PurchaseWrapper, so the plugin emits a PurchaseDetails with an empty
      // product ID. Handle those results before filtering by product ID.
      if (purchase.productID.isEmpty &&
          purchase.status == PurchaseStatus.canceled) {
        showMessage(
          'Pembayaran dibatalkan. Tidak ada biaya yang dikenakan dan akses tidak berubah.',
        );
        if (purchase.pendingCompletePurchase) {
          await gateway!.complete(purchase);
        }
        continue;
      }
      if (purchase.productID.isEmpty &&
          purchase.status == PurchaseStatus.error) {
        showMessage(
          'Google Play tidak dapat memulai pembayaran. Akses tidak berubah; silakan coba lagi.',
        );
        if (purchase.pendingCompletePurchase) {
          await gateway!.complete(purchase);
        }
        continue;
      }
      if (account == null || user == null) {
        waitingForLogin['${purchase.productID}:${purchase.purchaseID}'] =
            purchase;
        continue;
      }
      if (purchase.productID != account!.productId) {
        continue;
      }
      switch (purchase.status) {
        case PurchaseStatus.pending:
          showMessage(
            'Pembayaran menunggu persetujuan. Akses aktif setelah terverifikasi.',
            pending: true,
          );
          continue;
        case PurchaseStatus.canceled:
          showMessage(
            'Pembayaran dibatalkan. Tidak ada akses baru yang diaktifkan.',
          );
          if (purchase.pendingCompletePurchase) {
            await gateway!.complete(purchase);
          }
          continue;
        case PurchaseStatus.error:
          showMessage(
            'Pembayaran gagal. Coba lagi atau Restore jika sudah membeli.',
          );
          if (purchase.pendingCompletePurchase) {
            await gateway!.complete(purchase);
          }
          continue;
        case PurchaseStatus.purchased:
        case PurchaseStatus.restored:
          break;
      }
      final key =
          '${purchase.productID}:${purchase.purchaseID ?? purchase.verificationData.serverVerificationData}';
      if (completed.contains(key)) continue;
      retryPurchases[key] = purchase;
      final id = user;
      showMessage('Memverifikasi pembelian…', busy: true);
      try {
        final result = await repository.verify(
          platform: platform,
          productId: purchase.productID,
          proof: purchase.verificationData.serverVerificationData,
          transactionId: platform == 'ios' ? purchase.purchaseID : null,
        );
        if (!mounted || user != id) continue;
        if (result == 'pending') {
          showMessage('Pembayaran masih tertunda di store.', pending: true);
          continue;
        }
        // Finish only after server verification and durable ownership have succeeded.
        if (purchase.pendingCompletePurchase) await gateway!.complete(purchase);
        completed.add(key);
        retryPurchases.remove(key);
        restoring = false;
        await refresh();
        if (!mounted || user != id) continue;
        showMessage(
          result == 'purchased'
              ? 'Lifetime Access berhasil diaktifkan.'
              : 'Pembelian telah dicabut atau dikembalikan oleh store.',
        );
      } on DioException catch (error) {
        if (error.response?.statusCode == 401) await onChanged();
        if (!mounted || user != id) continue;
        showMessage(
          error.response?.statusCode == 402
              ? 'Pembelian ini terkait akun aplikasi lain. Masuk ke akun pembeli lalu Restore.'
              : 'Verifikasi belum selesai. Jangan membeli lagi; coba verifikasi atau Restore.',
        );
      } on Object {
        if (mounted && user == id) {
          showMessage(
            'Pembelian belum selesai. Coba verifikasi lagi atau Restore.',
          );
        }
      }
    }
  }

  @override
  void didChangeAppLifecycleState(AppLifecycleState state) {
    if (state == AppLifecycleState.resumed &&
        user != null &&
        !this.state.busy) {
      unawaited(retry());
    }
  }

  @override
  void dispose() {
    WidgetsBinding.instance.removeObserver(this);
    checkoutTimeout?.cancel();
    unawaited(subscription?.cancel());
    super.dispose();
  }
}
