package ocli

import (
	"testing"
	"time"

	om "github.com/dknelson9876/oregano/omoney"
	"github.com/dknelson9876/oregano/outil"
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

// func TestNewTrPartialPositional(t *testing.T) {
// 	inputs := [][]string{
// 		{"chase", "mcdonalds", "21.45", "3/18/2022", "fast food"},
// 		{"chase", "mcdonalds", "21.45", "3/18/2022"},
// 		{"chase", "mcdonalds", "21.45"},
// 	}
// 	expected := []*om.Transaction{
// 		om.NewTransaction("chase", "mcdonalds", 21.45,
// 			om.WithDate(time.Date(2022, time.March, 18, 0, 0, 0, 0, time.Local)),
// 			om.WithCategory("fast food"),
// 		),
// 		om.NewTransaction("chase", "mcdonalds", 21.45,
// 			om.WithDate(time.Date(2022, time.March, 18, 0, 0, 0, 0, time.Local)),
// 		),
// 		om.NewTransaction("chase", "mcdonalds", 21.45),
// 	}
// 	for i, input := range inputs {
// 		tr := CreateManualTransaction(input)
// 		if !expected[i].LooseEquals(tr) {
// 			t.Errorf(`Transaction("chase mcdonalds 21.45 3/18/2022 'fast food' 'big mac'")::partial`+
// 				"\ngot: %+v"+
// 				"\nneed %+v",
// 				tr, expected[i])
// 		}
// 	}
// }

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
			t.Errorf(`Transaction("chase mcdonalds 21.45 3/18/2022 'fast food' 'big mac'")::partial`+
				"\ngot: %+v"+
				"\nneed %+v",
				tr, expected)
		}
	}
}

func TestDateRounding(t *testing.T) {
	fakeNow := time.Date(2024, 02, 05, 13, 41, 26, 0, time.Local)
	var rec, exp time.Time

	rec = BeginningOfDay(fakeNow)
	exp = time.Date(2024, 02, 05, 0, 0, 0, 0, time.Local)
	if !exp.Equal(rec) {
		t.Errorf("BeginningOfDay failed"+
			"\ngot: %+v"+
			"\nneed: %+v",
			rec, exp)
	}

	rec = EndOfDay(fakeNow)
	exp = time.Date(2024, 02, 05, 23, 59, 59, 0, time.Local)
	if !exp.Equal(rec) {
		t.Errorf("EndOfDay failed"+
			"\ngot: %+v"+
			"\nneed: %+v",
			rec, exp)
	}

	rec = BeginningOfMonth(fakeNow)
	exp = time.Date(2024, 02, 01, 0, 0, 0, 0, time.Local)
	if !exp.Equal(rec) {
		t.Errorf("BeginningOfMonth failed"+
			"\ngot: %+v"+
			"\nneed: %+v",
			rec, exp)
	}

	rec = EndOfMonth(fakeNow)
	exp = time.Date(2024, 02, 29, 23, 59, 59, 0, time.Local)
	if !exp.Equal(rec) {
		t.Errorf("EndOfMonth failed"+
			"\ngot: %+v"+
			"\nneed: %+v",
			rec, exp)
	}

	rec = BeginningOfYear(fakeNow)
	exp = time.Date(2024, 01, 01, 0, 0, 0, 0, time.Local)
	if !exp.Equal(rec) {
		t.Errorf("BeginningOfYear failed"+
			"\ngot: %+v"+
			"\nneed: %+v",
			rec, exp)
	}

	rec = EndOfYear(fakeNow)
	exp = time.Date(2024, 12, 31, 23, 59, 59, 0, time.Local)
	if !exp.Equal(rec) {
		t.Errorf("EndOfMonth failed"+
			"\ngot: %+v"+
			"\nneed: %+v",
			rec, exp)
	}

}

func TestDateRangeInputParse(t *testing.T) {
	// 4:24pm May 23, 2023
	fakeNow := time.Date(2023, 05, 23, 16, 24, 26, 0, time.UTC)
	var emptyDate outil.Option[time.Time]

	// --- -> default to month range
	startRec, endRec, err := ParseDateInput(emptyDate, emptyDate, "", fakeNow)
	startExp := time.Date(2023, 05, 01, 0, 0, 0, 0, time.UTC)
	endExp := time.Date(2023, 05, 31, 23, 59, 59, 0, time.UTC)
	if err != nil {
		t.Errorf("Incorrect err on --- date case: %s", err)
	}
	if !(startRec.Equal(time.Date(2023, 05, 01, 0, 0, 0, 0, time.UTC)) && endRec.Equal(time.Date(2023, 05, 31, 23, 59, 59, 0, time.UTC))) {
		t.Errorf("Incorrect date case: ---"+
			"\ngot: %+v through %+v"+
			"\nneed: %+v through %+v",
			startRec, endRec, startExp, endExp)
	}

}
