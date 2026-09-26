import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:real_state_mobile/core/config/app_config.dart';
import 'package:real_state_mobile/core/network/dio_client.dart';
import 'package:real_state_mobile/core/storage/secure_storage_service.dart';
import 'package:real_state_mobile/features/auth/data/datasources/auth_remote_data_source.dart';
import 'package:real_state_mobile/features/auth/data/repositories/auth_repository_impl.dart';
import 'package:real_state_mobile/features/auth/domain/repositories/auth_repository.dart';
import 'package:real_state_mobile/features/auth/domain/usecases/login_usecase.dart';
import 'package:real_state_mobile/features/property/data/datasources/property_remote_data_source.dart';
import 'package:real_state_mobile/features/property/data/repositories/property_repository_impl.dart';
import 'package:real_state_mobile/features/property/domain/repositories/property_repository.dart';
import 'package:real_state_mobile/features/property/domain/usecases/delete_property_usecase.dart';
import 'package:real_state_mobile/features/property/domain/usecases/get_properties_usecase.dart';
import 'package:real_state_mobile/features/property/domain/usecases/upload_property_media_usecase.dart';

final secureStorageProvider = Provider<SecureStorageService>((ref) {
  return SecureStorageService();
});

final dioClientProvider = Provider<DioClient>((ref) {
  final storage = ref.watch(secureStorageProvider);
  return DioClient(
    baseUrl: AppConfig.apiBaseUrl,
    storage: storage,
  );
});

final authRemoteDataSourceProvider = Provider<AuthRemoteDataSource>((ref) {
  return AuthRemoteDataSourceImpl(ref.watch(dioClientProvider).dio);
});

final authRepositoryProvider = Provider<AuthRepository>((ref) {
  return AuthRepositoryImpl(
    remoteDataSource: ref.watch(authRemoteDataSourceProvider),
    storage: ref.watch(secureStorageProvider),
  );
});

final loginUseCaseProvider = Provider<LoginUseCase>((ref) {
  return LoginUseCase(ref.watch(authRepositoryProvider));
});

final propertyRemoteDataSourceProvider = Provider<PropertyRemoteDataSource>((ref) {
  return PropertyRemoteDataSourceImpl(ref.watch(dioClientProvider).dio);
});

final propertyRepositoryProvider = Provider<PropertyRepository>((ref) {
  return PropertyRepositoryImpl(ref.watch(propertyRemoteDataSourceProvider));
});

final getPropertiesUseCaseProvider = Provider<GetPropertiesUseCase>((ref) {
  return GetPropertiesUseCase(ref.watch(propertyRepositoryProvider));
});

final deletePropertyUseCaseProvider = Provider<DeletePropertyUseCase>((ref) {
  return DeletePropertyUseCase(ref.watch(propertyRepositoryProvider));
});

final uploadPropertyMediaUseCaseProvider = Provider<UploadPropertyMediaUseCase>((ref) {
  return UploadPropertyMediaUseCase(ref.watch(propertyRepositoryProvider));
});
