package models

type Material struct{
	ID int
	Names []string
	Unit string
	Description string
}

type Product struct{
	ID int
	Name string
	Description string
	Materials []int // ID's of materials
}