package scraper

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"reflect"
	"sort"
	"testing"
	"time"

	"google.golang.org/api/gmail/v1"
)

func Test_extractToken_shouldReturnToken(t *testing.T) {

	expectedToken := "sometokenhere"

	var bearer = "Bearer " + expectedToken
	r := httptest.NewRequest(http.MethodGet, "/urlhere", nil)
	r.Header.Add("Authorization", bearer)

	token, _ := extractToken(r)
	if token != expectedToken {
		t.Errorf("extractToken() = %v, want %v", token, expectedToken)
	}
}

func Test_extractToken_shouldReturnError(t *testing.T) {

	expectedError := "Bearer token not in proper format"

	var bearer = "sometokenhere"
	r := httptest.NewRequest(http.MethodGet, "/urlhere", nil)
	r.Header.Add("Authorization", bearer)

	_, err := extractToken(r)
	if err.Error() != expectedError {
		t.Errorf("extractToken() = %v, want %v", err.Error(), expectedError)
	}
}

type mockMessageSevice interface {
	fetchMessages(
		service *gmail.Service,
		query string) (*gmail.ListMessagesResponse, error)
	fetchNextPage(service *gmail.Service,
		query string,
		NextPageToken string) (*gmail.ListMessagesResponse, error)
}

type mockMessage struct {
}

func (m *mockMessage) fetchMessages(
	service *gmail.Service,
	query string) (*gmail.ListMessagesResponse, error) {
	gm := []*gmail.Message{
		{Id: "16c2"},
		{Id: "41ff9"},
		{Id: "41hfi"},
		{Id: "fgb"},
		{Id: "ifgh9"},
	}

	r := gmail.ListMessagesResponse{
		Messages: gm,
	}
	return &r, nil
}
func (m *mockMessage) fetchNextPage(
	service *gmail.Service,
	query string,
	NextPageToken string) (*gmail.ListMessagesResponse, error) {
	r := gmail.ListMessagesResponse{
		Messages: []*gmail.Message{},
	}
	return &r, nil
}

func Test_getIDsCanReturnIDsWithoutNextPageToken(t *testing.T) {
	tests := []struct {
		name string
		want []string
	}{
		{
			name: "test case without NextPageToken",
			want: []string{"16c2", "41ff9", "41hfi", "fgb", "ifgh9"},
		},
	}
	service := new(gmail.Service)
	testmail := "test@mail.com"
	var ms mockMessageSevice = &mockMessage{}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cgot, _ := getIDs(service, testmail, ms)
			var got []string

			for i := range cgot {
				got = append(got, i)
			}

			sort.Strings(sort.StringSlice(got))
			sort.Strings(sort.StringSlice(tt.want))

			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("getIDs() = %v, want %v", got, tt.want)
			}
		})
	}
}

type mockMessageWithNextPage struct {
}

func (m *mockMessageWithNextPage) fetchMessages(
	service *gmail.Service,
	query string) (*gmail.ListMessagesResponse, error) {
	gm := []*gmail.Message{
		{Id: "16c2"},
		{Id: "41ff9"},
		{Id: "41hfi"},
		{Id: "fgb"},
		{Id: "ifgh9"},
	}

	r := gmail.ListMessagesResponse{
		Messages:      gm,
		NextPageToken: "someToken",
	}
	return &r, nil
}

func (m *mockMessageWithNextPage) fetchNextPage(
	service *gmail.Service,
	query string,
	NextPageToken string) (*gmail.ListMessagesResponse, error) {
	gm := []*gmail.Message{
		{Id: "fgbmm"},
	}

	r := gmail.ListMessagesResponse{
		Messages: gm,
	}
	return &r, nil
}

