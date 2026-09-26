class PropertyMediaDto {
  factory PropertyMediaDto.fromJson(Map<String, dynamic> json) {
    return PropertyMediaDto(
      id: json['id'] as String? ?? '',
      propertyId: json['property_id'] as String? ?? json['propertyId'] as String? ?? '',
      url: json['url'] as String? ?? '',
      type: json['type'] as String? ?? 'IMAGE',
      thumbnailUrl: json['thumbnail_url'] as String? ?? json['thumbnailUrl'] as String?,
      width: json['width'] as int? ?? 0,
      height: json['height'] as int? ?? 0,
      sortOrder: json['sort_order'] as int? ?? 0,
      isPrimary: json['is_primary'] as bool? ?? json['isPrimary'] as bool? ?? false,
    );
  }

  const PropertyMediaDto({
    required this.id,
    required this.propertyId,
    required this.url,
    required this.type,
    this.thumbnailUrl,
    this.width,
    this.height,
    this.sortOrder = 0,
    this.isPrimary = false,
  });

  final String id;
  final String propertyId;
  final String url;
  final String type;
  final String? thumbnailUrl;
  final int? width;
  final int? height;
  final int sortOrder;
  final bool isPrimary;

  Map<String, dynamic> toJson() {
    return {
      'id': id,
      'property_id': propertyId,
      'url': url,
      'type': type,
      'thumbnail_url': thumbnailUrl,
      'width': width,
      'height': height,
      'sort_order': sortOrder,
      'is_primary': isPrimary,
    };
  }
}
