import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:real_state_mobile/core/providers/app_providers.dart';
import 'package:real_state_mobile/features/property/domain/entities/property.dart';
import 'package:real_state_mobile/features/property/domain/usecases/create_property_usecase.dart';
import 'package:real_state_mobile/features/property/domain/usecases/update_property_usecase.dart';

final createPropertyUseCaseProvider = Provider<CreatePropertyUseCase>((ref) {
  return CreatePropertyUseCase(ref.watch(propertyRepositoryProvider));
});

final updatePropertyUseCaseProvider = Provider<UpdatePropertyUseCase>((ref) {
  return UpdatePropertyUseCase(ref.watch(propertyRepositoryProvider));
});

final propertyFormProvider = StateNotifierProvider<PropertyFormNotifier, AsyncValue<void>>((ref) {
  return PropertyFormNotifier(
    createUseCase: ref.watch(createPropertyUseCaseProvider),
    updateUseCase: ref.watch(updatePropertyUseCaseProvider),
  );
});

class PropertyFormNotifier extends StateNotifier<AsyncValue<void>> {
  PropertyFormNotifier({
    required CreatePropertyUseCase createUseCase,
    required UpdatePropertyUseCase updateUseCase,
  })  : _createUseCase = createUseCase,
        _updateUseCase = updateUseCase,
        super(const AsyncValue.data(null));

  final CreatePropertyUseCase _createUseCase;
  final UpdatePropertyUseCase _updateUseCase;

  Future<void> create(Property property) async {
    state = const AsyncValue.loading();
    final result = await _createUseCase(property);

    result.fold(
      (failure) => state = AsyncValue.error(failure, StackTrace.current),
      (_) => state = const AsyncValue.data(null),
    );
  }

  Future<void> update({
    required String id,
    required Property property,
  }) async {
    state = const AsyncValue.loading();
    final result = await _updateUseCase(id: id, property: property);

    result.fold(
      (failure) => state = AsyncValue.error(failure, StackTrace.current),
      (_) => state = const AsyncValue.data(null),
    );
  }
}
