package ocli

import (
	"testing"
	"time"

	"github.com/dknelson9876/oregano/omoney"
)

// new tr [acc] [payee] [amount] (date) (cat)
//      (desc) (-t/--time date) (-c/--category cat) (-d/--description desc)

func TestNewTrFullPositional(t *testing.T) {
	tr := CreateManualTransaction([]string{"chase", "mcdonalds", "21.45", "3/18/2022", "fast food", "big mac"})
	need := omoney.NewTransaction(
		time.Date(2022, time.March, 18, 0, 0, 0, 0, time.Local),
		21.45,
		"chase",
		"mcdonalds",
		"fast food",
		"big mac",
	)
	if !need.LooseEquals(tr) {
		t.Fatalf(`Transaction("chase mcdonalds 21.45 3/18/2022 'fast food' 'big mac'")::`+
			"\ngot: %+v"+
			"\nneed %+v",
			tr, need)
	}

}
