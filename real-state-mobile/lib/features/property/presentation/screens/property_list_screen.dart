import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';
import 'package:real_state_mobile/features/property/presentation/providers/property_provider.dart';
import 'package:real_state_mobile/features/property/presentation/widgets/property_card.dart';

class PropertyListScreen extends ConsumerStatefulWidget {
  const PropertyListScreen({super.key});

  @override
  ConsumerState<PropertyListScreen> createState() => _PropertyListScreenState();
}

class _PropertyListScreenState extends ConsumerState<PropertyListScreen> {
  @override
  void initState() {
    super.initState();
    Future.microtask(() {
      ref.read(propertyProvider.notifier).loadProperties();
    });
  }

  @override
  Widget build(BuildContext context) {
    final state = ref.watch(propertyProvider);

    return Scaffold(
      appBar: AppBar(
        title: const Text('Propiedades'),
      ),
      body: state.when(
        data: (properties) {
          if (properties.isEmpty) {
            return const Center(child: Text('No hay propiedades disponibles'));
          }

          return ListView.builder(
            itemCount: properties.length,
            itemBuilder: (context, index) {
              final property = properties[index];
              return GestureDetector(
                onTap: () => context.push('/property/${property.id}'),
                child: PropertyCard(property: property),
              );
            },
          );
        },
        loading: () => const Center(child: CircularProgressIndicator()),
        error: (error, stackTrace) => Center(
          child: Column(
            mainAxisAlignment: MainAxisAlignment.center,
            children: [
              const Icon(Icons.error_outline, size: 44),
              const SizedBox(height: 12),
              Text('No se pudo cargar el catálogo: $error'),
            ],
          ),
        ),
      ),
    );
  }
}
