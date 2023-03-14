package main

import (
	"encoding/csv"
	"errors"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
	"bufio"
	"io"
	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
	"path/filepath"
)

type Users struct {
	ID           int
	Token        string
	Username     string
	Email        string
	Password     string
	Perms        string
	ProfilePhoto string
}

type Blogs struct {
	ID    int
	Token string
	Title string
	HTML  string
}

func main() {
	users, err := readUserCSV("./data/users.csv")
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	fmt.Println("Users:", users)

	blogs, err := readBlogsCSV("./data/blogs.csv")
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	fmt.Println("Blogs:", blogs)
	
	r := gin.Default()

	r.LoadHTMLGlob("src/*.tmpl")
	r.Static("/static", "./static/")
	r.Static("/components", "./components/")
	r.Static("/uploads", "./uploads/")

	r.GET("/", func(ctx *gin.Context) {
		token, err := ctx.Cookie("token")
		if err != nil {
			ctx.HTML(http.StatusOK, "index.tmpl", gin.H{
				"title":      "Anasayfa",
				"userStatus": "false",
			})
			return
		}

		user, err := findUserByToken(token)
		if err != nil {
			ctx.HTML(http.StatusOK, "index.tmpl", gin.H{
				"title":      "Anasayfa",
				"userStatus": "false",
			})
			return
		}

		ctx.HTML(http.StatusOK, "index.tmpl", gin.H{
			"title":      "Anasayfa",
			"userStatus": "true",
			"userId":     user.ID,
			"username":   user.Username,
			"email":      user.Email,
		})
	})
	
	r.GET("/admin/dashboard",func(ctx *gin.Context){
		token, err := ctx.Cookie("token")
		if err != nil {
			ctx.Redirect(http.StatusFound, "/")
			return
		}

		user, err := findUserByToken(token)
		if err != nil {
			ctx.Redirect(http.StatusFound, "/")
			return
		}
		if !HasRequiredPerms(ctx, []int{8}) {
			ctx.Redirect(http.StatusFound, "/")
			return
		}

		ctx.HTML(http.StatusOK, "dashboard.tmpl", gin.H{
			"title":      "Admin Dashboard",
			"userStatus": "true",
			"userId":     user.ID,
			"username":   user.Username,
			"email":      user.Email,
		})
	})

	r.POST("/register", func(ctx *gin.Context) {
		name := ctx.PostForm("name")
		email := ctx.PostForm("email")
		password := ctx.PostForm("password")
	
		name = strings.ToLower(name)
		email = strings.ToLower(email)
	
		token , err:= registerUser(name, email, password)
		if err != nil {
			ctx.String(http.StatusOK, "Kullanıcı adı Geçersiz ya da Bu Email Kullanılıyor.")	
		}
		ctx.SetCookie("token", token, 36000, "localhost", "", false, true)
	
		ctx.Redirect(http.StatusFound, "/")
	})

	r.POST("/login", func(ctx *gin.Context) {
		email := ctx.PostForm("email")
		password := ctx.PostForm("password")
	
		token, err := loginUser(email, password)
		if err != nil {
			ctx.String(http.StatusUnauthorized, "Email ya da şifre yanlış!")
			return
		}
	
		ctx.SetCookie("token", token, 36000, "/", "", false, true)
		ctx.Redirect(http.StatusFound, "/")
	})
	
	r.POST("/create/Blog",func(ctx *gin.Context){
		if !HasRequiredPerms(ctx, []int{8}) {
			ctx.Redirect(http.StatusFound, "/")
			return
		}

		title := ctx.PostForm("title")
		html := ctx.PostForm("html")
		err := addBlogToCSV("./data/blogs.csv", title, html)
		if err != nil {
			ctx.Redirect(http.StatusFound, "/")	
			return
		}
		ctx.Redirect(http.StatusFound, "/admin/dashboard")
	})

	r.POST("/delete/blog/:id", func(ctx *gin.Context){
		id := ctx.Param("id")
		if !HasRequiredPerms(ctx, []int{8}) {
			ctx.Redirect(http.StatusFound, "/")
			return
		}

		deleteBlog(id);

		ctx.Redirect(http.StatusFound, "/admin/dashboard")
	})

	r.POST("/delete/blogs/:ids",func(ctx *gin.Context){
		ids := ctx.Param("ids")
		idArr := strings.Split(ids, "-")
		if !HasRequiredPerms(ctx, []int{8}) {
			ctx.Redirect(http.StatusFound, "/")
			return
		}

		for _, id := range idArr {
			if err := deleteBlog(id); err != nil {
				ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Blog silinirken hata oluştu"})
				return
			}
		}

		ctx.Redirect(http.StatusFound, "/admin/dashboard")
	})

	r.GET("/api/blogs", func(ctx *gin.Context) {
		blogs, err := readBlogsFromFile("./data/blogs.csv")
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
	
		ctx.JSON(http.StatusOK, blogs)
	})

	r.GET("/api/blogs/:id", func(ctx *gin.Context) {
		id, err := strconv.Atoi(ctx.Param("id"))
		if err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "Geçersiz ID"})
			return
		}
	
		blog, err := readBlogByID("./data/blogs.csv", id)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
	
		ctx.JSON(http.StatusOK, blog)
	})

	r.POST("/upload/photo", func(ctx *gin.Context) {
		ctx.Request.Body = http.MaxBytesReader(ctx.Writer, ctx.Request.Body, 10<<20)

		file, header, err := ctx.Request.FormFile("upload")
		if err != nil {
			ctx.String(http.StatusBadRequest, "Kötü istek")
			return
		}
		defer file.Close()

		filename := header.Filename
		ext := filepath.Ext(filename)
		newFilename := "image" + ext

		out, err := os.Create("./uploads/" + newFilename)
		if err != nil {
			ctx.String(http.StatusInternalServerError, "Sunucu hatası")
			return
		}
		defer out.Close()

		_, err = io.Copy(out, file)
		if err != nil {
			ctx.String(http.StatusInternalServerError, "Sunucu hatası")
			return
		}

		ctx.String(http.StatusOK, "/uploads/"+newFilename)
	})
	r.Run(":3000")
}

