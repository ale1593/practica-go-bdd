package handlers

import (
	"encoding/json"
	"html/template"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"

	"prj_withdrawal/headers"
)

func nuevoServidorPrueba(t *testing.T) (*Servidor, string, string) {
	t.Helper()

	plantillas, err := template.ParseGlob(filepath.Join("..", "templates", "*.html"))
	if err != nil {
		t.Fatalf("no se pudieron cargar las plantillas: %v", err)
	}
	directorio := t.TempDir()
	rutaCuentas := filepath.Join(directorio, "cuentas.json")
	rutaHistorial := filepath.Join(directorio, "transacciones.json")
	if err := os.WriteFile(rutaHistorial, []byte(`[{"id":1,"nombre":"Histórico","tipo":"Retiro","monto":1,"resultado":"Éxito"}]`), 0600); err != nil {
		t.Fatalf("no se pudo crear el historial de prueba: %v", err)
	}
	servidor, err := nuevoServidorConDatos(plantillas, rutaCuentas, rutaHistorial)
	if err != nil {
		t.Fatalf("no se pudo iniciar el servidor de prueba: %v", err)
	}
	return servidor, rutaCuentas, rutaHistorial
}

func enviarFormulario(
	t *testing.T,
	handler http.HandlerFunc,
	ruta string,
	valores url.Values,
	cookie *http.Cookie,
) *httptest.ResponseRecorder {
	t.Helper()
	request := httptest.NewRequest(http.MethodPost, ruta, strings.NewReader(valores.Encode()))
	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	if cookie != nil {
		request.AddCookie(cookie)
	}
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, request)
	return response
}

func registrarUsuario(t *testing.T, servidor *Servidor, nombre, usuario, contrasena string) {
	t.Helper()
	formulario := url.Values{
		"nombre":     {nombre},
		"usuario":    {usuario},
		"contrasena": {contrasena},
	}
	response := enviarFormulario(t, servidor.Registro, "/registro", formulario, nil)
	if response.Code != http.StatusSeeOther {
		t.Fatalf("registro de %q falló: estado %d, respuesta %s", usuario, response.Code, response.Body)
	}
}

func iniciarSesion(t *testing.T, servidor *Servidor, usuario, contrasena string) (*httptest.ResponseRecorder, *http.Cookie) {
	t.Helper()
	formulario := url.Values{"usuario": {usuario}, "contrasena": {contrasena}}
	response := enviarFormulario(t, servidor.Login, "/login", formulario, nil)
	for _, cookie := range response.Result().Cookies() {
		if cookie.Name == nombreCookie {
			return response, cookie
		}
	}
	return response, nil
}

func TestRegistroLoginYHashDeContrasena(t *testing.T) {
	servidor, rutaCuentas, _ := nuevoServidorPrueba(t)
	registrarUsuario(t, servidor, "Ana Pérez", " Ana_1 ", "ClaveSegura123")

	datos, err := os.ReadFile(rutaCuentas)
	if err != nil {
		t.Fatalf("no se pudo leer el archivo de cuentas: %v", err)
	}
	var cuentas []cuentaPersistida
	if err := json.Unmarshal(datos, &cuentas); err != nil {
		t.Fatalf("el archivo de cuentas no contiene JSON válido: %v", err)
	}
	if len(cuentas) != 1 {
		t.Fatalf("se esperaba una cuenta, se obtuvieron %d", len(cuentas))
	}
	if cuentas[0].Saldo != 0 || cuentas[0].Usuario != "ana_1" || cuentas[0].ID == "" {
		t.Errorf("datos iniciales de la cuenta incorrectos: %+v", cuentas[0])
	}
	if strings.Contains(string(datos), "ClaveSegura123") || !strings.HasPrefix(cuentas[0].HashContrasena, "pbkdf2-sha256$") {
		t.Error("la contraseña no está almacenada exclusivamente como hash")
	}

	response, cookie := iniciarSesion(t, servidor, "ana_1", "Incorrecta123")
	if response.Code != http.StatusSeeOther || cookie != nil {
		t.Fatal("se aceptaron credenciales incorrectas")
	}

	request := httptest.NewRequest(http.MethodPost, "/login", strings.NewReader(url.Values{
		"usuario":    {"ana_1"},
		"contrasena": {"ClaveSegura123"},
	}.Encode()))
	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	request.Header.Set("X-Forwarded-Proto", "https")
	loginResponse := httptest.NewRecorder()
	servidor.Login(loginResponse, request)
	cookies := loginResponse.Result().Cookies()
	if loginResponse.Code != http.StatusSeeOther || len(cookies) == 0 {
		t.Fatal("no se pudo iniciar sesión con credenciales correctas")
	}
	cookie = cookies[0]
	if !cookie.HttpOnly || cookie.SameSite != http.SameSiteLaxMode || !cookie.Secure {
		t.Errorf("atributos de cookie inseguros: %+v", cookie)
	}

	panel := httptest.NewRecorder()
	homeRequest := httptest.NewRequest(http.MethodGet, "/", nil)
	homeRequest.AddCookie(cookie)
	servidor.Inicio(panel, homeRequest)
	if panel.Code != http.StatusOK || !strings.Contains(panel.Body.String(), "Ana Pérez") {
		t.Error("la sesión no abrió el panel de la cuenta autenticada")
	}
}

