import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:real_state_mobile/core/providers/app_providers.dart';
import 'package:real_state_mobile/features/property/domain/entities/property.dart';
import 'package:real_state_mobile/features/property/domain/usecases/get_property_by_id_usecase.dart';

final getPropertyByIdUseCaseProvider = Provider<GetPropertyByIdUseCase>((ref) {
  return GetPropertyByIdUseCase(ref.watch(propertyRepositoryProvider));
});

final propertyDetailProvider = FutureProvider.family<Property, String>((ref, id) async {
  final useCase = ref.watch(getPropertyByIdUseCaseProvider);
  final result = await useCase(id);

  return result.fold(
    (failure) => throw Exception(failure.message),
    (property) => property,
  );
});
