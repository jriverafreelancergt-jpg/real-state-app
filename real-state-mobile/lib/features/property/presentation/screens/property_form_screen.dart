import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:real_state_mobile/features/auth/presentation/providers/auth_provider.dart';
import 'package:real_state_mobile/features/property/domain/entities/property.dart';
import 'package:real_state_mobile/features/property/presentation/providers/property_form_provider.dart';

class PropertyFormScreen extends ConsumerStatefulWidget {
  const PropertyFormScreen({
    super.key,
    this.property,
  });

  final Property? property;

  @override
  ConsumerState<PropertyFormScreen> createState() => _PropertyFormScreenState();
}

class _PropertyFormScreenState extends ConsumerState<PropertyFormScreen> {
  final _formKey = GlobalKey<FormState>();
  late final TextEditingController _titleController;
  late final TextEditingController _priceController;
  late final TextEditingController _descriptionController;

  OperationType _operation = OperationType.sale;
  PropertyStatus _status = PropertyStatus.available;

  @override
  void initState() {
    super.initState();
    final property = widget.property;
    _titleController = TextEditingController(text: property?.title ?? '');
    _priceController = TextEditingController(
      text: property != null ? property.price.toStringAsFixed(0) : '0',
    );
    _descriptionController = TextEditingController(text: property?.description ?? '');

    if (property != null) {
      _operation = property.operation;
      _status = property.status;
    }
  }

  @override
  void dispose() {
    _titleController.dispose();
    _priceController.dispose();
    _descriptionController.dispose();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    final formState = ref.watch(propertyFormProvider);
    final session = ref.watch(authProvider).valueOrNull;
    final canSubmit = !formState.isLoading && session != null;

    return Scaffold(
      appBar: AppBar(
        title: Text(widget.property == null ? 'Nueva propiedad' : 'Editar propiedad'),
      ),
      body: Padding(
        padding: const EdgeInsets.all(16),
        child: Form(
          key: _formKey,
          child: ListView(
            children: [
              if (session == null)
                const Text('Debes iniciar sesión para administrar propiedades.'),
              TextFormField(
                controller: _titleController,
                decoration: const InputDecoration(labelText: 'Título'),
                validator: (value) => value == null || value.isEmpty ? 'Obligatorio' : null,
              ),
              const SizedBox(height: 12),
              TextFormField(
                controller: _priceController,
                keyboardType: const TextInputType.numberWithOptions(decimal: true),
                decoration: const InputDecoration(labelText: 'Precio'),
                validator: (value) {
                  if (value == null || value.trim().isEmpty) {
                    return 'Obligatorio';
                  }
                  if (double.tryParse(value.trim()) == null) {
                    return 'Precio inválido';
                  }
                  return null;
                },
              ),
              const SizedBox(height: 12),
              TextFormField(
                controller: _descriptionController,
                maxLines: 4,
                decoration: const InputDecoration(labelText: 'Descripción'),
                validator: (value) => value == null || value.trim().isEmpty ? 'Obligatorio' : null,
              ),
              const SizedBox(height: 12),
              DropdownButtonFormField<OperationType>(
                value: _operation,
                items: OperationType.values
                    .map(
                      (operation) => DropdownMenuItem(
                        value: operation,
                        child: Text(operation.name),
                      ),
                    )
                    .toList(),
                onChanged: (value) {
                  if (value != null) {
                    setState(() => _operation = value);
                  }
                },
              ),
              const SizedBox(height: 12),
              DropdownButtonFormField<PropertyStatus>(
                value: _status,
                items: PropertyStatus.values
                    .map(
                      (status) => DropdownMenuItem(
                        value: status,
                        child: Text(status.name),
                      ),
                    )
                    .toList(),
                onChanged: (value) {
                  if (value != null) {
                    setState(() => _status = value);
                  }
                },
              ),
              const SizedBox(height: 24),
              FilledButton(
                onPressed: canSubmit
                    ? () async {
                        if (!_formKey.currentState!.validate()) {
                          return;
                        }

                        final property = Property(
                          id: widget.property?.id ?? '',
                          title: _titleController.text.trim(),
                          description: _descriptionController.text.trim(),
                          price: double.tryParse(_priceController.text.trim()) ?? 0,
                          operation: _operation,
                          status: _status,
                          currency: widget.property?.currency ?? 'GTQ',
                          propertyType: widget.property?.propertyType ?? 'HOUSE',
                          address: widget.property?.address ?? '',
                        );

                        if (widget.property == null) {
                          await ref.read(propertyFormProvider.notifier).create(property);
                        } else {
                          await ref.read(propertyFormProvider.notifier).update(
                            id: widget.property!.id,
                            property: property,
                          );
                        }

                        if (context.mounted) {
                          Navigator.of(context).pop();
                        }
                      }
                    : null,
                child: formState.isLoading
                    ? const SizedBox(
                        width: 18,
                        height: 18,
                        child: CircularProgressIndicator(strokeWidth: 2),
                      )
                    : Text(widget.property == null ? 'Guardar propiedad' : 'Actualizar propiedad'),
              ),
            ],
          ),
        ),
      ),
    );
  }
}
