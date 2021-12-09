package main

import (
	"bytes"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"text/template"
	"time"

	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
	"github.com/sendgrid/sendgrid-go"
	"github.com/sendgrid/sendgrid-go/helpers/mail"
)

type Message struct {
	RecipientName  string `json:"name"`
	RecipientEmail string `json:"email"`
	Subject        string `json:"subject"`
	Title          string `json:"title"`
	MsgText        string `json:"message"`
}

func sendEmail(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		fmt.Fprintf(w, "service email is running...")
		return
	}

	decodeBody := json.NewDecoder(r.Body)
	var data Message
	err := decodeBody.Decode(&data)
	if err != nil {
		fmt.Fprintf(w, "invalid params")
		panic(err)
	}

	from := mail.NewEmail(os.Getenv("SENDGRID_API_NAME"), os.Getenv("SENDGRID_API_ADDRESS"))
	subject := data.Subject
	to := mail.NewEmail(data.RecipientName, data.RecipientEmail)

	tmp, err := template.ParseFiles("views/index.html")
	var html bytes.Buffer

	htmlData := map[string]interface{}{
		"recipientName":  data.RecipientName,
		"message":        data.MsgText,
		"subject":        data.Subject,
		"companyName":    os.Getenv("COMPANY_NAME"),
		"companyEmail":   os.Getenv("COMPANY_EMAIL"),
		"companyPhone":   os.Getenv("COMPANY_PHONE"),
		"companyAddress": os.Getenv("COMPANY_ADDRESS"),
	}

	err = tmp.Execute(&html, htmlData)
	if err != nil {
		json.NewEncoder(w).Encode(map[string]interface{}{"success": false, "message": err})
		panic(err)
	}

	message := mail.NewSingleEmail(from, subject, to, html.String(), html.String())
	client := sendgrid.NewSendClient(os.Getenv("SENDGRID_API_PASSWORD"))
	response, err := client.Send(message)

	if err != nil {
		json.NewEncoder(w).Encode(map[string]interface{}{"success": false, "message": err})
		panic(err)
	} else {
		json.NewEncoder(w).Encode(map[string]interface{}{"success": true, "message": "email sent"})
		fmt.Println(response.StatusCode)
		fmt.Println(response.Body)
		fmt.Println(response.Headers)
	}
}

func main() {
	err := godotenv.Load()
	if err != nil {
		fmt.Println("Error loading .env file")
		panic(err)
	}

	var wait time.Duration
	flag.DurationVar(&wait, "graceful-timeout", time.Second*15, "the duration for which the server gracefully wait for existing connections to finish - e.g. 15s or 1m")
	flag.Parse()
	r := mux.NewRouter()
	r.HandleFunc("/email", sendEmail).Methods("POST", "GET")

	srv := &http.Server{
		Addr:         ":" + os.Getenv("APP_PORT"),
		WriteTimeout: time.Second * 15,
		ReadTimeout:  time.Second * 15,
		IdleTimeout:  time.Second * 60,
		Handler:      r,
	}
	log.Println("server running at port: " + os.Getenv("APP_PORT"))

	go func() {
		if err := srv.ListenAndServe(); err != nil {
			panic(err)
		}
	}()

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)

	<-c

	ctx, cancel := context.WithTimeout(context.Background(), wait)
	defer cancel()
	srv.Shutdown(ctx)
	log.Println("shutting down")
	os.Exit(0)
}
