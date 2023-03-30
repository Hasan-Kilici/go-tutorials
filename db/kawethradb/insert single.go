package main

import (
	"fmt"
	kawethradb "github.com/Hasan-Kilici/kawethradb"
)

type Ogrenci struct {
	ID    int
	Ad    string
	Soyad string
	Sinif int
}

func main() {
	ogrenci := Ogrenci{ID: 1, Ad: "Ali", Soyad: "Veli", Sinif: 9}
	kawethradb.Insert("./data/Ogrenciler.csv", ogrenci)
        fmt.Println("Inserted!")
}
