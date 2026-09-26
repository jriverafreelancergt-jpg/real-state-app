import 'package:dio/dio.dart';
import 'package:real_state_mobile/core/network/auth_interceptor.dart';
import 'package:real_state_mobile/core/network/token_refresh_service.dart';
import 'package:real_state_mobile/core/storage/secure_storage_service.dart';

class DioClient {
  DioClient({
    required String baseUrl,
    SecureStorageService? storage,
  }) : _dio = Dio(BaseOptions(
          baseUrl: baseUrl,
          connectTimeout: const Duration(seconds: 10),
          receiveTimeout: const Duration(seconds: 10),
          sendTimeout: const Duration(seconds: 10),
          headers: {
            'Content-Type': 'application/json',
            'Accept': 'application/json',
          },
        )) {
    final resolvedStorage = storage ?? SecureStorageService();
    final refreshService = TokenRefreshService(storage: resolvedStorage, dio: _dio);

    _dio.interceptors.add(
      AuthInterceptor(
        storage: resolvedStorage,
        dio: _dio,
        refreshService: refreshService,
      ),
    );
  }

  final Dio _dio;

  Dio get dio => _dio;
}
