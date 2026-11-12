package headers

import (
	"math"
	"testing"
)

// Pruebas unitarias TDD
// TestRetirarExitosamente: Verifica que el retiro descuente el saldo correcto
func TestRetirarExitosamente(t *testing.T) {
	cuenta := NewCuenta("Carlos", 500.00)

	exito, err := cuenta.Retirar(150.00)

	if err != nil {
		t.Fatalf("No se esperaba un error: %v", err)
	}

	if !exito {
		t.Fatal("Se esperaba que el retiro fuera exitoso")
	}

	saldoEsperado := 350.00

	if math.Abs(cuenta.ObtenerSaldo()-saldoEsperado) > 0.001 {
		t.Errorf("Saldo incorrecto: se esperaba %.2f, se obtuvo %.2f",
			saldoEsperado, cuenta.ObtenerSaldo())
	}
}

// TestRetirarFondosInsuficientes: Verifica que el retiro sea rechazado
// cuando la cantidad supera el saldo disponible
func TestRetirarFondosInsuficientes(t *testing.T) {
	cuenta := NewCuenta("Maria", 50.00)

	exito, err := cuenta.Retirar(100.00)

	if err != nil {
		t.Fatalf("No se esperaba un error: %v", err)
	}

	if exito {
		t.Error("El retiro debía ser rechazado por fondos insuficientes")
	}

	if cuenta.ObtenerSaldo() != 50.00 {
		t.Errorf("El saldo debía permanecer en 50.00, se obtuvo %.2f",
			cuenta.ObtenerSaldo())
	}
}

// TestRetirarCantidadNegativa: Verifica que no se permitan
// cantidades negativas
func TestRetirarCantidadNegativa(t *testing.T) {
	cuenta := NewCuenta("Pedro", 300.00)

	exito, err := cuenta.Retirar(-20.00)

	if err == nil {
		t.Fatal("Se esperaba un error por cantidad negativa")
	}

	if exito {
		t.Error("El retiro no debía realizarse")
	}

	if cuenta.ObtenerSaldo() != 300.00 {
		t.Errorf("El saldo debía permanecer en 300.00, se obtuvo %.2f",
			cuenta.ObtenerSaldo())
	}
}

// TestAbonarExitosamente verifica que el abono incremente el saldo.
func TestAbonarExitosamente(t *testing.T) {
	cuenta := NewCuenta("Carlos", 100.00)

	err := cuenta.Abonar(50.00)

	if err != nil {
		t.Fatalf("No se esperaba un error: %v", err)
	}

	saldoEsperado := 150.00

	if math.Abs(cuenta.ObtenerSaldo()-saldoEsperado) > 0.001 {
		t.Errorf("Saldo incorrecto: se esperaba %.2f, se obtuvo %.2f",
			saldoEsperado, cuenta.ObtenerSaldo())
	}
}

// TestAbonarCantidadNegativa verifica que no se acepten montos negativos.
func TestAbonarCantidadNegativa(t *testing.T) {
	cuenta := NewCuenta("Maria", 100.00)

	err := cuenta.Abonar(-20.00)

	if err == nil {
		t.Fatal("Se esperaba un error por cantidad negativa")
	}

	if cuenta.ObtenerSaldo() != 100.00 {
		t.Errorf("El saldo debía permanecer en 100.00, se obtuvo %.2f",
			cuenta.ObtenerSaldo())
	}
}

// TestAbonarCantidadCero verifica que no se acepten montos iguales a cero.
func TestAbonarCantidadCero(t *testing.T) {
	cuenta := NewCuenta("Pedro", 100.00)

	err := cuenta.Abonar(0.00)

	if err == nil {
		t.Fatal("Se esperaba un error por cantidad igual a cero")
	}

	if cuenta.ObtenerSaldo() != 100.00 {
		t.Errorf("El saldo debía permanecer en 100.00, se obtuvo %.2f",
			cuenta.ObtenerSaldo())
	}
}
