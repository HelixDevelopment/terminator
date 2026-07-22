// Host-rendered visual proof for HelixTerminator's primary screens (§11.4.170).
//
// Flutter golden tests render the REAL screen widget to device-independent
// PNG pixels ON THE HOST — no device, emulator, or running app — for every
// screen x {light, dark} theme in this suite. This closes the gap flagged
// (honestly, not silently) in test/README.md: previously ONLY the single
// TerminalView widget (test/terminal_view_golden_test.dart) had host-rendered
// golden coverage; no full *screen* did.
//
// Screens covered in this pass (entry/auth/primary, per priority):
//   - SplashScreen            (brand entrance, real AnimationController)
//   - LoginScreen             (real Form, real AuthBloc wiring)
//   - DashboardScreen         ("Home" tab — real HostBloc-backed content)
//   - WorkspaceListScreen     ("Workspace" — real WorkspaceBloc-backed content)
//   - TerminalScreen          ("Terminal" — real xterm engine + real TerminalBloc)
//
// Dual validation per §11.4.170, for every test below:
//   (i)  golden image-diff — the committed PNG is asserted pixel-for-pixel
//        via matchesGoldenFile (device-independent host-rendered pixels).
//   (ii) content oracle — finder/semantics-based rendered-TEXT assertions
//        (find.text / find.widgetWithText / find.byIcon) proving every
//        expected label + control is genuinely on screen, not blank,
//        collapsed, or silently dropped. This is a Flutter-finder text
//        oracle, NOT a pixel-level OCR pass — see the honest gap noted in
//        the "Dual-validation, honestly" section of test/README.md.
//   Each test ALSO asserts tester.takeException() is null after settling,
//   so a RenderFlex overflow (a real regression class on this app — see
//   test/README.md defect #2 on LoginScreen) fails the test even if the
//   golden pixels alone would not make the mismatch obvious.
//
// Regenerate goldens (inside the rootless Podman Flutter container):
//   flutter test --update-goldens test/screens_golden_test.dart
// Verify (must pass without --update-goldens):
//   flutter test test/screens_golden_test.dart

import 'package:design_system/design_system.dart';
import 'package:flutter/material.dart';
import 'package:flutter_bloc/flutter_bloc.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:mocktail/mocktail.dart';
import 'package:xterm/xterm.dart' as xterm;

import 'package:helix_terminator/bloc/auth_bloc.dart';
import 'package:helix_terminator/bloc/host_bloc.dart';
import 'package:helix_terminator/bloc/terminal_bloc.dart';
import 'package:helix_terminator/bloc/workspace_bloc.dart';
import 'package:helix_terminator/models/host.dart';
import 'package:helix_terminator/models/workspace.dart';
import 'package:helix_terminator/screens/dashboard_screen.dart';
import 'package:helix_terminator/screens/login_screen.dart';
import 'package:helix_terminator/screens/splash_screen.dart';
import 'package:helix_terminator/screens/terminal_screen.dart';
import 'package:helix_terminator/screens/workspace_list_screen.dart';
import 'package:helix_terminator/services/auth_service.dart';
import 'package:helix_terminator/services/host_service.dart';
import 'package:helix_terminator/services/terminal_service.dart';
import 'package:helix_terminator/services/workspace_service.dart';

// ---------------------------------------------------------------------------
// Test doubles. Each mirrors an existing, already-proven pattern from the
// real bloc unit-test suites (host_bloc_test.dart / workspace_bloc_test.dart
// / app_boot_journey_test.dart / terminal_bloc_test.dart) so screen-level
// golden coverage composes with the existing test architecture rather than
// inventing a parallel one.
// ---------------------------------------------------------------------------

class MockAuthService extends Mock implements AuthService {}

class MockHostService extends Mock implements HostService {}

class MockWorkspaceService extends Mock implements WorkspaceService {}

/// Same technique as test/terminal_bloc_test.dart's FakeTerminalService:
/// TerminalService is a concrete class backed by a real dart:io WebSocket
/// (unsuitable for a host unit test), so this subclasses it and overrides
/// its public, non-final methods with an in-memory fake.
class FakeTerminalService extends TerminalService {
  TerminalMessageCallback? _callback;

  @override
  Future<void> connect(String hostId) async {}

  @override
  Future<void> sendCommand(String command) async {}

  @override
  Future<void> disconnect() async {}

  @override
  void onMessage(TerminalMessageCallback callback) {
    _callback = callback;
  }

