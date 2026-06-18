package summaries

import (
	"encoding/json"
	"testing"
)

func TestRawJSONOrNilFallsBackOnInvalidJSON(t *testing.T) {
	if got := rawJSONOrNil(`[{]`); got != nil {
		t.Fatalf("rawJSONOrNil invalid = %s, want nil", string(got))
	}
	if got := rawJSONOrNil(`[{"title":"ok"}]`); !json.Valid(got) {
		t.Fatalf("rawJSONOrNil valid returned invalid JSON: %s", string(got))
	}
}
