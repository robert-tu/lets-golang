package forms

// define new errors type - used for validation
type errors map[string][]string

// implement Add() function
func (e errors) Add(field, message string) {
	e[field] = append(e[field], message)
}

// implement Get() function to retrieve error
func (e errors) Get(field string) string {
	es := e[field]
	if len(es) == 0 {
		return ""
	}
	return es[0]
}
