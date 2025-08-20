package components

import (
	"math"

	"github.com/arturoeanton/go-echo-live-view/liveview"
)

type Pagination struct {
	*liveview.ComponentDriver[*Pagination]
	
	CurrentPage  int
	TotalPages   int
	TotalItems   int
	ItemsPerPage int
	MaxButtons   int
	OnPageChange func(page int)
}

func NewPagination(id string, totalItems, itemsPerPage, currentPage int) *Pagination {
	totalPages := int(math.Ceil(float64(totalItems) / float64(itemsPerPage)))
	
	return &Pagination{
		CurrentPage:  currentPage,
		TotalPages:   totalPages,
		TotalItems:   totalItems,
		ItemsPerPage: itemsPerPage,
		MaxButtons:   7,
	}
}

func (p *Pagination) Start() {
	// Events are registered directly on the ComponentDriver
	if p.ComponentDriver != nil {
		p.ComponentDriver.Events["GoToPage"] = func(c *Pagination, data interface{}) {
			c.GoToPage(data)
		}
		p.ComponentDriver.Events["Previous"] = func(c *Pagination, data interface{}) {
			c.Previous(data)
		}
		p.ComponentDriver.Events["Next"] = func(c *Pagination, data interface{}) {
			c.Next(data)
		}
		p.ComponentDriver.Events["First"] = func(c *Pagination, data interface{}) {
			c.First(data)
		}
		p.ComponentDriver.Events["Last"] = func(c *Pagination, data interface{}) {
			c.Last(data)
		}
	}
}

func (p *Pagination) GetTemplate() string {
	return `
<div class="pagination-container" id="{{.IdComponent}}">
	<style>
		.pagination-container {
			display: flex;
			align-items: center;
			gap: 0.5rem;
			padding: 1rem;
		}
		
		.pagination-info {
			margin-right: auto;
			color: #6b7280;
			font-size: 0.875rem;
		}
		
		.pagination-buttons {
			display: flex;
			gap: 0.25rem;
		}
		
		.page-btn {
			padding: 0.5rem 0.75rem;
			border: 1px solid #d1d5db;
			background: white;
			color: #374151;
			cursor: pointer;
			font-size: 0.875rem;
			border-radius: 0.375rem;
			transition: all 0.2s;
			min-width: 2.5rem;
		}
		
		.page-btn:hover:not(:disabled) {
			background: #f3f4f6;
			border-color: #9ca3af;
		}
		
		.page-btn.active {
			background: #3b82f6;
			color: white;
			border-color: #3b82f6;
		}
		
		.page-btn:disabled {
			opacity: 0.5;
			cursor: not-allowed;
		}
		
		.page-ellipsis {
			padding: 0.5rem 0.25rem;
			color: #9ca3af;
		}
	</style>
	
	<div class="pagination-info">
		Showing {{.StartItem}} to {{.EndItem}} of {{.TotalItems}} items
	</div>
	
	<div class="pagination-buttons">
		<button class="page-btn" 
				onclick="send_event('{{.IdComponent}}', 'First', null)"
				{{if eq .CurrentPage 1}}disabled{{end}}>
			First
		</button>
		
		<button class="page-btn" 
				onclick="send_event('{{.IdComponent}}', 'Previous', null)"
				{{if eq .CurrentPage 1}}disabled{{end}}>
			Previous
		</button>
		
		{{range .PageNumbers}}
			{{if eq . -1}}
				<span class="page-ellipsis">...</span>
			{{else}}
				<button class="page-btn {{if eq . $.CurrentPage}}active{{end}}"
						onclick="send_event('{{$.IdComponent}}', 'GoToPage', {{.}})">
					{{.}}
				</button>
			{{end}}
		{{end}}
		
		<button class="page-btn" 
				onclick="send_event('{{.IdComponent}}', 'Next', null)"
				{{if eq .CurrentPage .TotalPages}}disabled{{end}}>
			Next
		</button>
		
		<button class="page-btn" 
				onclick="send_event('{{.IdComponent}}', 'Last', null)"
				{{if eq .CurrentPage .TotalPages}}disabled{{end}}>
			Last
		</button>
	</div>
</div>
`
}

func (p *Pagination) GetDriver() liveview.LiveDriver {
	return p
}

func (p *Pagination) StartItem() int {
	if p.TotalItems == 0 {
		return 0
	}
	return (p.CurrentPage-1)*p.ItemsPerPage + 1
}

func (p *Pagination) EndItem() int {
	end := p.CurrentPage * p.ItemsPerPage
	if end > p.TotalItems {
		end = p.TotalItems
	}
	return end
}

func (p *Pagination) PageNumbers() []int {
	var pages []int
	
	if p.TotalPages <= p.MaxButtons {
		for i := 1; i <= p.TotalPages; i++ {
			pages = append(pages, i)
		}
		return pages
	}
	
	halfMax := p.MaxButtons / 2
	
	if p.CurrentPage <= halfMax {
		for i := 1; i <= p.MaxButtons-2; i++ {
			pages = append(pages, i)
		}
		pages = append(pages, -1)
		pages = append(pages, p.TotalPages)
	} else if p.CurrentPage >= p.TotalPages-halfMax {
		pages = append(pages, 1)
		pages = append(pages, -1)
		for i := p.TotalPages - p.MaxButtons + 3; i <= p.TotalPages; i++ {
			pages = append(pages, i)
		}
	} else {
		pages = append(pages, 1)
		pages = append(pages, -1)
		
		start := p.CurrentPage - halfMax + 2
		end := p.CurrentPage + halfMax - 2
		
		for i := start; i <= end; i++ {
			pages = append(pages, i)
		}
		
		pages = append(pages, -1)
		pages = append(pages, p.TotalPages)
	}
	
	return pages
}

func (p *Pagination) GoToPage(data interface{}) {
	page := int(data.(float64))
	if page >= 1 && page <= p.TotalPages {
		p.CurrentPage = page
		if p.OnPageChange != nil {
			p.OnPageChange(page)
		}
		p.Commit()
	}
}

func (p *Pagination) Previous(data interface{}) {
	if p.CurrentPage > 1 {
		p.CurrentPage--
		if p.OnPageChange != nil {
			p.OnPageChange(p.CurrentPage)
		}
		p.Commit()
	}
}

func (p *Pagination) Next(data interface{}) {
	if p.CurrentPage < p.TotalPages {
		p.CurrentPage++
		if p.OnPageChange != nil {
			p.OnPageChange(p.CurrentPage)
		}
		p.Commit()
	}
}

func (p *Pagination) First(data interface{}) {
	if p.CurrentPage != 1 {
		p.CurrentPage = 1
		if p.OnPageChange != nil {
			p.OnPageChange(1)
		}
		p.Commit()
	}
}

func (p *Pagination) Last(data interface{}) {
	if p.CurrentPage != p.TotalPages {
		p.CurrentPage = p.TotalPages
		if p.OnPageChange != nil {
			p.OnPageChange(p.TotalPages)
		}
		p.Commit()
	}
}

func (p *Pagination) UpdateTotal(totalItems int) {
	p.TotalItems = totalItems
	p.TotalPages = int(math.Ceil(float64(totalItems) / float64(p.ItemsPerPage)))
	
	if p.CurrentPage > p.TotalPages && p.TotalPages > 0 {
		p.CurrentPage = p.TotalPages
	}
	
	p.Commit()
}