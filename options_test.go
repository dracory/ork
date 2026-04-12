package ork

import (
	"testing"
	"time"
)

func TestDefaultOptions(t *testing.T) {
	opts := defaultOptions()

	if opts.port != "22" {
		t.Errorf("expected port '22', got '%s'", opts.port)
	}
	if opts.user != "root" {
		t.Errorf("expected user 'root', got '%s'", opts.user)
	}
	if opts.key != "id_rsa" {
		t.Errorf("expected key 'id_rsa', got '%s'", opts.key)
	}
	if opts.timeout != 30*time.Second {
		t.Errorf("expected timeout 30s, got %v", opts.timeout)
	}
	if opts.args == nil {
		t.Error("expected args map to be initialized")
	}
	if len(opts.args) != 0 {
		t.Errorf("expected empty args map, got %d entries", len(opts.args))
	}
	if opts.dryRun != false {
		t.Error("expected dryRun to be false")
	}
}

func TestWithPort(t *testing.T) {
	opts := defaultOptions()
	opt := WithPort("2222")
	opt(opts)

	if opts.port != "2222" {
		t.Errorf("expected port '2222', got '%s'", opts.port)
	}
}

func TestWithUser(t *testing.T) {
	opts := defaultOptions()
	opt := WithUser("deploy")
	opt(opts)

	if opts.user != "deploy" {
		t.Errorf("expected user 'deploy', got '%s'", opts.user)
	}
}

func TestWithKey(t *testing.T) {
	opts := defaultOptions()
	opt := WithKey("production.prv")
	opt(opts)

	if opts.key != "production.prv" {
		t.Errorf("expected key 'production.prv', got '%s'", opts.key)
	}
}

func TestWithArg(t *testing.T) {
	opts := defaultOptions()
	opt1 := WithArg("username", "alice")
	opt2 := WithArg("shell", "/bin/bash")
	opt1(opts)
	opt2(opts)

	if opts.args["username"] != "alice" {
		t.Errorf("expected args['username'] = 'alice', got '%s'", opts.args["username"])
	}
	if opts.args["shell"] != "/bin/bash" {
		t.Errorf("expected args['shell'] = '/bin/bash', got '%s'", opts.args["shell"])
	}
	if len(opts.args) != 2 {
		t.Errorf("expected 2 args, got %d", len(opts.args))
	}
}

func TestWithArg_NilMap(t *testing.T) {
	opts := &options{args: nil}
	opt := WithArg("key", "value")
	opt(opts)

	if opts.args == nil {
		t.Fatal("expected args map to be initialized")
	}
	if opts.args["key"] != "value" {
		t.Errorf("expected args['key'] = 'value', got '%s'", opts.args["key"])
	}
}

func TestWithArgs(t *testing.T) {
	opts := defaultOptions()
	args := map[string]string{
		"username": "alice",
		"shell":    "/bin/bash",
		"home":     "/home/alice",
	}
	opt := WithArgs(args)
	opt(opts)

	if opts.args["username"] != "alice" {
		t.Errorf("expected args['username'] = 'alice', got '%s'", opts.args["username"])
	}
	if opts.args["shell"] != "/bin/bash" {
		t.Errorf("expected args['shell'] = '/bin/bash', got '%s'", opts.args["shell"])
	}
	if opts.args["home"] != "/home/alice" {
		t.Errorf("expected args['home'] = '/home/alice', got '%s'", opts.args["home"])
	}
	if len(opts.args) != 3 {
		t.Errorf("expected 3 args, got %d", len(opts.args))
	}
}

func TestWithArgs_Merge(t *testing.T) {
	opts := defaultOptions()
	// Add initial arg
	WithArg("existing", "value1")(opts)

	// Merge new args
	args := map[string]string{
		"username": "alice",
		"existing": "value2", // Should override
	}
	WithArgs(args)(opts)

	if opts.args["existing"] != "value2" {
		t.Errorf("expected args['existing'] = 'value2' (overridden), got '%s'", opts.args["existing"])
	}
	if opts.args["username"] != "alice" {
		t.Errorf("expected args['username'] = 'alice', got '%s'", opts.args["username"])
	}
	if len(opts.args) != 2 {
		t.Errorf("expected 2 args, got %d", len(opts.args))
	}
}