  @override
  void removeOnMessage() {
    _callback = null;
  }

  @override
  Future<void> dispose() async {}

  /// Simulates the WebSocket pushing an inbound frame, exactly like the real
  /// `_socket!.listen((data) { ... _onMessageCallback?.call(message); })`.
  void pushMessage(String message) => _callback?.call(message);
}

// A representative phone viewport, shared by every screen below so the
// coverage matrix is visually comparable and the harness stays simple.
const Size _phoneSize = Size(390, 844);

/// Two fixed 120ms pumps (never pumpAndSettle): several of these screens
/// contain an indefinitely-repeating animation (CircularProgressIndicator /
/// xterm's cursor blink), which would hang pumpAndSettle forever. Fixed pumps
/// driven by the test binding's synthetic clock are what test/
/// terminal_view_golden_test.dart already established as deterministic for
/// this repo's golden suite.
Future<void> _settle(WidgetTester tester) async {
  await tester.pump(const Duration(milliseconds: 120));
  await tester.pump(const Duration(milliseconds: 120));
}

void main() {
  setUpAll(() {
    // mocktail fallback values for any() used with non-builtin argument
    // types across the mocked services below.
    registerFallbackValue(
      Host(id: 'fallback', name: 'fallback', address: '0.0.0.0', createdAt: DateTime.utc(2000)),
    );
  });

  // -------------------------------------------------------------------
  // SplashScreen — brand entrance, real AnimationController.
  // -------------------------------------------------------------------
  group('SplashScreen (host-rendered golden, §11.4.170)', () {
    Future<AuthBloc> pumpSplash(WidgetTester tester, ThemeData theme) async {
      final authService = MockAuthService();
      when(() => authService.isAuthenticated()).thenAnswer((_) async => false);
      final authBloc = AuthBloc(authService: authService);
      addTearDown(authBloc.close);

      await tester.binding.setSurfaceSize(_phoneSize);
      await tester.pumpWidget(
        BlocProvider<AuthBloc>.value(
          value: authBloc,
          child: MaterialApp(
            debugShowCheckedModeBanner: false,
            theme: theme,
            home: const SplashScreen(),
          ),
        ),
      );
      // Advance the entrance animation to its steady state (opacity/scale
      // fully settled) while staying under the real 2s delayed
      // AuthCheckRequested callback, so no navigation fires mid-capture —
      // this is the exact frame the golden below asserts.
      await tester.pump(const Duration(milliseconds: 1500));
      return authBloc;
    }

    /// Crosses the real 2s `Future.delayed` threshold inside SplashScreen's
    /// initState and lets its resulting AuthCheckRequested -> Navigator
    /// handoff settle, so the timer is fully drained BEFORE the test ends.
    /// Without this, AutomatedTestWidgetsFlutterBinding's post-test invariant
    /// check ("A Timer is still pending even after the widget tree was
    /// disposed") fails the test for a test-harness reason, not a product
    /// defect (§11.4.1) — the exact issue test/README.md already documents
    /// as defect-adjacent for widget_test.dart, fixed the same way there.
    Future<void> drainSplashTimer(WidgetTester tester) async {
      await tester.pump(const Duration(milliseconds: 600));
      await tester.pump();
      await tester.pump(const Duration(milliseconds: 100));
    }

    void assertContentOnScreen() {
      expect(find.text('HelixTerminator'), findsOneWidget);
      expect(find.text('Secure. Fast. Reliable.'), findsOneWidget);
      expect(find.byIcon(Icons.shield_outlined), findsOneWidget);
      expect(find.byType(CircularProgressIndicator), findsOneWidget);
    }

    testWidgets('renders real brand content — DARK theme', (tester) async {
      addTearDown(() => tester.binding.setSurfaceSize(null));
      await pumpSplash(tester, HTTheme.dark());

      await expectLater(
        find.byType(SplashScreen),
        matchesGoldenFile('goldens/splash_screen_dark.png'),
      );
      assertContentOnScreen();
      expect(tester.takeException(), isNull);

      await drainSplashTimer(tester);
      expect(tester.takeException(), isNull);
    });

    testWidgets('renders real brand content — LIGHT theme', (tester) async {
      addTearDown(() => tester.binding.setSurfaceSize(null));
      await pumpSplash(tester, HTTheme.light());

      await expectLater(
        find.byType(SplashScreen),
        matchesGoldenFile('goldens/splash_screen_light.png'),
      );
      assertContentOnScreen();
      expect(tester.takeException(), isNull);

      await drainSplashTimer(tester);
      expect(tester.takeException(), isNull);
    });
  });

  // -------------------------------------------------------------------
  // LoginScreen — real Form + real AuthBloc wiring.
  // -------------------------------------------------------------------
  group('LoginScreen (host-rendered golden, §11.4.170)', () {
    Future<AuthBloc> pumpLogin(WidgetTester tester, ThemeData theme) async {
      final authService = MockAuthService();
      final authBloc = AuthBloc(authService: authService);
      addTearDown(authBloc.close);

      await tester.binding.setSurfaceSize(_phoneSize);
      await tester.pumpWidget(
        BlocProvider<AuthBloc>.value(
          value: authBloc,
          child: MaterialApp(
            debugShowCheckedModeBanner: false,
            theme: theme,
            home: const LoginScreen(),
          ),
        ),
      );
      await tester.pump();
      return authBloc;
    }

    void assertContentOnScreen() {
      expect(find.text('Welcome to HelixTerminator'), findsOneWidget);
      expect(find.text('Sign in to your account'), findsOneWidget);
      expect(find.widgetWithText(TextFormField, 'Email'), findsOneWidget);
      expect(find.widgetWithText(TextFormField, 'Password'), findsOneWidget);
      expect(find.text('Forgot password?'), findsOneWidget);
      expect(find.widgetWithText(FilledButton, 'Sign In'), findsOneWidget);
      // The Row this text sits in previously overflowed its 400px-constrained
      // parent by ~58px (test/README.md defect #2) — its continued presence
      // + the exception-free assertion below together re-prove the fix.
      expect(find.text("Don't have an account?"), findsOneWidget);
      expect(find.widgetWithText(TextButton, 'Register'), findsOneWidget);
    }

    testWidgets('renders real login form — DARK theme', (tester) async {
      addTearDown(() => tester.binding.setSurfaceSize(null));
      await pumpLogin(tester, HTTheme.dark());

      await expectLater(
        find.byType(LoginScreen),
        matchesGoldenFile('goldens/login_screen_dark.png'),
      );
      assertContentOnScreen();
      // Proves no RenderFlex overflow (or any other framework exception) at
      // this 390px width — the regression class §11.4.170 exists to catch.
      expect(tester.takeException(), isNull);
    });

    testWidgets('renders real login form — LIGHT theme', (tester) async {
      addTearDown(() => tester.binding.setSurfaceSize(null));
      await pumpLogin(tester, HTTheme.light());

      await expectLater(
        find.byType(LoginScreen),
        matchesGoldenFile('goldens/login_screen_light.png'),
      );
      assertContentOnScreen();
      expect(tester.takeException(), isNull);
    });
  });

  // -------------------------------------------------------------------
  // DashboardScreen ("Home" tab) — real HostBloc-backed content.
  // -------------------------------------------------------------------
  group('DashboardScreen (host-rendered golden, §11.4.170)', () {
    List<Host> seedHosts() => [
          Host(
            id: '1',
            name: 'prod-db-01',
            address: '10.0.0.11',
            createdAt: DateTime.utc(2026, 1, 1),
            status: 'online',
            tags: const ['db', 'prod'],
          ),
          Host(
            id: '2',
            name: 'staging-web-02',
            address: '10.0.0.22',
            createdAt: DateTime.utc(2026, 1, 2),
            status: 'offline',
          ),
          Host(
            id: '3',
            name: 'edge-cache-03',
            address: '10.0.0.33',
            createdAt: DateTime.utc(2026, 1, 3),
            status: 'connecting',
          ),
        ];

    Future<HostBloc> pumpDashboard(WidgetTester tester, ThemeData theme) async {
      final hostService = MockHostService();
      when(() => hostService.getHosts()).thenAnswer((_) async => seedHosts());
      final hostBloc = HostBloc(hostService: hostService);
      addTearDown(hostBloc.close);

      await tester.binding.setSurfaceSize(_phoneSize);
      await tester.pumpWidget(
        BlocProvider<HostBloc>.value(
          value: hostBloc,
          child: MaterialApp(
            debugShowCheckedModeBanner: false,
            theme: theme,
            home: const DashboardScreen(),
          ),
        ),
      );
      // First pump mounts _DashboardBody, whose initState fires
      // HostLoadRequested; the following pumps flush the mocked async
      // getHosts() future and the resulting Loading -> Loaded rebuild.
      await tester.pump();
      await _settle(tester);
      return hostBloc;
    }

    void assertContentOnScreen(List<Host> hosts) {
      expect(find.text('Welcome back'), findsOneWidget);
      expect(find.text('Manage your hosts and connections'), findsOneWidget);
      expect(find.text('Hosts'), findsWidgets);
      expect(find.text('Active Sessions'), findsOneWidget);
      expect(find.text('Notifications'), findsWidgets);
      expect(find.text('Recent Hosts'), findsOneWidget);
      expect(find.text('Quick Actions'), findsOneWidget);
      for (final host in hosts) {
        expect(find.text(host.name), findsOneWidget);
        expect(find.text('${host.address}:${host.port}'), findsOneWidget);
      }
      // Bottom navigation — the persistent chrome around every dashboard tab.
      expect(find.text('Home'), findsOneWidget);
      expect(find.text('Search'), findsOneWidget);
      expect(find.text('Alerts'), findsOneWidget);
    }

    testWidgets('renders real host summary — DARK theme', (tester) async {
      addTearDown(() => tester.binding.setSurfaceSize(null));
      await pumpDashboard(tester, HTTheme.dark());

      await expectLater(
        find.byType(DashboardScreen),
        matchesGoldenFile('goldens/dashboard_screen_dark.png'),
      );
      assertContentOnScreen(seedHosts());
      expect(tester.takeException(), isNull);
    });

    testWidgets('renders real host summary — LIGHT theme', (tester) async {
      addTearDown(() => tester.binding.setSurfaceSize(null));
      await pumpDashboard(tester, HTTheme.light());

      await expectLater(
        find.byType(DashboardScreen),
        matchesGoldenFile('goldens/dashboard_screen_light.png'),
      );
      assertContentOnScreen(seedHosts());
      expect(tester.takeException(), isNull);
    });
  });

  // -------------------------------------------------------------------
  // WorkspaceListScreen — real WorkspaceBloc-backed content.
  // -------------------------------------------------------------------
  group('WorkspaceListScreen (host-rendered golden, §11.4.170)', () {
    List<Workspace> seedWorkspaces() => [
          Workspace(
            id: 'w1',
            name: 'Platform Infra',
            description: 'Core infrastructure hosts',
            hostIds: const ['1', '2', '3'],
            memberIds: const ['u1', 'u2'],
            createdAt: DateTime.utc(2026, 1, 1),
          ),
          Workspace(
            id: 'w2',
            name: 'Data Team',
            hostIds: const ['4'],
            memberIds: const ['u1', 'u3', 'u4'],
            createdAt: DateTime.utc(2026, 1, 2),
          ),
        ];

    Future<WorkspaceBloc> pumpWorkspaceList(WidgetTester tester, ThemeData theme) async {
      final workspaceService = MockWorkspaceService();
      when(() => workspaceService.getWorkspaces()).thenAnswer((_) async => seedWorkspaces());
      final workspaceBloc = WorkspaceBloc(workspaceService: workspaceService);
      addTearDown(workspaceBloc.close);

      await tester.binding.setSurfaceSize(_phoneSize);
      await tester.pumpWidget(
        BlocProvider<WorkspaceBloc>.value(
          value: workspaceBloc,
          child: MaterialApp(
            debugShowCheckedModeBanner: false,
            theme: theme,
            home: const WorkspaceListScreen(),
          ),
        ),
      );
      // WorkspaceListScreen's initState fires WorkspaceLoadRequested itself.
      await tester.pump();
      await _settle(tester);
      return workspaceBloc;
    }

    void assertContentOnScreen(List<Workspace> workspaces) {
      expect(find.text('Workspaces'), findsOneWidget);
      expect(find.text('Search workspaces...'), findsOneWidget);
      for (final workspace in workspaces) {
        expect(find.text(workspace.name), findsOneWidget);
        expect(find.text('${workspace.hostIds.length}'), findsWidgets);
        expect(find.text('${workspace.memberIds.length}'), findsWidgets);
      }
      expect(find.widgetWithText(FloatingActionButton, 'Create'), findsOneWidget);
    }

    testWidgets('renders real workspace cards — DARK theme', (tester) async {
      addTearDown(() => tester.binding.setSurfaceSize(null));
      await pumpWorkspaceList(tester, HTTheme.dark());

      await expectLater(
        find.byType(WorkspaceListScreen),
        matchesGoldenFile('goldens/workspace_list_screen_dark.png'),
      );
      assertContentOnScreen(seedWorkspaces());
      expect(tester.takeException(), isNull);
    });

    testWidgets('renders real workspace cards — LIGHT theme', (tester) async {
      addTearDown(() => tester.binding.setSurfaceSize(null));
      await pumpWorkspaceList(tester, HTTheme.light());

      await expectLater(
        find.byType(WorkspaceListScreen),
        matchesGoldenFile('goldens/workspace_list_screen_light.png'),
      );
      assertContentOnScreen(seedWorkspaces());
      expect(tester.takeException(), isNull);
    });
  });

  // -------------------------------------------------------------------
  // TerminalScreen — real xterm engine + real TerminalBloc.
  // -------------------------------------------------------------------
  group('TerminalScreen (host-rendered golden, §11.4.170)', () {
    const String hostId = 'host-42';
    const String hostName = 'prod-db-01';
    const String ansiLine = '\x1b[1;32muser@helix\x1b[0m:\x1b[1;34m~\x1b[0m\$ uptime\r\n'
        '10:15  up 3 days, load 0.42\r\n'
        '\x1b[1;32muser@helix\x1b[0m:\x1b[1;34m~\x1b[0m\$ ';

    Future<TerminalBloc> pumpTerminal(WidgetTester tester, ThemeData theme) async {
      final fakeService = FakeTerminalService();
      final terminalBloc = TerminalBloc(terminalService: fakeService);
      addTearDown(terminalBloc.close);

      await tester.binding.setSurfaceSize(const Size(420, 900));
      await tester.pumpWidget(
        BlocProvider<TerminalBloc>.value(
          value: terminalBloc,
          child: MaterialApp(
            debugShowCheckedModeBanner: false,
            theme: theme,
            home: const TerminalScreen(hostId: hostId, hostName: hostName),
          ),
        ),
      );
      // TerminalScreen's initState fires TerminalConnectRequested itself;
      // flush the mocked async connect() and the Connecting -> Connected
      // rebuild.
      await tester.pump();
      await _settle(tester);

      // Push a real server frame through the SAME onMessage callback wiring
      // TerminalBloc installs on connect (proven end-to-end already by
      // test/terminal_bloc_test.dart) so the real xterm engine genuinely
      // parses ANSI content, exactly like test/terminal_view_golden_test.dart
      // does for the standalone widget.
      fakeService.pushMessage(ansiLine);
      await _settle(tester);
      return terminalBloc;
    }

    void assertContentOnScreen() {
      expect(find.text(hostName), findsOneWidget);
      // "Connected" is rendered twice: the AppBar's ConnectionStatus chip and
      // the bottom status-bar label — both real, both must be on screen.
      expect(find.text('Connected'), findsNWidgets(2));
      expect(find.text(hostId), findsOneWidget);
      expect(find.byIcon(Icons.palette), findsOneWidget);
      expect(find.byIcon(Icons.fullscreen), findsOneWidget);
      expect(find.byIcon(Icons.settings), findsOneWidget);
      expect(find.byIcon(Icons.power_settings_new), findsOneWidget);
      expect(find.text('Enter command...'), findsOneWidget);
      // The real xterm painter widget is present (not a placeholder) — the
      // same content-oracle invariant test/terminal_view_golden_test.dart
      // asserts for the standalone TerminalView widget.
      expect(find.byType(xterm.TerminalView), findsOneWidget);
    }

    testWidgets('renders real connected terminal session — DARK theme', (tester) async {
      addTearDown(() => tester.binding.setSurfaceSize(null));
      await pumpTerminal(tester, HTTheme.dark());

      await expectLater(
        find.byType(TerminalScreen),
        matchesGoldenFile('goldens/terminal_screen_dark.png'),
      );
      assertContentOnScreen();
      expect(tester.takeException(), isNull);
    });

    testWidgets('renders real connected terminal session — LIGHT theme', (tester) async {
      addTearDown(() => tester.binding.setSurfaceSize(null));
      await pumpTerminal(tester, HTTheme.light());

      await expectLater(
        find.byType(TerminalScreen),
        matchesGoldenFile('goldens/terminal_screen_light.png'),
      );
      assertContentOnScreen();
      expect(tester.takeException(), isNull);
    });
  });
}
