package fs

import (
	"log/slog"
	"testing"
	"time"

	"github.com/dracory/ork/types"
)

// TestDirCopy_Run_DryRun verifies that dry-run mode reports the would-copy message.
func TestDirCopy_Run_DryRun(t *testing.T) {
	pb := NewDirCopy()

	cfg := types.NodeConfig{
		IsDryRunMode: true,
		Logger:       slog.Default(),
		Args:         map[string]string{},
	}

	pb.SetNodeConfig(cfg)
	pb.SetArg(ArgSrc, "/var/www/myapp")
	pb.SetArg(ArgDst, "/var/www/myapp.bak")
	pb.SetArg(ArgForce, "true")

	result := pb.Run()

	if !result.Changed {
		t.Error("Expected Changed to be true in dry-run mode")
	}

	expectedMessage := "Would copy directory: /var/www/myapp -> /var/www/myapp.bak"
	if result.Message != expectedMessage {
		t.Errorf("Expected message '%s', got '%s'", expectedMessage, result.Message)
	}

	if result.Error != nil {
		t.Errorf("Expected no error in dry-run mode, got: %v", result.Error)
	}
}

// TestDirCopy_Run_NoSrc verifies that missing ArgSrc returns an error.
func TestDirCopy_Run_NoSrc(t *testing.T) {
	pb := NewDirCopy()

	cfg := types.NodeConfig{
		IsDryRunMode: false,
		Logger:       slog.Default(),
		Args:         map[string]string{},
	}

	pb.SetNodeConfig(cfg)
	pb.SetArg(ArgDst, "/var/www/myapp.bak")

	result := pb.Run()

	if result.Changed {
		t.Error("Expected Changed to be false when no src specified")
	}

	if result.Error == nil {
		t.Error("Expected an error when no src is specified")
	}
}

// TestDirCopy_Run_NoDst verifies that missing ArgDst returns an error.
func TestDirCopy_Run_NoDst(t *testing.T) {
	pb := NewDirCopy()

	cfg := types.NodeConfig{
		IsDryRunMode: false,
		Logger:       slog.Default(),
		Args:         map[string]string{},
	}

	pb.SetNodeConfig(cfg)
	pb.SetArg(ArgSrc, "/var/www/myapp")

	result := pb.Run()

	if result.Changed {
		t.Error("Expected Changed to be false when no dst specified")
	}

	if result.Error == nil {
		t.Error("Expected an error when no dst is specified")
	}
}

// TestDirCopy_Run_RelativeSrc verifies that a relative src returns an error.
func TestDirCopy_Run_RelativeSrc(t *testing.T) {
	pb := NewDirCopy()

	cfg := types.NodeConfig{
		IsDryRunMode: true,
		Logger:       slog.Default(),
		Args:         map[string]string{},
	}

	pb.SetNodeConfig(cfg)
	pb.SetArg(ArgSrc, "relative/path")
	pb.SetArg(ArgDst, "/var/www/myapp.bak")

	result := pb.Run()

	if result.Changed {
		t.Error("Expected Changed to be false for relative src")
	}

	if result.Error == nil {
		t.Error("Expected an error for relative src")
	}
}

// TestDirCopy_Run_RelativeDst verifies that a relative dst returns an error.
func TestDirCopy_Run_RelativeDst(t *testing.T) {
	pb := NewDirCopy()

	cfg := types.NodeConfig{
		IsDryRunMode: true,
		Logger:       slog.Default(),
		Args:         map[string]string{},
	}

	pb.SetNodeConfig(cfg)
	pb.SetArg(ArgSrc, "/var/www/myapp")
	pb.SetArg(ArgDst, "relative/path")

	result := pb.Run()

	if result.Changed {
		t.Error("Expected Changed to be false for relative dst")
	}

	if result.Error == nil {
		t.Error("Expected an error for relative dst")
	}
}

