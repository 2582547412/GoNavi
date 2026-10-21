package plugin

import (
	"errors"
	"regexp"
	"strings"
)

var dataSourceID = regexp.MustCompile(`^[a-z][a-z0-9_.-]{0,63}$`)

// ValidateDataSource checks the public read-only declaration. It does not grant writes
// or prove that the command is declared on a backend.
func ValidateDataSource(source DataSource) error {
	if !source.ReadOnly {
		return errors.New("data source must declare readOnly true")
	}
	if source.Dialect != "mysql" && source.Dialect != "postgres" && source.Dialect != "sqlite" {
		return errors.New("unsupported data source dialect")
	}
	if source.DefaultPort < 0 || source.DefaultPort > 65535 {
		return errors.New("invalid data source port")
	}
	if !dataSourceID.MatchString(source.ID) || strings.TrimSpace(source.Name) == "" || len(source.Name) > 256 || strings.ContainsRune(source.Name, 0) {
		return errors.New("invalid data source identity")
	}
	return nil
}

// ValidateDataSourceCommand checks that the referenced command is the same one and does not advertise writes.
func ValidateDataSourceCommand(source DataSource, command Command) error {
	if err := ValidateDataSource(source); err != nil {
		return err
	}
	if source.Command == "" || source.Command != command.ID {
		return errors.New("data source command mismatch")
	}
	if command.Effect != "" && command.Effect != "read" && command.Effect != "unknown" {
		return errors.New("data source command effect must be read or unknown")
	}
	return nil
}
