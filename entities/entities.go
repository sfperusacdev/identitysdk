package entities

type JwtData struct {
	Session Session `json:"session"`
	Jwt     Jwt     `json:"jwt"`
}

type Jwt struct {
	Empresa           string `json:"empresa"`
	ReferenciaEmpresa string `json:"referencia"`
	UsuarioCodigo     string `json:"usuario_codigo"`
	Username          string `json:"username"`
	UsuarioReff       string `json:"usuario_reff"`
	IntegrationURL    string `json:"integration_url"`
	TabajadorCodigo   string `json:"tabajador_codigo"`
}

type Session struct {
	Company      string       `json:"company"`
	Username     string       `json:"username"`
	Supervisors  []string     `json:"supervisors"`
	Subordinates []string     `json:"subordinates"`
	Permissions  []Permission `json:"permissions"`
}

type Permission struct {
	ID             string   `json:"id"`
	CompanyBrances []string `json:"company_brances"`
}
