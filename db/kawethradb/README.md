# kawethradb <br>
Open Source CSV Database Module For Golang
<br><br><br>
## Commands <br>
### Create DB
```go
package main

import (
	"fmt"
	kawethradb "github.com/Hasan-Kilici/kawethradb"
)

type Ogrenci struct {
	ID    int
	Ad    stringQ
	Soyad string
	Sinif intA
}

func main(){
	ogrenciler := []Ogrenci{
		{ID: 1, Ad: "Ali", Soyad: "Veli", Sinif: 9},
		{ID: 2, Ad: "Ahmet", Soyad: "Mehmet", Sinif: 10},
		{ID: 3, Ad: "Ayşe", Soyad: "Fatma", Sinif: 11},
		{ID: 4, Ad: "Hasan", Soyad: "KILICI", Sinif: 12},
	}

	err := kawethradb.CreateDB("Ogrenciler", "./data/Ogrenciler.csv", ogrenciler)
	if err != nil {
		fmt.Println("err:", err)
		return
	}
}
```
### Insert Single
```go
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
```
### Insert Multiple
```go
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
	ogrenci := []Ogrenci{
		{ID: 1, Ad: "Ali", Soyad: "Veli", Sinif: 9},
		{ID: 2, Ad: "Ahmet", Soyad: "Mehmet", Sinif: 10},
		{ID: 3, Ad: "Ayşe", Soyad: "Fatma", Sinif: 11},
		{ID: 4, Ad: "Hasan", Soyad: "KILICI", Sinif: 12},
	}

	kawethradb.Insert("./data/Ogrenciler.csv", ogrenci)
        fmt.Println("Inserted!")
}
```
### Find
```go
package main

import (
	"fmt"
	kawethradb "github.com/Hasan-Kilici/kawethradb"
)

func main(){
	find, _ := kawethradb.Find("./data/Ogrenciler.csv", "ID", 3)
	fmt.Println(find)
}
```

## Delete
```go
package main

import (
	kawethradb "github.com/Hasan-Kilici/kawethradb"
)

func main(){
  kawethradb.Delete("./data/Ogrenciler.csv", "ID", 2)
}
```
### Update
```go
package main

import (
	"fmt"
	kawethradb "github.com/Hasan-Kilici/kawethradb"
)

func main(){
yeniVeri := []string{"2", "Hasan", "Kılıcı", "12"}
	err = kawethradb.Update("./data/Ogrenciler.csv", "ID", 2, yeniVeri)
	if err != nil {
		fmt.Println("Kayıt güncellenirken bir hata oluştu:", err)
		return
	}

	fmt.Println("Kayıt başarıyla güncellendi.")
}
```
