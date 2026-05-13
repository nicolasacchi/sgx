package commands

import "testing"

func TestBuildPulseParams_CorrectionParamNames(t *testing.T) {
	p := buildPulseParams(
		false, 95, "",
		true, true, 0.8, true, true,
		"ctrl", "test",
	)

	want := map[string]string{
		"applyBonferroniPerVariant":        "true",
		"applyBonferroniPerMetric":         "true",
		"applyBenjaminiHochbergPerMetric":  "true",
		"applyBenjaminiHochbergPerVariant": "true",
		"bonferroniPrimaryMetricWeight":    "0.800000",
		"cuped":                            "true",
		"confidence":                       "95",
		"control":                          "ctrl",
		"test":                             "test",
	}
	for k, v := range want {
		if got := p.Get(k); got != v {
			t.Errorf("param %q = %q, want %q", k, got, v)
		}
	}

	// Guard against regressing to the pre-fix names that Statsig's API
	// silently ignored, causing uncorrected results under --bonferroni/--bh.
	for _, dead := range []string{
		"bonferroniPerVariant",
		"bonferroniPerMetric",
		"bhPerMetric",
		"bhPerVariant",
		"bonferroniAlphaWeight",
	} {
		if _, ok := p[dead]; ok {
			t.Errorf("deprecated param %q is still being sent", dead)
		}
	}
}

func TestBuildPulseParams_NoCupedAndDate(t *testing.T) {
	p := buildPulseParams(
		true, 0, "2026-04-24",
		false, false, 0, false, false,
		"ctrl", "test",
	)
	if got := p.Get("cuped"); got != "false" {
		t.Errorf("cuped = %q, want %q", got, "false")
	}
	if got := p.Get("date"); got != "2026-04-24" {
		t.Errorf("date = %q, want %q", got, "2026-04-24")
	}
	if _, ok := p["confidence"]; ok {
		t.Errorf("confidence should be omitted when 0")
	}
	for _, corr := range []string{
		"applyBonferroniPerVariant",
		"applyBonferroniPerMetric",
		"applyBenjaminiHochbergPerMetric",
		"applyBenjaminiHochbergPerVariant",
		"bonferroniPrimaryMetricWeight",
	} {
		if _, ok := p[corr]; ok {
			t.Errorf("correction param %q should be omitted when flag is off", corr)
		}
	}
}
