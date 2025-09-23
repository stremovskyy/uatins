package uatins_test_test

import (
	"fmt"
	"strconv"
	"time"

	"github.com/stremovskyy/uatins"
)

// Example_chainMethods demonstrates using the fluent chain methods API
// as an alternative to functional options.
func Example_chainMethods() {
	dob := time.Date(1983, 2, 14, 13, 0, 0, 0, time.UTC)

	// Using the new chain methods API - more fluent and readable
	client := uatins.NewClient().
		MaxAge(130).
		Strict(true).
		Location(time.Local)

	res, err := client.Validate("3036045681", &dob)

	if err != nil {
		fmt.Println("err:", err)
		return
	}

	fmt.Println(res.Valid, res.Sex)
	// Output:
	// true female
}

// Example_chainMethodsWithCustomRules shows how to use chain methods with custom rules.
func Example_chainMethodsWithCustomRules() {
	// Create a custom rule using chain methods
	forbid1999 := uatins.Rule[string](func(s string) error {
		days, _ := strconv.Atoi(s[:5])
		date := uatins.DaysToDate(days)
		if date.Year() == 1999 {
			return fmt.Errorf("TINs from 1999 are not allowed")
		}
		return nil
	})

	// Chain multiple configuration calls
	client := uatins.NewClient().
		Rules(uatins.Rules[string]{forbid1999}).
		MaxAge(120).
		Strict(false)

	// This TIN has a birth date of 1999-12-31 and will fail custom validation
	res, err := client.Validate("3652412345", nil)

	if err != nil {
		fmt.Println("Custom rule error:", err)
		return
	}

	fmt.Println("valid:", res.Valid)
	// Output:
	// Custom rule error: TINs from 1999 are not allowed
}

// Example_chainMethodsComparison shows both API styles side by side.
func Example_chainMethodsComparison() {
	// Original functional options style
	functionalClient := uatins.NewClient(
		uatins.WithMaxAge(120),
		uatins.WithStrict(true),
		uatins.WithLocation(time.UTC),
	)

	// New chain methods style - more fluent
	chainClient := uatins.NewClient().
		MaxAge(120).
		Strict(true).
		Location(time.UTC)

	tin := "3036045681"
	dob := time.Date(1983, 2, 14, 0, 0, 0, 0, time.UTC)

	// Both produce identical results
	result1, _ := functionalClient.Validate(tin, &dob)
	result2, _ := chainClient.Validate(tin, &dob)

	fmt.Println("Functional:", result1.Valid, result1.Sex)
	fmt.Println("Chain:", result2.Valid, result2.Sex)
	fmt.Println("Results match:", result1.Valid == result2.Valid && result1.Sex == result2.Sex)
	// Output:
	// Functional: true female
	// Chain: true female
	// Results match: true
}