package systemctl

// Argument key constants for use with GetArg/SetArg.
const (
	// ArgService specifies the systemd unit name (e.g. "caddy", "mariadb",
	// "php8.5-fpm", "mariadb-backup.timer"). Required for all skills except
	// DaemonReload, which operates on the systemd manager itself.
	ArgService = "service"

	// ArgStart, when set to "true", causes Enable to also start the unit
	// (systemctl start) after enabling it. Any other value or empty means
	// enable only. This matches the common "enable && start" idiom.
	ArgStart = "start"

	// ArgStop, when set to "true", causes Disable to also stop the unit
	// (systemctl stop) after disabling it. Any other value or empty means
	// disable only.
	ArgStop = "stop"
)
