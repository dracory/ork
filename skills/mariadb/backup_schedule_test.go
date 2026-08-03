package mariadb

import (
	"log/slog"
	"testing"

	"github.com/dracory/ork/types"
)

// TestBackupSchedule_Run_DryRun verifies that dry-run mode correctly handles
// the backup schedule install without executing SSH commands.
func TestBackupSchedule_Run_DryRun(t *testing.T) {
	pb := NewBackupSchedule()
	pb.SetArg(ArgRootPassword, "testpass123")

	cfg := types.NodeConfig{
		IsDryRunMode: true,
		Logger:       slog.Default(),
	}

	pb.SetNodeConfig(cfg)

	result := pb.Run()

	if !result.Changed {
		t.Error("Expected Changed to be true in dry-run mode")
	}

	if result.Error != nil {
		t.Errorf("Expected no error in dry-run mode, got: %v", result.Error)
	}

	// Default schedule is daily at 02:00 with 7-day retention
	expected := "Would install MariaDB backup schedule (daily '*-*-* 02:00:00', 7-day retention, /var/backups/mariadb)"
	if result.Message != expected {
		t.Errorf("Expected message '%s', got '%s'", expected, result.Message)
	}
}

// TestBackupSchedule_Run_DryRun_NoPassword verifies dry-run without password
// returns an error.
func TestBackupSchedule_Run_DryRun_NoPassword(t *testing.T) {
	pb := NewBackupSchedule()

	cfg := types.NodeConfig{
		IsDryRunMode: true,
		Logger:       slog.Default(),
	}

	pb.SetNodeConfig(cfg)

	result := pb.Run()

	if result.Error == nil {
		t.Error("Expected error for missing root-password")
	}

	if result.Message != "MariaDB root password not provided" {
		t.Errorf("Expected message 'MariaDB root password not provided', got '%s'", result.Message)
	}
}

// TestBackupSchedule_Run_DryRun_CustomArgs verifies that custom args are
// reflected in the dry-run message.
func TestBackupSchedule_Run_DryRun_CustomArgs(t *testing.T) {
	pb := NewBackupSchedule().
		SetRootPassword("s3cret").
		SetSchedule("*-*-* 03:30:00").
		SetRetentionDays("14").
		SetBackupDir("/opt/backups")

	cfg := types.NodeConfig{
		IsDryRunMode: true,
		Logger:       slog.Default(),
	}

	pb.SetNodeConfig(cfg)

	result := pb.Run()

	if !result.Changed {
		t.Error("Expected Changed to be true in dry-run mode")
	}

	expected := "Would install MariaDB backup schedule (daily '*-*-* 03:30:00', 14-day retention, /opt/backups)"
	if result.Message != expected {
		t.Errorf("Expected message '%s', got '%s'", expected, result.Message)
	}
}

// TestBackupSchedule_Run_NotDryRun verifies that non-dry-run mode does not
// return the dry-run message.
func TestBackupSchedule_Run_NotDryRun(t *testing.T) {
	pb := NewBackupSchedule()
	pb.SetArg(ArgRootPassword, "testpass123")

	cfg := types.NodeConfig{
		IsDryRunMode: false,
		Logger:       slog.Default(),
	}

	pb.SetNodeConfig(cfg)

	result := pb.Run()

	// In non-dry-run mode, it will try to execute SSH commands and likely fail
	// since there's no real SSH server. We just verify it doesn't return the
	// dry-run message.
	if result.Message == "Would install MariaDB backup schedule (daily '*-*-* 02:00:00', 7-day retention, /var/backups/mariadb)" {
		t.Error("Should not return dry-run message when IsDryRunMode is false")
	}
}

// TestBackupSchedule_NewBackupSchedule verifies that NewBackupSchedule creates
// a properly configured skill.
func TestBackupSchedule_NewBackupSchedule(t *testing.T) {
	pb := NewBackupSchedule()

	if pb.GetID() != "mariadb-backup-schedule" {
		t.Errorf("Expected ID to be 'mariadb-backup-schedule', got '%s'", pb.GetID())
	}

	expectedDescription := "Install automated daily MariaDB backup with systemd timer"
	if pb.GetDescription() != expectedDescription {
		t.Errorf("Expected description '%s', got '%s'", expectedDescription, pb.GetDescription())
	}
}

