package reboot

// Argument key constants for the reboot skill.
const (
	// ArgWait enables wait-for-reconnect behaviour. Accepts "true"/"false"
	// (any value accepted by strconv.ParseBool). When true, the skill polls
	// the server until it responds again after rebooting.
	ArgWait = "wait"

	// ArgMaxWait sets the maximum total time to wait for the server to come
	// back online, including the initial post-reboot grace period and the
	// polling phase. Accepts a Go duration string (e.g. "5m", "10m30s").
	// Only applies when ArgWait is true.
	ArgMaxWait = "max-wait"

	// ArgInitialWait sets the grace period to wait after sending the reboot
	// command before beginning to poll. This gives the server time to actually
	// start shutting down. Accepts a Go duration string. Only applies when
	// ArgWait is true.
	ArgInitialWait = "initial-wait"

	// ArgPollInterval sets the delay between successive uptime probes while
	// waiting for the server to come back online. Accepts a Go duration
	// string. Only applies when ArgWait is true.
	ArgPollInterval = "poll-interval"
)

// Default configuration constants for the reboot skill.
const (
	// DefaultMaxWait is the default maximum total wait time for reconnect.
	DefaultMaxWait = "5m"

	// DefaultInitialWait is the default grace period after sending reboot
	// before polling begins.
	DefaultInitialWait = "10s"

	// DefaultPollInterval is the default delay between uptime probes.
	DefaultPollInterval = "5s"
)
