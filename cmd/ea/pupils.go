//go:build js && wasm

package main

import (
	"fmt"
	"time"

	"github.com/xuri/excelize/v2"
)

type Pupil struct {
	Name    string
	Surname string
}

func (a *Arrangements) GeneratePupilView() error {
	// get all pupil names

	rows, err := a.db.Query(selectPupilNamesQuery)
	if err != nil {
		return err
	}
	defer rows.Close()

	pupils := make([]Pupil, 0)
	for rows.Next() {
		var name string
		var surname string
		rows.Scan(&name, &surname)
		pupils = append(pupils, Pupil{Name: name, Surname: surname})
	}

	// create pupils workbook

	f := excelize.NewFile()
	defer f.Close()

	setupStyles(f)

	// add worksheet for each pupil

	for _, p := range pupils {
		sheetName := fmt.Sprintf("%s %s", p.Name, p.Surname)
		index, err := f.NewSheet(sheetName)
		if err != nil {
			return err
		}

		// page size / layout

		f.SetActiveSheet(index)
		f.SetPageLayout(sheetName, &excelize.PageLayoutOptions{
			Size:        &pageSize,
			Orientation: &pageOrientation,
		})

		// name

		f.SetCellStyle(sheetName, "A1", "A1", nameStyle)
		f.SetCellStr(sheetName, "A1", sheetName)
		f.MergeCell(sheetName, "A1", "G1")
		f.SetRowHeight(sheetName, 1, 40)

		// instructions

		f.MergeCell(sheetName, "A2", "G2")
		f.SetCellStr(sheetName, "A2", "Please meet Mrs O’Connor at the conference room 10 minutes before your prelim is due to start")
		f.SetCellStyle(sheetName, "A2", "A2", instructionsStyle)
		f.SetRowHeight(sheetName, 2, 30)

		// table header
		f.SetCellStyle(sheetName, "A4", "G4", tableHeaderStyle)
		f.SetSheetRow(sheetName, "A4", &[]any{"Subject", "Paper", "Level", "Date", "Location", "Start", "Finish"})

		// exam details

		rowNum := 5
		examRows, err := a.db.Query(selectPupilExamsQuery, p.Name, p.Surname)
		if err != nil {
			return err
		}
		maxSubjectWidth := 7
		maxPaperWidth := 6
		maxLevelWidth := 6
		for examRows.Next() {
			var subject string
			var paper string
			var level string
			var date string
			var location string
			var start int
			var finish int
			examRows.Scan(&subject, &paper, &level, &date, &location, &start, &finish)
			f.SetSheetRow(sheetName, fmt.Sprintf("A%d", rowNum), &[]any{subject, paper, level, date, location, minsToTime(start),
				minsToTime(finish)})
			rowNum++
			if len(subject) > maxSubjectWidth {
				maxSubjectWidth = len(subject)
			}
			if len(paper) > maxPaperWidth {
				maxPaperWidth = len(paper)
			}
			if len(level) > maxLevelWidth {
				maxLevelWidth = len(level)
			}
		}
		f.SetCellStyle(sheetName, "A5", fmt.Sprintf("G%d", rowNum-1), tableContentStyle)
		// no autowidth func, need to set manually
		f.SetColWidth(sheetName, "A", "A", float64(maxSubjectWidth)+2)
		f.SetColWidth(sheetName, "B", "B", float64(maxPaperWidth)+2)
		f.SetColWidth(sheetName, "C", "C", float64(maxLevelWidth)+2)
		f.SetColWidth(sheetName, "D", "D", 11) // date
		f.SetCellStyle(sheetName, "D5", fmt.Sprintf("D%d", rowNum-1), dateStyle)

		// footer

		footerCell := fmt.Sprintf("A%d", rowNum+1)
		f.SetCellStr(sheetName, footerCell, fmt.Sprintf("Table Generated at %s", time.Now().Format(time.RFC3339)))
		f.SetCellStyle(sheetName, footerCell, footerCell, footerStyle)
	}

	f.SetActiveSheet(0)
	f.DeleteSheet("Sheet1") // delete default sheet

	// save to bytes

	buf, err := f.WriteToBuffer()
	if err != nil {
		return err
	}
	a.PupilsSheet = buf.Bytes()

	return nil
}