// TestBackupSchedule_SetArgs_ReturnsConcreteType verifies that SetArgs returns
// the concrete BackupSchedule type.
func TestBackupSchedule_SetArgs_ReturnsConcreteType(t *testing.T) {
	skill := NewBackupSchedule()
	args := map[string]string{ArgRootPassword: "testpass"}

	result := skill.SetArgs(args)

	if _, ok := result.(*BackupSchedule); !ok {
		t.Error("SetArgs should return *BackupSchedule, not just RunnableInterface")
	}
}

// TestBackupSchedule_SetArg_ReturnsConcreteType verifies that SetArg returns
// the concrete BackupSchedule type.
func TestBackupSchedule_SetArg_ReturnsConcreteType(t *testing.T) {
	skill := NewBackupSchedule()

	result := skill.SetArg(ArgRootPassword, "testpass")

	if _, ok := result.(*BackupSchedule); !ok {
		t.Error("SetArg should return *BackupSchedule, not just RunnableInterface")
	}
}

// TestBackupSchedule_SetID_ReturnsConcreteType verifies that SetID returns the
// concrete BackupSchedule type.
func TestBackupSchedule_SetID_ReturnsConcreteType(t *testing.T) {
	skill := NewBackupSchedule()

	result := skill.SetID("custom-id")

	if _, ok := result.(*BackupSchedule); !ok {
		t.Error("SetID should return *BackupSchedule, not just RunnableInterface")
	}

	if skill.GetID() != "custom-id" {
		t.Error("SetID should set the ID")
	}
}

// TestBackupSchedule_SetDescription_ReturnsConcreteType verifies that
// SetDescription returns the concrete BackupSchedule type.
func TestBackupSchedule_SetDescription_ReturnsConcreteType(t *testing.T) {
	skill := NewBackupSchedule()

	result := skill.SetDescription("custom description")

	if _, ok := result.(*BackupSchedule); !ok {
		t.Error("SetDescription should return *BackupSchedule, not just RunnableInterface")
	}

	if skill.GetDescription() != "custom description" {
		t.Error("SetDescription should set the description")
	}
}

// TestBackupSchedule_SetTimeout_ReturnsConcreteType verifies that SetTimeout
// returns the concrete BackupSchedule type.
func TestBackupSchedule_SetTimeout_ReturnsConcreteType(t *testing.T) {
	skill := NewBackupSchedule()

	result := skill.SetTimeout(30 * 1000000000)

	if _, ok := result.(*BackupSchedule); !ok {
		t.Error("SetTimeout should return *BackupSchedule, not just RunnableInterface")
	}
}

// TestBackupSchedule_MethodChaining_PreservesType verifies that method
// chaining preserves the concrete type.
func TestBackupSchedule_MethodChaining_PreservesType(t *testing.T) {
	skill := NewBackupSchedule().
		SetID("custom-id").
		SetDescription("custom description").
		SetArg(ArgRootPassword, "testpass").
		SetArgs(map[string]string{ArgPort: "3307"}).
		SetTimeout(30 * 1000000000)

	if _, ok := skill.(*BackupSchedule); !ok {
		t.Error("Method chaining should preserve *BackupSchedule type")
	}

	if skill.GetID() != "custom-id" {
		t.Error("Method chaining should set ID")
	}

	if skill.GetDescription() != "custom description" {
		t.Error("Method chaining should set description")
	}
}

// TestBackupSchedule_SetRootPassword verifies that SetRootPassword sets the
// root-password arg.
func TestBackupSchedule_SetRootPassword(t *testing.T) {
	skill := NewBackupSchedule().SetRootPassword("s3cret")

	if skill.GetArg(ArgRootPassword) != "s3cret" {
		t.Errorf("Expected root-password 's3cret', got '%s'", skill.GetArg(ArgRootPassword))
	}
}

// TestBackupSchedule_SetPort verifies that SetPort sets the port arg.
func TestBackupSchedule_SetPort(t *testing.T) {
	skill := NewBackupSchedule().SetPort("3307")

	if skill.GetArg(ArgPort) != "3307" {
		t.Errorf("Expected port '3307', got '%s'", skill.GetArg(ArgPort))
	}
}

