package components

import (
	"fmt"
	"github.com/arturoeanton/go-echo-live-view/liveview"
)

type AnimationType string

const (
	AnimationFadeIn    AnimationType = "fadeIn"
	AnimationFadeOut   AnimationType = "fadeOut"
	AnimationSlideIn   AnimationType = "slideIn"
	AnimationSlideOut  AnimationType = "slideOut"
	AnimationBounce    AnimationType = "bounce"
	AnimationRotate    AnimationType = "rotate"
	AnimationPulse     AnimationType = "pulse"
	AnimationShake     AnimationType = "shake"
)

type Animation struct {
	*liveview.ComponentDriver[*Animation]
	Content      string
	Type         AnimationType
	Duration     string
	Delay        string
	IterationCount string
	IsPlaying    bool
}

func (a *Animation) Start() {
	if a.Duration == "" {
		a.Duration = "1s"
	}
	if a.Delay == "" {
		a.Delay = "0s"
	}
	if a.IterationCount == "" {
		a.IterationCount = "1"
	}
	a.Commit()
}

func (a *Animation) GetTemplate() string {
	return `
	<div id="{{.IdComponent}}" class="animation-container">
		<style>
			.animation-container { padding: 2rem; }
			.animated-element {
				display: inline-block;
				animation-duration: {{.Duration}};
				animation-delay: {{.Delay}};
				animation-iteration-count: {{.IterationCount}};
				animation-fill-mode: both;
			}
			{{if .IsPlaying}}
			.animated-element { animation-name: {{.Type}}; }
			{{end}}
			
			@keyframes fadeIn {
				from { opacity: 0; }
				to { opacity: 1; }
			}
			@keyframes fadeOut {
				from { opacity: 1; }
				to { opacity: 0; }
			}
			@keyframes slideIn {
				from { transform: translateX(-100%); }
				to { transform: translateX(0); }
			}
			@keyframes slideOut {
				from { transform: translateX(0); }
				to { transform: translateX(100%); }
			}
			@keyframes bounce {
				0%, 20%, 50%, 80%, 100% { transform: translateY(0); }
				40% { transform: translateY(-30px); }
				60% { transform: translateY(-15px); }
			}
			@keyframes rotate {
				from { transform: rotate(0deg); }
				to { transform: rotate(360deg); }
			}
			@keyframes pulse {
				0% { transform: scale(1); }
				50% { transform: scale(1.1); }
				100% { transform: scale(1); }
			}
			@keyframes shake {
				0%, 100% { transform: translateX(0); }
				10%, 30%, 50%, 70%, 90% { transform: translateX(-10px); }
				20%, 40%, 60%, 80% { transform: translateX(10px); }
			}
			
			.animation-controls { margin-top: 2rem; display: flex; gap: 1rem; flex-wrap: wrap; }
			.animation-btn { 
				padding: 0.5rem 1rem; 
				background: #4CAF50; 
				color: white; 
				border: none; 
				border-radius: 4px; 
				cursor: pointer; 
			}
			.animation-btn:hover { background: #45a049; }
			.animation-select {
				padding: 0.5rem;
				border: 1px solid #ddd;
				border-radius: 4px;
			}
		</style>
		
		<div class="animated-element">
			{{.Content}}
		</div>
		
		<div class="animation-controls">
			<button class="animation-btn" onclick="send_event('{{.IdComponent}}', 'Play', '')">
				Play Animation
			</button>
			<button class="animation-btn" onclick="send_event('{{.IdComponent}}', 'Stop', '')">
				Stop Animation
			</button>
			<select class="animation-select" onchange="send_event('{{.IdComponent}}', 'ChangeType', this.value)">
				<option value="fadeIn" {{if eq .Type "fadeIn"}}selected{{end}}>Fade In</option>
				<option value="fadeOut" {{if eq .Type "fadeOut"}}selected{{end}}>Fade Out</option>
				<option value="slideIn" {{if eq .Type "slideIn"}}selected{{end}}>Slide In</option>
				<option value="slideOut" {{if eq .Type "slideOut"}}selected{{end}}>Slide Out</option>
				<option value="bounce" {{if eq .Type "bounce"}}selected{{end}}>Bounce</option>
				<option value="rotate" {{if eq .Type "rotate"}}selected{{end}}>Rotate</option>
				<option value="pulse" {{if eq .Type "pulse"}}selected{{end}}>Pulse</option>
				<option value="shake" {{if eq .Type "shake"}}selected{{end}}>Shake</option>
			</select>
			<select class="animation-select" onchange="send_event('{{.IdComponent}}', 'ChangeDuration', this.value)">
				<option value="0.5s" {{if eq .Duration "0.5s"}}selected{{end}}>0.5s</option>
				<option value="1s" {{if eq .Duration "1s"}}selected{{end}}>1s</option>
				<option value="2s" {{if eq .Duration "2s"}}selected{{end}}>2s</option>
				<option value="3s" {{if eq .Duration "3s"}}selected{{end}}>3s</option>
			</select>
			<select class="animation-select" onchange="send_event('{{.IdComponent}}', 'ChangeIterations', this.value)">
				<option value="1" {{if eq .IterationCount "1"}}selected{{end}}>Once</option>
				<option value="2" {{if eq .IterationCount "2"}}selected{{end}}>Twice</option>
				<option value="3" {{if eq .IterationCount "3"}}selected{{end}}>3 times</option>
				<option value="infinite" {{if eq .IterationCount "infinite"}}selected{{end}}>Infinite</option>
			</select>
		</div>
	</div>
	`
}

func (a *Animation) GetDriver() liveview.LiveDriver {
	return a
}

func (a *Animation) Play(data interface{}) {
	a.IsPlaying = false
	a.Commit()
	a.EvalScript(fmt.Sprintf("setTimeout(() => send_event('%s', 'StartAnimation', ''), 10)", a.IdComponent))
}

func (a *Animation) StartAnimation(data interface{}) {
	a.IsPlaying = true
	a.Commit()
}

func (a *Animation) Stop(data interface{}) {
	a.IsPlaying = false
	a.Commit()
}

func (a *Animation) ChangeType(data interface{}) {
	a.Type = AnimationType(data.(string))
	a.IsPlaying = false
	a.Commit()
}

func (a *Animation) ChangeDuration(data interface{}) {
	a.Duration = data.(string)
	a.IsPlaying = false
	a.Commit()
}

func (a *Animation) ChangeIterations(data interface{}) {
	a.IterationCount = data.(string)
	a.IsPlaying = false
	a.Commit()
}

func (a *Animation) SetContent(content string) {
	a.Content = content
	a.Commit()
}

func (a *Animation) Animate(animType AnimationType, duration string) {
	a.Type = animType
	a.Duration = duration
	a.IsPlaying = true
	a.Commit()
}