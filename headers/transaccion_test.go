package headers

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"
)

// Verifica que una transacción se registre correctamente
func TestRegistrarTransaccion(t *testing.T) {
	rutaArchivo := filepath.Join(t.TempDir(), "transacciones.json")

	transaccion := Transaccion{
		Nombre:         "Carlos",
		Tipo:           "Abono",
		Monto:          50.00,
		Resultado:      "Éxito",
		Mensaje:        "Abono realizado correctamente",
		SaldoAnterior:  100.00,
		SaldoPosterior: 150.00,
	}

	err := RegistrarTransaccion(transaccion, rutaArchivo)
	if err != nil {
		t.Fatalf("Error al registrar transacción: %v", err)
	}

	datos, err := os.ReadFile(rutaArchivo)
	if err != nil {
		t.Fatalf("Error al leer archivo JSON: %v", err)
	}

	var transacciones []Transaccion
	if err := json.Unmarshal(datos, &transacciones); err != nil {
		t.Fatalf("Error al interpretar JSON: %v", err)
	}

	if len(transacciones) != 1 {
		t.Fatalf("Se esperaba 1 transacción, se obtuvieron %d", len(transacciones))
	}

	if transacciones[0].ID != 1 {
		t.Errorf("Se esperaba ID 1, se obtuvo %d", transacciones[0].ID)
	}

	if transacciones[0].Nombre != "Carlos" {
		t.Errorf("Nombre incorrecto: %s", transacciones[0].Nombre)
	}

	if transacciones[0].FechaHora.IsZero() {
		t.Error("La fecha y hora no fueron asignadas")
	}
}

// Verifica que se puedan registrar varias transacciones
func TestConservarHistorialTransacciones(t *testing.T) {
	rutaArchivo := filepath.Join(t.TempDir(), "transacciones.json")

	primera := Transaccion{
		Nombre:    "Carlos",
		Tipo:      "Abono",
		Monto:     50.00,
		Resultado: "Éxito",
	}

	segunda := Transaccion{
		Nombre:    "Maria",
		Tipo:      "Retiro",
		Monto:     200.00,
		Resultado: "Fallo",
	}

	if err := RegistrarTransaccion(primera, rutaArchivo); err != nil {
		t.Fatalf("Error al registrar primera transacción: %v", err)
	}

	if err := RegistrarTransaccion(segunda, rutaArchivo); err != nil {
		t.Fatalf("Error al registrar segunda transacción: %v", err)
	}

	datos, err := os.ReadFile(rutaArchivo)
	if err != nil {
		t.Fatalf("Error al leer historial: %v", err)
	}

	var transacciones []Transaccion
	if err := json.Unmarshal(datos, &transacciones); err != nil {
		t.Fatalf("Error al interpretar JSON: %v", err)
	}

	if len(transacciones) != 2 {
		t.Fatalf("Se esperaban 2 transacciones, se obtuvieron %d", len(transacciones))
	}

	if transacciones[0].ID != 1 || transacciones[1].ID != 2 {
		t.Error("Los identificadores no son consecutivos")
	}

	if transacciones[0].Nombre != "Carlos" || transacciones[1].Nombre != "Maria" {
		t.Error("No se conservaron correctamente las transacciones anteriores")
	}
}

// Verifica que una fecha especificada se conserve
func TestConservarFechaTransaccion(t *testing.T) {
	rutaArchivo := filepath.Join(t.TempDir(), "transacciones.json")

	fecha := time.Date(2026, 11, 9, 15, 30, 0, 0, time.UTC)

	transaccion := Transaccion{
		Nombre:    "Pedro",
		Tipo:      "Retiro",
		Monto:     30.00,
		Resultado: "Éxito",
		FechaHora: fecha,
	}

	if err := RegistrarTransaccion(transaccion, rutaArchivo); err != nil {
		t.Fatalf("Error al registrar transacción: %v", err)
	}

	datos, err := os.ReadFile(rutaArchivo)
	if err != nil {
		t.Fatalf("Error al leer archivo JSON: %v", err)
	}

	var transacciones []Transaccion
	if err := json.Unmarshal(datos, &transacciones); err != nil {
		t.Fatalf("Error al interpretar JSON: %v", err)
	}

	if !transacciones[0].FechaHora.Equal(fecha) {
		t.Error("La fecha original de la transacción no se conservó")
	}
}
