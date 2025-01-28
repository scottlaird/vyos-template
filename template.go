//go:build wasm

package main

import (
        "fmt"
        "syscall/js"
        "time"

        "github.com/scottlaird/vyos-parser/configmodel"
        "github.com/scottlaird/vyos-parser/parser"
        "github.com/scottlaird/vyos-parser/syntax"
        "honnef.co/go/js/dom/v2"
)

var (
        configModel *configmodel.VyOSConfigNode
)

func main() {
        // Avoid exiting so the Go runtime is still available from JS.
        done := make(chan struct{}, 0)
        var err error
        configModel, err = syntax.GetDefaultConfigModel()
        if err != nil {
                fmt.Printf("Error!: %v", err)
        }

        // Register Go functions with Javascript
        global := js.Global()
        global.Set("updatetarget", js.FuncOf(updatetarget))

        <-done
}

// updatetarget writes to InnerHTML on the .target element.
func updatetarget(this js.Value, args []js.Value) any {
        t := time.Now()
        
        fmt.Printf("Converting...\n")
        d := dom.GetWindow().Document()
        sourceElement := d.GetElementByID("source").(*dom.HTMLTextAreaElement)
        source := sourceElement.Value()

        fmt.Printf("Read: %s\n", source)

        ast, err := parser.ParseShowFormat(source, configModel)
        if err != nil {
                fmt.Printf("Error!: %v", err)
        }

        set, err := parser.WriteSetFormat(ast)
        setElement := d.GetElementByID("targetSet")
        setElement.SetInnerHTML(set)

        show, err := parser.WriteShowFormat(ast)
        showElement := d.GetElementByID("targetShow")
        showElement.SetInnerHTML(show)

        boot, err := parser.WriteConfigBootFormat(ast)
        bootElement := d.GetElementByID("targetBoot")
        bootElement.SetInnerHTML(boot)

        elapsed := time.Since(t).Milliseconds()
        fmt.Printf("Converted in %d ms\n", elapsed)
        
        return nil
}
