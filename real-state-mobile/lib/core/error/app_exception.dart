abstract class AppException implements Exception {
  const AppException(this.message);

  final String message;
}

class NetworkException extends AppException {
  const NetworkException(super.message);
}

class ServerException extends AppException {
  const ServerException(super.message);
}

class UnauthorizedException extends AppException {
  const UnauthorizedException(super.message);
}

class InvalidResponseException extends AppException {
  const InvalidResponseException(super.message);
}
