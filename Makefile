.PHONY: gen-favorite
gen-favorite:
	@protoc --go_out=./rpc_gen --go-grpc_out=./rpc_gen ./idl/favorite.proto