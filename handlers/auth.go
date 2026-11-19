package handlers

import (
	"crypto/pbkdf2"
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

const (
	iteracionesHash = 600_000
	longitudSal     = 16
	longitudHash    = 32
	longitudID      = 16
)

type cuentaPersistida struct {
	ID             string  `json:"id"`
	Nombre         string  `json:"nombre"`
	Usuario        string  `json:"usuario"`
	HashContrasena string  `json:"password_hash"`
	Saldo          float64 `json:"saldo"`
}

func crearID() (string, error) {
	datos := make([]byte, longitudID)
	if _, err := rand.Read(datos); err != nil {
		return "", err
	}
	return hex.EncodeToString(datos), nil
}

func hashContrasena(contrasena string) (string, error) {
	sal := make([]byte, longitudSal)
	if _, err := rand.Read(sal); err != nil {
		return "", err
	}
	hash, err := pbkdf2.Key(sha256.New, contrasena, sal, iteracionesHash, longitudHash)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf(
		"pbkdf2-sha256$%d$%s$%s",
		iteracionesHash,
		base64.RawStdEncoding.EncodeToString(sal),
		base64.RawStdEncoding.EncodeToString(hash),
	), nil
}

func verificarContrasena(contrasena, almacenado string) (bool, error) {
	partes := strings.Split(almacenado, "$")
	if len(partes) != 4 || partes[0] != "pbkdf2-sha256" {
		return false, errors.New("formato de hash de contraseña inválido")
	}
	iteraciones, err := strconv.Atoi(partes[1])
	if err != nil || iteraciones != iteracionesHash {
		return false, errors.New("parámetros de hash de contraseña inválidos")
	}
	sal, err := base64.RawStdEncoding.DecodeString(partes[2])
	if err != nil || len(sal) != longitudSal {
		return false, errors.New("sal de contraseña inválida")
	}
	hashGuardado, err := base64.RawStdEncoding.DecodeString(partes[3])
	if err != nil || len(hashGuardado) != longitudHash {
		return false, errors.New("hash de contraseña inválido")
	}
	hashIngresado, err := pbkdf2.Key(sha256.New, contrasena, sal, iteraciones, longitudHash)
	if err != nil {
		return false, err
	}
	return subtle.ConstantTimeCompare(hashGuardado, hashIngresado) == 1, nil
}

func cargarCuentas(ruta string) ([]cuentaPersistida, error) {
	datos, err := os.ReadFile(ruta)
	if errors.Is(err, os.ErrNotExist) {
		cuentas := []cuentaPersistida{}
		if err := guardarCuentas(ruta, cuentas); err != nil {
			return nil, err
		}
		return cuentas, nil
	}
	if err != nil {
		return nil, err
	}

	var cuentas []cuentaPersistida
	if err := json.Unmarshal(datos, &cuentas); err != nil {
		return nil, err
	}
	if cuentas == nil {
		cuentas = []cuentaPersistida{}
	}

	ids := make(map[string]bool, len(cuentas))
	usuarios := make(map[string]bool, len(cuentas))
	for _, cuenta := range cuentas {
		if cuenta.ID == "" || cuenta.Nombre == "" || cuenta.Usuario == "" || cuenta.HashContrasena == "" {
			return nil, errors.New("el archivo de cuentas contiene un registro incompleto")
		}
		usuarioNormalizado := normalizarUsuario(cuenta.Usuario)
		if usuarioNormalizado != cuenta.Usuario || validarUsuario(usuarioNormalizado) != "" ||
			cuenta.Saldo < 0 || math.IsNaN(cuenta.Saldo) || math.IsInf(cuenta.Saldo, 0) {
			return nil, errors.New("el archivo de cuentas contiene datos inválidos")
		}
		if ids[cuenta.ID] || usuarios[usuarioNormalizado] {
			return nil, errors.New("el archivo de cuentas contiene identificadores duplicados")
		}
		ids[cuenta.ID] = true
		usuarios[usuarioNormalizado] = true
	}
	return cuentas, nil
}

func guardarCuentas(ruta string, cuentas []cuentaPersistida) error {
	return guardarJSONAtomico(ruta, cuentas, 0600)
}

func guardarJSONAtomico(ruta string, valor any, permisos os.FileMode) error {
	directorio := filepath.Dir(ruta)
	if err := os.MkdirAll(directorio, 0700); err != nil {
		return err
	}
	datos, err := json.MarshalIndent(valor, "", "    ")
	if err != nil {
		return err
	}

	temporal, err := os.CreateTemp(directorio, ".cuentas-*.tmp")
	if err != nil {
		return err
	}
	rutaTemporal := temporal.Name()
	defer os.Remove(rutaTemporal)

	if err := temporal.Chmod(permisos); err != nil {
		temporal.Close()
		return err
	}
	if _, err := temporal.Write(datos); err != nil {
		temporal.Close()
		return err
	}
	if err := temporal.Sync(); err != nil {
		temporal.Close()
		return err
	}
	if err := temporal.Close(); err != nil {
		return err
	}
	return os.Rename(rutaTemporal, ruta)
}
