import 'package:flutter/material.dart';
import 'package:url_launcher/url_launcher.dart';

const _legalBaseUrl = String.fromEnvironment(
  'LEGAL_BASE_URL',
  defaultValue: 'https://nihongo.codeverta.com',
);

const _legalPages = {
  '/privacy-policy': 'Kebijakan Privasi',
  '/terms-of-service': 'Ketentuan Penggunaan',
  '/refund-policy': 'Refund',
  '/delete-account': 'Penghapusan Data',
  '/support': 'Bantuan',
};

Future<void> openLegalPage(BuildContext context, String path) async {
  try {
    final uri = Uri.parse(_legalBaseUrl).resolve(path);
    if (await launchUrl(uri, mode: LaunchMode.externalApplication)) return;
  } catch (_) {
    // The device may not have a browser capable of handling the URL.
  }
  if (context.mounted) {
    ScaffoldMessenger.of(context).showSnackBar(
      const SnackBar(content: Text('Halaman tidak dapat dibuka. Coba lagi.')),
    );
  }
}

class LegalLinks extends StatelessWidget {
  const LegalLinks({super.key});

  @override
  Widget build(BuildContext context) {
    return Wrap(
      alignment: WrapAlignment.center,
      children: [
        for (final page in _legalPages.entries)
          TextButton(
            onPressed: () => openLegalPage(context, page.key),
            child: Text(page.value),
          ),
      ],
    );
  }
}
