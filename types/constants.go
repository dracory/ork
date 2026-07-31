package types

// Constants for Atom property keys and map keys used by BaseSkill/BasePlaybook
// for state storage and cloning. Centralizing these prevents typos and makes
// refactoring safer.
const (
	// atomTypeSkill is the Atom type for skills.
	atomTypeSkill = "skill"
	// atomTypePlaybook is the Atom type for playbooks.
	atomTypePlaybook = "playbook"

	// propID is the Atom property key for the runnable's ID.
	propID = "id"
	// propDescription is the Atom property key for the description.
	propDescription = "description"
	// propDryRun is the Atom property key for the dry-run flag.
	propDryRun = "dryRun"
	// propTimeout is the Atom property key for the timeout duration (stored as int64 string).
	propTimeout = "timeout"
	// propBecomeUser is the Atom property key for the become user.
	propBecomeUser = "becomeUser"

	// argPrefix is prepended to argument keys to namespace them in the Atom's flat map.
	argPrefix = "arg_"

	// mapKeyNodeConfig is the ToMap/FromMap key for the NodeConfig struct.
	mapKeyNodeConfig = "nodeConfig"
	// mapKeyType is the ToMap/FromMap key for the Atom type string.
	mapKeyType = "type"
	// mapKeyProperties is the ToMap/FromMap key for the Atom properties sub-map.
	mapKeyProperties = "properties"

	// boolTrue is the string representation of true used in Atom string values.
	boolTrue = "true"
)

// Constants for commandImplementation's command-specific map keys.
// Exported because commandImplementation lives in the ork package, not types.
const (
	// MapKeyCommand is the ToMap/FromMap key for the command string.
	MapKeyCommand = "command"
	// MapKeyRequired is the ToMap/FromMap key for the required flag.
	MapKeyRequired = "required"
	// MapKeyChdir is the ToMap/FromMap key for the chdir working directory.
	MapKeyChdir = "chdir"
)

// Constants for struct field names used by reflection in cloneFromMap.
// Exported because cloneFromMap lives in the ork package, not types.
const (
	// FieldBaseSkill is the struct field name for embedded *BaseSkill.
	FieldBaseSkill = "BaseSkill"
	// FieldBasePlaybook is the struct field name for embedded *BasePlaybook.
	FieldBasePlaybook = "BasePlaybook"
)
