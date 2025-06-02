package main

type CompanyDataOld struct {
	ID      string `json:"id"`
	Company string `json:"company"`
}

type CompanyDataEnriched struct {
	ID      string `json:"id"`
	Company string `json:"company"`
	Website string `json:"website"`
}
