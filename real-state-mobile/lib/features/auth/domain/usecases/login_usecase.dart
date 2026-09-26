import 'package:dartz/dartz.dart';
import 'package:real_state_mobile/core/error/failures.dart';
import 'package:real_state_mobile/features/auth/domain/entities/auth_session.dart';
import 'package:real_state_mobile/features/auth/domain/repositories/auth_repository.dart';

class LoginUseCase {
  const LoginUseCase(this._repository);

  final AuthRepository _repository;

  Future<Either<Failure, AuthSession>> call({
    required String email,
    required String password,
  }) {
    return _repository.login(email: email, password: password);
  }
}
