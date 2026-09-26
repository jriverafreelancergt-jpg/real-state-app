import '../../domain/entities/property.dart';
import '../dto/property_dto.dart';

class PropertyMapper {
  static Property toEntity(PropertyDto dto) {
    return Property(
      id: dto.id,
      title: dto.title,
      price: dto.price,
      operation: _mapOperation(dto.operationType),
      status: _mapStatus(dto.status),
      primaryImageUrl: dto.primaryImageUrl,
      description: dto.description,
      currency: dto.currency,
      propertyType: dto.propertyType,
      address: dto.address,
      latitude: dto.latitude,
      longitude: dto.longitude,
      createdAt: dto.createdAt,
    );
  }

  static PropertyDto toDto(Property entity) {
    return PropertyDto(
      id: entity.id,
      title: entity.title,
      price: entity.price,
      operationType: entity.operation.name,
      status: entity.status.name,
      primaryImageUrl: entity.primaryImageUrl,
      description: entity.description,
      currency: entity.currency,
      propertyType: entity.propertyType,
      address: entity.address,
      latitude: entity.latitude,
      longitude: entity.longitude,
      createdAt: entity.createdAt,
    );
  }

  static OperationType _mapOperation(String op) {
    final normalized = op.trim().toLowerCase();
    return normalized == 'rent' ? OperationType.rent : OperationType.sale;
  }

  static PropertyStatus _mapStatus(String status) {
    final normalized = status.trim().toLowerCase();

    switch (normalized) {
      case 'reserved':
        return PropertyStatus.reserved;
      case 'sold':
        return PropertyStatus.sold;
      case 'rented':
        return PropertyStatus.rented;
      case 'inactive':
        return PropertyStatus.inactive;
      default:
        return PropertyStatus.available;
    }
  }
}
