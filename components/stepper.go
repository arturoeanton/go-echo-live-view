package components

import (
	"github.com/arturoeanton/go-echo-live-view/liveview"
)

type Step struct {
	ID          string
	Title       string
	Description string
	Completed   bool
	Active      bool
	Error       bool
	Icon        string
}

type Stepper struct {
	*liveview.ComponentDriver[*Stepper]
	
	Steps        []Step
	CurrentStep  int
	Orientation  string
	AllowSkip    bool
	OnStepChange func(stepIndex int)
}

func NewStepper(id string, steps []Step) *Stepper {
	if len(steps) > 0 {
		steps[0].Active = true
	}
	
	return &Stepper{
		Steps:       steps,
		CurrentStep: 0,
		Orientation: "horizontal",
		AllowSkip:   false,
	}
}

func (s *Stepper) Start() {
	// Events are registered directly on the ComponentDriver
	if s.ComponentDriver != nil {
		s.ComponentDriver.Events["GoToStep"] = func(c *Stepper, data interface{}) {
			c.GoToStep(data)
		}
		s.ComponentDriver.Events["NextStep"] = func(c *Stepper, data interface{}) {
			c.NextStep(data)
		}
		s.ComponentDriver.Events["PreviousStep"] = func(c *Stepper, data interface{}) {
			c.PreviousStep(data)
		}
	}
}

func (s *Stepper) GetTemplate() string {
	return `
<div class="stepper-container {{.Orientation}}" id="{{.IdComponent}}">
	<style>
		.stepper-container {
			width: 100%;
			padding: 1rem;
		}
		
		.stepper-container.vertical .steps-list {
			flex-direction: column;
		}
		
		.steps-list {
			display: flex;
			justify-content: space-between;
			margin-bottom: 2rem;
			position: relative;
		}
		
		.step-item {
			flex: 1;
			text-align: center;
			position: relative;
			cursor: pointer;
		}
		
		.step-item.disabled {
			cursor: not-allowed;
			opacity: 0.6;
		}
		
		.step-connector {
			position: absolute;
			top: 20px;
			left: 50%;
			width: 100%;
			height: 2px;
			background: #e5e7eb;
			z-index: -1;
		}
		
		.step-connector.completed {
			background: #10b981;
		}
		
		.step-number {
			width: 40px;
			height: 40px;
			border-radius: 50%;
			background: #e5e7eb;
			color: #6b7280;
			display: flex;
			align-items: center;
			justify-content: center;
			margin: 0 auto 0.5rem;
			font-weight: 600;
			transition: all 0.3s;
			position: relative;
			z-index: 1;
		}
		
		.step-item.active .step-number {
			background: #3b82f6;
			color: white;
			box-shadow: 0 0 0 4px rgba(59, 130, 246, 0.1);
		}
		
		.step-item.completed .step-number {
			background: #10b981;
			color: white;
		}
		
		.step-item.error .step-number {
			background: #ef4444;
			color: white;
		}
		
		.step-title {
			font-weight: 600;
			color: #374151;
			margin-bottom: 0.25rem;
		}
		
		.step-description {
			font-size: 0.875rem;
			color: #6b7280;
		}
		
		.step-content {
			background: white;
			padding: 2rem;
			border-radius: 0.5rem;
			box-shadow: 0 1px 3px rgba(0, 0, 0, 0.1);
			margin-bottom: 1.5rem;
		}
		
		.step-actions {
			display: flex;
			justify-content: space-between;
			gap: 1rem;
		}
		
		.step-btn {
			padding: 0.625rem 1.25rem;
			border-radius: 0.375rem;
			font-weight: 500;
			cursor: pointer;
			transition: all 0.2s;
			border: 1px solid transparent;
		}
		
		.step-btn.primary {
			background: #3b82f6;
			color: white;
		}
		
		.step-btn.primary:hover {
			background: #2563eb;
		}
		
		.step-btn.secondary {
			background: white;
			color: #374151;
			border-color: #d1d5db;
		}
		
		.step-btn.secondary:hover {
			background: #f3f4f6;
		}
		
		.step-btn:disabled {
			opacity: 0.5;
			cursor: not-allowed;
		}
		
		.checkmark {
			display: inline-block;
			width: 20px;
			height: 20px;
		}
	</style>
	
	<div class="steps-list">
		{{range $index, $step := .Steps}}
			{{if gt $index 0}}
				<div class="step-connector {{if le $index $.CurrentStep}}completed{{end}}"></div>
			{{end}}
			<div class="step-item {{if $step.Active}}active{{end}} {{if $step.Completed}}completed{{end}} {{if $step.Error}}error{{end}} {{if and (not $.AllowSkip) (gt $index $.CurrentStep)}}disabled{{end}}"
				 onclick="{{if or $.AllowSkip (le $index $.CurrentStep)}}send_event('{{$.IdComponent}}', 'GoToStep', {{$index}}){{end}}">
				<div class="step-number">
					{{if $step.Completed}}
						<span class="checkmark">✓</span>
					{{else if $step.Error}}
						<span>!</span>
					{{else}}
						{{add $index 1}}
					{{end}}
				</div>
				<div class="step-title">{{$step.Title}}</div>
				{{if $step.Description}}
					<div class="step-description">{{$step.Description}}</div>
				{{end}}
			</div>
		{{end}}
	</div>
	
	<div class="step-content">
		{{if lt .CurrentStep (len .Steps)}}
			<h3>{{(index .Steps .CurrentStep).Title}}</h3>
			<p>{{(index .Steps .CurrentStep).Description}}</p>
			
			<div style="padding: 2rem 0;">
				Content for step {{add .CurrentStep 1}} goes here
			</div>
		{{end}}
	</div>
	
	<div class="step-actions">
		<button class="step-btn secondary" 
				onclick="send_event('{{.IdComponent}}', 'PreviousStep', null)"
				{{if eq .CurrentStep 0}}disabled{{end}}>
			Previous
		</button>
		
		<button class="step-btn primary" 
				onclick="send_event('{{.IdComponent}}', 'NextStep', null)"
				{{if ge .CurrentStep (sub (len .Steps) 1)}}disabled{{end}}>
			{{if eq .CurrentStep (sub (len .Steps) 1)}}
				Finish
			{{else}}
				Next
			{{end}}
		</button>
	</div>
</div>

<script>
function add(a, b) { return a + b; }
function sub(a, b) { return a - b; }
</script>
`
}

