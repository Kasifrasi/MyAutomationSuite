package vorpruefung

import (
	"path/filepath"
	"reflect"
	"sort"
	"strings"
	"testing"

	"github.com/xuri/excelize/v2"
)

// TestAuditBindings prüft "Registry First": jeder in der TemplateRegistry
// deklarierte Input-/Output-Name (inkl. aller Factory-Instanzen) muss in der
// frisch generierten Vorlage als Defined Name existieren. Meldet zwei Klassen von
// Fehlern:
//   - "missing": Registry-Name deklariert, aber kein Defined Name erzeugt.
//   - "uninit":  Factory-Feld ohne Format-String (nie initialisiert).
func TestAuditBindings(t *testing.T) {
	dir := t.TempDir()
	out := filepath.Join(dir, "vorpruefung.xlsx")
	if err := GenerateVorpruefung(out, GeneratorConfig{}); err != nil {
		t.Fatalf("GenerateVorpruefung: %v", err)
	}

	f, err := excelize.OpenFile(out)
	if err != nil {
		t.Fatalf("OpenFile: %v", err)
	}
	defer f.Close()

	defined := map[string]bool{}
	for _, dn := range f.GetDefinedName() {
		defined[dn.Name] = true
	}

	reg := NewTemplateRegistry()
	rv := reflect.ValueOf(*reg)
	rt := rv.Type()

	var missing, uninit []string

	// check erfasst einen erwarteten Namen samt Feld-Kontext.
	check := func(field, name string) {
		if name == "" || defined[name] {
			return
		}
		missing = append(missing, field+" -> "+name)
	}

	for i := 0; i < rv.NumField(); i++ {
		field := rt.Field(i).Name
		switch fv := rv.Field(i).Interface().(type) {
		case InputField:
			check(field, fv.NamedRange)
		case OutputField:
			check(field, fv.NamedRange)
		case InputFactory:
			if fv.Base == "" {
				uninit = append(uninit, field)
				continue
			}
			for p := 1; p <= FBPeriodenAnzahl; p++ {
				check(field, fv.Get(p).NamedRange)
			}
		case OutputFactory:
			if fv.Base == "" {
				uninit = append(uninit, field)
				continue
			}
			for p := 1; p <= FBPeriodenAnzahl; p++ {
				check(field, fv.Get(p).NamedRange)
			}
		case MAInputFactory:
			if fv.Base == "" {
				uninit = append(uninit, field)
				continue
			}
			for id := 1; id <= MA_TABLE_COUNT; id++ {
				p := ((id - 1) % MA_PERIOD_COUNT) + 1
				lvl := ((id - 1) / MA_PERIOD_COUNT) + 1
				check(field, fv.Get(p, lvl).NamedRange)
			}
		case MAOutputFactory:
			if fv.Base == "" {
				uninit = append(uninit, field)
				continue
			}
			for id := 1; id <= MA_TABLE_COUNT; id++ {
				p := ((id - 1) % MA_PERIOD_COUNT) + 1
				lvl := ((id - 1) / MA_PERIOD_COUNT) + 1
				check(field, fv.Get(p, lvl).NamedRange)
			}
		case MAInputKatFactory:
			if fv.Base == "" {
				uninit = append(uninit, field)
				continue
			}
			for p := 1; p <= MA_PERIOD_COUNT; p++ {
				for lvl := 1; lvl <= MA_SLOT_COUNT; lvl++ {
					for row := 1; row <= len(MA_CATEGORIES); row++ {
						check(field, fv.Get(p, lvl, row).NamedRange)
					}
				}
			}
		case MAOutputKatFactory:
			if fv.Base == "" {
				uninit = append(uninit, field)
				continue
			}
			for p := 1; p <= MA_PERIOD_COUNT; p++ {
				for lvl := 1; lvl <= MA_SLOT_COUNT; lvl++ {
					for row := 1; row <= len(MA_CATEGORIES); row++ {
						check(field, fv.Get(p, lvl, row).NamedRange)
					}
				}
			}
		case TableField, TableFactory:
			// Excel-Tabellenobjekte, keine Defined Names.
		default:
			t.Fatalf("unbekannter Registry-Feldtyp %s: %T", field, fv)
		}
	}

	sort.Strings(missing)
	sort.Strings(uninit)
	if len(missing) > 0 || len(uninit) > 0 {
		t.Errorf("Binding-Audit fehlgeschlagen: %d missing, %d uninit\n--- uninit ---\n%s\n--- missing ---\n%s",
			len(missing), len(uninit), strings.Join(uninit, "\n"), strings.Join(missing, "\n"))
	}
}
