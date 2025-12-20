package email

import (
	"context"
	"crypto/sha256"
	"crypto/tls"
	"fmt"
	"log"
	"net/smtp"
	"os"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

type RedisEmailVerifier struct {
	client *redis.Client
}

func NewRedisEmailVerifier(client *redis.Client) *RedisEmailVerifier {
	return &RedisEmailVerifier{client: client}
}

func (r *RedisEmailVerifier) SendVerificationMail(ctx context.Context, email string) (string, string, error) {
	host := os.Getenv("SMTP_HOST")
	port := os.Getenv("SMTP_PORT")
	from := os.Getenv("SMTP_MAIL")
	password := os.Getenv("SMTP_PASSWORD")

	verificationKey := fmt.Sprintf(
		"%x", sha256.Sum256([]byte(email + "-" + uuid.New().String())[:]),
	)
	verificationLinkBase := "http://localhost:8080/verify-email?token="
	link := fmt.Sprintf("%s%s", verificationLinkBase, verificationKey)

	// Check if SMTP is configured
	if host == "" || port == "" {
		log.Printf("SMTP not configured. Skipping email send. Verification Link: %s\n", link)
		// Save to Redis so verification still works
		err := r.client.Set(ctx, verificationKey, email, 5*time.Minute).Err()
		if err != nil {
			return "", "", err
		}
		return link, verificationKey, nil
	}

	auth := smtp.PlainAuth("", from, password, host)

	// письмо
	body := fmt.Sprintf(`
	<html>
		<a href="%v" target="_blank">CLICK</a>
		<p>This link will expire in 24 hours.</p>
	</html>
	`, link)

	// smtp
	tlsconfig := &tls.Config{
		InsecureSkipVerify: true,
		ServerName:         host,
	}
	c, err := smtp.Dial(host + ":" + port)
	if err != nil {
		return "", "", err
	}
	defer c.Quit()

	if err := c.StartTLS(tlsconfig); err != nil {
		return "", "", err
	}
	if err := c.Auth(auth); err != nil {
		return "", "", err
	}
	if err := c.Mail(from); err != nil {
		return "", "", err
	}
	if err := c.Rcpt(email); err != nil {
		return "", "", err
	}
	w, err := c.Data()
	if err != nil {
		log.Println(err.Error())
		return "", "", err
	}

	msg := fmt.Sprintf("MIME-Version: 1.0\r\nContent-type: text/html; charset=UTF-8\r\nFrom: %s\r\nTo: %s\r\nSubject: Email Verification\r\n\r\n%s", from, email, body)
	if _, err := w.Write([]byte(msg)); err != nil {
		return "", "", err
	}
	w.Close()

	// сохраняем в Redis
	err = r.client.Set(ctx, verificationKey, email, 5*time.Minute).Err()
	if err != nil {
		return "", "", err
	}

	return link, verificationKey, nil
}

func (r *RedisEmailVerifier) VerifyEmail(ctx context.Context, token string) (string, error) {
	email, err := r.client.Get(ctx, token).Result()
	if err != nil {
		return "", err
	}
	_, _ = r.client.Del(ctx, token).Result()
	return email, nil
}
