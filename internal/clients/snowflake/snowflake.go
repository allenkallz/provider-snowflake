package snowflake

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	dbv1alpha1 "github.com/allenkallz/provider-snowflake/apis/database/v1alpha1"

	"github.com/allenkallz/provider-snowflake/apis/v1alpha1"
	"github.com/crossplane/crossplane-runtime/pkg/errors"
	"github.com/crossplane/crossplane-runtime/pkg/resource"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/golang-jwt/jwt/v5"
)

const (
	invalidName   = "Invalid name"
	requestFailed = "Failed to create request"
	configNotJson = "Spec Config not actually JSON"
	readCredError = "Can't read provider credential "
)

var ErrNotFound = errors.New("Not found")
var ErrBadRequest = errors.New("Bad request")

type Client interface {
	// TableClient
	DatabaseClient
}

type DatabaseClient interface {
	ListDatabase(ctx context.Context, dbinfo DbInfo)
	FetchDatabase(ctx context.Context, db *dbv1alpha1.DatabaseParameters) (DbInfo, error)
	CreateDatabase(ctx context.Context, db *dbv1alpha1.DatabaseParameters) (string, error)
	DeleteDatabase(ctx context.Context, db *dbv1alpha1.DatabaseParameters) error
	UpdateDatabase(ctx context.Context, dbinfo DbInfo)
}

type ClientInfo struct {
	SnowflakeAccount string
	Username         string
	FingerPrint      string
	PrivateKey       string
	httpClient       *http.Client
}

// all helper method
func GetClientInfo(ctx context.Context, c client.Client, mg resource.Managed) (*ClientInfo, error) {

	switch {
	case mg.GetProviderConfigReference() != nil:
		return UseProviderConfig(ctx, c, mg)
	default:
		return nil, errors.New("providerConfigRef is not given")
	}
}

func UseProviderConfig(ctx context.Context, c client.Client, mg resource.Managed) (*ClientInfo, error) {

	pc := &v1alpha1.ProviderConfig{}
	if err := c.Get(ctx, types.NamespacedName{Name: mg.GetProviderConfigReference().Name}, pc); err != nil {
		return nil, errors.Wrap(err, "cannot get referenced Provider")
	}

	t := resource.NewProviderConfigUsageTracker(c, &v1alpha1.ProviderConfigUsage{})
	if err := t.Track(ctx, mg); err != nil {
		return nil, errors.Wrap(err, "cannot track ProviderConfig usage")
	}

	// read private key from of secretRef
	privateKey, err := authFromCredentials(ctx, c, pc.Spec.PrivateKey)
	if err != nil {
		return nil, errors.Wrap(err, readCredError)
	}

	// finger print read
	fingerPrint, err := authFromCredentials(ctx, c, pc.Spec.FingerPrint)
	if err != nil {
		return nil, errors.Wrap(err, readCredError)
	}

	updatedAccount := strings.ReplaceAll(pc.Spec.SnowflakeAccount, ".", "-")
	return &ClientInfo{
		SnowflakeAccount: strings.ToUpper(updatedAccount),
		Username:         strings.ToUpper(pc.Spec.Username),
		FingerPrint:      fingerPrint,
		PrivateKey:       privateKey,
		httpClient:       &http.Client{},
	}, nil
}

// Read token from secret
func authFromCredentials(ctx context.Context, c client.Client, creds v1alpha1.ProviderCredentials) (string, error) {
	csr := creds.SecretRef
	if csr == nil {
		return "", errors.New("no credentials secret referenced")
	}

	s := &corev1.Secret{}

	if err := c.Get(ctx, types.NamespacedName{Namespace: csr.Namespace, Name: csr.Name}, s); err != nil {
		return "", errors.Wrap(err, "cannot get credentials secret")
	}

	return string(s.Data[csr.Key]), nil
}

// Generate JWT Token
// ToDo :  return active token if exist
func generateJWT(c ClientInfo) (string, error) {

	fmt.Println("Creating JWT token")

	// fmt.Println("username: ", c.Username)
	// fmt.Println("fingerprint : ", c.FingerPrint)
	// fmt.Println("account :", c.SnowflakeAccount)
	// fmt.Println("account :", c.PrivateKey)

	// Define expiration time
	expirationTime := time.Now().UTC().Add(1 * time.Hour).Unix()

	// Create custom claims
	claims := jwt.MapClaims{
		"iss": c.SnowflakeAccount + "." + c.Username + ".SHA256:" + c.FingerPrint,
		"exp": expirationTime,
		"iat": time.Now().UTC().Unix(),
		"sub": c.SnowflakeAccount + "." + c.Username,
	}

	// Create a token with claims
	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)

	// Parse the private key
	privateKey, err := jwt.ParseRSAPrivateKeyFromPEM([]byte(c.PrivateKey))
	if err != nil {
		return "", errors.Wrap(err, "Unable to parse private key")
	}
	// Sign the token with the private key
	tokenString, err := token.SignedString(privateKey)
	if err != nil {
		return "", errors.Wrap(err, "Unable to create token")
	}
	return tokenString, nil
}

func getBaseUrl(c ClientInfo) string {

	// baseUrl, _ := url.JoinPath("https://", c.SnowflakeAccount, "snowflakecomputing.com")

	baseUrl := "https://" + strings.ToLower(c.SnowflakeAccount) + ".snowflakecomputing.com"
	return baseUrl
}

// Setting all common header to request http
func setReqHeaders(req *http.Request, jwtToken string) {

	authToken := fmt.Sprintf("%s %s", "Bearer", jwtToken)

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Authorization", authToken)
	req.Header.Set("X-Snowflake-Authorization-Token-Type", "KEYPAIR_JWT")
}

func dclose(c io.Closer) {
	if err := c.Close(); err != nil {
		fmt.Println(err)
	}
}
