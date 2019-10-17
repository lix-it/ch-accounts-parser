package main

import "fmt"

func (c AccountsFilingEntry) String() string {
	var result string
	result = fmt.Sprintf("Name: %v\n", c.Name)
	result = fmt.Sprintf("%vID: %v\n", result, c.RegistrationID)
	result = fmt.Sprintf("%vAddress: %v, %v, %v, %v\n", result, c.AddressLine1, c.AddressLine2, c.CityOrTown, c.PostCode)
	result = fmt.Sprintf("%vApproval Date: %v\n", result, c.ApprovalDate)
	result = fmt.Sprintf("%vDormant: %v\n", result, c.Dormant)
	result = fmt.Sprintf("%vPeriod End Date: %v\n", result, c.PeriodEnd)
	result = fmt.Sprintf("%v\n", result)
	return result
}
