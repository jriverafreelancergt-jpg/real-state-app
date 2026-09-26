import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:real_state_mobile/core/router/app_router.dart';
import 'package:real_state_mobile/core/theme/app_theme.dart';
import 'package:real_state_mobile/features/auth/presentation/providers/auth_provider.dart';

class RealtyAstaracApp extends ConsumerStatefulWidget {
  const RealtyAstaracApp({super.key});

  @override
  ConsumerState<RealtyAstaracApp> createState() => _RealtyAstaracAppState();
}

class _RealtyAstaracAppState extends ConsumerState<RealtyAstaracApp> {
  @override
  void initState() {
    super.initState();
    Future.microtask(() {
      ref.read(authProvider.notifier).restoreSession();
    });
  }

  @override
  Widget build(BuildContext context) {
    return MaterialApp.router(
      title: 'Realty Astarac',
      debugShowCheckedModeBanner: false,
      theme: AppTheme.light(),
      routerConfig: AppRouter.router,
    );
  }
}
