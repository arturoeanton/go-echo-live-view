package components

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/arturoeanton/go-echo-live-view/liveview"
)

type ValidationRule struct {
	Type    string
	Value   interface{}
	Message string
}

type FormField struct {
	Name        string
	Label       string
	Type        string
	Value       string
	Placeholder string
	Required    bool
	Pattern     string
	MinLength   int
	MaxLength   int
	Min         string
	Max         string
	Error       string
	Rules       []ValidationRule
}

type Form struct {
	*liveview.ComponentDriver[*Form]
	Fields      []FormField
	SubmitLabel string
	Errors      map[string]string
	IsValid     bool
	OnSubmit    func(data map[string]string) error
}

func (f *Form) Start() {
	if f.SubmitLabel == "" {
		f.SubmitLabel = "Submit"
	}
	if f.Errors == nil {
		f.Errors = make(map[string]string)
	}
	f.Commit()
}

func (f *Form) GetTemplate() string {
	return `
	<form id="{{.IdComponent}}" class="form-component">
		<style>
			.form-component {
				display: flex;
				flex-direction: column;
				gap: 1rem;
				max-width: 500px;
			}
			.form-field {
				display: flex;
				flex-direction: column;
				gap: 0.25rem;
			}
			.form-label {
				font-weight: 600;
				font-size: 0.875rem;
			}
			.form-input {
				padding: 0.5rem;
				border: 1px solid #ccc;
				border-radius: 4px;
				font-size: 1rem;
			}
			.form-input:focus {
				outline: none;
				border-color: #4CAF50;
				box-shadow: 0 0 0 2px rgba(76, 175, 80, 0.2);
			}
			.form-input.error {
				border-color: #f44336;
			}
			.form-error {
				color: #f44336;
				font-size: 0.75rem;
				margin-top: 0.25rem;
			}
			.form-submit {
				padding: 0.75rem 1.5rem;
				background-color: #4CAF50;
				color: white;
				border: none;
				border-radius: 4px;
				font-size: 1rem;
				cursor: pointer;
				transition: background-color 0.3s;
			}
			.form-submit:hover {
				background-color: #45a049;
			}
			.form-submit:disabled {
				background-color: #ccc;
				cursor: not-allowed;
			}
		</style>
		{{range .Fields}}
		<div class="form-field">
			<label class="form-label" for="{{$.IdComponent}}_{{.Name}}">
				{{.Label}}
				{{if .Required}}<span style="color: red;">*</span>{{end}}
			</label>
			<input 
				id="{{$.IdComponent}}_{{.Name}}"
				class="form-input {{if .Error}}error{{end}}"
				type="{{.Type}}"
				name="{{.Name}}"
				value="{{.Value}}"
				placeholder="{{.Placeholder}}"
				{{if .Required}}required{{end}}
				{{if .Pattern}}pattern="{{.Pattern}}"{{end}}
				{{if .MinLength}}minlength="{{.MinLength}}"{{end}}
				{{if .MaxLength}}maxlength="{{.MaxLength}}"{{end}}
				{{if .Min}}min="{{.Min}}"{{end}}
				{{if .Max}}max="{{.Max}}"{{end}}
				onblur="send_event('{{$.IdComponent}}', 'Validate', JSON.stringify({field: '{{.Name}}', value: this.value}))"
				oninput="send_event('{{$.IdComponent}}', 'Input', JSON.stringify({field: '{{.Name}}', value: this.value}))"
			>
			{{if .Error}}
			<span class="form-error">{{.Error}}</span>
			{{end}}
		</div>
		{{end}}
		<button 
			type="button" 
			class="form-submit"
			onclick="send_event('{{.IdComponent}}', 'Submit', JSON.stringify(Object.fromEntries(new FormData(this.closest('form')))))"
		>
			{{.SubmitLabel}}
		</button>
	</form>
	`
}

func (f *Form) GetDriver() liveview.LiveDriver {
	return f
}

func (f *Form) Validate(data interface{}) {
	params := data.(map[string]interface{})
	fieldName := params["field"].(string)
	value := params["value"].(string)
	
	for i, field := range f.Fields {
		if field.Name == fieldName {
			f.Fields[i].Value = value
			f.Fields[i].Error = f.validateField(field, value)
			break
		}
	}
	
	f.checkFormValidity()
	f.Commit()
}

