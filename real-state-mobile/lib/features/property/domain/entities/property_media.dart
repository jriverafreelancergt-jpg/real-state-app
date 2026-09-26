class PropertyMedia {
  const PropertyMedia({
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
}
