//go:build !linux

package firewall

import "errors"

var errNotImplemented = errors.New("not implemented for non-linux systems")

// AddOrUpdateRedirect updates the firewall using NFTables to redirect traffic from, to.
func AddOrUpdateRedirect(from int, to int) error {
	return errNotImplemented
}

// DeleteRules deletes any created rules by deleting the custom table created
func DeleteRules() error {
	return errNotImplemented
}
