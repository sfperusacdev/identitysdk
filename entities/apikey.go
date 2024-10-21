package entities

type ApikeyData struct {
	Apikey Apikey `json:"jwt"`
}

type Apikey struct {
	Empresa string `json:"empresa"`
	Zona    string `json:"zona"`
}
