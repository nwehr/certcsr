package main

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math/rand"
	"net/http"
	"net/smtp"
	"os"
	"os/exec"
	"time"

	"github.com/alexsasharegan/dotenv"
	// "github.com/gorilla/mux"
	// "github.com/rs/cors"
)

//  Subject: C = US, ST = Virginia, L = Roanoke, O = "Wehr Holdings, LLC", CN = Nathan Wehr, emailAddress = nathan@wehrholdings.com

type Subject struct {
	Country string `json:"country"`
	State   string `json:"state"`
	City    string `json:"city"`
	Company string `json:"company"`
	Name    string `json:"name"`
	Email   string `json:"email"`
}

func main() {
	_ = dotenv.Load()

	static := http.FileServer(http.Dir(os.Getenv("ServeRoot")))

	http.Handle("/", static)
	http.HandleFunc("/post-csr", func(w http.ResponseWriter, r *http.Request) {
		subj := Subject{}

		err := json.NewDecoder(r.Body).Decode(&subj)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		password, err := clientCert(subj)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		response := struct {
			Password string `json:"password"`
		}{password}

		json.NewEncoder(w).Encode(response)

		go func() {
			_, err := emailWithAttachment(os.Getenv("SmtpUser"), "New Certificate Request", fmt.Sprintf("%s <%s> is requesting a client cert.", subj.Name, subj.Email), "/tmp/", subj.Email+".pfx")
			if err != nil {
				fmt.Println(err)
			}
		}()

	})

	fmt.Println(http.ListenAndServe(":8080", nil))
}

func emailWithAttachment(to, subject, content, fileDir, filename string) (bool, error) {
	fileBytes, err := ioutil.ReadFile(fileDir + filename)
	if err != nil {
		return false, err
	}

	fileMIMEType := http.DetectContentType(fileBytes)
	_ = fileMIMEType

	fileData := base64.StdEncoding.EncodeToString(fileBytes)
	_ = fileData

	boundary := generatePassword()

	messageBody := []byte("Content-Type: multipart/mixed; boundary=" + boundary + " \n" +
		"MIME-Version: 1.0\n" +
		"to: " + to + "\n" +
		"subject: " + subject + "\n\n" +

		"--" + boundary + "\n" +
		"Content-Type: text/plain; charset=" + string('"') + "UTF-8" + string('"') + "\n" +
		"MIME-Version: 1.0\n" +
		"Content-Transfer-Encoding: 7bit\n\n" +
		content + "\n\n" +
		"--" + boundary + "\n" +

		"Content-Type: " + fileMIMEType + "; name=" + string('"') + filename + string('"') + " \n" +
		"MIME-Version: 1.0\n" +
		"Content-Transfer-Encoding: base64\n" +
		"Content-Disposition: attachment; filename=" + string('"') + filename + string('"') + " \n\n" +
		chunkSplit(fileData, 76, "\n") +
		"--" + boundary + "--")

	auth := smtp.PlainAuth("", os.Getenv("SmtpUser"), os.Getenv("SmtpPassword"), os.Getenv("SmtpHost"))
	err = smtp.SendMail(fmt.Sprintf("%s:%s", os.Getenv("SmtpHost"), os.Getenv("SmtpPort")), auth, os.Getenv("SmtpUser"), []string{to}, messageBody)

	if err != nil {
		return false, err
	}

	return true, nil
}

func chunkSplit(body string, limit int, end string) string {
	var charSlice []rune

	// push characters to slice
	for _, char := range body {
		charSlice = append(charSlice, char)
	}

	var result = ""

	for len(charSlice) >= 1 {
		// convert slice/array back to string
		// but insert end at specified limit
		result = result + string(charSlice[:limit]) + end

		// discard the elements that were copied over to result
		charSlice = charSlice[limit:]

		// change the limit
		// to cater for the last few words in
		if len(charSlice) < limit {
			limit = len(charSlice)
		}
	}
	return result
}

func clientCert(subj Subject) (string, error) {
	password := generatePassword()

	createKey := exec.Command("openssl", "genrsa", "-des3", "-out", "/tmp/"+subj.Email+".key", "-passout", "pass:"+password, "4096")
	out, err := createKey.CombinedOutput()
	if err != nil {
		fmt.Println(createKey.String())
		fmt.Println(string(out))
		return "", err
	}

	createCsr := exec.Command("openssl", "req", "-new", "-key", "/tmp/"+subj.Email+".key", "-out", "/tmp/"+subj.Email+".csr", "-subj", fmt.Sprintf("/C=%s/ST=%s/L=%s/O=%s/CN=%s/emailAddress=%s", subj.Country, subj.State, subj.City, subj.Company, subj.Name, subj.Email), "-passin", "pass:"+password)
	out, err = createCsr.CombinedOutput()
	if err != nil {
		fmt.Println(createCsr.String())
		fmt.Println(string(out))
		return "", err
	}

	signCsr := exec.Command("openssl", "x509", "-req", "-days", "365", "-in", "/tmp/"+subj.Email+".csr", "-CA", os.Getenv("CA"), "-CAkey", os.Getenv("CAKey"), "-set_serial", "01", "-out", "/tmp/"+subj.Email+".crt")
	out, err = signCsr.CombinedOutput()
	if err != nil {
		fmt.Println(signCsr.String())
		fmt.Println(string(out))
		return "", err
	}

	createPfx := exec.Command("openssl", "pkcs12", "-export", "-out", "/tmp/"+subj.Email+".pfx", "-inkey", "/tmp/"+subj.Email+".key", "-in", "/tmp/"+subj.Email+".crt", "-certfile", os.Getenv("CA"), "-passin", "pass:"+password, "-passout", "pass:"+password)
	out, err = createPfx.CombinedOutput()
	if err != nil {
		fmt.Println(createPfx.String())
		fmt.Println(string(out))
		return "", err
	}

	return password, nil
}

func generatePassword() string {
	rand.Seed(time.Now().UnixNano())

	chars := "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789"

	segment := func() string {
		segment := []byte{}

		for {
			segment = append(segment, chars[rand.Intn(len(chars))])

			if len(segment) == 3 {
				break
			}
		}

		return string(segment)
	}

	return fmt.Sprintf("%s-%s-%s", segment(), segment(), segment())
}
