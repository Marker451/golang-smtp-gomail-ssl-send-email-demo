package main

//抓取京东新奇页面 预约数超过设定值的商品信息，发送信息到邮箱
import (
	"crypto/tls"
	"encoding/json"
	"flag"
	"fmt"
	"gopkg.in/gomail.v2"
	"io/ioutil"
	"log"
	"net/http"
	"net/smtp"
	"regexp"
	"strings"
)

type SkuInfo struct {
	Type       string `json:"type"`
	YueEtime   string `json:"yueEtime"`
	PlusType   int    `json:"plusType"`
	D          int    `json:"d"`
	Stime      string `json:"stime"`
	PlusEtime  string `json:"plusEtime"`
	State      int    `json:"state"`
	QiangEtime string `json:"qiangEtime"`
	Sku        int    `json:"sku"`
	URL        string `json:"url"`
	IsJ        int    `json:"isJ"`
	Info       string `json:"info"`
	Category   string `json:"category"`
	IsBefore   int    `json:"isBefore"`
	Num        int    `json:"num"`
	Flag       bool   `json:"flag"`
	Etime      string `json:"etime"`
	QiangStime string `json:"qiangStime"`
	PlusD      int    `json:"plusD"`
	YueStime   string `json:"yueStime"`
	PlusStime  string `json:"plusStime"`
}

var bookNum = flag.Int("booknumber", 500000, "")

func main() {
	flag.Parse()
	getSkuInfoUrl := "https://yushou.jd.com/youshouinfo.action?sku="
	jdItemUrl := "https://item.jd.com/%d.html"
	re := regexp.MustCompile(`p-id="([0-9]*)"`)
	addrs := []string{
		"https://xinpin.jd.com/presalelist/801.html",
		"https://xinpin.jd.com/presalelist/1401.html"}
	emailMsg := ""

	for _, addr := range addrs {
		result, _ := getUrl(addr)
		matchs := re.FindAllStringSubmatch(string(result), -1)
		skus := make([]string, 0)
		for _, match := range matchs {
			skus = append(skus, match[1])
		}
		for _, sku := range skus {
			url := getSkuInfoUrl + sku
			rawInfo, err := getUrl(url)
			if err != nil {
				log.Println("geturl err ", err)
				continue
			}
			info := SkuInfo{}
			err = json.Unmarshal(rawInfo, &info)
			if err != nil {
				log.Println("unmarshal err ", err)
				continue
			}
			if info.Type != "1" {
				continue
			}
			if info.Num > *bookNum {
				itemInfoUrl := fmt.Sprintf(jdItemUrl, info.Sku)
				itemName := ""

				emailMsg += fmt.Sprintf("商品名称：%s, \n 预约截止时间：%s,\n 抢购时间: %s \n 预约人数: %d \n 商品地址: %s \n \n", itemName, info.YueEtime, info.QiangStime, info.Num, itemInfoUrl)
			}

		}
	}
	if emailMsg == "" {
		log.Println("no items")
		emailMsg = "No items today!"
	}
	for i := 0; i < 3; i++ {
		if err := SendEmailWithGomail(TO, "jd预约筛选脚本", emailMsg); err != nil {
			log.Println("send email err ", err)
		} else {
			break
		}
	}

}

//默认utf8 jd详情页是gbk 编码 不处理会乱码
func getItemName(s string) (name string) {
	re := regexp.MustCompile(`title(.{15})`)
	matchs := re.FindStringSubmatch(s)
	if len(matchs) > 1 {
		name = matchs[1]
	}
	return
}

func getUrl(addr string) (result []byte, err error) {
	resp, err := http.Get(addr)
	if err != nil {
		log.Println("http get err", err)
		return
	}
	defer resp.Body.Close()
	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Println("read resp err", err)
		return
	}
	return data, nil
}

const (
	HOST        = "smtp.163.com"
	SERVER_ADDR = "smtp.163.com:465"
	USER        = "youxileiting@163.com"  //发送邮件的邮箱
	PASSWORD    = ""               //发送邮件邮箱的密码
	TO          = "zhongjizhizun@126.com" //默认收件
)

//25端口 非加密模式
func SendEmail(to string, subject string, msg string) error {
	//auth := smtp.CRAMMD5Auth(USER, PASSWORD)
	auth := smtp.PlainAuth("", USER, PASSWORD, HOST)
	str := strings.Replace("From: "+USER+"~To: "+to+"~Subject: "+subject+"~~", "~", "\r\n", -1) + msg
	err := smtp.SendMail(
		SERVER_ADDR,
		auth,
		USER,
		[]string{to},
		[]byte(str),
	)

	return err
}
//有问题
//smtp 包支持 先建立普通链接，然后STARTTLS
//conn,_ := net.Dial()
//cli, err := smtp.NewClient(conn, HOST)
//cli.StartTLS

// 如果直接建立的tls链接  在plainAuth Auth()时会调用cli.Start() 会有一个检查cli.tls 是否为true
//如果为false 则认为这不是一个tls链接， 不安全， 直接出错返回 "unencrypted connection"
//因为没有调用过cli.StartTLS 所以cli.tls == false, 但其实这是一个tls安全链接
//so 自己封一层Auth 接口， 去掉对cli.tls的检查
//https://stackoverflow.com/questions/11065913/send-email-through-unencrypted-connection
func SendEmailSSL(to string, subject string, msg string) error {
	tlsconfig := &tls.Config{
		InsecureSkipVerify: true,
		ServerName:         HOST,
	}
	conn, err := tls.Dial("tcp", SERVER_ADDR, tlsconfig)
	if err != nil {
		log.Println("Dialing Error:", err)
		return err
	}
	cli, err := smtp.NewClient(conn, HOST)
	if err != nil {
		log.Println("new client err ", err)
		return err
	}
	auth := unencryptedAuth{
		smtp.PlainAuth("", USER, PASSWORD, HOST),
	}
	if err = cli.Auth(auth); err != nil {
		log.Println("auth err", err)
		return err
	}
	cli.Mail(USER)
	cli.Rcpt(TO)
	w, err := cli.Data()
	if err != nil {
		log.Println(err)
		return err
	}

	_, err = w.Write([]byte(msg))
	if err != nil {
		log.Println(err)
		return err
	}
	w.Close()
	return nil
}

//is ok
func SendEmailWithGomail(to string, subject string, msg string) error {
	m := gomail.NewMessage()

	m.SetHeader("From", USER)
	m.SetHeader("To", to)
	m.SetHeader("Subject", subject)
	m.SetBody("text/html", msg)

	d := gomail.NewDialer(HOST, 465, USER, PASSWORD)
	d.SSL = true
	d.Auth = unencryptedAuth{
		smtp.PlainAuth("", USER, PASSWORD, HOST),
	}

	if err := d.DialAndSend(m); err != nil {
		log.Println("send gomail err ", err)
	}
	return nil
}

type unencryptedAuth struct {
	smtp.Auth
}

func (a unencryptedAuth) Start(server *smtp.ServerInfo) (string, []byte, error) {
	s := *server
	s.TLS = true
	return a.Auth.Start(&s)
}
