package cgroups_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestCGroups(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Cgroup suite")
}
