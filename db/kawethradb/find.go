package main

import (
	"fmt"
	kawethradb "github.com/Hasan-Kilici/kawethradb"
)

func main(){
	find, _ := kawethradb.Find("./Ogrenciler.csv", "ID", 3)
	fmt.Println(find)
}
