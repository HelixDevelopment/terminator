class SshKey {
  final String id;
  final String name;
  final String fingerprint;
  final DateTime createdAt;

  SshKey({
    required this.id,
    required this.name,
    required this.fingerprint,
    required this.createdAt,
  });
  factory SshKey.fromJson(Map<String, dynamic> json) {
    return SshKey(
      id: json['id'] as String,
      name: json['name'] as String,
      fingerprint: json['fingerprint'] as String,
      createdAt: DateTime.parse(json['created_at'] as String),
    );
  }

  Map<String, dynamic> toJson() {
    return {
      'id': id,
      'name': name,
      'fingerprint': fingerprint,
      'created_at': createdAt.toIso8601String(),
    };
  }
}
