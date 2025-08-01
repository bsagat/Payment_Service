SET TIMEZONE = 'Asia/Almaty';

CREATE TYPE status_enum as enum ('CREATED','DEPOSITED','REVERSED','failed');
CREATE TYPE operation as enum ('COF_payment','URL_payment');

CREATE TABLE Transactions(
    Payment_id VARCHAR(256) NOT NULL UNIQUE, 
    User_id VARCHAR(256) NOT NULL,
    Order_id VARCHAR(256) NOT NULL UNIQUE,
    Amount float NOT NULL,
    Currency VARCHAR(10) NOT NULL,
    Broker VARCHAR(256) NOT NULL,
    Operation operation NOT NULL
);

CREATE TABLE Status(
    Order_id VARCHAR(256) NOT NULL references Transactions(Order_id) ON DELETE CASCADE,
    Created_at TIMESTAMPTZ Default Now(),
    Status status_enum NOT NULL
);
