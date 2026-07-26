class Session {
  final String id;
  final String hostId;
  final DateTime startedAt;
  final DateTime? endedAt;
  final String protocol;

  Session({
    required this.id,
    required this.hostId,
    required this.startedAt,
    this.endedAt,
    this.protocol = 'ssh',
  });
  factory Session.fromJson(Map<String, dynamic> json) {
    return Session(
      id: json['id'] as String,
      hostId: json['host_id'] as String,
      startedAt: DateTime.parse(json['started_at'] as String),
      endedAt: json['ended_at'] == null ? null : DateTime.parse(json['ended_at'] as String),
      protocol: json['protocol'] as String,
    );
  }

  Map<String, dynamic> toJson() {
    return {
      'id': id,
      'host_id': hostId,
      'started_at': startedAt.toIso8601String(),
      'ended_at': endedAt,
      'protocol': protocol,
    };
  }
}
