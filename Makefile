.PHONY: gen-favorite
gen-favorite:
	@protoc --go_out=./api/rpc_gen --go-grpc_out=./api/rpc_gen ./api/favorite.proto