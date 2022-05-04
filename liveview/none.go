package liveview

type None struct {
	*ComponentDriver[*None]
	Template string
}

func (t *None) GetDriver() LiveDriver {
	return t
}

func (t *None) Start() {
	t.Commit()
}

func (t *None) GetTemplate() string {
	return t.Template
}
