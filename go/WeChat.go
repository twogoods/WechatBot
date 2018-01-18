package main

import (
	. "github.com/twogoods/golib/gohttp"
	"time"
	"strconv"
	"fmt"
	"regexp"
	"errors"
	"io/ioutil"
	"log"
	"strings"
	"math/rand"
	"bytes"
	"github.com/json-iterator/go"
	"os"
)

const (
	TIME_WAIT         = 5
	UUID_URL          = "https://login.weixin.qq.com/jslogin"
	QRCODE_URL        = "https://login.weixin.qq.com/qrcode/"
	LOGIN_URL         = "https://login.weixin.qq.com/cgi-bin/mmwebwx-bin/login"
	INIT_URL          = "https://%s/cgi-bin/mmwebwx-bin/webwxinit?pass_ticket=%s&skey=%s&r=%s"
	STATUS_NOTIFY_URL = "https://%s/cgi-bin/mmwebwx-bin/webwxstatusnotify?lang=zh_CN&pass_ticket=%s"
	CONTACT_URL       = "https://%s/cgi-bin/mmwebwx-bin/webwxgetcontact?pass_ticket=%s&skey=%s&r=%s"
	BATCH_CONTACT_URL = "https://%s/cgi-bin/mmwebwx-bin/webwxbatchgetcontact?type=ex&r=%s&pass_ticket=%s"
	SYNC_CHECK        = "https://%s/cgi-bin/mmwebwx-bin/synccheck"
	MSG_URL           = "https://%s/cgi-bin/mmwebwx-bin/webwxsync?sid=%s&skey=%s&pass_ticket=%s"
	SNED_MSG_URL      = "https://%s/cgi-bin/mmwebwx-bin/webwxsendmsg?pass_ticket=%s"

	TUING_BOT_URL = "http://www.tuling123.com/openapi/api"
	TUING_BOT_KEY = "*****"
)

type TulingMsg struct {
	Key    string `json:"key"`
	Info   string `json:"info"`
	Userid string `json:"userid"`
}

type TulingReply struct {
	Code int    `json:"code"`
	Text string `json:"text"`
	Url  string `json:"url"`
	List string `json:"list"`
}

type Session struct {
	uuid        string
	skey        string
	wxsid       string
	wxuin       string
	pass_ticket string
}

type BaseRequest struct {
	Uin      string
	Sid      string
	Skey     string
	DeviceID string
}

type BaseResponse struct {
	Ret    int
	ErrMsg string
}

type Key struct {
	Key int
	Val int
}

type SyncKey struct {
	Count int
	List  []Key
}

type User struct {
	Uin               int
	UserName          string
	NickName          string
	HeadImgUrl        string
	RemarkName        string
	PYInitial         string
	PYQuanPin         string
	RemarkPYInitial   string
	RemarkPYQuanPin   string
	HideInputBarFlag  int
	StarFriend        int
	Sex               int
	Signature         string
	AppAccountFlag    int
	VerifyFlag        int
	ContactFlag       int
	WebWxPluginSwitch int
	HeadImgFlag       int
	SnsFlag           int
}

type WXOrigin struct {
	BaseResponse        BaseResponse
	Count               int
	SyncKey             SyncKey
	User                User
	ChatSet             string
	SKey                string
	ClientVersion       int
	SystemTime          int
	GrayScale           int
	InviteStartCount    int
	ClickReportInterval int
}

type Room struct {
	UserName        string
	EncryChatRoomId string
}

type Member struct {
	UserName   string
	NickName   string
	RemarkName string
}

type MemberData struct {
	BaseResponse BaseResponse
	MemberCount  int
	MemberList   []Member
}
type AddMsg struct {
	FromUserName         string
	PlayLength           int
	Content              string
	StatusNotifyUserName string
	StatusNotifyCode     int
	Status               int
	VoiceLength          int
	ToUserName           string
	ForwardFlag          int
	AppMsgType           int
	Url                  string
	ImgStatus            int
	MsgType              int
	ImgHeight            int
	MediaId              string
	FileName             string
	FileSize             string
}

