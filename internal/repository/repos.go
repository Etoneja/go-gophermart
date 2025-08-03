package repository

type Repositories struct {
	OrderRepo       *OrderRepository
	UserRepo        *UserRepository
	TransactionRepo *TransactionRepository
}

func NewRepositories() *Repositories {
	return &Repositories{
		OrderRepo:       NewOrderRepository(),
		UserRepo:        NewUserRepository(),
		TransactionRepo: NewTransactionRepository(),
	}
}
