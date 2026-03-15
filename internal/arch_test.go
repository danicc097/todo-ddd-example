package internal_test

import (
	"fmt"
	"go/build"
	"testing"

	"github.com/matthewmcnew/archtest"
)

func TestArchitecture(t *testing.T) {
	modules := []string{"todo", "user", "auth", "workspace", "schedule", "audit"}

	for _, mod := range modules {
		t.Run(mod, func(t *testing.T) {
			domainPkg := fmt.Sprintf("github.com/danicc097/todo-ddd-example/internal/modules/%s/domain", mod)
			appPkg := fmt.Sprintf("github.com/danicc097/todo-ddd-example/internal/modules/%s/application", mod)
			infraPkg := fmt.Sprintf("github.com/danicc097/todo-ddd-example/internal/modules/%s/infrastructure", mod)

			hasDomain := pkgExists(domainPkg)
			hasApp := pkgExists(appPkg)
			hasInfra := pkgExists(infraPkg)

			if hasDomain && hasApp {
				// Domain should not depend on Application
				archtest.Package(t, domainPkg).ShouldNotDependOn(appPkg)
			}

			if hasDomain && hasInfra {
				// Domain should not depend on Infrastructure
				archtest.Package(t, domainPkg).ShouldNotDependOn(infraPkg)
			}

			if hasApp && hasInfra {
				// Application should not depend on Infrastructure
				archtest.Package(t, appPkg).ShouldNotDependOn(infraPkg)
			}
		})
	}
}

func pkgExists(pkg string) bool {
	_, err := build.Import(pkg, ".", build.FindOnly)
	return err == nil
}
