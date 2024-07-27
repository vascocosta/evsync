package main

import (
	"fmt"
	"strings"
)

func f1Formatter(source EventSource, summary string, dateTime string) string {
	if strings.Contains(strings.ToLower(summary), "quali") {
		source.Tags += " quali qualy f1quali f1qualy qualifying f1qualifying"
	}

	if strings.Contains(strings.ToLower(summary), "race") {
		source.Tags += " race f1race"
	}

	if strings.Contains(summary, " - ") {
		splitSummary := strings.Split(summary, " - ")

		return fmt.Sprintf(
			"[Formula 1],%v,%v,%v,%v,%v,%v",
			strings.ReplaceAll(
				strings.ReplaceAll(
					splitSummary[0],
					" FORMULA 1 ",
					"",
				),
				"GRAND PRIX 2024",
				"GP",
			),
			splitSummary[1],
			dateTime,
			source.Channel,
			source.Tags,
			source.Notify,
		)
	} else {
		return fmt.Sprintf(
			"[Formula 1],%v,NA,%v,%v,%v,%v",
			summary,
			dateTime,
			source.Channel,
			source.Tags,
			source.Notify,
		)
	}
}

func f2Formatter(source EventSource, summary string, dateTime string) string {
	if strings.Contains(summary, " - ") {
		splitSummary := strings.Split(summary, " - ")

		return fmt.Sprintf(
			"[Formula 2],%v,%v,%v,%v,%v,%v",
			strings.ReplaceAll(splitSummary[0], " FIA FORMULA 2: The Championship ", ""),
			splitSummary[1],
			dateTime,
			source.Channel,
			source.Tags,
			source.Notify,
		)
	} else {
		return fmt.Sprintf(
			"[Formula 2],%v,NA,%v,%v,%v,%v",
			summary,
			dateTime,
			source.Channel,
			source.Tags,
			source.Notify,
		)
	}
}

func f3Formatter(source EventSource, summary string, dateTime string) string {
	if strings.Contains(summary, " - ") {
		splitSummary := strings.Split(summary, " - ")

		return fmt.Sprintf(
			"[Formula 3],%v,%v,%v,%v,%v,%v",
			strings.ReplaceAll(splitSummary[0], " FIA FORMULA 3: The Championship ", ""),
			splitSummary[1],
			dateTime,
			source.Channel,
			source.Tags,
			source.Notify,
		)
	} else {
		return fmt.Sprintf(
			"[Formula 3],%v,NA,%v,%v,%v,%v",
			summary,
			dateTime,
			source.Channel,
			source.Tags,
			source.Notify,
		)
	}
}

func indyCarFormatter(source EventSource, summary string, dateTime string) string {
	return fmt.Sprintf(
		"[IndyCar],%v,Race,%v,%v,%v,%v",
		summary[5:],
		dateTime,
		source.Channel,
		source.Tags,
		source.Notify,
	)
}

func motoGPFormatter(source EventSource, summary string, dateTime string) string {
	splitSummary := strings.Split(summary, "(")

	return fmt.Sprintf(
		"[MotoGP],%v,%v,%v,%v,%v,%v",
		strings.ReplaceAll(splitSummary[1], ")", ""),
		strings.TrimSpace(strings.ReplaceAll(splitSummary[0], "MOTOGP: ", "")),
		dateTime,
		source.Channel,
		source.Tags,
		source.Notify,
	)
}

func spaceFormatter(source EventSource, summary string, dateTime string) string {
	if strings.Contains(strings.ToLower(summary), "falcon") {
		source.Tags += " spacex"
	}

	return fmt.Sprintf(
		"[Space],%v,Launch,%v,%v,%v,%v",
		summary,
		dateTime,
		source.Channel,
		source.Tags,
		source.Notify,
	)
}

func defaultFormatter(source EventSource, summary string, dateTime string) string {
	return fmt.Sprintf(
		"[%v],%v, ,%v,%v,%v,%v",
		source.Name,
		summary,
		dateTime,
		source.Channel,
		source.Tags,
		source.Notify,
	)
}