func GenerateToken(length int) string {
	rand.Seed(time.Now().UnixNano())

	const letters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	result := make([]byte, length)
	for i := range result {
		result[i] = letters[rand.Intn(len(letters))]
	}
	return string(result)
}

func readUserCSV(filePath string) ([]Users, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	reader := csv.NewReader(file)
	var users []Users
	header := true
	for {
		record, err := reader.Read()
		if err != nil {
			if err.Error() == "EOF" {
				break
			}
			return nil, err
		}

		if header {
			header = false
			continue
		}

		id, err := strconv.Atoi(record[0])
		if err != nil {
			return nil, err
		}

		user := Users{
			ID:           id,
			Token:        record[1],
			Username:     record[2],
			Email:        record[3],
			Password:     record[4],
			Perms:        record[5],
			ProfilePhoto: record[6],
		}

		users = append(users, user)
	}
	return users, nil
}

func readBlogsCSV(filePath string) ([]Blogs, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	reader := csv.NewReader(file)
	var blogs []Blogs
	header := true
	for {
		record, err := reader.Read()
		if err != nil {
			if err.Error() == "EOF" {
				break
			}
			return nil, err
		}

		if header {
			header = false
			continue
		}

		id, err := strconv.Atoi(record[0])
		if err != nil {
			return nil, err
		}

		blog := Blogs{
			ID:    id,
			Token: record[1],
			Title: record[2],
			HTML:  record[3],
		}

		blogs = append(blogs, blog)
	}
	return blogs, nil
}

func HasRequiredPerms(ctx *gin.Context, requiredPerms []int) bool {
    token, err := ctx.Cookie("token")
    if err != nil {
        return false
    }
    
    file, err := os.Open("./data/users.csv")
    if err != nil {
        log.Fatal(err)
    }
    defer file.Close()

    reader := csv.NewReader(file)
    lines, err := reader.ReadAll()
    if err != nil {
        log.Fatal(err)
    }

    var perms []string
    for _, line := range lines {
        if line[1] == token {
            perms = strings.Split(line[5], ">")
            break
        }
    }

    for _, requiredPerm := range requiredPerms {
        found := false
        for _, perm := range perms {
            permInt, err := strconv.Atoi(strings.TrimSpace(perm))
            if err != nil {
                log.Fatal(err)
            }
            if permInt == requiredPerm {
                found = true
                break
            }
        }
        if !found {
            return false
        }
    }
    return true
}

func registerUser(username, email, password string) (string, error) {
	file, err := os.OpenFile("./data/users.csv", os.O_RDWR|os.O_CREATE, 0755)
	if err != nil {
		return "", err
	}
	defer file.Close()

	reader := csv.NewReader(file)

	usernameKey := strings.ToLower(username)
	emailKey := strings.ToLower(email)

	records, err := reader.ReadAll()
	if err != nil {
		return "", err
	}
	for _, record := range records {
		if len(record) != 7 {
			continue
		}
		if record[2] == usernameKey {
			return "", errors.New("Bu Kullanıcı adı kullanılıyor")
		}
		if record[3] == emailKey {
			return "", errors.New("Bu Email Kullanılıyor")
		}
	}

	token := GenerateToken(16)

	newRecord := []string{
		strconv.Itoa(len(records) + 1),
		token,
		username,
		email,
		hashPassword(password),
		"0>1",
		"./images/pp.png",
	}

	writer := csv.NewWriter(file)
	if err := writer.Write(newRecord); err != nil {
		return "", err
	}
	writer.Flush()

	return token, nil
}

