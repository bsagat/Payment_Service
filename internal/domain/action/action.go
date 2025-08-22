package action

const (
	// Важный жизненный цикл сервиса
	ServiceSetup     = "service_setup"
	ServiceStarted   = "service_started"
	ServiceStartFail = "service_start_failed"
	GracefulShutdown = "service_shutdown"

	// Действия с сервером
	ServerStarted   = "server_started"
	ServerStartFail = "server_start_failed"
	ServerClosed    = "server_closed"

	// Взаимодействие сервиса
	DbConnected         = "db_connected"
	DbTransactionFailed = "db_transaction_failed"
	HealthCheck         = "health_check"

	PaymentReqReceived  = "payment_received"
	PaymentReqProcessed = "payment_processed"
	ValidationFailed    = "validation_failed"

	// CRUD операций с платежами
	CreatePayment    = "create_payment"
	GetPayment       = "get_payment"
	GetPaymentStatus = "get_payment_status"
	ListPayments     = "list_payments"

	// Изменение состояния платежа
	RefundPayment  = "refund_payment"
	SuccessPayment = "success_payment"
	AuthPayment    = "auth_payment"
	DepositPayment = "deposit_payment"
	ReversePayment = "reversal_payment"

	PaymentTransactionFail = "payment_broker_transaction_failed"
)