func TestWithArgs_NilMap(t *testing.T) {
	opts := &options{args: nil}
	args := map[string]string{"key": "value"}
	opt := WithArgs(args)
	opt(opts)

	if opts.args == nil {
		t.Fatal("expected args map to be initialized")
	}
	if opts.args["key"] != "value" {
		t.Errorf("expected args['key'] = 'value', got '%s'", opts.args["key"])
	}
}

func TestWithDryRun(t *testing.T) {
	opts := defaultOptions()

	// Enable dry-run
	WithDryRun(true)(opts)
	if opts.dryRun != true {
		t.Error("expected dryRun to be true")
	}

	// Disable dry-run
	WithDryRun(false)(opts)
	if opts.dryRun != false {
		t.Error("expected dryRun to be false")
	}
}

func TestWithTimeout(t *testing.T) {
	opts := defaultOptions()
	opt := WithTimeout(5 * time.Minute)
	opt(opts)

	if opts.timeout != 5*time.Minute {
		t.Errorf("expected timeout 5m, got %v", opts.timeout)
	}
}

func TestOptionOrdering(t *testing.T) {
	opts := defaultOptions()

	// Apply options in order - last one wins
	WithPort("2222")(opts)
	WithPort("3333")(opts)

	if opts.port != "3333" {
		t.Errorf("expected port '3333' (last wins), got '%s'", opts.port)
	}
}

func TestMultipleOptions(t *testing.T) {
	opts := defaultOptions()

	// Apply multiple different options
	WithPort("2222")(opts)
	WithUser("deploy")(opts)
	WithKey("prod.prv")(opts)
	WithArg("env", "production")(opts)
	WithTimeout(2 * time.Minute)(opts)
	WithDryRun(true)(opts)

	if opts.port != "2222" {
		t.Errorf("expected port '2222', got '%s'", opts.port)
	}
	if opts.user != "deploy" {
		t.Errorf("expected user 'deploy', got '%s'", opts.user)
	}
	if opts.key != "prod.prv" {
		t.Errorf("expected key 'prod.prv', got '%s'", opts.key)
	}
	if opts.args["env"] != "production" {
		t.Errorf("expected args['env'] = 'production', got '%s'", opts.args["env"])
	}
	if opts.timeout != 2*time.Minute {
		t.Errorf("expected timeout 2m, got %v", opts.timeout)
	}
	if opts.dryRun != true {
		t.Error("expected dryRun to be true")
	}
}

func TestApplyOptions_Defaults(t *testing.T) {
	cfg := applyOptions("testhost.example.com")

	if cfg.SSHHost != "testhost.example.com" {
		t.Errorf("expected SSHHost 'testhost.example.com', got '%s'", cfg.SSHHost)
	}
	if cfg.SSHPort != "22" {
		t.Errorf("expected SSHPort '22', got '%s'", cfg.SSHPort)
	}
	if cfg.RootUser != "root" {
		t.Errorf("expected RootUser 'root', got '%s'", cfg.RootUser)
	}
	if cfg.SSHKey != "id_rsa" {
		t.Errorf("expected SSHKey 'id_rsa', got '%s'", cfg.SSHKey)
	}
	if cfg.Args == nil {
		t.Error("expected Args map to be initialized")
	}
	if len(cfg.Args) != 0 {
		t.Errorf("expected empty Args map, got %d entries", len(cfg.Args))
	}
}

func TestApplyOptions_WithPort(t *testing.T) {
	cfg := applyOptions("testhost.example.com", WithPort("2222"))

	if cfg.SSHHost != "testhost.example.com" {
		t.Errorf("expected SSHHost 'testhost.example.com', got '%s'", cfg.SSHHost)
	}
	if cfg.SSHPort != "2222" {
		t.Errorf("expected SSHPort '2222', got '%s'", cfg.SSHPort)
	}
}

func TestApplyOptions_WithUser(t *testing.T) {
	cfg := applyOptions("testhost.example.com", WithUser("deploy"))

	if cfg.RootUser != "deploy" {
		t.Errorf("expected RootUser 'deploy', got '%s'", cfg.RootUser)
	}
}

