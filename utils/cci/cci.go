package cci

import (
	"errors"
	"unicode"
)

type Bank struct {
	ID          string
	Description string
	Swift       string
}

var (
	ErrInvalidCCILength = errors.New("cci must be 20 digits long")
	ErrCCINotNumeric    = errors.New("cci must contain only digits")
	ErrUnknownBankCode  = errors.New("unknown bank code")
)

var banksByCode = map[string]Bank{
	"002": {ID: "BCP", Description: "Banco de Credito del Peru", Swift: "BCPLPEPL"},
	"003": {ID: "BINT", Description: "Interbank", Swift: "BINPPEPL"},
	"009": {ID: "BSC", Description: "Scotiabank Peru", Swift: "BSUDPEPL"},
	"011": {ID: "BCON", Description: "BBVA Peru", Swift: "BCONPEPL"},
	"018": {ID: "BN", Description: "Banco de la Nacion", Swift: "BNAPPEPL"},
	"038": {ID: "BBF", Description: "BanBif", Swift: "BANBPEPL"},
	"043": {ID: "BFAL", Description: "Banco Falabella Peru", Swift: "FALAPEP"},
	"053": {ID: "BRIP", Description: "Banco Ripley", Swift: ""},
	"054": {ID: "BCST", Description: "Banco Santander Peru", Swift: ""},
	"056": {ID: "BPCH", Description: "Banco Pichincha Peru", Swift: "PICHP EPL"},
	"058": {ID: "BGNB", Description: "Banco GNB Peru", Swift: "GNBPPEPL"},
	"060": {ID: "BICB", Description: "ICBC Peru", Swift: "ICBKPEPL"},
	"061": {ID: "BALF", Description: "Banco Alfin", Swift: ""},
	"062": {ID: "BCOM", Description: "Banco Compartamos", Swift: ""},
	"063": {ID: "BCOM", Description: "Banco de Comercio", Swift: "BCOMPEPL"},
	"064": {ID: "BCCB", Description: "China Construction Bank Peru", Swift: "PCBCPEPL"},
	"065": {ID: "BBTG", Description: "BTG Pactual Peru", Swift: "BBTGPEPL"},
}

func ResolveBankFromCCI(cci string) (Bank, error) {
	if len(cci) != 20 {
		return Bank{}, ErrInvalidCCILength
	}

	for _, r := range cci {
		if !unicode.IsDigit(r) {
			return Bank{}, ErrCCINotNumeric
		}
	}

	code := cci[:3]

	bank, ok := banksByCode[code]
	if !ok {
		return Bank{}, ErrUnknownBankCode
	}

	return bank, nil
}
