package utils

import (
	"errors"
)

func Error(prefix string) func(string) error {
	return func(s string) error {
		return errors.New(prefix + s)
	}
}
