package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"html/template"
	"io/ioutil"
	"net/http"

	//	"strconv"

	_ "github.com/lib/pq"

	"github.com/google/uuid"
	"github.com/gramework/gramework"
)

const (
	host     = "localhost"
	port     = 5432
	user     = "postgres"
	password = "123123"
	dbname   = "volhina_site"
)

type PageData struct {
	StatisticsYAMetrika []string
	StatusAuthorization bool
}

func main() {
	psqlInfo := fmt.Sprintf("host=%s port=%d user=%s "+"password=%s dbname=%s sslmode=disable", host, port, user, password, dbname)
	db, err := sql.Open("postgres", psqlInfo)
	if err != nil {
		panic(err)
	}
	defer db.Close()

	err = db.Ping()
	if err != nil {
		panic(err)
	}

	app := gramework.New()
	app.Use(app.CORSMiddleware())
	app.GET("/css/*any", app.ServeDir("./"))
	app.GET("/js/*any", app.ServeDir("./"))
	app.GET("/", func(ctx *gramework.Context) {
		/*visits, views, visitors := getStatisticsByYandexMetrics()

		str_visits := strconv.FormatInt(int64(visits), 10)
		str_views := strconv.FormatInt(int64(views), 10)
		str_visitors := strconv.FormatInt(int64(visitors), 10) */
		fmt.Println(ctx.Cookies.Exists("sk"))
		tmpl := template.Must(template.ParseFiles("templates/index.html"))
		data := PageData{
			StatusAuthorization: ctx.Cookies.Exists("sk"),
			//StatisticsYAMetrika: []string{"Визиты: " + str_visits, "Просмотры: " + str_views, "Посетители: " + str_visitors},
		}
		tmpl.Execute(ctx.HTML(), data)

	})
	app.GET("/photo", func(ctx *gramework.Context) {
		tmpl := template.Must(template.ParseFiles("templates/photo.html"))
		data := PageData{
			StatusAuthorization: ctx.Cookies.Exists("sk"),
		}
		tmpl.Execute(ctx.HTML(), data)
	})
	app.GET("/signin", func(ctx *gramework.Context) {
		tmpl := template.Must(template.ParseFiles("templates/registration.html"))
		data := PageData{
			StatusAuthorization: ctx.Cookies.Exists("sk"),
		}
		tmpl.Execute(ctx.HTML(), data)
	})
	app.GET("/movie", func(ctx *gramework.Context) {
		tmpl := template.Must(template.ParseFiles("templates/movie.html"))
		data := PageData{
			StatusAuthorization: ctx.Cookies.Exists("sk"),
		}
		tmpl.Execute(ctx.HTML(), data)
	})
	app.GET("/authorization", func(ctx *gramework.Context) {
		tmpl := template.Must(template.ParseFiles("templates/authorization.html"))
		data := PageData{
			StatusAuthorization: ctx.Cookies.Exists("sk"),
		}
		tmpl.Execute(ctx.HTML(), data)
	})
	app.POST("/params", func(ctx *gramework.Context) {
		if string(ctx.FormValue("name")) == "" {
			ctx.Error("Error! Empty username", 200)
			return
		} else if string(ctx.FormValue("email")) == "" {
			ctx.Error("Error! Empty email", 200)
			return
		} else if string(ctx.FormValue("username")) == "" {
			ctx.Error("Error! Empty username", 200)
			return
		} else if string(ctx.FormValue("password")) == "" {
			ctx.Error("Error! Empty password", 200)
			return
		} else if string(ctx.FormValue("confirm_password")) == "" {
			ctx.Error("Error! Empty confirim password", 200)
			return
		}
		if string(ctx.FormValue("password")) != string(ctx.FormValue("confirm_password")) {
			ctx.Error("Error! password != confirim_password", 200)
			return
		}

		sqlStatement := `INSERT INTO site (email, nick_name, password) VALUES ($1, $2, $3)`
		_, err = db.Exec(sqlStatement, string(ctx.FormValue("email")),
			string(ctx.FormValue("username")),
			string(ctx.FormValue("password")))
		if err != nil {
			panic(err)
		}
		ctx.Redirect("/", 200)

	})
	app.POST("/params1", func(ctx *gramework.Context) {
		if string(ctx.FormValue("name")) == "" {
			ctx.Error("Error! Empty username", 200)
			return
		} else if string(ctx.FormValue("password")) == "" {
			ctx.Error("Error! Empty email", 200)
			return
		}
		row := db.QueryRow("SELECT * FROM site WHERE nick_name = $1", string(ctx.FormValue("name")))
		var id int
		var email string
		var nickname string
		var password string
		err := row.Scan(&id, &email, &nickname, &password)
		if err == sql.ErrNoRows {
			fmt.Println("юзер не найден!")
			ctx.Redirect("authorization", 200)
			return
		} else if err != nil {
			return
		}
		if password != string(ctx.FormValue("password")) {
			// ctx.Redirect("authorization", 200)
			return
		}
		sessID := uuid.New().String()
		fmt.Println("session_key", sessID)
		ctx.Cookies.Set("sk", sessID)
		ctx.Redirect("/", 200)
	})
	app.ListenAndServe(":3333")
}
func getStatisticsByYandexMetrics() (int, int, int) {
	var client http.Client
	resp, err1 := client.Get("http://api-metrika.yandex.ru/stat/v1/data?ids=51066758&oauth_token=token&metrics=ym:s:visits,ym:s:pageviews,ym:s:users&dimensions=ym:s:date")
	if err1 != nil {
		// err
	}
	defer resp.Body.Close()
	bodyString := ""

	if resp.StatusCode == http.StatusOK {
		bodyBytes, _ := ioutil.ReadAll(resp.Body)
		bodyString = string(bodyBytes)
	}
	type JSONBody struct {
		Query struct {
			Ids           []int    `json:"ids"`
			Dimensions    []string `json:"dimensions"`
			Metrics       []string `json:"metrics"`
			Sort          []string `json:"sort"`
			Date1         string   `json:"date1"`
			Date2         string   `json:"date2"`
			Limit         int      `json:"limit"`
			Offset        int      `json:"offset"`
			Group         string   `json:"group"`
			AutoGroupSize string   `json:"auto_group_size"`
			Quantile      string   `json:"quantile"`
			OfflineWindow string   `json:"offline_window"`
			Attribution   string   `json:"attribution"`
			Currency      string   `json:"currency"`
		} `json:"query"`
		Data []struct {
			Dimensions []struct {
				Name string `json:"name"`
			} `json:"dimensions"`
			Metrics []float64 `json:"metrics"`
		} `json:"data"`
		TotalRows        int       `json:"total_rows"`
		TotalRowsRounded bool      `json:"total_rows_rounded"`
		Sampled          bool      `json:"sampled"`
		SampleShare      float64   `json:"sample_share"`
		SampleSize       int       `json:"sample_size"`
		SampleSpace      int       `json:"sample_space"`
		DataLag          int       `json:"data_lag"`
		Totals           []float64 `json:"totals"`
		Min              []float64 `json:"min"`
		Max              []float64 `json:"max"`
	}
	var body JSONBody
	if err := json.Unmarshal([]byte(bodyString), &body); err != nil {
		fmt.Println(err)
	}
	return int(body.Totals[0]), int(body.Totals[1]), int(body.Totals[2])
}
