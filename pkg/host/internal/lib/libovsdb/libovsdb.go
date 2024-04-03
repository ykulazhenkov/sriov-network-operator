package libovsdb

import (
	"github.com/ovn-org/libovsdb/client"
)

// Client is a wrapper for libovsdb.Client interface,
// it is required to generate mocks
//
//go:generate ../../../../../bin/mockgen -destination mock/mock_libovsdb.go -source libovsdb.go
type Client interface {
	client.Client
}

// Client is a wrapper for libovsdb.ConditionalAPI interface,
// it is required to generate mocks
type ConditionalAPI interface {
	client.ConditionalAPI
}
