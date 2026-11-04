package fxmddecimal

import (
	"bytes"
	"encoding/json"
	"fmt"
	sdk "github.com/fxmacrodata/fxmacrodata-go"
	"github.com/shopspring/decimal"
	"reflect"
	"strings"
	"sync"
	"testing"
)

func TestNativeExactDecimalWithoutFloatIntermediate(t *testing.T) {
	raw := []byte(` {"currency":"USD","data":[{"value":9007199254740993.1200,"revision":0},{"value":9007199254740993.1300,"revision":1},{"value":null,"revision":2}],"provenance":{"source":"fictional offline fixture"},"future":true} `)
	doc, err := ParseJSON(raw)
	if err != nil {
		t.Fatal(err)
	}
	value, err := doc.NumberAt("/data/0/value")
	if err != nil {
		t.Fatal(err)
	}
	if value.StringFixed(4) != "9007199254740993.1200" || value.Exponent() != -4 || value.Coefficient().String() != "90071992547409931200" {
		t.Fatal("native precision lost", value, value.Exponent())
	}
	extra, _ := decimal.NewFromString("0.88")
	if got := value.Add(extra).StringFixed(4); got != "9007199254740994.0000" {
		t.Fatal("native exact addition lost", got)
	}
	text, err := doc.NumberTextAt("/data/0/value")
	if err != nil || text != "9007199254740993.1200" {
		t.Fatal(text, err)
	}
	if !bytes.Equal(doc.Raw(), raw) {
		t.Fatal("response changed")
	}
	raw[0] = 'X'
	if doc.Raw()[0] == 'X' {
		t.Fatal("document aliases input")
	}
	copy := doc.Raw()
	copy[0] = 'X'
	if doc.Raw()[0] == 'X' {
		t.Fatal("Raw exposed internal bytes")
	}
	view, err := doc.JSONAt("/data")
	if err != nil || !strings.Contains(string(view), "9007199254740993.1200") || !strings.Contains(string(view), `"revision":2`) {
		t.Fatal(string(view), err)
	}
}
func TestNullMissingStringAndMetadataSemantics(t *testing.T) {
	raw := []byte(`{"nullable":null,"zero":0,"flag":false,"text":"001.2300","date":"2026-11-01","clock":"unknown","empty":[]}`)
	doc, err := ParseJSON(raw)
	if err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct {
		path  string
		found bool
	}{{"/nullable", true}, {"/missing", false}, {"/empty/0", false}, {"/zero", true}} {
		found, err := doc.Exists(tc.path)
		if err != nil || found != tc.found {
			t.Fatal(tc, found, err)
		}
	}
	for _, path := range []string{"/nullable", "/missing", "/flag", "/text", "/date", "/empty"} {
		if _, err := doc.NumberAt(path); err == nil {
			t.Fatal("nonnumeric value guessed", path)
		}
	}
	zero, err := doc.NumberAt("/zero")
	if err != nil || !zero.Equal(decimal.Zero) {
		t.Fatal(zero, err)
	}
	value, err := doc.NumberStringAt("/text")
	if err != nil || value.StringFixed(4) != "1.2300" {
		t.Fatal(value, err)
	}
	if _, err := doc.NumberStringAt("/date"); err == nil {
		t.Fatal("date interpreted as number")
	}
	if _, err := doc.NumberStringAt("/zero"); err == nil {
		t.Fatal("number implicitly treated as string")
	}
	null, err := doc.JSONAt("/nullable")
	if err != nil || string(null) != "null" {
		t.Fatal(string(null), err)
	}
	if !bytes.Equal(doc.Raw(), raw) {
		t.Fatal("string/null metadata changed")
	}
}
func TestPointersAndNativeScale(t *testing.T) {
	doc, err := ParseJSON([]byte(`{"a/b":{"~key":1.2300},"":-2.5e-4,"unicode":"\u00c5\ud83d\ude42"}`))
	if err != nil {
		t.Fatal(err)
	}
	n, err := doc.NumberAt("/a~1b/~0key")
	if err != nil || n.Exponent() != -4 {
		t.Fatal(n, err)
	}
	n, err = doc.NumberAt("/")
	if err != nil || n.String() != "-0.00025" {
		t.Fatal(n, err)
	}
	view, err := doc.JSONAt("/unicode")
	if err != nil {
		t.Fatal(err)
	}
	var text string
	if err := json.Unmarshal(view, &text); err != nil || len([]rune(text)) != 2 {
		t.Fatal("Unicode changed", err)
	}
	for _, pointer := range []string{"not-a-pointer", "/a~2b"} {
		if _, err := doc.Exists(pointer); err == nil {
			t.Fatal("bad pointer accepted")
		}
	}
	array, _ := ParseJSON([]byte(`[1]`))
	for _, pointer := range []string{"/01", "/-", "/+1"} {
		if _, err := array.Exists(pointer); err == nil {
			t.Fatal("noncanonical array index accepted")
		}
	}
}
func TestUnsupportedNumbersRemainAvailableAsOriginalText(t *testing.T) {
	raw := []byte(`{"huge":1e10001,"large":` + strings.Repeat("1", MaxCoefficientDigits+1) + `,"deep_scale":1.1e-10000}`)
	doc, err := ParseJSON(raw)
	if err != nil {
		t.Fatal(err)
	}
	for _, pointer := range []string{"/huge", "/large", "/deep_scale"} {
		if _, err := doc.NumberAt(pointer); err == nil {
			t.Fatal("unbounded conversion accepted")
		}
		if text, err := doc.NumberTextAt(pointer); err != nil || text == "" {
			t.Fatal("original numeric text lost", err)
		}
	}
	if !bytes.Equal(doc.Raw(), raw) {
		t.Fatal("unsupported numeric removed or rounded")
	}
	for _, text := range []string{"", "NaN", "Infinity", "1e999999999999999999999999", "1.2.3", " 1", "1%", "1_000"} {
		raw, _ := json.Marshal(map[string]string{"value": text})
		doc, _ := ParseJSON(raw)
		if _, err := doc.NumberStringAt("/value"); err == nil {
			t.Fatal("invalid native numeric accepted")
		}
	}
}
func TestMalformedBoundsAndNilReceiver(t *testing.T) {
	for _, raw := range [][]byte{nil, []byte(`{"x":1,"x":2}`), []byte(`{} {}`), []byte(`{"x":`), {0xff}, []byte(strings.Repeat("[", MaxJSONDepth+2) + "0" + strings.Repeat("]", MaxJSONDepth+2)), []byte(`{"x":"` + strings.Repeat("x", MaxResponseBytes) + `"}`), []byte("[" + strings.Repeat("0,", MaxJSONValues) + "0]")} {
		if _, err := ParseJSON(raw); err == nil {
			t.Fatal("invalid/bounded input accepted")
		}
	}
	var doc *Document
	if doc.Raw() != nil {
		t.Fatal("nil Raw not empty")
	}
	if _, err := doc.Exists(""); err == nil {
		t.Fatal("nil receiver accepted")
	}
	if _, err := doc.NumberAt(""); err == nil {
		t.Fatal("nil numeric access accepted")
	}
	if _, err := Parse(make(chan int)); err == nil {
		t.Fatal("non-JSON model accepted")
	}
}
func TestGlobalNativeRoundingSettingsAreUnchanged(t *testing.T) {
	precision := decimal.DivisionPrecision
	quoted := decimal.MarshalJSONWithoutQuotes
	doc, err := ParseJSON([]byte(`{"n":1.234500}`))
	if err != nil {
		t.Fatal(err)
	}
	if _, err := doc.NumberAt("/n"); err != nil {
		t.Fatal(err)
	}
	if decimal.DivisionPrecision != precision || decimal.MarshalJSONWithoutQuotes != quoted {
		t.Fatal("adapter mutated native global settings")
	}
}
func TestEveryResponseFamilyKeepsFullJSON(t *testing.T) {
	fixtures := map[string]string{
		"catalogue":                              `{"currency":"USD","data":[],"coverage":{"future":null}}`,
		"announcements_revisions":                `{"data":[{"revision":0,"value":1.2300},{"revision":1,"value":null}],"publication_time_status":"unknown"}`,
		"calendar":                               `{"data":[{"release_time_status":"not_announced","date":"2026-11-01"}]}`,
		"predictions_coverage":                   `{"data":[],"coverage":{"available":false}}`,
		"central_bank":                           `{"data":[{"headline":"fictional fixture","provenance":{"official":false}}]}`,
		"forex":                                  `{"base":"USD","quote":"EUR","data":[{"rate":1.234500}]}`,
		"sources":                                `{"data":[],"provenance":{"source":"fictional offline fixture"}}`,
		"cot_commodities":                        `{"data":[{"net":-1.500}],"future_quality":{"kept":true}}`,
		"sessions_prices":                        `{"data":[],"sessions":[],"prices":[]}`,
		"curves_factors_differentials_sentiment": `{"data":[],"curve":[],"factors":[],"sentiment":null}`,
		"research":                               `{"panels":[{"data":[],"model":"fictional"}],"request":{"as_of":"2026-11-01"}}`,
		"seasonality_news_future":                `{"future":{"unknown":[null,true,9007199254740993]},"data":[]}`,
		"array":                                  `[null,{"unchanged":true}]`, "scalar": `"unchanged"`, "boolean": `true`, "null": `null`,
	}
	for name, raw := range fixtures {
		t.Run(name, func(t *testing.T) {
			doc, err := ParseJSON([]byte(raw))
			if err != nil {
				t.Fatal(err)
			}
			if found, err := doc.Exists(""); err != nil || !found {
				t.Fatal(found, err)
			}
			if !bytes.Equal(doc.Raw(), []byte(raw)) {
				t.Fatal("response narrowed")
			}
		})
	}
}
func TestPublishedSDKAvailabilityDoesNotIntroduceAnOperationList(t *testing.T) {
	names := []string{"GetAnnouncementChanges", "GetAnnouncements", "GetCalendar", "GetCommodities", "GetCot", "GetCurves", "GetDataCatalogue", "GetFactor", "GetFinancialPrices", "GetForex", "GetFxIntradayReferenceRates", "GetFxReferenceRates", "GetFxSessionRates", "GetFxSourceUniverse", "GetFxSources", "GetHealth", "GetLatestAnnouncements", "GetLatestCommodities", "GetMarketSessions", "GetPredictionCoverage", "GetPredictionCoverageForCurrency", "GetPredictions", "GetPressReleases", "GetRateDifferentials", "GetRiskSentiment", "Ping", "ResearchPanel"}
	api := reflect.TypeOf(sdk.NewAPIClient(sdk.NewConfiguration()).FXMacroDataAPI)
	for _, name := range names {
		if _, ok := api.MethodByName(name); !ok {
			t.Fatal("public SDK method unavailable", name)
		}
	}
}
func TestConcurrentDocumentReadsAreIndependent(t *testing.T) {
	raw := []byte(`{"value":9007199254740993.1200}`)
	doc, err := ParseJSON(raw)
	if err != nil {
		t.Fatal(err)
	}
	var wg sync.WaitGroup
	for i := 0; i < 32; i++ {
		wg.Add(1)
		go func(n int) {
			defer wg.Done()
			value, err := doc.NumberAt("/value")
			if err != nil || value.StringFixed(4) != "9007199254740993.1200" {
				t.Errorf("native read %d lost precision: %v", n, err)
			}
			copy := doc.Raw()
			copy[0] = 'X'
			if !bytes.Equal(doc.Raw(), raw) {
				t.Error("Raw alias exposed")
			}
		}(i)
	}
	wg.Wait()
}
func TestTypedJSONRepresentationAndRootNumber(t *testing.T) {
	doc, err := Parse(struct {
		Value  json.Number `json:"value"`
		Hidden string      `json:"-"`
	}{Value: json.Number("1.2300"), Hidden: "fictional"})
	if err != nil {
		t.Fatal(err)
	}
	value, err := doc.NumberAt("/value")
	if err != nil || value.Exponent() != -4 {
		t.Fatal(value, err)
	}
	if exists, _ := doc.Exists("/Hidden"); exists {
		t.Fatal("non-JSON field exposed")
	}
	root, err := ParseJSON([]byte(`9007199254740993`))
	if err != nil {
		t.Fatal(err)
	}
	value, err = root.NumberAt("")
	if err != nil || fmt.Sprint(value) != "9007199254740993" {
		t.Fatal(value, err)
	}
}
