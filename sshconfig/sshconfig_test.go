package sshconfig_test

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/nitrous-io/nsync/pkg/sshconfig"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func Test(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "sshconfig")
}

var _ = Describe("sshconfig", func() {
	Describe("Config", func() {
		var (
			sshKeyPath string
		)

		BeforeEach(func() {
			sshKeyFile, err := ioutil.TempFile("", "ssh-key")
			Expect(err).To(BeNil())
			defer sshKeyFile.Close()

			privateKey := `-----BEGIN RSA PRIVATE KEY-----
MIIBPAIBAAJBAN55NcYKZeInyTuhcCwFMhDHCmwaIUSdtXdcbItRB/yfXGBhiex0
0IaLXQnSU+QZPRZWYqeTEbFSgihqi1PUDy8CAwEAAQJBAQdUx66rfh8sYsgfdcvV
NoafYpnEcB5s4m/vSVe6SU7dCK6eYec9f9wpT353ljhDUHq3EbmE4foNzJngh35d
AekCIQDhRQG5Li0Wj8TM4obOnnXUXf1jRv0UkzE9AHWLG5q3AwIhAPzSjpYUDjVW
MCUXgckTpKCuGwbJk7424Nb8bLzf3kllAiA5mUBgjfr/WtFSJdWcPQ4Zt9KTMNKD
EUO0ukpTwEIl6wIhAMbGqZK3zAAFdq8DD2jPx+UJXnh0rnOkZBzDtJ6/iN69AiEA
1Aq8MJgTaYsDQWyU/hDq5YkDJc9e9DSCvUIzqxQWMQE=
-----END RSA PRIVATE KEY-----`

			_, err = sshKeyFile.WriteString(privateKey)
			Expect(err).To(BeNil())

			sshKeyPath = sshKeyFile.Name()
		})

		AfterEach(func() {
			os.Remove(sshKeyPath)
		})

		It("returns ssh host and ssh config", func() {
			sshHost, config, err := sshconfig.Config("ssh://testuser@example.com:12345", sshKeyPath)

			Expect(err).To(BeNil())
			Expect(sshHost).To(Equal("example.com:12345"))
			Expect(config.User).To(Equal("testuser"))
			Expect(config.Auth).NotTo(BeNil())
		})
	})
})
