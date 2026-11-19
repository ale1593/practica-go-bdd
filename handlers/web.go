package handlers

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"errors"
	"html/template"
	"log"
	"math"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"
	"unicode/utf8"

	"prj_withdrawal/headers"
)

const (
	rutaCuentas       = "datos/cuentas.json"
	rutaTransacciones = "datos/transacciones.json"
	nombreCookie      = "sesion_bancaria"
	duracionSesion    = 8 * time.Hour
)

// Servidor conecta los manejadores HTTP con las cuentas y el historial.
type Servidor struct {
	Plantillas    *template.Template
	cuentas       []cuentaPersistida
	sesiones      map[string]sesion
	rutaCuentas   string
	rutaHistorial string
	mu            sync.Mutex
}

type sesion struct {
	CuentaID string
	Expira   time.Time
}

type datosInicio struct {
	Nombre        string
	Usuario       string
	IDCuenta      string
	Saldo         float64
	Transacciones []headers.Transaccion
	Estado        string
	Mensaje       string
}

type datosAutenticacion struct {
	Estado  string
	Mensaje string
}

// NuevoServidor carga las plantillas y las cuentas persistidas.
func NuevoServidor() (*Servidor, error) {
	var plantillas *template.Template
	var err error
	for _, ruta := range []string{
		filepath.Join("templates", "*.html"),
		filepath.Join("..", "templates", "*.html"),
	} {
		plantillas, err = template.ParseGlob(ruta)
		if err == nil {
			break
		}
	}
	if err != nil {
		return nil, err
	}
	return nuevoServidorConDatos(plantillas, rutaCuentas, rutaTransacciones)
}

func nuevoServidorConDatos(plantillas *template.Template, rutaArchivoCuentas, rutaArchivoTransacciones string) (*Servidor, error) {
	cuentas, err := cargarCuentas(rutaArchivoCuentas)
	if err != nil {
		return nil, err
	}
	return &Servidor{
		Plantillas:    plantillas,
		cuentas:       cuentas,
		sesiones:      make(map[string]sesion),
		rutaCuentas:   rutaArchivoCuentas,
		rutaHistorial: rutaArchivoTransacciones,
	}, nil
}

// Inicio presenta el panel privado o el formulario de inicio de sesión.
func (s *Servidor) Inicio(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.Header().Set("Allow", http.MethodGet)
		http.Error(w, "Método no permitido", http.StatusMethodNotAllowed)
		return
	}
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}

	s.mu.Lock()
	cuenta, autenticado := s.cuentaDeSesionLocked(r)
	if !autenticado {
		s.mu.Unlock()
		s.renderizar(w, "login.html", datosAutenticacion{
			Estado:  estadoValido(r.URL.Query().Get("estado")),
			Mensaje: r.URL.Query().Get("mensaje"),
		})
		return
	}
	transacciones, err := s.cargarTransacciones()
	if err == nil {
		transacciones = filtrarTransacciones(transacciones, cuenta.ID)
	}
	s.mu.Unlock()
	if err != nil {
		log.Printf("no se pudo cargar el historial de transacciones: %v", err)
		http.Error(w, "No se pudo cargar la información bancaria.", http.StatusInternalServerError)
		return
	}

	s.renderizar(w, "index.html", datosInicio{
		Nombre:        cuenta.Nombre,
		Usuario:       cuenta.Usuario,
		IDCuenta:      cuenta.ID,
		Saldo:         cuenta.Saldo,
		Transacciones: transacciones,
		Estado:        estadoValido(r.URL.Query().Get("estado")),
		Mensaje:       r.URL.Query().Get("mensaje"),
	})
}

