package main

import (
	"html/template"
	"net/http"
)

func checkrpdb(postid string, userid int, action string) bool {
	var ac string
	ac = ""
	// On vérifie si une réaction existe déjà pour cet utilisateur et ce post
	err := db.QueryRow(`
        select action
        FROM postreaction
        WHERE post_id = ? AND user_id = ?`, postid, userid).Scan(&ac)
	if err != nil {
		// Si aucune réaction n'existe, on insère la nouvelle action
		if err = insertReaction(postid, userid, action); err != nil {
			return false
		}
		return true
	}

	// Si l'utilisateur a déjà réagi
	if ac == action {
		// Si l'action est la même, on la supprime
		_, err = db.Exec(`
            DELETE FROM postreaction
            WHERE post_id = ? AND user_id = ?`, postid, userid)
		if err != nil {
			return false
		}
		return true
	}

	// Si l'action est différente (ex: like -> dislike), on la met à jour
	if ac != "" && ac != action {
		// Mettre à jour la réaction
		_, err = db.Exec(`
            UPDATE postreaction
            SET action = ?
            WHERE post_id = ? AND user_id = ?`, action, postid, userid)
		if err != nil {
			return false
		}
		return true
	}
	return false
}

func insertReaction(postid string, userid int, action string) error {
	// Insérer une nouvelle réaction si aucune n'existe pour cet utilisateur et ce post
	_, err := db.Exec(`
        INSERT INTO postreaction (post_id, user_id, action)
        VALUES (?, ?, ?)`, postid, userid, action)
	return err
}

func postHandler(w http.ResponseWriter, r *http.Request) {
	postID := r.URL.Query().Get("id")
	if postID == "" {
		http.Error(w, "Post ID is required", http.StatusBadRequest)
		return
	}

	var post Post
	// Fetch the post details
	err := db.QueryRow(`
        SELECT 
            posts.id, 
            posts.title, 
            posts.content, 
            posts.created_at, 
            category.name AS category_name,
            users.username AS Username
        FROM posts
        LEFT JOIN post_category ON posts.id = post_category.post_id
        LEFT JOIN category ON post_category.catego_id = category.id
        INNER JOIN users ON posts.id_users = users.id
        WHERE posts.id = ?`, postID).Scan(&post.ID, &post.Title, &post.Content, &post.CreatedAt, &post.CategoryName, &post.Username)

	if err != nil {
		http.Error(w, "Post not found", http.StatusNotFound)
		return
	}

	// Fetch comments
	rows, err := db.Query(`
        SELECT comments.id, comments.content, comments.created_at, users.username,
        COALESCE((SELECT COUNT(*) FROM commentreaction WHERE commentreaction.comment_id = comments.id AND action = 'like'), 0) AS like_count,
        COALESCE((SELECT COUNT(*) FROM commentreaction WHERE commentreaction.comment_id = comments.id AND action = 'dislike'), 0) AS dislike_count
        FROM comments 
        JOIN users ON comments.user_id = users.id
        WHERE comments.post_id = ?
        ORDER BY comments.created_at ASC`, postID)

	if err != nil {
		http.Error(w, "Error fetching comments", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var comments []struct {
		ID           int
		Content      string
		CreatedAt    string
		Username     string
		LikeCount    int
		DislikeCount int
	}

	for rows.Next() {
		var comment struct {
			ID           int
			Content      string
			CreatedAt    string
			Username     string
			LikeCount    int
			DislikeCount int
		}
		if err := rows.Scan(&comment.ID, &comment.Content, &comment.CreatedAt, &comment.Username, &comment.LikeCount, &comment.DislikeCount); err != nil {
			continue
		}
		comments = append(comments, comment)
	}

	// Determine if the user is logged in
	isLoggedIn := false
	cookie, err := r.Cookie("userId")
	if err == nil && cookie != nil {
		isLoggedIn = true
	}

	data := struct {
		Post     Post
		Comments []struct {
			ID           int
			Content      string
			CreatedAt    string
			Username     string
			LikeCount    int
			DislikeCount int
		}
		IsLoggedIn bool
	}{
		Post:       post,
		Comments:   comments,
		IsLoggedIn: isLoggedIn,
	}

	tmpl := template.Must(template.ParseFiles("templates/post.html"))
	tmpl.Execute(w, data)
}
