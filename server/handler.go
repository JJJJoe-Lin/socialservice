package server

import (
	"database/sql"
	"net"
	"strings"
)

func registerHandler(conn net.Conn, db *sql.DB, cmdlist []string) {
	res := response{
		Status:  0,
		Message: "Success!",
	}
	user := cmdlist[1]
	passwd := cmdlist[2]

	// 檢查重複使用者
	rows, err := db.Query("SELECT * FROM user WHERE username=?", user)
	checkErrPanic(err)
	defer rows.Close()
	if rows.Next() {
		sendJSONErrorResponse(conn, user+" is already used")
		return
	}
	checkErrPanic(rows.Close())

	// 新增使用者
	_, err = db.Exec("INSERT user SET username=?, password=password(?)", user, passwd)
	checkErrPanic(err)

	sendJSONResponse(conn, res)
}

func loginHandler(conn net.Conn, db *sql.DB, cmdlist []string) {
	res := response{
		Status:  0,
		Message: "Success!",
	}
	user := cmdlist[1]
	passwd := cmdlist[2]

	// 檢查使用者名稱及密碼
	rows, err := db.Query("SELECT * FROM user WHERE username=? and password=password(?)", user, passwd)
	checkErrPanic(err)
	defer rows.Close()
	if !rows.Next() {
		sendJSONErrorResponse(conn, "No such user or password error")
		return
	}
	checkErrPanic(rows.Close())

	// 檢查並產生 token
	var token string
	rows, err = db.Query("SELECT token FROM login WHERE username=?", user)
	checkErrPanic(err)
	if !rows.Next() {
		// 新增 token
		token = getToken()
		_, err = db.Exec("INSERT login SET username=?, token=?", user, token)
		checkErrPanic(err)
	} else {
		checkErrPanic(rows.Scan(&token))
	}
	checkErrPanic(rows.Close())
	res.Token = token
	sendJSONResponse(conn, res)
}

func deleteHandler(conn net.Conn, db *sql.DB, cmdlist []string) {
	res := response{
		Status:  0,
		Message: "Success!",
	}
	user := cmdlist[1]

	// 刪除使用者
	_, err := db.Exec("DELETE FROM user WHERE username=?", user)
	checkErrPanic(err)

	// 刪除 invite
	_, err = db.Exec("DELETE FROM invite WHERE inviter=? or invitee=?", user, user)
	checkErrPanic(err)

	// 刪除 posts
	_, err = db.Exec("DELETE FROM post WHERE author=?", user)
	checkErrPanic(err)

	// 刪除 friends
	_, err = db.Exec("DELETE FROM friend WHERE user1=? or user2=?", user, user)
	checkErrPanic(err)

	// 刪除 token
	_, err = db.Exec("DELETE FROM login WHERE username=?", user)
	checkErrPanic(err)

	sendJSONResponse(conn, res)
}

func logoutHandler(conn net.Conn, db *sql.DB, cmdlist []string) {
	res := response{
		Status:  0,
		Message: "Bye!",
	}

	// 刪除 token
	_, err := db.Exec("DELETE FROM login WHERE username=?", cmdlist[1])
	checkErrPanic(err)

	sendJSONResponse(conn, res)
}

func inviteHandler(conn net.Conn, db *sql.DB, cmdlist []string) {
	res := response{
		Status:  0,
		Message: "Success!",
	}
	user := cmdlist[1]
	target := cmdlist[2]

	// 檢查對方是否為自己
	if user == target {
		sendJSONErrorResponse(conn, "You cannot invite yourself")
		return
	}

	// 檢查對方是否已存在
	rows, err := db.Query("SELECT * FROM user WHERE username=?", target)
	checkErrPanic(err)
	defer rows.Close()
	if !rows.Next() {
		sendJSONErrorResponse(conn, target+" does not exist")
		return
	}
	checkErrPanic(rows.Close())

	// 檢查是否已經是朋友
	query := `SELECT * FROM friend WHERE user1=? and user2=?
	          UNION ALL
	          SELECT * FROM friend WHERE user2=? and user1=?`
	rows, err = db.Query(query, user, target, user, target)
	checkErrPanic(err)
	if rows.Next() {
		sendJSONErrorResponse(conn, target+" is already your friend")
		return
	}
	checkErrPanic(rows.Close())

	// 檢查是否已邀請
	rows, err = db.Query("SELECT * FROM invite WHERE inviter=? and invitee=?", user, target)
	checkErrPanic(err)
	if rows.Next() {
		sendJSONErrorResponse(conn, "Already invited")
		return
	}
	checkErrPanic(rows.Close())

	// 檢查是否已被邀請
	rows, err = db.Query("SELECT * FROM invite WHERE inviter=? and invitee=?", target, user)
	checkErrPanic(err)
	if rows.Next() {
		sendJSONErrorResponse(conn, target+" has invited you")
		return
	}
	checkErrPanic(rows.Close())

	// 新增邀請資料
	_, err = db.Exec("INSERT invite SET inviter=?,invitee=?", user, target)
	checkErrPanic(err)

	sendJSONResponse(conn, res)
}