// Registro muestra y procesa el formulario de alta de cuentas.
func (s *Servidor) Registro(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet && r.Method != http.MethodPost {
		w.Header().Set("Allow", "GET, POST")
		http.Error(w, "Método no permitido", http.StatusMethodNotAllowed)
		return
	}
	if s.estaAutenticado(r) {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}
	if r.Method == http.MethodGet {
		s.renderizar(w, "registro.html", datosAutenticacion{
			Estado:  estadoValido(r.URL.Query().Get("estado")),
			Mensaje: r.URL.Query().Get("mensaje"),
		})
		return
	}
	if err := r.ParseForm(); err != nil {
		http.Error(w, "No se pudo leer el formulario.", http.StatusBadRequest)
		return
	}

	nombre := strings.TrimSpace(r.FormValue("nombre"))
	usuario := normalizarUsuario(r.FormValue("usuario"))
	contrasena := r.FormValue("contrasena")
	if mensaje := validarRegistro(nombre, usuario, contrasena); mensaje != "" {
		redireccionConMensaje(w, r, "/registro", "Error", mensaje)
		return
	}
	hash, err := hashContrasena(contrasena)
	if err != nil {
		log.Printf("no se pudo generar el hash de contraseña: %v", err)
		http.Error(w, "No se pudo crear la cuenta. Inténtalo más tarde.", http.StatusInternalServerError)
		return
	}
	id, err := crearID()
	if err != nil {
		log.Printf("no se pudo generar el identificador de cuenta: %v", err)
		http.Error(w, "No se pudo crear la cuenta. Inténtalo más tarde.", http.StatusInternalServerError)
		return
	}

	s.mu.Lock()
	for _, existente := range s.cuentas {
		if existente.Usuario == usuario {
			s.mu.Unlock()
			redireccionConMensaje(w, r, "/registro", "Error", "Ese nombre de usuario ya está registrado.")
			return
		}
	}
	nuevaCuenta := cuentaPersistida{
		ID:             id,
		Nombre:         nombre,
		Usuario:        usuario,
		HashContrasena: hash,
		Saldo:          0,
	}
	for s.idDeCuentaExisteLocked(id) {
		id, err = crearID()
		if err != nil {
			s.mu.Unlock()
			log.Printf("no se pudo generar un identificador de cuenta único: %v", err)
			http.Error(w, "No se pudo crear la cuenta. Inténtalo más tarde.", http.StatusInternalServerError)
			return
		}
		nuevaCuenta.ID = id
	}
	cuentasActualizadas := append(append([]cuentaPersistida(nil), s.cuentas...), nuevaCuenta)
	if err := guardarCuentas(s.rutaCuentas, cuentasActualizadas); err != nil {
		s.mu.Unlock()
		log.Printf("no se pudo guardar la nueva cuenta: %v", err)
		http.Error(w, "No se pudo crear la cuenta. Inténtalo más tarde.", http.StatusInternalServerError)
		return
	}
	s.cuentas = cuentasActualizadas
	s.mu.Unlock()

	redireccionConMensaje(w, r, "/", "Éxito", "Cuenta creada. Ya puedes iniciar sesión.")
}

// Login valida las credenciales e inicia una sesión de servidor.
func (s *Servidor) Login(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.Header().Set("Allow", http.MethodPost)
		http.Error(w, "Método no permitido", http.StatusMethodNotAllowed)
		return
	}
	if err := r.ParseForm(); err != nil {
		http.Error(w, "No se pudo leer el formulario.", http.StatusBadRequest)
		return
	}

	usuario := normalizarUsuario(r.FormValue("usuario"))
	contrasena := r.FormValue("contrasena")
	s.mu.Lock()
	var cuenta cuentaPersistida
	encontrada := false
	for _, existente := range s.cuentas {
		if existente.Usuario == usuario {
			cuenta = existente
			encontrada = true
			break
		}
	}
	s.mu.Unlock()

	valida := false
	if encontrada {
		var err error
		valida, err = verificarContrasena(contrasena, cuenta.HashContrasena)
		if err != nil {
			log.Printf("no se pudo verificar el hash de contraseña de la cuenta %s: %v", cuenta.ID, err)
		}
	}
	if !valida {
		redireccionConMensaje(w, r, "/", "Error", "Usuario o contraseña incorrectos.")
		return
	}

	token, err := crearTokenSesion()
	if err != nil {
		log.Printf("no se pudo crear la sesión: %v", err)
		http.Error(w, "No se pudo iniciar sesión. Inténtalo más tarde.", http.StatusInternalServerError)
		return
	}
	s.mu.Lock()
	ahora := time.Now()
	for token, sesionActual := range s.sesiones {
		if ahora.After(sesionActual.Expira) {
			delete(s.sesiones, token)
		}
	}
	s.sesiones[token] = sesion{CuentaID: cuenta.ID, Expira: time.Now().Add(duracionSesion)}
	s.mu.Unlock()
	http.SetCookie(w, nuevaCookieSesion(r, token, int(duracionSesion.Seconds())))
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

