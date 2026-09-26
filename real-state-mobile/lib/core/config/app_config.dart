import 'environment.dart';

class AppConfig {
  const AppConfig._();

  static AppEnvironment environment = AppEnvironment.development;

  static String get apiBaseUrl => switch (environment) {
        AppEnvironment.development => Environment.development.apiBaseUrl,
        AppEnvironment.staging => Environment.staging.apiBaseUrl,
        AppEnvironment.production => Environment.production.apiBaseUrl,
      };
}
