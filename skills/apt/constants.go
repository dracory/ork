package apt

// ArgPackages is the argument key for specifying packages to install.
// Value should be a space-separated list of package names, e.g. "nodejs npm".
const ArgPackages = "packages"

// ArgPackage is the argument key for specifying a single package name to
// filter on for read-only queries such as apt list --installed.
// Value should be a single package name, e.g. "nginx". Optional.
const ArgPackage = "package"