func TestAislamientoDeCuentasOperacionesYHistorial(t *testing.T) {
	servidor, rutaCuentas, _ := nuevoServidorPrueba(t)
	registrarUsuario(t, servidor, "Ana", "ana", "ClaveSegura123")
	registrarUsuario(t, servidor, "Bruno", "bruno", "OtraClave1234")

	_, cookieAna := iniciarSesion(t, servidor, "ana", "ClaveSegura123")
	if cookieAna == nil {
		t.Fatal("no se pudo iniciar sesión con Ana")
	}
	_, cookieBruno := iniciarSesion(t, servidor, "bruno", "OtraClave1234")
	if cookieBruno == nil {
		t.Fatal("no se pudo iniciar sesión con Bruno")
	}

	formulario := url.Values{"monto": {"100"}, "usuario": {"bruno"}, "saldo": {"999999"}}
	response := enviarFormulario(t, servidor.Abonar, "/abono", formulario, cookieAna)
	if response.Code != http.StatusSeeOther {
		t.Fatalf("el abono de Ana falló: %d", response.Code)
	}
	response = enviarFormulario(t, servidor.Retirar, "/retiro", url.Values{"monto": {"40"}}, cookieAna)
	if response.Code != http.StatusSeeOther {
		t.Fatalf("el retiro de Ana falló: %d", response.Code)
	}
	response = enviarFormulario(t, servidor.Retirar, "/retiro", url.Values{"monto": {"1000"}}, cookieAna)
	if response.Code != http.StatusSeeOther || !strings.Contains(response.Header().Get("Location"), "Fallo") {
		t.Fatal("no se clasificó correctamente el retiro rechazado")
	}

	historialAna := httptest.NewRecorder()
	requestAna := httptest.NewRequest(http.MethodGet, "/historial", nil)
	requestAna.AddCookie(cookieAna)
	servidor.Historial(historialAna, requestAna)
	if historialAna.Code != http.StatusOK {
		t.Fatalf("el historial de Ana devolvió %d", historialAna.Code)
	}
	var transaccionesAna []headers.Transaccion
	if err := json.Unmarshal(historialAna.Body.Bytes(), &transaccionesAna); err != nil {
		t.Fatalf("JSON de historial inválido: %v", err)
	}
	if len(transaccionesAna) != 3 {
		t.Fatalf("el historial de Ana debe tener 3 operaciones, tiene %d", len(transaccionesAna))
	}
	for _, transaccion := range transaccionesAna {
		if transaccion.CuentaID != servidor.cuentas[0].ID {
			t.Errorf("transacción asociada a una cuenta incorrecta: %q", transaccion.CuentaID)
		}
	}

	historialBruno := httptest.NewRecorder()
	requestBruno := httptest.NewRequest(http.MethodGet, "/historial", nil)
	requestBruno.AddCookie(cookieBruno)
	servidor.Historial(historialBruno, requestBruno)
	var transaccionesBruno []headers.Transaccion
	if err := json.Unmarshal(historialBruno.Body.Bytes(), &transaccionesBruno); err != nil {
		t.Fatalf("JSON de historial inválido: %v", err)
	}
	if len(transaccionesBruno) != 0 {
		t.Fatalf("el historial de Bruno expuso operaciones ajenas: %d", len(transaccionesBruno))
	}

	panelBruno := httptest.NewRecorder()
	homeBruno := httptest.NewRequest(http.MethodGet, "/", nil)
	homeBruno.AddCookie(cookieBruno)
	servidor.Inicio(panelBruno, homeBruno)
	if !strings.Contains(panelBruno.Body.String(), "Bruno") ||
		!strings.Contains(panelBruno.Body.String(), "$0.00") ||
		strings.Contains(panelBruno.Body.String(), "Retiro realizado correctamente.") {
		t.Error("el panel de Bruno no está aislado del saldo o historial de Ana")
	}

	datos, err := os.ReadFile(rutaCuentas)
	if err != nil {
		t.Fatalf("no se pudieron leer las cuentas: %v", err)
	}
	var cuentas []cuentaPersistida
	if err := json.Unmarshal(datos, &cuentas); err != nil {
		t.Fatalf("JSON de cuentas inválido: %v", err)
	}
	if len(cuentas) != 2 || cuentas[0].ID == cuentas[1].ID {
		t.Fatal("las cuentas no tienen identificadores únicos")
	}
	if cuentas[0].Saldo != 60 || cuentas[1].Saldo != 0 {
		t.Errorf("los saldos no son independientes: %.2f y %.2f", cuentas[0].Saldo, cuentas[1].Saldo)
	}

	servidorReiniciado, err := nuevoServidorConDatos(servidor.Plantillas, rutaCuentas, servidor.rutaHistorial)
	if err != nil {
		t.Fatalf("no se pudieron volver a cargar las cuentas: %v", err)
	}
	if servidorReiniciado.cuentas[0].Saldo != 60 || servidorReiniciado.cuentas[1].Saldo != 0 {
		t.Fatal("los saldos no persistieron al reiniciar el servidor")
	}

	logout := httptest.NewRecorder()
	requestLogout := httptest.NewRequest(http.MethodPost, "/logout", nil)
	requestLogout.AddCookie(cookieAna)
	servidor.CerrarSesion(logout, requestLogout)
	if logout.Code != http.StatusSeeOther || !strings.Contains(logout.Header().Get("Set-Cookie"), "Max-Age=0") {
		t.Error("el cierre de sesión no invalidó la cookie")
	}
	private := httptest.NewRecorder()
	requestPrivate := httptest.NewRequest(http.MethodGet, "/historial", nil)
	requestPrivate.AddCookie(cookieAna)
	servidor.Historial(private, requestPrivate)
	if private.Code != http.StatusSeeOther {
		t.Error("se permitió consultar el historial con una sesión cerrada")
	}
}

