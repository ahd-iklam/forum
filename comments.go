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

	// if reaction exists and update accordingly
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
		// insert the new reactionif there is no reaction exist
		err = insertCommentReaction(commentID, userID, action)
		return err == nil
	}

	if currentAction == action {
		// delete the reaction if it's the same
		_, err = db.Exec(`
            DELETE FROM commentreaction
            WHERE comment_id = ? AND user_id = ?`, commentID, userID)
		return err == nil
	}

	// change the reaction if it's different
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
