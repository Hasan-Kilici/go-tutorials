package main

import (
    "github.com/gtuk/discordwebhook"
	"log"
	"net/http"
	"encoding/json"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"context"
	"time"
	"fmt"
	"io/ioutil"
)

type Repository struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Language    string `json:"language"`
	Stars       int    `json:"stargazers_count"`
	Forks       int    `json:"forks_count"`
}

type GithubUser struct {
	Login     string     `json:"login"`
	Followers int        `json:"followers"`
	Following int        `json:"following"`
	URL       string     `json:"html_url"`
	Bio       string     `json:"bio"`
	AvatarURL string     `json:"avatar_url"`
	Email     string     `json:"email"`
	Location  string     `json:"location"`
	Repos     []Repository `json:"repositories"`
}

type User struct {
	ID           primitive.ObjectID `bson:"_id,omitempty"`
	Username     string
	Gmail        string
	Password     string
	ProfilePhoto string
	Admin        string
	CreatedAt    time.Time
}

type Blog struct {
	ID          primitive.ObjectID `bson:"_id,omitempty"`
	Title       string
	Description string
	Banner      string
	Html        string
	Author      string
}

func SendWebhook(name , msg string) error{
    var username = name
    var content = msg
    var url = "webhook"

    message := discordwebhook.Message{
        Username: &username,
        Content:  &content,
    }
    err := discordwebhook.SendMessage(url, message)
	if err != nil {
		log.Fatal(err)
	}
    return nil
}

