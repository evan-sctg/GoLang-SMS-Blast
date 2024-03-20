package main

type PhoneNumbers struct {
	Numbers []string `json:"numbers"`
	Uuid    string   `json:"uuid"`
	Message string   `json:"message"`
}

type BroadcastUUID struct {
	Uuid    string `json:"uuid"`
	Message string `json:"message"`
}

type Bucket struct {
	Uuid string `json:"uuid"`
}
