import 'package:flutter_test/flutter_test.dart';
import 'package:nihongo_master/core/network/remote_asset_image.dart';

void main() {
  test('learning images resolve against API origin or retain CDN URL', () {
    const api = 'https://api.example.com/v1';
    const asset = 'assets/images/reading/story-01.png';
    expect(
      resolveAssetImageUrl(api, '', asset),
      'https://api.example.com/$asset',
    );
    expect(
      resolveAssetImageUrl(api, '/$asset', asset),
      'https://api.example.com/$asset',
    );
    expect(
      resolveAssetImageUrl(api, 'https://cdn.example.com/$asset', asset),
      'https://cdn.example.com/$asset',
    );
    expect(resolveAssetImageUrl(api, '', ''), isEmpty);
  });
}
