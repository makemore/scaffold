// Package config handles scaffold configuration files
package config

// Manifest represents a scaffold.yaml configuration file
type Manifest struct {
	Name        string            `yaml:"name"`
	Description string            `yaml:"description,omitempty"`
	Type        string            `yaml:"type"` // "base" or "module"
	Version     string            `yaml:"version,omitempty"`
	Variables   []Variable        `yaml:"variables,omitempty"`
	Files       FileConfig        `yaml:"files,omitempty"`
	Actions     []Action          `yaml:"actions,omitempty"`
	Requires    []string          `yaml:"requires,omitempty"` // Required modules
	Conflicts   []string          `yaml:"conflicts,omitempty"` // Incompatible modules
}

// Variable represents a template variable
type Variable struct {
	Name        string   `yaml:"name"`
	Description string   `yaml:"description,omitempty"`
	Type        string   `yaml:"type,omitempty"` // string, bool, choice
	Default     string   `yaml:"default,omitempty"`
	Required    bool     `yaml:"required,omitempty"`
	Choices     []string `yaml:"choices,omitempty"` // For type: choice
	Pattern     string   `yaml:"pattern,omitempty"` // Regex validation
}

// FileConfig specifies file handling rules
type FileConfig struct {
	Include []string          `yaml:"include,omitempty"` // Glob patterns to include
	Exclude []string          `yaml:"exclude,omitempty"` // Glob patterns to exclude
	Rename  map[string]string `yaml:"rename,omitempty"`  // File rename mappings
}

// Action represents a post-generation action
type Action struct {
	Name        string   `yaml:"name"`
	Description string   `yaml:"description,omitempty"`
	Type        string   `yaml:"type"` // command, message
	Command     string   `yaml:"command,omitempty"`
	Args        []string `yaml:"args,omitempty"`
	Message     string   `yaml:"message,omitempty"`
	Condition   string   `yaml:"condition,omitempty"` // Variable-based condition
	Optional    bool     `yaml:"optional,omitempty"`  // User can skip
}

// Lockfile represents a scaffold.lock file for reproducibility
type Lockfile struct {
	Version   string       `yaml:"version"`
	Generated string       `yaml:"generated"`
	Base      LockedSource `yaml:"base"`
	Modules   []LockedSource `yaml:"modules,omitempty"`
	Variables map[string]string `yaml:"variables"`
}

// LockedSource represents a locked template/module source
type LockedSource struct {
	Name   string `yaml:"name"`
	Source string `yaml:"source"`
	Ref    string `yaml:"ref,omitempty"`
	Commit string `yaml:"commit,omitempty"` // Resolved commit SHA
	Hash   string `yaml:"hash,omitempty"`   // Content hash for non-git sources
}

