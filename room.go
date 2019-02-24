package main

type RoomState struct {
	Num         int
	Status      string
	GuestInfo   string
	CheckinTime string
}

type RoomDetails struct {
	RoomNum string
	NumBeds int
	BedSize string
}
