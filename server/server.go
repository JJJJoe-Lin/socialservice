package socialserviceserver

import (
	"database/sql"
	"encoding/json"
	"log"
	"net"
	"strings"

	"github.com/satori/go.uuid"
)

type handler func(conn net.Conn, db *sql.DB, cmdlist []string)
type srvinfo struct {
	handleFunc handler // Message handler
	Usage      string  // Command usage
	tokenPos   int     // Token position in command (-1 is no token)
	commandLen int     // Number of string token splited by space (-1 is indefinite)
}
type response struct {
	Status  int       `json:"status"`
	Message string    `json:"message,omitempty"`
	Token   string    `json:"token,omitempty"`
	Invite  *[]string `json:"invite,omitempty"`
	Friend  *[]string `json:"friend,omitempty"`
	Post    *[]post   `json:"post,omitempty"`
}
type post struct {
	ID      string `json:"id"`
	Message string `json:"message"`
}

var address *net.TCPAddr

var serviceInfos = map[string]srvinfo{
	"register":      srvinfo{registerHandler, "register <id> <password>", -1, 3},
	"login":         srvinfo{loginHandler, "login <id> <password>", -1, 3},
	"delete":        srvinfo{deleteHandler, "delete <user>", 1, 2},
	"logout":        srvinfo{logoutHandler, "logout <user>", 1, 2},
	"invite":        srvinfo{inviteHandler, "invite <user> <id>", 1, 3},
	"list-invite":   srvinfo{listinviteHandler, "list-invite <user>", 1, 2},
	"accept-invite": srvinfo{acceptinviteHandler, "accept-invite <user> <id>", 1, 3},
	"list-friend":   srvinfo{listfriendHandler, "list-friend <user>", 1, 2},
	"post":          srvinfo{postHandler, "post <user> <message>", 1, -1},
	"receive-post":  srvinfo{receivepostHandler, "receive-post <user>", 1, 2},
}

// SetAddr 設定 server 的 addr
func SetAddr(ip, port string) error {
	addr, err := net.ResolveTCPAddr("tcp", (ip + ":" + port))
	if err != nil {
		return err
	}
	address = addr
	return nil
}

// Run 啟動 server
func Run(db *sql.DB) {
	listener, err := net.ListenTCP("tcp", address)
	if err != nil {
		log.Fatalln(err)
	}
	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Panicln(err)
			continue
		}
		handleClient(conn, db)
		conn.Close()
	}
}

func handleClient(conn net.Conn, db *sql.DB) {
	defer func() {
		if x := recover(); x != nil {
			log.Println(x)
		}
	}()

	req := make([]byte, 4096)
	num, err := conn.Read(req)
	checkErrPanic(err)

	cmd := string(req[:num])
	cmdlist := strings.Split(cmd, " ")
	serviceInfo, ok := serviceInfos[cmdlist[0]]
	if ok {
		// 檢查 token 並將其替換成 username
		if pos := serviceInfo.tokenPos; pos != -1 {
			if len(cmdlist) <= pos {
				sendJSONErrorResponse(conn, "Not login yet")
				return
			}
			rows, err := db.Query("SELECT username FROM login WHERE token=?", cmdlist[pos])
			checkErrPanic(err)
			if !rows.Next() {
				sendJSONErrorResponse(conn, "Not login yet")
				checkErrPanic(rows.Close())
				return
			}
			checkErrPanic(rows.Scan(&cmdlist[pos]))
			checkErrPanic(rows.Close())
		}
		// 檢查格式
		if l := serviceInfo.commandLen; l != -1 {
			if len(cmdlist) != l {
				sendJSONErrorResponse(conn, "Usage: "+serviceInfo.Usage)
				return
			}
		}
		serviceInfo.handleFunc(conn, db, cmdlist)
		return
	}
	sendJSONErrorResponse(conn, "Unknown command "+cmdlist[0])
}

func sendJSONResponse(conn net.Conn, res response) {
	resJSONb, err := json.Marshal(res)
	checkErrPanic(err)
	_, err = conn.Write(resJSONb)
	checkErrPanic(err)
}

func sendJSONErrorResponse(conn net.Conn, errMsg string) {
	res := response{
		Status:  1,
		Message: errMsg,
	}
	sendJSONResponse(conn, res)
}

func getToken() string {
	return uuid.Must(uuid.NewV4()).String()
}

func checkErrPanic(err error) {
	if err != nil {
		panic(err)
	}
}
