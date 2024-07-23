# Поднять контейнеры Docker
up:
	docker-compose up -d

# Остановить контейнеры Docker
down:
	docker-compose down

create-migration:
	goose -dir ./db/migrations create orders sql 

# Выполнить миграции
migrate:
	goose -dir ./db/migrations postgres "host=localhost port=5432 user=postgres password=examplepassword dbname=oms sslmode=disable" up

# Выполнить SQL-запрос
# Пример использования: make query SQL="SELECT * FROM orders"
query:
	docker exec -it postgres psql -U postgres -d oms -c "$(SQL)"

select-all-orders:
	docker exec -it postgres psql -U postgres -d oms -c "SELECT * FROM orders"

select-all-outbox:
	docker exec -it postgres psql -U postgres -d oms -c "SELECT * FROM outbox"

# Запуск интеграционных тестов
integration-test:
	go test -tags=integration -run ^TestFindOrder_NotFound$$ ./tests
	go test -tags=integration -run ^TestFindOrder$$ ./tests
	go test -tags=integration -run ^TestAddOrder$$ ./tests
	go test -tags=integration -run ^TestRemoveOrder$$ ./tests
	go test -tags=integration -run ^TestIssueOrder$$ ./tests
	go test -tags=integration -run ^TestReturnOrder$$ ./tests

# Запуск Unit тестов
unit-test:
	go test ./...

create-topic:
	docker-compose exec kafka /usr/bin/kafka-topics --bootstrap-server localhost:9092 --topic logs --create --partitions 3 --replication-factor 1

consume:
	docker-compose exec kafka /usr/bin/kafka-console-consumer --bootstrap-server localhost:9092 --topic logs

run-service:
	go run cmd/grpc/service/main.go

run-client:
	go run cmd/grpc/client/main.go

moc-gen:
	mockgen -source=internal/api/service.go -destination=internal/api/mocks/module_mock.go -package=mocks


# Используем bin в текущей директории для установки плагинов protoc
LOCAL_BIN:=$(CURDIR)/bin

# Добавляем bin в текущей директории в PATH при запуске protoc
PROTOC = PATH="$$PATH:$(LOCAL_BIN)" protoc

PICKPOINT_PROTO_PATH:="api/proto/pickpoint/v1"

# Установка всех необходимых зависимостей
.PHONY: .bin-deps
.bin-deps:
	$(info Installing binary dependencies...)

	GOBIN=$(LOCAL_BIN) go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
	GOBIN=$(LOCAL_BIN) go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
	GOBIN=$(LOCAL_BIN) go install github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-grpc-gateway@latest
	GOBIN=$(LOCAL_BIN) go install github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-openapiv2@latest
	GOBIN=$(LOCAL_BIN) go install github.com/envoyproxy/protoc-gen-validate@latest


# Вендоринг внешних proto файлов
.vendor-proto: vendor-proto/google/protobuf vendor-proto/google/api vendor-proto/protoc-gen-openapiv2/options vendor-proto/validate

# Устанавливаем proto описания protoc-gen-openapiv2/options
vendor-proto/protoc-gen-openapiv2/options:
	git clone -b main --single-branch -n --depth=1 --filter=tree:0 \
 		https://github.com/grpc-ecosystem/grpc-gateway vendor.proto/grpc-ecosystem && \
 	cd vendor.proto/grpc-ecosystem && \
	git sparse-checkout set --no-cone protoc-gen-openapiv2/options && \
	git checkout
	mkdir -p vendor.proto/protoc-gen-openapiv2
	mv vendor.proto/grpc-ecosystem/protoc-gen-openapiv2/options vendor.proto/protoc-gen-openapiv2
	rm -rf vendor.proto/grpc-ecosystem


# Устанавливаем proto описания google/protobuf
vendor-proto/google/protobuf:
	git clone -b main --single-branch -n --depth=1 --filter=tree:0 \
		https://github.com/protocolbuffers/protobuf vendor.proto/protobuf &&\
	cd vendor.proto/protobuf &&\
	git sparse-checkout set --no-cone src/google/protobuf &&\
	git checkout
	mkdir -p vendor.proto/google
	mv vendor.proto/protobuf/src/google/protobuf vendor.proto/google
	rm -rf vendor.proto/protobuf

vendor-proto/google/api:
	git clone -b master --single-branch -n --depth=1 --filter=tree:0 \
 		https://github.com/googleapis/googleapis vendor.proto/googleapis && \
 	cd vendor.proto/googleapis && \
	git sparse-checkout set --no-cone google/api && \
	git checkout
	mkdir -p  vendor.proto/google
	mv vendor.proto/googleapis/google/api vendor.proto/google
	rm -rf vendor.proto/googleapis

vendor-proto/validate:
	git clone -b main --single-branch --depth=2 --filter=tree:0 \
		https://github.com/bufbuild/protoc-gen-validate vendor.proto/tmp && \
		cd vendor.proto/tmp && \
		git sparse-checkout set --no-cone validate &&\
		git checkout
		mkdir -p vendor.proto/validate
		mv vendor.proto/tmp/validate vendor.proto/
		rm -rf vendor.proto/tmp

.PHONY: generate
generate:
	rm -rf ./vendor.proto
	make .generate

.generate: .bin-deps .vendor-proto
	mkdir -p pkg/${PICKPOINT_PROTO_PATH}
	protoc -I api/proto \
		-I vendor.proto \
		${PICKPOINT_PROTO_PATH}/pickpoint.proto \
		--plugin=protoc-gen-go=$(LOCAL_BIN)/protoc-gen-go --go_out=./pkg/${PICKPOINT_PROTO_PATH} --go_opt=paths=source_relative\
		--plugin=protoc-gen-go-grpc=$(LOCAL_BIN)/protoc-gen-go-grpc --go-grpc_out=./pkg/${PICKPOINT_PROTO_PATH} --go-grpc_opt=paths=source_relative --experimental_allow_proto3_optional\
		--plugin=protoc-gen-grpc-gateway=$(LOCAL_BIN)/protoc-gen-grpc-gateway --grpc-gateway_out ./pkg/api/proto/pickpoint/v1  --grpc-gateway_opt  paths=source_relative --grpc-gateway_opt generate_unbound_methods=true \
		--plugin=protoc-gen-openapiv2=$(LOCAL_BIN)/protoc-gen-openapiv2 --openapiv2_out=./docs \
		--plugin=protoc-gen-validate=$(LOCAL_BIN)/protoc-gen-validate --validate_out="lang=go,paths=source_relative:pkg/api/proto/pickpoint/v1"

run-prometheus:
	./prometheus-2.43.0.linux-amd64/prometheus --config.file=prometheus.yml
