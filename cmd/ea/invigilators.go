//go:build js && wasm

package main

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/xuri/excelize/v2"
)

type LocDate struct {
	Location string
	Date     string
}

type InvigilatorRow struct {
	Name    string
	Surname string
	Subject string
	Paper   string
	Level   string
	Start   string
	Finish  string
	Extime  string
	Notes   string
}

func (a *Arrangements) GenerateInvigilatorsView() error {
	rows, err := a.db.Query(selectLocationsByDateQuery)
	if err != nil {
		return err
	}
	locdates := make([]LocDate, 0)
	for rows.Next() {
		var date string
		var location string
		rows.Scan(&date, &location)
		locdates = append(locdates, LocDate{Location: location, Date: date})
	}
	rows.Close()

	morningStart := 8 * 60
	morningEnd := 12 * 60
	afternoonEnd := 17 * 60

	f := excelize.NewFile()
	defer f.Close()

	setupStyles(f)

	for _, ld := range locdates {
		for _, s := range []Session{
			{start: morningStart, end: morningEnd, name: "morning"},
			{start: morningEnd, end: afternoonEnd, name: "afternoon"},
		} {
			irows := make([]InvigilatorRow, 0)
			pupilRows, err := a.db.Query(selectPupilsByDateAndLocationQuery, ld.Date, ld.Location, s.start, s.end)
			if err != nil {
				return err
			}

			for pupilRows.Next() {
				var name string
				var surname string
				var subject string
				var paper string
				var level string
				var start int
				var finish int
				var extraTime string
				var notes string
				pupilRows.Scan(&name, &surname, &subject, &paper, &level, &start, &finish, &extraTime, &notes)
				irows = append(irows, InvigilatorRow{Name: name, Surname: surname, Subject: subject, Paper: paper,
					Level: level, Start: minsToTime(start), Finish: minsToTime(finish), Extime: extraTime, Notes: notes})
			}
			pupilRows.Close()

			if len(irows) == 0 {
				continue // no exams on this date for this session
			}

			date, err := time.Parse("01-02-06", ld.Date)
			if err != nil {
				log.Println(ld.Date)
				return err
			}
			dateStr := date.Format("02/Jan/2006")

			// create worksheets

			sheetName := fmt.Sprintf("%s %s %s", ld.Location, strings.ReplaceAll(dateStr, "/", "_"), s.name)
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
			f.SetCellStr(sheetName, "A1", fmt.Sprintf("%s (Room %s, %s)", dateStr, ld.Location, s.name))
			f.MergeCell(sheetName, "A1", "I1")
			f.SetRowHeight(sheetName, 1, 40)

			// table header

			f.SetCellStyle(sheetName, "A3", "I3", tableHeaderStyle)
			f.SetSheetRow(sheetName, "A3", &[]any{"Name", "Surname", "Subject", "Paper", "Level", "Start", "Finish",
				"Ex Time", "Additional Arrangements"})

			// table content

			maxNameWidth := 5
			maxSurnameWidth := 8
			maxSubjectWidth := 8
			maxLevelWidth := 5
			maxExTimeWidth := 6
			maxNotesWidth := 30
			rowNum := 4
			for _, r := range irows {

				if len(r.Name) > maxNameWidth {
					maxNameWidth = len(r.Name)
				}
				if len(r.Surname) > maxSurnameWidth {
					maxSurnameWidth = len(r.Surname)
				}
				if len(r.Subject) > maxSubjectWidth {
					maxSubjectWidth = len(r.Subject)
				}
				if len(r.Level) > maxLevelWidth {
					maxLevelWidth = len(r.Level)
				}
				if len(r.Extime) > maxExTimeWidth {
					maxExTimeWidth = len(r.Extime)
				}
				f.SetSheetRow(sheetName, fmt.Sprintf("A%d", rowNum), &[]any{r.Name, r.Surname, r.Subject, r.Paper, r.Level,
					r.Start, r.Finish, r.Extime, r.Notes})
				rowNum++
			}

			f.SetCellStyle(sheetName, "A4", fmt.Sprintf("I%d", rowNum-1), tableContentStyle)
			// no autowidth func, need to set manually
			f.SetColWidth(sheetName, "A", "A", float64(maxNameWidth)+2)
			f.SetColWidth(sheetName, "B", "B", float64(maxSurnameWidth)+2)
			f.SetColWidth(sheetName, "C", "C", float64(maxSubjectWidth)+2)
			f.SetColWidth(sheetName, "E", "E", float64(maxLevelWidth)+2)
			f.SetColWidth(sheetName, "H", "H", float64(maxExTimeWidth)+2)
			f.SetColWidth(sheetName, "I", "I", float64(maxNotesWidth)+2)
			f.SetCellStyle(sheetName, "I4", fmt.Sprintf("I%d", rowNum-1), additionalArrStyle)

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
	a.InvigilatorsSheet = buf.Bytes()

	return nil
}
