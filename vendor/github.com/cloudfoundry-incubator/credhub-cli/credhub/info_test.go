package credhub_test

import (
	"net/http"
	"net/http/httptest"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "github.com/cloudfoundry-incubator/credhub-cli/credhub"
)

var _ = Describe("Info", func() {
	Context("Info()", func() {
		It("should return auth-server url from the /info endpoint", func() {
			testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Path == "/info" {
					w.Write([]byte(
						`{
							"auth-server": {
								"url": "https://uaa.example.com:8443"
							},
							"app": {
								"name": "CredHub",
								"version": "0.7.0"
							}
						}`,
					))
				}
			}))

			defer testServer.Close()

			ch, _ := New(testServer.URL, ServerVersion("1.2.3"))

			info, err := ch.Info()
			Expect(err).To(BeNil())

			Expect(info.App.Name).To(Equal("CredHub"))
			Expect(info.App.Version).To(Equal("0.7.0"))
			Expect(info.AuthServer.URL).To(Equal("https://uaa.example.com:8443"))
		})

		Context("when the info endpoint cannot be parsed", func() {
			It("returns an error", func() {
				testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					if r.URL.Path == "/info" {
						w.Write([]byte(`INVALID JSON`))
					}
				}))

				defer testServer.Close()

				ch, _ := New(testServer.URL, ServerVersion("1.2.3"))

				info, err := ch.Info()

				Expect(err).To(HaveOccurred())
				Expect(info).To(BeNil())
			})
		})
	})

	Context("AuthURL()", func() {
		Context("Errors", func() {

			Specify("when ApiURL is inaccessible", func() {
				ch, _ := New("http://localhost:1", ServerVersion("1.2.3"))
				_, err := ch.AuthURL()
				Expect(err).ToNot(BeNil())
			})

			Specify("when auth-server is not returned", func() {
				testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					if r.URL.Path == "/info" {
						w.Write([]byte(`{}`))
					}
				}))
				defer testServer.Close()

				ch, _ := New(testServer.URL, ServerVersion("1.2.3"))
				_, err := ch.AuthURL()

				Expect(err).ToNot(BeNil())
			})
		})
	})
})
