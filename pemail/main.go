package main

import (
	"flag"
	"log"
	// FIX "time"
	gomail "gopkg.in/gomail.v2"
	// FIX textTemplate "text/template"
)

type User struct {
	EmailAddr  string
	Realname   string
	ExternalID string // ID used for Google analytics custom dimension
}

type Mail struct {
	From    string
	ToUsers []User
	ReplyTo string
	Subject string
	TxtBody string
	Files   map[string]string
}

type Client interface {
	Send(m Mail) error
}

type SMTPClient struct {
	mailer *gomail.Dialer
	//templateInfo TemplateInfo
}

var (
	smtpServer = flag.String("smtp-server", "127.0.0.1", "SMTP Server")
	smtpPort   = flag.Int("smtp-port", 25, "SMTP Port")
	smtpLogin  = flag.String("smtp-login", "", "SMTP Login")
	smtpPasswd = flag.String("smtp-passwd", "", "SMTP Password")
	devMod     = flag.Bool("dev-mode", false, "dev mode")

	fromAddr   = flag.String("from", "", "From")
	toUser     = flag.String("touser", "", "To User")
	toUserName = flag.String("toname", "", "Real Name of To User address")
	replyTo    = flag.String("replyto", "", "Reply To")
	subject    = flag.String("subject", "", "Subject")
	txtBody    = flag.String("body", "", "Body")
	//File = flag.String("filename", "", "File Path")
)

func NewSMTPClient(server string, port int, login string, passwd string) (Client, error) {
	mailer := gomail.NewPlainDialer(server, port, login, passwd)
	return SMTPClient{mailer}, nil
}

func (c SMTPClient) Send(m Mail) error {
	for _, user := range m.ToUsers {
		addr := user.EmailAddr
		name := user.Realname
		if name == "" {
			name = user.EmailAddr
		}
		err := c.smtpSend(m, addr, name, user.ExternalID)
		if err != nil {
			log.Println("ERROR: failed to send email to user=", name, " : err=", err)
		}
	}
	return nil
}

func (c SMTPClient) smtpSend(m Mail, toAddr, toName, toID string) error {

	msg := gomail.NewMessage()
	msg.SetHeader("From", m.From)
	msg.SetHeader("To", msg.FormatAddress(toAddr, toName))
	msg.SetHeader("Reply-To", m.ReplyTo)
	msg.SetHeader("Subject", m.Subject)
	msg.SetBody("text/plain", m.TxtBody)

	for _, fileName := range m.Files {
		msg.Attach(fileName)
	}

	sender, err := c.mailer.Dial()
	if err != nil {
		log.Println("ERROR: failed to dial-up email for ", toAddr, " : err=", err)
		return err
	}
	defer sender.Close()

	err = gomail.Send(sender, msg)
	if err != nil {
		log.Println("ERROR: failed to send email to ", toAddr, " : err=", err)
		return err
	}

	return nil
}

func main() {
	flag.Parse()

	mailClient, err := NewSMTPClient(*smtpServer, *smtpPort, *smtpLogin, *smtpPasswd)
	if err != nil {
		log.Fatalf("ERROR: could not load SMTP client: err=%v", err)
	}

	user := User{
		EmailAddr:  *toUser,
		Realname:   *toUserName,
		ExternalID: "todo", // ID used for Google analytics custom dimension
	}
	toUsers := make([]User, 1)
	toUsers = append(toUsers, user)
	mail := Mail{
		From:    *fromAddr,
		ToUsers: toUsers,
		ReplyTo: *replyTo,
		Subject: *subject,
		TxtBody: *txtBody,
		//Files map[string]string
	}

	mailClient.Send(mail)
}
