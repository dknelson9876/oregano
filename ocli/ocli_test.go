package ocli

import (
	"testing"
	"time"

	om "github.com/dknelson9876/oregano/omoney"
)

// new tr [acc] [payee] [amount] (date) (cat)
//      (desc) (-t/--time date) (-c/--category cat) (-d/--description desc)

func TestNewTrFullPositional(t *testing.T) {
	tr := CreateManualTransaction([]string{"chase", "mcdonalds", "21.45", "3/18/2022", "fast food", "big mac"})
	need := om.NewTransaction("chase", "mcdonalds", 21.45,
		om.WithDate(time.Date(2022, time.March, 18, 0, 0, 0, 0, time.Local)),
		om.WithCategory("fast food"),
		om.WithDescription("big mac"),
	)
	if !need.LooseEquals(tr) {
		t.Fatalf(`Transaction("chase mcdonalds 21.45 3/18/2022 'fast food' 'big mac'")::`+
			"\ngot: %+v"+
			"\nneed %+v",
			tr, need)
	}

}

func TestNewTrPartialPositional(t *testing.T) {
	inputs := [][]string{
		{"chase", "mcdonalds", "21.45", "3/18/2022", "fast food"},
		{"chase", "mcdonalds", "21.45", "3/18/2022"},
		{"chase", "mcdonalds", "21.45"},
	}
	expected := []*om.Transaction{
		om.NewTransaction("chase", "mcdonalds", 21.45,
			om.WithDate(time.Date(2022, time.March, 18, 0, 0, 0, 0, time.Local)),
			om.WithCategory("fast food"),
		),
		om.NewTransaction("chase", "mcdonalds", 21.45,
			om.WithDate(time.Date(2022, time.March, 18, 0, 0, 0, 0, time.Local)),
		),
		om.NewTransaction("chase", "mcdonalds", 21.45),
	}
	for i, input := range inputs {
		tr := CreateManualTransaction(input)
		if !expected[i].LooseEquals(tr) {
			t.Fatalf(`Transaction("chase mcdonalds 21.45 3/18/2022 'fast food' 'big mac'")::partial`+
				"\ngot: %+v"+
				"\nneed %+v",
				tr, expected[i])
		}
	}
}

func TestNewTrMixedPositionalAndFlag(t *testing.T) {
	inputs := [][]string{
		{"chase", "mcdonalds", "21.45", "3/18/2022", "-c", "fast food", "-d", "big mac"},
		{"chase", "mcdonalds", "21.45", "3/18/2022", "-d", "big mac", "-c", "fast food"},
		{"chase", "mcdonalds", "21.45", "-d", "big mac", "-c", "fast food", "-t", "3/18/2022"},
	}
	expected := om.NewTransaction("chase", "mcdonalds", 21.45,
		om.WithDate(time.Date(2022, time.March, 18, 0, 0, 0, 0, time.Local)),
		om.WithCategory("fast food"),
		om.WithDescription("big mac"),
	)
	for _, input := range inputs {
		tr := CreateManualTransaction(input)
		if !expected.LooseEquals(tr) {
			t.Fatalf(`Transaction("chase mcdonalds 21.45 3/18/2022 'fast food' 'big mac'")::partial`+
				"\ngot: %+v"+
				"\nneed %+v",
				tr, expected)
		}
	}
}
