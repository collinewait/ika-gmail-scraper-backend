package scraper

import (
	"archive/zip"
	"encoding/base64"
	"errors"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/collinewait/ika-gmail-scraper/oauth"
	"google.golang.org/api/gmail/v1"
)

const userID = "me"

// create fileAddress based on current time
var fileAddress string

// Scrape will extract attachments contained in mails sent by a specific email.
func Scrape(w http.ResponseWriter, r *http.Request) {
	emailThatSentAttach := r.FormValue("emailThatSentAttach")
	fileAddress = "attachments" + strconv.FormatInt(time.Now().Unix(), 10) + ".zip"
	defer os.Remove(fileAddress)

	token, err := extractToken(r)
	if err != nil {
		fmt.Println(err.Error())
	}
	claim, err := oauth.DecodeJwtToken(token)
	if err != nil {
		errorResponse(w, err.Error())
		return //nolint
	}

	oauthTkn, oauthRtkn, expiry, err := oauth.RetrieveTokensFromSession(r, claim.RandomID)
	if err != nil {
		errorResponse(w, err.Error())
		return //nolint
	}

	service := oauth.GetGmailService(oauthTkn, oauthRtkn, expiry)

	var ms messageSevice
	var cont content
	var as attachmentService

	ms = &message{}
	cont = &messageContent{}
	as = &attachment{}

	attachErrChannel := make(chan *messageError, 1)
	doneChannel := make(chan bool)
	messagesChannel, getIDsErr := getIDs(service, emailThatSentAttach, ms)
	if len(getIDsErr) != 0 {
		err := <-getIDsErr
		msg := err.msg + " " + err.err.Error()
		errorResponse(w, msg)
		return //nolint
	}
	messageContentChannel, getMsgCErr := getMessageContent(messagesChannel, service, cont)
	if len(getMsgCErr) != 0 {
		err := <-getMsgCErr
		msg := err.msg + " " + err.err.Error()
		errorResponse(w, msg)
		return //nolint
	}
	attachmentChannel, getAttachErr := getAttachment(messageContentChannel, service, as)
	if len(getAttachErr) != 0 {
		err := <-getAttachErr
		msg := err.msg + " " + err.err.Error()
		errorResponse(w, msg)
		return //nolint
	}

	go saveAttachment(attachmentChannel, attachErrChannel, doneChannel)
	for err := range attachErrChannel {
		if err != nil {
			msg := err.msg + " " + err.err.Error()
			errorResponse(w, msg)
			return //nolint
		}
	}

	<-doneChannel

	w.Header().Set("Content-type", "application/zip")
	http.ServeFile(w, r, fileAddress)
}

func extractToken(r *http.Request) (string, error) {
	reqToken := r.Header.Get("Authorization")
	splitToken := strings.Split(reqToken, "Bearer")
	if len(splitToken) != 2 {
		return "", errors.New("Bearer token not in proper format")
	}

	reqToken = strings.TrimSpace(splitToken[1])
	return reqToken, nil
}

func errorResponse(w http.ResponseWriter, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusBadRequest)
	errMsg := fmt.Sprintf(`{"error": "%s"}`, message)
	w.Write([]byte(errMsg)) // nolint
}

type messageSevice interface {
	fetchMessages(service *gmail.Service,
		query string) (*gmail.ListMessagesResponse, error)
	fetchNextPage(service *gmail.Service,
		query string,
		NextPageToken string) (*gmail.ListMessagesResponse, error)
}
type content interface {
	getContent(
		service *gmail.Service, id string) (*gmail.Message, error)
}
type attachmentService interface {
	fetchAttachment(
		service *gmail.Service,
		msgID string, attachID string) (*gmail.MessagePartBody, error)
}

type message struct{}
type messageContent struct{}
type messageError struct {
	err error
	msg string
}
type attachment struct {
	data     string
	fileName string
}

func (m *message) fetchMessages(
	service *gmail.Service,
	query string) (*gmail.ListMessagesResponse, error) {
	r, err := service.Users.Messages.List(userID).Q(query).Do()
	return r, err
}

func (m *message) fetchNextPage(
	service *gmail.Service,
	query string,
	NextPageToken string) (*gmail.ListMessagesResponse, error) {
	r, err := service.Users.Messages.List(userID).Q(query).
		PageToken(NextPageToken).Do()
	return r, err
}

