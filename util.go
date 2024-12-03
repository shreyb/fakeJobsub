package main

import "errors"

func checkSubmitForGroup(group string) error {
	if group == "" {
		return errors.New("failed submit group")
	}
	return nil
}
