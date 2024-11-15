package constants

const (
	DependencyTypePython = "python"
	DependencyTypeNode   = "node"
)

const (
	DependencyStatusInstalling   = "installing"
	DependencyStatusInstalled    = "installed"
	DependencyStatusUninstalling = "uninstalling"
	DependencyStatusUninstalled  = "uninstalled"
	DependencyStatusError        = "error"
	DependencyStatusAbnormal     = "abnormal"
)

const (
	DependencyFileTypeRequirementsTxt = "requirements.txt"
	DependencyFileTypePackageJson     = "package.json"
)