func getIDs(service *gmail.Service,
	email string,
	m messageSevice) (<-chan string, <-chan *messageError) {
	errs := new(messageError)
	errorsCh := make(chan *messageError, 1)
	defer close(errorsCh)
	query := fmt.Sprintf("from:%s", email)

	msgs := []*gmail.Message{}

	r, err := m.fetchMessages(service, query)
	if err != nil {
		msg := "Unable to retrieve Messages"
		populateErrorChan(errs, msg, err, errorsCh)
		return nil, errorsCh
	}
	msgs = append(msgs, r.Messages...)

	for len(r.NextPageToken) != 0 {
		r, err = m.fetchNextPage(service, query, r.NextPageToken)
		if err != nil {
			msg := "Unable to retrieve Messages on the next page"
			populateErrorChan(errs, msg, err, errorsCh)
			return nil, errorsCh
		}
		msgs = append(msgs, r.Messages...)
	}

	if len(r.Messages) == 0 {
		fmt.Println("No messages found.")
	}

	ids := make(chan string)

	var wg sync.WaitGroup
	for _, msg := range msgs {
		wg.Add(1)
		go func(msg *gmail.Message) {
			defer wg.Done()
			ids <- msg.Id
		}(msg)
	}

	go func() {
		wg.Wait()
		close(ids)
	}()

	return ids, nil
}

func (mc *messageContent) getContent(
	service *gmail.Service, id string) (*gmail.Message, error) {
	return service.Users.Messages.Get(userID, id).Do()
}

func getMessageContent(
	ids <-chan string,
	service *gmail.Service,
	c content) (<-chan *gmail.Message, <-chan *messageError) {
	msgCh := make(chan *gmail.Message)
	errs := new(messageError)
	errorsCh := make(chan *messageError, 1)
	var wg sync.WaitGroup
	for id := range ids {
		fmt.Println("Getting MessageContent....")
		wg.Add(1)
		go func(id string) {
			defer wg.Done()
			msgContent, err := c.getContent(service, id)
			if err != nil {
				msg := "Unable to retrieve Message Contents"
				populateErrorChan(errs, msg, err, errorsCh)
				close(errorsCh)
			}
			msgCh <- msgContent
		}(id)
	}
	go func() {
		wg.Wait()
		close(msgCh)
	}()
	return msgCh, errorsCh
}

func (a *attachment) fetchAttachment(
	service *gmail.Service, msgID string, attachID string) (*gmail.MessagePartBody, error) {
	return service.Users.Messages.Attachments.
		Get(userID, msgID, attachID).Do()
}

func getAttachment(
	msgContentCh <-chan *gmail.Message,
	service *gmail.Service,
	as attachmentService,
) (<-chan *attachment, <-chan *messageError) {
	var wg sync.WaitGroup
	attachCh := make(chan *attachment)
	errs := new(messageError)
	errorsCh := make(chan *messageError, 1)

	for msgContent := range msgContentCh {
		fmt.Println("Getting attachment....")
		wg.Add(1)
		go func(msgContent *gmail.Message) {
			defer wg.Done()
			attach := new(attachment)
			tm := time.Unix(0, msgContent.InternalDate*1e6)
			for _, part := range msgContent.Payload.Parts {
				if len(part.Filename) != 0 {
					newFileName := tm.Format("Jan-02-2006") + "-" + part.Filename
					msgPartBody, err := as.fetchAttachment(service, msgContent.Id, part.Body.AttachmentId)
					if err != nil {
						msg := "Unable to retrieve Attachment"
						populateErrorChan(errs, msg, err, errorsCh)
						close(errorsCh)
					}
					attach.data = msgPartBody.Data
					attach.fileName = newFileName
					attachCh <- attach
				}
			}
		}(msgContent)
	}
	go func() {
		wg.Wait()
		close(attachCh)
	}()

	return attachCh, errorsCh
}

func saveAttachment(
	attachCh <-chan *attachment,
	attachErrCh chan *messageError,
	doneCh chan bool,
) {

	errs := new(messageError)
	outFile, err := os.Create(fileAddress)
	if err != nil {
		msg := "Unable to create a file"
		populateErrorChan(errs, msg, err, attachErrCh)
		return
	}
	defer outFile.Close()

	zw := zip.NewWriter(outFile)

	for attach := range attachCh {
		fmt.Println("Saving attachment....")
		decoded, _ := base64.URLEncoding.DecodeString(attach.data)
		f, err := zw.Create(attach.fileName)
		if err != nil {
			msg := "Unable to create a zip writer"
			populateErrorChan(errs, msg, err, attachErrCh)
			return // nolint
		}
		if _, err := f.Write(decoded); err != nil {
			msg := "Unable to write a file to the disk"
			populateErrorChan(errs, msg, err, attachErrCh)
		}
	}

	if err := zw.Close(); err != nil {
		msg := "failed to close zip writer."
		populateErrorChan(errs, msg, err, attachErrCh)
	}

	close(attachErrCh)
	doneCh <- true
}

func populateErrorChan(
	msgErr *messageError,
	msg string,
	err error,
	errorsCh chan *messageError,
) {
	msgErr.msg = msg
	msgErr.err = err
	errorsCh <- msgErr
}
