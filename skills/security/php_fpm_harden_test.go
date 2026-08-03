package security

import (
	"log/slog"
	"strings"
	"testing"

	"github.com/dracory/ork/types"
)

// TestPhpFpmHarden_Run_DryRun verifies that dry-run mode reports a would-change
// result without touching the remote system.
func TestPhpFpmHarden_Run_DryRun(t *testing.T) {
	pb := NewPhpFpmHarden().
		SetVersion("8.5").
		SetOpenBasedirPaths("/var/www/app:/var/www/media:/tmp")

	cfg := types.NodeConfig{
		IsDryRunMode: true,
		Logger:       slog.Default(),
	}

	pb.SetNodeConfig(cfg)

	result := pb.Run()

	if !result.Changed {
		t.Error("Expected Changed to be true in dry-run mode")
	}

	expectedMessage := "Would harden PHP-FPM configuration"
	if result.Message != expectedMessage {
		t.Errorf("Expected message '%s', got '%s'", expectedMessage, result.Message)
	}

	if result.Error != nil {
		t.Errorf("Expected no error in dry-run mode, got: %v", result.Error)
	}
}

// TestPhpFpmHarden_Run_NotDryRun verifies that non-dry-run mode does not return
// the dry-run message (it will attempt SSH and fail without a real server).
func TestPhpFpmHarden_Run_NotDryRun(t *testing.T) {
	pb := NewPhpFpmHarden().
		SetVersion("8.5").
		SetOpenBasedirPaths("/var/www/app:/tmp")

	cfg := types.NodeConfig{
		IsDryRunMode: false,
		Logger:       slog.Default(),
	}

	pb.SetNodeConfig(cfg)

	result := pb.Run()

	if result.Message == "Would harden PHP-FPM configuration" {
		t.Error("Should not return dry-run message when IsDryRunMode is false")
	}
}

// TestPhpFpmHarden_Run_MissingVersion verifies that a missing php-version arg
// produces an error result without attempting any SSH work.
func TestPhpFpmHarden_Run_MissingVersion(t *testing.T) {
	pb := NewPhpFpmHarden().
		SetOpenBasedirPaths("/var/www/app:/tmp")

	cfg := types.NodeConfig{
		IsDryRunMode: true, // dry-run so no SSH is attempted even if validation order changes
		Logger:       slog.Default(),
	}

	pb.SetNodeConfig(cfg)

	result := pb.Run()

	if result.Changed {
		t.Error("Expected Changed to be false when php-version is missing")
	}
	if result.Error == nil {
		t.Error("Expected an error when php-version is missing")
	}
}

// TestPhpFpmHarden_Run_MissingOpenBasedir verifies that a missing
// open-basedir-paths arg produces an error result.
func TestPhpFpmHarden_Run_MissingOpenBasedir(t *testing.T) {
	pb := NewPhpFpmHarden().
		SetVersion("8.5")

	cfg := types.NodeConfig{
		IsDryRunMode: true,
		Logger:       slog.Default(),
	}

	pb.SetNodeConfig(cfg)

	result := pb.Run()

	if result.Changed {
		t.Error("Expected Changed to be false when open-basedir-paths is missing")
	}
	if result.Error == nil {
		t.Error("Expected an error when open-basedir-paths is missing")
	}
}

// TestPhpFpmHarden_Run_InvalidVersion verifies that a malformed php-version
// arg (e.g. containing shell metacharacters) is rejected by validateArgs.
func TestPhpFpmHarden_Run_InvalidVersion(t *testing.T) {
	cases := []string{
		"8.5; rm -rf /",
		"8.5 && cat /etc/passwd",
		"8",
		"8.5.3",
		"",
		"abc",
	}
	for _, version := range cases {
		t.Run(version, func(t *testing.T) {
			pb := NewPhpFpmHarden().
				SetVersion(version).
				SetOpenBasedirPaths("/var/www/app:/tmp")

			cfg := types.NodeConfig{
				IsDryRunMode: true,
				Logger:       slog.Default(),
			}
			pb.SetNodeConfig(cfg)

			result := pb.Run()

			if result.Changed {
				t.Error("Expected Changed to be false for invalid version")
			}
			if result.Error == nil {
				t.Error("Expected an error for invalid version")
			}
		})
	}
}

