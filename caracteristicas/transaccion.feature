
Feature: Registro de transacciones bancarias
  Como sistema bancario
  Quiero registrar las operaciones de retiro y abono
  Para mantener un historial de las transacciones realizadas

  Rule: Toda operación debe registrarse con su resultado en el archivo JSON

  Scenario: Registrar un retiro exitoso
    Given Carlos tiene 100.00 en su cuenta
    When Carlos realiza un retiro de 50.00
    Then el retiro se realiza con éxito
    And la transacción se registra con resultado "Éxito"
    And el saldo actualizado es 50.00

  Scenario: Registrar un retiro rechazado por fondos insuficientes
    Given Maria tiene 100.00 en su cuenta
    When Maria intenta retirar 200.00
    Then el retiro es rechazado por fondos insuficientes
    And la transacción se registra con resultado "Fallo"
    And el saldo permanece en 100.00

  Scenario: Registrar un retiro con monto negativo
    Given Pedro tiene 100.00 en su cuenta
    When Pedro intenta retirar -20.00
    Then el sistema muestra un error de validación
    And la transacción se registra con resultado "Error"
    And el saldo permanece en 100.00

  Scenario: Registrar un abono exitoso
    Given Ana tiene 100.00 en su cuenta
    When Ana realiza un abono de 50.00
    Then el abono se realiza con éxito
    And la transacción se registra con resultado "Éxito"
    And el saldo actualizado es 150.00

  Scenario: Registrar un abono con monto negativo
    Given Luis tiene 100.00 en su cuenta
    When Luis intenta abonar -20.00
    Then el sistema muestra un error de validación
    And la transacción se registra con resultado "Error"
    And el saldo permanece en 100.00

  Scenario: Registrar un abono con monto cero
    Given Sofia tiene 100.00 en su cuenta
    When Sofia intenta abonar 0.00
    Then el sistema muestra un error de validación
    And la transacción se registra con resultado "Error"
    And el saldo permanece en 100.00
