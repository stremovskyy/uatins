# uatins: A Go Validator for Ukrainian Taxpayer IDs (RNOKPP)

[![Go Reference](https://pkg.go.dev/badge/github.com/stremovskyy/uatins.svg)](https://pkg.go.dev/github.com/stremovskyy/uatins)
[![Go Report Card](https://goreportcard.com/badge/github.com/stremovskyy/uatins)](https://goreportcard.com/report/github.com/stremovskyy/uatins)

`uatins` is a lightweight, zero-dependency Go library for validating Ukrainian taxpayer identification numbers (РНОКПП - Registration Number of the Taxpayer's Account Card). It provides a simple, reusable client to perform validation, decode birth dates, and determine the sex of the TIN holder.

## Features

*   **Checksum Validation:** Ensures the TIN is arithmetically correct.
*   **Birth Date Decoding:** Extracts the holder's birth date from the TIN.
*   **Sex Determination:** Determines the holder's sex (male/female).
*   **Plausibility Checks:** Verifies that the decoded birth date is within a reasonable range.
*   **Reusable Client:** Configure a validation client once and reuse it throughout your application.
*   **Extensible Rules:** Add your own custom validation logic.
*   **Zero Dependencies:** Pure Go, no external libraries needed.

## Installation

```bash
go get github.com/stremovskyy/uatins
```

## Usage

### Basic Validation

Create a reusable client and use its `Validate` method to check a TIN. You can optionally provide a date of birth to verify it against the date encoded in the number.

```go
package main

import (
	"fmt"
	"github.com/stremovskyy/uatins"
	"time"
)

func main() {
	// Create a client with default settings.
	// It's recommended to create this once and reuse it.
	validator := uatins.NewClient()

	tin := "3036045681"
	dob := time.Date(1983, 2, 14, 0, 0, 0, 0, time.UTC)

	// Validate the TIN, optionally providing the date of birth to check for a match.
	res, err := validator.Validate(tin, &dob)
	if err != nil {
		// This typically happens for structural errors, like invalid length or non-digit characters.
		fmt.Printf("Validation error: %v\n", err)
		return
	}

	if res.Valid {
		fmt.Printf("TIN %s is valid!\n", tin)
		fmt.Printf("Sex: %s\n", res.Sex)
		fmt.Printf("Birth Date: %s\n", res.BirthDate.Format("2006-01-02"))
		fmt.Printf("Checksum OK: %t\n", res.ChecksumOK)
		fmt.Printf("Provided DOB Matched: %t\n", res.DOBMatched)
	} else {
		fmt.Printf("TIN %s is invalid.\n", tin)
	}
}
```

### Advanced Configuration

The client can be configured with functional options to customize its behavior.

```go
import (
    "github.com/stremovskyy/uatins"
    "time"
)

// Create a client with custom settings.
validator := uatins.NewClient(
    // Require the provided DOB to match the one encoded in the TIN.
    // If they don't match, Validate will return an ErrDOBMismatch error.
    uatins.WithStrict(true),

    // Return the decoded birth date in the local time zone.
    uatins.WithLocation(time.Local),

    // Set a maximum plausible age for the TIN holder (default is 130).
    uatins.WithMaxAge(120),
)

// Use the configured client...
// tin := "3036045681"
// dob := time.Date(1983, 2, 15, 0, 0, 0, 0, time.UTC) // Wrong DOB
// res, err := validator.Validate(tin, &dob) 
// -> err: tin: provided DOB does not match encoded date
```

### Custom Validation Rules

You can extend the validator with your own rules. A rule is a simple function that accepts the TIN string and returns an error if validation fails.

```go
import (
    "fmt"
    "strconv"
    "github.com/stremovskyy/uatins"
)

// Forbid TINs issued in a specific year.
var forbid1999 = uatins.Rule[string](func(s string) error {
    days, _ := strconv.Atoi(s[:5])
    date := uatins.DaysToDate(days)
    if date.Year() == 1999 {
        return fmt.Errorf("TINs from 1999 are not allowed")
    }
    return nil
})

// Create a client with the custom rule.
validator := uatins.NewClient(
    uatins.WithRules(uatins.Rules[string]{forbid1999}),
)

// This TIN has a birth date of 1999-12-31 and will fail validation.
res, err := validator.Validate("3652412345", nil)
// -> err: TINs from 1999 are not allowed
```

## Error Handling

The `Validate` method returns a custom error type that you can inspect. Use `errors.Is` to check against the exported error variables (`ErrLength`, `ErrNonDigit`, `ErrDOBMismatch`, etc.).

```go
import (
    "errors"
    "github.com/stremovskyy/uatins"
)

_, err := validator.Validate("12345", nil)
if errors.Is(err, uatins.ErrLength) {
    fmt.Println("The TIN has an invalid length.")
}
```

## Running Tests

To run the full suite of tests:

```bash
go test ./...
```

## Benchmarks

To run the benchmarks:

```bash
go test -bench=. -benchmem
```

## License

MIT

# LICENSE

MIT License

Copyright (c) 2025 Megakit Systems

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
SOFTWARE.
