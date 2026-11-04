# FXMacroData for shopspring/decimal

Read native exact decimals from complete FXMacroData JSON responses in Go. The library performs no network calls and accepts no credentials. Obtain data using your own FXMacroData SDK/client and entitlement, then pass the complete response JSON to the adapter.

```go
import fxmddecimal "github.com/fxmacrodata/decimal-fxmacrodata"

doc, err := fxmddecimal.ParseJSON(responseJSON)
if err != nil { return err }
value, err := doc.NumberAt("/data/0/value")
if err != nil { return err }
// value is a native decimal.Decimal. doc.Raw() retains the complete response.
```

JSON numbers go directly from their original number text to `decimal.NewFromString`, without a float64 intermediate. Original number spelling and trailing zeros remain available through `NumberTextAt`; the native Decimal retains its coefficient/exponent. Native display methods and arithmetic may choose their own formatting or rounding. Derived calculations stay separate from the original observations, revisions and metadata.

`NumberAt` accepts JSON numbers only. Null, absence, booleans, strings and containers are errors rather than guessed zeroes. `NumberStringAt` explicitly opts into the native library's numeric-string syntax when the requested field really is a numeric string. Dates and units are never stripped or inferred. Check every returned error.

## Complete response access

`Raw()` returns an independent copy of the entire original JSON body, including provenance, quality flags, pagination, nulls, unknown future fields and timestamps. `JSONAt(pointer)` creates a compact subtree view without normalizing number lexemes; the original body remains separate. `Exists(pointer)` distinguishes a missing field from an explicit null. Pointers use RFC 6901 escaping (`~1` for `/`, `~0` for `~`) and canonical nonnegative array indices. The empty pointer selects the root. No rows are flattened, revisions collapsed, fields renamed or values/time inferred.

`Parse` accepts a JSON-serializable SDK model; it preserves that model's JSON representation. Prefer `ParseJSON` for original HTTP bytes, full numeric precision and fields a typed model may not retain. The adapter supports every JSON response family and introduces no endpoint or request-parameter list. All public SDK/client paths, query/body parameters, entitlement checks and explicit pagination remain with the customer-owned client. [CAPABILITIES.md](CAPABILITIES.md) describes the native response surfaces. Streaming and non-JSON transport are outside this response parser.

## Boundaries and examples

Inputs are limited to 4 MiB, JSON nesting to 64 and values to 50,000. Duplicate keys, malformed JSON and invalid UTF-8 are rejected. Compact views are limited to 8 MiB. Native numeric access accepts up to 2,048 characters, 1,024 coefficient digits and an absolute exponent/final scale of 10,000. Unsupported numeric ranges remain intact in the original response and JSON/text views; access returns an explicit error. These limits prevent unbounded conversion, and do not impose a history window or mutate data. The adapter does not alter the decimal library's global precision or rounding settings. Callers choose arithmetic precision, rounding and error handling for their calculations.

`examples/precision` is an explicitly fictional offline calculation. Native tests cover precise arithmetic beyond float64's exact integer range, number spelling, null/absence, strings, metadata/history preservation, pointer escaping, Unicode, limits and concurrent reads. Separate public Go SDK v0.1.0 method-availability checks issue no requests and do not certify live endpoints or customer installation.

This release targets shopspring/decimal v1.6.0 and public Go SDK v0.1.0. It is an independent companion without upstream endorsement. [FXMacroData](https://fxmacrodata.com/?utm_source=github&utm_medium=integration&utm_campaign=open_source&utm_content=decimal_fxm) supplies official-source macroeconomic data. Code is MIT licensed; API data access and redistribution rights remain governed by each customer's plan and agreement.
