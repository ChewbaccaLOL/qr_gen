package config

import "sort"

func EnabledVariantNames(variants map[string]Variant) []string {
	names := make([]string, 0, len(variants))
	for name, variant := range variants {
		if variant.Disabled {
			continue
		}
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

func EnabledVariants(variants map[string]Variant) []Variant {
	names := EnabledVariantNames(variants)
	out := make([]Variant, 0, len(names))
	for _, name := range names {
		out = append(out, variants[name])
	}
	return out
}
