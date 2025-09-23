package uatins_test_test

import (
	"fmt"
	"time"

	"github.com/stremovskyy/uatins"
)

func Example() {
	dob := time.Date(1983, 2, 14, 13, 0, 0, 0, time.UTC)

	client := uatins.NewClient(
		uatins.WithMaxAge(130),
		uatins.WithStrict(true),
		uatins.WithLocation(time.Local),
	)

	res, err := client.Validate("3036045681", &dob)

	if err != nil {
		fmt.Println("err:", err)
		return
	}

	fmt.Println(res.Valid, res.Sex)
	// Output:
	// true female
}
