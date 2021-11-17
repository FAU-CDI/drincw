package drincw

import "sort"

type sortableBundles []*Bundle

func (bundles sortableBundles) Len() int {
	return len(bundles)
}

func (bundles sortableBundles) Swap(i, j int) {
	bundles[i], bundles[j] = bundles[j], bundles[i]
}
func (bundles sortableBundles) Less(i, j int) bool {
	return bundles[i].Group.Weight < bundles[j].Group.Weight
}

func (s sortableBundles) Sort() {
	sort.Sort(s)
}

type sortableFields []Field

func (fields sortableFields) Len() int {
	return len(fields)
}

func (fields sortableFields) Swap(i, j int) {
	fields[i], fields[j] = fields[j], fields[i]
}
func (fields sortableFields) Less(i, j int) bool {
	return fields[i].Weight < fields[j].Weight
}

func (fields sortableFields) Sort() {
	sort.Sort(fields)
}
