class AuditLog {
  final String id;
  final String actor;
  final String action;
  final String target;
  final DateTime timestamp;

  AuditLog({
    required this.id,
    required this.actor,
    required this.action,
    required this.target,
    required this.timestamp,
  });
  factory AuditLog.fromJson(Map<String, dynamic> json) {
    return AuditLog(
      id: json['id'] as String,
      actor: json['actor'] as String,
      action: json['action'] as String,
      target: json['target'] as String,
      timestamp: DateTime.parse(json['timestamp'] as String),
    );
  }

  Map<String, dynamic> toJson() {
    return {
      'id': id,
      'actor': actor,
      'action': action,
      'target': target,
      'timestamp': timestamp.toIso8601String(),
    };
  }
}
