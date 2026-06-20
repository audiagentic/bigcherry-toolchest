package models

import "testing"

func TestDetectReasoning(t *testing.T) {
	tests := []struct {
		name        string
		template    string
		wantSupport bool
		wantToggle  string
		wantKwarg   string
		wantDefault bool
	}{
		{
			name:        "qwen enable_thinking",
			template:    "{%- if enable_thinking is defined and enable_thinking is false %}...",
			wantSupport: true,
			wantToggle:  ReasoningToggleChatTemplateKwargs,
			wantKwarg:   "enable_thinking",
			wantDefault: true,
		},
		{
			name:        "openai reasoning_effort",
			template:    "{{ reasoning_effort }} system prompt...",
			wantSupport: true,
			wantToggle:  ReasoningToggleReasoningEffort,
			wantKwarg:   "",
			wantDefault: true,
		},
		{
			name:        "plain template, no reasoning",
			template:    "{{ messages }}",
			wantSupport: false,
			wantToggle:  ReasoningToggleNone,
			wantKwarg:   "",
			wantDefault: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := detectReasoning(tt.template)
			if got.Supported != tt.wantSupport {
				t.Errorf("Supported = %v, want %v", got.Supported, tt.wantSupport)
			}
			if got.Toggle != tt.wantToggle {
				t.Errorf("Toggle = %q, want %q", got.Toggle, tt.wantToggle)
			}
			if got.Kwarg != tt.wantKwarg {
				t.Errorf("Kwarg = %q, want %q", got.Kwarg, tt.wantKwarg)
			}
			if got.DefaultEnabled != tt.wantDefault {
				t.Errorf("DefaultEnabled = %v, want %v", got.DefaultEnabled, tt.wantDefault)
			}
		})
	}
}

func TestEffectiveReasoning(t *testing.T) {
	// Override wins over detected value.
	m := &Model{Reasoning: ReasoningCapability{Supported: true, Toggle: ReasoningToggleChatTemplateKwargs, Kwarg: "enable_thinking", DefaultEnabled: true}}
	override := &ReasoningCapability{Supported: false, Toggle: ReasoningToggleNone}
	cfg := &ModelConfig{ReasoningOverride: override}
	if got := m.EffectiveReasoning(cfg); got.Supported {
		t.Errorf("override should disable reasoning, got %+v", got)
	}

	// No override → detected value, with empty toggle normalized to "none".
	m2 := &Model{Reasoning: ReasoningCapability{Supported: false}}
	if got := m2.EffectiveReasoning(nil); got.Toggle != ReasoningToggleNone {
		t.Errorf("empty toggle should normalize to none, got %q", got.Toggle)
	}

	// Detected reasoning surfaces when no override is set.
	if got := m.EffectiveReasoning(nil); !got.Supported || got.Kwarg != "enable_thinking" {
		t.Errorf("expected detected reasoning, got %+v", got)
	}
}
