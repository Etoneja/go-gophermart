package utils

import (
	"fmt"
	"strconv"
	"strings"
)

const minLengh = 6
const maxLengh = 19

func LuhnCheck(s string) (bool, error) {
	strValue := strings.TrimSpace(string(s))
	if strValue == "" {
		return false, fmt.Errorf("bad luhn: empty value")
	}

	_, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		return false, fmt.Errorf("bad luhn: invalid integer format")
	}

	if len(strValue) < minLengh || len(strValue) > maxLengh {
		return false, fmt.Errorf("bad luhn: invalid len format")
	}

	sum := 0
	for i := range len(strValue) {
		digit := int(strValue[len(strValue)-1-i] - '0')

		if i%2 == 1 {
			digit *= 2
			if digit > 9 {
				digit = digit%10 + digit/10
			}
		}
		sum += digit
	}
	return sum%10 == 0, nil
}
