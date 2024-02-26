package slo

//go:generate ../../../bin/go-enum --nocase --lower --names --values

// WhenDelayed represents enum for behavior of Composite SLO objectives
/* ENUM(
CountAsGood
CountAsBad
Ignore
)*/
type WhenDelayed string
