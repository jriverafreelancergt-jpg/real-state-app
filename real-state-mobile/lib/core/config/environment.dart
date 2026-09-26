enum AppEnvironment { development, staging, production }

class Environment {
  const Environment._({
    required this.name,
    required this.apiBaseUrl,
  });

  final AppEnvironment name;
  final String apiBaseUrl;

  static const Environment development = Environment._(
    name: AppEnvironment.development,
    apiBaseUrl: 'http://192.168.0.19:8080',
  );

  static const Environment staging = Environment._(
    name: AppEnvironment.staging,
    apiBaseUrl: 'https://stg-api.realtyastarac.com/api/v1',
  );

  static const Environment production = Environment._(
    name: AppEnvironment.production,
    apiBaseUrl: 'https://api.realtyastarac.com/api/v1',
  );
}
