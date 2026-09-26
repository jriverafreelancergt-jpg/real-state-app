import 'package:dartz/dartz.dart';
import 'package:real_state_mobile/core/error/failures.dart';
import 'package:real_state_mobile/features/property/domain/entities/property_media.dart';
import 'package:real_state_mobile/features/property/domain/repositories/property_repository.dart';

class UploadPropertyMediaUseCase {
  const UploadPropertyMediaUseCase(this._repository);

  final PropertyRepository _repository;

  Future<Either<Failure, PropertyMedia>> call({
    required String propertyId,
    required String url,
    required String type,
    bool isPrimary = false,
  }) {
    return _repository.uploadMedia(
      propertyId: propertyId,
      url: url,
      type: type,
      isPrimary: isPrimary,
    );
  }
}
