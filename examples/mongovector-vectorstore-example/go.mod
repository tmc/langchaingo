module github.com/vendasta/langchaingo/examples/mongovector-vectorstore-example

go 1.24

toolchain go1.24.6

require (
	github.com/vendasta/langchaingo v0.1.14-pre.4
	go.mongodb.org/mongo-driver/v2 v2.0.0
)

require (
	github.com/dlclark/regexp2 v1.10.0 // indirect
	github.com/golang/snappy v1.0.0 // indirect
	github.com/google/uuid v1.6.0 // indirect
	github.com/klauspost/compress v1.18.0 // indirect
	github.com/pkoukk/tiktoken-go v0.1.8 // indirect
	github.com/xdg-go/pbkdf2 v1.0.0 // indirect
	github.com/xdg-go/scram v1.1.2 // indirect
	github.com/xdg-go/stringprep v1.0.4 // indirect
	github.com/youmark/pkcs8 v0.0.0-20240726163527-a2c0da244d78 // indirect
	golang.org/x/crypto v0.41.0 // indirect
	golang.org/x/sync v0.16.0 // indirect
	golang.org/x/text v0.28.0 // indirect
)

replace github.com/vendasta/langchaingo => ../..
