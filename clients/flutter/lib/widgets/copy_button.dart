import 'package:flutter/material.dart';
import 'package:flutter/services.dart';

class CopyButton extends StatelessWidget {
  final String textToCopy;

  const CopyButton({super.key, required this.textToCopy});

  @override
  Widget build(BuildContext context) {
    return IconButton(
      icon: const Icon(Icons.copy),
      onPressed: () async {
        await Clipboard.setData(ClipboardData(text: textToCopy));
        if (context.mounted) {
          ScaffoldMessenger.of(context).showSnackBar(
            const SnackBar(content: Text('Copied to clipboard')),
          );
        }
      },
    );
  }
}
