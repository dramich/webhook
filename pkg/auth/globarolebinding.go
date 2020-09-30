package auth

import (
	"time"

	rancherv3 "github.com/rancher/rancher/pkg/apis/management.cattle.io/v3"
	v3 "github.com/rancher/webhook/pkg/generated/controllers/management.cattle.io/v3"
	"github.com/rancher/wrangler/pkg/webhook"
	admissionv1 "k8s.io/api/admission/v1"
	authenticationv1 "k8s.io/api/authentication/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/utils/trace"
)

func NewGRBValidator(grClient v3.GlobalRoleCache, escalationChecker *EscalationChecker) webhook.Handler {
	return &globalRoleBindingValidator{
		escalationChecker: escalationChecker,
		globalRoles:       grClient,
	}
}

type globalRoleBindingValidator struct {
	escalationChecker *EscalationChecker
	globalRoles       v3.GlobalRoleCache
}

func (grbv *globalRoleBindingValidator) Admit(response *webhook.Response, request *webhook.Request) error {
	listTrace := trace.New("globalRoleBindingValidator Admit", trace.Field{Key: "user", Value: request.UserInfo.Username})
	defer listTrace.LogIfLong(2 * time.Second)

	newGRB, err := grbObject(request)
	if err != nil {
		return err
	}

	// Pull the global role to get the rules
	globalRole, err := grbv.globalRoles.Get(newGRB.GlobalRoleName)
	if err != nil {
		return err
	}

	return grbv.escalationChecker.confirmNoEscalation(response, request, globalRole.Rules, "")
}

func grbObject(request *webhook.Request) (*rancherv3.GlobalRoleBinding, error) {
	var grb runtime.Object
	var err error
	if request.Operation == admissionv1.Delete {
		grb, err = request.DecodeOldObject()
	} else {
		grb, err = request.DecodeObject()
	}
	return grb.(*rancherv3.GlobalRoleBinding), err
}

func toExtraString(extra map[string]authenticationv1.ExtraValue) map[string][]string {
	result := make(map[string][]string)
	for k, v := range extra {
		result[k] = v
	}
	return result
}
