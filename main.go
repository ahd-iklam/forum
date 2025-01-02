package main

import (
	"database/sql"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"strconv"

	_ "github.com/mattn/go-sqlite3"
)

type Post struct {
	ID           int
	Title        string
	Content      string
	CreatedAt    string
	Username     string
	Categories   string // To store the concatenated categories
	CommentCount int
	LikeCount    int
	DislikeCount int
}

type User struct {
	ID       int
	Username string
	Email    string
}

func main() {
	http.HandleFunc("/", homeHandler)
	http.HandleFunc("/add-post", addPostHandler)
	http.HandleFunc("/new-post", newPostHandler)
	http.HandleFunc("/login", loginHandler)
	http.HandleFunc("/logout", logoutHandler)
	http.HandleFunc("/register", registerHandler)
	http.HandleFunc("/post", postHandler)
	http.HandleFunc("/add-comment", addCommentHandler)
	http.HandleFunc("/profile", profileHandler)
	http.HandleFunc("/categories", categoriesHandler)
	http.HandleFunc("/category-posts", categoryPostsHandler)
	http.HandleFunc("/comment-reaction", commentReactionHandler)
	fmt.Println("Server is running at http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func profileHandler(w http.ResponseWriter, r *http.Request) {
	// Retrieve the userId from the cookie
	cookie, err := r.Cookie("userId")
	if err != nil {
		if err == http.ErrNoCookie {
			// if there is no logged in user display a message and add a buttonredirect to login page
			tmpl, err := template.ParseFiles("templates/profile.html")
			if err != nil {
				log.Println("Error parsing profile template:", err)
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
				return
			}

			//  flag to indicate that the user is not logged in
			data := struct {
				IsLoggedIn bool
			}{
				IsLoggedIn: false,
			}

			if err := tmpl.Execute(w, data); err != nil {
				log.Println("Error executing profile template:", err)
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			}
			return
		}
		// For any other error, return bad request
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}

	userIdStr := cookie.Value
	userId, err := strconv.Atoi(userIdStr)
	if err != nil {
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	// Fetch user information from the database
	var user User
	err = db.QueryRow("SELECT id, username, email FROM users WHERE id = ?", userId).
		Scan(&user.ID, &user.Username, &user.Email)
	if err != nil {
		if err == sql.ErrNoRows {
			// User not found, redirect to login
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}
		log.Println("Database error:", err)
		return
	}

	tmpl, err := template.ParseFiles("templates/profile.html")
	if err != nil {
		log.Println("Error parsing profile template:", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	data := struct {
		Username   string
		Email      string
		IsLoggedIn bool
	}{
		Username:   user.Username,
		Email:      user.Email,
		IsLoggedIn: true,
	}

	if err := tmpl.Execute(w, data); err != nil {
		log.Println("Error executing profile template:", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}
