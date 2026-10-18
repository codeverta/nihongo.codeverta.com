import 'package:flutter/material.dart';

class ProfileAvatar extends StatelessWidget {
  const ProfileAvatar({
    required this.displayName,
    required this.photoUrl,
    required this.radius,
    super.key,
  });

  final String displayName;
  final String? photoUrl;
  final double radius;

  @override
  Widget build(BuildContext context) {
    final colorScheme = Theme.of(context).colorScheme;
    final imageUrl = photoUrl?.trim() ?? '';
    Widget initial() => Text(
      displayName.trim().isEmpty
          ? 'N'
          : displayName.trim().characters.first.toUpperCase(),
      style: TextStyle(
        color: colorScheme.primary,
        fontSize: radius * 0.8,
        fontWeight: FontWeight.w900,
      ),
    );

    return CircleAvatar(
      radius: radius,
      backgroundColor: colorScheme.primary.withValues(alpha: 0.10),
      child: imageUrl.isEmpty
          ? initial()
          : ClipOval(
              child: Image.network(
                imageUrl,
                width: radius * 2,
                height: radius * 2,
                fit: BoxFit.cover,
                errorBuilder: (_, __, ___) => Center(child: initial()),
                loadingBuilder: (_, child, progress) => progress == null
                    ? child
                    : Center(
                        child: SizedBox(
                          width: radius * 0.55,
                          height: radius * 0.55,
                          child: CircularProgressIndicator(
                            strokeWidth: 2,
                            color: colorScheme.primary,
                          ),
                        ),
                      ),
              ),
            ),
    );
  }
}