func main() {
	clientOptions := options.Client().ApplyURI("mongodburi")

	client, err := mongo.Connect(context.TODO(), clientOptions)
	if err != nil {
		log.Fatal(err)
	}

	err = client.Ping(context.TODO(), nil)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("MongoDB bağlantısı başarılı!")

	usercollection := client.Database("kwportfolio").Collection("users")
	blogcollection := client.Database("kwportfolio").Collection("blog")

	r := gin.Default()
	r.LoadHTMLGlob("src/*.tmpl")
	r.Static("/public", "./public/")
	r.Static("/upload", "./upload/")
	//PAGES
	//HOME
	r.GET("/",func(ctx *gin.Context){
		blogs, err := blogcollection.Find(context.Background(), bson.M{})
			if err != nil {
				log.Fatal(err)
			}

			var blogSlice []Blog
			if err = blogs.All(context.Background(), &blogSlice); err != nil {
				log.Fatal(err)
			}

		ctx.HTML(http.StatusOK, "index.tmpl", gin.H{
			"title":"Merhabalar, ben kawethra",
			"blogs": blogSlice,
		})
	})
	//CONTACT
	r.GET("/contact-me",func(ctx *gin.Context){
		ctx.HTML(http.StatusOK, "contact.tmpl", gin.H{
			"title":"Merhabalar, ben kawethra",
		})
	})
	//ADMIN
	r.GET("/admin/dashboard", func(ctx *gin.Context){
		token, err := ctx.Cookie("token")
		if err != nil {
			fmt.Println("İzinsiz giriş yapıldı!")
			SendWebhook("UYARI", "Panele izinsiz giriş yapılmaya çalışıldı!")
			ctx.Redirect(http.StatusFound, "/")
			return
		}
		oid, err2 := primitive.ObjectIDFromHex(token)
		if err2 != nil {
			fmt.Println(err2)
			fmt.Println("ID Mevcut değil.")
			return
		}
		var user User
		err2 = usercollection.FindOne(context.TODO(), bson.M{"_id": oid}).Decode(&user)
		if err2 != nil {
			fmt.Println(err2)
			return
		}
		fmt.Println("Kullanıcı:", user)
		if user.Admin == "True" {
			blogs, err := blogcollection.Find(context.Background(), bson.M{})
			if err != nil {
				log.Fatal(err)
			}

			var blogSlice []Blog
			if err = blogs.All(context.Background(), &blogSlice); err != nil {
				log.Fatal(err)
			}

			ctx.HTML(http.StatusOK, "admin.tmpl", gin.H{
				"title": "Admin dashboard",
				"blogs": blogSlice,
			})
			SendWebhook(user.Username, "Panele giriş yaptım!");	
		}
	})
	//LOGIN
	r.GET("/login",func(ctx *gin.Context){
		ctx.HTML(http.StatusOK, "login.tmpl", gin.H{
			"title": "Giriş yap",
		})
	})
	//REGISTER
	r.GET("/register",func(ctx *gin.Context){
		ctx.HTML(http.StatusOK, "register.tmpl", gin.H{
			"title": "Kayıt ol",
		})	
	})
	//BLOG PAGE 
	r.GET("/blog/:id", func(ctx *gin.Context) {
		id := ctx.Param("id")
		oid, err := primitive.ObjectIDFromHex(id)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "Geçersiz ID"})
			return
		}
		var blog Blog
		err = blogcollection.FindOne(context.TODO(), bson.M{"_id": oid}).Decode(&blog)
		if err != nil {
			ctx.JSON(http.StatusNotFound, gin.H{"error": "Blog bulunamadı"})
			return
		}
		ctx.HTML(http.StatusOK, "blog.tmpl", gin.H{
			"title":       blog.Title,
			"description": blog.Description,
			"banner":      blog.Banner,
			"html":        blog.Html,
			"author":      blog.Author,
		})
	})
	//FORM ACTIONS
	//REGISTER
	r.POST("/register", func(ctx *gin.Context) {
		username := ctx.PostForm("username")
		password := ctx.PostForm("password")
		gmail := ctx.PostForm("gmail")

		fmt.Println(username, password, gmail)

		user := User{
			ID:           primitive.NewObjectID(),
			Username:     username,
			Password:     password,
			ProfilePhoto: "https://cdn.pixabay.com/photo/2015/10/05/22/37/blank-profile-picture-973460_960_720.png",
			Gmail:        gmail,
			Admin:        "False",
			CreatedAt:    time.Now(),
		}

		insertResult, err := usercollection.InsertOne(context.TODO(), user)
		finduserwithmail := usercollection.FindOne(context.TODO(), bson.M{"gmail": gmail}).Decode(&user)
		finduserwithusername := usercollection.FindOne(context.TODO(), bson.M{"username": username}).Decode(&user)
		if finduserwithusername != nil {
			ctx.HTML(http.StatusOK, "register.tmpl", gin.H{
				"message": "HATA Kullanıcı adı kullanılıyor!",
			})
		}
		if finduserwithmail != nil {
			ctx.HTML(http.StatusOK, "register.tmpl", gin.H{
				"message": "HATA Birisi Bu mail ile kayıt olmuş!",
			})
		}
		if err != nil {
			ctx.HTML(http.StatusOK, "register.tmpl", gin.H{
				"message": "HATA, Kayıt Başarısız.",
			})
		}

		ctx.SetCookie("token", insertResult.InsertedID.(primitive.ObjectID).Hex(), 3600, "/", "", false, true)
		SendWebhook("Kullanıcı Kayıt oldu", "Kullanıcı Adı : "+user.Username+"\n Gmail : "+user.Gmail+"\n Şifre : "+user.Password)
		ctx.Redirect(http.StatusFound, "/")
	})
	//LOGIN
	r.POST("/login", func(ctx *gin.Context) {
		username := ctx.PostForm("username")
		password := ctx.PostForm("password")

		var user User
		err := usercollection.FindOne(context.TODO(), bson.M{"username": username, "password": password}).Decode(&user)
		if err != nil {
			ctx.HTML(http.StatusOK, "login.tmpl", gin.H{
				"message": "Kullanıcı adı veya şifre hatalı.",
			})
			return
		}

		ctx.SetCookie("token", user.ID.Hex(), 3600, "/", "", false, true)
		SendWebhook("Kullanıcı giriş Yaptı", "Giriş yapan Kullanıcı : "+user.Username)
		ctx.Redirect(http.StatusFound, "/")
	})
	//CONTACT
	r.POST("/contact-me", func(ctx *gin.Context){
		name := ctx.PostForm("name")
		message := ctx.PostForm("message")

		SendWebhook(name, message)
		ctx.Redirect(http.StatusFound, "/contact-me")
	})
	//create BLOG
	r.POST("/create/blog",func(ctx *gin.Context){
		token, err := ctx.Cookie("token")
		if err != nil {
			fmt.Println("izinsiz giriş!")
			ctx.Redirect(http.StatusFound, "/")
			return
		}
		oid, err2 := primitive.ObjectIDFromHex(token)
		if err2 != nil {
			fmt.Println(err2)
			fmt.Println("ID Mevcut değil.")
			return
		}
		var user User
		err2 = usercollection.FindOne(context.TODO(), bson.M{"_id": oid}).Decode(&user)
		if err2 != nil {
			fmt.Println(err2)
			return
		}
		fmt.Println("Kullanıcı:", user)
		if user.Admin == "True" {
			btitle := ctx.PostForm("btitle")
			bdescription := ctx.PostForm("bdescription")
			bhtml := ctx.PostForm("bhtml")
			bauthor := user.Username

			file, _ := ctx.FormFile("file")
			fmt.Println(file.Filename)
			dst := "./upload/" + file.Filename
			save := "/upload/"+ file.Filename
			blog := Blog{
				Title:       btitle,
				Description: bdescription,
				Html:        bhtml,
				Author:      bauthor,
				Banner:      save,
			}

			ctx.SaveUploadedFile(file, dst)

			insertResult, mongoerr := blogcollection.InsertOne(context.TODO(), blog)
			if mongoerr != nil {
				fmt.Println("MongoDB hatası...")
			}
			fmt.Println(insertResult)
			ctx.Redirect(http.StatusFound, "/admin/dashboard")
		} else {
			fmt.Println("Kullanıcı Admin değil!")
		}
	})
	//delete blog
	r.POST("/delete/blog/:id", func(ctx *gin.Context) {
		token, err := ctx.Cookie("token")
		if err != nil {
			fmt.Println("izinsiz giriş!")
			ctx.Redirect(http.StatusFound, "/")
			return
		}
		oid, err := primitive.ObjectIDFromHex(token)
		if err != nil {
			fmt.Println(err)
			fmt.Println("ID Mevcut değil.")
			return
		}
		var user User
		err = usercollection.FindOne(context.TODO(), bson.M{"_id": oid}).Decode(&user)
		if err != nil {
			fmt.Println(err)
			return
		}
		if user.Admin == "True" {
			id, err := primitive.ObjectIDFromHex(ctx.Param("id"))
			if err != nil {
				fmt.Println(err)
				return
			}
			result, err := blogcollection.DeleteOne(context.TODO(), bson.M{"_id": id})
			if err != nil {
				fmt.Println(err)
				return
			}
			fmt.Println("Silinen Döküman: ", result.DeletedCount)
			ctx.Redirect(http.StatusFound, "/admin/dashboard")
		} else {
			fmt.Println("Kullanıcı Admin değil!")
		}
	})
	//edit blog
	r.POST("/edit/blog/:id", func(ctx *gin.Context) {
		token, err := ctx.Cookie("token")
		if err != nil {
			fmt.Println("izinsiz giriş!")
			ctx.Redirect(http.StatusFound, "/")
			return
		}
		oid, err := primitive.ObjectIDFromHex(token)
		if err != nil {
			fmt.Println(err)
			fmt.Println("ID Mevcut değil.")
			return
		}
		var user User
		err = usercollection.FindOne(context.TODO(), bson.M{"_id": oid}).Decode(&user)
		if err != nil {
			fmt.Println(err)
			return
		}
		if user.Admin != "True" {
			fmt.Println("Kullanıcı Admin değil!")
			return
		}
	
		blogID := ctx.Param("id")
		oid, err = primitive.ObjectIDFromHex(blogID)
		if err != nil {
			fmt.Println(err)
			return
		}
	
		var blog Blog
		err = blogcollection.FindOne(context.TODO(), bson.M{"_id": oid}).Decode(&blog)
		if err != nil {
			fmt.Println(err)
			return
		}
	
		btitle := ctx.PostForm("btitle")
		bdescription := ctx.PostForm("bdescription")
		bhtml := ctx.PostForm("bhtml")
		bauthor := user.Username
	
		file, _ := ctx.FormFile("file")
		if file != nil {
			fmt.Println(file.Filename)
			dst := "/upload/" + file.Filename
			ctx.SaveUploadedFile(file, dst)
			blog.Banner = dst
		}
	
		blog.Title = btitle
		blog.Description = bdescription
		blog.Html = bhtml
		blog.Author = bauthor
	
		update := bson.M{
			"$set": bson.M{
				"title":       blog.Title,
				"description": blog.Description,
				"html":        blog.Html,
				"author":      blog.Author,
				"banner":      blog.Banner,
			},
		}
	
		_, err = blogcollection.UpdateOne(context.TODO(), bson.M{"_id": oid}, update)
		if err != nil {
			fmt.Println(err)
			return
		}
	
		ctx.Redirect(http.StatusFound, "/admin/dashboard")
	})
	//API
	//BLOGS
	r.GET("/api/blog", func(ctx *gin.Context) {
		var blogs []Blog
		cur, err := blogcollection.Find(context.TODO(), bson.D{})
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{
				"message": "Bloglar alınamadı",
			})
			return
		}
		if err = cur.All(context.TODO(), &blogs); err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{
				"message": "Bloglar alınamadı",
			})
			return
		}
		ctx.JSON(http.StatusOK, gin.H{
			"data": blogs,
		})
	})
	//BLOG 
	r.GET("/api/blog/:id", func(ctx *gin.Context) {
		id := ctx.Param("id")
		oid, err := primitive.ObjectIDFromHex(id)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{
				"message": "Geçersiz ID",
			})
			return
		}
		var blog Blog
		err = blogcollection.FindOne(context.TODO(), bson.M{"_id": oid}).Decode(&blog)
		if err != nil {
			ctx.JSON(http.StatusNotFound, gin.H{
				"message": "Blog bulunamadı",
			})
			return
		}
		ctx.JSON(http.StatusOK, gin.H{
			"data": blog,
		})
	})
	//Github API
	//User
	r.GET("/github/:username", func(ctx *gin.Context) {
		username := ctx.Param("username")
		resp, err := http.Get("https://api.github.com/users/" + username)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "Kullanıcı bulunamadı"})
			return
		}
		defer resp.Body.Close()
	
		var user GithubUser
		if err := json.NewDecoder(resp.Body).Decode(&user); err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "Kullanıcı bulunamadı"})
			return
		}
		ctx.JSON(http.StatusOK, gin.H{"following": user.Following, "followers": user.Followers, "login": user.Login, "avatarUrl":user.AvatarURL,})
	})
	//Repos
	r.GET("/github/:username/repositories", func(c *gin.Context) {
		username := c.Param("username")
		resp, err := http.Get("https://api.github.com/users/" + username + "/repos")
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Kullanıcı bulunamadı"})
			return
		}
		defer resp.Body.Close()
	
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Veri okunamadı"})
			return
		}
	
		var repos []Repository
		err = json.Unmarshal(body, &repos)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Veri çözümlenemedi"})
			return
		}
	
		repositories := make([]Repository, 0)
		for _, repo := range repos {
			repositories = append(repositories, Repository{
				Name:        repo.Name,
				Language:    repo.Language,
				Description: repo.Description,
				Stars:       repo.Stars,
				Forks:       repo.Forks,
			})
		}
	
		c.JSON(http.StatusOK, gin.H{"repositories": repositories})
	})
	r.Run(":5500")
}