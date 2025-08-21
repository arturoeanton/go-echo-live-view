package components

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/arturoeanton/go-echo-live-view/liveview"
)

type FileInfo struct {
	Name     string
	Size     int64
	Type     string
	Data     string
	Preview  string
	Progress int
}

type FileUpload struct {
	*liveview.ComponentDriver[*FileUpload]
	Multiple       bool
	Accept         string
	MaxSize        int64
	MaxFiles       int
	Files          []FileInfo
	DragActive     bool
	UploadProgress int
	Label          string
	OnUpload       func(files []FileInfo) error
}

func (f *FileUpload) Start() {
	if f.Label == "" {
		f.Label = "Choose files or drag and drop"
	}
	if f.MaxSize == 0 {
		f.MaxSize = 10 * 1024 * 1024
	}
	if f.MaxFiles == 0 {
		f.MaxFiles = 10
	}
	f.Commit()
}

func (f *FileUpload) GetTemplate() string {
	return `
	<div id="{{.IdComponent}}" class="file-upload-container">
		<style>
			.file-upload-container {
				width: 100%;
				max-width: 600px;
			}
			.file-drop-zone {
				border: 2px dashed #ccc;
				border-radius: 8px;
				padding: 2rem;
				text-align: center;
				cursor: pointer;
				transition: all 0.3s;
				background-color: #fafafa;
			}
			.file-drop-zone:hover {
				border-color: #4CAF50;
				background-color: #f0f8f0;
			}
			.file-drop-zone.drag-active {
				border-color: #4CAF50;
				background-color: #e8f5e9;
			}
			.file-input-hidden {
				display: none;
			}
			.file-icon {
				font-size: 3rem;
				color: #999;
				margin-bottom: 1rem;
			}
			.file-label {
				font-size: 1.1rem;
				color: #666;
				margin-bottom: 0.5rem;
			}
			.file-hint {
				font-size: 0.875rem;
				color: #999;
			}
			.file-list {
				margin-top: 1.5rem;
			}
			.file-item {
				display: flex;
				align-items: center;
				padding: 0.75rem;
				background: white;
				border: 1px solid #e0e0e0;
				border-radius: 4px;
				margin-bottom: 0.5rem;
			}
			.file-preview {
				width: 40px;
				height: 40px;
				margin-right: 1rem;
				border-radius: 4px;
				object-fit: cover;
			}
			.file-info {
				flex: 1;
			}
			.file-name {
				font-weight: 500;
				color: #333;
			}
			.file-size {
				font-size: 0.875rem;
				color: #666;
			}
			.file-remove {
				padding: 0.25rem 0.5rem;
				background: #f44336;
				color: white;
				border: none;
				border-radius: 4px;
				cursor: pointer;
				font-size: 0.875rem;
			}
			.file-remove:hover {
				background: #d32f2f;
			}
			.file-progress {
				width: 100%;
				height: 4px;
				background: #e0e0e0;
				border-radius: 2px;
				margin-top: 0.5rem;
				overflow: hidden;
			}
			.file-progress-bar {
				height: 100%;
				background: #4CAF50;
				transition: width 0.3s;
			}
			.upload-button {
				margin-top: 1rem;
				padding: 0.75rem 1.5rem;
				background: #4CAF50;
				color: white;
				border: none;
				border-radius: 4px;
				cursor: pointer;
				font-size: 1rem;
				width: 100%;
			}
			.upload-button:hover {
				background: #45a049;
			}
			.upload-button:disabled {
				background: #ccc;
				cursor: not-allowed;
			}
		</style>
		
		<div class="file-drop-zone {{if .DragActive}}drag-active{{end}}"
			ondragover="event.preventDefault(); send_event('{{.IdComponent}}', 'DragOver', '')"
			ondragleave="event.preventDefault(); send_event('{{.IdComponent}}', 'DragLeave', '')"
			ondrop="event.preventDefault(); 
				var files = Array.from(event.dataTransfer.files).map(f => ({
					name: f.name,
					size: f.size,
					type: f.type
				}));
				var readers = [];
				event.dataTransfer.files.forEach((file, i) => {
					var reader = new FileReader();
					readers.push(new Promise(resolve => {
						reader.onload = e => {
							// Use text for JSON files, data URL for others
							if (file.type === 'application/json' || file.name.endsWith('.json')) {
								files[i].data = e.target.result;
							} else {
								files[i].data = e.target.result;
							}
							resolve();
						};
						// Read as text for JSON files, as data URL for others
						if (file.type === 'application/json' || file.name.endsWith('.json')) {
							reader.readAsText(file);
						} else {
							reader.readAsDataURL(file);
						}
					}));
				});
				Promise.all(readers).then(() => {
					send_event('{{.IdComponent}}', 'Drop', JSON.stringify(files));
				});
			"
			onclick="document.getElementById('{{.IdComponent}}_input').click()"
		>
			<div class="file-icon">📁</div>
			<div class="file-label">{{.Label}}</div>
			<div class="file-hint">
				{{if .Accept}}Accepted: {{.Accept}}{{end}}
				{{if .MaxSize}} | Max size: {{.MaxSize}} bytes{{end}}
				{{if .Multiple}} | Multiple files allowed{{end}}
			</div>
		</div>
		
		<input 
			id="{{.IdComponent}}_input"
			type="file" 
			class="file-input-hidden"
			{{if .Multiple}}multiple{{end}}
			{{if .Accept}}accept="{{.Accept}}"{{end}}
			onchange="
				var files = Array.from(this.files).map(f => ({
					name: f.name,
					size: f.size,
					type: f.type
				}));
				var readers = [];
				Array.from(this.files).forEach((file, i) => {
					var reader = new FileReader();
					readers.push(new Promise(resolve => {
						reader.onload = e => {
							files[i].data = e.target.result;
							resolve();
						};
						// Read as text for JSON files, as data URL for others
						if (file.type === 'application/json' || file.name.endsWith('.json')) {
							reader.readAsText(file);
						} else {
							reader.readAsDataURL(file);
						}
					}));
				});
				Promise.all(readers).then(() => {
					send_event('{{.IdComponent}}', 'Select', JSON.stringify(files));
				});
			"
		>
		
		{{if .Files}}
		<div class="file-list">
			{{range $i, $file := .Files}}
			<div class="file-item">
				{{if $file.Preview}}
				<img src="{{$file.Preview}}" class="file-preview">
				{{else}}
				<div class="file-preview" style="background: #f0f0f0; display: flex; align-items: center; justify-content: center;">
					📄
				</div>
				{{end}}
				<div class="file-info">
					<div class="file-name">{{$file.Name}}</div>
					<div class="file-size">{{$file.Size}} bytes</div>
					{{if $file.Progress}}
					<div class="file-progress">
						<div class="file-progress-bar" style="width: {{$file.Progress}}%"></div>
					</div>
					{{end}}
				</div>
				<button class="file-remove" onclick="event.stopPropagation(); send_event('{{$.IdComponent}}', 'RemoveFile', '{{$i}}')">
					Remove
				</button>
			</div>
			{{end}}
		</div>
		{{end}}
		
		{{if .Files}}
		<button class="upload-button" onclick="send_event('{{.IdComponent}}', 'Upload', '')">
			Upload {{len .Files}} file(s)
		</button>
		{{end}}
	</div>
	`
}

