SET TIMEZONE = 'Asia/Almaty';

CREATE TYPE status_enum AS ENUM ('CREATED','REVERSED','APPROVED','DEPOSITED','DECLINED','REFUNDED');
CREATE TYPE operation_enum AS ENUM ('URL_payment');

CREATE TABLE Transactions (
    Payment_id VARCHAR(256) PRIMARY KEY,
    User_id VARCHAR(256) NOT NULL,
    Order_id VARCHAR(256) NOT NULL UNIQUE,
    Amount NUMERIC(18,2) NOT NULL,
    Currency CHAR(3) NOT NULL, 
    Broker VARCHAR(100) NOT NULL,
    Operation operation_enum NOT NULL,
    Current_status status_enum NOT NULL DEFAULT 'CREATED',
    Created_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE TABLE TransactionStatus (
    Payment_id VARCHAR(256) NOT NULL REFERENCES Transactions(Payment_id) ON DELETE CASCADE,
    Created_at TIMESTAMPTZ DEFAULT NOW(),
    Status status_enum NOT NULL,
    PRIMARY KEY (Payment_id, Created_at)
);

CREATE TABLE Refunds (
    Refund_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    Payment_id VARCHAR(256) REFERENCES Transactions(Payment_id) ON DELETE CASCADE,
    Amount NUMERIC(18,2) NOT NULL,
    Reason TEXT,
    Created_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_transactions_user ON Transactions(User_id);
CREATE INDEX idx_status_created_at ON TransactionStatus(Created_at DESC);
