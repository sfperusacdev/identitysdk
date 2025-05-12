package models

type BasicSystemProperty struct {
	ID    string `json:"id"`
	Value string `json:"value,omitempty"`
}

type DetailedSystemProperty struct {
	ID          string `xml:"id" json:"id"`
	Title       string `xml:"title" json:"title"`
	Group       string `xml:"group,omitempty" json:"group,omitempty"`
	Description string `xml:"description,omitempty" json:"description,omitempty"`
	Type        string `xml:"type,omitempty" json:"type,omitempty"`
	Value       string `xml:"value,omitempty" json:"value,omitempty"`
}
