package fs

import (
	"log/slog"
	"testing"
	"time"

	"github.com/dracory/ork/types"
)

// TestChangeMode_Run_DryRun verifies that dry-run mode reports the would-change message.
func TestChangeMode_Run_DryRun(t *testing.T) {
	pb := NewChangeMode()

	cfg := types.NodeConfig{
		IsDryRunMode: true,
		Logger:       slog.Default(),
		Args:         map[string]string{},
	}

	pb.SetNodeConfig(cfg)
	pb.SetArg(ArgPath, "/var/www/myapp/.ssh")
	pb.SetArg(ArgMode, "700")

	result := pb.Run()

	if !result.Changed {
		t.Error("Expected Changed to be true in dry-run mode")
	}

	expectedMessage := "Would change mode to 700 on /var/www/myapp/.ssh"
	if result.Message != expectedMessage {
		t.Errorf("Expected message '%s', got '%s'", expectedMessage, result.Message)
	}

	if result.Error != nil {
		t.Errorf("Expected no error in dry-run mode, got: %v", result.Error)
	}
}

// TestChangeMode_Run_NoPath verifies that missing ArgPath returns an error.
func TestChangeMode_Run_NoPath(t *testing.T) {
	pb := NewChangeMode()

	cfg := types.NodeConfig{
		IsDryRunMode: false,
		Logger:       slog.Default(),
		Args:         map[string]string{},
	}

	pb.SetNodeConfig(cfg)
	pb.SetArg(ArgMode, "755")

	result := pb.Run()

	if result.Changed {
		t.Error("Expected Changed to be false when no path specified")
	}

	if result.Error == nil {
		t.Error("Expected an error when no path is specified")
	}
}

// TestChangeMode_Run_NoMode verifies that missing ArgMode returns an error.
func TestChangeMode_Run_NoMode(t *testing.T) {
	pb := NewChangeMode()

	cfg := types.NodeConfig{
		IsDryRunMode: false,
		Logger:       slog.Default(),
		Args:         map[string]string{},
	}

	pb.SetNodeConfig(cfg)
	pb.SetArg(ArgPath, "/var/www/myapp")

	result := pb.Run()

	if result.Changed {
		t.Error("Expected Changed to be false when no mode specified")
	}

	if result.Error == nil {
		t.Error("Expected an error when no mode is specified")
	}
}

// TestChangeMode_Run_InvalidMode verifies that an invalid mode returns an error.
func TestChangeMode_Run_InvalidMode(t *testing.T) {
	pb := NewChangeMode()

	cfg := types.NodeConfig{
		IsDryRunMode: true,
		Logger:       slog.Default(),
		Args:         map[string]string{},
	}

	pb.SetNodeConfig(cfg)
	pb.SetArg(ArgPath, "/var/www/myapp")
	pb.SetArg(ArgMode, "999") // 9 is not a valid octal digit

	result := pb.Run()

	if result.Changed {
		t.Error("Expected Changed to be false for invalid mode")
	}

	if result.Error == nil {
		t.Error("Expected an error for invalid mode '999'")
	}
}

// TestChangeMode_Run_NotDryRun verifies that non-dry-run mode does not return the dry-run message.
func TestChangeMode_Run_NotDryRun(t *testing.T) {
	pb := NewChangeMode()

	cfg := types.NodeConfig{
		IsDryRunMode: false,
		Logger:       slog.Default(),
		Args:         map[string]string{},
	}

	pb.SetNodeConfig(cfg)
	pb.SetArg(ArgPath, "/var/www/myapp")
	pb.SetArg(ArgMode, "755")

	result := pb.Run()

	if result.Message == "Would change mode to 755 on /var/www/myapp" {
		t.Error("Should not return dry-run message when IsDryRunMode is false")
	}
}

// TestChangeMode_NewChangeMode verifies that NewChangeMode creates a properly configured skill.
func TestChangeMode_NewChangeMode(t *testing.T) {
	pb := NewChangeMode()

	if pb.GetID() != "fs-change-mode" {
		t.Errorf("Expected ID to be 'fs-change-mode', got '%s'", pb.GetID())
	}

	expectedDescription := "Change file/directory permissions (chmod)"
	if pb.GetDescription() != expectedDescription {
		t.Errorf("Expected description '%s', got '%s'", expectedDescription, pb.GetDescription())
	}
}

// TestChangeMode_SetArgs_ReturnsConcreteType verifies that SetArgs returns the concrete ChangeMode type.
func TestChangeMode_SetArgs_ReturnsConcreteType(t *testing.T) {
	skill := NewChangeMode()
	args := map[string]string{ArgPath: "/tmp/test"}

	result := skill.SetArgs(args)

	if _, ok := result.(*ChangeMode); !ok {
		t.Error("SetArgs should return *ChangeMode, not just RunnableInterface")
	}
}

