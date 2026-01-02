//go:build js && wasm

package main

import (
	"errors"
	"fmt"
	"strings"
)

func (a *Arrangements) AllPupilsHaveBothPapers() error {
	rows, err := a.db.Query(pupilsHaveBothPartsCheckQuery)
	if err != nil {
		return err
	}

	errs := make([]string, 0)
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

	if len(errs) == 0 {
		return nil
	}

	return errors.New(strings.Join(errs, "<br />"))
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

func (a *Arrangements) CheckP1P2Gap() error {
	rows, err := a.db.Query(timeBetweenPapersCheckQuery)
	if err != nil {
		return err
	}

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
	for i := 0; i < len(sittings); i += 2 {
		gap := sittings[i+1].StartMins - sittings[i].FinishMins
		if gap < 35 {
			errs = append(errs, fmt.Sprintf("Error: %s %s sitting %s %s on %s only has %d mins between papers",
				sittings[i].Name, sittings[i].Surname, sittings[i].Paper, sittings[i].Subject, sittings[i].Date, gap))
		}
	}

	if len(errs) == 0 {
		return nil
	}

	return errors.New(strings.Join(errs, "<br />"))
}