func Test_getIDsCanReturnIDsWithNextPageToken(t *testing.T) {
	tests := []struct {
		name string
		want []string
	}{
		{
			name: "test case with NextPageToken",
			want: []string{"16c2", "41ff9", "41hfi", "fgb", "ifgh9", "fgbmm"},
		},
	}
	service := new(gmail.Service)
	testmail := "test@mail.com"
	var ms mockMessageSevice = &mockMessageWithNextPage{}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cgot, _ := getIDs(service, testmail, ms)
			var got []string

			for i := range cgot {
				got = append(got, i)
			}

			sort.Strings(sort.StringSlice(got))
			sort.Strings(sort.StringSlice(tt.want))

			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("getIDs() = %v, want %v", got, tt.want)
			}
		})
	}
}

type mockMessageWithFetchMessagesError struct {
}

func (m *mockMessageWithFetchMessagesError) fetchMessages(
	service *gmail.Service,
	query string) (*gmail.ListMessagesResponse, error) {
	r := gmail.ListMessagesResponse{
		Messages: []*gmail.Message{},
	}
	return &r, errors.New("Couldn't fetch messages")
}

func (m *mockMessageWithFetchMessagesError) fetchNextPage(
	service *gmail.Service,
	query string,
	NextPageToken string) (*gmail.ListMessagesResponse, error) {

	r := gmail.ListMessagesResponse{
		Messages: []*gmail.Message{},
	}
	return &r, nil
}

func Test_getIDsShouldReturnErrorsReturnedByFetchMessages(t *testing.T) {
	service := new(gmail.Service)
	testmail := "test@mail.com"
	var ms mockMessageSevice = &mockMessageWithFetchMessagesError{}

	_, err := getIDs(service, testmail, ms)
	expected := "Unable to retrieve Messages"
	for e := range err {
		if e.msg != expected {
			t.Errorf("getIDs() = %v, want %v", e.msg, expected)
		}
	}

}

type mockMessageWithFetchNextPageError struct {
}

func (m *mockMessageWithFetchNextPageError) fetchMessages(
	service *gmail.Service,
	query string) (*gmail.ListMessagesResponse, error) {
	r := gmail.ListMessagesResponse{
		Messages:      []*gmail.Message{{Id: "16c2"}},
		NextPageToken: "someTokenHere",
	}
	return &r, nil
}

func (m *mockMessageWithFetchNextPageError) fetchNextPage(
	service *gmail.Service,
	query string,
	NextPageToken string) (*gmail.ListMessagesResponse, error) {

	r := gmail.ListMessagesResponse{
		Messages: []*gmail.Message{},
	}
	return &r, errors.New("Couldn't fetch messages on next page")
}

func Test_getIDsShouldReturnErrorsReturnedByFetchNextPage(t *testing.T) {
	service := new(gmail.Service)
	testmail := "test@mail.com"
	var ms mockMessageSevice = &mockMessageWithFetchNextPageError{}

	_, err := getIDs(service, testmail, ms)
	expected := "Unable to retrieve Messages on the next page"
	for e := range err {
		if e.msg != expected {
			t.Errorf("getIDs() = %v, want %v", e.msg, expected)
		}
	}

}

type mockMessageWithoutMessages struct {
}

func (m *mockMessageWithoutMessages) fetchMessages(
	service *gmail.Service,
	query string) (*gmail.ListMessagesResponse, error) {
	r := gmail.ListMessagesResponse{
		Messages: []*gmail.Message{},
	}
	return &r, nil
}

func (m *mockMessageWithoutMessages) fetchNextPage(
	service *gmail.Service,
	query string,
	NextPageToken string) (*gmail.ListMessagesResponse, error) {

	r := gmail.ListMessagesResponse{
		Messages: []*gmail.Message{},
	}
	return &r, nil
}

func Test_getIDsWithoutMessages(t *testing.T) {
	service := new(gmail.Service)
	testmail := "test@mail.com"
	var ms mockMessageSevice = &mockMessageWithoutMessages{}

	msgs, _ := getIDs(service, testmail, ms)
	if len(msgs) != 0 {
		t.Errorf("getIDs() = %v, want %v", len(msgs), 0)
	}
}

type mockMessageContent struct {
}

type mockContent interface {
	getContent(
		service *gmail.Service, id string) (*gmail.Message, error)
}

