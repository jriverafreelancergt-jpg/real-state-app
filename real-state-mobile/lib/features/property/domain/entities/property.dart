enum OperationType { sale, rent }
enum PropertyStatus { available, reserved, sold, rented, inactive }

class Property {
  const Property({
    required this.id,
    required this.title,
    required this.price,
    required this.operation,
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
  final OperationType operation;
  final PropertyStatus status;
  final String? primaryImageUrl;
  final String? description;
  final String currency;
  final String propertyType;
  final String? address;
  final double? latitude;
  final double? longitude;
  final DateTime? createdAt;
}
