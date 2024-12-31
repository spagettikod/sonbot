package energy

import (
	"testing"
	"time"
)

func TestTimestampAsTime(t *testing.T) {
	type TestCase struct {
		Expected  string
		Timestamp string
		Offset    int
	}

	cases := []TestCase{
		{Expected: "2024-12-26T14:38:03+01:00", Timestamp: "2024-12-26 14:38:03", Offset: 1},
		{Expected: "2024-12-26T14:38:03-01:00", Timestamp: "2024-12-26 14:38:03", Offset: -1},
		{Expected: "2024-12-26T14:38:03Z", Timestamp: "2024-12-26 14:38:03", Offset: 0},
	}

	for _, tc := range cases {
		status := sonnenBatteryStatus{Timestamp: tc.Timestamp, UTCOffset: tc.Offset}
		ts, err := status.TimestampAsTime()
		if err != nil {
			t.Fatal(err)
		}
		if ts.Format(time.RFC3339) != tc.Expected {
			t.Fatalf("expected %s but got %s", tc.Expected, ts.Format(time.RFC3339))
		}
	}
}
