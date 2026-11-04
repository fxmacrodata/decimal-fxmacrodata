// This example uses fictional data entirely offline.
package main

import (
	"fmt"
	fxmddecimal "github.com/fxmacrodata/decimal-fxmacrodata"
	"github.com/shopspring/decimal"
	"log"
)

func main() {
	raw := []byte(`{"currency":"USD","data":[{"value":9007199254740993.1200}],"fixture":true}`)
	doc, err := fxmddecimal.ParseJSON(raw)
	if err != nil {
		log.Fatal(err)
	}
	value, err := doc.NumberAt("/data/0/value")
	if err != nil {
		log.Fatal(err)
	}
	extra, err := decimal.NewFromString("0.88")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Fictional exact sum:", value.Add(extra).StringFixed(4))
}