// TestDirCopy_Run_NotDryRun verifies that non-dry-run mode does not return the dry-run message.
func TestDirCopy_Run_NotDryRun(t *testing.T) {
	pb := NewDirCopy()

	cfg := types.NodeConfig{
		IsDryRunMode: false,
		Logger:       slog.Default(),
		Args:         map[string]string{},
	}

	pb.SetNodeConfig(cfg)
	pb.SetArg(ArgSrc, "/var/www/myapp")
	pb.SetArg(ArgDst, "/var/www/myapp.bak")

	result := pb.Run()

	if result.Message == "Would copy directory: /var/www/myapp -> /var/www/myapp.bak" {
		t.Error("Should not return dry-run message when IsDryRunMode is false")
	}
}

// TestDirCopy_NewDirCopy verifies that NewDirCopy creates a properly configured skill.
func TestDirCopy_NewDirCopy(t *testing.T) {
	pb := NewDirCopy()

	if pb.GetID() != "fs-dir-copy" {
		t.Errorf("Expected ID to be 'fs-dir-copy', got '%s'", pb.GetID())
	}

	expectedDescription := "Copy directory recursively on remote server (cp -Rp)"
	if pb.GetDescription() != expectedDescription {
		t.Errorf("Expected description '%s', got '%s'", expectedDescription, pb.GetDescription())
	}
}

// TestDirCopy_SetArgs_ReturnsConcreteType verifies that SetArgs returns the concrete DirCopy type.
func TestDirCopy_SetArgs_ReturnsConcreteType(t *testing.T) {
	skill := NewDirCopy()
	args := map[string]string{ArgSrc: "/tmp/test"}

	result := skill.SetArgs(args)

	if _, ok := result.(*DirCopy); !ok {
		t.Error("SetArgs should return *DirCopy, not just RunnableInterface")
	}
}

// TestDirCopy_SetArg_ReturnsConcreteType verifies that SetArg returns the concrete DirCopy type.
func TestDirCopy_SetArg_ReturnsConcreteType(t *testing.T) {
	skill := NewDirCopy()

	result := skill.SetArg(ArgSrc, "/tmp/test")

	if _, ok := result.(*DirCopy); !ok {
		t.Error("SetArg should return *DirCopy, not just RunnableInterface")
	}
}

// TestDirCopy_SetID_ReturnsConcreteType verifies that SetID returns the concrete DirCopy type.
func TestDirCopy_SetID_ReturnsConcreteType(t *testing.T) {
	skill := NewDirCopy()

	result := skill.SetID("custom-id")

	if _, ok := result.(*DirCopy); !ok {
		t.Error("SetID should return *DirCopy, not just RunnableInterface")
	}

	if skill.GetID() != "custom-id" {
		t.Error("SetID should set the ID")
	}
}

// TestDirCopy_SetDescription_ReturnsConcreteType verifies that SetDescription returns the concrete DirCopy type.
func TestDirCopy_SetDescription_ReturnsConcreteType(t *testing.T) {
	skill := NewDirCopy()

	result := skill.SetDescription("custom description")

	if _, ok := result.(*DirCopy); !ok {
		t.Error("SetDescription should return *DirCopy, not just RunnableInterface")
	}

	if skill.GetDescription() != "custom description" {
		t.Error("SetDescription should set the description")
	}
}

// TestDirCopy_SetTimeout_ReturnsConcreteType verifies that SetTimeout returns the concrete DirCopy type.
func TestDirCopy_SetTimeout_ReturnsConcreteType(t *testing.T) {
	skill := NewDirCopy()

	result := skill.SetTimeout(30 * time.Second)

	if _, ok := result.(*DirCopy); !ok {
		t.Error("SetTimeout should return *DirCopy, not just RunnableInterface")
	}
}

// TestDirCopy_MethodChaining_PreservesType verifies that method chaining preserves the concrete type.
func TestDirCopy_MethodChaining_PreservesType(t *testing.T) {
	skill := NewDirCopy().
		SetID("custom-id").
		SetDescription("custom description").
		SetArg(ArgSrc, "/tmp/test").
		SetArgs(map[string]string{ArgSrc: "/tmp/other"}).
		SetTimeout(30 * time.Second)

	if _, ok := skill.(*DirCopy); !ok {
		t.Error("Method chaining should preserve *DirCopy type")
	}

	if skill.GetID() != "custom-id" {
		t.Error("Method chaining should set ID")
	}

	if skill.GetDescription() != "custom description" {
		t.Error("Method chaining should set description")
	}
}

