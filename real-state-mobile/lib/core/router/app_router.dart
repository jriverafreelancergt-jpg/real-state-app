import 'package:go_router/go_router.dart';
import 'package:real_state_mobile/core/session/auth_session_manager.dart';
import 'package:real_state_mobile/core/storage/secure_storage_service.dart';
import 'package:real_state_mobile/features/auth/presentation/screens/login_screen.dart';
import 'package:real_state_mobile/features/property/presentation/screens/property_detail_screen.dart';
import 'package:real_state_mobile/features/property/presentation/screens/property_list_screen.dart';

abstract final class AppRouter {
  static final AuthSessionManager _sessionManager = AuthSessionManager(
    storage: SecureStorageService(),
  );

  static Future<String?> redirectHandler(_, GoRouterState state) async {
    final session = await _sessionManager.restoreSession();
    final isAuthRoute = state.matchedLocation == '/login';

    if (session == null) {
      return isAuthRoute ? null : '/login';
    }

    return isAuthRoute ? '/' : null;
  }

  static final router = GoRouter(
    initialLocation: '/',
    redirect: redirectHandler,
    routes: [
      GoRoute(
        path: '/',
        builder: (context, state) => const PropertyListScreen(),
      ),
      GoRoute(
        path: '/login',
        builder: (context, state) => const LoginScreen(),
      ),
      GoRoute(
        path: '/property/:propertyId',
        builder: (context, state) {
          final propertyId = state.pathParameters['propertyId'] ?? '';
          return PropertyDetailScreen(propertyId: propertyId);
        },
      ),
    ],
  );
}
