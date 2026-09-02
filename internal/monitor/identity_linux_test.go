//go:build linux

package monitor

import "testing"

func TestParseROCmProductCSVUsesBDFIdentity(t *testing.T) {
	// Header casing intentionally mirrors variants seen across rocm-smi
	// releases; parsing must not depend on one capitalization.
	csv := `device,PCI Bus,Card series,Card Model,Card Vendor,Card SKU,Subsystem ID,Device Rev,Node ID,GUID,GFX Version
card0,0000:03:00.0,Radeon RX 7900 XTX,0x744c,AMD,D70201,0x0e3b,0xc8,1,123,gfx1100
card1,0000:09:00.0,Radeon AI PRO R9700,0x7550,AMD,R9700,0x0000,0x00,2,456,gfx1201
`
	// Deliberately reverse physical row order vs ROCm index order: the parser
	// must trust BDF, never card0/card1.
	byBDF := map[string]int{
		"0000:03:00.0": 1,
		"0000:09:00.0": 0,
	}

	got := parseROCmProductCSV(csv, byBDF)
	if len(got) != 2 {
		t.Fatalf("got %d identities, want 2", len(got))
	}
	if got[0].Name != "Radeon AI PRO R9700" || got[0].Arch != "gfx1201" || got[0].BDF != "0000:09:00.0" {
		t.Fatalf("ROCm0 identity mismatch: %+v", got[0])
	}
	if got[1].Name != "Radeon RX 7900 XTX" || got[1].Arch != "gfx1100" || got[1].BDF != "0000:03:00.0" {
		t.Fatalf("ROCm1 identity mismatch: %+v", got[1])
	}
}

func TestParseROCmProductCSVIgnoresUnknownBDF(t *testing.T) {
	csv := `device,PCI Bus,Card Series,GFX Version
card0,0000:03:00.0,Radeon RX 7900 XTX,gfx1100
`
	got := parseROCmProductCSV(csv, map[string]int{"0000:09:00.0": 0})
	if len(got) != 0 {
		t.Fatalf("unexpected identities: %+v", got)
	}
}
