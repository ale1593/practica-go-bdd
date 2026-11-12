package headers

import (
	"fmt"
	"math"
)

// imprimirSaldo muestra el saldo formateado a dos decimales.
func imprimirSaldo(cuenta *Cuenta) {
	fmt.Printf("Saldo de %s: %.2f\n", cuenta.Nombre(), cuenta.ObtenerSaldo())
}

// ------------------------------------------------------------
// Validación de escenarios Gherkin dinámicos
// ------------------------------------------------------------

// TestEscenario1Exito: Simula un retiro exitoso con datos dinámicos
func TestEscenario1Exito(nombre string, saldoInicial float64, cantidadRetirar float64) {
	fmt.Printf("--- Escenario 1: Successful withdrawal within balance (%s) ---\n", nombre)

	// Given
	cuenta := NewCuenta(nombre, saldoInicial)
	fmt.Printf("Given: %s tiene %.2f en su cuenta\n", nombre, saldoInicial)

	// When
	fmt.Printf("When: %s intenta retirar %.2f\n", nombre, cantidadRetirar)
	exito, err := cuenta.Retirar(cantidadRetirar)

	// Then
	if err != nil {
		fmt.Printf("Resultado: FALLO. Error inesperado: %v\n", err)
		return
	}

	if exito {
		fmt.Println("Resultado: ÉXITO. La retirada fue procesada.")
		imprimirSaldo(cuenta)

		saldoEsperado := saldoInicial - cantidadRetirar

		if math.Abs(cuenta.ObtenerSaldo()-saldoEsperado) < 0.001 {
			fmt.Printf("Validación: El saldo restante es correcto (%.2f).\n\n", saldoEsperado)

			// Registrar la transacción exitosa.
			registrarResultadoTransaccion(
				cuenta,
				"Retiro",
				cantidadRetirar,
				"Éxito",
				"Retiro realizado correctamente",
				saldoInicial,
			)
		} else {
			fmt.Println("ERROR: El saldo restante no coincide.")
		}
	} else {
		fmt.Println("Resultado: FALLO. La retirada fue rechazada.")
	}
}

// TestEscenario2Fallo: Simula un retiro rechazado con datos dinámicos
func TestEscenario2Fallo(nombre string, saldoInicial float64, cantidadRetirar float64) {
	fmt.Printf("--- Escenario 2: Declined withdrawal in excess of balance (%s) ---\n", nombre)

	// Given
	cuenta := NewCuenta(nombre, saldoInicial)
	fmt.Printf("Given: %s tiene %.2f en su cuenta\n", nombre, saldoInicial)

	// When
	fmt.Printf("When: %s intenta retirar %.2f\n", nombre, cantidadRetirar)
	exito, err := cuenta.Retirar(cantidadRetirar)

	// Then
	if err != nil {
		fmt.Printf("Resultado: FALLO. Error inesperado: %v\n", err)
		return
	}

	if !exito {
		fmt.Println("Resultado: ÉXITO. La retirada fue rechazada correctamente.")
		imprimirSaldo(cuenta)

		if math.Abs(cuenta.ObtenerSaldo()-saldoInicial) < 0.001 {
			fmt.Printf("Validación: El saldo no ha cambiado (%.2f).\n\n", saldoInicial)

			// Registrar el retiro rechazado por fondos insuficientes.
			registrarResultadoTransaccion(
				cuenta,
				"Retiro",
				cantidadRetirar,
				"Fallo",
				"Retiro rechazado por fondos insuficientes",
				saldoInicial,
			)
		} else {
			fmt.Println("ERROR: El saldo se modificó incorrectamente.")
		}
	} else {
		fmt.Println("Resultado: FALLO. La retirada fue procesada (debería haber sido rechazada).")
	}
}

// TestEscenario3Error: Simula un error de validación (ej. cantidad negativa)
func TestEscenario3Error(nombre string, saldoInicial float64, cantidadRetirar float64) {
	fmt.Printf("--- Escenario 3: Validation Error (%s) ---\n", nombre)

	// Given
	cuenta := NewCuenta(nombre, saldoInicial)
	fmt.Printf("Given: %s tiene %.2f en su cuenta\n", nombre, saldoInicial)

	// When
	exito, err := cuenta.Retirar(cantidadRetirar)

	// Then
	if err != nil {
		fmt.Printf("Resultado: ÉXITO. El sistema atrapó el error correctamente: %v\n", err)
		imprimirSaldo(cuenta)

		// Registrar el retiro con error de validación.
		registrarResultadoTransaccion(
			cuenta,
			"Retiro",
			cantidadRetirar,
			"Error",
			"Retiro rechazado por monto negativo",
			saldoInicial,
		)

	} else if !exito {
		fmt.Println("Resultado: RECHAZADO. Fondos insuficientes (no es un error de validación).")
	} else {
		fmt.Println("Resultado: FALLO. Se procesó un retiro que debía dar error.")
	}
}

