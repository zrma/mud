package pb

//go:generate go install google.golang.org/grpc/cmd/protoc-gen-go-grpc
//go:generate go install google.golang.org/protobuf/cmd/protoc-gen-go
//go:generate protoc --go_out=. --go_opt=paths=source_relative --go-grpc_out=. --go-grpc_opt=paths=source_relative mud.proto