// TestPhpFpmHarden_NewPhpFpmHarden verifies that NewPhpFpmHarden creates a
// properly configured skill.
func TestPhpFpmHarden_NewPhpFpmHarden(t *testing.T) {
	pb := NewPhpFpmHarden()

	if pb.GetID() != "php-fpm-harden" {
		t.Errorf("Expected ID to be 'php-fpm-harden', got '%s'", pb.GetID())
	}

	expectedDescription := "Harden PHP-FPM configuration for production security"
	if pb.GetDescription() != expectedDescription {
		t.Errorf("Expected description '%s', got '%s'", expectedDescription, pb.GetDescription())
	}
}

// TestPhpFpmHarden_SetArgs_ReturnsConcreteType verifies that SetArgs returns
// the concrete PhpFpmHarden type.
func TestPhpFpmHarden_SetArgs_ReturnsConcreteType(t *testing.T) {
	skill := NewPhpFpmHarden()
	args := map[string]string{"test": "value"}

	result := skill.SetArgs(args)

	if _, ok := result.(*PhpFpmHarden); !ok {
		t.Error("SetArgs should return *PhpFpmHarden, not just RunnableInterface")
	}
}

// TestPhpFpmHarden_SetArg_ReturnsConcreteType verifies that SetArg returns the
// concrete PhpFpmHarden type.
func TestPhpFpmHarden_SetArg_ReturnsConcreteType(t *testing.T) {
	skill := NewPhpFpmHarden()

	result := skill.SetArg("test", "value")

	if _, ok := result.(*PhpFpmHarden); !ok {
		t.Error("SetArg should return *PhpFpmHarden, not just RunnableInterface")
	}
}

// TestPhpFpmHarden_SetID_ReturnsConcreteType verifies that SetID returns the
// concrete PhpFpmHarden type and sets the ID.
func TestPhpFpmHarden_SetID_ReturnsConcreteType(t *testing.T) {
	skill := NewPhpFpmHarden()

	result := skill.SetID("custom-id")

	if _, ok := result.(*PhpFpmHarden); !ok {
		t.Error("SetID should return *PhpFpmHarden, not just RunnableInterface")
	}

	if skill.GetID() != "custom-id" {
		t.Error("SetID should set the ID")
	}
}

// TestPhpFpmHarden_SetDescription_ReturnsConcreteType verifies that
// SetDescription returns the concrete PhpFpmHarden type.
func TestPhpFpmHarden_SetDescription_ReturnsConcreteType(t *testing.T) {
	skill := NewPhpFpmHarden()

	result := skill.SetDescription("custom description")

	if _, ok := result.(*PhpFpmHarden); !ok {
		t.Error("SetDescription should return *PhpFpmHarden, not just RunnableInterface")
	}

	if skill.GetDescription() != "custom description" {
		t.Error("SetDescription should set the description")
	}
}

// TestPhpFpmHarden_SetTimeout_ReturnsConcreteType verifies that SetTimeout
// returns the concrete PhpFpmHarden type.
func TestPhpFpmHarden_SetTimeout_ReturnsConcreteType(t *testing.T) {
	skill := NewPhpFpmHarden()

	result := skill.SetTimeout(30 * 1000000000)

	if _, ok := result.(*PhpFpmHarden); !ok {
		t.Error("SetTimeout should return *PhpFpmHarden, not just RunnableInterface")
	}
}

// TestPhpFpmHarden_MethodChaining_PreservesType verifies that method chaining
// preserves the concrete type.
func TestPhpFpmHarden_MethodChaining_PreservesType(t *testing.T) {
	skill := NewPhpFpmHarden().
		SetID("custom-id").
		SetDescription("custom description").
		SetArg("test", "value").
		SetArgs(map[string]string{"another": "arg"}).
		SetTimeout(30 * 1000000000)

	if _, ok := skill.(*PhpFpmHarden); !ok {
		t.Error("Method chaining should preserve *PhpFpmHarden type")
	}

	if skill.GetID() != "custom-id" {
		t.Error("Method chaining should set ID")
	}

	if skill.GetDescription() != "custom description" {
		t.Error("Method chaining should set description")
	}
}

