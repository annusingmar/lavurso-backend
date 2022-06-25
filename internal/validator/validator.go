package validator

type Validator struct {
	Errors map[string][]string
}

func NewValidator() *Validator {
	return &Validator{
		Errors: make(map[string][]string),
	}
}

func (v *Validator) Add(key, msg string) {
	v.Errors[key] = append(v.Errors[key], msg)
}

func (v *Validator) Check(result bool, key, msg string) {
	if !result {
		v.Add(key, msg)
	}
}

func (v *Validator) Valid() bool {
	return len(v.Errors) == 0
}
