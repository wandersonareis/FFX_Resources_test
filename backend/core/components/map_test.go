package components_test

import (
	"sync/atomic"
	"testing"

	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"

	"ffxresources/backend/core/components"
)

func TestMap(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "Map Suite")
}

var _ = ginkgo.Describe("Map", func() {
	var m *components.Map[string, int]

	ginkgo.BeforeEach(func() {
		m = components.NewEmptyMap[string, int]()
	})

	ginkgo.It("should add and get values", func() {
		m.Add("a", 1)
		val, ok := m.Get("a")
		gomega.Expect(ok).To(gomega.BeTrue())
		gomega.Expect(val).To(gomega.Equal(1))
	})

	ginkgo.It("should remove values", func() {
		m.Add("b", 2)
		m.Remove("b")
		_, ok := m.Get("b")
		gomega.Expect(ok).To(gomega.BeFalse())
	})

	ginkgo.It("should get default when key not present", func() {
		def := m.GetOrDefault("c", 3)
		gomega.Expect(def).To(gomega.Equal(3))
		m.Add("c", 4)
		gomega.Expect(m.GetOrDefault("c", 3)).To(gomega.Equal(4))
	})

	ginkgo.It("should check for keys and values", func() {
		entries := map[string]int{"x": 5, "y": 6}
		m.AddAll(entries)
		gomega.Expect(m.ContainsKey("x")).To(gomega.BeTrue())
		gomega.Expect(m.ContainsValue(6, func(a, b int) bool { return a == b })).To(gomega.BeTrue())
	})

	ginkgo.It("should return keys and values", func() {
		m.Add("k1", 10)
		m.Add("k2", 20)
		gomega.Expect(m.Keys()).To(gomega.ContainElements("k1", "k2"))
		gomega.Expect(m.Values()).To(gomega.ContainElements(10, 20))
	})

	ginkgo.It("should clear all entries", func() {
		m.Add("d", 7)
		m.Clear()
		gomega.Expect(m.Count()).To(gomega.Equal(0))
		gomega.Expect(m.Keys()).To(gomega.BeEmpty())
	})

	ginkgo.It("should iterate with Range", func() {
		entries := map[string]int{"p": 1, "q": 2, "r": 3}
		m.AddAll(entries)
		seen := make(map[string]int)
		m.ForEach(func(k string, v int) {
			seen[k] = v
		})
		gomega.Expect(seen).To(gomega.Equal(entries))
	})

	ginkgo.It("should iterate in parallel with RangeParallel", func() {
		entries := map[string]int{"a1": 100, "b2": 200, "c3": 300}
		m.AddAll(entries)
		var counter atomic.Int32
		m.ParallelForEach(func(k string, v int) {
			counter.Add(1)
		})
		gomega.Expect(int(counter.Load())).To(gomega.Equal(len(entries)))
	})
})
