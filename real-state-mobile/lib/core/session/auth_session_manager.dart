import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:real_state_mobile/core/storage/secure_storage_service.dart';
import 'package:real_state_mobile/features/auth/domain/entities/auth_session.dart';

class AuthSessionManager {
  AuthSessionManager({
    required SecureStorageService storage,
  }) : _storage = storage;

  final SecureStorageService _storage;

  Future<AuthSession?> restoreSession() async {
    final accessToken = await _storage.readAccessToken();
    final refreshToken = await _storage.readRefreshToken();
    final email = await _storage.readUserEmail();

    if (accessToken == null || refreshToken == null || email == null) {
      return null;
    }

    return AuthSession(
      accessToken: accessToken,
      refreshToken: refreshToken,
      userEmail: email,
      expiresIn: 0,
    );
  }

  Future<void> persistSession(AuthSession session) async {
    await _storage.saveAccessToken(session.accessToken);
    await _storage.saveRefreshToken(session.refreshToken);
    await _storage.saveUserEmail(session.userEmail);
  }

  Future<void> clearSession() async {
    await _storage.clearSession();
  }
}

final authSessionManagerProvider = Provider<AuthSessionManager>((ref) {
  return AuthSessionManager(
    storage: SecureStorageService(),
  );
});
