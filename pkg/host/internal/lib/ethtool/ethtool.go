package ethtool

import (
	"github.com/safchain/ethtool"
)

func New() EthtoolLib {
	return &libWrapper{}
}

//go:generate ../../../../../bin/mockgen -destination mock/mock_ethtool.go -source ethtool.go
type EthtoolLib interface {
	// Features retrieves features of the given interface name.
	Features(intf string) (map[string]bool, error)
	// FeatureNames shows supported features by their name.
	FeatureNames(intf string) (map[string]uint, error)
	// Change requests a change in the given device's features.
	Change(intf string, config map[string]bool) error
}

type libWrapper struct{}

// Features retrieves features of the given interface name.
func (w *libWrapper) Features(intf string) (map[string]bool, error) {
	e, err := ethtool.NewEthtool()
	if err != nil {
		return nil, err
	}
	defer e.Close()
	return e.Features(intf)
}

// FeatureNames shows supported features by their name.
func (w *libWrapper) FeatureNames(intf string) (map[string]uint, error) {
	e, err := ethtool.NewEthtool()
	if err != nil {
		return nil, err
	}
	defer e.Close()
	return e.FeatureNames(intf)
}

// Change requests a change in the given device's features.
func (w *libWrapper) Change(intf string, config map[string]bool) error {
	e, err := ethtool.NewEthtool()
	if err != nil {
		return err
	}
	defer e.Close()
	return e.Change(intf, config)
}
