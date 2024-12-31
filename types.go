package main

import (
	"math/rand"
	"time"
)

type User struct {
	ID         uint        `json:"id"`
	Email      string      `json:"email"`
	Password   string      `json:"-"`
	FirstName  string      `json:"firstName"`
	LastName   string      `json:"lastName"`
	Containers []Container `json:"containers"`
	CreatedAt  time.Time   `json:"createdAt"`
	UpdatedAt  time.Time   `json:"updatedAt"`
}

type CreateContainerRequest struct {
	Name     string `json:"name"`
	Location string `json:"location"`
}

type Container struct {
	ID       int    `json:"id"`
	Name     string `json:"name"`
	QRCode   int    `json:"qrCode"`
	Number   int    `json:"number"`
	Location string `json:"location"`
	//images
	//Items []Ite,
	UserId    int       `json:"userId"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

func NewContainer(name, location string) *Container {
	return &Container{
		ID:        rand.Intn(10000),
		Name:      name,
		QRCode:    rand.Intn(10000),
		Number:    rand.Intn(1000),
		Location:  location,
		UserId:    rand.Intn(10000),
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
	}

}