// TestChangeMode_SetArg_ReturnsConcreteType verifies that SetArg returns the concrete ChangeMode type.
func TestChangeMode_SetArg_ReturnsConcreteType(t *testing.T) {
	skill := NewChangeMode()

	result := skill.SetArg(ArgPath, "/tmp/test")

	if _, ok := result.(*ChangeMode); !ok {
		t.Error("SetArg should return *ChangeMode, not just RunnableInterface")
	}
}

// TestChangeMode_SetID_ReturnsConcreteType verifies that SetID returns the concrete ChangeMode type.
func TestChangeMode_SetID_ReturnsConcreteType(t *testing.T) {
	skill := NewChangeMode()

	result := skill.SetID("custom-id")

	if _, ok := result.(*ChangeMode); !ok {
		t.Error("SetID should return *ChangeMode, not just RunnableInterface")
	}

	if skill.GetID() != "custom-id" {
		t.Error("SetID should set the ID")
	}
}

// TestChangeMode_SetDescription_ReturnsConcreteType verifies that SetDescription returns the concrete ChangeMode type.
func TestChangeMode_SetDescription_ReturnsConcreteType(t *testing.T) {
	skill := NewChangeMode()

	result := skill.SetDescription("custom description")

	if _, ok := result.(*ChangeMode); !ok {
		t.Error("SetDescription should return *ChangeMode, not just RunnableInterface")
	}

	if skill.GetDescription() != "custom description" {
		t.Error("SetDescription should set the description")
	}
}

// TestChangeMode_SetTimeout_ReturnsConcreteType verifies that SetTimeout returns the concrete ChangeMode type.
func TestChangeMode_SetTimeout_ReturnsConcreteType(t *testing.T) {
	skill := NewChangeMode()

	result := skill.SetTimeout(30 * time.Second)

	if _, ok := result.(*ChangeMode); !ok {
		t.Error("SetTimeout should return *ChangeMode, not just RunnableInterface")
	}
}

// TestChangeMode_MethodChaining_PreservesType verifies that method chaining preserves the concrete type.
func TestChangeMode_MethodChaining_PreservesType(t *testing.T) {
	skill := NewChangeMode().
		SetID("custom-id").
		SetDescription("custom description").
		SetArg(ArgPath, "/tmp/test").
		SetArgs(map[string]string{ArgPath: "/tmp/other"}).
		SetTimeout(30 * time.Second)

	if _, ok := skill.(*ChangeMode); !ok {
		t.Error("Method chaining should preserve *ChangeMode type")
	}

	if skill.GetID() != "custom-id" {
		t.Error("Method chaining should set ID")
	}

	if skill.GetDescription() != "custom description" {
		t.Error("Method chaining should set description")
	}
}

// TestChangeMode_SetPath verifies that SetPath sets the path arg and returns *ChangeMode.
func TestChangeMode_SetPath(t *testing.T) {
	skill := NewChangeMode().SetPath("/var/www/myapp")

	if skill.GetArg(ArgPath) != "/var/www/myapp" {
		t.Errorf("Expected path '/var/www/myapp', got '%s'", skill.GetArg(ArgPath))
	}
}

// TestChangeMode_SetMode verifies that SetMode sets the mode arg and returns *ChangeMode.
func TestChangeMode_SetMode(t *testing.T) {
	skill := NewChangeMode().SetMode("755")

	if skill.GetArg(ArgMode) != "755" {
		t.Errorf("Expected mode '755', got '%s'", skill.GetArg(ArgMode))
	}
}

// TestChangeMode_SetRecursive verifies that SetRecursive sets the recursive arg as a string bool and returns *ChangeMode.
func TestChangeMode_SetRecursive(t *testing.T) {
	skill := NewChangeMode().SetRecursive(true)

	if skill.GetArg(ArgRecursive) != "true" {
		t.Errorf("Expected recursive 'true', got '%s'", skill.GetArg(ArgRecursive))
	}

	skill.SetRecursive(false)
	if skill.GetArg(ArgRecursive) != "false" {
		t.Errorf("Expected recursive 'false', got '%s'", skill.GetArg(ArgRecursive))
	}
}

// TestChangeMode_TypedSetters_Chaining verifies that all typed setters chain correctly.
func TestChangeMode_TypedSetters_Chaining(t *testing.T) {
	skill := NewChangeMode().
		SetPath("/var/www/myapp").
		SetMode("755").
		SetRecursive(true)

	if skill.GetArg(ArgPath) != "/var/www/myapp" {
		t.Errorf("Expected path '/var/www/myapp', got '%s'", skill.GetArg(ArgPath))
	}
	if skill.GetArg(ArgMode) != "755" {
		t.Errorf("Expected mode '755', got '%s'", skill.GetArg(ArgMode))
	}
	if skill.GetArg(ArgRecursive) != "true" {
		t.Errorf("Expected recursive 'true', got '%s'", skill.GetArg(ArgRecursive))
	}
}

// --- Symbolic mode tests ---

