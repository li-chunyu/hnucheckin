package main

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	urllib "net/url"
	"strings"
)

// TODO(wangmengting), error handling

var Code = flag.String("code", "", "stu number")
var Password = flag.String("passwd", "", "pass word")
var FstAddr = flag.String("first address", "", "first address")
var SecAddr = flag.String("second adddress", "", " second adress")
var ThiAddr = flag.String("Third address", "", "Third address")
var FourthAddr = flag.String("fourth address", "", "fourth adress")

const (
	BD_CLIENT_ID = "uWuLMhwrtNygkORwAD9Y9bX1"
	BD_SECRET_ID = "3eiPO1GWD9hGhpyFvHs78jhmdcurSdfT"
)

type LoginData struct {
	Code               string `json:"Code"`
	Password           string `json:"Password"`
	Token              string `json:"Token"`
	VerCode            string `json:"VerCode"`
	WechatUserinfoCode string `json:"WechatUserinfoCode"`
}

// simply call baidu API
func ocr(image []byte) string {
	// Get baidu API acess Token
	clientID := BD_CLIENT_ID
	client_secret := BD_SECRET_ID
	// TODO(wangmengting), token存在一个文件里面，检查文件创建时间
	urlTpl := "https://aip.baidubce.com/oauth/2.0/token?grant_type=client_credentials&client_id=%v&client_secret=%v"
	acessUrl := fmt.Sprintf(urlTpl, clientID, client_secret)
	resp, _ := http.Get(acessUrl)
	acessTokenPlayload, _ := ioutil.ReadAll(resp.Body)
	var acessTokenData map[string]interface{}
	_ = json.Unmarshal(acessTokenPlayload, &acessTokenData)
	acessToken := acessTokenData["access_token"].(string)

	// Do real ocr shit
	ocrUrl := "https://aip.baidubce.com/rest/2.0/ocr/v1/numbers?access_token=" + acessToken
	encoded := base64.StdEncoding.EncodeToString(image)
	encoded = urllib.QueryEscape(encoded)
	ocrBody := "image=" + encoded
	httpClient := &http.Client{}
	ocrReq, _ := http.NewRequest("POST", ocrUrl, bytes.NewReader([]byte(ocrBody)))
	ocrResp, _ := httpClient.Do(ocrReq)

	ocrRespBody, _ := ioutil.ReadAll(ocrResp.Body)
	var ocrResult map[string]interface{}
	json.Unmarshal(ocrRespBody, &ocrResult)
	words := ocrResult["words_result"].([]interface{})[0].(map[string]interface{})["words"].(string)
	return words
}

// TODO(wangmengting), 登陆失败 重复尝试
func login() ([]*http.Cookie, bool) {
	httpClient := &http.Client{}

	// step1 Get vcode token
	getImageVcodeUrl := "https://fangkong.hnu.edu.cn/api/v1/account/getimgvcode"
	getimagevcodeReq, _ := http.NewRequest("GET", getImageVcodeUrl, nil)
	getimagevcodeReq.Header.Add("User-Agent",
		"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/88.0.4324.146 Safari/537.36")
	getimagevcodeResp, _ := httpClient.Do(getimagevcodeReq)
	getimagevcodeBody, _ := ioutil.ReadAll(getimagevcodeResp.Body)
	var getimagevcodeData map[string]interface{}
	json.Unmarshal(getimagevcodeBody, &getimagevcodeData)
	vcodeToken := getimagevcodeData["data"].(map[string]interface{})["Token"].(string)

	// step2 Get vcode pic
	getimageUrl := "https://fangkong.hnu.edu.cn/imagevcode?token=" + vcodeToken
	getimageReq, _ := http.NewRequest("GET", getimageUrl, nil)
	getimageReq.Header.Add("User-Agent",
		"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/88.0.4324.146 Safari/537.36")
	getimageResp, _ := httpClient.Do(getimageReq)
	imageData, _ := ioutil.ReadAll(getimageResp.Body)
	vcode := ocr(imageData)

	// step3 login
	loginUrl := "https://fangkong.hnu.edu.cn/api/v1/account/login"
	loginData := LoginData{
		Code:               "S1810W0721",
		Password:           "Lichunyu521",
		Token:              vcodeToken,
		VerCode:            vcode,
		WechatUserinfoCode: "",
	}
	loginPlayLoad, _ := json.Marshal(loginData)
	loginReq, _ := http.NewRequest("POST", loginUrl, bytes.NewReader(loginPlayLoad))
	loginReq.Header.Add("User-Agent",
		"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/88.0.4324.146 Safari/537.36")
	loginReq.Header.Add("Content-Type", "application/json;charset=UTF-8")
	loginResp, _ := httpClient.Do(loginReq)
	loginRespBody, _ := ioutil.ReadAll(loginResp.Body)
	// TODO(wangmengting), 处理验证码失败的情况
	fmt.Println(string(loginRespBody))
	cookies := loginResp.Cookies()
	cookieMap := make(map[string]*http.Cookie, 0)

	// remove duplications. Since login set cookie .ASPXAUTH 3 times.
	for _, cookie := range cookies {
		cookieMap[cookie.Name] = cookie
	}
	retCookies := make([]*http.Cookie, 0, 0)
	for _, v := range cookieMap {
		retCookies = append(retCookies, v)
	}
	return retCookies, true
}

func cookieToStr(cookies []*http.Cookie) string {
	cookiesStrs := make([]string, 0, 0)
	for _, c := range cookies {
		cookiesStrs = append(cookiesStrs, c.Name+"="+c.Value)
	}
	return strings.Join(cookiesStrs, "; ")
}

func add(cookies []*http.Cookie) {
	httpClient := &http.Client{}
	addUrl := "https://fangkong.hnu.edu.cn/api/v1/clockinlog/add"
	addJsonTpl := `{"Temperature":null,"RealProvince":"%v","RealCity":"%v","RealCounty":"%v","RealAddress":"%v","IsUnusual":"0","UnusualInfo":"","IsTouch":"0","IsInsulated":"0","IsSuspected":"0","IsDiagnosis":"0","tripinfolist":[{"aTripDate":"","FromAdr":"","ToAdr":"","Number":"","trippersoninfolist":[]}],"toucherinfolist":[],"dailyinfo":{"IsVia":"0","DateTrip":""},"IsInCampus":"0","IsViaHuBei":"0","IsViaWuHan":"0","InsulatedAddress":"","TouchInfo":"","IsNormalTemperature":"1","Longitude":null,"Latitude":null}`
	addJson := fmt.Sprintf(addJsonTpl, "吉林省", "延边朝鲜族自治州", "延吉市", "1101")

	cookieStr := cookieToStr(cookies)
	addReq, _ := http.NewRequest("POST", addUrl, bytes.NewReader([]byte(addJson)))
	addReq.Header.Add("User-Agent",
		"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/88.0.4324.104 Safari/537.36")
	addReq.Header.Add("Content-Type", "application/json;charset=UTF-8")
	addReq.Header.Add("Cookie", cookieStr)

	addResp, _ := httpClient.Do(addReq)
	addRespBody, _ := ioutil.ReadAll(addResp.Body)
	fmt.Println(string(addRespBody))
}

func main() {
	cookies, _ := login()
	add(cookies)
	// check()
	// f, _ := os.Open("imagevcode.jpg")
	// data, _ := ioutil.ReadAll(f)
	// fmt.Println(ocr(data))
}
