import 'package:dio/dio.dart';
import 'package:real_state_mobile/features/auth/data/dto/auth_tokens_dto.dart';
import 'package:real_state_mobile/features/auth/data/dto/login_request_dto.dart';

abstract class AuthRemoteDataSource {
  Future<AuthTokensDto> login({
    required String email,
    required String password,
  });

  Future<AuthTokensDto> refreshToken({
    required String refreshToken,
  });
}

class AuthRemoteDataSourceImpl implements AuthRemoteDataSource {
  const AuthRemoteDataSourceImpl(this._dio);

  final Dio _dio;

  @override
  Future<AuthTokensDto> login({
    required String email,
    required String password,
  }) async {
    final response = await _dio.post(
      '/auth/login',
      data: LoginRequestDto(email: email, password: password).toJson(),
    );

    return AuthTokensDto.fromJson(response.data as Map<String, dynamic>);
  }

  @override
  Future<AuthTokensDto> refreshToken({
    required String refreshToken,
  }) async {
    final response = await _dio.post(
      '/auth/refresh',
      data: {'refreshToken': refreshToken},
    );

    return AuthTokensDto.fromJson(response.data as Map<String, dynamic>);
  }
}
