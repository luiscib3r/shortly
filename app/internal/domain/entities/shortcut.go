package entities

type Shortcut struct {
	id  string
	url string
}

func NewShortcut(
	id string,
	url string,
) *Shortcut {
	return &Shortcut{
		id:  id,
		url: url,
	}
}

func (s Shortcut) Id() string {
	return s.id
}

func (s Shortcut) Url() string {
	return s.url
}