func TestIsSymbolicMode(t *testing.T) {
	tests := []struct {
		mode string
		want bool
	}{
		{"u+x", true},
		{"g-w", true},
		{"o=rwx", true},
		{"a+rw", true},
		{"ug+rw", true},
		{"+x", true},
		{"=rw", true},
		{"u+x,g-w", true},
		{"g+rwX", true},
		{"755", false},
		{"0600", false},
		{"", false},
		{"abc", false},
	}
	for _, tt := range tests {
		t.Run(tt.mode, func(t *testing.T) {
			if got := isSymbolicMode(tt.mode); got != tt.want {
				t.Errorf("isSymbolicMode(%q) = %v, want %v", tt.mode, got, tt.want)
			}
		})
	}
}

func TestValidateSymbolicMode(t *testing.T) {
	valid := []string{
		"u+x", "g-w", "o=rwx", "a+rw", "ug+rw", "+x", "=rw",
		"u+x,g-w", "g+rwX", "go-rwx", "u=s", "a+t",
		"u+", "+", // valid chmod no-ops
	}
	for _, mode := range valid {
		t.Run("valid/"+mode, func(t *testing.T) {
			if err := validateSymbolicMode(mode); err != nil {
				t.Errorf("validateSymbolicMode(%q) returned unexpected error: %v", mode, err)
			}
		})
	}
	invalid := []string{
		"u+xyz",   // invalid permission char
		"x+rw",    // invalid who char
		"u+x,",    // trailing comma = empty clause
		",u+x",    // leading comma = empty clause
		"u+x g-w", // space instead of comma
		"u",       // no operator
	}
	for _, mode := range invalid {
		t.Run("invalid/"+mode, func(t *testing.T) {
			if err := validateSymbolicMode(mode); err == nil {
				t.Errorf("validateSymbolicMode(%q) expected error, got nil", mode)
			}
		})
	}
}

func TestValidateMode_Symbolic(t *testing.T) {
	// Symbolic modes should pass validateMode
	valid := []string{"u+x", "g+rwX", "a-w", "+x", "u+x,g-w"}
	for _, mode := range valid {
		t.Run("valid/"+mode, func(t *testing.T) {
			if err := validateMode(mode); err != nil {
				t.Errorf("validateMode(%q) returned unexpected error: %v", mode, err)
			}
		})
	}
	// Invalid symbolic modes should fail validateMode
	invalid := []string{"u+xyz", "x+rw", "u+x,"}
	for _, mode := range invalid {
		t.Run("invalid/"+mode, func(t *testing.T) {
			if err := validateMode(mode); err == nil {
				t.Errorf("validateMode(%q) expected error, got nil", mode)
			}
		})
	}
}

func TestChangeMode_Check_SymbolicMode_AlwaysTrue(t *testing.T) {
	pb := NewChangeMode()
	cfg := types.NodeConfig{
		IsDryRunMode: false,
		Logger:       slog.Default(),
		Args:         map[string]string{},
	}
	pb.SetNodeConfig(cfg)
	pb.SetArg(ArgPath, "/var/www/myapp")
	pb.SetArg(ArgMode, "g+rwX")

	needsChange, err := pb.Check()
	if err != nil {
		t.Fatalf("Check() returned unexpected error: %v", err)
	}
	if !needsChange {
		t.Error("Check() should return true for symbolic modes (can't compare against stat output)")
	}
}

func TestChangeMode_Run_DryRun_SymbolicMode(t *testing.T) {
	pb := NewChangeMode()
	cfg := types.NodeConfig{
		IsDryRunMode: true,
		Logger:       slog.Default(),
		Args:         map[string]string{},
	}
	pb.SetNodeConfig(cfg)
	pb.SetArg(ArgPath, "/var/www/myapp")
	pb.SetArg(ArgMode, "g+rwX")

	result := pb.Run()
	if !result.Changed {
		t.Error("Expected Changed to be true in dry-run mode with symbolic mode")
	}
	expectedMessage := "Would change mode to g+rwX on /var/www/myapp"
	if result.Message != expectedMessage {
		t.Errorf("Expected message '%s', got '%s'", expectedMessage, result.Message)
	}
	if result.Error != nil {
		t.Errorf("Expected no error in dry-run mode, got: %v", result.Error)
	}
}

func TestChangeMode_Run_InvalidSymbolicMode(t *testing.T) {
	pb := NewChangeMode()
	cfg := types.NodeConfig{
		IsDryRunMode: true,
		Logger:       slog.Default(),
		Args:         map[string]string{},
	}
	pb.SetNodeConfig(cfg)
	pb.SetArg(ArgPath, "/var/www/myapp")
	pb.SetArg(ArgMode, "u+xyz")

	result := pb.Run()
	if result.Changed {
		t.Error("Expected Changed to be false for invalid symbolic mode")
	}
	if result.Error == nil {
		t.Error("Expected an error for invalid symbolic mode 'u+xyz'")
	}
}
