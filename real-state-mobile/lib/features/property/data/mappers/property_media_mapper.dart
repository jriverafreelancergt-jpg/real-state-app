import '../../domain/entities/property_media.dart';
import '../dto/property_media_dto.dart';

class PropertyMediaMapper {
  static PropertyMedia toEntity(PropertyMediaDto dto) {
    return PropertyMedia(
      id: dto.id,
      propertyId: dto.propertyId,
      url: dto.url,
      type: dto.type,
      thumbnailUrl: dto.thumbnailUrl,
      width: dto.width,
      height: dto.height,
      sortOrder: dto.sortOrder,
      isPrimary: dto.isPrimary,
    );
  }
}
