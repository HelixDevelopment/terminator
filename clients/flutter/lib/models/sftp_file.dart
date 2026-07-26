class SftpFile {
  final String name;
  final String path;
  final bool isDirectory;
  final int size;
  final DateTime modifiedAt;

  SftpFile({
    required this.name,
    required this.path,
    required this.isDirectory,
    required this.size,
    required this.modifiedAt,
  });
  factory SftpFile.fromJson(Map<String, dynamic> json) {
    return SftpFile(
      name: json['name'] as String,
      path: json['path'] as String,
      isDirectory: json['is_directory'] as bool,
      size: json['size'] as int,
      modifiedAt: DateTime.parse(json['modified_at'] as String),
    );
  }

  Map<String, dynamic> toJson() {
    return {
      'name': name,
      'path': path,
      'is_directory': isDirectory,
      'size': size,
      'modified_at': modifiedAt.toIso8601String(),
    };
  }
}
