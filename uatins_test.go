package uatins

import (
	"errors"
	"testing"
	"time"
)

func TestValidate_OK(t *testing.T) {
	// Example valid TIN: 00101xxxxx? You must provide a real valid code.
	// For test, construct a synthetic TIN with known DOB and checksum.
	// DOB 1990-01-01 => days since 1899-12-31 = 32874
	dob := DaysToDate(32874)
	prefix := "32874"
	body := "123" // placeholder; we will adjust checksum digit to match
	for d9 := 0; d9 <= 9; d9++ {
		for d10 := 0; d10 <= 9; d10++ {
			tin := prefix + body + string('0'+byte(d9)) + string('0'+byte(d10))
			if ChecksumOK(tin) {
				client := NewClient(WithStrict(true))
				res, err := client.Validate(tin, &dob)
				if err != nil {
					t.Fatalf("unexpected err: %v", err)
				}
				if !res.Valid || !res.DOBMatched || !res.ChecksumOK || !res.BirthDatePlausible {
					t.Fatalf("unexpected result: %+v", res)
				}
				return
			}
		}
	}
	t.Fatalf("could not synthesize a valid TIN for test; ChecksumOK never true")
}

func TestErrors(t *testing.T) {
	client := NewClient()
	_, err := client.Validate("12A", nil)
	if err == nil || !errorsIs(err, ErrLength) {
		t.Fatalf("expected ErrLength, got %v", err)
	}

	_, err = client.Validate("1111111111", nil)
	if err == nil || !errorsIs(err, ErrAllSame) {
		t.Fatalf("expected ErrAllSame, got %v", err)
	}

	res, err := client.Validate("1234567890", nil)
	if err != nil {
		t.Fatalf("unexpected err for checksum test: %v", err)
	}
	if res.Valid {
		t.Fatalf("expected invalid checksum result, got valid")
	}
}

func TestDOBStrictMismatch(t *testing.T) {
	dob := time.Date(1980, 7, 10, 0, 0, 0, 0, time.UTC)
	// Make a TIN sharing DOB but wrong checksum to force crafting loop
	prefix := "29318" // 1980-07-10 days since 1899-12-31
	body := "5678"
	var tin string
	for d9 := 1; d9 <= 9; d9 += 2 { // odd male
		for d10 := 0; d10 <= 9; d10++ {
			candidate := prefix + body + string('0'+byte(d9)) + string('0'+byte(d10))
			if ChecksumOK(candidate) {
				tin = candidate
				break
			}
		}
		if tin != "" {
			break
		}
	}
	if tin == "" {
		t.Skip("could not craft candidate with valid checksum; skip")
	}

	wrong := dob.AddDate(0, 0, 1)
	client := NewClient(WithStrict(true))
	_, err := client.Validate(tin, &wrong)
	if err == nil || !errorsIs(err, ErrDOBMismatch) {
		t.Fatalf("expected ErrDOBMismatch, got %v", err)
	}
}

func errorsIs(err error, target error) bool {
	return err != nil && (err == target || errors.Is(err, target))
}
