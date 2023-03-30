package main

import (
	kawethradb "github.com/Hasan-Kilici/kawethradb"
)

func main(){
  kawethradb.Delete("./Ogrenciler.csv", "ID", 2)
}
