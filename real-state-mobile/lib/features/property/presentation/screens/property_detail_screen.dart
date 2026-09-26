import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:real_state_mobile/features/property/presentation/providers/property_detail_provider.dart';

class PropertyDetailScreen extends ConsumerWidget {
  const PropertyDetailScreen({
    super.key,
    required this.propertyId,
  });

  final String propertyId;

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final propertyAsync = ref.watch(propertyDetailProvider(propertyId));

    return Scaffold(
      appBar: AppBar(title: const Text('Detalle')),
      body: propertyAsync.when(
        data: (property) => Padding(
          padding: const EdgeInsets.all(20),
          child: Column(
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              Text(
                property.title,
                style: Theme.of(context).textTheme.headlineSmall,
              ),
              const SizedBox(height: 12),
              Text('Precio: \$${property.price.toStringAsFixed(0)}'),
              const SizedBox(height: 8),
              Text('Operación: ${property.operation.name}'),
              const SizedBox(height: 8),
              Text('Estado: ${property.status.name}'),
            ],
          ),
        ),
        loading: () => const Center(child: CircularProgressIndicator()),
        error: (error, stackTrace) => Center(
          child: Text('Error: $error'),
        ),
      ),
    );
  }
}
