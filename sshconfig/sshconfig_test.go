package sshconfig_test

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/nitrous-io/solo-tunnel/sshconfig"
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
			sshKeyPath             string
			defaultSSHKeyPath      string
			OrigDefaultSSHKeyPaths *[]string
		)

		createPrivateKey := func(prefix string) string {
			sshKeyFile, err := ioutil.TempFile("", prefix)
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

			return sshKeyFile.Name()
		}

		BeforeEach(func() {
			OrigDefaultSSHKeyPaths = &sshconfig.DefaultSSHKeyPaths

			sshKeyPath = createPrivateKey("ssh-key")
			defaultSSHKeyPath = createPrivateKey("default-ssh-key")
			sshconfig.DefaultSSHKeyPaths = []string{defaultSSHKeyPath}
		})

		AfterEach(func() {
			sshconfig.DefaultSSHKeyPaths = *OrigDefaultSSHKeyPaths
			os.Remove(sshKeyPath)
			os.Remove(defaultSSHKeyPath)
		})

		It("returns ssh host and ssh config", func() {
			sshHost, config, err := sshconfig.Config("ssh://testuser@example.com:12345", sshKeyPath)
			Expect(err).To(BeNil())

			Expect(sshHost).To(Equal("example.com:12345"))
			Expect(config.User).To(Equal("testuser"))
			Expect(config.Auth).NotTo(BeNil())
		})

		Context("ssh key path is not specified", func() {
			It("falls back to default ssh key path", func() {
				sshHost, config, err := sshconfig.Config("ssh://testuser@example.com:12345", "")
				Expect(err).To(BeNil())

				Expect(sshHost).To(Equal("example.com:12345"))
				Expect(config.User).To(Equal("testuser"))
				Expect(config.Auth).NotTo(BeNil())
			})

			Context("sshconfig.DefaultSSHKeyPaths includes non-exist file", func() {
				BeforeEach(func() {
					sshconfig.DefaultSSHKeyPaths = append([]string{"no-ssh-key"}, sshconfig.DefaultSSHKeyPaths...)
				})

				It("returns an error", func() {
					sshHost, config, err := sshconfig.Config("ssh://testuser@example.com:12345", "")
					Expect(err).To(BeNil())

					Expect(sshHost).To(Equal("example.com:12345"))
					Expect(config.User).To(Equal("testuser"))
					Expect(config.Auth).NotTo(BeNil())
				})
			})

			Context("ssh key file does not exist", func() {
				BeforeEach(func() {
					os.Remove(defaultSSHKeyPath)
				})

				It("returns an error", func() {
					_, _, err := sshconfig.Config("ssh://testuser@example.com:12345", "")
					Expect(err).To(Equal(sshconfig.ErrNoSSHKey))
				})
			})
		})

		Context("ssh key file does not exist", func() {
			It("returns an error", func() {
				_, _, err := sshconfig.Config("ssh://testuser@example.com:12345", "no-ssh-key")
				Expect(err).To(Equal(sshconfig.ErrNoSSHKey))
			})
		})
	})
})
