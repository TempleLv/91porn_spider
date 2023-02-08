package mailSend

import (
	"fmt"
	yaml "gopkg.in/yaml.v2"
	"io/ioutil"
	"log"
	"net/smtp"
	"strings"
)

type MailInfo struct {
	User     string `yaml:"user"`
	Password string `yaml:"password"`
	Host     string `yaml:"host"`
	Addr     string `yaml:"addr"`
}

func SendMailByYaml(subject, content, mailtype string) error {

	conf := new(MailInfo)
	yamlFile, err := ioutil.ReadFile("./save/mailConfig.yaml")
	if err != nil {
		yamlFile, err = ioutil.ReadFile("mailConfig.yaml")
	}
	err = yaml.Unmarshal(yamlFile, conf)
	// err = yaml.Unmarshal(yamlFile, &resultMap)
	if err != nil {
		log.Println("can't get mail config!!!")
		return err
	}

	hp := strings.Split(conf.Host, ":")
	auth := smtp.PlainAuth("", conf.User, conf.Password, hp[0])
	var content_type string
	if mailtype == "html" {
		content_type = "Content-Type: text/" + mailtype + "; charset=UTF-8"
	} else {
		content_type = "Content-Type: text/plain" + "; charset=UTF-8"
	}
	send_to := strings.Split(conf.Addr, ";")
	rfc822_to := strings.Join(send_to, ",")

	body := `
		<html>
		<body>
		<h3>
		"%s"
		</h3>
		</body>
		</html>
		`
	body = fmt.Sprintf(body, content)

	msg := []byte("To: " + rfc822_to + "\r\nFrom: " + conf.User + ">\r\nSubject: " + subject + "\r\n" + content_type + "\r\n\r\n" + body)

	err = smtp.SendMail(conf.Host, auth, conf.User, send_to, msg)
	return err
}
