package spider

import "github.com/crawlab-team/crawlab/core/interfaces"

var templateSvcInstance interfaces.SpiderTemplateService

func SetSpiderTemplateRegistryService(svc interfaces.SpiderTemplateService) {
	templateSvcInstance = svc
}

func GetSpiderTemplateRegistryService() interfaces.SpiderTemplateService {
	return templateSvcInstance
}
