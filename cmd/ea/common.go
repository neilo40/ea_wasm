package main

import "github.com/xuri/excelize/v2"

var (
	nameStyle         int
	tableHeaderStyle  int
	tableContentStyle int
	instructionsStyle int
	footerStyle       int
	dateStyle         int
	err               error
	pageSize          = 9 // A4
	pageOrientation   = "landscape"
)

func setupStyles(f *excelize.File) error {
	nameStyle, err = f.NewStyle(&excelize.Style{
		Font: &excelize.Font{
			Bold: true,
			Size: 30,
		},
	})
	if err != nil {
		return err
	}

	tableHeaderStyle, err = f.NewStyle(&excelize.Style{
		Font: &excelize.Font{
			Bold: true,
		},
		Border: []excelize.Border{
			{Type: "left", Style: 1, Color: "000000"},
			{Type: "right", Style: 1, Color: "000000"},
			{Type: "top", Style: 1, Color: "000000"},
			{Type: "bottom", Style: 1, Color: "000000"},
		},
	})
	if err != nil {
		return err
	}

	tableContentStyle, err = f.NewStyle(&excelize.Style{
		Border: []excelize.Border{
			{Type: "left", Style: 1, Color: "000000"},
			{Type: "right", Style: 1, Color: "000000"},
			{Type: "top", Style: 1, Color: "000000"},
			{Type: "bottom", Style: 1, Color: "000000"},
		},
	})
	if err != nil {
		return err
	}

	instructionsStyle, err = f.NewStyle(&excelize.Style{
		Alignment: &excelize.Alignment{
			WrapText: true,
		},
	})
	if err != nil {
		return err
	}

	footerStyle, err = f.NewStyle(&excelize.Style{
		Font: &excelize.Font{
			Size: 8,
		},
	})
	if err != nil {
		return err
	}

	dateStyle, err = f.NewStyle(&excelize.Style{
		NumFmt: 15,
		Border: []excelize.Border{
			{Type: "left", Style: 1, Color: "000000"},
			{Type: "right", Style: 1, Color: "000000"},
			{Type: "top", Style: 1, Color: "000000"},
			{Type: "bottom", Style: 1, Color: "000000"},
		},
	})
	if err != nil {
		return err
	}

	return nil
}