func TestRegistroValidaNombreDeUsuarioDuplicado(t *testing.T) {
	servidor, _, _ := nuevoServidorPrueba(t)
	registrarUsuario(t, servidor, "Ana", "ana", "ClaveSegura123")
	response := enviarFormulario(t, servidor.Registro, "/registro", url.Values{
		"nombre":     {"Otra Ana"},
		"usuario":    {"ANA"},
		"contrasena": {"OtraClave1234"},
	}, nil)
	if !strings.Contains(response.Header().Get("Location"), "ya+est%C3%A1+registrado") {
		t.Errorf("no se rechazó el nombre de usuario duplicado: %s", response.Header().Get("Location"))
	}
}

func TestOperacionesPrivadasRequierenAutenticacion(t *testing.T) {
	servidor, _, _ := nuevoServidorPrueba(t)
	response := enviarFormulario(t, servidor.Retirar, "/retiro", url.Values{"monto": {"10"}}, nil)
	if response.Code != http.StatusSeeOther {
		t.Fatalf("se esperaba redirección al login, se obtuvo %d", response.Code)
	}
	if len(servidor.cuentas) != 0 {
		t.Error("una operación no autenticada alteró las cuentas")
	}
}

func TestOperacionesConcurrentesConservanSaldoEHistorial(t *testing.T) {
	servidor, _, _ := nuevoServidorPrueba(t)
	registrarUsuario(t, servidor, "Cuenta Concurrente", "concurrente", "ClaveSegura123")
	_, cookie := iniciarSesion(t, servidor, "concurrente", "ClaveSegura123")
	if cookie == nil {
		t.Fatal("no se pudo iniciar sesión")
	}

	const operaciones = 12
	var grupo sync.WaitGroup
	errores := make(chan int, operaciones)
	for i := 0; i < operaciones; i++ {
		grupo.Add(1)
		go func() {
			defer grupo.Done()
			request := httptest.NewRequest(http.MethodPost, "/abono", strings.NewReader("monto=1"))
			request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
			request.AddCookie(cookie)
			response := httptest.NewRecorder()
			servidor.Abonar(response, request)
			if response.Code != http.StatusSeeOther {
				errores <- response.Code
			}
		}()
	}
	grupo.Wait()
	close(errores)
	for status := range errores {
		t.Errorf("operación concurrente devolvió %d", status)
	}

	if saldo := servidor.cuentas[0].Saldo; saldo != operaciones {
		t.Errorf("saldo final incorrecto: esperado %d, obtenido %.2f", operaciones, saldo)
	}
	historial := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/historial", nil)
	request.AddCookie(cookie)
	servidor.Historial(historial, request)
	var transacciones []headers.Transaccion
	if err := json.Unmarshal(historial.Body.Bytes(), &transacciones); err != nil {
		t.Fatalf("historial concurrente inválido: %v", err)
	}
	if len(transacciones) != operaciones {
		t.Errorf("faltan transacciones: esperado %d, obtenido %d", operaciones, len(transacciones))
	}
}
