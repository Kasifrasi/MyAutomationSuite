package api

// Dashboard Layout
const (
	ColDashChecklist      = 4 // Spalte D
	RowDashChecklistStart = 16
	RowDashChecklistEnd   = 22
)

// Budget Layout
const (
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
