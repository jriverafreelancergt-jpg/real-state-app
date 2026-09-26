import 'package:dartz/dartz.dart';
import 'package:real_state_mobile/core/error/failures.dart';
import 'package:real_state_mobile/features/property/domain/entities/property.dart';
import 'package:real_state_mobile/features/property/domain/entities/property_media.dart';

abstract class PropertyRepository {
  Future<Either<Failure, List<Property>>> getProperties({
    int page = 1,
    int limit = 20,
    String? operation,
    double? minPrice,
    double? maxPrice,
  });

  Future<Either<Failure, Property>> getPropertyById(String id);

  Future<Either<Failure, Property>> createProperty(Property property);

  Future<Either<Failure, Property>> updateProperty(String id, Property property);

  Future<Either<Failure, bool>> deleteProperty(String id);

  Future<Either<Failure, PropertyMedia>> uploadMedia({
    required String propertyId,
    required String url,
    required String type,
    bool isPrimary = false,
  });
}