// CerrarSesion invalida la sesión actual.
func (s *Servidor) CerrarSesion(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.Header().Set("Allow", http.MethodPost)
		http.Error(w, "Método no permitido", http.StatusMethodNotAllowed)
		return
	}
	if cookie, err := r.Cookie(nombreCookie); err == nil {
		s.mu.Lock()
		delete(s.sesiones, cookie.Value)
		s.mu.Unlock()
	}
	cookie := nuevaCookieSesion(r, "", -1)
	cookie.Expires = time.Unix(1, 0)
	http.SetCookie(w, cookie)
	redireccionConMensaje(w, r, "/", "Éxito", "Has cerrado sesión.")
}

// Retirar procesa un retiro para la cuenta autenticada.
func (s *Servidor) Retirar(w http.ResponseWriter, r *http.Request) {
	s.procesarOperacion(w, r, "Retiro")
}

// Abonar procesa un abono para la cuenta autenticada.
func (s *Servidor) Abonar(w http.ResponseWriter, r *http.Request) {
	s.procesarOperacion(w, r, "Abono")
}

func (s *Servidor) procesarOperacion(w http.ResponseWriter, r *http.Request, tipo string) {
	if r.Method != http.MethodPost {
		w.Header().Set("Allow", http.MethodPost)
		http.Error(w, "Método no permitido", http.StatusMethodNotAllowed)
		return
	}
	if !s.estaAutenticado(r) {
		redireccionConMensaje(w, r, "/", "Error", "Inicia sesión para continuar.")
		return
	}
	if err := r.ParseForm(); err != nil {
		http.Error(w, "No se pudo leer el monto enviado.", http.StatusBadRequest)
		return
	}
	monto, errMonto := strconv.ParseFloat(strings.TrimSpace(r.FormValue("monto")), 64)
	if errMonto == nil && (math.IsNaN(monto) || math.IsInf(monto, 0)) {
		errMonto = errors.New("el monto no es finito")
	}

	s.mu.Lock()
	cuenta, autenticado := s.cuentaDeSesionLocked(r)
	if !autenticado {
		s.mu.Unlock()
		redireccionConMensaje(w, r, "/", "Error", "Inicia sesión para continuar.")
		return
	}

	cuentaNegocio := headers.NewCuenta(cuenta.Nombre, cuenta.Saldo)
	resultado := "Éxito"
	mensaje := ""
	operacionExitosa := false
	if errMonto != nil {
		resultado = "Error"
		mensaje = "El monto ingresado no es válido."
		monto = 0
	} else if tipo == "Retiro" {
		exito, err := cuentaNegocio.Retirar(monto)
		switch {
		case err != nil:
			resultado = "Error"
			mensaje = err.Error()
		case !exito:
			resultado = "Fallo"
			mensaje = "Retiro rechazado por fondos insuficientes."
		default:
			mensaje = "Retiro realizado correctamente."
			operacionExitosa = true
		}
	} else {
		if err := cuentaNegocio.Abonar(monto); err != nil {
			resultado = "Error"
			mensaje = err.Error()
		} else {
			mensaje = "Abono realizado correctamente."
			operacionExitosa = true
		}
	}

	saldoAnterior := cuenta.Saldo
	cuenta.Saldo = cuentaNegocio.ObtenerSaldo()
	if operacionExitosa {
		cuentasActualizadas := append([]cuentaPersistida(nil), s.cuentas...)
		for i := range cuentasActualizadas {
			if cuentasActualizadas[i].ID == cuenta.ID {
				cuentasActualizadas[i].Saldo = cuenta.Saldo
				break
			}
		}
		if err := guardarCuentas(s.rutaCuentas, cuentasActualizadas); err != nil {
			s.mu.Unlock()
			log.Printf("no se pudo guardar el saldo de la cuenta %s: %v", cuenta.ID, err)
			http.Error(w, "No se pudo completar la operación. Inténtalo más tarde.", http.StatusInternalServerError)
			return
		}
		s.cuentas = cuentasActualizadas
	}

	transaccion := headers.Transaccion{
		CuentaID:       cuenta.ID,
		Nombre:         cuenta.Nombre,
		Tipo:           tipo,
		Monto:          monto,
		Resultado:      resultado,
		Mensaje:        mensaje,
		SaldoAnterior:  saldoAnterior,
		SaldoPosterior: cuenta.Saldo,
	}
	if err := headers.RegistrarTransaccion(transaccion, s.rutaHistorial); err != nil {
		if operacionExitosa {
			if errRollback := guardarCuentas(s.rutaCuentas, s.cuentasPreviasLocked(cuenta.ID, saldoAnterior)); errRollback != nil {
				log.Printf("no se pudo revertir el saldo de la cuenta %s: %v", cuenta.ID, errRollback)
			} else {
				s.actualizarSaldoLocked(cuenta.ID, saldoAnterior)
			}
		}
		s.mu.Unlock()
		log.Printf("no se pudo registrar la transacción: %v", err)
		http.Error(w, "No se pudo registrar la operación. Inténtalo más tarde.", http.StatusInternalServerError)
		return
	}
	s.mu.Unlock()

	redireccionConMensaje(w, r, "/", resultado, mensaje)
}

