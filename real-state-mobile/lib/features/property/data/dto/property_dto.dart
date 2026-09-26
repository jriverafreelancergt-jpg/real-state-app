class PropertyDto {
  factory PropertyDto.fromJson(Map<String, dynamic> json) {
    return PropertyDto(
      id: json['id'] as String? ?? '',
      title: json['title'] as String? ?? '',
      price: (json['price'] as num?)?.toDouble() ?? 0.0,
      operationType: (json['operation_type'] as String?) ??
          (json['operation'] as String?) ??
          'sale',
      status: (json['status'] as String?) ?? 'available',
      primaryImageUrl: json['primary_image_url'] as String? ??
          json['primaryImageUrl'] as String?,
      description: json['description'] as String?,
      currency: json['currency'] as String? ?? 'GTQ',
      propertyType: json['type'] as String? ??
          json['property_type'] as String? ??
          'HOUSE',
      address: json['address'] as String?,
      latitude: (json['latitude'] as num?)?.toDouble(),
      longitude: (json['longitude'] as num?)?.toDouble(),
      createdAt: json['created_at'] != null
          ? DateTime.tryParse(json['created_at'] as String)
          : null,
    );
  }

  const PropertyDto({
    required this.id,
    required this.title,
    required this.price,
    required this.operationType,
    required this.status,
    this.primaryImageUrl,
    this.description,
    this.currency = 'GTQ',
    this.propertyType = 'HOUSE',
    this.address,
    this.latitude,
    this.longitude,
    this.createdAt,
  });

  final String id;
  final String title;
  final double price;
  final String operationType;
  final String status;
  final String? primaryImageUrl;
  final String? description;
  final String currency;
  final String propertyType;
  final String? address;
  final double? latitude;
  final double? longitude;
  final DateTime? createdAt;

  Map<String, dynamic> toJson() {
    return {
      'id': id,
      'title': title,
      'price': price,
      'operation_type': operationType,
      'status': status,
      'primary_image_url': primaryImageUrl,
      'description': description,
      'currency': currency,
      'type': propertyType,
      'address': address,
      'latitude': latitude,
      'longitude': longitude,
      'created_at': createdAt?.toUtc().toIso8601String(),
    };
  }
}
