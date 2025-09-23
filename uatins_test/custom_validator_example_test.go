package uatins_test_test

import (
	"fmt"
	"strconv"

	"github.com/stremovskyy/uatins"
)

// CustomValidator shows composing your own validator using the exposed Rules.
func Example_customValidator() {
	// Rule: forbid DOBs during a blackout window
	forbidBlackout := uatins.Rule[string](
		func(s string) error {
			dd, _ := strconv.Atoi(s[:5])
			d := uatins.DaysToDate(dd)
			if d.Year() == 1999 && d.Month() == 12 && d.Day() == 31 {
				return fmt.Errorf("blackout date not allowed")
			}
			return nil
		},
	)

	client := uatins.NewClient(
		uatins.WithRules(uatins.Rules[string]{forbidBlackout}),
	)
	res, _ := client.Validate("3652412345", nil) // DOB 1999-12-31, but checksum is wrong

	fmt.Println("valid:", res.Valid)
	// Output:
	// valid: false
}
