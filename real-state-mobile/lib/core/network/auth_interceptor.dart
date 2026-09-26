import 'package:dio/dio.dart';
import 'package:real_state_mobile/core/network/token_refresh_service.dart';
import 'package:real_state_mobile/core/storage/secure_storage_service.dart';

class AuthInterceptor extends Interceptor {
  AuthInterceptor({
    required SecureStorageService storage,
    required Dio dio,
    required TokenRefreshService refreshService,
  })  : _storage = storage,
        _dio = dio,
        _refreshService = refreshService;

  final SecureStorageService _storage;
  final Dio _dio;
  final TokenRefreshService _refreshService;
  bool _isRefreshing = false;

  @override
  Future<void> onRequest(
    RequestOptions options,
    RequestInterceptorHandler handler,
  ) async {
    final token = await _storage.readAccessToken();
    if (token != null && token.isNotEmpty) {
      options.headers['Authorization'] = 'Bearer $token';
    }
    return handler.next(options);
  }

  @override
  Future<void> onError(
    DioException err,
    ErrorInterceptorHandler handler,
  ) async {
    if (err.response?.statusCode == 401 && !_isRefreshing) {
      _isRefreshing = true;
      try {
        await _refreshService.refreshAccessToken();

        final requestOptions = err.requestOptions;
        final token = await _storage.readAccessToken();
        if (token != null && token.isNotEmpty) {
          requestOptions.headers['Authorization'] = 'Bearer $token';
        }

        final response = await _dio.fetch(requestOptions);
        _isRefreshing = false;
        return handler.resolve(response);
      } catch (_) {
        _isRefreshing = false;
        return handler.next(err);
      }
    }

    return handler.next(err);
  }
}
