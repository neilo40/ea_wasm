//go:build js && wasm

package main

import (
	"fmt"
	"strings"
)

func (a *Arrangements) AllPupilsHaveBothPapers() string {
	rows, err := a.db.Query(pupilsHaveBothPartsCheckQuery)
	if err != nil {
		return err.Error()
	}
	defer rows.Close()

	errs := make([]string, 0)
	errs = append(errs, "<h2>Checking all pupils have both parts...</h2>")
	for rows.Next() {
		var name string
		var surname string
		var subject string
		var date string
		var count int
		rows.Scan(&name, &surname, &subject, &date, &count)
		if count < 2 {
			errs = append(errs, fmt.Sprintf("%s %s sitting %s on %s is missing a paper", name, surname, subject, date))
		}
	}

	if len(errs) == 1 {
		errs = append(errs, "OK")
	}
	return strings.Join(errs, "<br />")
}

type Sitting struct {
	Name       string
	Surname    string
	Subject    string
	Paper      string
	Date       string
	StartMins  int
	FinishMins int
}

func (a *Arrangements) CheckExamGaps() string {
	rows, err := a.db.Query(timeBetweenPapersCheckQuery)
	if err != nil {
		return err.Error()
	}
	defer rows.Close()

	sittings := make([]Sitting, 0)
	for rows.Next() {
		var name string
		var surname string
		var subject string
		var paper string
		var date string
		var start int
		var finish int
		rows.Scan(&name, &surname, &subject, &paper, &date, &start, &finish)
		sittings = append(sittings, Sitting{name, surname, subject, paper, date, start, finish})
	}

	errs := make([]string, 0)
	errs = append(errs, "<h2>Checking gap between papers...</h2>")
	for i := 0; i < len(sittings)-1; i++ {
		gap := sittings[i+1].StartMins - sittings[i].FinishMins
		if sittings[i].Paper == "P1" {
			// need 30 min gap between P1 and P2 of the same subject
			if gap < 30 {
				errs = append(errs, fmt.Sprintf("%s %s sitting %s on %s only has %d mins between papers",
					sittings[i].Name, sittings[i].Surname, sittings[i].Subject, sittings[i].Date, gap))
			}
		} else if sittings[i].Date == sittings[i+1].Date && sittings[i].Subject != sittings[i+1].Subject {
			// otherwise need 60 min gap between different subjects on the same day.
			if gap < 60 {
				errs = append(errs, fmt.Sprintf("%s %s sitting %s %s followed by %s %s on %s only has %d mins between subjects",
					sittings[i].Name, sittings[i].Surname, sittings[i].Paper, sittings[i].Subject, sittings[i+1].Paper, sittings[i+1].Subject, sittings[i].Date, gap))
			}
		}
	}

	if len(errs) == 1 {
		errs = append(errs, "OK")
	}

	return strings.Join(errs, "<br />")
}

func (a *Arrangements) CheckPupilsAreNotInTwoPlacesAtOnce() string {

	rows, err := a.db.Query(selectAllDatesQuery)
	if err != nil {
		return err.Error()
	}

	dates := make([]string, 0)
	for rows.Next() {
		var date string
		rows.Scan(&date)
		dates = append(dates, date)
	}
	rows.Close()

	errs := make([]string, 0)
	errs = append(errs, "<h2>Checking for pupils that need to be in two places at once...</h2>")
	for _, d := range dates {
		rows, err = a.db.Query(pupilsForDateQuery, d)
		if err != nil {
			return err.Error()
		}

		pupils := make([]Pupil, 0)
		for rows.Next() {
			var name string
			var surname string
			rows.Scan(&name, &surname)
			pupils = append(pupils, Pupil{Name: name, Surname: surname})
		}
		rows.Close()

		for _, p := range pupils {
			rows, err = a.db.Query(examsForPupilAndDateCheckQuery, d, p.Name, p.Surname)
			if err != nil {
				return err.Error()
			}

			exams := make([]Sitting, 0)
			var subject string
			var paper string
			var start int
			var finish int
			for rows.Next() {
				rows.Scan(&subject, &paper, &start, &finish)
				exams = append(exams, Sitting{Subject: subject, Paper: paper, StartMins: start, FinishMins: finish})
			}
			rows.Close()

			for i := 0; i < len(exams)-1; i++ {
				if exams[i].FinishMins > exams[i+1].StartMins {
					errs = append(errs, fmt.Sprintf("%s %s needs to be in two places at once on %s", p.Name, p.Surname, d))
				}
			}
		}

	}

	if len(errs) == 1 {
		errs = append(errs, "OK")
	}

	return strings.Join(errs, "<br />")
}
