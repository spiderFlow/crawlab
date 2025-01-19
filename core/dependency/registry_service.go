package dependency

import (
	"github.com/crawlab-team/crawlab/core/interfaces"
)

var serviceInstance interfaces.DependencyInstallerService

func SetDependencyInstallerRegistryService(svc interfaces.DependencyInstallerService) {
	serviceInstance = svc
}

func GetDependencyInstallerRegistryService() interfaces.DependencyInstallerService {
	return serviceInstance
}
