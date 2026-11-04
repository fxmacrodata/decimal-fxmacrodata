# FXMacroData response capabilities

This is an offline native decimal/data reader. Customers provide responses from their own entitled SDK/client. Every original endpoint and request parameter stays with that client. No transport, key, endpoint selector, automatic pagination or replacement FXMacroData schema is introduced.

| FXMacroData capability | Target platform surface | Auth model | Implemented files | Tests | Status |
| --- | --- | --- | --- | --- | --- |
| Catalogue/discovery | Full JSON and native decimal access where numeric | Customer client; no adapter credentials | decimal.go | catalogue fixture | native response reader |
| Macro history/releases/revisions | Exact number text/native Decimal plus unchanged raw history | Customer client/API entitlement | decimal.go | revisions, nulls, original bytes | native response reader |
| Calendar | Complete JSON with dates/clocks kept intact | Customer client/API entitlement | decimal.go | calendar/date preservation | native response reader |
| Predictions/coverage | Complete JSON and optional numeric access | Customer client/API entitlement | decimal.go | predictions/coverage fixture | native response reader |
| Central-bank press releases | Full metadata/text JSON | Customer client/API entitlement | decimal.go | central-bank fixture | native response reader |
| FX/source/reference metadata | Exact numeric parsing and all original fields | Customer client/API entitlement | decimal.go | FX precision and sources fixtures | native response reader |
| COT/commodities | Exact signed decimals plus full JSON | Customer client/API entitlement | decimal.go | COT/commodity fixture | native response reader |
| Sessions/financial prices | Native decimal prices and complete nonnumeric data | Customer client/API entitlement | decimal.go | sessions/prices fixture | native response reader |
| Curves/factors/differentials/sentiment | Full nested JSON and numeric selectors | Customer client/API entitlement | decimal.go | structured-response fixture | native response reader |
| Research panels | Full nested data without flattening | Customer client/API entitlement | decimal.go | research fixture | native response reader |
| Other public v1 JSON, including seasonality/news where entitled | Generic JSON input without an endpoint allowlist | Customer client/API entitlement | decimal.go | future fields and root-shape fixtures | native response reader |

Numeric access is bounded as documented in README. Values outside conversion limits are retained in full raw/JSON/text form, not rounded or dropped. Native Decimal arithmetic/display behavior is separate from the original response. No clocks, missing values, forecasts or future events are inferred.

Fictional offline native fixtures verify JSON mechanisms and supported response shapes, not every live endpoint. The public SDK's 27 method names are checked separately for compatibility; these checks do not exercise all request parameters or make any network call. Typed SDK serialization retains the model's representation; raw input is required for original HTTP bytes and unknown fields. Customer access, rights and explicit pagination remain the client's responsibility.