func (f *FileUpload) GetDriver() liveview.LiveDriver {
	return f
}

func (f *FileUpload) DragOver(data interface{}) {
	f.DragActive = true
	f.Commit()
}

func (f *FileUpload) DragLeave(data interface{}) {
	f.DragActive = false
	f.Commit()
}

func (f *FileUpload) Drop(data interface{}) {
	f.DragActive = false
	f.handleFiles(data)
}

func (f *FileUpload) Select(data interface{}) {
	f.handleFiles(data)
}

func (f *FileUpload) handleFiles(data interface{}) {
	var files []map[string]interface{}
	if str, ok := data.(string); ok {
		files = parseJSONFiles(str)
	} else if arr, ok := data.([]map[string]interface{}); ok {
		files = arr
	}
	
	for _, file := range files {
		if len(f.Files) >= f.MaxFiles {
			f.EvalScript(fmt.Sprintf("alert('Maximum %d files allowed')", f.MaxFiles))
			break
		}
		
		size := int64(0)
		if s, ok := file["size"].(float64); ok {
			size = int64(s)
		}
		
		if size > f.MaxSize {
			f.EvalScript(fmt.Sprintf("alert('File %s exceeds maximum size of %d bytes')", file["name"], f.MaxSize))
			continue
		}
		
		fileInfo := FileInfo{
			Name: fmt.Sprint(file["name"]),
			Size: size,
			Type: fmt.Sprint(file["type"]),
			Data: fmt.Sprint(file["data"]),
		}
		
		if strings.HasPrefix(fileInfo.Type, "image/") {
			fileInfo.Preview = fileInfo.Data
		}
		
		f.Files = append(f.Files, fileInfo)
	}
	
	f.Commit()
}

func (f *FileUpload) RemoveFile(data interface{}) {
	index := 0
	if str, ok := data.(string); ok {
		fmt.Sscanf(str, "%d", &index)
	}
	
	if index >= 0 && index < len(f.Files) {
		f.Files = append(f.Files[:index], f.Files[index+1:]...)
		f.Commit()
	}
}

func (f *FileUpload) Upload(data interface{}) {
	if f.OnUpload != nil {
		if err := f.OnUpload(f.Files); err != nil {
			f.EvalScript(fmt.Sprintf("alert('Upload failed: %s')", err.Error()))
		} else {
			f.Files = []FileInfo{}
			f.EvalScript("alert('Upload successful!')")
			f.Commit()
		}
	}
}

func (f *FileUpload) Clear() {
	f.Files = []FileInfo{}
	f.UploadProgress = 0
	f.Commit()
}

func (f *FileUpload) GetFileData(index int) ([]byte, error) {
	if index < 0 || index >= len(f.Files) {
		return nil, fmt.Errorf("invalid file index")
	}
	
	data := f.Files[index].Data
	
	// Check if it's a data URL (base64 encoded)
	if strings.HasPrefix(data, "data:") {
		// Extract base64 part after the comma
		if idx := strings.IndexByte(data, ','); idx != -1 {
			data = data[idx+1:]
		}
		return base64.StdEncoding.DecodeString(data)
	}
	
	// If not a data URL, assume it's plain text (like JSON)
	return []byte(data), nil
}

func parseJSONFiles(data string) []map[string]interface{} {
	data = strings.TrimSpace(data)
	if !strings.HasPrefix(data, "[") {
		data = "[" + data + "]"
	}
	
	var files []map[string]interface{}
	// Parse JSON data
	if err := json.Unmarshal([]byte(data), &files); err != nil {
		// If parsing fails, return empty array
		return []map[string]interface{}{}
	}
	return files
}