import 'package:flutter_secure_storage/flutter_secure_storage.dart';

class SecureStorageService {
  SecureStorageService({
    FlutterSecureStorage? storage,
  }) : _storage = storage ?? const FlutterSecureStorage();

  final FlutterSecureStorage _storage;

  static const String accessTokenKey = 'access_token';
  static const String refreshTokenKey = 'refresh_token';
  static const String userEmailKey = 'user_email';

  Future<void> saveAccessToken(String token) async {
    await _storage.write(key: accessTokenKey, value: token);
  }

  Future<void> saveRefreshToken(String token) async {
    await _storage.write(key: refreshTokenKey, value: token);
  }

  Future<void> saveUserEmail(String email) async {
    await _storage.write(key: userEmailKey, value: email);
  }

  Future<String?> readAccessToken() async {
    return _storage.read(key: accessTokenKey);
  }

  Future<String?> readRefreshToken() async {
    return _storage.read(key: refreshTokenKey);
  }

  Future<String?> readUserEmail() async {
    return _storage.read(key: userEmailKey);
  }

  Future<void> clearSession() async {
    await Future.wait([
      _storage.delete(key: accessTokenKey),
      _storage.delete(key: refreshTokenKey),
      _storage.delete(key: userEmailKey),
    ]);
  }
}
