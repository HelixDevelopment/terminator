class Notification {
  final String id;
  final String title;
  final String body;
  final bool read;
  final DateTime createdAt;

  Notification({
    required this.id,
    required this.title,
    required this.body,
    this.read = false,
    required this.createdAt,
  });
  factory Notification.fromJson(Map<String, dynamic> json) {
    return Notification(
      id: json['id'] as String,
      title: json['title'] as String,
      body: json['body'] as String,
      read: json['read'] as bool,
      createdAt: DateTime.parse(json['created_at'] as String),
    );
  }

  Map<String, dynamic> toJson() {
    return {
      'id': id,
      'title': title,
      'body': body,
      'read': read,
      'created_at': createdAt.toIso8601String(),
    };
  }
}
