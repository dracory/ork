package swap

// Argument key constants for use with GetArg.
const (
	// ArgSize is the swap file size argument key
	ArgSize = "size"

	// ArgUnit is the size unit argument key (gb|mb)
	ArgUnit = "unit"

	// ArgSwappiness is the kernel swappiness argument key
	ArgSwappiness = "swappiness"

	// ArgSwapFilePath is the path for the swap file
	ArgSwapFilePath = "swapfile-path"
)

// Default configuration constants for swap playbooks.
const (
	// DefaultSize is the default swap file size (1)
	DefaultSize = "1"

	// DefaultUnit is the default size unit (gb)
	DefaultUnit = "gb"

	// DefaultSwappiness is the default kernel swappiness value (10)
	// Lower values prefer RAM over swap, better for database workloads
	DefaultSwappiness = "10"

	// DefaultSwapFilePath is the default path for the swap file
	DefaultSwapFilePath = "/swapfile"
)
