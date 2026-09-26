import 'package:dio/dio.dart';
import 'package:real_state_mobile/core/config/app_config.dart';
import 'package:real_state_mobile/core/storage/secure_storage_service.dart';

class TokenRefreshService {
  TokenRefreshService({
    required SecureStorageService storage,
    Dio? dio,
  })  : _storage = storage,
        _dio = dio ?? Dio(BaseOptions(baseUrl: AppConfig.apiBaseUrl));

  final SecureStorageService _storage;
  final Dio _dio;

  Future<void> refreshAccessToken() async {
    final refreshToken = await _storage.readRefreshToken();
    if (refreshToken == null || refreshToken.isEmpty) {
      throw Exception('Refresh token no disponible');
    }

    final response = await _dio.post(
      '/auth/refresh',
      data: {'refreshToken': refreshToken},
    );

    final accessToken = response.data['accessToken'] as String?;
    final newRefreshToken = response.data['refreshToken'] as String?;

    if (accessToken == null || newRefreshToken == null) {
      throw Exception('La renovación de sesión falló');
    }

    await _storage.saveAccessToken(accessToken);
    await _storage.saveRefreshToken(newRefreshToken);
  }
}
