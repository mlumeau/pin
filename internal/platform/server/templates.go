package server

import "html/template"

func applyTemplateFuncs(base template.FuncMap, extras ...template.FuncMap) {
	if base == nil {
		return
	}
	for _, extra := range extras {
		if extra == nil {
			continue
		}
		for name, fn := range extra {
			if _, exists := base[name]; exists {
				continue
			}
			base[name] = fn
		}
	}
}
