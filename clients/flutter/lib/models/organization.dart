class Organization {
  final String id;
  final String name;
  final String slug;
  final String? logoUrl;
  final DateTime createdAt;

  Organization({
    required this.id,
    required this.name,
    required this.slug,
    this.logoUrl,
    required this.createdAt,
  });
  factory Organization.fromJson(Map<String, dynamic> json) {
    return Organization(
      id: json['id'] as String,
      name: json['name'] as String,
      slug: json['slug'] as String,
      logoUrl: json['logo_url'] as String?,
      createdAt: DateTime.parse(json['created_at'] as String),
    );
  }

  Map<String, dynamic> toJson() {
    return {
      'id': id,
      'name': name,
      'slug': slug,
      'logo_url': logoUrl,
      'created_at': createdAt.toIso8601String(),
    };
  }
}
