package sql

import (
	"sort"
	"strings"
)

type Builder struct {
	Selectors map[string]Selector
}

func (builder *Builder) AddSelector(column string, line string) {
	selector := NewSelector(line)
	if selector == nil {
		return
	}
	builder.Selectors[column] = selector
}

func (builder *Builder) SetDefault(name string) {
	if builder.Selectors[name] != nil {
		return
	}
	builder.Selectors[name] = Column(Escape(name))
}

func (builder *Builder) Build(table string) (selectclause string, appendclause string) {
	table = Escape(table)

	selects := make([]string, 0, len(builder.Selectors))
	for _, column := range builder.SelectorKeys() {
		selector := builder.Selectors[column]
		selects = append(selects, selector.selectClause(table, Escape(column)))
	}

	appends := make([]string, 0, len(builder.Selectors))
	for _, column := range builder.SelectorKeys() {
		selector := builder.Selectors[column]
		aClause := selector.appendClause(table, Escape(column))
		if aClause == "" {
			continue
		}
		appends = append(appends, aClause)
	}
	return strings.Join(selects, ",\n"), strings.Join(appends, "\n")

}

func (builder *Builder) SelectorKeys() []string {
	names := make([]string, 0, len(builder.Selectors))
	for k := range builder.Selectors {
		names = append(names, k)
	}
	sort.Strings(names)
	return names
}
