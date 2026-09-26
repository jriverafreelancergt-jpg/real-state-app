import 'package:dio/dio.dart';

class ApiClient {
  ApiClient() : _dio = Dio(BaseOptions(
          baseUrl: 'https://api.realtyastarac.com/api/v1',
          connectTimeout: const Duration(seconds: 10),
          receiveTimeout: const Duration(seconds: 10),
          headers: {
            'Content-Type': 'application/json',
            'Accept': 'application/json',
          },
        ));

  final Dio _dio;

  Dio get dio => _dio;
}