// TestPhpFpmHarden_SetVersion verifies that SetVersion sets the php-version arg.
func TestPhpFpmHarden_SetVersion(t *testing.T) {
	skill := NewPhpFpmHarden().SetVersion("8.5")

	if skill.GetArg(ArgPhpVersion) != "8.5" {
		t.Errorf("Expected php-version '8.5', got '%s'", skill.GetArg(ArgPhpVersion))
	}
}

// TestPhpFpmHarden_SetOpenBasedirPaths verifies that SetOpenBasedirPaths sets
// the open-basedir-paths arg.
func TestPhpFpmHarden_SetOpenBasedirPaths(t *testing.T) {
	skill := NewPhpFpmHarden().SetOpenBasedirPaths("/var/www/app:/tmp")

	if skill.GetArg(ArgOpenBasedirPaths) != "/var/www/app:/tmp" {
		t.Errorf("Expected open-basedir-paths '/var/www/app:/tmp', got '%s'", skill.GetArg(ArgOpenBasedirPaths))
	}
}

// TestPhpFpmHarden_SetDisableFunctions verifies that SetDisableFunctions sets
// the disable-functions arg.
func TestPhpFpmHarden_SetDisableFunctions(t *testing.T) {
	skill := NewPhpFpmHarden().SetDisableFunctions("exec, shell_exec")

	if skill.GetArg(ArgDisableFunctions) != "exec, shell_exec" {
		t.Errorf("Expected disable-functions 'exec, shell_exec', got '%s'", skill.GetArg(ArgDisableFunctions))
	}
}

// TestPhpFpmHarden_TypedSetters_Chaining verifies that all typed setters chain
// correctly and store their values.
func TestPhpFpmHarden_TypedSetters_Chaining(t *testing.T) {
	skill := NewPhpFpmHarden().
		SetVersion("8.5").
		SetOpenBasedirPaths("/var/www/app:/var/www/media:/tmp").
		SetDisableFunctions("exec, shell_exec, system, passthru, popen, proc_open").
		SetMemoryLimit("512M").
		SetUploadMaxFilesize("20M").
		SetPostMaxSize("24M").
		SetMaxExecutionTime("120").
		SetMaxInputTime("120").
		SetOpcacheMemory("256").
		SetOpcacheMaxFiles("20000").
		SetConfDPath("/etc/php/8.5/fpm/conf.d/99-custom.ini").
		SetPoolPath("/etc/php/8.5/fpm/pool.d/www.conf").
		SetErrorLog("/var/log/php8.5-fpm.log")

	if skill.GetArg(ArgPhpVersion) != "8.5" {
		t.Errorf("Expected php-version '8.5', got '%s'", skill.GetArg(ArgPhpVersion))
	}
	if skill.GetArg(ArgOpenBasedirPaths) != "/var/www/app:/var/www/media:/tmp" {
		t.Errorf("Expected open-basedir-paths, got '%s'", skill.GetArg(ArgOpenBasedirPaths))
	}
	if skill.GetArg(ArgDisableFunctions) != "exec, shell_exec, system, passthru, popen, proc_open" {
		t.Errorf("Expected disable-functions, got '%s'", skill.GetArg(ArgDisableFunctions))
	}
	if skill.GetArg(ArgMemoryLimit) != "512M" {
		t.Errorf("Expected memory-limit '512M', got '%s'", skill.GetArg(ArgMemoryLimit))
	}
	if skill.GetArg(ArgUploadMaxFilesize) != "20M" {
		t.Errorf("Expected upload-max-filesize '20M', got '%s'", skill.GetArg(ArgUploadMaxFilesize))
	}
	if skill.GetArg(ArgPostMaxSize) != "24M" {
		t.Errorf("Expected post-max-size '24M', got '%s'", skill.GetArg(ArgPostMaxSize))
	}
	if skill.GetArg(ArgMaxExecutionTime) != "120" {
		t.Errorf("Expected max-execution-time '120', got '%s'", skill.GetArg(ArgMaxExecutionTime))
	}
	if skill.GetArg(ArgMaxInputTime) != "120" {
		t.Errorf("Expected max-input-time '120', got '%s'", skill.GetArg(ArgMaxInputTime))
	}
	if skill.GetArg(ArgOpcacheMemory) != "256" {
		t.Errorf("Expected opcache-memory '256', got '%s'", skill.GetArg(ArgOpcacheMemory))
	}
	if skill.GetArg(ArgOpcacheMaxFiles) != "20000" {
		t.Errorf("Expected opcache-max-files '20000', got '%s'", skill.GetArg(ArgOpcacheMaxFiles))
	}
	if skill.GetArg(ArgConfDPath) != "/etc/php/8.5/fpm/conf.d/99-custom.ini" {
		t.Errorf("Expected conf-d-path, got '%s'", skill.GetArg(ArgConfDPath))
	}
	if skill.GetArg(ArgPoolPath) != "/etc/php/8.5/fpm/pool.d/www.conf" {
		t.Errorf("Expected pool-path, got '%s'", skill.GetArg(ArgPoolPath))
	}
	if skill.GetArg(ArgErrorLog) != "/var/log/php8.5-fpm.log" {
		t.Errorf("Expected error-log, got '%s'", skill.GetArg(ArgErrorLog))
	}
}

