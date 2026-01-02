//go:build js && wasm

package main

import (
	"fmt"
	"time"

	"github.com/xuri/excelize/v2"
)

func (a *Arrangements) GenerateRoomUtilView() error {
	rows, err := a.db.Query(selectAllDatesQuery)
	if err != nil {
		return err
	}

	dates := make([]string, 0)
	for rows.Next() {
		var date string
		rows.Scan(&date)
		dates = append(dates, date)
	}
	rows.Close()

	rows, err = a.db.Query(selectAllLocationsQuery)
	if err != nil {
		return err
	}

	locs := make([]string, 0)
	for rows.Next() {
		var loc string
		rows.Scan(&loc)
		locs = append(locs, loc)
	}
	rows.Close()

	morningStart := 8 * 60
	morningEnd := 12 * 60
	afternoonEnd := 17 * 60

	roomMap := make(map[string]map[string]int) // session -> room -> count
	sessions := make([]string, 0, len(dates)*2)
	for _, d := range dates {
		for _, s := range []Session{
			{start: morningStart, end: morningEnd, name: "morning"},
			{start: morningEnd, end: afternoonEnd, name: "afternoon"},
		} {
			sessionKey := fmt.Sprintf("%s %s", d, s.name)
			sessions = append(sessions, sessionKey)
			for _, l := range locs {
				pupilRows, err := a.db.Query(selectPupilsForLocationQuery, l, d, s.start, s.end)
				if err != nil {
					return err
				}
				pc := 0
				for pupilRows.Next() {
					pc++
				}
				pupilRows.Close()
				if pc == 0 {
					continue
				}
				_, ok := roomMap[sessionKey]
				if !ok {
					roomMap[sessionKey] = make(map[string]int)
				}
				roomMap[sessionKey][l] = pc
			}
		}
	}

	f := excelize.NewFile()
	defer f.Close()

	setupStyles(f)

	_, err = f.NewSheet("rooms")
	if err != nil {
		return err
	}

	// page size / layout

	f.SetPageLayout("rooms", &excelize.PageLayoutOptions{
		Size:        &pageSize,
		Orientation: &pageOrientation,
	})

	// title
	f.SetCellStyle("rooms", "A1", "A1", nameStyle)
	f.SetCellStr("rooms", "A1", "Room Usage (pupils per session)")
	f.MergeCell("rooms", "A1", "M1")
	f.SetRowHeight("rooms", 1, 40)

	// table header

	tableWidthCol, _ := excelize.ColumnNumberToName(len(locs) + 1)
	f.SetCellStyle("rooms", "A3", fmt.Sprintf("%s3", tableWidthCol), tableHeaderStyle)
	cols := make([]any, 0)
	cols = append(cols, "Date/Session")
	for _, l := range locs {
		cols = append(cols, l)
	}
	f.SetSheetRow("rooms", "A3", &cols)

	// table content
	rowNum := 4
	for _, s := range sessions {
		row := make([]any, 0, len(locs)+1)
		row = append(row, s)
		for _, l := range locs {
			c, ok := roomMap[s][l]
			if ok {
				row = append(row, c)
			} else {
				row = append(row, nil)
			}
		}
		f.SetSheetRow("rooms", fmt.Sprintf("A%d", rowNum), &row)
		rowNum++
	}
	f.SetCellStyle("rooms", "A4", fmt.Sprintf("%s%d", tableWidthCol, rowNum-1), tableContentStyle)
	f.SetColWidth("rooms", "A", "A", 30)

	// footer
	footerCell := fmt.Sprintf("A%d", rowNum+1)
	f.SetCellStr("rooms", footerCell, fmt.Sprintf("Table Generated at %s", time.Now().Format(time.RFC3339)))
	f.SetCellStyle("rooms", footerCell, footerCell, footerStyle)

	f.SetActiveSheet(0)
	f.DeleteSheet("Sheet1") // delete default sheet

	// save to bytes

	buf, err := f.WriteToBuffer()
	if err != nil {
		return err
	}
	a.RoomUtilSheet = buf.Bytes()

	return nil
}
