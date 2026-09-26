import 'package:dartz/dartz.dart';
import 'package:real_state_mobile/core/error/failures.dart';
import 'package:real_state_mobile/features/property/domain/entities/property.dart';
import 'package:real_state_mobile/features/property/domain/repositories/property_repository.dart';

class CreatePropertyUseCase {
  const CreatePropertyUseCase(this._repository);

  final PropertyRepository _repository;

  Future<Either<Failure, Property>> call(Property property) {
    return _repository.createProperty(property);
  }
}
