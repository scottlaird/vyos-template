package config

import (
        "cmp"
        "slices"
        "text/template"
        "embed"
        "regexp"
        "fmt"

        "gopkg.in/yaml.v3"
)
//var (
//        configModel *configmodel.VyOSConfigNode
//        templates   *Templates
//        vars VariableDefinitionMap
//)

var (
        // These are default regular expressions for validating
        // certain types of parameters.  They're not perfect; they'll
        // accept 999.999.999.999 as an IP address right now, but
        // they're good enough for now.
        TypeRegex = map[string]string{
                "boolean": "",
                "hostname": "^[-a-z0-9]+$",
                "integer": "^[0-9]+$",
                "ipaddress": "^([0-9]{1,3}[.]){3}([0-9]{1,3})$",
                "ipv4address": "^([0-9]{1,3}[.]){3}([0-9]{1,3})$",
                "ipv4prefix": "^([0-9]{1,3}[.]){3}([0-9]{1,3})/([1-3]?[0-9])$",
        }
)

type Templates struct {
        BaseTemplates  []*Template
        AddonTemplates []*Template
}

func (t *Templates) Variables() VariableDefinitionMap {
        vars := make(VariableDefinitionMap)

        for _, template := range t.BaseTemplates {
                for k, v := range template.Config.Variables {
                        vars[k] = v
                }
        }
        for _, template := range t.AddonTemplates {
                for k, v := range template.Config.Variables {
                        vars[k] = v
                }
        }
        return vars
}

type Template struct {
        Config *TemplateConfig
        Body   *template.Template
}

type TemplateConfig struct {
        Description string                `yaml:"Description"`
        VyOSVersion string                `yaml:"VyOSVersion"`
        Variables   VariableDefinitionMap `yaml:"Variables"`
}

type VariableDefinitionMap map[string]VariableDefinition

func (vdm VariableDefinitionMap) KeysInPriorityOrder() []string {
        keys := []string{}
        for k := range vdm {
                keys = append(keys, k)
        }
        slices.SortFunc(keys, func(a, b string) int {
                return cmp.Compare(vdm[a].Priority, vdm[b].Priority)
        })
        return keys
}

type VariableDefinition struct {
        Priority int    `yaml:"Priority"` // Where the var shows up in the config form
        Type     string `yaml:"Type"`     // Variable type (string, etc).  Eventually used for validation.
        Label    string `yaml:"Label"`    // The text shown (in English) on the web form
        HelpText string `yaml:"HelpText"` // Longer help text for the web form
        Default  string `yaml:"Default"`  // The default value
        Regex    string `yaml:"Regex"`    // A validation regex (not used yet)
        Unless   string `yaml:"Unless"`   // Disable if the var named is true.
        OnlyIf   string `yaml:"OnlyIf"`   // Disable if the var named is false.
}


type Values map[string]interface{}

// Validate verifies that all values are valid, given the constraints
// provided in the provided VariableDefinitionMap.  It returns a map
// of variable name->error message.  Error messages may be blank, in
// which case no error has occured.
func (vals Values) Validate(vars VariableDefinitionMap) map[string]string {
        ret := make(map[string]string)
        
        for k, _ := range vars {
                ret[k]=""
                if vals.FieldIsEnabled(vars, k) {
                        regex := vars[k].Regex
                        if regex == "" {
                                regex = TypeRegex[vars[k].Type]
                        }
                        if regex != "" {
                                matched, err :=  regexp.MatchString(regex, vals[k].(string))
                                if err != nil {
                                        ret[k]=fmt.Sprintf("Definition error: regex fault: %v", err)
                                }

                                if matched == false {
                                        // TODO: this needs a real message
                                        ret[k]="Invalid!"
                                }
                        }
                }
        }

        return ret
}

func (vals Values) FieldIsEnabled(vars VariableDefinitionMap, field string) bool {
        enable := true
        if vars[field].Unless != "" {
                enable = false
                val := vals[vars[field].Unless]
                if val == false  {
                        enable = true
                }
        }
        if vars[field].OnlyIf != "" {
                val := vals[vars[field].OnlyIf]
                enable = false
                if val == true {
                        enable = true
                }
        }

        return enable
}


func LoadTemplates(f embed.FS) (*Templates, error) {
        t := &Templates{}

        d, err := LoadTemplate(f, "templates/base/default.yml", "templates/base/default.show")
        if err != nil {
                return t, err
        }
        t.BaseTemplates = append(t.BaseTemplates, d)

        return t, nil
}

func LoadTemplate(f embed.FS, ymlFile, showFile string) (*Template, error) {
        t := &Template{}

        ymlBytes, err := f.ReadFile(ymlFile)
        if err != nil {
                return nil, err
        }
        showBytes, err := f.ReadFile(showFile)
        if err != nil {
                return nil, err
        }

        t.Config = &TemplateConfig{}
        err = yaml.Unmarshal(ymlBytes, t.Config)
        if err != nil {
                return nil, err
        }

        tmp := template.New(showFile)
        t.Body, err = tmp.Parse(string(showBytes))
        if err != nil {
                return nil, err
        }
        return t, nil
}