// ------------------------------------------------------------
// Validación de escenarios Gherkin: Abono en efectivo
// ------------------------------------------------------------
// TestEscenario1AbonoExitoso: Verifica un abono válido
func TestEscenario4AbonoExitoso(nombre string, saldoInicial float64, cantidadAbonar float64) {
	fmt.Printf("--- Escenario 4: Successful cash deposit (%s) ---\n", nombre)

	// Given
	cuenta := NewCuenta(nombre, saldoInicial)
	fmt.Printf("Given: %s tiene %.2f en su cuenta\n", nombre, saldoInicial)

	// When
	fmt.Printf("When: %s realiza un abono de %.2f\n", nombre, cantidadAbonar)
	err := cuenta.Abonar(cantidadAbonar)

	// Then
	if err != nil {
		fmt.Printf("Resultado: FALLO. Error inesperado: %v\n", err)
		return
	}

	saldoEsperado := saldoInicial + cantidadAbonar

	if math.Abs(cuenta.ObtenerSaldo()-saldoEsperado) < 0.001 {
		fmt.Println("Resultado: ÉXITO. El abono fue procesado correctamente.")
		imprimirSaldo(cuenta)
		fmt.Printf("Validación: El saldo actualizado es correcto (%.2f).\n", saldoEsperado)

		// Registrar el abono exitoso.
		registrarResultadoTransaccion(
			cuenta,
			"Abono",
			cantidadAbonar,
			"Éxito",
			"Abono realizado correctamente",
			saldoInicial,
		)
	} else {
		fmt.Println("Resultado: FALLO. El saldo actualizado no coincide.")
	}
}

// TestEscenario2AbonoNegativo: Verifica el rechazo de un monto negativo
func TestEscenario5AbonoNegativo(nombre string, saldoInicial float64, cantidadAbonar float64) {
	fmt.Printf("--- Escenario 5: Negative cash deposit (%s) ---\n", nombre)

	// Given
	cuenta := NewCuenta(nombre, saldoInicial)
	fmt.Printf("Given: %s tiene %.2f en su cuenta\n", nombre, saldoInicial)

	// When
	fmt.Printf("When: %s intenta abonar %.2f\n", nombre, cantidadAbonar)
	err := cuenta.Abonar(cantidadAbonar)

	// Then
	if err != nil {
		fmt.Printf("Resultado: ÉXITO. El sistema rechazó el monto: %v\n", err)
		imprimirSaldo(cuenta)

		// Registrar el abono con error de validación.
		registrarResultadoTransaccion(
			cuenta,
			"Abono",
			cantidadAbonar,
			"Error",
			"Abono rechazado por monto negativo",
			saldoInicial,
		)
	} else {
		fmt.Println("Resultado: FALLO. Se aceptó un monto negativo.")
	}
}

// TestEscenario3AbonoCero: Verifica el rechazo de un monto igual a cero
func TestEscenario6AbonoCero(nombre string, saldoInicial float64, cantidadAbonar float64) {
	fmt.Printf("--- Escenario 6: Zero cash deposit (%s) ---\n", nombre)

	// Given
	cuenta := NewCuenta(nombre, saldoInicial)
	fmt.Printf("Given: %s tiene %.2f en su cuenta\n", nombre, saldoInicial)

	// When
	fmt.Printf("When: %s intenta abonar %.2f\n", nombre, cantidadAbonar)
	err := cuenta.Abonar(cantidadAbonar)

	// Then
	if err != nil {
		fmt.Printf("Resultado: ÉXITO. El sistema rechazó el monto: %v\n", err)
		imprimirSaldo(cuenta)

		// Registrar el abono con error de validación.
		registrarResultadoTransaccion(
			cuenta,
			"Abono",
			cantidadAbonar,
			"Error",
			"Abono rechazado por monto cero",
			saldoInicial,
		)
	} else {
		fmt.Println("Resultado: FALLO. Se aceptó un monto igual a cero.")
	}
}

// registrarResultadoTransaccion: Guarda el resultado de una operación bancaria
func registrarResultadoTransaccion(
	cuenta *Cuenta,
	tipo string,
	monto float64,
	resultado string,
	mensaje string,
	saldoAnterior float64,
) {
	transaccion := Transaccion{
		Nombre:         cuenta.Nombre(),
		Tipo:           tipo,
		Monto:          monto,
		Resultado:      resultado,
		Mensaje:        mensaje,
		SaldoAnterior:  saldoAnterior,
		SaldoPosterior: cuenta.ObtenerSaldo(),
	}

	err := RegistrarTransaccion(transaccion, "datos/transacciones.json")

	if err != nil {
		fmt.Printf("ADVERTENCIA: No se pudo registrar la transacción: %v\n", err)
	} else {
		fmt.Println("Registro de transacción guardado correctamente.")
	}
}