type AddMsgResponse struct {
	BaseResponse BaseResponse
	SyncKey      SyncKey
	ContinueFlag int
	AddMsgCount  int
	AddMsgList   []AddMsg
}

type Msg struct {
	Type         int
	Content      string
	FromUserName string
	ToUserName   string
	LocalID      string
	ClientMsgId  string
}

var session *Session
var wxSyncHost = "webpush.weixin.qq.com"
var wxHost string
var deviceID string
var cookieUrl string
var members []Member
var synckey SyncKey
var synckeyStr string
var me string
var twogoods string
var weGroup = "@@03a689f59245f2f6a4b431e291da45f49db657f1f2fe31f3c4e4b701ef3de82e"

var client = HttpClientBuilder().EnableCookie(true).Build()

func init() {
	deviceID = getDeviceId()
}

func now() string {
	return strconv.FormatInt(time.Now().Unix()*1000, 10)
}

func nowWitRandom() string {
	second := time.Now().Unix()
	rnd := rand.New(rand.NewSource(second))
	code := fmt.Sprintf("%06v", rnd.Int31n(1000000))
	return strconv.FormatInt(second, 10) + code

}

func getDeviceId() string {
	rnd := rand.New(rand.NewSource(time.Now().UnixNano()))
	pre := fmt.Sprintf("%08v", rnd.Int31n(100000000))
	end := fmt.Sprintf("%07v", rnd.Int31n(10000000))
	return "e" + pre + end
}

func match(express string, content string) (string, error) {
	r, _ := regexp.Compile(express)
	arr := r.FindStringSubmatch(content)
	if (len(arr) != 2) {
		return "", errors.New("prase content error : " + content)
	}
	return arr[1], nil
}

func UUID() (string, error) {
	param := make(map[string][]string)
	param["appid"] = []string{"wx782c26e4c19acffb"}
	param["fun"] = []string{"new"}
	param["lang"] = []string{"zh_CN"}
	param["_"] = []string{now()}
	url := BuildGetUrl(UUID_URL, param)
	req, _ := RequestBuilder().Header("User-agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_11_3) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/48.0.2564.109 Safari/537.36").Url(url).Build()
	resp, err := client.Execute(req)
	if err == nil {
		result, _ := resp.BodyString()
		r, e := match("window.QRLogin.code = (\\d+);", result)
		if e != nil {
			return "", e
		} else if (r != "200") {
			return "", errors.New("wechat get msg error : " + result)
		}
		r, e = match("window.QRLogin.uuid = \"(.*)\";", result)
		if e != nil {
			return "", e
		}
		return r, nil
	} else {
		return "", err
	}
}

func ShowQrCode(uuid string) {
	postbody := FormBodyBuilder().AddParam("t", "webwx").AddParam("_", now()).Build()
	req, _ := RequestBuilder().Header("User-agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_11_3) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/48.0.2564.109 Safari/537.36").Url(QRCODE_URL + uuid).Post(postbody).Build()
	resp, err := client.Execute(req)
	if err == nil {
		bytes, _ := resp.BodyByte()
		ioutil.WriteFile("qrcode.jpg", bytes, 0644)
	} else {
		log.Println("get qrcode error:", err)
	}
}

func waitForLogin(uuid string, time4Wait time.Duration) bool {
	time.Sleep(time4Wait * time.Second)
	param := make(map[string][]string)
	param["tip"] = []string{"1"}
	param["uuid"] = []string{uuid}
	param["_"] = []string{now()}
	url := BuildGetUrl(LOGIN_URL, param)
	req, _ := RequestBuilder().Header("User-agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_11_3) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/48.0.2564.109 Safari/537.36").Url(url).Build()
	resp, err := client.Execute(req)
	if err != nil {
		log.Println("get qrcode error:", err)

	}
	result, _ := resp.BodyString()
	//window.code=200;
	//window.redirect_uri="https://wx.qq.com/cgi-bin/mmwebwx-bin/webwxnewloginpage?ticket=AeXFeczuwqZ9LU_nHzfeUGuA@qrticket_0&uuid=obpi9Rft-A==&lang=zh_CN&scan=1514813932";

	r, e := match("window.code=(\\d+);", result)
	if (e != nil) {
		log.Println("prase response error", e)
		return false
	} else if (r != "200") {
		log.Println("login fail", result)
		return false
	}
	r, e = match("window.redirect_uri=\"(\\S+?)\";", result)
	if (e != nil) {
		log.Println("prase response error", e)
		return false
	}
	cookieUrl = r + "&fun=new"
	wxHost = strings.Split(strings.Split(r, "://")[1], "/")[0]
	setWxSyncHost()
	return true
}

