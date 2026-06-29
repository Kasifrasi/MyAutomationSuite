package api

// Dashboard Layout
const (
	CellDashProjektnummer       = "C5"
	CellDashVorprojekt          = "E5"
	CellDashProjekttitel        = "C6"
	CellDashProjekttraeger      = "C7"
	CellDashBerichtswaehrung    = "E7"
	CellDashProjektstart        = "C8"
	CellDashProjektende         = "E8"
	CellDashVPNummer            = "C10"
	CellDashVPBerichtswaehrung  = "E10"
	CellDashVPEnde              = "C11"
	CellDashVPWechselkurs       = "E11"
	CellDashVPSaldoLC           = "C12"
	CellDashVPSaldoEUR          = "E12"
	CellDashVPFolgeprojektstart = "C13"
	CellDashVPFolgeWechselkurs  = "E13"
	CellDashVPFolgeSaldoLC      = "C14"
	CellDashVPFolgeSaldoEUR     = "E14"

	ColDashChecklist      = 4 // Spalte D
	RowDashChecklistStart = 16
	RowDashChecklistEnd   = 22
)

// Budget Layout
const (
	CellBudgetReserveFreigabe = "K5"
	CellBudgetDrittmittelY1   = "F10"
	CellBudgetDrittmittelY2   = "G10"
	CellBudgetDrittmittelY3   = "H10"

	RowBudgetEigenmittel = 7
	RowBudgetKMWMittel   = 13

	ColBudgetLC  = 5 // E
	ColBudgetY1  = 6 // F
	ColBudgetY2  = 7 // G
	ColBudgetY3  = 8 // H
	ColBudgetEUR = 9 // I

	ColBudgetOffsetID       = 1
	ColBudgetOffsetPosition = 2
	ColBudgetOffsetLC       = 3
	ColBudgetOffsetY1       = 4
	ColBudgetOffsetY2       = 5
	ColBudgetOffsetY3       = 6
	ColBudgetOffsetEUR      = 7
)

// KMW Layout
const (
	RowKMWStart    = 5
	RowKMWEnd      = 22
	ColKMWPeriode  = 2
	ColKMWWaehrung = 3
	ColKMWBetrag   = 4
	ColKMWDatum    = 5
)

// MA Layout (IV. MA)
const (
	ColMAStart          = 3
	ColMAStep           = 4
	RowMAVon            = 5
	RowMABis            = 6
	RowMAKurs           = 8
	RowMAOffsetEigen    = 4
	RowMAOffsetDritt    = 5
	TableMAOffsetEbene2 = 18
	TableMAOffsetEbene3 = 36
)

// FB Layout (III. Finanzberichte)
const (
	ColFBStart        = 2
	ColFBStep         = 7
	RowFBVon          = 5
	RowFBBis          = 6
	RowFBEinnahmen    = 12
	OffsetFBBank      = 6
	OffsetFBKasse     = 7
	OffsetFBSonstiges = 8
)

// Pruefung Layout
const (
	CellFBPruefungAuswahl = "C9"
	CellMAPruefungAuswahl = "C9"
)
