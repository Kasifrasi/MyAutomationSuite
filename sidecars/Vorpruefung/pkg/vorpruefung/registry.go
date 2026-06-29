package vorpruefung

import "fmt"

type ValidationList []string

// Zentral definierte Dropdown-Werte, damit die API sie nutzen kann
var (
	ListJaNein           ValidationList = []string{"Ja", "Nein"}
	ListAbzug            ValidationList = []string{"Abzug", "Kein Abzug"}
	ListWaehrung         ValidationList = []string{"EUR", "USD", "CHF"}
	ListMonate           ValidationList = []string{"0", "1", "2", "3", "4", "5", "6", "7", "8", "9", "10", "11", "12"}
	ListKostenkategorien ValidationList = []string{"Bauausgaben", "Investitionen", "Personalkosten", "Projektaktivitaeten", "Projektverwaltung", "Evaluierung", "Audit", "Reserve"}
)

type InputField struct {
	NamedRange string
	Validation ValidationList
}

// Registry aller skalaren gelben Felder
var (
	// Dashboard
	FieldDashProjektnummer       = InputField{NamedRange: "Inp_Dash_Projektnummer"}
	FieldDashVorprojekt          = InputField{NamedRange: "Inp_Dash_Vorprojekt", Validation: ListJaNein}
	FieldDashProjekttitel        = InputField{NamedRange: "Inp_Dash_Projekttitel"}
	FieldDashProjekttraeger      = InputField{NamedRange: "Inp_Dash_Projekttraeger"}
	FieldDashBerichtswaehrung    = InputField{NamedRange: "Inp_Dash_Berichtswaehrung"}
	FieldDashProjektstart        = InputField{NamedRange: "Inp_Dash_Projektstart"}
	FieldDashProjektende         = InputField{NamedRange: "Inp_Dash_Projektende"}
	FieldDashVPNummer            = InputField{NamedRange: "Inp_Dash_VPNummer"}
	FieldDashVPBerichtswaehrung  = InputField{NamedRange: "Inp_Dash_VPBerichtswaehrung"}
	FieldDashVPEnde              = InputField{NamedRange: "Inp_Dash_VPEnde"}
	FieldDashVPWechselkurs       = InputField{NamedRange: "Inp_Dash_VPWechselkurs"}
	FieldDashVPSaldoLC           = InputField{NamedRange: "Inp_Dash_VPSaldoLC"}
	FieldDashVPSaldoEUR          = InputField{NamedRange: "Inp_Dash_VPSaldoEUR"}
	FieldDashVPFolgeprojektstart = InputField{NamedRange: "Inp_Dash_VPFolgeprojektstart"}
	FieldDashVPFolgeWechselkurs  = InputField{NamedRange: "Inp_Dash_VPFolgeWechselkurs"}
	FieldDashVPFolgeSaldoLC      = InputField{NamedRange: "Inp_Dash_VPFolgeSaldoLC"}
	FieldDashVPFolgeSaldoEUR     = InputField{NamedRange: "Inp_Dash_VPFolgeSaldoEUR"}

	// Budget
	FieldBudgetReserveFreigabe = InputField{NamedRange: "Inp_Budget_ReserveFreigabe", Validation: ListJaNein}
	FieldBudgetDrittmittelY1   = InputField{NamedRange: "Inp_Budget_DrittmittelY1"}
	FieldBudgetDrittmittelY2   = InputField{NamedRange: "Inp_Budget_DrittmittelY2"}
	FieldBudgetDrittmittelY3   = InputField{NamedRange: "Inp_Budget_DrittmittelY3"}

	FieldBudgetEigenmittelLC  = InputField{NamedRange: "Inp_Budget_EigenmittelLC"}
	FieldBudgetEigenmittelY1  = InputField{NamedRange: "Inp_Budget_EigenmittelY1"}
	FieldBudgetEigenmittelY2  = InputField{NamedRange: "Inp_Budget_EigenmittelY2"}
	FieldBudgetEigenmittelY3  = InputField{NamedRange: "Inp_Budget_EigenmittelY3"}
	FieldBudgetEigenmittelEUR = InputField{NamedRange: "Inp_Budget_EigenmittelEUR"}

	FieldBudgetKMWLC  = InputField{NamedRange: "Inp_Budget_KMWLC"}
	FieldBudgetKMWY1  = InputField{NamedRange: "Inp_Budget_KMWY1"}
	FieldBudgetKMWY2  = InputField{NamedRange: "Inp_Budget_KMWY2"}
	FieldBudgetKMWY3  = InputField{NamedRange: "Inp_Budget_KMWY3"}
	FieldBudgetKMWEUR = InputField{NamedRange: "Inp_Budget_KMWEUR"}

	// Pruefung FB
	FieldFBPruefungAuswahl    = InputField{NamedRange: "Inp_FBPruefung_Auswahl"}
	FieldFBPruefungAbzugSaldo = InputField{NamedRange: "Inp_FBPruefung_AbzugSaldo", Validation: ListAbzug}
	FieldFBPruefungAbzugMehr  = InputField{NamedRange: "Inp_FBPruefung_AbzugMehr", Validation: ListAbzug}

	// Pruefung MA
	FieldMAPruefungAuswahl       = InputField{NamedRange: "Inp_MAPruefung_Auswahl"}
	FieldMAPruefungAbzugSaldo    = InputField{NamedRange: "Inp_MAPruefung_AbzugSaldo", Validation: ListAbzug}
	FieldMAPruefungAbzugMehr     = InputField{NamedRange: "Inp_MAPruefung_AbzugMehr", Validation: ListAbzug}
	FieldMAPruefungAbzugPrognose = InputField{NamedRange: "Inp_MAPruefung_AbzugPrognose", Validation: ListAbzug}
	FieldMAPruefungMonateY1      = InputField{NamedRange: "Inp_MAPruefung_MonateY1", Validation: ListMonate}
	FieldMAPruefungMonateY2      = InputField{NamedRange: "Inp_MAPruefung_MonateY2", Validation: ListMonate}
	FieldMAPruefungMonateY3      = InputField{NamedRange: "Inp_MAPruefung_MonateY3", Validation: ListMonate}
)