func (f *Form) Input(data interface{}) {
	params := data.(map[string]interface{})
	fieldName := params["field"].(string)
	value := params["value"].(string)
	
	for i, field := range f.Fields {
		if field.Name == fieldName {
			f.Fields[i].Value = value
			break
		}
	}
}

func (f *Form) Submit(data interface{}) {
	params := data.(map[string]interface{})
	formData := make(map[string]string)
	
	for key, val := range params {
		formData[key] = fmt.Sprint(val)
	}
	
	f.validateAll(formData)
	
	if f.IsValid && f.OnSubmit != nil {
		if err := f.OnSubmit(formData); err != nil {
			f.EvalScript(fmt.Sprintf("alert('Error: %s')", err.Error()))
		}
	}
	
	f.Commit()
}

func (f *Form) validateField(field FormField, value string) string {
	if field.Required && value == "" {
		return fmt.Sprintf("%s is required", field.Label)
	}
	
	for _, rule := range field.Rules {
		switch rule.Type {
		case "email":
			emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
			if !emailRegex.MatchString(value) {
				return rule.Message
			}
		case "phone":
			phoneRegex := regexp.MustCompile(`^[+]?[(]?[0-9]{3}[)]?[-\s.]?[0-9]{3}[-\s.]?[0-9]{4,6}$`)
			if !phoneRegex.MatchString(value) {
				return rule.Message
			}
		case "min":
			if len(value) < rule.Value.(int) {
				return rule.Message
			}
		case "max":
			if len(value) > rule.Value.(int) {
				return rule.Message
			}
		case "regex":
			regex := regexp.MustCompile(rule.Value.(string))
			if !regex.MatchString(value) {
				return rule.Message
			}
		}
	}
	
	return ""
}

func (f *Form) validateAll(formData map[string]string) {
	f.IsValid = true
	f.Errors = make(map[string]string)
	
	for i, field := range f.Fields {
		value := formData[field.Name]
		error := f.validateField(field, value)
		f.Fields[i].Error = error
		if error != "" {
			f.Errors[field.Name] = error
			f.IsValid = false
		}
	}
}

func (f *Form) checkFormValidity() {
	f.IsValid = true
	for _, field := range f.Fields {
		if field.Error != "" {
			f.IsValid = false
			break
		}
		if field.Required && field.Value == "" {
			f.IsValid = false
			break
		}
	}
}

func (f *Form) AddField(field FormField) *Form {
	f.Fields = append(f.Fields, field)
	return f
}

func (f *Form) SetField(name string, value string) *Form {
	for i, field := range f.Fields {
		if field.Name == name {
			f.Fields[i].Value = value
			break
		}
	}
	return f
}

func (f *Form) GetField(name string) *FormField {
	for i, field := range f.Fields {
		if field.Name == name {
			return &f.Fields[i]
		}
	}
	return nil
}

func (f *Form) ClearErrors() *Form {
	f.Errors = make(map[string]string)
	for i := range f.Fields {
		f.Fields[i].Error = ""
	}
	f.IsValid = true
	return f
}

func (f *Form) Reset() *Form {
	for i := range f.Fields {
		f.Fields[i].Value = ""
		f.Fields[i].Error = ""
	}
	f.Errors = make(map[string]string)
	f.IsValid = true
	return f
}

func (f *Form) GetValues() map[string]string {
	values := make(map[string]string)
	for _, field := range f.Fields {
		values[field.Name] = field.Value
	}
	return values
}

func (f *Form) SetValues(values map[string]string) *Form {
	for i, field := range f.Fields {
		if value, ok := values[field.Name]; ok {
			f.Fields[i].Value = value
		}
	}
	return f
}

func (f *Form) ValidateEmail(email string) bool {
	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	return emailRegex.MatchString(email)
}

func (f *Form) ValidatePhone(phone string) bool {
	phone = strings.ReplaceAll(phone, " ", "")
	phone = strings.ReplaceAll(phone, "-", "")
	phoneRegex := regexp.MustCompile(`^[+]?[0-9]{10,15}$`)
	return phoneRegex.MatchString(phone)
}