# Captchure
Flexible client for [anti-captcha](http://anti-captcha.com) written in Golang.

## Features

* It supports ReCaptcha, Usual (image) captcha and FunCaptcha out of the box
* It is not hardcoded for happy path, you can define your own parameters for tasks
* It is friendly to debugging & and provide behavior logs
* It is easy to write your own solution logic around the engine

## Installation
```
$ go get github.com/narrator69/captchure
```

## Documentation
See [overview](https://anti-captcha.com/apidoc/image) and [detailed docs](https://anticaptcha.atlassian.net/wiki/spaces/API/pages/5079073/createTask+captcha+task+creating).

## Usage

### Solving typical captcha

```go
package main

import (
	"github.com/narrator69/captchure"
	"fmt"
)

func main()  {
	c := captchure.Captchure{ClientKey: "e1671797c52e15f763380b45e841ec32"}

	base64image, _ := captchure.LocalFileToBase64("cap.jpeg")

	taskParameters := make(map[string]interface{})

	taskParameters["minLength"] = 5 // optional parameter, see docs

	word, _ := c.SolveImage(base64image, taskParameters)

	fmt.Println(word) // qGphJD
}
```
### Solving typical captcha with additional parameters

```go
package main

import (
	"github.com/narrator69/captchure"
	"fmt"
)

func main()  {
	c := captchure.Captchure{
		ClientKey: "e1671797c52e15f763380b45e841ec32",
		LanguagePool: "rn", // solver should speak russian, see docs
		Interval: 3, // interval between requests in seconds, by default is 1
		Verbose: true, // output behavior
	}

	base64image, _ := captchure.LocalFileToBase64("cap.jpeg")

	taskParameters := make(map[string]interface{})

	taskParameters["minLength"] = 3 // optional parameter, see docs
	taskParameters["case"] = true // optional parameter, see docs

	word, _ := c.SolveImage(base64image, taskParameters)

	fmt.Println(word) // qGphJD
}
```

### Solving ReCaptcha

```go
package main

import (
	"github.com/narrator69/captchure"
	"fmt"
)

func main()  {
	c := captchure.Captchure{ClientKey: "e1671797c52e15f763380b45e841ec32"}

	websiteUrl := "http://samplesite.com" // required parameter, see docs
	websiteKey := "6Lc_aCMTAAAAAFx7u2W0WRXnVbI_v6ZdbM6rYf16" // required parameter, see docs

	word, _ := c.SolveRecaptcha(websiteUrl, websiteKey, make(map[string]interface{}))

	fmt.Println(word) // 03ANcjosrl9NKz99gDz...
}
```
