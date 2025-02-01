ALL: template.wasm html/wasm_exec.js 

template.wasm: template.go
	$(shell GOOS=js GOARCH=wasm go build -ldflags="-s -w" -o template.wasm template.go)
	cp template.wasm html/template.wasm
	gzip -f9k html/template.wasm

html/wasm_exec.js:
	$(shell cp "`go env GOROOT`/misc/wasm/wasm_exec.js" html/wasm_exec.js)
