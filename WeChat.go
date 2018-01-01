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
)

const (
	TIME_WAIT  = 10
	UUID_URL   = "https://login.weixin.qq.com/jslogin"
	QRCODE_URL = "https://login.weixin.qq.com/qrcode/"
	LOGIN_URL  = "https://login.weixin.qq.com/cgi-bin/mmwebwx-bin/login"
)

var client = HttpClientBuilder().Build()

func now() string {
	return strconv.FormatInt(time.Now().Unix(), 10)
}

func UUID() (string, error) {
	param := make(map[string][]string)
	param["appid"] = []string{"wx782c26e4c19acffb"}
	param["fun"] = []string{"new"}
	param["lang"] = []string{"zh_CN"}
	param["_"] = []string{now()}
	url := BuildGetUrl(UUID_URL, param)
	req, _ := RequestBuilder().Url(url).Build()
	resp, err := client.Execute(req)
	if err == nil {
		result, _ := resp.BodyString()
		r, e := regexp.Compile("window.QRLogin.code = (\\d+);")
		arr := r.FindStringSubmatch(result)
		if e != nil {
			return "", e
		} else if (len(arr) != 2 || arr[1] != "200") {
			return "", errors.New("wechat get msg error : " + result)
		}
		r, e = regexp.Compile("window.QRLogin.uuid = \"(.*)\";")
		arr = r.FindStringSubmatch(result)
		if e != nil {
			return "", e
		} else if (len(arr) != 2) {
			return "", errors.New("wechat get msg error : " + result)
		}
		return arr[1], nil
	} else {
		return "", err
	}
}

func ShowQrCode(uuid string) {
	postbody := FormBodyBuilder().AddParam("t", "webwx").AddParam("_", now()).Build()
	req, _ := RequestBuilder().Url(QRCODE_URL + uuid).Post(postbody).Build()
	resp, err := client.Execute(req)
	if err == nil {
		bytes, _ := resp.BodyByte()
		ioutil.WriteFile("qrcode.jpg", bytes, 0644)
	} else {
		log.Println("get qrcode error:", err)
	}
}

func waitForLogin(uuid string, time4Wait time.Duration) {
	time.Sleep(time4Wait * time.Second)
	param := make(map[string][]string)
	param["tip"] = []string{"1"}
	param["uuid"] = []string{uuid}
	param["_"] = []string{now()}
	url := BuildGetUrl(LOGIN_URL, param)
	req, _ := RequestBuilder().Url(url).Build()
	resp, err := client.Execute(req)
	if err == nil {
		result, _ := resp.BodyString()
		//window.code=200;
		//window.redirect_uri="https://wx.qq.com/cgi-bin/mmwebwx-bin/webwxnewloginpage?ticket=AeXFeczuwqZ9LU_nHzfeUGuA@qrticket_0&uuid=obpi9Rft-A==&lang=zh_CN&scan=1514813932";
		log.Println(result)
	} else {
		log.Println("get qrcode error:", err)
	}
}

// 文档 https://my.oschina.net/biezhi/blog/618493
func main() {
	uuid, err := UUID()
	if (err == nil) {
		ShowQrCode(uuid)
		waitForLogin(uuid,TIME_WAIT)
	} else {
		fmt.Println(err)
	}
}
