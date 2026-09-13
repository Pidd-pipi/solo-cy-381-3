package splitcalc

import (
	"testing"
)

func TestCalculateSharesEqual(t *testing.T) {
	tests := []struct {
		name     string
		total    float64
		users    []uint
		wantSum  float64
		wantEach float64
	}{
		{"three way 90", 90, []uint{1, 2, 3}, 90, 30},
		{"two way 100", 100, []uint{1, 2}, 100, 50},
		{"odd cents 100.01 three", 100.01, []uint{1, 2, 3}, 100.01, 33.34},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ps := make([]Participant, 0, len(tt.users))
			for _, u := range tt.users {
				ps = append(ps, Participant{UserID: u})
			}
			shares, err := CalculateShares(tt.total, SplitEqual, ps)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			sum := 0.0
			for _, s := range shares {
				sum += s.ShareAmount
			}
			if round2(sum) != tt.wantSum {
				t.Fatalf("sum = %.2f, want %.2f", sum, tt.wantSum)
			}
		})
	}
}

func TestCalculateSharesRatio(t *testing.T) {
	shares, err := CalculateShares(100, SplitRatio, []Participant{
		{UserID: 1, Ratio: 1},
		{UserID: 2, Ratio: 2},
		{UserID: 3, Ratio: 1},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := map[uint]float64{1: 25, 2: 50, 3: 25}
	for _, s := range shares {
		if s.ShareAmount != want[s.UserID] {
			t.Fatalf("user %d amount = %.2f, want %.2f", s.UserID, s.ShareAmount, want[s.UserID])
		}
	}
}

func TestCalculateSharesAmountMismatch(t *testing.T) {
	_, err := CalculateShares(100, SplitAmount, []Participant{
		{UserID: 1, Amount: 30},
		{UserID: 2, Amount: 30},
	})
	if err != ErrSumMismatch {
		t.Fatalf("err = %v, want ErrSumMismatch", err)
	}
}

func TestCalculateSharesInvalid(t *testing.T) {
	if _, err := CalculateShares(0, SplitEqual, []Participant{{UserID: 1}}); err != ErrInvalidSplit {
		t.Fatalf("err = %v, want ErrInvalidSplit", err)
	}
	if _, err := CalculateShares(100, SplitRatio, []Participant{{UserID: 1, Ratio: 0}}); err != ErrInvalidSplit {
		t.Fatalf("err = %v, want ErrInvalidSplit", err)
	}
}

func TestOptimizeTransfers(t *testing.T) {
	balances := []Balance{
		{UserID: 1, Amount: 100},
		{UserID: 2, Amount: -30},
		{UserID: 3, Amount: -70},
	}
	transfers := OptimizeTransfers(balances)
	sum := 0.0
	for _, tr := range transfers {
		sum += tr.Amount
		if tr.FromUserID == 0 || tr.ToUserID == 0 {
			t.Fatalf("transfer has zero user: %+v", tr)
		}
	}
	if round2(sum) != 100 {
		t.Fatalf("transfer sum = %.2f, want 100", sum)
	}
	if len(transfers) > 2 {
		t.Fatalf("transfers = %d, want <= 2", len(transfers))
	}
}

func TestOptimizeTransfersEmpty(t *testing.T) {
	if got := OptimizeTransfers(nil); len(got) != 0 {
		t.Fatalf("expected no transfers, got %d", len(got))
	}
}
