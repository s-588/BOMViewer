package models

type Material struct {
	ID          int64
	Names       []string
	PrimaryName string
	Unit        Unit
	Description string
	Quantity    string
	Products    []Product
}

type Unit struct {
	ID   int64
	Name string
}

type Product struct {
	ID          int64
	Name        string
	Description string
// How many of material used in this product. 
// This field used only in MaterialView situation where list of products use same material.
	Quantity    string
	Materials   []Material
}

type File struct {
	ID   int64
	Name string
	Path string
}
