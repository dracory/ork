// Package php provides skills for managing PHP installations on Debian/Ubuntu systems.
// It supports installing and removing PHP with extensions, configuring PHP-FPM,
// and managing Composer (PHP dependency manager).
package php

// Argument key constants for use with GetArg.
const (
	// ArgVersion specifies the PHP version (e.g. "8.3", "8.4")
	ArgVersion = "version"

	// ArgUser specifies the user to run PHP-FPM as
	ArgUser = "user"

	// ArgExtensions specifies additional PHP extensions (space-separated)
	ArgExtensions = "extensions"
)

// Default configuration constants.
const (
	// DefaultExtensions is a convenience set of common PHP extensions.
	// It is NOT auto-applied by Install; pass it explicitly via
	// SetExtensions(php.DefaultExtensions) when you want the bundled set.
	// Each entry maps to a php<version>-<ext> package on Debian/Ubuntu.
	DefaultExtensions = "cli fpm mysql mbstring xml curl gd zip intl bcmath readline"

	// DefaultFpmPoolPath is the default FPM pool config path pattern
	// The actual path is /etc/php/<version>/fpm/pool.d/www.conf
	DefaultFpmPoolPath = "/etc/php/%s/fpm/pool.d/www.conf"

	// ComposerBinaryPath is the path where Composer is installed
	ComposerBinaryPath = "/usr/local/bin/composer"

	// ComposerInstallerPath is the temporary path for the Composer installer
	ComposerInstallerPath = "/tmp/composer-setup.php"

	// ComposerInstallerSigPath is the temporary path for the installer signature
	ComposerInstallerSigPath = "/tmp/composer-setup.sig"

	// ComposerInstallerUrl is the URL for the Composer installer
	ComposerInstallerUrl = "https://getcomposer.org/installer"

	// ComposerSigUrl is the URL for the Composer installer signature
	ComposerSigUrl = "https://composer.github.io/installer.sig"
)
