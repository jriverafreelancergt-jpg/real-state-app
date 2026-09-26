import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:real_state_mobile/core/providers/app_providers.dart';
import 'package:real_state_mobile/features/property/domain/entities/property.dart';
import 'package:real_state_mobile/features/property/domain/usecases/get_properties_usecase.dart';

final propertyProvider = StateNotifierProvider<PropertyNotifier, AsyncValue<List<Property>>>((ref) {
  final useCase = ref.watch(getPropertiesUseCaseProvider);
  return PropertyNotifier(useCase);
});

class PropertyNotifier extends StateNotifier<AsyncValue<List<Property>>> {
  PropertyNotifier(this._useCase) : super(const AsyncValue.data([]));

  final GetPropertiesUseCase _useCase;

  Future<void> loadProperties() async {
    state = const AsyncValue.loading();

    final result = await _useCase();

    result.fold(
      (failure) => state = AsyncValue.error(failure, StackTrace.current),
      (properties) => state = AsyncValue.data(properties),
    );
  }
}
