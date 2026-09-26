import 'package:dartz/dartz.dart';
import 'package:real_state_mobile/core/error/failures.dart';
import 'package:real_state_mobile/features/property/domain/repositories/property_repository.dart';

class DeletePropertyUseCase {
  const DeletePropertyUseCase(this._repository);

  final PropertyRepository _repository;

  Future<Either<Failure, bool>> call(String id) {
    return _repository.deleteProperty(id);
  }
}
