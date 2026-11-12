
Feature: Abono de efectivo
  Como cliente titular de una cuenta bancaria
  Quiero depositar dinero en efectivo
  Para incrementar mi saldo disponible

  Rule: El monto del abono debe ser mayor que cero

  Scenario: Abono exitoso
    Given Carlos tiene 100.00 en su cuenta
    When Carlos realiza un abono de 50.00
    Then el abono se realiza con éxito
    And el saldo actualizado es 150.00

  Scenario: Abono rechazado por monto negativo
    Given Maria tiene 100.00 en su cuenta
    When Maria intenta abonar -20.00
    Then el sistema muestra un error de validación
    And el saldo permanece en 100.00

  Scenario: Abono rechazado por monto cero
    Given Pedro tiene 100.00 en su cuenta
    When Pedro intenta abonar 0.00
    Then el sistema muestra un error de validación
    And el saldo permanece en 100.00