func setWxSyncHost() {
	if strings.Index(wxHost, "wx2.qq.com") > -1 {
		wxSyncHost = "webpush.wx2.qq.com"
	} else if strings.Index(wxHost, "wx8.qq.com") > -1 {
		wxSyncHost = "webpush.wx8.qq.com"
	} else if strings.Index(wxHost, "qq.com") > -1 {
		wxSyncHost = "webpush.wx.qq.com"
	} else if strings.Index(wxHost, "web2.wechat.com") > -1 {
		wxSyncHost = "webpush.web2.wechat.com"
	} else if strings.Index(wxHost, "wechat.com") > -1 {
		wxSyncHost = "webpush.web.wechat.com"
	}
}

func getCookie() *Session {
	req, _ := RequestBuilder().Header("User-agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_11_3) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/48.0.2564.109 Safari/537.36").Url(cookieUrl).Build()
	resp, err := client.Execute(req)
	if err != nil {
		return nil
	}
	result, _ := resp.BodyString()
	return praseCookie(result)
}

func praseCookie(content string) *Session {
	fmt.Println("cookie: ", content)
	session := &Session{}
	flag, r := match("<ret>(\\S+)</ret>", content)
	if r != nil || flag != "0" {
		return session
	}
	skey, _ := match("<skey>(\\S+)</skey>", content)
	wxsid, _ := match("<wxsid>(\\S+)</wxsid>", content)
	wxuin, _ := match("<wxuin>(\\S+)</wxuin>", content)
	pass_ticket, _ := match("<pass_ticket>(\\S+)</pass_ticket>", content)
	session.skey = skey
	session.wxsid = wxsid
	session.wxuin = wxuin
	session.pass_ticket = pass_ticket

	fmt.Println("skey : ", skey)
	fmt.Println("sid : ", wxsid)
	fmt.Println("uin : ", wxuin)
	fmt.Println("pass_ticket : ", pass_ticket)

	return session
}

func wxInit() *WXOrigin {
	obj := make(map[string]BaseRequest)
	obj["BaseRequest"] = BaseRequest{session.wxuin, session.wxsid, session.skey, deviceID}
	jsonData, _ := jsoniter.Marshal(obj)
	jsonbody := JsonBodyBuilder().Json(jsonData).Build()
	url := fmt.Sprintf(INIT_URL, wxHost, session.pass_ticket, session.skey, now())
	req, _ := RequestBuilder().Header("User-agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_11_3) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/48.0.2564.109 Safari/537.36").Url(url).Post(jsonbody).Build()
	resp, err := client.Execute(req)
	if err == nil {
		response, _ := resp.BodyByte()
		var wXOrigin WXOrigin
		jsoniter.Unmarshal(response, &wXOrigin)
		return &wXOrigin
	} else {
		log.Println("wxinit error ", err)
	}
	return nil
}

func wxstatusnotify(user *User) {
	obj := make(map[string]interface{})
	obj["BaseRequest"] = BaseRequest{session.wxuin, session.wxsid, session.skey, deviceID}
	obj["Code"] = 3
	obj["FromUserName"] = user.UserName
	obj["ToUserName"] = user.UserName
	obj["ClientMsgId"] = time.Now().UnixNano() / 1000000
	jsonData, _ := jsoniter.Marshal(obj)
	jsonbody := JsonBodyBuilder().Json(jsonData).Build()
	url := fmt.Sprintf(STATUS_NOTIFY_URL, wxHost, session.pass_ticket)
	req, _ := RequestBuilder().Header("User-agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_11_3) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/48.0.2564.109 Safari/537.36").Url(url).Post(jsonbody).Build()
	_, err := client.Execute(req)
	if err != nil {
		log.Println("wxstatusnotify error ", err)
	}
}

