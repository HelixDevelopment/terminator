class Recording {
  final String id;
  final String sessionId;
  final String title;
  final Duration duration;
  final DateTime createdAt;

  Recording({
    required this.id,
    required this.sessionId,
    required this.title,
    required this.duration,
    required this.createdAt,
  });
  factory Recording.fromJson(Map<String, dynamic> json) {
    return Recording(
      id: json['id'] as String,
      sessionId: json['session_id'] as String,
      title: json['title'] as String,
      duration: Duration(milliseconds: json['duration'] as int),
      createdAt: DateTime.parse(json['created_at'] as String),
    );
  }

  Map<String, dynamic> toJson() {
    return {
      'id': id,
      'session_id': sessionId,
      'title': title,
      'duration': duration.inMilliseconds,
      'created_at': createdAt.toIso8601String(),
    };
  }
}
