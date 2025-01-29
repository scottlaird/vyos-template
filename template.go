//go:build wasm

package main

import (
        "embed"
        "fmt"
        "syscall/js"
        "time"
        "bytes"

        "github.com/scottlaird/vyos-parser/configmodel"
        "github.com/scottlaird/vyos-parser/parser"
        "github.com/scottlaird/vyos-parser/syntax"
        "github.com/scottlaird/vyos-template/config"
        "honnef.co/go/js/dom/v2"
)

// Embed template files directly into the generated .wasm file.  These
// are accessed via the `f` global var, below.
//
// //go:embed templates/addon/*.yml
// //go:embed templates/addon/*.show
//
//go:embed templates/base/*.yml
//go:embed templates/base/*.show
var f embed.FS

var (
        configModel *configmodel.VyOSConfigNode
        templates   *config.Templates
        vars config.VariableDefinitionMap
)

func main() {
        // Avoid exiting so the Go runtime is still available from JS.
        done := make(chan struct{}, 0)
        var err error
        configModel, err = syntax.GetDefaultConfigModel()
        if err != nil {
                fmt.Printf("Error!: %v\n", err)
        }

        // Load templates
        t, err := config.LoadTemplates(f)
        if err != nil {
                fmt.Printf("Error!: %v\n", err)
        }
        templates = t
        vars = templates.Variables()
        CreateForm("configForm", vars)

        // Register Go functions with Javascript
        global := js.Global()
        global.Set("updatetarget", js.FuncOf(updatetarget))

        <-done
}

func UpdateNotice(key string, notice string) {
        d := dom.GetWindow().Document()
        noticeElement := d.GetElementByID("notice-"+key)
        noticeElement.SetInnerHTML(notice)
}

func UpdateNotices(notices map[string]string) {
        for k,v := range notices {
                UpdateNotice(k,v)
        }
}

func CreateForm(formID string, vars config.VariableDefinitionMap) {
        d := dom.GetWindow().Document()
        formElement := d.GetElementByID(formID)

        keys := vars.KeysInPriorityOrder()
        fmt.Printf("Keys: %#v\n", keys)

        for _, key := range keys {
                div := d.CreateElement("div")
                div.SetID("div-"+key)

                label := d.CreateElement("label").(*dom.HTMLLabelElement)
                //label.SetFor(key)
                label.SetID("label-" + key)
                if vars[key].HelpText != "" {
                        label.SetTitle(vars[key].HelpText)
                }
                labelText := vars[key].Label
                if labelText == "" {
                        labelText = key
                }
                label.SetTextContent(labelText)
                div.AppendChild(label)

                input := d.CreateElement("input").(*dom.HTMLInputElement)
                input.SetID(key)

                if vars[key].Type == "boolean" {
                        input.SetType("checkbox")
                        if vars[key].Default == "true" {
                                input.SetDefaultChecked(true)
                        }
                        div.AppendChild(input)
                } else {
                        input.SetDefaultValue(vars[key].Default)
                        div.AppendChild(input)
                }

                notice := d.CreateElement("span")
                notice.SetID("notice-"+key)
                notice.Class().SetString("notice")
                div.AppendChild(notice)
                
                formElement.AppendChild(div)
        }

        values := ReadForm(formID, vars)
        EnableDisableForm(formID, vars, values)
        UpdateNotice("SerialConsolePort", "This is a test")
}

func ReadForm(formID string, vars config.VariableDefinitionMap) config.Values {
        values := make(config.Values)
        d := dom.GetWindow().Document()

        for k := range vars {
                input := d.GetElementByID(k).(*dom.HTMLInputElement)
                if vars[k].Type != "boolean" {
                        values[k] = input.Value()
                } else {
                        values[k] = input.Checked()
                }
        }

        return values
}

func EnableDisableForm(formID string, vars config.VariableDefinitionMap, values config.Values) {
        d := dom.GetWindow().Document()

        for k := range vars {
                enable := values.FieldIsEnabled(vars, k)
                
                input := d.GetElementByID(k).(*dom.HTMLInputElement)
                input.SetDisabled(!enable) 
                if vars[k].Type != "boolean" {
                        values[k] = input.Value()
                } else {
                        values[k] = input.Checked()
                }
        }
}


// updatetarget writes to InnerHTML on the .target element.
func updatetarget(this js.Value, args []js.Value) any {
        t := time.Now()

        d := dom.GetWindow().Document()

        values := ReadForm("configForm", vars)
        EnableDisableForm("configForm", vars, values)
        fieldErrors := values.Validate(vars)
        UpdateNotices(fieldErrors)

        template := templates.BaseTemplates[0].Body
        configText := bytes.Buffer{}
        err := template.Execute(&configText, values)
        if err != nil {
                fmt.Printf("Error!: %v", err)
        }
        
        ast, err := parser.ParseShowFormat(configText.String(), configModel)
        if err != nil {
                fmt.Printf("Error!: %v", err)
        }
        ast.Sort()

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
