package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"log"
	"os"
	"path/filepath"
)

func main() {
	// Start ETL process
	startETL()
}

func startETL() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	mongoURI := os.Getenv("MONGO_URI")
	if mongoURI == "" {
		log.Fatalln("No URI has been set.")
	}

	clientOptions := options.Client().ApplyURI(mongoURI)
	client, err := mongo.Connect(context.TODO(), clientOptions)
	if err != nil {
		log.Panicf("error connecting to MongoDB: %v", err)
	}
	defer client.Disconnect(context.TODO())

	rawCompanies, _ := FetchDocumentsFromMongo(client)

	companyToWebsiteMap := make(map[string]string)

	for _, data := range rawCompanies {
		company := data.Company // assuming Company is a field in the CompanyDataOld struct
		prompt := fmt.Sprintf("Give me the website for this company, return only a string and no humanlike response, the company is %s", company)
		// I use a European model to increase the chance of matching with European websites but you can also specify this in the prompt
		response := startModelReturnResponse("mistral-nemo:latest", prompt)
		companyToWebsiteMap[company] = response
		log.Printf("Company: %s Response: %s", company, response)
	}

	var enrichedCompanyData []CompanyDataEnriched
	for _, item := range enrichedCompanyData {
		website := companyToWebsiteMap[item.Company]
		if website == "" {
			// An extra check if the LLM does not add https://
			website = "https://" + item.Company
		}

		enrichedData := CompanyDataEnriched{
			ID:      item.ID,
			Company: item.Company,
			Website: website,
		}

		enrichedCompanyData = append(enrichedCompanyData, enrichedData)
	}

	dbErr := UploadToMongo(enrichedCompanyData)
	if dbErr != nil {
		log.Printf("Error uploading coupons to MongoDB: %v", dbErr)
	}

	if wErr := writeJSONFile("output.json", enrichedCompanyData); wErr != nil {
		log.Printf("Error writing JSON file: %v", err)
	}

	log.Println("ETL process finished.")
}

func writeJSONFile(filename string, data []CompanyDataEnriched) error {
	file, err := os.Create(filepath.Clean(filename))
	if err != nil {
		return fmt.Errorf("error creating file: %v", err)
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(data); err != nil {
		return fmt.Errorf("error encoding JSON: %v", err)
	}
	return nil
}
