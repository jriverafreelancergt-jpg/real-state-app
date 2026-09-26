import 'property_dto.dart';

class PropertyListResponseDto {
  const PropertyListResponseDto({
    required this.items,
    required this.page,
    required this.limit,
    required this.total,
  });

  factory PropertyListResponseDto.fromJson(Map<String, dynamic> json) {
    final items = ((json['items'] as List?) ?? const [])
        .map((item) => PropertyDto.fromJson(item as Map<String, dynamic>))
        .toList();

    return PropertyListResponseDto(
      items: items,
      page: json['page'] as int? ?? 1,
      limit: json['limit'] as int? ?? items.length,
      total: json['total'] as int? ?? items.length,
    );
  }

  final List<PropertyDto> items;
  final int page;
  final int limit;
  final int total;
}
