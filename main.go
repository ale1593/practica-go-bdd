package main

import (
	"fmt"
	"log"
	"net/http"

	"prj_withdrawal/handlers"
)

func main() {
	servidor, err := handlers.NuevoServidor()
	if err != nil {
		log.Fatalf("no se pudo iniciar la aplicación: %v", err)
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/", servidor.Inicio)
	mux.HandleFunc("/registro", servidor.Registro)
	mux.HandleFunc("/login", servidor.Login)
	mux.HandleFunc("/logout", servidor.CerrarSesion)
	mux.HandleFunc("/retiro", servidor.Retirar)
	mux.HandleFunc("/abono", servidor.Abonar)
	mux.HandleFunc("/historial", servidor.Historial)
	mux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))

	fmt.Println("Sistema Bancario disponible en http://127.0.0.1:8080")
	if err := http.ListenAndServe("127.0.0.1:8080", mux); err != nil {
		log.Fatalf("el servidor web se detuvo: %v", err)
	}
}
