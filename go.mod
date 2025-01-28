module github.com/scottlaird/vyos-template

go 1.22.2

replace github.com/scottlaird/vyos-parser => ../vyos-parser

require (
	github.com/scottlaird/vyos-parser v0.0.0-00010101000000-000000000000
	gopkg.in/yaml.v3 v3.0.1
	honnef.co/go/js/dom/v2 v2.0.0-20241221162326-00dae5193c3f
)

require github.com/kballard/go-shellquote v0.0.0-20180428030007-95032a82bc51 // indirect
