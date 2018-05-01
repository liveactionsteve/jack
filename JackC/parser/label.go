package parser

import "fmt"

type LabelGenerator map[string]int

func newLabelGenerator() LabelGenerator {
	return make(map[string]int)
}

func (lbl LabelGenerator) generateLabel(prefix string) string {
	result := fmt.Sprintf("%s%d", prefix, lbl[prefix])
	lbl[prefix] += 1
	return result
}
