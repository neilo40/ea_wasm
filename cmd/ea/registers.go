//go:build js && wasm

package main

import (
	"fmt"
	"strings"
	"time"

	"github.com/xuri/excelize/v2"
)

type Session struct {
	start int
	end   int
	name  string
}

type RegisterRow struct {
	Name     string
	Surname  string
	Subject  string
	Location string
}

func (a *Arrangements) GenerateRegistersView() error {
	rows, err := a.db.Query(selectAllDatesQuery)
	if err != nil {
		return err
	}

	morningStart := 8 * 60
	morningEnd := 12 * 60
	afternoonEnd := 17 * 60
	dates := make([]string, 0)
	for rows.Next() {
		var date string
		rows.Scan(&date)
		dates = append(dates, date)
	}
	rows.Close()

	f := excelize.NewFile()
	defer f.Close()

	setupStyles(f)

	for _, date := range dates {
		for _, s := range []Session{
			{start: morningStart, end: morningEnd, name: "morning"},
			{start: morningEnd, end: afternoonEnd, name: "afternoon"},
		} {
			rrows := make([]RegisterRow, 0)
			pupilRows, err := a.db.Query(selectPupilsForDateAndTime, date, s.start, s.end)
			if err != nil {
				return err
			}

			for pupilRows.Next() {
				var name string
				var surname string
				var subject string
				var location string
				pupilRows.Scan(&name, &surname, &subject, &location)
				rrows = append(rrows, RegisterRow{Name: name, Surname: surname, Subject: subject, Location: location})
			}
			pupilRows.Close()

			if len(rrows) == 0 {
				continue // no exams on this date for this session
			}

			// create worksheets

			sheetName := fmt.Sprintf("%s %s", strings.ReplaceAll(date, "/", "_"), s.name)
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

			// title

			f.SetCellStyle(sheetName, "A1", "A1", nameStyle)
			f.SetCellStr(sheetName, "A1", fmt.Sprintf("%s (%s)", date, s.name))
			f.MergeCell(sheetName, "A1", "G1")
			f.SetRowHeight(sheetName, 1, 40)

			// table header

			f.SetCellStyle(sheetName, "A3", "G3", tableHeaderStyle)
			f.SetSheetRow(sheetName, "A3", &[]any{"Name", "Surname", "Subject", "Location", "Laptop #", "Seat #", "Present"})

			// table content

			maxNameWidth := 5
			maxSurnameWidth := 8
			maxSubjectWidth := 8
			rowNum := 4
			for _, r := range rrows {

				if len(r.Name) > maxNameWidth {
					maxNameWidth = len(r.Name)
				}
				if len(r.Surname) > maxSurnameWidth {
					maxSurnameWidth = len(r.Surname)
				}
				if len(r.Subject) > maxSubjectWidth {
					maxSubjectWidth = len(r.Subject)
				}
				f.SetSheetRow(sheetName, fmt.Sprintf("A%d", rowNum), &[]any{r.Name, r.Surname, r.Subject, r.Location})
				rowNum++
			}

			f.SetCellStyle(sheetName, "A4", fmt.Sprintf("G%d", rowNum-1), tableContentStyle)
			// no autowidth func, need to set manually
			f.SetColWidth(sheetName, "A", "A", float64(maxNameWidth)+2)
			f.SetColWidth(sheetName, "B", "B", float64(maxSurnameWidth)+2)
			f.SetColWidth(sheetName, "C", "C", float64(maxSubjectWidth)+2)

			// footer

			footerCell := fmt.Sprintf("A%d", rowNum+1)
			f.SetCellStr(sheetName, footerCell, fmt.Sprintf("Table Generated at %s", time.Now().Format(time.RFC3339)))
			f.SetCellStyle(sheetName, footerCell, footerCell, footerStyle)

		}
	}

	f.SetActiveSheet(0)
	f.DeleteSheet("Sheet1") // delete default sheet

	// save to bytes

	buf, err := f.WriteToBuffer()
	if err != nil {
		return err
	}
	a.RegistersSheet = buf.Bytes()

	return nil
}
