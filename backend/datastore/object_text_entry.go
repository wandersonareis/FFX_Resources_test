package datastore

import "ffxresources/backend/models"

// ObjectTextEntry é a representação original, neutra de formato, do texto de
// um objeto de localização: um ID mais os textos por campo e idioma. Cada
// formato de saída (JSON, CSV, ...) organiza essas entries à sua maneira.
//
// Movido de objectsfile (antes JSONEntry): o tipo é o dado original que os
// formatters recebem, não uma estrutura JSON.
type ObjectTextEntry struct {
	ID                    int                    `json:"id"`
	Name                  map[string]string      `json:"name,omitempty"`
	SimplifiedName        map[string]string      `json:"simplifiedName,omitempty"`
	Description           map[string]string      `json:"description,omitempty"`
	SimplifiedDescription map[string]string      `json:"simplifiedDescription,omitempty"`
	Effect                map[string]string      `json:"effect,omitempty"`
	EffectDescription     map[string]string      `json:"effectDescription,omitempty"`
	Bonus                 map[string]string      `json:"bonus,omitempty"`
	BonusIconA            map[string]string      `json:"bonusIconA,omitempty"`
	BonusIconB            map[string]string      `json:"bonusIconB,omitempty"`
	BonusReserve          map[string]string      `json:"bonusReserve,omitempty"`
	Abilities             []map[string]string    `json:"abilities,omitempty"`
	SensorText            map[string]string      `json:"sensorText,omitempty"`
	SimplifiedSensorText  map[string]string      `json:"simplifiedSensorText,omitempty"`
	ScanText              map[string]string      `json:"scanText,omitempty"`
	SimplifiedScanText    map[string]string      `json:"simplifiedScanText,omitempty"`
	Weapons               map[string]WeaponTexts `json:"weapons,omitempty"`
}

// WeaponTexts agrupa os textos de uma arma por campo.
type WeaponTexts struct {
	Name           map[string]string `json:"name"`
	SimplifiedName map[string]string `json:"simplifiedName"`
}

// HasContent indica se a entry carrega algum texto em qualquer campo.
func (e *ObjectTextEntry) HasContent() bool {
	if len(e.Name) > 0 || len(e.SimplifiedName) > 0 ||
		len(e.Description) > 0 || len(e.SimplifiedDescription) > 0 ||
		len(e.Effect) > 0 || len(e.EffectDescription) > 0 ||
		len(e.Bonus) > 0 || len(e.BonusIconA) > 0 ||
		len(e.BonusIconB) > 0 || len(e.BonusReserve) > 0 ||
		len(e.SensorText) > 0 || len(e.SimplifiedSensorText) > 0 ||
		len(e.ScanText) > 0 || len(e.SimplifiedScanText) > 0 {
		return true
	}
	for _, ab := range e.Abilities {
		if len(ab) > 0 {
			return true
		}
	}
	for _, wt := range e.Weapons {
		if len(wt.Name) > 0 || len(wt.SimplifiedName) > 0 {
			return true
		}
	}
	return false
}

// ObjectTextData carrega as entries originais mais os metadados do binário de
// origem. O domínio monta, cada formato decide como renderizar.
type ObjectTextData struct {
	Entries  []*ObjectTextEntry
	Metadata *models.FileMetadata
}