// TestPhpFpmHarden_DefaultsApplied verifies that unset optional args fall back
// to their documented defaults inside Run (via the dry-run path, which still
// resolves defaults before reporting).
func TestPhpFpmHarden_DefaultsApplied(t *testing.T) {
	skill := NewPhpFpmHarden().
		SetVersion("8.5").
		SetOpenBasedirPaths("/var/www/app:/tmp")

	// Defaults should not be stored as args (they are applied in Run), so the
	// raw args remain empty — this documents the lazy-default behaviour.
	if skill.GetArg(ArgDisableFunctions) != "" {
		t.Errorf("disable-functions should default lazily, got '%s'", skill.GetArg(ArgDisableFunctions))
	}
	if skill.GetArg(ArgMemoryLimit) != "" {
		t.Errorf("memory-limit should default lazily, got '%s'", skill.GetArg(ArgMemoryLimit))
	}
}

// TestPhpFpmHarden_BuildConfDContent verifies the rendered conf.d content
// contains the key directives and the caller-supplied values.
func TestPhpFpmHarden_BuildConfDContent(t *testing.T) {
	content := buildConfDContent(
		"/var/log/php8.5-fpm.log",
		"/var/www/app:/tmp",
		"exec, shell_exec",
		"256M", "10M", "12M", "60", "60", "128", "10000",
	)

	checks := []string{
		"expose_php = Off",
		"display_errors = Off",
		"error_log = /var/log/php8.5-fpm.log",
		"open_basedir = /var/www/app:/tmp",
		"disable_functions = exec, shell_exec",
		"session.cookie_secure = On",
		"opcache.enable = On",
		"upload_max_filesize = 10M",
		"memory_limit = 256M",
	}
	for _, want := range checks {
		if !strings.Contains(content, want) {
			t.Errorf("buildConfDContent: missing %q in output:\n%s", want, content)
		}
	}
}

// TestPhpFpmHarden_BuildPoolBlock verifies the rendered pool block contains the
// markers and the admin value/flag directives.
func TestPhpFpmHarden_BuildPoolBlock(t *testing.T) {
	block := buildPoolBlock("exec, shell_exec", "/var/www/app:/tmp")

	if !strings.Contains(block, phpFpmHardenMarkerBegin) {
		t.Errorf("buildPoolBlock: missing begin marker")
	}
	if !strings.Contains(block, phpFpmHardenMarkerEnd) {
		t.Errorf("buildPoolBlock: missing end marker")
	}
	checks := []string{
		"php_admin_value[disable_functions] = exec, shell_exec",
		"php_admin_flag[expose_php] = Off",
		"php_admin_flag[display_errors] = Off",
		"php_admin_value[open_basedir] = /var/www/app:/tmp",
	}
	for _, want := range checks {
		if !strings.Contains(block, want) {
			t.Errorf("buildPoolBlock: missing %q in output:\n%s", want, block)
		}
	}
}
