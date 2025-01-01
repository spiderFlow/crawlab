package interfaces

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
	"os/exec"
)

type DependencyInstallerService interface {
	GetInstallDependencyRequirementsCmdBySpiderId(id primitive.ObjectID) (cmd *exec.Cmd, err error)
}
