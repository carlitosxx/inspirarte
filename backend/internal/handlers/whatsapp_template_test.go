package handlers

import (
	"strings"
	"testing"
)

func TestValidateWhatsAppTemplate(t *testing.T) {
	tests := []struct {
		name     string
		template string
		valid    bool
	}{
		{"texto y variables repetidas", "Hola {nombre}, {nombre}: {fecha} a las {hora} para {personas} personas", true},
		{"texto manual", "Hola, confirma tu reserva por favor", true},
		{"variable desconocida", "Hola {correo}", false},
		{"variable abierta", "Hola {nombre", false},
		{"cierre suelto", "Hola nombre}", false},
		{"vacia", "", false},
		{"demasiado larga", strings.Repeat("a", 2001), false},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got := validateWhatsAppTemplate(test.template)
			if (got == "") != test.valid {
				t.Fatalf("validación inesperada: %q", got)
			}
		})
	}
}
