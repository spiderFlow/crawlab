package user

import (
	"github.com/apex/log"
	"github.com/crawlab-team/crawlab/core/constants"
	"github.com/crawlab-team/crawlab/core/errors"
	"github.com/crawlab-team/crawlab/core/models/models"
	"github.com/crawlab-team/crawlab/core/models/service"
	"github.com/crawlab-team/crawlab/core/utils"
	mongo2 "github.com/crawlab-team/crawlab/db/mongo"
	"github.com/crawlab-team/crawlab/trace"
	"github.com/golang-jwt/jwt/v5"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"sync"
	"time"
)

type Service struct {
	jwtSecret        string
	jwtSigningMethod jwt.SigningMethod
	modelSvc         *service.ModelService[models.User]
}

func (svc *Service) Init() (err error) {
	_, err = svc.modelSvc.GetOne(bson.M{"username": constants.DefaultAdminUsername}, nil)
	if err == nil {
		return nil
	}
	if err.Error() != mongo.ErrNoDocuments.Error() {
		return err
	}
	return svc.Create(
		constants.DefaultAdminUsername,
		constants.DefaultAdminPassword,
		constants.RoleAdmin,
		"",
		primitive.NilObjectID,
	)
}

func (svc *Service) SetJwtSecret(secret string) {
	svc.jwtSecret = secret
}

func (svc *Service) SetJwtSigningMethod(method jwt.SigningMethod) {
	svc.jwtSigningMethod = method
}

func (svc *Service) Create(username, password, role, email string, by primitive.ObjectID) (err error) {
	// validate options
	if username == "" || password == "" {
		return trace.TraceError(errors.ErrorUserMissingRequiredFields)
	}
	if len(password) < 5 {
		return trace.TraceError(errors.ErrorUserInvalidPassword)
	}

	// normalize options
	if role == "" {
		role = constants.RoleNormal
	}

	// check if user exists
	if u, err := svc.modelSvc.GetOne(bson.M{"username": username}, nil); err == nil && u != nil && !u.Id.IsZero() {
		return trace.TraceError(errors.ErrorUserAlreadyExists)
	}

	// transaction
	return mongo2.RunTransaction(func(ctx mongo.SessionContext) error {
		// add user
		u := models.User{
			Username: username,
			Role:     role,
			Password: utils.EncryptMd5(password),
			Email:    email,
		}
		u.SetCreated(by)
		u.SetUpdated(by)
		_, err = svc.modelSvc.InsertOne(u)

		return err
	})
}

func (svc *Service) Login(username, password string) (token string, u *models.User, err error) {
	u, err = svc.modelSvc.GetOne(bson.M{"username": username}, nil)
	if err != nil {
		return "", nil, err
	}
	if u.Password != utils.EncryptMd5(password) {
		return "", nil, errors.ErrorUserMismatch
	}
	token, err = svc.makeToken(u)
	if err != nil {
		return "", nil, err
	}
	return token, u, nil
}

func (svc *Service) CheckToken(tokenStr string) (u *models.User, err error) {
	return svc.checkToken(tokenStr)
}

func (svc *Service) ChangePassword(id primitive.ObjectID, password string, by primitive.ObjectID) (err error) {
	u, err := svc.modelSvc.GetById(id)
	if err != nil {
		return err
	}
	u.Password = utils.EncryptMd5(password)
	u.SetCreatedBy(by)
	return svc.modelSvc.ReplaceById(id, *u)
}

func (svc *Service) MakeToken(user *models.User) (tokenStr string, err error) {
	return svc.makeToken(user)
}

func (svc *Service) makeToken(user *models.User) (tokenStr string, err error) {
	token := jwt.NewWithClaims(svc.jwtSigningMethod, jwt.MapClaims{
		"id":       user.Id,
		"username": user.Username,
		"nbf":      time.Now().Unix(),
	})
	return token.SignedString([]byte(svc.jwtSecret))
}

func (svc *Service) checkToken(tokenStr string) (user *models.User, err error) {
	token, err := jwt.Parse(tokenStr, svc.getSecretFunc())
	if err != nil {
		return
	}

	claim, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		err = errors.ErrorUserInvalidType
		return
	}

	if !token.Valid {
		err = errors.ErrorUserInvalidToken
		return
	}

	id, err := primitive.ObjectIDFromHex(claim["id"].(string))
	if err != nil {
		return user, err
	}
	username := claim["username"].(string)
	user, err = svc.modelSvc.GetById(id)
	if err != nil {
		err = errors.ErrorUserNotExists
		return
	}

	if username != user.Username {
		err = errors.ErrorUserMismatch
		return
	}

	return
}

func (svc *Service) getSecretFunc() jwt.Keyfunc {
	return func(token *jwt.Token) (interface{}, error) {
		return []byte(svc.jwtSecret), nil
	}
}

func newUserService() (svc *Service, err error) {
	// service
	svc = &Service{
		modelSvc:         service.NewModelService[models.User](),
		jwtSecret:        "crawlab",
		jwtSigningMethod: jwt.SigningMethodHS256,
	}

	// initialize
	if err := svc.Init(); err != nil {
		log.Errorf("failed to initialize user service: %v", err)
		return nil, trace.TraceError(err)
	}

	return svc, nil
}

var userSvc *Service
var userSvcOnce sync.Once

func GetUserService() (svc *Service, err error) {
	userSvcOnce.Do(func() {
		userSvc, err = newUserService()
		if err != nil {
			return
		}
	})
	return userSvc, nil
}