func generateIds() <-chan string {
	idsCh := make(chan string, 1)
	ids := []string{"someId"}
	defer close(idsCh)
	for _, v := range ids {
		idsCh <- v
	}

	return idsCh
}

func (m *mockMessageContent) getContent(
	service *gmail.Service, id string) (*gmail.Message, error) {
	gm := gmail.Message{
		Payload: &gmail.MessagePart{
			Filename: "somename.pdf",
		},
	}
	return &gm, nil
}

func Test_getMessageContent(t *testing.T) {
	service := new(gmail.Service)
	var mc mockContent = &mockMessageContent{}
	msgCh := generateIds()
	filename := "somename.pdf"

	msgs, _ := getMessageContent(msgCh, service, mc)

	for m := range msgs {
		if m.Payload.Filename != filename {
			t.Errorf("getMessageContent() = %v, want %v", m.Payload.Filename, filename)
		}
	}
}

type mockMessageContentWithGetContentError struct {
}

func (m *mockMessageContentWithGetContentError) getContent(
	service *gmail.Service, id string) (*gmail.Message, error) {
	gm := gmail.Message{
		Payload: &gmail.MessagePart{
			Filename: "",
		},
	}
	return &gm, errors.New("Error when getting contents")
}

func Test_getMessageContentWithGetContentError(t *testing.T) {
	service := new(gmail.Service)
	var mc mockContent = &mockMessageContentWithGetContentError{}
	msgCh := generateIds()

	_, err := getMessageContent(msgCh, service, mc)

	expected := "Unable to retrieve Message Contents"
	for e := range err {
		if e.msg != expected {
			t.Errorf("getMessageContent() = %v, want %v", e.msg, expected)
		}
	}
}

type mockAttachment struct {
}

type mockAttachmentService interface {
	fetchAttachment(
		service *gmail.Service,
		msgID string, attachID string) (*gmail.MessagePartBody, error)
}

func generateMsgsContents() <-chan *gmail.Message {
	msgsCh := make(chan *gmail.Message, 1)
	layout := "01/02/2006 3:04:05 PM"
	t, _ := time.Parse(layout, "11/20/2019 2:03:46 PM")
	msgs := []*gmail.Message{
		{
			InternalDate: t.UnixNano(),
			Payload: &gmail.MessagePart{
				Parts: []*gmail.MessagePart{
					{
						Filename: "attachmentfile.pdf",
						Body: &gmail.MessagePartBody{
							AttachmentId: "attachmentId",
						},
					},
				},
			},
			Id: "msgIdhere",
		},
	}
	defer close(msgsCh)
	for _, v := range msgs {
		msgsCh <- v
	}

	return msgsCh
}

func (a *mockAttachment) fetchAttachment(
	service *gmail.Service,
	msgID string, attachID string) (*gmail.MessagePartBody, error) {
	attachment := gmail.MessagePartBody{
		Data: "some attachment Data Here",
	}

	return &attachment, nil
}

func Test_getAttachment(t *testing.T) {
	service := new(gmail.Service)
	var as mockAttachmentService = &mockAttachment{}
	msgContents := generateMsgsContents()
	data := "some attachment Data Here"

	atts, _ := getAttachment(msgContents, service, as)

	for a := range atts {
		if a.data != data {
			t.Errorf("getAttachment() = %v, want %v", a.data, data)
		}
	}
}

type mockAttachmentWithFetchError struct {
}

func (a *mockAttachmentWithFetchError) fetchAttachment(
	service *gmail.Service,
	msgID string, attachID string) (*gmail.MessagePartBody, error) {
	attachment := gmail.MessagePartBody{}

	return &attachment, errors.New("Error fetching attachment")
}

func Test_getAttachmentWithFetchError(t *testing.T) {
	service := new(gmail.Service)
	var as mockAttachmentService = &mockAttachmentWithFetchError{}
	msgContents := generateMsgsContents()

	_, err := getAttachment(msgContents, service, as)

	expected := "Unable to retrieve Attachment"
	for e := range err {
		if e.msg != expected {
			t.Errorf("getMessageContent() = %v, want %v", e.msg, expected)
		}
	}
}