func getContact() {
	obj := make(map[string]BaseRequest)
	obj["BaseRequest"] = BaseRequest{session.wxuin, session.wxsid, session.skey, deviceID}
	jsonData, _ := jsoniter.Marshal(obj)
	jsonbody := JsonBodyBuilder().Json(jsonData).Build()
	url := fmt.Sprintf(CONTACT_URL, wxHost, session.pass_ticket, session.skey, now())
	req, _ := RequestBuilder().Header("User-agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_11_3) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/48.0.2564.109 Safari/537.36").Url(url).Post(jsonbody).Build()
	resp, err := client.Execute(req)
	if err == nil {
		bytes, _ := resp.BodyByte()
		memberData := &MemberData{}
		jsoniter.Unmarshal(bytes, &memberData)
		members = memberData.MemberList
		filterContact()
	} else {
		log.Println("wxinit error ", err)
	}
}

func filterContact() {
	for _, member := range members {
		if member.NickName == "twogoods" && strings.Index(member.UserName, "@") == 0 {
			twogoods = member.UserName
			fmt.Println("me : ", member.UserName)
		}
		if member.NickName == "小目标要有，养身是第一位的！" && strings.Index(member.UserName, "@@") == 0 {
			weGroup = member.UserName
			fmt.Println("we : ", member.UserName)
		}
	}

}

func batchGetContact() {
	list := make([]Room, 5)
	for _, member := range members {
		if member.UserName != "" && strings.Index(member.UserName, "@@") == 0 {
			list = append(list, Room{member.UserName, ""})
		}
	}

	obj := make(map[string]interface{})
	obj["BaseRequest"] = BaseRequest{session.wxuin, session.wxsid, session.skey, deviceID}
	obj["Count"] = len(list)
	obj["List"] = list
	jsonData, _ := jsoniter.Marshal(obj)
	jsonbody := JsonBodyBuilder().Json(jsonData).Build()
	url := fmt.Sprintf(BATCH_CONTACT_URL, wxHost, now(), session.pass_ticket)
	req, _ := RequestBuilder().Header("User-agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_11_3) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/48.0.2564.109 Safari/537.36").Url(url).Post(jsonbody).Build()
	resp, err := client.Execute(req)
	if err == nil {
		resp.BodyString()
	} else {
		log.Println("wxinit error ", err)
	}
}

func testsyncCheck(session *Session) (int, int) {
	param := make(map[string][]string)
	param["sid"] = []string{session.wxsid}
	param["uin"] = []string{session.wxuin}
	param["skey"] = []string{session.skey}
	param["deviceid"] = []string{deviceID}
	param["synckey"] = []string{synckeyStr}
	param["_"] = []string{now()}
	param["r"] = []string{now()}

	hosts := []string{
		"webpush.wx.qq.com",
		"wx2.qq.com",
		"webpush.wx2.qq.com",
		"wx8.qq.com",
		"webpush.wx8.qq.com",
		"webpush.wx.qq.com",
		"web2.wechat.com",
		"webpush.web2.wechat.com",
		"webpush.web.wechat.com",
		"webpush.weixin.qq.com",
		"webpush.wechat.com",
		"webpush1.wechat.com",
		"webpush2.wechat.com",
		"webpush2.wx.qq.com"}

	for _, host := range hosts {
		url := fmt.Sprintf(SYNC_CHECK, host)
		url = BuildGetUrl(url, param)
		fmt.Println("syncCheck url: " + url)
		req, _ := RequestBuilder().Header("User-agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_11_3) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/48.0.2564.109 Safari/537.36").Url(url).Build()
		resp, err := client.Execute(req)
		if err == nil {
			response, _ := resp.BodyString()
			fmt.Println("syncCheck response: " + response)
			retcodeStr, _ := match("retcode:\"(\\d+)\",", response)
			retcode, e := strconv.Atoi(retcodeStr)
			if (e != nil) {
				return -1, -1
			}
			selectorStr, e := match("selector:\"(\\d+)\"}", response)
			selector, e := strconv.Atoi(selectorStr)
			if (e != nil) {
				return -1, -1
			}
			if (retcode == 0) {
				wxSyncHost = host
				return retcode, selector
			}
		} else {
			log.Println("wxinit error ", err)
		}
	}
	return -1, -1
}

