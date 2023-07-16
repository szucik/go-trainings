package client

import "go-training-repo/101mistakes/interface/store"

type customersGetter interface {
	GetAllCustomers() ([]store.Customer, error)
}
