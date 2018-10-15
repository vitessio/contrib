package mysql

import (
	"encoding/json"
	"flag"
	"fmt"
	"net"
	"os"

	"k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"vitess.io/vitess/go/vt/log"
	querypb "vitess.io/vitess/go/vt/proto/query"
)

var (
	kubeconfig *string

	/*
		To implement authorization through Kubernetes secrets, you must have created a secret with the following in the Data specification: {MysqlNativePassword, UserData}
		Below, we observe three command line flags. The namespace and name of the secret are passed in normally. Only pass in a value for config if you have an external configuration.
	*/
	mysqlAuthServerK8sSecret    = flag.String("mysql_auth_server_k8s_secret", "", "Name of the Kubernetes secret that contains the auth information for vtgate")
	mysqlAuthServerK8sConfig    = flag.String("mysql_auth_server_k8s_config", "", "The Kubernetes configuration that sets information for the clientset")
	mysqlAuthServerK8sNamespace = flag.String("mysql_auth_server_k8s_namespace", "default", "The namespace in which the Kubernetes secret lives in")
)

// SecretGetter implements K8sAuthSecretGetter
type SecretGetter interface {
	Get(getOptions metav1.GetOptions) (*v1.Secret, error)
}

// K8sAuthSecretGetter takes information from command line flags and holds the clientset
type K8sAuthSecretGetter struct {
	namespace string
	secret    string
	clientset *kubernetes.Clientset
}

// Get uses the K8sAuthSecretGetter to fetch the information from the Kubernetes secret specified
func (c *K8sAuthSecretGetter) Get(getOptions metav1.GetOptions) (*v1.Secret, error) {
	return c.clientset.CoreV1().Secrets(c.namespace).Get(c.secret, getOptions)
}

// AuthServerK8s indicates that the user will provide authentication through a Kubernetes secret
type AuthServerK8s struct {
	// Method can only be set to:
	// - MysqlNativePassword
	Method string
	// Entries contains the users, passwords and user data.
	Entries map[string][]*AuthServerK8sEntry
	Getter  SecretGetter
}

// AuthServerK8sEntry stores user values
type AuthServerK8sEntry struct {
	// MysqlNativePassword is generated by password hashing methods in MySQL.
	MysqlNativePassword string
	UserData            string
	Username            string
}

// InitAuthServerK8s registers AuthServerK8s
func InitAuthServerK8s() {
	if mysqlAuthServerK8sSecret == nil || *mysqlAuthServerK8sSecret == "" {
		log.Infof("Not configuring AuthServerK8s, as you must provide a Kubernetes secret name")
		return
	}

	authServerK8s := NewAuthServerK8s()
	if authServerK8s != nil {
		log.Errorf("Successfully created AuthServerK8s")
	}

	RegisterAuthServerImpl("k8s", authServerK8s)
}

func (a *AuthServerK8s) parseJSON() error {
	secret, err := a.Getter.Get(metav1.GetOptions{})
	data := string(secret.Data["UserData"])
	jsonConfig := []byte(data)

	if errors.IsNotFound(err) {
		return fmt.Errorf("Secret not found")
	} else if statusError, isStatus := err.(*errors.StatusError); isStatus {
		return fmt.Errorf("Error getting secret %v", statusError.ErrStatus.Message)
	} else if err != nil {
		return fmt.Errorf("Error getting secret %v", err)
	}

	a.Entries = make(map[string][]*AuthServerK8sEntry)
	if err := json.Unmarshal(jsonConfig, &a.Entries); err != nil {
		log.Exitf("Error parsing json from K8s secret: %v", err)
	}
	return nil
}

// NewAuthServerK8s returns a new empty AuthServerK8s
func NewAuthServerK8s() *AuthServerK8s {
	var config *rest.Config
	var err error

	if mysqlAuthServerK8sSecret != nil && *mysqlAuthServerK8sConfig != "" {
		config, err = clientcmd.BuildConfigFromFlags("", *mysqlAuthServerK8sConfig)
	} else {
		config, err = rest.InClusterConfig()
	}

	if err != nil {
		return nil
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil
	}

	return &AuthServerK8s{
		Method:  MysqlNativePassword,
		Entries: make(map[string][]*AuthServerK8sEntry),
		Getter: &K8sAuthSecretGetter{
			namespace: *mysqlAuthServerK8sNamespace,
			secret:    *mysqlAuthServerK8sSecret,
			clientset: clientset,
		},
	}
}

// AuthMethod is part of the AuthServer interface.
func (a *AuthServerK8s) AuthMethod(user string) (string, error) {
	log.Errorf("The method is: %v", a.Method)
	return a.Method, nil
}

// Salt is part of the AuthServer interface.
func (a *AuthServerK8s) Salt() ([]byte, error) {
	return NewSalt()
}

// ValidateHash is part of the AuthServer interface.
func (a *AuthServerK8s) ValidateHash(salt []byte, user string, authResponse []byte, _ net.Addr) (Getter, error) {

	a.parseJSON()

	/*
		secret, err := a.Getter.Get(metav1.GetOptions{})

		if errors.IsNotFound(err) {
			return nil, fmt.Errorf("Secret not found")
		} else if statusError, isStatus := err.(*errors.StatusError); isStatus {
			return nil, fmt.Errorf("Error getting secret %v", statusError.ErrStatus.Message)
		} else if err != nil {
			return nil, fmt.Errorf("Error getting secret %v", err)
		}

			entries := []*AuthServerK8sEntry{
				{
					MysqlNativePassword: string(secret.Data["MysqlNativePassword"]),
					UserData:            string(secret.Data["UserData"]),
				},
			}
	*/
	entries, ok := a.Entries[user]
	if !ok {
		return &K8sUserData{""}, NewSQLError(ERAccessDeniedError, SSAccessDeniedError, "Access denied for user '%v'", user)
	}

	for _, entry := range entries {
		if entry.MysqlNativePassword != "" {
			isPass := isPassScrambleMysqlNativePassword(authResponse, salt, entry.MysqlNativePassword)
			if isPass {
				return &K8sUserData{entry.UserData}, nil
			}
		}
	}
	return &K8sUserData{""}, NewSQLError(ERAccessDeniedError, SSAccessDeniedError, "Access denied for user '%v'", user)
}

// Negotiate returns an error becuase we only support MysqlNativePassword for the K8s AuthServer
func (a *AuthServerK8s) Negotiate(c *Conn, user string, remoteAddr net.Addr) (Getter, error) {
	return &K8sUserData{""}, NewSQLError(ERAccessDeniedError, SSAccessDeniedError, "I'm sorry Dave, I'm afraid I can't do that.", user)
}

// K8sUserData holds the username
type K8sUserData struct {
	value string
}

// Get returns the wrapped username
func (sud *K8sUserData) Get() *querypb.VTGateCallerID {
	return &querypb.VTGateCallerID{Username: sud.value}
}

func homeDir() string {
	if h := os.Getenv("HOME"); h != "" {
		return h
	}
	return os.Getenv("USERPROFILE") // windows
}