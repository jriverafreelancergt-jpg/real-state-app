import 'package:flutter/material.dart';
import 'package:real_state_mobile/features/property/domain/entities/property.dart';

class PropertyCard extends StatelessWidget {
  const PropertyCard({
    super.key,
    required this.property,
  });

  final Property property;

  @override
  Widget build(BuildContext context) {
    return Card(
      margin: const EdgeInsets.symmetric(vertical: 8, horizontal: 12),
      child: Padding(
        padding: const EdgeInsets.all(16),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            Text(
              property.title,
              style: Theme.of(context).textTheme.titleMedium,
            ),
            const SizedBox(height: 8),
            Text(
              'USD \${property.price.toStringAsFixed(0)}',
            ),
            const SizedBox(height: 8),
            Text(
              'Operación: ${property.operation.name}',
              style: Theme.of(context).textTheme.bodySmall,
            ),
            Text(
              'Estado: ${property.status.name}',
              style: Theme.of(context).textTheme.bodySmall,
            ),
          ],
        ),
      ),
    );
  }
}