// TestBackupSchedule_SetBackupDir verifies that SetBackupDir sets the
// backup-dir arg.
func TestBackupSchedule_SetBackupDir(t *testing.T) {
	skill := NewBackupSchedule().SetBackupDir("/var/backups/db")

	if skill.GetArg(ArgBackupDir) != "/var/backups/db" {
		t.Errorf("Expected backup-dir '/var/backups/db', got '%s'", skill.GetArg(ArgBackupDir))
	}
}

// TestBackupSchedule_SetSchedule verifies that SetSchedule sets the schedule arg.
func TestBackupSchedule_SetSchedule(t *testing.T) {
	skill := NewBackupSchedule().SetSchedule("*-*-* 03:00:00")

	if skill.GetArg(ArgSchedule) != "*-*-* 03:00:00" {
		t.Errorf("Expected schedule '*-*-* 03:00:00', got '%s'", skill.GetArg(ArgSchedule))
	}
}

// TestBackupSchedule_SetRetentionDays verifies that SetRetentionDays sets the
// retention-days arg.
func TestBackupSchedule_SetRetentionDays(t *testing.T) {
	skill := NewBackupSchedule().SetRetentionDays("14")

	if skill.GetArg(ArgRetentionDays) != "14" {
		t.Errorf("Expected retention-days '14', got '%s'", skill.GetArg(ArgRetentionDays))
	}
}

// TestBackupSchedule_SetScriptPath verifies that SetScriptPath sets the
// script-path arg.
func TestBackupSchedule_SetScriptPath(t *testing.T) {
	skill := NewBackupSchedule().SetScriptPath("/usr/local/bin/backup.sh")

	if skill.GetArg(ArgScriptPath) != "/usr/local/bin/backup.sh" {
		t.Errorf("Expected script-path '/usr/local/bin/backup.sh', got '%s'", skill.GetArg(ArgScriptPath))
	}
}

// TestBackupSchedule_SetServiceName verifies that SetServiceName sets the
// service-name arg.
func TestBackupSchedule_SetServiceName(t *testing.T) {
	skill := NewBackupSchedule().SetServiceName("custom-backup")

	if skill.GetArg(ArgServiceName) != "custom-backup" {
		t.Errorf("Expected service-name 'custom-backup', got '%s'", skill.GetArg(ArgServiceName))
	}
}

// TestBackupSchedule_TypedSetters_Chaining verifies that all typed setters
// chain correctly.
func TestBackupSchedule_TypedSetters_Chaining(t *testing.T) {
	skill := NewBackupSchedule().
		SetRootPassword("s3cret").
		SetPort("3307").
		SetBackupDir("/var/backups/db").
		SetSchedule("*-*-* 03:00:00").
		SetRetentionDays("14").
		SetScriptPath("/usr/local/bin/backup.sh").
		SetServiceName("custom-backup")

	if skill.GetArg(ArgRootPassword) != "s3cret" {
		t.Errorf("Expected root-password 's3cret', got '%s'", skill.GetArg(ArgRootPassword))
	}
	if skill.GetArg(ArgPort) != "3307" {
		t.Errorf("Expected port '3307', got '%s'", skill.GetArg(ArgPort))
	}
	if skill.GetArg(ArgBackupDir) != "/var/backups/db" {
		t.Errorf("Expected backup-dir '/var/backups/db', got '%s'", skill.GetArg(ArgBackupDir))
	}
	if skill.GetArg(ArgSchedule) != "*-*-* 03:00:00" {
		t.Errorf("Expected schedule '*-*-* 03:00:00', got '%s'", skill.GetArg(ArgSchedule))
	}
	if skill.GetArg(ArgRetentionDays) != "14" {
		t.Errorf("Expected retention-days '14', got '%s'", skill.GetArg(ArgRetentionDays))
	}
	if skill.GetArg(ArgScriptPath) != "/usr/local/bin/backup.sh" {
		t.Errorf("Expected script-path '/usr/local/bin/backup.sh', got '%s'", skill.GetArg(ArgScriptPath))
	}
	if skill.GetArg(ArgServiceName) != "custom-backup" {
		t.Errorf("Expected service-name 'custom-backup', got '%s'", skill.GetArg(ArgServiceName))
	}
}

