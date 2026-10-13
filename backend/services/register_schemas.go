package services

import "ffxresources/backend/fileFormats/objectsfile/weaponnames"

// init registra os schemas chunkmap de pacotes de política de domínio que
// o chunkmap não pode importar (circular). weaponnames importa chunkmap;
// services importa os dois e é o ponto natural do wiring.
func init() {
	weaponnames.Register()
}