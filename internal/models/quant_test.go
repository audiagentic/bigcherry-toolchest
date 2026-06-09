package models

import "testing"

func TestParseQuant(t *testing.T) {
	cases := []struct {
		filename string
		want     string
	}{
		// UD ultra-dynamic quants — prefix must be preserved consistently.
		{"UD-IQ1_M/MiniMax-M2.5-UD-IQ1_M-00001-of-00003.gguf", "UD_IQ1_M"},
		{"UD-IQ1_S/MiniMax-M2.5-UD-IQ1_S-00001-of-00003.gguf", "UD_IQ1_S"},
		{"UD-IQ2_M/MiniMax-M2.5-UD-IQ2_M-00001-of-00003.gguf", "UD_IQ2_M"},
		{"UD-IQ2_XXS/MiniMax-M2.5-UD-IQ2_XXS-00001-of-00003.gguf", "UD_IQ2_XXS"},
		{"UD-IQ3_XXS/MiniMax-M2.5-UD-IQ3_XXS-00001-of-00003.gguf", "UD_IQ3_XXS"},
		// Previously dropped the XL suffix and reported a bare Q2_K.
		{"UD-Q2_K_XL/MiniMax-M2.5-UD-Q2_K_XL-00001-of-00003.gguf", "UD_Q2_K_XL"},
		{"UD-Q3_K_XL/MiniMax-M2.5-UD-Q3_K_XL-00001-of-00004.gguf", "UD_Q3_K_XL"},
		{"UD-Q4_K_XL/MiniMax-M2.5-UD-Q4_K_XL-00001-of-00004.gguf", "UD_Q4_K_XL"},
		{"UD-Q5_K_XL/MiniMax-M2.5-UD-Q5_K_XL-00001-of-00005.gguf", "UD_Q5_K_XL"},
		{"UD-Q6_K_XL/MiniMax-M2.5-UD-Q6_K_XL-00001-of-00005.gguf", "UD_Q6_K_XL"},

		// MXFP4 — previously reported "unknown".
		{"gpt-oss-20b-MXFP4.gguf", "MXFP4"},
		{"openai-gpt-oss-120b-UD-MXFP4.gguf", "UD_MXFP4"},

		// Plain (non-UD) quants.
		{"model-Q4_K_M.gguf", "Q4_K_M"},
		{"model-Q2_K.gguf", "Q2_K"},
		{"model-Q2_K_S.gguf", "Q2_K_S"},
		{"model-Q8_0.gguf", "Q8_0"},
		{"model-Q4_0.gguf", "Q4_0"},
		{"model-IQ4_NL.gguf", "IQ4_NL"},
		{"model-IQ2_XS.gguf", "IQ2_XS"},
		{"model-TQ1_0.gguf", "TQ1_0"},
		{"model-F16.gguf", "F16"},
		{"model-BF16.gguf", "BF16"},
		{"model-F32.gguf", "F32"},

		// No recognizable quant.
		{"some-random-model.gguf", "unknown"},
	}

	for _, c := range cases {
		if got := ParseQuant(c.filename); got != c.want {
			t.Errorf("ParseQuant(%q) = %q, want %q", c.filename, got, c.want)
		}
	}
}
