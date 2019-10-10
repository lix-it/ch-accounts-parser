package main

import "reflect"

func calculateDataQuality(input []AccountsFilingEntry) float32 {
	var totalScore float32
	for _, entry := range input {
		// calcualate number of rows with non-null data
		var currentQualScore float32
		t := reflect.ValueOf(entry)
		for index := 0; index < t.NumField(); index++ {
			// Assume all fields are strings
			str := t.Field(index).Interface()
			if str != "" {
				currentQualScore++
			}
		}
		// divide by total number of fields
		currentQualScore = currentQualScore / float32(t.NumField())
		totalScore += currentQualScore
	}
	// take average
	totalScore /= float32(len(input))
	return totalScore * 100 // return %age
}
