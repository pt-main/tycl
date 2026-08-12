package generation

import (
	"fmt"
	"strings"

	"github.com/pt-main/tycl/shared"
)

func GeneratePlainTextDocs(wd shared.WithDocumentation) string {
	return generatePlainTextDocs("", "main", wd)
}

func generatePlainTextDocs(dtype, name string, wd shared.WithDocumentation) string {
	docs := []string{}
	addHeader := func(htype, name string, tabs int) {
		if htype != "" {
			htype += ": "
		}
		docs = append(docs, fmt.Sprintf(`%v-== %v[%v] ========----`, strings.Repeat("|   ", tabs), htype, name))
	}
	addDocs := func(doc string, tabs int) {
		for _, line := range strings.Split(doc, "\n") {
			docs = append(docs, fmt.Sprintf(strings.Repeat(`|   `, tabs)+`%v`, strings.TrimSpace(line)))
		}
	}
	addHeader(dtype, name, 0)
	comms := wd.GetComments()
	if len(comms) > 0 {
		addDocs(comms[0], 1)
	}
	for name, obj := range wd.GetInnerV() {
		addDocs(generatePlainTextDocs("object", name, obj), 1)
	}
	for name, obj := range wd.GetInnerA() {
		addHeader("array", name, 1)
		for idx, aObj := range obj {
			docs := generatePlainTextDocs(fmt.Sprintf(`%v element %v`, name, idx), name, aObj)
			addDocs(docs, 2)
		}
	}
	if len(comms) > 1 {
		for i := 1; i < len(comms); i++ {
			addDocs(comms[i], 1)
		}
	}
	return strings.Join(docs, "\n")
}
