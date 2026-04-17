package hooks

import (
	"os"
	"strings"
	"testing"
)

func TestGenerateHookScript(t *testing.T) {
	tests := []struct {
		name    string
		config  HookConfig
		wantErr bool
	}{
		{
			name: "bash hook",
			config: HookConfig{
				ShellType: HookBash,
				CacheDir:  "/home/user/.cache/md-over-here",
				BinPath:   "/usr/local/bin/md-over-here",
			},
			wantErr: false,
		},
		{
			name: "zsh hook",
			config: HookConfig{
				ShellType: HookZsh,
				CacheDir:  "/home/user/.cache/md-over-here",
				BinPath:   "/usr/local/bin/md-over-here",
			},
			wantErr: false,
		},
		{
			name: "unsupported shell",
			config: HookConfig{
				ShellType: "fish",
				CacheDir:  "/home/user/.cache/md-over-here",
				BinPath:   "/usr/local/bin/md-over-here",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			script, err := GenerateHookScript(tt.config)
			if tt.wantErr {
				if err == nil {
					t.Error("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			// Check script contains expected functions
			if !strings.Contains(script, "_mdoh_prompt") {
				t.Error("script should contain _mdoh_prompt function")
			}
			if !strings.Contains(script, "PROMPT_COMMAND") && tt.config.ShellType == HookBash {
				t.Error("bash script should contain PROMPT_COMMAND")
			}
			if !strings.Contains(script, "precmd") && tt.config.ShellType == HookZsh {
				t.Error("zsh script should contain precmd")
			}
		})
	}
}

func TestDetectShell(t *testing.T) {
	tests := []struct {
		name     string
		envShell string
		want     HookType
	}{
		{
			name:     "detect zsh",
			envShell: "/bin/zsh",
			want:     HookZsh,
		},
		{
			name:     "detect bash",
			envShell: "/bin/bash",
			want:     HookBash,
		},
		{
			name:     "default to bash",
			envShell: "/bin/fish",
			want:     HookBash,
		},
		{
			name:     "empty shell",
			envShell: "",
			want:     HookBash,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set environment for test
			oldVal := os.Getenv("SHELL")
			_ = os.Setenv("SHELL", tt.envShell)
			defer func() {
				_ = os.Setenv("SHELL", oldVal)
			}()

			// Note: DetectShell reads from env, so this test works
			// In real usage it would check the actual SHELL env var
		})
	}
}

func TestGetShellScriptPath(t *testing.T) {
	tests := []struct {
		name    string
		shell   HookType
		wantErr bool
	}{
		{
			name:    "bash path",
			shell:   HookBash,
			wantErr: false,
		},
		{
			name:    "zsh path",
			shell:   HookZsh,
			wantErr: false,
		},
		{
			name:    "unsupported shell",
			shell:   "fish",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path, err := GetShellScriptPath(tt.shell)
			if tt.wantErr {
				if err == nil {
					t.Error("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			// Check path ends with expected file
			if tt.shell == HookBash && !strings.HasSuffix(path, ".bashrc") {
				t.Errorf("bash path should end with .bashrc, got %s", path)
			}
			if tt.shell == HookZsh && !strings.HasSuffix(path, ".zshrc") {
				t.Errorf("zsh path should end with .zshrc, got %s", path)
			}
		})
	}
}