// TestDirCopy_SetSrc verifies that SetSrc sets the src arg and returns *DirCopy.
func TestDirCopy_SetSrc(t *testing.T) {
	skill := NewDirCopy().SetSrc("/var/www/myapp")

	if skill.GetArg(ArgSrc) != "/var/www/myapp" {
		t.Errorf("Expected src '/var/www/myapp', got '%s'", skill.GetArg(ArgSrc))
	}
}

// TestDirCopy_SetDst verifies that SetDst sets the dst arg and returns *DirCopy.
func TestDirCopy_SetDst(t *testing.T) {
	skill := NewDirCopy().SetDst("/var/www/myapp.bak")

	if skill.GetArg(ArgDst) != "/var/www/myapp.bak" {
		t.Errorf("Expected dst '/var/www/myapp.bak', got '%s'", skill.GetArg(ArgDst))
	}
}

// TestDirCopy_SetForce verifies that SetForce sets the force arg and returns *DirCopy.
func TestDirCopy_SetForce(t *testing.T) {
	skill := NewDirCopy().SetForce(true)

	if skill.GetArg(ArgForce) != "true" {
		t.Errorf("Expected force 'true', got '%s'", skill.GetArg(ArgForce))
	}

	skill.SetForce(false)
	if skill.GetArg(ArgForce) != "false" {
		t.Errorf("Expected force 'false', got '%s'", skill.GetArg(ArgForce))
	}
}

// TestDirCopy_TypedSetters_Chaining verifies that all typed setters chain correctly.
func TestDirCopy_TypedSetters_Chaining(t *testing.T) {
	skill := NewDirCopy().
		SetSrc("/var/www/myapp").
		SetDst("/var/www/myapp.bak").
		SetForce(true)

	if skill.GetArg(ArgSrc) != "/var/www/myapp" {
		t.Errorf("Expected src '/var/www/myapp', got '%s'", skill.GetArg(ArgSrc))
	}
	if skill.GetArg(ArgDst) != "/var/www/myapp.bak" {
		t.Errorf("Expected dst '/var/www/myapp.bak', got '%s'", skill.GetArg(ArgDst))
	}
	if skill.GetArg(ArgForce) != "true" {
		t.Errorf("Expected force 'true', got '%s'", skill.GetArg(ArgForce))
	}
}

// TestDirCopy_WithNodeConfig verifies that WithNodeConfig sets the config and returns *DirCopy.
func TestDirCopy_WithNodeConfig(t *testing.T) {
	cfg := types.NodeConfig{
		IsDryRunMode: true,
		Logger:       slog.Default(),
		Args:         map[string]string{},
	}

	skill := NewDirCopy().WithNodeConfig(cfg)

	// WithNodeConfig returns *DirCopy directly, so just verify it works
	if skill.GetArg(ArgSrc) != "" {
		t.Error("WithNodeConfig should return a usable *DirCopy")
	}
}

// TestDirCopy_Run_DestructivePath_RejectsRoot verifies that rm -rf on "/" is rejected.
func TestDirCopy_Run_DestructivePath_RejectsRoot(t *testing.T) {
	pb := NewDirCopy()

	cfg := types.NodeConfig{
		IsDryRunMode: true,
		Logger:       slog.Default(),
		Args:         map[string]string{},
	}

	pb.SetNodeConfig(cfg)
	pb.SetArg(ArgSrc, "/var/www/myapp")
	pb.SetArg(ArgDst, "/")
	pb.SetArg(ArgForce, "true")

	result := pb.Run()

	if result.Changed {
		t.Error("Expected Changed to be false for destructive dst path '/'")
	}

	if result.Error == nil {
		t.Error("Expected an error for destructive dst path '/'")
	}
}

