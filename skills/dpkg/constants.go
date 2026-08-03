// Package dpkg provides skills for managing Debian packages via dpkg-level
// tools (dpkg-query, dpkg-reconfigure, etc.). These are lower-level than the
// apt skills — they interact directly with the dpkg package database rather
// than going through the apt frontend.
package dpkg

// ArgPackage is the argument key for specifying a single package name.
// Value should be a single package name, e.g. "nginx".
const ArgPackage = "package"
