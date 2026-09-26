class CurrencyFormatter {
  static String format(double value) {
    return '\u00A0${value.toStringAsFixed(0)}';
  }
}
