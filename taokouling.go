package main

import (
	"fmt"
	"sort"
	"crypto/md5"
	"crypto/hmac"
	"time"
	"net/url"
	"net/http"
	"crypto/tls"
	"io"
	"strings"
	"io/ioutil"
	"encoding/json"
)

/*  Taobao开放API基础包
*   请知晓，该包并未做任何的判断，请确保传入的参数是正确的，否则报错查不出原因请别怪我没说。
*   这个基础包可以做淘宝开放API的多数请求（至少我看到阿里大鱼，可以可以直接弄了。）
*   恩，基本上没写好XML解析，所以，XML用户请不要玩。虽然我在开头设置了format，但是你最好自己写xml解析
*   Written By Rffanlab
*
*
*/



type Client struct {
	AppKey string
	AppSecret string
	UseHttps bool
	Method string
	Sign_method string
	Format string
}


const (
	// httpURL is for HTTP REST API URL.
	httpURL string = "http://gw.api.taobao.com/router/rest"
	// httpsURL is for HTTPS REST API URL.
	httpsURL string = "https://eco.taobao.com/router/rest"
)



func main()  {
	requestParams := make(map[string]string)
	requestParams["tpwd_param"] = "{\"url\":\"\",\"text\":\"\",\"logo\":\"\" }"
	client := Client{
		AppKey:"",
		AppSecret:"",
		UseHttps:true,
		Method:"taobao.wireless.share.tpwd.create",
		Sign_method:"md5",
		Format:"json",
	}
	result,err :=client.DoRequest(requestParams)
	if err != nil{
		fmt.Println(err)
	}
	fmt.Println(result["wireless_share_tpwd_create_response"])
}


// 设置通用的Map
func (c *Client)SetCommonParams() map[string]string {
	params := make(map[string]string)
	t := time.Now()
	params["method"] = c.Method
	params["format"] = c.Format
	params["v"] = "2.0"
	params["sign_method"] = c.Sign_method
	params["app_key"] = c.AppKey
	params["timestamp"] = fmt.Sprintf("%04d-%02d-%02d %02d:%02d:%02d", t.Year(), t.Month(), t.Day(), t.Hour(), t.Minute(), t.Second())
	return params
}


// 生成全部参数
func (c *Client)SetRequestParams(requestParams,commonParams map[string]string) map[string]string {
	for k := range commonParams{
		requestParams[k] = commonParams[k]
	}
	return requestParams
}

// 给参数排序，并返回数值
func (c *Client)SortParamsToStr(params map[string]string) string {
	var keys []string
	for key := range params{
		keys = append(keys,key)
	}
	sort.Strings(keys)
	str := ""
	for _,k := range keys{
		str += k + params[k]
	}
	return str
}


// 签名算法
/*
* API算法的URL：http://open.taobao.com/docs/doc.htm?spm=a219a.7395905.0.0.hsp22E&articleId=101617&docType=1&treeId=1
*
*/
func (c *Client)SignMD5(params map[string]string) (string) {
	str := fmt.Sprintf("%s%s%s",c.AppSecret,c.SortParamsToStr(params),c.AppSecret)
	return fmt.Sprintf("%X",md5.Sum([]byte(str)))
}
/*
* HMAC 加密算法
*
*/
func (c *Client)SignHMAC(params map[string]string) (string) {
	str := c.SortParamsToStr(params)
	mac := hmac.New(md5.New,[]byte(c.AppSecret))
	mac.Write([]byte(str))
	return fmt.Sprintf("%X",mac.Sum(nil))
}

// 创建请求体

func (c *Client)MakeRequestBody(params map[string]string) (io.Reader,error) {
	values := url.Values{}
	for k,v := range params{
		values.Set(k,v)
	}
	return strings.NewReader(values.Encode()),nil
}
// 开始请求
/*
*  传入参数：一个包含所有参数的map
*  返回参数：返回一个map
*/
func (c *Client)DoRequest(params map[string]string) (map[string]interface{},error) {
	commonParams := c.SetCommonParams()
	requestParams := c.SetRequestParams(params,commonParams)
	if requestParams["sign_method"] == "md5" {
		sign := c.SignMD5(requestParams)
		requestParams["sign"] = sign
	}else if requestParams["sign_method"] == "hmac"{
		sign := c.SignHMAC(requestParams)
		requestParams["sign"] = sign
	}else {
		fmt.Errorf("签名方法配置错误")
	}
	values := url.Values{}
	for k,v := range requestParams{
		values.Set(k,v)
	}
	requestUrl := ""
	if c.UseHttps {
		requestUrl = httpsURL
	}else {
		requestUrl = httpURL
	}
	tr := &http.Transport{
		TLSClientConfig:&tls.Config{InsecureSkipVerify:true},
	}
	requestClient := &http.Client{Transport:tr}
	requestBody,err := c.MakeRequestBody(requestParams)
	if err != nil{
		return nil,err
	}
	req,err := http.NewRequest("POST",requestUrl,requestBody)
	if err != nil{
		return nil,err
	}
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	resp,err := requestClient.Do(req)
	if err != nil{
		return nil,err
	}
	defer resp.Body.Close()
	body,err := ioutil.ReadAll(resp.Body)
	if err != nil{
		return nil,err
	}
	responseStr := string(body)
	if strings.Contains(responseStr,"error"){

	}
	tbPwd := make(map[string]interface{})
	json.Unmarshal(body,&tbPwd)
	return tbPwd,nil
}






