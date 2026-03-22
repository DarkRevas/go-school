PROTOC := $(shell which protoc 2>/dev/null || echo "$(PWD)/.protoc/bin/protoc")
PROTO_PATH := proto
PROTO_FILES := $(wildcard $(PROTO_PATH)/*.proto)
GO_OUT := .

.PHONY: proto
proto:
	$(PROTOC) --go_out=$(GO_OUT) --go_opt=paths=source_relative \
		--go-grpc_out=$(GO_OUT) --go-grpc_opt=paths=source_relative \
		-I=$(PROTO_PATH) $(PROTO_FILES)

.PHONY: clean
clean:
	rm -rf $(PROTO_PATH)/*.pb.go $(PROTO_PATH)/*_grpc.pb.go

.PHONY: generate
generate: proto
