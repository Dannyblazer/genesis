package main

import (
	"genesis/controllers"
	"genesis/initializers"
	"genesis/middleware"

	"github.com/gin-gonic/gin"
	_ "github.com/go-sql-driver/mysql"
)

// func main() {
// 	r := mux.NewRouter()
// 	r.HandleFunc("/books/{title}/page/{page}", func(w http.ResponseWriter, r *http.Request) {
// 		vars := mux.Vars(r)
// 		first := vars["title"]
// 		last := vars["page"]
// 		fmt.Fprintln(w, "Welcome my people: ", first, last)
// 	}).Methods("GET")

// 	fs := http.FileServer(http.Dir("static/"))
// 	http.Handle("/static/", http.StripPrefix("/static/", fs))

//		if err := http.ListenAndServe(":8000", r); err != nil {
//			log.Fatalf("Server failed: %v", err)
//		}
//	}
// type user struct {
// 	id         int
// 	username   string
// 	password   string
// 	created_at time.Time
// }

// func main() {
// 	connStr := "genesis:golangforlife@(127.0.0.1:3306)/genesis_db?parseTime=true"

// 	db, err := sql.Open("mysql", connStr)

// 	errChecker(err)
// 	fmt.Println("Connection Successfull")
// 	defer db.Close()

// 	{ // Create a new table
// 		query := `
//             CREATE TABLE IF NOT EXISTs users (
//                 id INT AUTO_INCREMENT,
//                 username TEXT NOT NULL,
//                 password TEXT NOT NULL,
//                 created_at DATETIME,
//                 PRIMARY KEY (id)
//             );`

// 		if _, err := db.Exec(query); err != nil {
// 			log.Fatal(err)
// 		}
// 	}

// 	{ // Insert a new user
// 		username := "Danny"
// 		password := "secret"
// 		createdAt := time.Now()

// 		result, err := db.Exec(`INSERT INTO users (username, password, created_at) VALUES (?, ?, ?)`, username, password, createdAt)
// 		if err != nil {
// 			log.Fatal(err)
// 		}

// 		id, err := result.LastInsertId()
// 		errChecker(err)
// 		fmt.Printf("UserID: %v \n", id)
// 	}

// 	{ // Query a single user
// 		var (
// 			id        int
// 			username  string
// 			password  string
// 			createdAt time.Time
// 		)

// 		query := "SELECT id, username, password, created_at FROM users WHERE id = ?"
// 		if err := db.QueryRow(query, 1).Scan(&id, &username, &password, &createdAt); err != nil {
// 			log.Fatal(err)
// 		}

// 		fmt.Println(id, username, password, createdAt)
// 	}

// 	{ // Query all users
// 		type user struct {
// 			id        int
// 			username  string
// 			password  string
// 			createdAt time.Time
// 		}

// 		rows, err := db.Query(`SELECT id, username, password, created_at FROM users`)
// 		if err != nil {
// 			log.Fatal(err)
// 		}
// 		defer rows.Close()

// 		var users []user
// 		for rows.Next() {
// 			var u user

// 			err := rows.Scan(&u.id, &u.username, &u.password, &u.createdAt)
// 			if err != nil {
// 				log.Fatal(err)
// 			}
// 			users = append(users, u)
// 		}
// 		if err := rows.Err(); err != nil {
// 			log.Fatal(err)
// 		}

// 		fmt.Printf("%#v \n", users)
// 	}

// 	{
// 		_, err := db.Exec(`DELETE FROM users WHERE id = ?`, 1)
// 		if err != nil {
// 			log.Fatal(err)
// 		}
// 	}
// }

// Basic Middleware
// func logging(f http.HandlerFunc) http.HandlerFunc {
// 	return func(w http.ResponseWriter, r *http.Request) {
// 		log.Println(r.URL.Path)
// 		f(w, r)
// 	}
// }

// func foo(w http.ResponseWriter, r *http.Request) {
// 	fmt.Fprintln(w, "foo")
// }

// func bar(w http.ResponseWriter, r *http.Request) {
// 	fmt.Fprintln(w, "bar")
// }
// func main() {
// 	http.HandleFunc("/foo", logging(foo))
// 	http.HandleFunc("/bar", logging(bar))

// 	fmt.Println("Starting Server at :8000")
// 	if err := http.ListenAndServe(":8000", nil); err != nil {
// 		log.Fatalf("Server failed: %v", err)
// 	}
// }
// func errChecker(err error) {
// 	if err != nil {
// 		log.Fatal(err)
// 	}
// }

// func main() {
// 	const url = "https://www.nairaland.com/"
// 	response, err := http.Get(url)

// 	if err != nil {
// 		log.Fatal(err)
// 		return
// 	}

// 	fmt.Printf("Response Type: %T\n", response)
// 	defer response.Body.Close()

// 	if response.StatusCode != http.StatusOK {
// 		fmt.Printf("Failed with status %v\n", response.Status)
// 		return
// 	}

// 	body, err := io.ReadAll(response.Body)

//		if err != nil {
//			fmt.Printf("Failed to read response body %v\n", err)
//			return
//		}
//		fmt.Printf("Response: %T\n", string(body))
//	}
func init() {
	initializers.LoadEnvVariables()
	initializers.ConnectDB()
}
func main() {
	router := gin.Default()
	// Account Handlers
	router.GET("account/", middleware.RequireAuth, controllers.AccountDetail)
	router.POST("account/create/", controllers.AccountCreate)
	router.POST("account/login/", controllers.AccountLogin)
	router.PUT("account/", middleware.RequireAuth, controllers.AccountUpdate)
	router.DELETE("account/", middleware.RequireAuth, controllers.AccountDelete)

	// Post Handlers
	router.POST("posts/", middleware.RequireAuth, controllers.PostsCreate)
	router.GET("posts/:id", middleware.RequireAuth, controllers.PostGet)
	router.PUT("posts/:id", middleware.RequireAuth, controllers.PostUpdate)
	router.GET("posts/", middleware.RequireAuth, controllers.PostList)
	router.DELETE("posts/:id", middleware.RequireAuth, controllers.PostDelete)

	//Bank
	router.Run() // listen and serve on 0.0.0.0:8080
}
