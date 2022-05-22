# structs2map

Convert struct into a `map[string]interface{}`. 

## Getting started 

Example Function

```go
func struct2map(app *cli.App) map[string]interface{} {
	s := structs.New(app)
	m := s.Map()
	return m
}
```