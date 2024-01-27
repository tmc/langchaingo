package detectschemaprompts

func GetExamples() []map[string]string {
	return []map[string]string{
		{
			"i":         "1",
			"file_type": "CSV",
			"sample_data": `Index, Name, Age, City
			1, John Doe, 35, New York
			2, Jane Smith, 28, Los Angeles`,
			"response": `[
				{
					"name": "Index",
					"description": "Unique identifier",
					"type": "integer"
				},
				{
					"name": "Name",
					"description": "Full name of the individual",
					"type": "string"
				},
				{
					"name": "Age",
					"description": "Age of the individual",
					"type": "integer"
				},
				{
					"name": "City",
					"description": "City where the individual resides",
					"type": "string"
				}
			]`,
		},
		{
			"i":         "2",
			"file_type": "Excel",
			"sample_data": `Order ID, Product, Quantity, Price
			001, Apples, 10, 1.5
			002, Bananas, 20, 0.75`,
			"response": `[
				{
					"name": "Order ID",
					"description": "Unique identifier for the order",
					"type": "string"
				},
				{
					"name": "Product",
					"description": "Name of the product",
					"type": "string"
				},
				{
					"name": "Quantity",
					"description": "Quantity of the product ordered",
					"type": "integer"
				},
				{
					"name": "Price",
					"description": "Price per unit of the product",
					"type": "float"
				}
			]`,
		},
	}
}
