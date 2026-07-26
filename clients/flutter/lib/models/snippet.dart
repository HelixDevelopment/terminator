class Snippet {
  final String id;
  final String title;
  final String content;
  final String language;
  final DateTime createdAt;

  Snippet({
    required this.id,
    required this.title,
    required this.content,
    required this.language,
    required this.createdAt,
  });
  factory Snippet.fromJson(Map<String, dynamic> json) {
    return Snippet(
      id: json['id'] as String,
      title: json['title'] as String,
      content: json['content'] as String,
      language: json['language'] as String,
      createdAt: DateTime.parse(json['created_at'] as String),
    );
  }

  Map<String, dynamic> toJson() {
    return {
      'id': id,
      'title': title,
      'content': content,
      'language': language,
      'created_at': createdAt.toIso8601String(),
    };
  }
}
