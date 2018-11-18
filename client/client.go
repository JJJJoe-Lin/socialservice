package socialserviceclient

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net"
	"strings"
)

type response struct {
	Status  int
	Message string
	Token   string
	Invite  []string
	Friend  []string
	Post    []post
}

type post struct {
	ID      string
	Message string
}

// TCPClient 用來與 Social Service Server 溝通（使用 TCP）
type TCPClient struct {
	addr      *net.TCPAddr
	userToken map[string]string
}

// NewTCPClient 用來建立新 TCPClient
func NewTCPClient(ip, port string) (client *TCPClient, err error) {
	client = new(TCPClient)
	addr, err := net.ResolveTCPAddr("tcp", (ip + ":" + port))
	if err != nil {
		return nil, err
	}
	client.addr = addr
	client.userToken = make(map[string]string)
	return
}

// Execute 是 TCPClient 執行指令的邏輯
func (c *TCPClient) Execute(cmd string) string {
	cmdlist := strings.Split(cmd, " ")

	var user string
	needToken := true
	for _, v := range []string{"register", "login"} {
		if cmdlist[0] == v {
			needToken = false
		}
	}
	if needToken && len(cmdlist) > 1 {
		token, ok := c.userToken[cmdlist[1]]
		if !ok {
			token = ""
		}
		user = cmdlist[1]
		cmdlist[1] = token
	}

	var resJSON response
	resByte, err := CommunicateTCP(strings.Join(cmdlist, " "), c.addr)
	checkErr(err)
	err = json.Unmarshal(resByte, &resJSON)
	checkErr(err)

	// login
	if cmdlist[0] == "login" && resJSON.Status == 0 {
		c.userToken[cmdlist[1]] = resJSON.Token
	}

	// logout and delete
	for _, v := range []string{"logout", "delete"} {
		if cmdlist[0] == v && resJSON.Status == 0 {
			delete(c.userToken, user)
		}
	}

	var str string
	if resJSON.Message != "" {
		str = resJSON.Message
	} else if resJSON.Invite != nil {
		if len(resJSON.Invite) == 0 {
			str = "No invitations"
		} else {
			str = strings.Join(resJSON.Invite, "\n")
		}
	} else if resJSON.Friend != nil {
		if len(resJSON.Friend) == 0 {
			str = "No friends"
		} else {
			str = strings.Join(resJSON.Friend, "\n")
		}
	} else if resJSON.Post != nil {
		if len(resJSON.Post) == 0 {
			str = "No posts"
		} else {
			var msglist []string
			for _, v := range resJSON.Post {
				msglist = append(msglist, (v.ID + ": " + v.Message))
			}
			str = strings.Join(msglist, "\n")
		}
	}
	return str
}

// CommunicateTCP 送出 request 並取得 response
func CommunicateTCP(cmd string, raddr *net.TCPAddr) (data []byte, err error) {
	conn, err := net.DialTCP("tcp", nil, raddr)
	if err != nil {
		return nil, err
	}
	defer conn.Close()
	conn.Write([]byte(cmd))

	data, err = ioutil.ReadAll(conn)
	if err != nil {
		return nil, err
	}
	return data, nil
}

func checkErr(err error) {
	if err != nil {
		log.Fatalln(err)
	}
}
