import 'package:dartz/dartz.dart';
import 'package:real_state_mobile/core/error/failures.dart';
import 'package:real_state_mobile/features/property/domain/entities/property.dart';
import 'package:real_state_mobile/features/property/domain/repositories/property_repository.dart';

class GetPropertiesUseCase {
  const GetPropertiesUseCase(this._repository);

  final PropertyRepository _repository;

  Future<Either<Failure, List<Property>>> call({
    int page = 1,
    int limit = 20,
    String? operation,
    double? minPrice,
    double? maxPrice,
  }) {
    return _repository.getProperties(
      page: page,
      limit: limit,
      operation: operation,
      minPrice: minPrice,
      maxPrice: maxPrice,
    );
  }
}