func listinviteHandler(conn net.Conn, db *sql.DB, cmdlist []string) {
	res := response{
		Status: 0,
	}
	invite := make([]string, 0, 0)
	user := cmdlist[1]

	// 取得 invite 資料
	rows, err := db.Query("SELECT inviter FROM invite WHERE invitee=?", user)
	checkErrPanic(err)
	defer rows.Close()
	for rows.Next() {
		var inviter string
		checkErrPanic(rows.Scan(&inviter))
		invite = append(invite, inviter)
	}
	checkErrPanic(rows.Close())
	res.Invite = &invite
	sendJSONResponse(conn, res)
}

func acceptinviteHandler(conn net.Conn, db *sql.DB, cmdlist []string) {
	res := response{
		Status:  0,
		Message: "Success!",
	}
	user := cmdlist[1]
	target := cmdlist[2]

	// 檢查是否被邀請
	rows, err := db.Query("SELECT * FROM invite WHERE inviter=? and invitee=?", target, user)
	checkErrPanic(err)
	defer rows.Close()
	if !rows.Next() {
		sendJSONErrorResponse(conn, target+" did not invite you")
		return
	}
	checkErrPanic(rows.Close())

	// 新增朋友資料
	_, err = db.Exec("INSERT friend SET user1=?, user2=?", user, target)
	checkErrPanic(err)

	// 刪除邀請資料
	_, err = db.Exec("DELETE FROM invite WHERE inviter=? and invitee=?", target, user)
	checkErrPanic(err)

	sendJSONResponse(conn, res)
}

func listfriendHandler(conn net.Conn, db *sql.DB, cmdlist []string) {
	res := response{
		Status: 0,
	}
	friends := make([]string, 0, 0)
	user := cmdlist[1]

	// 取得 friends 資料
	query := `SELECT user2 as user FROM friend WHERE user1=?
	          UNION ALL
	          SELECT user1 as user FROM friend WHERE user2=?`
	rows, err := db.Query(query, user, user)
	checkErrPanic(err)
	defer rows.Close()
	for rows.Next() {
		var friend string
		checkErrPanic(rows.Scan(&friend))
		friends = append(friends, friend)
	}
	checkErrPanic(rows.Close())
	res.Friend = &friends
	sendJSONResponse(conn, res)
}

func postHandler(conn net.Conn, db *sql.DB, cmdlist []string) {
	res := response{
		Status:  0,
		Message: "Success!",
	}
	user := cmdlist[1]

	// 檢查格式
	if len(cmdlist) < 3 {
		sendJSONErrorResponse(conn, "Usage: post <user> <message>")
		return
	}

	// 新增 post 資料
	_, err := db.Exec("INSERT post SET author=?, message=?", user, strings.Join(cmdlist[2:], " "))
	checkErrPanic(err)

	sendJSONResponse(conn, res)
}

func receivepostHandler(conn net.Conn, db *sql.DB, cmdlist []string) {
	res := response{
		Status: 0,
	}
	posts := make([]post, 0, 0)
	user := cmdlist[1]

	// 取得 post 資料
	query := `SELECT author, message FROM post 
			  WHERE author IN ( SELECT user2 as user FROM friend WHERE user1=?
	                            UNION ALL
								SELECT user1 as user FROM friend WHERE user2=?)`
	rows, err := db.Query(query, user, user)
	checkErrPanic(err)
	defer rows.Close()
	for rows.Next() {
		var p post
		checkErrPanic(rows.Scan(&p.ID, &p.Message))
		posts = append(posts, p)
	}
	checkErrPanic(rows.Close())
	res.Post = &posts
	sendJSONResponse(conn, res)
}
