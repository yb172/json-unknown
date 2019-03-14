# JSON Unknown

Imagine we have following Go structs:

```go
type Config struct {
  Name   string  `json:"name,omitempty"`
  Params []Param `json:"params,omitempty"`
}

type Param struct {
  Name  string `json:"name,omitempty"`
  Value string `json:"value,omitempty"`
}
```

and following json:

```json
{
  "name": "parabolic",
  "subdir": "pb",
  "params": [{
    "name": "input",
    "value": "in.csv"
  }, {
    "name": "output",
    "value": "out.csv",
    "tune": "fine"
  }]
}
```

and we do unmarshalling:

```go
cfg := Config{}
if err := json.Unmarshal([]byte(cfgString), &cfg); err != nil {
  log.Fatalf("Error unmarshalling json: %v", err)
}
fmt.Println(cfg)
```

[https://play.golang.org/p/HZgo0jxbQrp](https://play.golang.org/p/HZgo0jxbQrp)

Output would be `{parabolic [{input in.csv} {output out.csv}]}` which makes sense - unknown fields were ignored.

Question: how to find out which fields were ignored?

I.e. `getIgnoredFields(cfg, cfgString)` would return `["subdir", "params[1].tune"]`

(There is a [`DisallowUnknownFields`](https://godoc.org/encoding/json#Decoder.DisallowUnknownFields) option but it's different: this option would result `Unmarshal` in error while question is how to still parse json without errors and find out which fields were ignored)

(Also see on [StackOverflow](https://stackoverflow.com/q/55029335/518469))

This library attempts to provide an answer to this question.

## Usage

Library exports `ValidateUnknownFields` function which returns list of unknown fields and error:

```go
unknowns, err := ValidateUnknownFields(jsonCfgBytes, cfg)
```

JSON is valid and has no unknown fields if both returned values are `nil`

If unknown fields were found full path to the field would be specified, e.g. `triggers[1].repoz`
