package main

import (
	"net/http"
	"strconv"
)

func commentReactionHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	commentID := r.FormValue("comment_id")
	action := r.FormValue("action")
	cookie, err := r.Cookie("userId")
	if err != nil {
		http.Error(w, "User not logged in", http.StatusUnauthorized)
		return
	}
	userID, _ := strconv.Atoi(cookie.Value)

	if action != "like" && action != "dislike" {
		http.Error(w, "Invalid action", http.StatusBadRequest)
		return
	}

	// Check if reaction exists and update accordingly
	if checkCommentReaction(commentID, userID, action) {
		http.Redirect(w, r, r.Header.Get("Referer"), http.StatusSeeOther)
		return
	}

	// Insert a new reaction
	_, err = db.Exec(`
        INSERT INTO commentreaction (comment_id, user_id, action)
        VALUES (?, ?, ?)`, commentID, userID, action)
	if err != nil {
		http.Error(w, "Error saving reaction", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, r.Header.Get("Referer"), http.StatusSeeOther)
}
func checkCommentReaction(commentID string, userID int, action string) bool {
	var currentAction string
	err := db.QueryRow(`
        SELECT action
        FROM commentreaction
        WHERE comment_id = ? AND user_id = ?`, commentID, userID).Scan(&currentAction)

	if err != nil {
		// Insert a new reaction if no reaction exists
		err = insertCommentReaction(commentID, userID, action)
		return err == nil
	}

	if currentAction == action {
		// Remove the reaction if it's the same
		_, err = db.Exec(`
            DELETE FROM commentreaction
            WHERE comment_id = ? AND user_id = ?`, commentID, userID)
		return err == nil
	}

	// Update the reaction if it's different
	_, err = db.Exec(`
        UPDATE commentreaction
        SET action = ?
        WHERE comment_id = ? AND user_id = ?`, action, commentID, userID)
	return err == nil
}
func insertCommentReaction(commentID string, userID int, action string) error {
	_, err := db.Exec(`
        INSERT INTO commentreaction (comment_id, user_id, action)
        VALUES (?, ?, ?)`, commentID, userID, action)
	return err
}
