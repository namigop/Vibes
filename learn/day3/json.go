package main

import (
	"encoding/json"
	"fmt"
)

type Order struct {
	OrderId     int64    `json:"orderId"`
	CustomerId  int64    `json:"customerId"`
	TotalAmount float64  `json:"totalAmount"`
	Customer    Customer `json:"customer"`
}

type Customer struct {
	Id     int64  `json:"id"`
	Name   string `json:"name"`
	Adress string `json:"address"`
	Email  string `json:"email"`
}

func jsonExample() {
	order := Order{
		OrderId:     12345,
		CustomerId:  67890,
		TotalAmount: 100.00,
		Customer: Customer{
			Id:     67890,
			Name:   "John Doe",
			Adress: "123 Main St",
			Email:  "john.doe@example.com",
		},
	}

	jsonStr, err := Serialize(order, true)
	if err != nil {
		panic(err)
	}
	fmt.Println(jsonStr)
}

func Serialize[T any](source T, prettyPrint bool) (string, error) {
	var bytes []byte
	var err error
	if prettyPrint {
		bytes, err = json.MarshalIndent(source, "", "  ")
	} else {
		bytes, err = json.Marshal(source)
	}

	if err != nil {
		return "", err
	}

	return string(bytes), nil
}

func Deserialize[T any](jsonStr string) (T, error) {
	var instance T
	bytes := []byte(jsonStr)
	error := json.Unmarshal(bytes, &instance)
	if error != nil {
		return instance, error
	}

	return instance, nil
}
