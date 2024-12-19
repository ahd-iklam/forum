package main

import (
	"html/template"
	"log"
	"net/http"
)

// func categoryPostsHandler(w http.ResponseWriter, r *http.Request) {
// 	categoryID := r.URL.Query().Get("id")
// 	if categoryID == "" {
// 		http.Error(w, "Category ID is missing", http.StatusBadRequest)
// 		return
// 	}

// 	type PostData struct {
// 		Title        string
// 		CreatedAt    string
// 		Username     string
// 		LikeCount    int
// 		DislikeCount int
// 		CommentCount int
// 	}

// 	posts := []PostData{}

// 	query := `
//         SELECT p.title, p.created_at, u.username,
//                (SELECT COUNT(*) FROM likes WHERE post_id = p.id) AS like_count,
//                (SELECT COUNT(*) FROM dislikes WHERE post_id = p.id) AS dislike_count,
//                (SELECT COUNT(*) FROM comments WHERE post_id = p.id) AS comment_count
//         FROM posts p
//         JOIN users u ON p.id_users = u.id
//         JOIN post_category pc ON p.id = pc.post_id
//         WHERE pc.catego_id = ?
//     `

// 	rows, err := db.Query(query, categoryID)
// 	if err != nil {
// 		http.Error(w, "Unable to fetch posts", http.StatusInternalServerError)
// 		return
// 	}
// 	defer rows.Close()

// 	for rows.Next() {
// 		var post PostData
// 		if err := rows.Scan(&post.Title, &post.CreatedAt, &post.Username, &post.LikeCount, &post.DislikeCount, &post.CommentCount); err != nil {
// 			http.Error(w, "Error processing posts", http.StatusInternalServerError)
// 			return
// 		}
// 		posts = append(posts, post)
// 	}

// 	tmpl, err := template.ParseFiles("templates/category_posts.html")
// 	if err != nil {
// 		http.Error(w, "Unable to load template", http.StatusInternalServerError)
// 		return
// 	}

// 	tmpl.Execute(w, posts)
// }

func categoriesHandler(w http.ResponseWriter, r *http.Request) {
	type CategoryData struct {
		ID        int
		Name      string
		PostCount int
	}

	categories := []CategoryData{}

	rows, err := db.Query(`
        SELECT c.id, c.name, COUNT(pc.post_id) AS post_count 
        FROM category c
        LEFT JOIN post_category pc ON c.id = pc.catego_id
        GROUP BY c.id, c.name
    `)
	if err != nil {
		http.Error(w, "Unable to fetch categories", http.StatusInternalServerError)
		log.Println("Error fetching categories:", err) // Debug log
		return
	}
	defer rows.Close()

	for rows.Next() {
		var category CategoryData
		if err := rows.Scan(&category.ID, &category.Name, &category.PostCount); err != nil {
			http.Error(w, "Unable to process category data", http.StatusInternalServerError)
			log.Println("Error scanning category data:", err) // Debug log
			return
		}
		categories = append(categories, category)
	}

	tmpl, err := template.ParseFiles("templates/categories.html")
	if err != nil {
		http.Error(w, "Unable to load categories template", http.StatusInternalServerError)
		log.Println("Error loading template:", err) // Debug log
		return
	}

	tmpl.Execute(w, categories)
}

func categoryPostsHandler(w http.ResponseWriter, r *http.Request) {
	categoryID := r.URL.Query().Get("id")
	if categoryID == "" {
		http.Error(w, "Category ID is missing", http.StatusBadRequest)
		log.Println("Category ID is missing") // Debug log
		return
	}

	type PostData struct {
		ID           int
		Title        string
		CreatedAt    string
		Username     string
		LikeCount    int
		DislikeCount int
		CommentCount int
	}

	posts := []PostData{}

	query := `
        SELECT 
            p.id,
            p.title, 
            p.created_at, 
            u.username,
            COALESCE(COUNT(comments.content), 0) AS comment_count,
            COALESCE(SUM(CASE WHEN postreaction.action = 'like' THEN 1 ELSE 0 END), 0) AS post_likes,
            COALESCE(SUM(CASE WHEN postreaction.action = 'dislike' THEN 1 ELSE 0 END), 0) AS post_dislikes
        FROM posts p
        JOIN users u ON p.id_users = u.id
        JOIN post_category pc ON p.id = pc.post_id
        LEFT JOIN comments ON comments.post_id = p.id
        LEFT JOIN postreaction ON postreaction.post_id = p.id
        WHERE pc.catego_id = ?
        GROUP BY p.id, u.username
    `

	rows, err := db.Query(query, categoryID)
	if err != nil {
		http.Error(w, "Unable to fetch posts", http.StatusInternalServerError)
		log.Println("Error fetching posts:", err)
		return
	}
	defer rows.Close()

	for rows.Next() {
		var post PostData
		if err := rows.Scan(&post.ID, &post.Title, &post.CreatedAt, &post.Username, &post.CommentCount, &post.LikeCount, &post.DislikeCount); err != nil {
			http.Error(w, "Error processing posts", http.StatusInternalServerError)
			log.Println("Error scanning post data:", err)
			return
		}
		posts = append(posts, post)
	}

	tmpl, err := template.ParseFiles("templates/category_posts.html")
	if err != nil {
		http.Error(w, "Unable to load category posts template", http.StatusInternalServerError)
		log.Println("Error loading template:", err)
		return
	}

	tmpl.Execute(w, posts)
}

// func categoriesHandler(w http.ResponseWriter, r *http.Request) {
// 	type CategoryData struct {
// 		ID        int
// 		Name      string
// 		PostCount int
// 	}

// 	categories := []CategoryData{}

// 	rows, err := db.Query(`
//     SELECT c.id, c.name, COUNT(pc.post_id) AS post_count
//     FROM category c
//     LEFT JOIN post_category pc ON c.id = pc.catego_id
//     GROUP BY c.id, c.name
// `)

// 	if err != nil {
// 		http.Error(w, "Unable to fetch categories", http.StatusInternalServerError)
// 		return
// 	}
// 	defer rows.Close()

// 	for rows.Next() {
// 		var category CategoryData
// 		if err := rows.Scan(&category.ID, &category.Name, &category.PostCount); err != nil {
// 			http.Error(w, "Unable to process data", http.StatusInternalServerError)
// 			return
// 		}
// 		categories = append(categories, category)
// 	}

// 	tmpl, err := template.ParseFiles("templates/categories.html")
// 	if err != nil {
// 		http.Error(w, "Unable to load template", http.StatusInternalServerError)
// 		return
// 	}

// 	tmpl.Execute(w, categories)
// }