func syncCheck(session *Session) (int, int) {
	param := make(map[string][]string)
	param["sid"] = []string{session.wxsid}
	param["uin"] = []string{session.wxuin}
	param["skey"] = []string{session.skey}
	param["deviceid"] = []string{deviceID}
	param["synckey"] = []string{synckeyStr}
	param["_"] = []string{now()}
	param["r"] = []string{now()}

	url := fmt.Sprintf(SYNC_CHECK, wxSyncHost)
	url = BuildGetUrl(url, param)
	fmt.Println("syncCheck url: " + url)
	req, _ := RequestBuilder().Header("User-agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_11_3) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/48.0.2564.109 Safari/537.36").Url(url).Build()
	resp, err := client.Execute(req)
	if err == nil {
		response, _ := resp.BodyString()
		fmt.Println("syncCheck response: " + response)
		retcodeStr, _ := match("retcode:\"(\\d+)\",", response)
		retcode, e := strconv.Atoi(retcodeStr)
		if (e != nil) {
			return -1, -1
		}
		selectorStr, e := match("selector:\"(\\d+)\"}", response)
		selector, e := strconv.Atoi(selectorStr)
		if (e != nil) {
			return -1, -1
		}
		return retcode, selector
	} else {
		log.Println("syncCheck error ", err)
	}
	return -1, -1
}

func getNewMessage() {
	obj := make(map[string]interface{})
	obj["BaseRequest"] = BaseRequest{session.wxuin, session.wxsid, session.skey, deviceID}
	obj["SyncKey"] = synckey
	obj["rr"] = ^time.Now().Unix() + 1
	jsonData, _ := jsoniter.Marshal(obj)
	fmt.Println(string(jsonData))
	jsonbody := JsonBodyBuilder().Json(jsonData).Build()
	url := fmt.Sprintf(MSG_URL, wxHost, session.wxsid, session.skey, session.pass_ticket)
	fmt.Println(url)
	req, _ := RequestBuilder().Header("User-agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_11_3) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/48.0.2564.109 Safari/537.36").Url(url).Post(jsonbody).Build()
	resp, err := client.Execute(req)
	if err == nil {
		bytes, _ := resp.BodyByte()
		addMsgResponse := AddMsgResponse{}
		jsoniter.Unmarshal(bytes, &addMsgResponse)
		generateSyncKey(addMsgResponse.SyncKey)
		if addMsgResponse.AddMsgCount > 0 {
			filterTxtMsg(addMsgResponse.AddMsgList)
		}
	} else {
		log.Println("getNewMessage error ", err)
	}
}

func filterTxtMsg(messages []AddMsg) {
	for _, msg := range messages {
		if msg.MsgType == 1 {
			log.Println("msg: ", msg.Content)
			log.Println("from: ", msg.FromUserName)
			log.Println("to: ", msg.ToUserName)

			if strings.Contains(msg.Content, "@twogoods") {
				content := ""
				sender := ""
				replyReceiver := msg.FromUserName
				if msg.FromUserName == me {
					replyReceiver = msg.ToUserName
				}
				if strings.Index(msg.FromUserName, "@@") == 0 {
					//群消息
					items := strings.Split(msg.Content, ":<br/>")
					sender = items[0]
					content = strings.Replace(items[1], "@twogoods", "", -1)
				} else {
					//私聊消息 或者 我发的消息
					content = strings.Replace(msg.Content, "@twogoods", "", -1)
					sender = msg.FromUserName
				}
				reply, e := getReplyFromTuling(content, sender)
				if (e == nil) {
					fmt.Println(reply, replyReceiver)
					sendTxtMsg(reply, replyReceiver)
				}
			}
		}
	}
}

