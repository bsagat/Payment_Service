# Makefile

PROTO_DIR = docs
PROTO_INCLUDE = "C:/Users/sagat/Downloads/protoc-31.1-win64/include"
GO_OUT = internal/adapters/grpc

.PHONY: up down nuke run proto

up:
	docker-compose up --build -d

down:
	docker-compose down

nuke:
	docker-compose down -v

run:
	go run cmd/main.go

proto:
	protoc \
		--proto_path=$(PROTO_DIR) \
		--proto_path=$(PROTO_INCLUDE) \
		--go_out=$(GO_OUT) \
		--go-grpc_out=$(GO_OUT) \
		--grpc-gateway_out=$(GO_OUT) \
		$(PROTO_DIR)/payment.proto