// Historial devuelve únicamente las transacciones de la cuenta autenticada.
func (s *Servidor) Historial(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.Header().Set("Allow", http.MethodGet)
		http.Error(w, "Método no permitido", http.StatusMethodNotAllowed)
		return
	}
	s.mu.Lock()
	cuenta, autenticado := s.cuentaDeSesionLocked(r)
	if !autenticado {
		s.mu.Unlock()
		redireccionConMensaje(w, r, "/", "Error", "Inicia sesión para consultar el historial.")
		return
	}
	transacciones, err := s.cargarTransacciones()
	if err == nil {
		transacciones = filtrarTransacciones(transacciones, cuenta.ID)
	}
	s.mu.Unlock()
	if err != nil {
		log.Printf("no se pudo cargar el historial de transacciones: %v", err)
		http.Error(w, "No se pudo cargar el historial.", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("Cache-Control", "no-store")
	if err := json.NewEncoder(w).Encode(transacciones); err != nil {
		log.Printf("no se pudo enviar el historial de transacciones: %v", err)
	}
}

func (s *Servidor) cargarTransacciones() ([]headers.Transaccion, error) {
	datos, err := os.ReadFile(s.rutaHistorial)
	if errors.Is(err, os.ErrNotExist) {
		return []headers.Transaccion{}, nil
	}
	if err != nil {
		return nil, err
	}
	if len(datos) == 0 {
		return []headers.Transaccion{}, nil
	}
	var transacciones []headers.Transaccion
	if err := json.Unmarshal(datos, &transacciones); err != nil {
		return nil, err
	}
	return transacciones, nil
}

func (s *Servidor) cuentaDeSesionLocked(r *http.Request) (cuentaPersistida, bool) {
	cookie, err := r.Cookie(nombreCookie)
	if err != nil {
		return cuentaPersistida{}, false
	}
	sesionActual, existe := s.sesiones[cookie.Value]
	if !existe {
		return cuentaPersistida{}, false
	}
	if time.Now().After(sesionActual.Expira) {
		delete(s.sesiones, cookie.Value)
		return cuentaPersistida{}, false
	}
	for _, cuenta := range s.cuentas {
		if cuenta.ID == sesionActual.CuentaID {
			return cuenta, true
		}
	}
	delete(s.sesiones, cookie.Value)
	return cuentaPersistida{}, false
}

func (s *Servidor) estaAutenticado(r *http.Request) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	_, autenticado := s.cuentaDeSesionLocked(r)
	return autenticado
}

