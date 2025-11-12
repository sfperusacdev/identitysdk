package utils

import (
	"slices"
	"strings"
)

func IsAffirmative(input string) bool {
	v := strings.ToLower(strings.TrimSpace(input))
	affirmatives := []string{
		"yes", "y", "true", "t", "1", "on", "ok", "sure", "yeah", "affirmative",
		"si", "sí", "s", "vale", "de acuerdo", "afirmativo",
	}
	return slices.Contains(affirmatives, v)
}

func IsActive(input string) bool {
	v := strings.ToLower(strings.TrimSpace(input))
	actives := []string{
		"activo", "activa", "activos", "on", "si", "sí", "true", "1", "habilitado", "habilitada",
		"enable", "enabled", "ok", "operativo", "funcional", "vigente",
	}
	inactives := []string{
		"inactivo", "inactiva", "inactivos", "off", "no", "false", "0", "deshabilitado", "deshabilitada",
		"disable", "disabled", "inoperativo", "no operativo", "caducado", "suspendido", "baja",
	}
	if slices.Contains(actives, v) {
		return true
	}
	if slices.Contains(inactives, v) {
		return false
	}
	return false
}
