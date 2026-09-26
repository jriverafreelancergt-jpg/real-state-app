import 'package:dartz/dartz.dart';
import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:real_state_mobile/core/error/failures.dart';
import 'package:real_state_mobile/core/providers/app_providers.dart';
import 'package:real_state_mobile/core/session/auth_session_manager.dart';
import 'package:real_state_mobile/core/storage/secure_storage_service.dart';
import 'package:real_state_mobile/features/auth/domain/entities/auth_session.dart';
import 'package:real_state_mobile/features/auth/domain/repositories/auth_repository.dart';
import 'package:real_state_mobile/features/auth/domain/usecases/login_usecase.dart';
import 'package:real_state_mobile/features/auth/presentation/providers/auth_provider.dart';
import 'package:real_state_mobile/features/auth/presentation/screens/login_screen.dart';
import 'package:real_state_mobile/features/property/domain/entities/property.dart';
import 'package:real_state_mobile/features/property/domain/entities/property_media.dart';
import 'package:real_state_mobile/features/property/domain/repositories/property_repository.dart';
import 'package:real_state_mobile/features/property/domain/usecases/get_properties_usecase.dart';
import 'package:real_state_mobile/features/property/presentation/providers/property_provider.dart';
import 'package:real_state_mobile/features/property/presentation/screens/property_list_screen.dart';

class FakeAuthRepository implements AuthRepository {
  FakeAuthRepository();

  int loginCalls = 0;

  @override
  Future<Either<Failure, AuthSession>> login({
    required String email,
    required String password,
  }) async {
    loginCalls++;

    return Right(
      AuthSession(
        accessToken: 'fake_access_token',
        refreshToken: 'fake_refresh_token',
        userEmail: email,
        expiresIn: 900,
      ),
    );
  }

  @override
  Future<Either<Failure, AuthSession>> refreshToken({
    required String refreshToken,
  }) async {
    return const Right(
      AuthSession(
        accessToken: 'fresh_access_token',
        refreshToken: 'fresh_refresh_token',
        userEmail: 'demo@realtyastarac.com',
        expiresIn: 900,
      ),
    );
  }

  @override
  Future<void> logout() async {}
}

class FakePropertyRepository implements PropertyRepository {
  @override
  Future<Either<Failure, List<Property>>> getProperties({
    int page = 1,
    int limit = 20,
    String? operation,
    double? minPrice,
    double? maxPrice,
  }) async {
    return Right([
      const Property(
        id: 'prop-1',
        title: 'Casa de prueba',
        price: 125000,
        operation: OperationType.sale,
        status: PropertyStatus.available,
        description: 'Propiedad cargada desde el fake repository',
        propertyType: 'HOUSE',
        address: 'Guatemala City',
      ),
    ]);
  }

  @override
  Future<Either<Failure, Property>> getPropertyById(String id) async {
    return Right(
      const Property(
        id: 'prop-1',
        title: 'Casa de prueba',
        price: 125000,
        operation: OperationType.sale,
        status: PropertyStatus.available,
        description: 'Detalle de prueba',
        propertyType: 'HOUSE',
        address: 'Guatemala City',
      ),
    );
  }

  @override
  Future<Either<Failure, Property>> createProperty(Property property) async {
    return Right(property);
  }

  @override
  Future<Either<Failure, Property>> updateProperty(String id, Property property) async {
    return Right(property);
  }

  @override
  Future<Either<Failure, bool>> deleteProperty(String id) async {
    return const Right(true);
  }

  @override
  Future<Either<Failure, PropertyMedia>> uploadMedia({
    required String propertyId,
    required String url,
    required String type,
    bool isPrimary = false,
  }) async {
    throw UnimplementedError();
  }
}

class _FakeAuthSessionManager extends AuthSessionManager {
  _FakeAuthSessionManager()
      : super(
          storage: InMemorySecureStorage(),
        );

  AuthSession? currentSession;

  @override
  Future<AuthSession?> restoreSession() async => currentSession;

  @override
  Future<void> persistSession(AuthSession session) async {
    currentSession = session;
  }

  @override
  Future<void> clearSession() async {
    currentSession = null;
  }
}

class TestAuthNotifier extends AuthNotifier {
  TestAuthNotifier() : super(_FakeAuthSessionManager());
}

class InMemorySecureStorage implements SecureStorageService {
  final Map<String, String> _storage = {};

  @override
  Future<void> saveAccessToken(String token) async => _storage['access_token'] = token;

  @override
  Future<void> saveRefreshToken(String token) async => _storage['refresh_token'] = token;

  @override
  Future<void> saveUserEmail(String email) async => _storage['user_email'] = email;

  @override
  Future<String?> readAccessToken() async => _storage['access_token'];

  @override
  Future<String?> readRefreshToken() async => _storage['refresh_token'];

  @override
  Future<String?> readUserEmail() async => _storage['user_email'];

  @override
  Future<void> clearSession() async => _storage.clear();
}

void main() {
  testWidgets('login flow stores a session and property list shows data', (tester) async {
    final authRepository = FakeAuthRepository();
    final propertyRepository = FakePropertyRepository();

    final container = ProviderContainer(
      overrides: [
        loginUseCaseProvider.overrideWith(
          (ref) => LoginUseCase(authRepository),
        ),
        authProvider.overrideWith(
          (ref) => TestAuthNotifier(),
        ),
        getPropertiesUseCaseProvider.overrideWith(
          (ref) => GetPropertiesUseCase(propertyRepository),
        ),
        propertyProvider.overrideWith(
          (ref) => PropertyNotifier(GetPropertiesUseCase(propertyRepository)),
        ),
      ],
    );
    addTearDown(container.dispose);

    await tester.pumpWidget(
      UncontrolledProviderScope(
        container: container,
        child: const MaterialApp(home: LoginScreen()),
      ),
    );

    await tester.enterText(find.byType(TextField).at(0), 'admin@realtyastarac.com');
    await tester.enterText(find.byType(TextField).at(1), 'secret123');
    await tester.tap(find.text('Entrar'));
    await tester.pumpAndSettle();

    expect(authRepository.loginCalls, 1);
    expect(container.read(authProvider).valueOrNull?.userEmail,
        'admin@realtyastarac.com');

    await tester.pumpWidget(
      UncontrolledProviderScope(
        container: container,
        child: const MaterialApp(home: PropertyListScreen()),
      ),
    );
    await tester.pumpAndSettle();

    expect(find.text('Casa de prueba'), findsOneWidget);
  });
}
