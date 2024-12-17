package main

import (
	"database/sql"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"strconv"
	"time"
)

func homeHandler(w http.ResponseWriter, r *http.Request) {
	rows, err := db.Query(`
SELECT 
	posts.id,
    posts.title, 
    posts.created_at, 
    users.username, 
    category.name, 
    COALESCE(COUNT(comments.content), 0) AS comment_count,
    COALESCE(SUM(CASE WHEN postreaction.action = 'like' THEN 1 ELSE 0 END), 0) AS post_likes,
    COALESCE(SUM(CASE WHEN postreaction.action = 'dislike' THEN 1 ELSE 0 END), 0) AS post_dislikes
	FROM posts
	INNER JOIN users ON posts.id_users = users.id
	INNER JOIN post_category ON posts.id = post_category.post_id
	INNER JOIN category ON post_category.catego_id = category.id
	LEFT JOIN comments ON posts.id = comments.post_id
	LEFT JOIN postreaction ON postreaction.post_id = posts.id
	GROUP BY posts.id
	ORDER BY posts.created_at DESC;
    `)
	if err != nil {
		log.Println("error in querying posts:", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var posts []Post
	for rows.Next() {
		var post Post
		if err := rows.Scan(&post.ID, &post.Title, &post.CreatedAt, &post.Username, &post.CategoryName, &post.CommentCount, &post.LikeCount, &post.DislikeCount); err != nil {
			log.Println("Error scanning post:", err)
			continue
		}
		posts = append(posts, post)
	}

	if err := rows.Err(); err != nil {
		log.Println("error iterating over rows:", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	tmpl := template.Must(template.ParseFiles("templates/home.html"))
	if err := tmpl.Execute(w, posts); err != nil {
		log.Println("Error rendering template:", err)
	}
}

func newPostHandler(w http.ResponseWriter, r *http.Request) {
	tmpl := template.Must(template.ParseFiles("templates/new_post.html"))
	tmpl.Execute(w, nil)
}
func addPostHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Invalid method", http.StatusMethodNotAllowed)
		return
	}

	// Get the user ID from the cookie
	cookie, err := r.Cookie("userId")
	if err != nil {
		http.Error(w, "User not logged in", http.StatusUnauthorized)
		return
	}
	userIdStr := cookie.Value
	userID, err := strconv.Atoi(userIdStr)
	if err != nil {
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	// Retrieve form data
	title := r.FormValue("title")
	content := r.FormValue("content")
	categoryName := r.FormValue("category") // Get the category name (not ID)

	// Ensure the category name is valid (not empty)
	if categoryName == "" {
		http.Error(w, "Category is required", http.StatusBadRequest)
		return
	}

	// Begin a transaction to ensure consistency
	tx, err := db.Begin()
	if err != nil {
		http.Error(w, "Error starting transaction", http.StatusInternalServerError)
		return
	}
	defer tx.Rollback()

	// Insert the post into the posts table
	result, err := tx.Exec("INSERT INTO posts (id_users, title, content) VALUES (?, ?, ?)", userID, title, content)
	if err != nil {
		http.Error(w, "Error saving post", http.StatusInternalServerError)
		return
	}

	// Get the post ID of the newly inserted post
	postID, err := result.LastInsertId()
	if err != nil {
		http.Error(w, "Error retrieving post ID", http.StatusInternalServerError)
		return
	}

	// Insert the category if it does not exist
	var categoryID int64
	err = tx.QueryRow("SELECT id FROM category WHERE name = ?", categoryName).Scan(&categoryID)
	if err != nil {
		// If the category does not exist, create a new category
		if err == sql.ErrNoRows {
			result, err := tx.Exec("INSERT INTO category (name) VALUES (?)", categoryName)
			if err != nil {
				http.Error(w, "Error creating category", http.StatusInternalServerError)
				return
			}

			// Get the category ID of the newly inserted category
			categoryID, err = result.LastInsertId()
			if err != nil {
				http.Error(w, "Error retrieving category ID", http.StatusInternalServerError)
				return
			}
		} else {
			http.Error(w, "Error fetching category", http.StatusInternalServerError)
			return
		}
	}

	// Insert the post-category relation into the post_category table
	_, err = tx.Exec("INSERT INTO post_category (catego_id, post_id) VALUES (?, ?)", categoryID, postID)
	if err != nil {
		http.Error(w, "Error saving post category relation", http.StatusInternalServerError)
		return
	}

	// Commit the transaction
	if err := tx.Commit(); err != nil {
		http.Error(w, "Error committing transaction", http.StatusInternalServerError)
		return
	}

	// Redirect to home page after successful post creation
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func registerHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		tmpl := template.Must(template.ParseFiles("templates/register.html"))
		tmpl.Execute(w, nil)
		return
	}

	username := r.FormValue("username")
	email := r.FormValue("email")
	password := r.FormValue("password")

	// Check if the username or email already exists
	var exists bool
	err := db.QueryRow(
		"SELECT EXISTS(SELECT 1 FROM users WHERE username = ? OR email = ?)",
		username, email,
	).Scan(&exists)
	if err != nil {
		log.Println("Error checking existing user:", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	if exists {
		tmpl := template.Must(template.ParseFiles("templates/register.html"))
		errMsg := "Username or email already exists. Please choose another."
		tmpl.Execute(w, map[string]string{"ErrorMessage": errMsg})
		return
	}

	// enter the new user into the database
	_, err = db.Exec(
		"INSERT INTO users (username, email, password) VALUES (?, ?, ?)",
		username, email, password,
	)
	if err != nil {
		log.Println("Error inserting user:", err)
		// http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/login", http.StatusSeeOther)
}

func getUserId(username string) allInfos {
	var infos allInfos

	// QueryRow returns a row, we need to scan the result
	err := db.QueryRow("SELECT id FROM users WHERE username = ?", username).Scan(&infos.userId)
	if err != nil {
		// Handle error (possibly no rows found or other errors)
		return infos // return empty infos on error
	}
	fmt.Println(info.userIdAsstring, "tsting")
	// Continue processing if no error
	return infos
}

func loginHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		tmpl := template.Must(template.ParseFiles("templates/login.html"))
		tmpl.Execute(w, nil)
		return
	}

	username := r.FormValue("username")
	password := r.FormValue("password")
	if username == "" || password == "" {
		http.Error(w, "Username and password are required.", http.StatusBadRequest)
		return
	}

	var storedPassword string
	var userId string

	err := db.QueryRow("SELECT password, id FROM users WHERE username = ?", username).Scan(&storedPassword, &userId)
	if err != nil || storedPassword != password {
		http.Error(w, "Invalid username or password", http.StatusUnauthorized)
		return
	}

	cookie := &http.Cookie{
		Name:     "userId",
		Value:    userId,
		Path:     "/",                             // Cookie is valid for the entire site
		HttpOnly: true,                            // Prevents JavaScript access
		Secure:   false,                           // Set to true in production (requires HTTPS)
		Expires:  time.Now().Add(168 * time.Hour), // Cookie expires in 7 days
		SameSite: http.SameSiteLaxMode,            // Helps protect against CSRF
	}
	http.SetCookie(w, cookie)

	info = getUserId(username)
	// fmt.Println(info.userId, "this is a new user")

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func addCommentHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}
	cookie, err := r.Cookie("userId")
	if err != nil {
		// For any other error, return bad request
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}

	userIdStr := cookie.Value
	userID, err := strconv.Atoi(userIdStr)
	postID := r.FormValue("post_id")
	commentContent := r.FormValue("comment")

	if postID == "" || commentContent == "" || userID == 0 {
		http.Error(w, "Invalid comment or user", http.StatusBadRequest)
		return
	}

	_, erro := db.Exec(`
        INSERT INTO comments (content, post_id, user_id)
        VALUES (?, ?, ?)`, commentContent, postID, userID)
	if erro != nil {
		http.Error(w, "Error adding comment", http.StatusInternalServerError)
		return
	}

	// Redirect back to the post page after adding the comment
	http.Redirect(w, r, "/post?id="+postID, http.StatusFound)
}

func profileHander(w http.ResponseWriter, r *http.Request) {

}
