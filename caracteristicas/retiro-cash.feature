
Feature: Retiro de efectivo
  Como cliente titular de una cuenta bancaria
  Quiero retirar dinero de mi cuenta
  Para disponer de efectivo cuando tenga saldo suficiente

  Rule: Los clientes no pueden retirar más de su saldo

  Scenario: Retiro exitoso dentro del saldo
    Given Alice tiene 234.56 en su cuenta
    When Alice intenta retirar 200.00
    Then el retiro se realiza con éxito
    And el saldo restante es 34.56

  Scenario: Retiro rechazado por saldo insuficiente
    Given Hamza tiene 198.76 en su cuenta
    When Hamza intenta retirar 200.00
    Then el retiro se rechaza
    And el saldo permanece en 198.76

  Scenario: Retiro rechazado por monto negativo
    Given Pedro tiene 300.00 en su cuenta
    When Pedro intenta retirar -20.00
    Then el sistema muestra un error de validación
    And el saldo permanece en 300.00
