class PortForward {
  final String id;
  final String hostId;
  final int localPort;
  final int remotePort;
  final String remoteHost;
  final bool active;

  PortForward({
    required this.id,
    required this.hostId,
    required this.localPort,
    required this.remotePort,
    this.remoteHost = 'localhost',
    this.active = false,
  });
  factory PortForward.fromJson(Map<String, dynamic> json) {
    return PortForward(
      id: json['id'] as String,
      hostId: json['host_id'] as String,
      localPort: json['local_port'] as int,
      remotePort: json['remote_port'] as int,
      remoteHost: json['remote_host'] as String,
      active: json['active'] as bool,
    );
  }

  Map<String, dynamic> toJson() {
    return {
      'id': id,
      'host_id': hostId,
      'local_port': localPort,
      'remote_port': remotePort,
      'remote_host': remoteHost,
      'active': active,
    };
  }
}
