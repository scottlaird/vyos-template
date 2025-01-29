ALL: html/template.wasm html/wasm_exec.js

template.wasm: template.go
	$(shell GOOS=js GOARCH=wasm go build -o template.wasm template.go)

template-tiny.wasm: template.go
	$(shell GOOS=js GOARCH=wasm tinygo build -o template-tiny.wasm template.go)

html/wasm_exec.js:
	$(shell cp "`go env GOROOT`/misc/wasm/wasm_exec.js" html/wasm_exec.js)

html/template.wasm: template.wasm
	cp template.wasm html/template.wasm
