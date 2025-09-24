package uatins

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"
)

// Sex indicates the gender extracted from the TIN.
type Sex string

const (
	Male   Sex = "male"
	Female Sex = "female"
)

// Result holds parsed information about a TIN.
type Result struct {
	TIN                string
	BirthDate          time.Time
	Sex                Sex
	ChecksumOK         bool
	BirthDatePlausible bool
	DOBMatched         bool
	Valid              bool
}

// Custom errors for various validation failures.
var (
	ErrLength          = errors.New("tin: invalid length")
	ErrNonDigit        = errors.New("tin: contains non-digit")
	ErrAllSame         = errors.New("tin: all digits identical")
	ErrChecksum        = errors.New("tin: checksum failed")
	ErrBirthOutOfRange = errors.New("tin: birth date not plausible")
	ErrDOBMismatch     = errors.New("tin: provided DOB does not match encoded date")
	ErrUnknown         = errors.New("tin: unknown error")
)

// Error contains context for validation errors.
type Error struct {
	Code        string
	TIN         string
	Msg         string
	DecodedDOB  *time.Time
	ProvidedDOB *time.Time
}

func (e *Error) Error() string {
	if e.Msg != "" {
		return e.Msg
	}
	return e.Code
}

func (e *Error) Is(target error) bool {
	switch target {
	case ErrLength, ErrNonDigit, ErrAllSame, ErrChecksum, ErrBirthOutOfRange, ErrDOBMismatch:
		return e.Code == target.Error()
	default:
		return false
	}
}

// wrapErr constructs a detailed Error from a sentinel.
func wrapErr(sentinel error, tin string, msg string, dec, prov *time.Time) *Error {
	return &Error{
		Code:        sentinel.Error(),
		TIN:         tin,
		Msg:         msg,
		DecodedDOB:  dec,
		ProvidedDOB: prov,
	}
}

// Rule represents a generic validation rule for any type.
type Rule[T any] func(T) error

// Rules is a slice of Rule values with helper methods.
type Rules[T any] []Rule[T]

// Add appends rules and enables chaining.
func (r Rules[T]) Add(more ...Rule[T]) Rules[T] {
	return append(r, more...)
}

// Validate runs rules until the first error is returned.
func (r Rules[T]) Validate(v T) error {
	for _, rule := range r {
		if err := rule(v); err != nil {
			return err
		}
	}
	return nil
}

// Client is a reusable TIN validator.
type Client struct {
	now         time.Time
	maxAgeYears int
	strict      bool
	loc         *time.Location
	custom      Rules[string]
}

// NewClient returns a new Client with sane defaults.
func NewClient(opts ...Option) *Client {
	c := &Client{
		now:         time.Now().UTC(),
		maxAgeYears: 130,
		loc:         time.UTC,
	}
	for _, opt := range opts {
		opt(c)
	}
	return c
}

// Option configures a Client.
type Option func(*Client)

// WithMaxAge sets an age cap; 0 disables the cap.
func WithMaxAge(years int) Option {
	return func(c *Client) {
		c.maxAgeYears = years
	}
}

// WithStrict enforces DOB mismatch as a validation error.
func WithStrict(on bool) Option {
	return func(c *Client) {
		c.strict = on
	}
}

// WithLocation sets the time zone used to expose the BirthDate.
func WithLocation(loc *time.Location) Option {
	return func(c *Client) {
		if loc != nil {
			c.loc = loc
		}
	}
}

// WithRules allows callers to extend or override rules.
func WithRules(r Rules[string]) Option {
	return func(c *Client) {
		c.custom = r
	}
}

// WithNow overrides the current time (useful for tests).
func WithNow(t time.Time) Option {
	return func(c *Client) {
		c.now = t.In(time.UTC)
	}
}

// MaxAge sets an age cap; 0 disables the cap. Returns the client for chaining.
func (c *Client) MaxAge(years int) *Client {
	c.maxAgeYears = years
	return c
}

// Strict enforces DOB mismatch as a validation error. Returns the client for chaining.
func (c *Client) Strict(on bool) *Client {
	c.strict = on
	return c
}

// Location sets the time zone used to expose the BirthDate. Returns the client for chaining.
func (c *Client) Location(loc *time.Location) *Client {
	if loc != nil {
		c.loc = loc
	}
	return c
}

// Rules allows callers to extend or override rules. Returns the client for chaining.
func (c *Client) Rules(r Rules[string]) *Client {
	c.custom = r
	return c
}

// Now overrides the current time (useful for tests). Returns the client for chaining.
func (c *Client) Now(t time.Time) *Client {
	c.now = t.In(time.UTC)
	return c
}

