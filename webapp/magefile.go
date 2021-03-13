//go:build mage
// +build mage

package main

import (
	"os"

	"github.com/magefile/mage/sh"
)

func Test() error {

	var err error

	ternEnvars := make(map[string]string)
	ternEnvars["PGHOST"] = "127.0.0.1"

	if os.Getenv("CI") != "true" {
		err = sh.RunWith(ternEnvars, "tern", "migrate", "-m", "../migrations/", "-c", "../migrations/tern.conf", "-d", "0")
		if err != nil {
			return err
		}
	}

	err = sh.RunWith(ternEnvars, "tern", "migrate", "-m", "../migrations/", "-c", "../migrations/tern.conf")
	if err != nil {
		return err
	}

	return sh.Run("gotestsum", "--junitfile=unit-tests.xml", "--", "-coverprofile=coverage.txt", "-covermode=atomic", "./...")
}
