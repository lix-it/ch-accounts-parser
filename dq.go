package main

import "reflect"

func calculateDataQuality(input []AccountsFilingEntry) float32 {
	var totalScore float32
	for _, entry := range input {
		// calcualate number of rows with non-null data
		var currentQualScore float32
		val := reflect.ValueOf(entry)
		t := val.Type()
		numFields := val.NumField()
		for index := 0; index < numFields; index++ {
			// ignore fields that aren't exported
			field := val.Field(index)
			// get struct tag from the type's field
			if t.Field(index).Tag.Get("csv") == "-" {
				continue
			}
			// Assume all fields are strings
			str := field.Interface()
			if str != "" {
				currentQualScore++
			}
		}
		// divide by total number of fields
		currentQualScore = currentQualScore / float32(numFields)
		totalScore += currentQualScore
	}
	// take average
	totalScore /= float32(len(input))
	return totalScore * 100 // return %age
}
