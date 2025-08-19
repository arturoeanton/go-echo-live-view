package components

import (
	"github.com/arturoeanton/go-echo-live-view/liveview"
)

type RichEditor struct {
	*liveview.ComponentDriver[*RichEditor]
	Content    string
	Placeholder string
	Height     string
	OnChange   func(content string)
}

func (r *RichEditor) Start() {
	if r.Height == "" {
		r.Height = "300px"
	}
	if r.Placeholder == "" {
		r.Placeholder = "Start typing..."
	}
	r.Commit()
}

func (r *RichEditor) GetTemplate() string {
	return `
	<div id="{{.IdComponent}}" class="rich-editor">
		<style>
			.rich-editor { border: 1px solid #ddd; border-radius: 8px; overflow: hidden; background: white; }
			.editor-toolbar { display: flex; gap: 0.5rem; padding: 0.75rem; background: #f5f5f5; border-bottom: 1px solid #ddd; flex-wrap: wrap; }
			.editor-btn { padding: 0.5rem; background: white; border: 1px solid #ddd; border-radius: 4px; cursor: pointer; font-size: 1rem; min-width: 32px; }
			.editor-btn:hover { background: #e0e0e0; }
			.editor-btn.active { background: #4CAF50; color: white; }
			.editor-separator { width: 1px; background: #ccc; margin: 0 0.5rem; }
			.editor-content { min-height: {{.Height}}; padding: 1rem; outline: none; font-family: -apple-system, sans-serif; line-height: 1.6; }
			.editor-content:focus { background: #fafafa; }
			.editor-content h1 { font-size: 2rem; margin: 1rem 0; }
			.editor-content h2 { font-size: 1.5rem; margin: 0.75rem 0; }
			.editor-content ul, .editor-content ol { margin-left: 1.5rem; }
			.editor-content blockquote { border-left: 4px solid #ddd; margin: 1rem 0; padding-left: 1rem; color: #666; }
		</style>
		
		<div class="editor-toolbar">
			<button class="editor-btn" onclick="send_event('{{.IdComponent}}', 'Format', 'bold')" title="Bold">
				<b>B</b>
			</button>
			<button class="editor-btn" onclick="send_event('{{.IdComponent}}', 'Format', 'italic')" title="Italic">
				<i>I</i>
			</button>
			<button class="editor-btn" onclick="send_event('{{.IdComponent}}', 'Format', 'underline')" title="Underline">
				<u>U</u>
			</button>
			<button class="editor-btn" onclick="send_event('{{.IdComponent}}', 'Format', 'strikethrough')" title="Strikethrough">
				<s>S</s>
			</button>
			<div class="editor-separator"></div>
			<button class="editor-btn" onclick="send_event('{{.IdComponent}}', 'Format', 'h1')" title="Heading 1">
				H1
			</button>
			<button class="editor-btn" onclick="send_event('{{.IdComponent}}', 'Format', 'h2')" title="Heading 2">
				H2
			</button>
			<div class="editor-separator"></div>
			<button class="editor-btn" onclick="send_event('{{.IdComponent}}', 'Format', 'ul')" title="Bullet List">
				•
			</button>
			<button class="editor-btn" onclick="send_event('{{.IdComponent}}', 'Format', 'ol')" title="Numbered List">
				1.
			</button>
			<button class="editor-btn" onclick="send_event('{{.IdComponent}}', 'Format', 'quote')" title="Quote">
				"
			</button>
			<div class="editor-separator"></div>
			<button class="editor-btn" onclick="send_event('{{.IdComponent}}', 'Format', 'link')" title="Insert Link">
				🔗
			</button>
			<button class="editor-btn" onclick="send_event('{{.IdComponent}}', 'Format', 'clear')" title="Clear Formatting">
				✕
			</button>
		</div>
		
		<div class="editor-content" 
			contenteditable="true"
			placeholder="{{.Placeholder}}"
			oninput="send_event('{{.IdComponent}}', 'Change', this.innerHTML)"
			onpaste="setTimeout(() => send_event('{{.IdComponent}}', 'Change', this.innerHTML), 10)"
		>{{.Content}}</div>
	</div>
	`
}

func (r *RichEditor) GetDriver() liveview.LiveDriver {
	return r
}

func (r *RichEditor) Format(data interface{}) {
	format := data.(string)
	switch format {
	case "bold":
		r.EvalScript("document.execCommand('bold', false, null)")
	case "italic":
		r.EvalScript("document.execCommand('italic', false, null)")
	case "underline":
		r.EvalScript("document.execCommand('underline', false, null)")
	case "strikethrough":
		r.EvalScript("document.execCommand('strikeThrough', false, null)")
	case "h1":
		r.EvalScript("document.execCommand('formatBlock', false, '<h1>')")
	case "h2":
		r.EvalScript("document.execCommand('formatBlock', false, '<h2>')")
	case "ul":
		r.EvalScript("document.execCommand('insertUnorderedList', false, null)")
	case "ol":
		r.EvalScript("document.execCommand('insertOrderedList', false, null)")
	case "quote":
		r.EvalScript("document.execCommand('formatBlock', false, '<blockquote>')")
	case "link":
		r.EvalScript("var url = prompt('Enter URL:'); if(url) document.execCommand('createLink', false, url)")
	case "clear":
		r.EvalScript("document.execCommand('removeFormat', false, null)")
	}
}

func (r *RichEditor) Change(data interface{}) {
	r.Content = data.(string)
	if r.OnChange != nil {
		r.OnChange(r.Content)
	}
}

func (r *RichEditor) SetContent(content string) {
	r.Content = content
	r.Commit()
}

func (r *RichEditor) GetContent() string {
	return r.Content
}

func (r *RichEditor) Clear() {
	r.Content = ""
	r.Commit()
}