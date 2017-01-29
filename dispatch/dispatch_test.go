package dispatch

import (
	"fmt"
	"html"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"
)

type httpmux struct {
	text string
}

func TestAddRoot(t *testing.T) {

	d := NewDispatch()
	d.AddNamespace("/", &httpmux{"ROOT"})

	handler := d
	server := httptest.NewServer(handler)
	defer server.Close()

	resp, err := http.Get(server.URL)
	if err != nil {
		t.Fatal(err)
	}

	if resp.StatusCode != 200 {
		t.Fatalf("Received non-200 response: %d\n", resp.StatusCode)
	}

	expected := fmt.Sprintf("ROOT: \"/\"")
	actual, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		t.Fatal(err)
	}
	if expected != string(actual) {
		t.Errorf("Expected the message '%s' got '%s'\n", expected, string(actual))
	}
}

func TestRedirect(t *testing.T) {

	d := NewDispatch()
	d.AddNamespace("/redirect/", &httpmux{"REDIRECT"})

	handler := d
	server := httptest.NewServer(handler)
	defer server.Close()

	for _, url := range []string{"/redirect", "/redirect/"} {
		resp, err := http.Get(server.URL + url)
		if err != nil {
			t.Fatal(err)
		}

		if resp.StatusCode != 200 {
			t.Fatalf("Received non-200 response: %d\n", resp.StatusCode)
		}

		expected := fmt.Sprintf("REDIRECT: \"/redirect/\"")
		actual, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			t.Fatal(err)
		}
		if expected != string(actual) {
			t.Errorf("Expected the message '%s' got '%s'\n", expected, string(actual))
		}
	}

}

func (h *httpmux) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "%v: %q", h.text, html.EscapeString(r.URL.Path))
}
