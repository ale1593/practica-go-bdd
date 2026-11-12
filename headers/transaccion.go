package headers

import (
	"encoding/json"
	"os"
	"path/filepath"
	"time"
)

// Transaccion representa una operación bancaria realizada
type Transaccion struct {
	ID             int       `json:"id"`
	Nombre         string    `json:"nombre"`
	Tipo           string    `json:"tipo"`
	Monto          float64   `json:"monto"`
	Resultado      string    `json:"resultado"`
	Mensaje        string    `json:"mensaje"`
	SaldoAnterior  float64   `json:"saldo_anterior"`
	SaldoPosterior float64   `json:"saldo_posterior"`
	FechaHora      time.Time `json:"fecha_hora"`
}

// GuardarTransacciones: Almacena el historial de transacciones en un archivo JSON
func GuardarTransacciones(transacciones []Transaccion, rutaArchivo string) error {
	// Crear el directorio si todavía no existe
	directorio := filepath.Dir(rutaArchivo)

	if err := os.MkdirAll(directorio, 0755); err != nil {
		return err
	}

	// Convertir las transacciones a formato JSON.
	datos, err := json.MarshalIndent(transacciones, "", "    ")
	if err != nil {
		return err
	}

	// Guardar el contenido en el archivo.
	return os.WriteFile(rutaArchivo, datos, 0644)
}

// RegistrarTransaccion: Agrega una operación al historial existente
func RegistrarTransaccion(transaccion Transaccion, rutaArchivo string) error {
	var transacciones []Transaccion

	// Leer el historial existente
	datos, err := os.ReadFile(rutaArchivo)

	if err != nil && !os.IsNotExist(err) {
		return err
	}

	// Recuperar las transacciones anteriores
	if err == nil && len(datos) > 0 {
		if err := json.Unmarshal(datos, &transacciones); err != nil {
			return err
		}
	}

	// Asignar un identificador consecutivo
	transaccion.ID = len(transacciones) + 1

	// Registrar la fecha y hora si no fue especificada
	if transaccion.FechaHora.IsZero() {
		transaccion.FechaHora = time.Now()
	}

	// Agregar la nueva transacción al historial
	transacciones = append(transacciones, transaccion)

	// Guardar el historial actualizado
	return GuardarTransacciones(transacciones, rutaArchivo)
}
