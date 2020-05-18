package nificlient

import (
	"github.com/erdrix/nifikop/pkg/apis/nifi/v1alpha1"
)

// CreateUserACLs creates Nifi ACLs for the given access type and user
// `literal` patternType will be used if patternType == ""
func (n *nifiClient) CreateUserACLs(accessType v1alpha1.NifiAccessType, dn string, topic string) (err error) {
	//csuserName := fmt.Sprintf("User:%s", dn)

	/*switch accessType {
	case v1alpha1.KafkaAccessTypeRead:
		return k.createReadACLs(userName, topic, aclPatternType)
	case v1alpha1.KafkaAccessTypeWrite:
		return k.createWriteACLs(userName, topic, aclPatternType)
	default:
		return errorfactory.New(errorfactory.InternalError{}, fmt.Errorf("unknown type: %s", accessType), "unrecognized access type")
	}*/
	return
}

// DeleteUserACLs removes all ACLs for a given user
func (n *nifiClient) DeleteUserACLs(dn string) (err error) {
	/*matches, err := k.admin.DeleteACL(sarama.AclFilter{
		Principal: &dn,
	}, false)
	if err != nil {
		return
	}
	for _, x := range matches {
		if x.Err != sarama.ErrNoError {
			return x.Err
		}
	}*/
	return
}