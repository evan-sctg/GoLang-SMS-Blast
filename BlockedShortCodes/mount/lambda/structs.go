package main

type Bucket struct {
	Uuid   	 string   `json:"uuid"`
}

type Message struct {
	Broadcast 	 string   `json:"uuid"`
	Message   	 string   `json:"message"`
	PhoneNumbers []string `json:"numbers"`
}

type Number struct {
	PhoneNumber   string `json:"phone_number"`
}

