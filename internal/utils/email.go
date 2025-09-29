package utils

import (
	"bytes"
	"crypto/tls"
	"encoding/base64"
	"errors"
	"fmt"
	"gin-blog-server/internal/global"
	"github.com/k3a/html2text"
	"github.com/thanhpk/randstr"
	"github.com/vanng822/go-premailer/premailer"
	"gopkg.in/gomail.v2"
	"html/template"
	"io/fs"
	"log/slog"
	"path/filepath"
	"strconv"
	"strings"
)

type EmailData struct {
	URL      template.URL //验证链接
	UserName string       //用户名即邮箱地址
	Subject  string       //邮箱主题
}

// Format 将邮箱地址转换成小写，并去除空格
// 格式化邮件可以防止写错大小写重复注册，同时给用户预留犯错空间，输入空格和大小写错误也能正常处理
func Format(email string) string {
	return strings.ToLower(strings.TrimSpace(email))
}

func GenEmailVerificationInfo(email, password string) string {
	code := GetCode()
	info := Encode(email + "|" + password + "|" + code)
	return info
}

// Encode  返回生成 base64 编码
func Encode(s string) string {
	data := base64.StdEncoding.EncodeToString([]byte(s))
	return data
}

// GetCode 生成随机字符串
func GetCode() string {
	code := randstr.String(24)
	return code
}

// ParseEmailVerificationInfo 返回解析base64字符串后的 邮箱地址和code
func ParseEmailVerificationInfo(info string) (string, string, error) {
	data, err := Decode(info)
	if err != nil {
		return "", "", err
	}
	str := strings.Split(data, "|")
	if len(str) != 3 {
		return "", "", errors.New("wrong verification info format")
	}
	return str[0], str[1], nil
}

// Decode 返回解码 base64
func Decode(s string) (string, error) {
	data, err := base64.StdEncoding.DecodeString(s)
	if err != nil {
		return "", errors.New("email verify failed, decode error")
	}
	return string(data), nil
}

// GetEmailData 生成邮件数据
func GetEmailData(email string, info string) *EmailData {
	return &EmailData{
		URL:      template.URL(GetEmailVerifyURL(info)),
		UserName: email,
		Subject:  "邮箱验证",
	}
}

// GetEmailVerifyURL 获取邮件验证链接
func GetEmailVerifyURL(info string) string {
	//baseurl := global.GetConfig().Server.Port
	//if baseurl[0] == ':' {
	//	baseurl = fmt.Sprintf("localhost%s", baseurl)
	//}
	// FIXME:如果是用docker部署,则 注释上面的代码，使用下面的代码
	baseurl := "8.138.134.122:8765" //切记不需要加端口
	// 点击该链接可以触发 api/email/verify -> 进一步将账号存储到对应数据库中，完成账号注册
	return fmt.Sprintf("%s/api/email/verify?info=%s", baseurl, info)
}

// SendEmail 发送邮件
// 发送邮件需要配置邮箱服务器信息， 可以在config.yaml中配置
// 以下情况会发生错误: 1. 邮箱配置错误,smtp信息错误 2. 修改模板后,解析模板失败!
func SendEmail(email string, data *EmailData) error {
	config := global.GetConfig().Email
	from := config.From
	Pass := config.SmtpPass
	User := config.SmtpUser
	to := email
	Host := config.Host
	Port := config.Port

	slog.Info("User: " + User + "Pass " + Pass + "Host " + Host + "Port: " + strconv.Itoa(Port))

	var body bytes.Buffer
	//解析模板
	Template, err := ParseTemplateDir("./assets/templates")

	if err != nil {
		return errors.New("解析模版失败")
	}
	slog.Info("解析模版成功！")

	fmt.Println("URL:", data.URL)
	// 执行模版
	// 把html数据存储在body中， 第二个参数是模板名称， 第三个参数是模板数据（把模板中的占位符换成data数据）
	Template.ExecuteTemplate(&body, "email-verify.tpl", &data)

	//为了确保html文件在各个邮件客户端都能正常显示，把html转换成内联模式
	htmlString := body.String()
	prem, _ := premailer.NewPremailerFromString(htmlString, nil) //很多邮箱客户端对 <style> 标签支持不好，所以为了确保 HTML 邮件显示正常，通常会把 CSS 样式直接写到每个标签的 style 属性中（即“内联”样式）。
	htmlline, _ := prem.Transform()                              //执行 HTML 的“内联化”转换。

	//创建一个gomail.Message对象
	m := gomail.NewMessage()
	slog.Info("准备发送邮件")
	//设定m头
	m.SetHeader("From", from)
	m.SetHeader("To", to)
	m.SetHeader("Subject", data.Subject)
	//设定邮件内容
	m.SetBody("text/html", htmlline)
	//添加纯文本样式的内容作为备选
	m.AddAlternative("text/plain", html2text.HTML2Text(body.String()))

	//配置SMTP连接
	d := gomail.NewDialer(Host, Port, User, Pass)
	//FIXME 生产环境可更换腾讯云等方式获取证书
	d.TLSConfig = &tls.Config{InsecureSkipVerify: true} //跳过 TLS 证书验证，允许连接到使用非法或自签名证书的 SMTP 服务器
	slog.Info("smtp连接已建立")
	if err := d.DialAndSend(m); err != nil {
		return err
	}
	return nil
}

// ParseTemplateDir 解析模版目录
func ParseTemplateDir(dir string) (*template.Template, error) {
	var paths []string
	// 遍历模版目录, 把所有模版文件路径保存到paths中
	err := filepath.Walk(dir, func(path string, info fs.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			paths = append(paths, path)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return template.ParseFiles(paths...)
}
