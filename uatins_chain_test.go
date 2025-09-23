package uatins

import (
	"strconv"
	"testing"
	"time"
)

// TestChainMethods verifies that chain methods work and produce the same results as functional options
func TestChainMethods(t *testing.T) {
	// Test basic chaining
	client := NewClient().
		MaxAge(120).
		Strict(true).
		Location(time.Local)

	// Verify the settings are applied
	if client.maxAgeYears != 120 {
		t.Errorf("MaxAge chaining failed: expected 120, got %d", client.maxAgeYears)
	}
	if !client.strict {
		t.Errorf("Strict chaining failed: expected true, got %t", client.strict)
	}
	if client.loc != time.Local {
		t.Errorf("Location chaining failed: expected time.Local, got %v", client.loc)
	}

	// Compare with functional options approach
	functionalClient := NewClient(
		WithMaxAge(120),
		WithStrict(true),
		WithLocation(time.Local),
	)

	// Both clients should have the same configuration
	if client.maxAgeYears != functionalClient.maxAgeYears {
		t.Errorf("MaxAge mismatch: chain=%d, functional=%d", client.maxAgeYears, functionalClient.maxAgeYears)
	}
	if client.strict != functionalClient.strict {
		t.Errorf("Strict mismatch: chain=%t, functional=%t", client.strict, functionalClient.strict)
	}
	if client.loc != functionalClient.loc {
		t.Errorf("Location mismatch: chain=%v, functional=%v", client.loc, functionalClient.loc)
	}
}

// TestChainMethodsValidation ensures both chaining and functional options produce identical validation results
func TestChainMethodsValidation(t *testing.T) {
	testTIN := "3036045681"
	testDOB := time.Date(1983, 2, 14, 0, 0, 0, 0, time.UTC)

	// Create clients using both approaches
	chainClient := NewClient().
		MaxAge(130).
		Strict(true).
		Location(time.UTC)

	functionalClient := NewClient(
		WithMaxAge(130),
		WithStrict(true),
		WithLocation(time.UTC),
	)

	// Both should produce identical results
	chainResult, chainErr := chainClient.Validate(testTIN, &testDOB)
	functionalResult, functionalErr := functionalClient.Validate(testTIN, &testDOB)

	// Compare errors
	if (chainErr == nil) != (functionalErr == nil) {
		t.Errorf("Error mismatch: chain error=%v, functional error=%v", chainErr, functionalErr)
	}
	if chainErr != nil && functionalErr != nil {
		if chainErr.Error() != functionalErr.Error() {
			t.Errorf("Error message mismatch: chain=%q, functional=%q", chainErr.Error(), functionalErr.Error())
		}
	}

	// Compare results
	if chainResult.Valid != functionalResult.Valid {
		t.Errorf("Valid mismatch: chain=%t, functional=%t", chainResult.Valid, functionalResult.Valid)
	}
	if chainResult.Sex != functionalResult.Sex {
		t.Errorf("Sex mismatch: chain=%s, functional=%s", chainResult.Sex, functionalResult.Sex)
	}
	if !chainResult.BirthDate.Equal(functionalResult.BirthDate) {
		t.Errorf("BirthDate mismatch: chain=%s, functional=%s", chainResult.BirthDate, functionalResult.BirthDate)
	}
}

// TestChainMethodsWithRules tests the Rules chaining method
func TestChainMethodsWithRules(t *testing.T) {
	// Create a custom rule
	forbid1999 := Rule[string](func(s string) error {
		days, _ := strconv.Atoi(s[:5])
		date := DaysToDate(days)
		if date.Year() == 1999 {
			return &Error{
				Code: "custom",
				TIN:  s,
				Msg:  "TINs from 1999 are not allowed",
			}
		}
		return nil
	})

	rules := Rules[string]{forbid1999}

	// Test chaining approach
	chainClient := NewClient().Rules(rules)

	// Test functional approach
	functionalClient := NewClient(WithRules(rules))

	// Use a TIN that should trigger the custom rule (1999-12-31)
	testTIN := "3652412345"

	chainResult, chainErr := chainClient.Validate(testTIN, nil)
	functionalResult, functionalErr := functionalClient.Validate(testTIN, nil)

	// Both should produce identical results
	if (chainErr == nil) != (functionalErr == nil) {
		t.Errorf("Custom rule error mismatch: chain error=%v, functional error=%v", chainErr, functionalErr)
	}

	if chainResult.Valid != functionalResult.Valid {
		t.Errorf("Custom rule valid mismatch: chain=%t, functional=%t", chainResult.Valid, functionalResult.Valid)
	}
}

// TestChainMethodsWithNow tests the Now chaining method
func TestChainMethodsWithNow(t *testing.T) {
	testTime := time.Date(2020, 1, 1, 12, 0, 0, 0, time.UTC)

	// Test chaining approach
	chainClient := NewClient().Now(testTime)

	// Test functional approach
	functionalClient := NewClient(WithNow(testTime))

	// Both should have the same time set
	if !chainClient.now.Equal(functionalClient.now) {
		t.Errorf("Now time mismatch: chain=%s, functional=%s", chainClient.now, functionalClient.now)
	}

	// Verify the time is actually used in UTC
	expectedTime := testTime.In(time.UTC)
	if !chainClient.now.Equal(expectedTime) {
		t.Errorf("Now time not converted to UTC: expected=%s, got=%s", expectedTime, chainClient.now)
	}
}

// TestChainMethodsReturnValue ensures all chain methods return the client for chaining
func TestChainMethodsReturnValue(t *testing.T) {
	client := NewClient()
	
	// Each method should return the same client instance
	if client.MaxAge(100) != client {
		t.Error("MaxAge should return the same client instance")
	}
	if client.Strict(true) != client {
		t.Error("Strict should return the same client instance")
	}
	if client.Location(time.UTC) != client {
		t.Error("Location should return the same client instance")
	}
	if client.Rules(Rules[string]{}) != client {
		t.Error("Rules should return the same client instance")
	}
	if client.Now(time.Now()) != client {
		t.Error("Now should return the same client instance")
	}
}