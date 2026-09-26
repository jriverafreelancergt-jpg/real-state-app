import 'package:dio/dio.dart';
import '../dto/property_dto.dart';
import '../dto/property_list_response_dto.dart';
import '../dto/property_media_dto.dart';

abstract class PropertyRemoteDataSource {
  Future<List<PropertyDto>> getProperties({
    required int page,
    required int limit,
    String? operation,
    double? minPrice,
    double? maxPrice,
  });

  Future<PropertyDto> getPropertyById(String id);

  Future<PropertyDto> createProperty(PropertyDto property);

  Future<PropertyDto> updateProperty(String id, PropertyDto property);

  Future<bool> deleteProperty(String id);

  Future<PropertyMediaDto> uploadMedia({
    required String propertyId,
    required String url,
    required String type,
    bool isPrimary = false,
  });
}

class PropertyRemoteDataSourceImpl implements PropertyRemoteDataSource {
  const PropertyRemoteDataSourceImpl(this._dio);

  final Dio _dio;

  @override
  Future<List<PropertyDto>> getProperties({
    required int page,
    required int limit,
    String? operation,
    double? minPrice,
    double? maxPrice,
  }) async {
    final response = await _dio.get(
      '/properties',
      queryParameters: {
        'page': page,
        'limit': limit,
        if (operation != null) 'operation': operation,
        if (minPrice != null) 'min_price': minPrice,
        if (maxPrice != null) 'max_price': maxPrice,
      },
    );

    final payload = response.data is Map<String, dynamic>
        ? response.data as Map<String, dynamic>
        : {'items': <Map<String, dynamic>>[]};

    final dto = PropertyListResponseDto.fromJson(payload);
    return dto.items;
  }

  @override
  Future<PropertyDto> getPropertyById(String id) async {
    final response = await _dio.get('/properties/$id');
    final payload = response.data is Map<String, dynamic>
        ? response.data as Map<String, dynamic>
        : <String, dynamic>{};
    return PropertyDto.fromJson(payload);
  }

  @override
  Future<PropertyDto> createProperty(PropertyDto property) async {
    final response = await _dio.post(
      '/properties',
      data: property.toJson(),
    );

    final payload = response.data is Map<String, dynamic>
        ? response.data as Map<String, dynamic>
        : <String, dynamic>{};

    return PropertyDto.fromJson(payload);
  }

  @override
  Future<PropertyDto> updateProperty(String id, PropertyDto property) async {
    final response = await _dio.put(
      '/properties/$id',
      data: property.toJson(),
    );

    final payload = response.data is Map<String, dynamic>
        ? response.data as Map<String, dynamic>
        : <String, dynamic>{};

    return PropertyDto.fromJson(payload);
  }

  @override
  Future<bool> deleteProperty(String id) async {
    final response = await _dio.delete('/properties/$id');
    return response.statusCode != null && response.statusCode! >= 200 && response.statusCode! < 300;
  }

  @override
  Future<PropertyMediaDto> uploadMedia({
    required String propertyId,
    required String url,
    required String type,
    bool isPrimary = false,
  }) async {
    final response = await _dio.post(
      '/properties/$propertyId/media',
      data: {
        'url': url,
        'type': type,
        'is_primary': isPrimary,
      },
    );

    final payload = response.data is Map<String, dynamic>
        ? response.data as Map<String, dynamic>
        : <String, dynamic>{};

    return PropertyMediaDto.fromJson(payload);
  }
}
