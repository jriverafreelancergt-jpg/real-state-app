import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:real_state_mobile/core/session/auth_session_manager.dart';
import 'package:real_state_mobile/features/auth/domain/entities/auth_session.dart';
import 'package:real_state_mobile/features/auth/domain/usecases/login_usecase.dart';

final authProvider = StateNotifierProvider<AuthNotifier, AsyncValue<AuthSession?>>((ref) {
  final manager = ref.watch(authSessionManagerProvider);
  return AuthNotifier(manager);
});

class AuthNotifier extends StateNotifier<AsyncValue<AuthSession?>> {
  AuthNotifier(this._manager) : super(const AsyncValue.data(null));

  final AuthSessionManager _manager;

  Future<void> restoreSession() async {
    final session = await _manager.restoreSession();
    state = AsyncValue.data(session);
  }

  Future<void> login(
    LoginUseCase useCase, {
    required String email,
    required String password,
  }) async {
    state = const AsyncValue.loading();

    final result = await useCase(email: email, password: password);

    result.fold(
      (failure) {
        state = AsyncValue.error(failure, StackTrace.current);
      },
      (session) async {
        await _manager.persistSession(session);
        state = AsyncValue.data(session);
      },
    );
  }

  Future<void> logout() async {
    await _manager.clearSession();
    state = const AsyncValue.data(null);
  }
}