// Validate runs all checks and returns a Result and an error (if any).
func (c *Client) Validate(tin string, providedDOB *time.Time) (Result, error) {
	var res Result
	tin = digitsOnly(tin)
	res.TIN = tin

	// Core string rules: non-digit first, then length, then all-same.
	var core Rules[string]
	core = core.Add(
		ruleAllDigits(),  // ensure only digits first
		ruleLength(10),   // enforce exact length next
		ruleNotAllSame(), // disallow all-same digits or all zeros
	)
	if err := core.Validate(tin); err != nil {
		return res, err
	}

	// Custom rules, if any.
	if c.custom != nil {
		if err := c.custom.Validate(tin); err != nil {
			return res, err
		}
	}

	// Decode DOB from digits 1..5 and sex from digit 9.
	ddays, _ := strconv.Atoi(tin[:5])
	utcDOB := DaysToDate(ddays)
	res.BirthDate = utcDOB.In(c.loc)

	if int(tin[8]-'0')%2 == 0 {
		res.Sex = Female
	} else {
		res.Sex = Male
	}

	// Check if the birth date is plausible.
	if !IsBirthDatePlausible(utcDOB, c.now, c.maxAgeYears) {
		return res, wrapErr(
			ErrBirthOutOfRange, tin,
			"encoded birth date out of plausible range", &utcDOB, providedDOB,
		)
	}
	res.BirthDatePlausible = true

	// Compute the checksum result. Do not return an error if it fails;
	// just set res.Valid accordingly below.
	res.ChecksumOK = ChecksumOK(tin)

	// Compare provided DOB if supplied.
	if providedDOB != nil {
		res.DOBMatched = sameYMD(utcDOB, providedDOB.In(time.UTC))
		if c.strict && !res.DOBMatched {
			return res, wrapErr(
				ErrDOBMismatch, tin,
				"provided DOB does not match encoded date",
				&utcDOB, providedDOB,
			)
		}
	} else {
		// If strict mode is on but no DOB is given, there is no mismatch.
		res.DOBMatched = true
	}

	// TIN is valid only if checksum matches, DOB plausible,
	// and (if strict, the provided DOB matches).
	res.Valid = res.ChecksumOK && res.BirthDatePlausible
	if c.strict && providedDOB != nil {
		res.Valid = res.Valid && res.DOBMatched
	}
	return res, nil
}

// --- Rule implementations ---

// ruleLength ensures a string has exactly n characters.
func ruleLength(n int) Rule[string] {
	return func(s string) error {
		if len(s) != n {
			return wrapErr(
				ErrLength, s,
				fmt.Sprintf("need %d digits", n),
				nil, nil,
			)
		}
		return nil
	}
}

// ruleAllDigits ensures all characters are digits 0–9.
func ruleAllDigits() Rule[string] {
	return func(s string) error {
		for i := 0; i < len(s); i++ {
			if s[i] < '0' || s[i] > '9' {
				return wrapErr(
					ErrNonDigit, s,
					"only digits allowed",
					nil, nil,
				)
			}
		}
		return nil
	}
}

// ruleNotAllSame disallows TINs with all identical digits or all zeros.
func ruleNotAllSame() Rule[string] {
	return func(s string) error {
		if s == "" {
			return wrapErr(ErrLength, s, "empty", nil, nil)
		}
		all := true
		for i := 1; i < len(s); i++ {
			if s[i] != s[0] {
				all = false
				break
			}
		}
		if all || s == "0000000000" {
			return wrapErr(
				ErrAllSame, s,
				"implausible: all digits identical or zero",
				nil, nil,
			)
		}
		return nil
	}
}

// ruleChecksum is kept for external use; it returns ErrChecksum if checksum fails.
func ruleChecksum() Rule[string] {
	return func(s string) error {
		if !ChecksumOK(s) {
			return wrapErr(
				ErrChecksum, s,
				"checksum mismatch",
				nil, nil,
			)
		}
		return nil
	}
}

// --- Helpers ---

// digitsOnly filters non-digit characters from a string.
func digitsOnly(s string) string {
	var b strings.Builder
	b.Grow(len(s))
	for i := 0; i < len(s); i++ {
		c := s[i]
		if c >= '0' && c <= '9' {
			b.WriteByte(c)
		}
	}
	return b.String()
}

// ChecksumOK implements the official RNOKPP checksum using
// weights [-1,5,7,9,4,6,10,5,7], computing ctrl=((sum mod 11) mod 10).
func ChecksumOK(tin string) bool {
	if len(tin) != 10 {
		return false
	}
	weights := [...]int{-1, 5, 7, 9, 4, 6, 10, 5, 7}
	sum := 0
	for i := 0; i < 9; i++ {
		sum += int(tin[i]-'0') * weights[i]
	}
	ctrl := sum % 11
	if ctrl < 0 {
		ctrl += 11
	}
	return (ctrl % 10) == int(tin[9]-'0')
}

// DaysToDate converts days since 1899-12-31 to UTC midnight.
func DaysToDate(days int) time.Time {
	base := time.Date(1899, 12, 31, 0, 0, 0, 0, time.UTC)
	return base.AddDate(0, 0, days)
}

// DecodeDOBFromTIN extracts the encoded birth date from a TIN.
func DecodeDOBFromTIN(tin string) (time.Time, error) {
	if len(tin) < 5 {
		return time.Time{}, fmt.Errorf("tin too short")
	}
	dd, err := strconv.Atoi(tin[:5])
	if err != nil {
		return time.Time{}, err
	}
	return DaysToDate(dd), nil
}

// IsBirthDatePlausible checks that the date is within a plausible range.
func IsBirthDatePlausible(d, now time.Time, maxAgeYears int) bool {
	if d.IsZero() {
		return false
	}
	min := time.Date(1900, 1, 1, 0, 0, 0, 0, time.UTC)
	if d.Before(min) || d.After(now) {
		return false
	}
	if maxAgeYears > 0 && d.Before(now.AddDate(-maxAgeYears, 0, 0)) {
		return false
	}
	return true
}

// sameYMD compares two dates by year, month, and day only.
func sameYMD(a, b time.Time) bool {
	ay, am, ad := a.Date()
	by, bm, bd := b.Date()
	return ay == by && am == bm && ad == bd
}