// TestDirCopy_Run_DestructivePath_RejectsSingleComponent verifies that rm -rf on "/var" is rejected.
func TestDirCopy_Run_DestructivePath_RejectsSingleComponent(t *testing.T) {
	pb := NewDirCopy()

	cfg := types.NodeConfig{
		IsDryRunMode: true,
		Logger:       slog.Default(),
		Args:         map[string]string{},
	}

	pb.SetNodeConfig(cfg)
	pb.SetArg(ArgSrc, "/var/www/myapp")
	pb.SetArg(ArgDst, "/var")
	pb.SetArg(ArgForce, "true")

	result := pb.Run()

	if result.Changed {
		t.Error("Expected Changed to be false for destructive dst path '/var'")
	}

	if result.Error == nil {
		t.Error("Expected an error for destructive dst path '/var'")
	}
}

// TestDirCopy_Run_DestructivePath_AllowsTwoComponents verifies that "/var/www" is allowed.
func TestDirCopy_Run_DestructivePath_AllowsTwoComponents(t *testing.T) {
	pb := NewDirCopy()

	cfg := types.NodeConfig{
		IsDryRunMode: true,
		Logger:       slog.Default(),
		Args:         map[string]string{},
	}

	pb.SetNodeConfig(cfg)
	pb.SetArg(ArgSrc, "/var/www/myapp")
	pb.SetArg(ArgDst, "/var/www")
	pb.SetArg(ArgForce, "true")

	result := pb.Run()

	// In dry-run mode with valid paths, should proceed (Changed=true)
	if !result.Changed {
		t.Error("Expected Changed to be true for valid dst path '/var/www'")
	}

	if result.Error != nil {
		t.Errorf("Expected no error for valid dst path, got: %v", result.Error)
	}
}

// TestDirCopy_Run_DotDotTraversal_RejectsSrc verifies that '..' in src is rejected.
func TestDirCopy_Run_DotDotTraversal_RejectsSrc(t *testing.T) {
	pb := NewDirCopy()

	cfg := types.NodeConfig{
		IsDryRunMode: true,
		Logger:       slog.Default(),
		Args:         map[string]string{},
	}

	pb.SetNodeConfig(cfg)
	pb.SetArg(ArgSrc, "/var/../etc")
	pb.SetArg(ArgDst, "/var/www/myapp.bak")

	result := pb.Run()

	if result.Changed {
		t.Error("Expected Changed to be false for '..' traversal in src")
	}

	if result.Error == nil {
		t.Error("Expected an error for '..' traversal in src")
	}
}

// TestDirCopy_Run_DotDotTraversal_RejectsDst verifies that '..' in dst is rejected.
func TestDirCopy_Run_DotDotTraversal_RejectsDst(t *testing.T) {
	pb := NewDirCopy()

	cfg := types.NodeConfig{
		IsDryRunMode: true,
		Logger:       slog.Default(),
		Args:         map[string]string{},
	}

	pb.SetNodeConfig(cfg)
	pb.SetArg(ArgSrc, "/var/www/myapp")
	pb.SetArg(ArgDst, "/var/../etc")
	pb.SetArg(ArgForce, "true")

	result := pb.Run()

	if result.Changed {
		t.Error("Expected Changed to be false for '..' traversal in dst")
	}

	if result.Error == nil {
		t.Error("Expected an error for '..' traversal in dst")
	}
}

// TestDirCopy_Run_DotDotTraversal_RejectsRootEscape verifies that '/../..' is rejected.
func TestDirCopy_Run_DotDotTraversal_RejectsRootEscape(t *testing.T) {
	pb := NewDirCopy()

	cfg := types.NodeConfig{
		IsDryRunMode: true,
		Logger:       slog.Default(),
		Args:         map[string]string{},
	}

	pb.SetNodeConfig(cfg)
	pb.SetArg(ArgSrc, "/var/www/myapp")
	pb.SetArg(ArgDst, "/../..")
	pb.SetArg(ArgForce, "true")

	result := pb.Run()

	if result.Changed {
		t.Error("Expected Changed to be false for '/../..' dst path")
	}

	if result.Error == nil {
		t.Error("Expected an error for '/../..' dst path")
	}
}