func (s *Stepper) GetDriver() liveview.LiveDriver {
	return s
}

func (s *Stepper) GoToStep(data interface{}) {
	stepIndex := int(data.(float64))
	
	if stepIndex < 0 || stepIndex >= len(s.Steps) {
		return
	}
	
	if !s.AllowSkip && stepIndex > s.CurrentStep {
		return
	}
	
	s.Steps[s.CurrentStep].Active = false
	s.CurrentStep = stepIndex
	s.Steps[s.CurrentStep].Active = true
	
	if s.OnStepChange != nil {
		s.OnStepChange(stepIndex)
	}
	
	s.Commit()
}

func (s *Stepper) NextStep(data interface{}) {
	if s.CurrentStep < len(s.Steps)-1 {
		s.Steps[s.CurrentStep].Completed = true
		s.Steps[s.CurrentStep].Active = false
		s.CurrentStep++
		s.Steps[s.CurrentStep].Active = true
		
		if s.OnStepChange != nil {
			s.OnStepChange(s.CurrentStep)
		}
		
		s.Commit()
	}
}

func (s *Stepper) PreviousStep(data interface{}) {
	if s.CurrentStep > 0 {
		s.Steps[s.CurrentStep].Active = false
		s.CurrentStep--
		s.Steps[s.CurrentStep].Active = true
		s.Steps[s.CurrentStep].Completed = false
		
		if s.OnStepChange != nil {
			s.OnStepChange(s.CurrentStep)
		}
		
		s.Commit()
	}
}

func (s *Stepper) SetStepError(stepIndex int, hasError bool) {
	if stepIndex >= 0 && stepIndex < len(s.Steps) {
		s.Steps[stepIndex].Error = hasError
		s.Commit()
	}
}

func (s *Stepper) CompleteCurrentStep() {
	if s.CurrentStep < len(s.Steps) {
		s.Steps[s.CurrentStep].Completed = true
		s.Commit()
	}
}

func (s *Stepper) Reset() {
	for i := range s.Steps {
		s.Steps[i].Active = false
		s.Steps[i].Completed = false
		s.Steps[i].Error = false
	}
	
	s.CurrentStep = 0
	if len(s.Steps) > 0 {
		s.Steps[0].Active = true
	}
	
	s.Commit()
}