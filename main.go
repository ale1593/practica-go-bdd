package main

import (
	"flag"
	"fmt"
	"os"

	"prj_withdrawal/headers"
)

func main() {
	// Definir banderas para los argumentos de línea de comandos.
	nombre := flag.String("nombre", "", "Nombre del titular de la cuenta")
	saldo := flag.Float64("saldo", 0.0, "Saldo inicial de la cuenta")
	retiro := flag.Float64("retiro", 0.0, "Cantidad a retirar")

	// Nuevas banderas para la operación de abono.
	abono := flag.Float64("abono", 0.0, "Cantidad a abonar")
	operacion := flag.String("operacion", "retiro", "Operación: retiro o abono")

	// Bandera para seleccionar el escenario de prueba.
	escenario := flag.String("test", "", "Tipo de test: exito, fallo, error")

	flag.Parse()

	// Si no se pasan argumentos, ejecutamos los escenarios por defecto.
	if len(os.Args) == 1 {
		fmt.Println("=== EJECUTANDO PRUEBAS POR DEFECTO (Gherkin v1.0.0) ===")
		headers.TestEscenario1Exito("Alice", 234.56, 200.00)
		headers.TestEscenario2Fallo("Hamza", 198.76, 200.00)
		os.Exit(0)
	}

	// Validar que se haya ingresado el nombre.
	if *nombre == "" {
		fmt.Println("Error: Debes proporcionar al menos -nombre. Usa -h para ayuda.")
		os.Exit(1)
	}

	// NUEVO: Enrutamiento de la operación de abono.
	if *operacion == "abono" {
		if *abono > 0 {
			headers.TestEscenario4AbonoExitoso(*nombre, *saldo, *abono)
		} else if *abono < 0 {
			headers.TestEscenario5AbonoNegativo(*nombre, *saldo, *abono)
		} else {
			headers.TestEscenario6AbonoCero(*nombre, *saldo, *abono)
		}
		os.Exit(0)
	}

	// Validar el tipo de operación.
	if *operacion != "retiro" {
		fmt.Println("Error: Operación no válida. Usa 'retiro' o 'abono'.")
		os.Exit(1)
	}

	// Enrutamiento de escenarios de retiro.
	switch *escenario {
	case "exito":
		headers.TestEscenario1Exito(*nombre, *saldo, *retiro)

	case "fallo":
		headers.TestEscenario2Fallo(*nombre, *saldo, *retiro)

	case "error":
		headers.TestEscenario3Error(*nombre, *saldo, *retiro)

	default:
		// Si no se especifica escenario, deducimos qué pasará según los datos.
		fmt.Println("=== EJECUTANDO TEST DINÁMICO AUTOMÁTICO ===")

		if *retiro < 0 {
			headers.TestEscenario3Error(*nombre, *saldo, *retiro)
		} else if *retiro > *saldo {
			headers.TestEscenario2Fallo(*nombre, *saldo, *retiro)
		} else {
			headers.TestEscenario1Exito(*nombre, *saldo, *retiro)
		}
	}
	os.Exit(0)
}
