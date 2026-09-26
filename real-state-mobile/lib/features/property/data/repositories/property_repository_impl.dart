import 'package:dartz/dartz.dart';
import 'package:real_state_mobile/core/error/failures.dart';
import 'package:real_state_mobile/features/property/data/datasources/property_remote_data_source.dart';
import 'package:real_state_mobile/features/property/data/mappers/property_mapper.dart';
import 'package:real_state_mobile/features/property/data/mappers/property_media_mapper.dart';
import 'package:real_state_mobile/features/property/domain/entities/property.dart';
import 'package:real_state_mobile/features/property/domain/entities/property_media.dart';
import 'package:real_state_mobile/features/property/domain/repositories/property_repository.dart';

class PropertyRepositoryImpl implements PropertyRepository {
  const PropertyRepositoryImpl(this._remoteDataSource);

  final PropertyRemoteDataSource _remoteDataSource;

  @override
  Future<Either<Failure, List<Property>>> getProperties({
    int page = 1,
    int limit = 20,
    String? operation,
    double? minPrice,
    double? maxPrice,
  }) async {
    try {
      final dtoList = await _remoteDataSource.getProperties(
        page: page,
        limit: limit,
        operation: operation,
        minPrice: minPrice,
        maxPrice: maxPrice,
      );

      final entities = dtoList.map(PropertyMapper.toEntity).toList();
      return Right(entities);
    } catch (_) {
      return const Left(ServerFailure('No se pudo cargar la lista de propiedades'));
    }
  }

  @override
  Future<Either<Failure, Property>> getPropertyById(String id) async {
    try {
      final dto = await _remoteDataSource.getPropertyById(id);
      return Right(PropertyMapper.toEntity(dto));
    } catch (_) {
      return Left(ServerFailure('No se pudo cargar la propiedad $id'));
    }
  }

  @override
  Future<Either<Failure, Property>> createProperty(Property property) async {
    try {
      final dto = await _remoteDataSource.createProperty(PropertyMapper.toDto(property));
      return Right(PropertyMapper.toEntity(dto));
    } catch (_) {
      return const Left(ServerFailure('No se pudo crear la propiedad'));
    }
  }

  @override
  Future<Either<Failure, Property>> updateProperty(String id, Property property) async {
    try {
      final dto = await _remoteDataSource.updateProperty(id, PropertyMapper.toDto(property));
      return Right(PropertyMapper.toEntity(dto));
    } catch (_) {
      return Left(ServerFailure('No se pudo actualizar la propiedad $id'));
    }
  }

  @override
  Future<Either<Failure, bool>> deleteProperty(String id) async {
    try {
      final deleted = await _remoteDataSource.deleteProperty(id);
      return Right(deleted);
    } catch (_) {
      return Left(ServerFailure('No se pudo eliminar la propiedad $id'));
    }
  }

  @override
  Future<Either<Failure, PropertyMedia>> uploadMedia({
    required String propertyId,
    required String url,
    required String type,
    bool isPrimary = false,
  }) async {
    try {
      final dto = await _remoteDataSource.uploadMedia(
        propertyId: propertyId,
        url: url,
        type: type,
        isPrimary: isPrimary,
      );
      return Right(PropertyMediaMapper.toEntity(dto));
    } catch (_) {
      return const Left(ServerFailure('No se pudo subir la imagen'));
    }
  }
}
