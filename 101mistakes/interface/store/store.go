package store

type CustomerStorage interface {
	StoreCustomer() error
	GetCustomer(Customer) error
	GetAllCustomers([]Customer) error
	UpdateCustomer() error
}

type Customer struct {
	name     string
	lastName string
	age      int
}
