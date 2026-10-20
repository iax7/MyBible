package tui

import (
	"os"
	"strings"
)

// lang is an interface language.
type lang string

const (
	langEnglish lang = "en"
	langSpanish lang = "es"
)

// spanish maps each English interface string to its Spanish translation. The
// English text is embedded in the code and is its own key, so a missing entry
// falls back to English with no separate string table.
var spanish = map[string]string{
	// Top bar and Help modal.
	"Book":           "Libro",
	"Chapter":        "Capítulo",
	"Go To":          "Ir a",
	"Translation":    "Traducción",
	"Search":         "Buscar",
	"History":        "Historial",
	"Help":           "Ayuda",
	"Quit":           "Salir",
	"Keys":           "Teclas",
	"Movement":       "Movimiento",
	"Commands":       "Comandos",
	"Back":           "Atrás",
	"Forward":        "Adelante",
	"Toggle top bar": "Mostrar u ocultar la barra",
	"Close":          "Cerrar",

	// Movement keys, shown in the Help modal.
	"Verse down":        "Versículo abajo",
	"Verse up":          "Versículo arriba",
	"Page down":         "Página abajo",
	"Page up":           "Página arriba",
	"Previous Chapter":  "Capítulo anterior",
	"Next Chapter":      "Capítulo siguiente",
	"Top of Chapter":    "Inicio del capítulo",
	"Bottom of Chapter": "Final del capítulo",

	// Modal titles.
	"Select Translation": "Seleccionar traducción",
	"Select Book":        "Seleccionar libro",
	"Select Chapter":     "Seleccionar capítulo",
	"Go to Reference":    "Ir a la referencia",
	"Log":                "Registro",

	// Modal bodies.
	"Type to search":                       "Escribe para buscar",
	"Nothing visited yet":                  "Aún no has visitado nada",
	"Today":                                "Hoy",
	"Yesterday":                            "Ayer",
	"No matches":                           "Sin resultados",
	"No warnings":                          "Sin avisos",
	"enter to select":                      "enter para seleccionar",
	"For example: John, John 3, John 3:16": "Por ejemplo: John, John 3, John 3:16",
	"enter to go · tab to complete":        "enter para ir · tab para completar",
	"That is not a Reference. Try John 3:16.":      "Eso no es una referencia. Prueba con John 3:16.",
	"No such Reference in the Active translation.": "Esa referencia no existe en la traducción activa.",

	// Reading pane.
	"Chapter %d": "Capítulo %d",

	// Empty states and warnings.
	"No Module directory configured.":                            "No hay un directorio de módulos configurado.",
	"Set module_directory in %s":                                 "Define module_directory en %s",
	"to the folder holding your MyBible *.SQLite3 modules.":      "en la carpeta que contiene tus módulos MyBible *.SQLite3.",
	"No Bible Modules found in %s.":                              "No se encontraron módulos de la Biblia en %s.",
	"Put your *.SQLite3 modules there, then restart the reader.": "Coloca allí tus módulos *.SQLite3 y reinicia el lector.",
	"No Active translation.":                                     "No hay una traducción activa.",
	"Press %s to choose one.":                                    "Pulsa %s para elegir una.",
}

// resolveLanguage picks the interface language: the config value first, then
// the system locale, then English.
func resolveLanguage(configured string) lang {
	for _, value := range []string{configured, os.Getenv("LC_ALL"), os.Getenv("LC_MESSAGES"), os.Getenv("LANG")} {
		switch {
		case strings.HasPrefix(value, "es"):
			return langSpanish
		case strings.HasPrefix(value, "en"):
			return langEnglish
		}
	}
	return langEnglish
}

// translate returns text in the reader's language, falling back to the English
// source text when no translation exists.
func (m Model) translate(text string) string {
	if translated, ok := spanish[text]; ok && m.lang == langSpanish {
		return translated
	}
	return text
}