func TestApplyOptions_WithKey(t *testing.T) {
	cfg := applyOptions("testhost.example.com", WithKey("production.prv"))

	if cfg.SSHKey != "production.prv" {
		t.Errorf("expected SSHKey 'production.prv', got '%s'", cfg.SSHKey)
	}
}

func TestApplyOptions_WithArg(t *testing.T) {
	cfg := applyOptions("testhost.example.com",
		WithArg("username", "alice"),
		WithArg("shell", "/bin/bash"),
	)

	if cfg.Args["username"] != "alice" {
		t.Errorf("expected Args['username'] = 'alice', got '%s'", cfg.Args["username"])
	}
	if cfg.Args["shell"] != "/bin/bash" {
		t.Errorf("expected Args['shell'] = '/bin/bash', got '%s'", cfg.Args["shell"])
	}
	if len(cfg.Args) != 2 {
		t.Errorf("expected 2 args, got %d", len(cfg.Args))
	}
}

func TestApplyOptions_WithArgs(t *testing.T) {
	args := map[string]string{
		"username": "alice",
		"shell":    "/bin/bash",
		"home":     "/home/alice",
	}
	cfg := applyOptions("testhost.example.com", WithArgs(args))

	if cfg.Args["username"] != "alice" {
		t.Errorf("expected Args['username'] = 'alice', got '%s'", cfg.Args["username"])
	}
	if cfg.Args["shell"] != "/bin/bash" {
		t.Errorf("expected Args['shell'] = '/bin/bash', got '%s'", cfg.Args["shell"])
	}
	if cfg.Args["home"] != "/home/alice" {
		t.Errorf("expected Args['home'] = '/home/alice', got '%s'", cfg.Args["home"])
	}
	if len(cfg.Args) != 3 {
		t.Errorf("expected 3 args, got %d", len(cfg.Args))
	}
}

func TestApplyOptions_MultipleOptions(t *testing.T) {
	cfg := applyOptions("testhost.example.com",
		WithPort("2222"),
		WithUser("deploy"),
		WithKey("prod.prv"),
		WithArg("env", "production"),
		WithTimeout(2*time.Minute),
		WithDryRun(true),
	)

	if cfg.SSHHost != "testhost.example.com" {
		t.Errorf("expected SSHHost 'testhost.example.com', got '%s'", cfg.SSHHost)
	}
	if cfg.SSHPort != "2222" {
		t.Errorf("expected SSHPort '2222', got '%s'", cfg.SSHPort)
	}
	if cfg.RootUser != "deploy" {
		t.Errorf("expected RootUser 'deploy', got '%s'", cfg.RootUser)
	}
	if cfg.SSHKey != "prod.prv" {
		t.Errorf("expected SSHKey 'prod.prv', got '%s'", cfg.SSHKey)
	}
	if cfg.Args["env"] != "production" {
		t.Errorf("expected Args['env'] = 'production', got '%s'", cfg.Args["env"])
	}
}

func TestApplyOptions_LastWins(t *testing.T) {
	cfg := applyOptions("testhost.example.com",
		WithPort("2222"),
		WithPort("3333"),
		WithUser("user1"),
		WithUser("user2"),
	)

	if cfg.SSHPort != "3333" {
		t.Errorf("expected SSHPort '3333' (last wins), got '%s'", cfg.SSHPort)
	}
	if cfg.RootUser != "user2" {
		t.Errorf("expected RootUser 'user2' (last wins), got '%s'", cfg.RootUser)
	}
}

func TestApplyOptions_ArgsMerge(t *testing.T) {
	cfg := applyOptions("testhost.example.com",
		WithArg("key1", "value1"),
		WithArgs(map[string]string{
			"key2": "value2",
			"key1": "overridden",
		}),
	)

	if cfg.Args["key1"] != "overridden" {
		t.Errorf("expected Args['key1'] = 'overridden', got '%s'", cfg.Args["key1"])
	}
	if cfg.Args["key2"] != "value2" {
		t.Errorf("expected Args['key2'] = 'value2', got '%s'", cfg.Args["key2"])
	}
	if len(cfg.Args) != 2 {
		t.Errorf("expected 2 args, got %d", len(cfg.Args))
	}
}
