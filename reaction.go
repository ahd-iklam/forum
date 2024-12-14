package main

import (
	"database/sql"
	"net/http"
	"strconv"
)

func likePostHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	// Extract post_id and user_id from the form
	postID := r.FormValue("post_id")
	userID := r.FormValue("user_id")
	action := r.FormValue("action") // action should be "like"

	if postID == "" || userID == "" || action != "like" {
		http.Error(w, "Invalid form data", http.StatusBadRequest)
		return
	}

	// Convert userID and postID to integer (if necessary)
	postIDInt, err := strconv.Atoi(postID)
	if err != nil {
		http.Error(w, "Invalid post ID", http.StatusBadRequest)
		return
	}

	userIDInt, err := strconv.Atoi(userID)
	if err != nil {
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	// Check if the user has already reacted (either like or dislike)
	var existingReaction string
	err = db.QueryRow(`
        SELECT action FROM postreaction WHERE post_id = ? AND user_id = ?`, postIDInt, userIDInt).Scan(&existingReaction)

	if err == nil {
		// If a reaction exists, we can update it
		_, err := db.Exec(`
            UPDATE postreaction SET action = ? WHERE post_id = ? AND user_id = ?`, action, postIDInt, userIDInt)
		if err != nil {
			http.Error(w, "Failed to update reaction", http.StatusInternalServerError)
			return
		}
	} else if err == sql.ErrNoRows {
		// If no reaction exists, insert a new record
		_, err := db.Exec(`
            INSERT INTO postreaction (post_id, user_id, action) 
            VALUES (?, ?, ?)`, postIDInt, userIDInt, action)
		if err != nil {
			http.Error(w, "Failed to add reaction", http.StatusInternalServerError)
			return
		}
	} else {
		http.Error(w, "Error checking existing reaction", http.StatusInternalServerError)
		return
	}

	// Redirect to the same page (reload)
	http.Redirect(w, r, r.Referer(), http.StatusSeeOther)
}

func dislikePostHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	// Extract post_id and user_id from the form
	postID := r.FormValue("post_id")
	userID := r.FormValue("user_id")
	action := r.FormValue("action") // action should be "dislike"

	if postID == "" || userID == "" || action != "dislike" {
		http.Error(w, "Invalid form data", http.StatusBadRequest)
		return
	}

	// Convert userID and postID to integer (if necessary)
	postIDInt, err := strconv.Atoi(postID)
	if err != nil {
		http.Error(w, "Invalid post ID", http.StatusBadRequest)
		return
	}

	userIDInt, err := strconv.Atoi(userID)
	if err != nil {
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	// Check if the user has already reacted (either like or dislike)
	var existingReaction string
	err = db.QueryRow(`
        SELECT action FROM postreaction WHERE post_id = ? AND user_id = ?`, postIDInt, userIDInt).Scan(&existingReaction)

	if err == nil {
		// If a reaction exists, we can update it
		_, err := db.Exec(`
            UPDATE postreaction SET action = ? WHERE post_id = ? AND user_id = ?`, action, postIDInt, userIDInt)
		if err != nil {
			http.Error(w, "Failed to update reaction", http.StatusInternalServerError)
			return
		}
	} else if err == sql.ErrNoRows {
		// If no reaction exists, insert a new record
		_, err := db.Exec(`
            INSERT INTO postreaction (post_id, user_id, action) 
            VALUES (?, ?, ?)`, postIDInt, userIDInt, action)
		if err != nil {
			http.Error(w, "Failed to add reaction", http.StatusInternalServerError)
			return
		}
	} else {
		http.Error(w, "Error checking existing reaction", http.StatusInternalServerError)
		return
	}

	// Redirect to the same page (reload)
	http.Redirect(w, r, r.Referer(), http.StatusSeeOther)
}
