build:
	GOOS=wasip1 GOARCH=wasm go build -buildmode=c-shared -o structured-output.wasm ./cmd/structured-output

test:
	go test -v ./...

clean:
	rm -f structured-output.wasm
