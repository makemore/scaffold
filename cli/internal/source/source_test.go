package source

import (
	"testing"
)

func TestParse(t *testing.T) {
	tests := []struct {
		name       string
		uri        string
		wantType   Type
		wantURL    string
		wantRef    string
		wantSubdir string
		wantErr    bool
	}{
		{
			name:     "github shorthand",
			uri:      "github:org/repo",
			wantType: TypeGit,
			wantURL:  "https://github.com/org/repo",
		},
		{
			name:     "gitlab shorthand",
			uri:      "gitlab:org/repo",
			wantType: TypeGit,
			wantURL:  "https://gitlab.com/org/repo",
		},
		{
			name:     "bitbucket shorthand",
			uri:      "bitbucket:org/repo",
			wantType: TypeGit,
			wantURL:  "https://bitbucket.org/org/repo",
		},
		{
			name:     "git prefix with https",
			uri:      "git:https://github.com/org/repo",
			wantType: TypeGit,
			wantURL:  "https://github.com/org/repo",
		},
		{
			name:       "github with subdir",
			uri:        "github:org/repo//templates/base",
			wantType:   TypeGit,
			wantURL:    "https://github.com/org/repo",
			wantSubdir: "templates/base",
		},
		{
			name:     "github with ref",
			uri:      "github:org/repo#v1.0.0",
			wantType: TypeGit,
			wantURL:  "https://github.com/org/repo",
			wantRef:  "v1.0.0",
		},
		{
			name:       "github with subdir and ref",
			uri:        "github:org/repo//templates/base#main",
			wantType:   TypeGit,
			wantURL:    "https://github.com/org/repo",
			wantSubdir: "templates/base",
			wantRef:    "main",
		},
		{
			name:     "file prefix relative",
			uri:      "file:./templates/base",
			wantType: TypeFile,
			wantURL:  "./templates/base",
		},
		{
			name:     "file prefix home",
			uri:      "file:~/templates/base",
			wantType: TypeFile,
			wantURL:  "~/templates/base",
		},
		{
			name:     "file prefix absolute",
			uri:      "file:/absolute/path",
			wantType: TypeFile,
			wantURL:  "/absolute/path",
		},
		{
			name:     "plain https URL",
			uri:      "https://example.com/template.tar.gz",
			wantType: TypeURL,
			wantURL:  "https://example.com/template.tar.gz",
		},
		{
			name:    "empty uri",
			uri:     "",
			wantErr: true,
		},
		{
			name:    "unknown format",
			uri:     "unknown:something",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Parse(tt.uri)
			if (err != nil) != tt.wantErr {
				t.Errorf("Parse() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr {
				return
			}
			if got.Type != tt.wantType {
				t.Errorf("Parse() Type = %v, want %v", got.Type, tt.wantType)
			}
			if got.URL != tt.wantURL {
				t.Errorf("Parse() URL = %v, want %v", got.URL, tt.wantURL)
			}
			if got.Ref != tt.wantRef {
				t.Errorf("Parse() Ref = %v, want %v", got.Ref, tt.wantRef)
			}
			if got.Subdir != tt.wantSubdir {
				t.Errorf("Parse() Subdir = %v, want %v", got.Subdir, tt.wantSubdir)
			}
		})
	}
}

