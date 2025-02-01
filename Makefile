ALL: template.wasm html/wasm_exec.js html/wasm_exec_tiny.js

tinygo: template-tiny.wasm html/wasm_exec_tiny.js

template.wasm: template.go
	$(shell GOOS=js GOARCH=wasm go build -ldflags="-s -w" -o template.wasm template.go)
	cp template.wasm html/template.wasm
	gzip -f9k html/template.wasm

template-tiny.wasm: template.go
	$(shell GOOS=js GOARCH=wasm tinygo build -o template-tiny.wasm .)
	cp template-tiny.wasm html/

html/wasm_exec.js:
	$(shell cp "`go env GOROOT`/misc/wasm/wasm_exec.js" html/wasm_exec.js)

html/wasm_exec_tiny.js:
	cp /usr/local/lib/tinygo/targets/wasm_exec.js html/wasm_exec_tiny.js
