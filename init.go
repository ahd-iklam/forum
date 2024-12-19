package main

import (
	"database/sql"
	"log"
)

var db *sql.DB

func init() {
	var err error
	db, err = sql.Open("sqlite3", "./data.db")
	if err != nil {
		log.Fatal(err)
	}
	createCategoriesTable := ` 
		CREATE TABLE IF NOT EXISTS category(
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT NOT NULL,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP
	 );
	`

	_, err = db.Exec(createCategoriesTable)
	if err != nil {
		log.Fatal("Failed to create posts table:", err)
	}

	createPostCategories := ` 
	CREATE TABLE IF NOT EXISTS post_category(
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	catego_id INTEGER NOT NULL,
	post_id INTEGER NOT NULL,
  	FOREIGN KEY(catego_id) REFERENCES category(id),
   	FOREIGN KEY(post_id) REFERENCES posts(id)
);
`
	_, err = db.Exec(createPostCategories)
	if err != nil {
		log.Fatal("Failed to create posts table:", err)
	}
	createPostsTable := `
		CREATE TABLE IF NOT EXISTS posts (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		id_users INTEGER NOT NULL,
		title TEXT NOT NULL,
		content TEXT NOT NULL,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY(id_users) REFERENCES users(id)

	);`
	_, err = db.Exec(createPostsTable)
	if err != nil {
		log.Fatal("Failed to create posts table:", err)
	}

	createUsersTable := `
	CREATE TABLE IF NOT EXISTS users (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    username TEXT UNIQUE NOT NULL,
    email TEXT UNIQUE NOT NULL,
    password TEXT NOT NULL
);`
	_, err = db.Exec(createUsersTable)
	if err != nil {
		log.Fatal("Failed to create users table:", err)
	}
	createCommentsTable := `
	CREATE TABLE IF NOT EXISTS comments (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    content TEXT NOT NULL,
	post_id INTEGER NOT NULL,
	user_id INTEGER NOT NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY(post_id) REFERENCES posts(id),
	FOREIGN KEY(user_id) REFERENCES users(id)
);`
	_, err = db.Exec(createCommentsTable)
	if err != nil {
		log.Fatal("Failed to create comments table:", err)
	}
	createPostReactionTable := `
		CREATE TABLE IF NOT EXISTS postreaction (
	    id INTEGER PRIMARY KEY AUTOINCREMENT,
		 post_id INTEGER NOT NULL,
		 user_id INTEGER NOT NULL,
	    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		action TEXT NOT NULL CHECK(action IN ('like', 'dislike')) ,
	    FOREIGN KEY(post_id) REFERENCES posts(id),
		FOREIGN KEY(user_id) REFERENCES users(id)
	);`
	_, err = db.Exec(createPostReactionTable)
	if err != nil {
		log.Fatal("Failed to create lis table:", err)
	}

	createCommentReactionTable := `
	CREATE TABLE IF NOT EXISTS commentreaction (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	 comment_id INTEGER NOT NULL,
	 user_id INTEGER NOT NULL,
	created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
	action TEXT NOT NULL CHECK(action IN ('like', 'dislike')),

	FOREIGN KEY(user_id) REFERENCES users(id),
	FOREIGN KEY(comment_id) REFERENCES comments(id)
);`
	_, err = db.Exec(createCommentReactionTable)
	if err != nil {
		log.Fatal("Failed to create likes table:", err)
	}
	createSessionsTable := `
	CREATE TABLE IF NOT EXISTS sessions (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	user_id INTEGER NOT NULL,
	created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
	ended_at DATETIME ,
	FOREIGN KEY(user_id) REFERENCES users(id)
);`
	_, err = db.Exec(createSessionsTable)
	if err != nil {
		log.Fatal("Failed to create likes table:", err)
	}

}