func FieldFBVon(period int) InputField {
	return InputField{NamedRange: fmt.Sprintf("Inp_FB_Von_%d", period)}
}

func FieldFBBis(period int) InputField {
	return InputField{NamedRange: fmt.Sprintf("Inp_FB_Bis_%d", period)}
}

func FieldFBAufschlBank(period int) InputField {
	return InputField{NamedRange: fmt.Sprintf("Inp_FB_aufschl_Bank_%d", period)}
}

func FieldFBAufschlKasse(period int) InputField {
	return InputField{NamedRange: fmt.Sprintf("Inp_FB_aufschl_Kasse_%d", period)}
}

func FieldFBAufschlSonstiges(period int) InputField {
	return InputField{NamedRange: fmt.Sprintf("Inp_FB_aufschl_Sonstiges_%d", period)}
}

func FieldMAVon(tableId int) InputField {
	return InputField{NamedRange: fmt.Sprintf("Inp_MA_Von_%d", tableId)}
}

func FieldMABis(tableId int) InputField {
	return InputField{NamedRange: fmt.Sprintf("Inp_MA_Bis_%d", tableId)}
}

func FieldMAKurs(tableId int) InputField {
	return InputField{NamedRange: fmt.Sprintf("Inp_MA_Kurs_%d", tableId)}
}

func FieldMAManBetrag(tableId int) InputField {
	return InputField{NamedRange: fmt.Sprintf("Inp_MA_ManBetrag_%d", tableId)}
}

func FieldMAEigenmittelLC(tableId int) InputField {
	return InputField{NamedRange: fmt.Sprintf("Inp_MA_EigenmittelLC_%d", tableId)}
}

func FieldMADrittmittelLC(tableId int) InputField {
	return InputField{NamedRange: fmt.Sprintf("Inp_MA_DrittmittelLC_%d", tableId)}
}

func FieldMASaldoLC(tableId int) InputField {
	return InputField{NamedRange: fmt.Sprintf("Inp_MA_SaldoLC_%d", tableId)}
}

func FieldDashChecklist(index int) InputField {
	return InputField{NamedRange: fmt.Sprintf("Inp_Dash_Checklist_%d", index), Validation: ListJaNein}
}