func getReplyFromTuling(content string, user string) (string, error) {
	userid := user[5:37]
	param := make(map[string]string)
	param["key"] = TUING_BOT_KEY
	param["info"] = content
	param["userid"] = userid
	jsonData, _ := jsoniter.Marshal(param)
	fmt.Println(string(jsonData))
	jsonbody := JsonBodyBuilder().Json(jsonData).Build()
	req, _ := RequestBuilder().Url(TUING_BOT_URL).Post(jsonbody).Build()
	resp, err := client.Execute(req)
	if err == nil {
		bytes, _ := resp.BodyByte()
		reply := &TulingReply{}
		jsoniter.Unmarshal(bytes, &reply)
		fmt.Println(reply)
		if reply.Code == 100000 {
			return "机器人: " + reply.Text, nil
		} else if reply.Code == 200000 {
			return "机器人: " + reply.Text + "  " + reply.Url, nil
		} else {
			return "机器人: 自己人脑解析 →_→ " + reply.Url + "   " + reply.List, nil
		}
	} else {
		log.Println("getReplyFromTuling error ", err)
		return "", err
	}
}

func sendTxtMsg(content string, to string) {
	time.Sleep(5 * time.Second)
	obj := make(map[string]interface{})
	obj["BaseRequest"] = BaseRequest{session.wxuin, session.wxsid, session.skey, deviceID}
	clientMsgId := nowWitRandom()
	obj["Msg"] = Msg{1, content, me, to, clientMsgId, clientMsgId}
	jsonData, _ := jsoniter.Marshal(obj)
	jsonbody := JsonBodyBuilder().Json(jsonData).Build()
	url := fmt.Sprintf(SNED_MSG_URL, wxHost, session.pass_ticket)
	req, _ := RequestBuilder().Header("User-agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_11_3) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/48.0.2564.109 Safari/537.36").Url(url).Post(jsonbody).Build()
	resp, err := client.Execute(req)
	if err == nil {
		response, _ := resp.BodyString()
		log.Println("send msg : ", response)
	} else {
		log.Println("sendTxtMsg error ", err)
	}
}

func polling() {
	retcode, selector := testsyncCheck(session)
	if retcode == 0 {
		handleMsg(selector)
	}
	flag := 0
	for {
		if flag > 10 {
			os.Exit(-1)
		}
		retcode, selector = syncCheck(session)
		switch (retcode) {
		case 1100:
			log.Println("在手机上退出了登录", retcode, selector)
			flag++
			break
		case 1101:
			log.Println("你在其他地方登录了 WEB 版微信", retcode, selector)
			flag++
			break
		case 1102:
			log.Println("你在手机上主动退出了", retcode, selector)
			flag++
			break
		case 0:
			handleMsg(selector)
			break;
		default:
			log.Println("未知返回值", retcode, selector)
			flag++
			break
		}
	}
}

func handleMsg(selector int) {
	if selector != 2 {
		return
	}
	getNewMessage()
}

func generateSyncKey(key SyncKey) {
	synckey = key
	keyArr := key.List
	var buf bytes.Buffer
	for i := 0; i < len(keyArr); i++ {
		if i == 0 {
			buf.WriteString(strconv.Itoa(keyArr[i].Key))
			buf.WriteString("_")
			buf.WriteString(strconv.Itoa(keyArr[i].Val))
		} else {
			buf.WriteString("|")
			buf.WriteString(strconv.Itoa(keyArr[i].Key))
			buf.WriteString("_")
			buf.WriteString(strconv.Itoa(keyArr[i].Val))
		}
	}
	synckeyStr = buf.String()
}

// 文档 https://my.oschina.net/biezhi/blog/618493
func main() {
	uuid, err := UUID()
	if (err == nil) {
		ShowQrCode(uuid)
		for {
			if waitForLogin(uuid, TIME_WAIT) {
				break;
			}
		}
		session = getCookie()
		originData := wxInit()
		me = originData.User.UserName
		generateSyncKey(originData.SyncKey)
		wxstatusnotify(&originData.User)
		getContact()
		batchGetContact()
		polling()
	} else {
		fmt.Println(err)
	}
}