func (s *Servidor) cuentasPreviasLocked(id string, saldoAnterior float64) []cuentaPersistida {
	cuentas := append([]cuentaPersistida(nil), s.cuentas...)
	for i := range cuentas {
		if cuentas[i].ID == id {
			cuentas[i].Saldo = saldoAnterior
			break
		}
	}
	return cuentas
}

func (s *Servidor) actualizarSaldoLocked(id string, saldo float64) {
	for i := range s.cuentas {
		if s.cuentas[i].ID == id {
			s.cuentas[i].Saldo = saldo
			return
		}
	}
}

func (s *Servidor) idDeCuentaExisteLocked(id string) bool {
	for _, cuenta := range s.cuentas {
		if cuenta.ID == id {
			return true
		}
	}
	return false
}

func (s *Servidor) renderizar(w http.ResponseWriter, plantilla string, datos any) {
	w.Header().Set("Cache-Control", "no-store")
	if err := s.Plantillas.ExecuteTemplate(w, plantilla, datos); err != nil {
		log.Printf("no se pudo renderizar %s: %v", plantilla, err)
		http.Error(w, "No se pudo cargar la página.", http.StatusInternalServerError)
	}
}

func crearTokenSesion() (string, error) {
	datos := make([]byte, 32)
	if _, err := rand.Read(datos); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(datos), nil
}

func nuevaCookieSesion(r *http.Request, valor string, maxAge int) *http.Cookie {
	return &http.Cookie{
		Name:     nombreCookie,
		Value:    valor,
		Path:     "/",
		HttpOnly: true,
		Secure:   r.TLS != nil || strings.EqualFold(r.Header.Get("X-Forwarded-Proto"), "https"),
		SameSite: http.SameSiteLaxMode,
		MaxAge:   maxAge,
	}
}

func normalizarUsuario(usuario string) string {
	return strings.ToLower(strings.TrimSpace(usuario))
}

func validarRegistro(nombre, usuario, contrasena string) string {
	if nombre == "" || utf8.RuneCountInString(nombre) > 100 {
		return "El nombre debe tener entre 1 y 100 caracteres."
	}
	if mensaje := validarUsuario(usuario); mensaje != "" {
		return mensaje
	}
	if utf8.RuneCountInString(contrasena) < 8 || utf8.RuneCountInString(contrasena) > 128 {
		return "La contraseña debe tener entre 8 y 128 caracteres."
	}
	return ""
}

func validarUsuario(usuario string) string {
	if len(usuario) < 3 || len(usuario) > 32 {
		return "El usuario debe tener entre 3 y 32 caracteres."
	}
	for _, caracter := range usuario {
		if !(caracter >= 'a' && caracter <= 'z') &&
			!(caracter >= '0' && caracter <= '9') &&
			caracter != '.' && caracter != '_' && caracter != '-' {
			return "El usuario solo puede contener letras, números, punto, guion y guion bajo."
		}
	}
	return ""
}

func filtrarTransacciones(transacciones []headers.Transaccion, cuentaID string) []headers.Transaccion {
	filtradas := make([]headers.Transaccion, 0)
	for _, transaccion := range transacciones {
		if transaccion.CuentaID == cuentaID {
			filtradas = append(filtradas, transaccion)
		}
	}
	return filtradas
}

func estadoValido(estado string) string {
	if estado == "Éxito" || estado == "Fallo" || estado == "Error" {
		return estado
	}
	return ""
}

func redireccionConMensaje(w http.ResponseWriter, r *http.Request, destino, estado, mensaje string) {
	valores := url.Values{
		"estado":  {estado},
		"mensaje": {mensaje},
	}
	http.Redirect(w, r, destino+"?"+valores.Encode(), http.StatusSeeOther)
}
