package sunat

type NumeroALetras interface {
	ToWords(number float64, decimals int) (string, error)
	ToMoney(number float64, decimals int, currency, cents string) (string, error)
	ToInvoice(number float64, decimals int, currency string) (string, error)
}