func hashPassword(password string) string {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		log.Println(err)
		return ""
	}
	return string(hash)
}

func loginUser(email, password string) (string, error) {
    file, err := os.Open("./data/users.csv")
    if err != nil {
        return "", err
    }
    defer file.Close()

    reader := csv.NewReader(file)
    records, err := reader.ReadAll()
    if err != nil {
        return "", err
    }

    email = strings.ToLower(email)
    for _, record := range records {
        if len(record) != 7 {
            continue
        }
        if strings.ToLower(record[3]) == email {
            if checkPassword(record[4], password) {
                token := record[1]

                file, err := os.OpenFile("./data/users.csv", os.O_WRONLY, 0755)
                if err != nil {
                    return "", err
                }
                defer file.Close()

                writer := csv.NewWriter(file)
                if err := writer.WriteAll(records); err != nil {
                    return "", err
                }
                return token, nil
            } else {
                return "", errors.New("Geçersiz şifre")
            }
        }
    }

    return "", errors.New("Kullanıcı Bulunamadı")
}

func checkPassword(hashedPassword, password string) bool {
    err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
    return err == nil
}

func loadUsersFromCSV() ([]Users, error) {
	file, err := os.Open("./data/users.csv")
	if err != nil {
		return nil, err
	}
	defer file.Close()

	csvReader := csv.NewReader(file)
	csvData, err := csvReader.ReadAll()
	if err != nil {
		return nil, err
	}

	var users []Users
	for _, row := range csvData {
		userID, err := strconv.Atoi(row[0])
		if err != nil {
			return nil, err
		}
		user := Users{
			ID:           userID,
			Token:        row[1],
			Username:     row[2],
			Email:        row[3],
			Password:     row[4],
			Perms:        row[5],
			ProfilePhoto: row[6],
		}
		users = append(users, user)
	}

	return users, nil
}

func findUserByToken(token string) (*Users, error) {
	users, err := loadUsersFromCSV()
	if err != nil {
		return nil, err
	}

	for _, user := range users {
		if user.Token == token {
			return &user, nil
		}
	}

	return nil, errors.New("Kullanıcı Bulunamadı")
}

func addBlogToCSV(filePath string, title string, html string) error {
	file, err := os.OpenFile(filePath, os.O_APPEND|os.O_WRONLY, os.ModeAppend)
	if err != nil {
		return err
	}
	defer file.Close()

	token := GenerateToken(16)

	lines := 0
	f, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		lines++
	}

	writer := csv.NewWriter(file)

	err = writer.Write([]string{strconv.Itoa(lines + 1), token, title, html})
	if err != nil {
		return err
	}

	writer.Flush()

	return nil
}

func deleteBlog(id string) error {
	file, err := os.OpenFile("./data/blogs.csv", os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		fmt.Println("Dosya açılırken hata oluştu")
	}
	defer file.Close()

	reader := csv.NewReader(file)
	lines, err := reader.ReadAll()
	if err != nil {
		fmt.Println("Dosya okunurken hata oluştu")
	}
	
	found := false
	for i, line := range lines {
		if line[0] == id {
			lines = append(lines[:i], lines[i+1:]...)
			found = true
			break
		}
	}

	if !found {
		fmt.Println("ID bulunamadı")
	}

	file.Truncate(0)
	file.Seek(0, 0)

	writer := csv.NewWriter(file)
	defer writer.Flush()

	err = writer.WriteAll(lines)
	if err != nil {
		fmt.Println("Dosya yazılırken hata oluştu")
	}
	return nil
}

func readBlogsFromFile(filePath string) ([]Blogs, error) {
	f, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	reader := csv.NewReader(f)

	var blogs []Blogs

	for {
		row, err := reader.Read()
		if err == io.EOF {
			break
		} else if err != nil {
			return nil, err
		}

		id, _ := strconv.Atoi(row[0])
		blog := Blogs{
			ID:    id,
			Title: row[2],
			HTML:  row[3],
		}

		blogs = append(blogs, blog)
	}

	return blogs, nil
}
func readBlogByID(filePath string, id int) (*Blogs, error) {
	f, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	reader := csv.NewReader(f)

	for {
		row, err := reader.Read()
		if err == io.EOF {
			break
		} else if err != nil {
			return nil, err
		}

		if row[0] == strconv.Itoa(id) {
			blog := &Blogs{
				ID:    id,
				Title: row[2],
				HTML:  row[3],
			}
			return blog, nil
		}
	}

	return nil, fmt.Errorf("Blog Bulunamadı")
}