// TestBackupSchedule_Run_DryRun_InvalidPort verifies that a non-numeric port
// is rejected before any commands run.
func TestBackupSchedule_Run_DryRun_InvalidPort(t *testing.T) {
	pb := NewBackupSchedule().
		SetRootPassword("testpass").
		SetPort("abc")

	cfg := types.NodeConfig{IsDryRunMode: true, Logger: slog.Default()}
	pb.SetNodeConfig(cfg)

	result := pb.Run()

	if result.Error == nil {
		t.Error("Expected error for non-numeric port")
	}
	if result.Message != "Invalid port: must be numeric" {
		t.Errorf("Expected 'Invalid port: must be numeric', got '%s'", result.Message)
	}
}

// TestBackupSchedule_Run_DryRun_InvalidRetentionDays verifies that a
// non-numeric retention-days is rejected.
func TestBackupSchedule_Run_DryRun_InvalidRetentionDays(t *testing.T) {
	pb := NewBackupSchedule().
		SetRootPassword("testpass").
		SetRetentionDays("abc")

	cfg := types.NodeConfig{IsDryRunMode: true, Logger: slog.Default()}
	pb.SetNodeConfig(cfg)

	result := pb.Run()

	if result.Error == nil {
		t.Error("Expected error for non-numeric retention-days")
	}
	if result.Message != "Invalid retention-days: must be numeric" {
		t.Errorf("Expected 'Invalid retention-days: must be numeric', got '%s'", result.Message)
	}
}

// TestBackupSchedule_Run_DryRun_InvalidServiceName verifies that a service-name
// with path traversal characters is rejected.
func TestBackupSchedule_Run_DryRun_InvalidServiceName(t *testing.T) {
	pb := NewBackupSchedule().
		SetRootPassword("testpass").
		SetServiceName("../evil")

	cfg := types.NodeConfig{IsDryRunMode: true, Logger: slog.Default()}
	pb.SetNodeConfig(cfg)

	result := pb.Run()

	if result.Error == nil {
		t.Error("Expected error for invalid service-name")
	}
	if result.Message != "Invalid service-name: only alphanumeric, hyphens, and underscores allowed" {
		t.Errorf("Expected invalid service-name message, got '%s'", result.Message)
	}
}

// TestBackupSchedule_Run_DryRun_InvalidSchedule verifies that a schedule with
// newlines (injection attempt) is rejected.
func TestBackupSchedule_Run_DryRun_InvalidSchedule(t *testing.T) {
	pb := NewBackupSchedule().
		SetRootPassword("testpass").
		SetSchedule("daily\n[Timer]\nOnCalendar=always")

	cfg := types.NodeConfig{IsDryRunMode: true, Logger: slog.Default()}
	pb.SetNodeConfig(cfg)

	result := pb.Run()

	if result.Error == nil {
		t.Error("Expected error for invalid schedule")
	}
	if result.Message != "Invalid schedule: contains illegal characters" {
		t.Errorf("Expected invalid schedule message, got '%s'", result.Message)
	}
}

// TestBackupSchedule_Run_DryRun_InvalidBackupDir verifies that a backup-dir
// with shell-dangerous characters is rejected.
func TestBackupSchedule_Run_DryRun_InvalidBackupDir(t *testing.T) {
	pb := NewBackupSchedule().
		SetRootPassword("testpass").
		SetBackupDir("/tmp/$(whoami)")

	cfg := types.NodeConfig{IsDryRunMode: true, Logger: slog.Default()}
	pb.SetNodeConfig(cfg)

	result := pb.Run()

	if result.Error == nil {
		t.Error("Expected error for invalid backup-dir")
	}
	if result.Message != "Invalid backup-dir: contains illegal characters" {
		t.Errorf("Expected invalid backup-dir message, got '%s'", result.Message)
	}
}

// TestBackupSchedule_Run_DryRun_InvalidScriptPath verifies that a script-path
// with shell-dangerous characters is rejected.
func TestBackupSchedule_Run_DryRun_InvalidScriptPath(t *testing.T) {
	pb := NewBackupSchedule().
		SetRootPassword("testpass").
		SetScriptPath("/tmp/$(whoami)/script.sh")

	cfg := types.NodeConfig{IsDryRunMode: true, Logger: slog.Default()}
	pb.SetNodeConfig(cfg)

	result := pb.Run()

	if result.Error == nil {
		t.Error("Expected error for invalid script-path")
	}
	if result.Message != "Invalid script-path: contains illegal characters" {
		t.Errorf("Expected invalid script-path message, got '%s'", result.Message)
	}
}
