import 'package:dartz/dartz.dart';
import 'package:real_state_mobile/core/error/failures.dart';
import 'package:real_state_mobile/core/storage/secure_storage_service.dart';
import 'package:real_state_mobile/features/auth/data/datasources/auth_remote_data_source.dart';
import 'package:real_state_mobile/features/auth/domain/entities/auth_session.dart';
import 'package:real_state_mobile/features/auth/domain/repositories/auth_repository.dart';

class AuthRepositoryImpl implements AuthRepository {
  const AuthRepositoryImpl({
    required AuthRemoteDataSource remoteDataSource,
    required SecureStorageService storage,
  })  : _remoteDataSource = remoteDataSource,
        _storage = storage;

  final AuthRemoteDataSource _remoteDataSource;
  final SecureStorageService _storage;

  @override
  Future<Either<Failure, AuthSession>> login({
    required String email,
    required String password,
  }) async {
    try {
      final result = await _remoteDataSource.login(
        email: email,
        password: password,
      );

      await _storage.saveAccessToken(result.accessToken);
      await _storage.saveRefreshToken(result.refreshToken);
      await _storage.saveUserEmail(email);

      return Right(
        AuthSession(
          accessToken: result.accessToken,
          refreshToken: result.refreshToken,
          userEmail: email,
          expiresIn: result.expiresIn,
        ),
      );
    } catch (_) {
      return const Left(ServerFailure('No se pudo iniciar sesión'));
    }
  }

  @override
  Future<Either<Failure, AuthSession>> refreshToken({
    required String refreshToken,
  }) async {
    try {
      final result = await _remoteDataSource.refreshToken(
        refreshToken: refreshToken,
      );

      final userEmail = await _storage.readUserEmail() ?? '';

      await _storage.saveAccessToken(result.accessToken);
      await _storage.saveRefreshToken(result.refreshToken);

      return Right(
        AuthSession(
          accessToken: result.accessToken,
          refreshToken: result.refreshToken,
          userEmail: userEmail,
          expiresIn: result.expiresIn,
        ),
      );
    } catch (_) {
      return const Left(UnauthorizedFailure('La sesión expiró o el refresh token es inválido'));
    }
  }

  @override
  Future<void> logout() async {
    await _storage.clearSession();
  }
}
