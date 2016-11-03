package tunnel_test

import (
	"testing"

	"github.com/nitrous-io/go-tunnel/tunnel"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func Test(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "tunnel")
}

var _ = Describe("sshtunnel", func() {
	Describe("ParsePort", func() {
		It("parses a port", func() {
			ports, err := tunnel.ParsePort("3000")
			Expect(err).NotTo(HaveOccurred())

			Expect(ports).To(HaveLen(1))
			Expect(ports[0].Local).To(Equal(3000))
			Expect(ports[0].Remote).To(Equal(3000))
		})

		It("parses ports", func() {
			ports, err := tunnel.ParsePort("3000,5000")
			Expect(err).NotTo(HaveOccurred())

			expectedResult := []tunnel.Port{
				{Local: 3000, Remote: 3000},
				{Local: 5000, Remote: 5000},
			}
			Expect(ports).To(HaveLen(2))
			Expect(ports).To(ConsistOf(expectedResult))
		})

		It("returns an error if any ports lower than minimum port number", func() {
			ports, err := tunnel.ParsePort("900-3000")
			Expect(err).To(HaveOccurred())
			Expect(ports).To(BeNil())
		})

		It("works with range", func() {
			ports, err := tunnel.ParsePort("3000-3002,5000")
			Expect(err).NotTo(HaveOccurred())

			expectedResult := []tunnel.Port{
				{Local: 3000, Remote: 3000},
				{Local: 3001, Remote: 3001},
				{Local: 3002, Remote: 3002},
				{Local: 5000, Remote: 5000},
			}
			Expect(ports).To(HaveLen(4))
			Expect(ports).To(ConsistOf(expectedResult))
		})

		It("returns an error when start port is bigger than end port", func() {
			ports, err := tunnel.ParsePort("3002-3000,5000")
			Expect(err).To(HaveOccurred())
			Expect(ports).To(BeNil())
		})

		It("ignores duplicated ports", func() {
			ports, err := tunnel.ParsePort("3000-3002,3001")
			Expect(err).To(HaveOccurred())
			Expect(ports).To(BeNil())
		})

		It("ignores duplicated ports with range", func() {
			ports, err := tunnel.ParsePort("3000-3002,3001-3004")
			Expect(err).To(HaveOccurred())
			Expect(ports).To(BeNil())
		})

		It("works when first and last port are same", func() {
			ports, err := tunnel.ParsePort("3000-3000")
			Expect(err).NotTo(HaveOccurred())

			Expect(ports).To(HaveLen(1))
			Expect(ports[0].Local).To(Equal(3000))
			Expect(ports[0].Remote).To(Equal(3000))
		})

		It("works with port mapping", func() {
			ports, err := tunnel.ParsePort("3000:4000")
			Expect(err).NotTo(HaveOccurred())

			Expect(ports).To(HaveLen(1))
			Expect(ports[0].Local).To(Equal(3000))
			Expect(ports[0].Remote).To(Equal(4000))
		})

		It("works with port mapping", func() {
			ports, err := tunnel.ParsePort("3000:4000, 4000:5000")
			Expect(err).NotTo(HaveOccurred())

			expectedResult := []tunnel.Port{
				{Local: 3000, Remote: 4000},
				{Local: 4000, Remote: 5000},
			}

			Expect(ports).To(HaveLen(2))
			Expect(ports).To(ConsistOf(expectedResult))
		})

		It("returns error when there is any conflicts with mappings", func() {
			ports, err := tunnel.ParsePort("3000:4000, 4000")
			Expect(err).To(HaveOccurred())
			Expect(ports).To(BeNil())
		})
	})
})
