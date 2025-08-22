# Сервис Платежей

Сервис для обработки **создания платежей, авторизации, депозита, возврата, реверса** и отслеживания статусов. Поддерживает работу с разными брокерами и банками.

---

## Возможности

✅ Создание, авторизация, депозит, возврат и реверс платежей  
✅ Отслеживание статуса и истории платежей  
✅ REST & gRPC API через gRPC-Gateway  
✅ Эндпоинт HealthCheck для проверки базы данных и брокера  
✅ Поддержка произвольных метаданных для платежей  

---

## Обзор API

- **Base URL:** `http://localhost:8080`  
- **Secured URL:** `https://localhost:8080`  
- **Спецификация:** OpenAPI 3.0 (можно сгенерировать из proto через grpc-gateway)

### Основные эндпоинты

| Метод | Эндпоинт                        | Описание                                    |
|-------|---------------------------------|---------------------------------------------|
| POST  | `/v1/payments`                  | Создание платежа                            |
| POST  | `/v1/payments/auth`             | Авторизация платежа                         |
| POST  | `/v1/payments/deposit`          | Депозит (списание средств)                  |
| POST  | `/v1/payments/refund`           | Возврат средств                             |
| POST  | `/v1/payments/reversal`         | Реверс платежа (отмена или ошибка)         |
| GET   | `/v1/payments/{payment_id}`     | Получение информации о платеже             |
| GET   | `/v1/payments/{payment_id}/status` | Получение текущего статуса платежа      |
| POST  | `/v1/payments/success`          | Пометить платеж как успешный               |
| GET   | `/v1/payments`                  | Список платежей (с пагинацией)             |
| GET   | `/v1/health`                    | Проверка состояния сервиса                  |

---

## Настройка

### 1️⃣ Настройка PostgreSQL

Создайте базу данных PostgreSQL, соответствующую настройкам в `.env` перед запуском сервиса.

---

### 2️⃣ Файл конфигурации `.env`

Пример `.env`:

```env
# Конфигурация сервера
GRPC_MAX_MESSAGE_SIZE_MIB=12
GRPC_MAX_CONNECTION_AGE=30s
GRPC_MAX_CONNECTION_AGE_GRACE=10s
GRPC_PORT=5433
LEVEL=debug # debug | prod | dev

# Настройки базы данных
DB_HOST=localhost
DB_PORT=5432
DB_NAME=paymentServiceDB
DB_USER=Bacoonti
DB_PASSWORD=SuperSecretPassword
POSTGRES_MAX_OPEN_CONN=25
POSTGRES_MAX_IDLE_TIME=15m

# API Банка Bereke
BEREKE_MERCHANT_LOGIN=SuperSecretLogin
BEREKE_MERCHANT_PASSWORD=SuperSecretPassword
BEREKE_MERCHANT_MODE=TEST
```


##  🎨 Визуализация процесса оплаты
Ниже приведена последовательность действий между клиентом, сервисом, банком (Merchant Adapter) и брокером платежей. Диаграммы разделены на три фазы.

**Фаза 1: Предавторизация заказа:**

```mermaid
sequenceDiagram
    actor Client as Client (User)
    participant Service as Service (Shop Backend)
    participant BMA as Merchant (Bank Merchant Adapter)
    participant Broker as Broker (Payment Broker)

    Note over Client,Broker: Phase 1: Order Pre-Authorization

    Client->>Service: (1) Create Order Request <br> Product_ID=123, qty=2
    Service->>Service: Calculate amount = 2000 KZT
    Service->>BMA: (2) Create Order PreAuth <br> {amount=2000, currency=KZT, orderId=ORD-1001}
    BMA->>Broker: (3) POST /registerPreAuth.do <br> payload: {amount, currency, orderId}
    Broker-->>BMA: (4) Response 200 OK <br> {transactionId=TX-abc123, paymentUrl=[https://pay.kz/](https://pay.kz/)...}
    BMA-->>Service: (5) Return PreAuth Result
    Service-->>Client: (6) Provide paymentUrl to redirect
    Client->>Broker: (7) Redirect to paymentUrl (card entry, 3DS, OTP)
    Broker-->>Client: (8) Payment Success Screen
    Client->>Service: (9) Callback/redirect success?orderId=ORD-1001
    Service->>BMA: (10) Verify transaction status
    BMA->>Broker: (11) GET /status.do?tx=TX-abc123
    Broker-->>BMA: (12) status=AUTHORIZED
    BMA-->>Service: (13) status=AUTHORIZED
    Service-->>Client: (14) Order status updated to AUTHORIZED
```

**Фаза 2: Депозит заказа (списание средств)**

```mermaid
sequenceDiagram
    actor Client as Client (User)
    participant Service as Service (Shop Backend)
    participant BMA as Merchant (Bank Merchant Adapter)
    participant Broker as Broker (Payment Broker)

    Note over Client,Broker: Phase 2: Order Deposition (Capture funds)

    Service-->>Client: (1) Provide Product/Service ✅
    Service->>BMA: (2) Deposit Order <br> {transactionId=TX-abc123, amount=2000}
    BMA->>Broker: (3) POST /deposit.do
    Broker-->>BMA: (4) Response 200 OK <br> {status=DEPOSITED}
    BMA-->>Service: (5) Order deposited
    Service-->>Client: (6) Notify payment completed
```

**Фаза 3: Реверсирование заказа (отмена или сбой)**

```mermaid
sequenceDiagram
    actor Client as Client (User)
    participant Service as Service (Shop Backend)
    participant BMA as Merchant (Bank Merchant Adapter)
    participant Broker as Broker (Payment Broker)

    Note over Client,Broker: Phase 3: Order Reversal (Cancel or failure)

    Service-->>Client: (1) Failure occurred ❌
    Service->>BMA: (2) Reversal Order <br> {transactionId=TX-abc123}
    BMA->>Broker: (3) POST /reverse.do
    Broker-->>BMA: (4) Response 200 OK <br> {status=REVERSED}
    BMA-->>Service: (5) Reversal confirmed
    Service-->>Client: (6) Order has been reversed
```