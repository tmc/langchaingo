install-compiler:
	go install github.com/nihei9/vartan/cmd/vartan@latest
	go install github.com/nihei9/vartan/cmd/vartan-go@latest
compile:
	vartan compile expr.vartan -o expr.json
resolve-conflict:
	vartan show expr-report.json
test:
	vartan test
generate-parser:
	vartan-go expr.json