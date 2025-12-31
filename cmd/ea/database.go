package main

const (
	createTableQuery = `CREATE TABLE input (
    id INTEGER PRIMARY KEY,
    name TEXT NOT NULL,
    surname TEXT NOT NULL,
    subject TEXT NOT NULL,
    paper TEXT,
    level TEXT NOT NULL,
    date TEXT NOT NULL,
    location TEXT NOT NULL,
    start INTEGER NOT NULL,
    finish INTEGER NOT NULL,
    extra_time TEXT,
    notes TEXT)`

	insertQuery = `INSERT INTO input (name,surname,subject,paper,level,date,location,start,finish,extra_time,notes) 
    VALUES (?,?,?,?,?,?,?,?,?,?,?)`

	// pupil view

	selectPupilNamesQuery = `SELECT DISTINCT name,surname 
    FROM input
    ORDER BY surname ASC`

	selectPupilExamsQuery = `SELECT DISTINCT subject,paper,level,date,location,start,finish 
    FROM input 
    WHERE name=? AND surname=?`

	// invigilator view

	selectLocationsByDateQuery = `SELECT DISTINCT date,location 
    FROM input`

	selectPupilsByDateAndLocationQuery = `SELECT name,surname,subject,paper,level,start,finish,extra_time,notes 
    FROM input 
    WHERE date=? AND location=? AND start>=? AND start<?
    ORDER BY subject ASC, finish ASC, surname ASC `

	// register view

	selectAllDatesQuery = `SELECT DISTINCT date 
    FROM input`

	selectPupilsForDateAndTime = `SELECT DISTINCT name,surname,subject,location 
    FROM input 
    WHERE date=? AND start>=? AND start<?
    ORDER BY surname ASC`

	// checks

	pupilsHaveBothPartsCheckQuery = `SELECT name,surname,subject,date, COUNT(*)
    FROM input
    WHERE paper IN ("P1","P2")
    GROUP BY name,surname,subject,date
    `

	timeBetweenPapersCheckQuery = `SELECT name,surname,subject,paper,date,start,finish 
    FROM input
    WHERE paper IN ("P1","P2")
    ORDER BY surname ASC, subject ASC, paper ASC, date ASC, start ASC`

	// room view

	selectAllLocationsQuery = `SELECT DISTINCT location
    FROM input
    WHERE location != 'TBC'
    ORDER BY location ASC`

	selectPupilsForLocationQuery = `SELECT COUNT(*)
    FROM input
    WHERE location=? AND date=? AND start>=? AND start<?
    GROUP BY name,surname`
)
