package helpers

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"os/user"
	"net/http"
	"crypto/rand"
	
	"github.com/Azure/azure-sdk-for-go/storage"	
	"github.com/Azure/azure-sdk-for-go/Godeps/_workspace/src/github.com/Azure/go-autorest/autorest"
	"github.com/Azure/azure-sdk-for-go/Godeps/_workspace/src/github.com/Azure/go-autorest/autorest/azure"
)

const (
	credentialsPath = "/.azure/credentials.json"
)

// ToJSON returns the passed item as a pretty-printed JSON string. If any JSON error occurs,
// it returns the empty string.
func ToJSON(v interface{}) string {
	j, _ := json.MarshalIndent(v, "", "  ")
	return string(j)
}

// LoadCredentials reads credentials from a ~/.azure/credentials.json file. See the accompanying
// credentials_sample.json file for an example.
//
// Note: Storing crendentials in a local file must be secured and not shared. It is used here
// simply to reduce code in the examples, but it is not suggested as a best (or even good)
// practice.
func LoadCredentials() (map[string]string, error) {
	u, err := user.Current()
	if err != nil {
		return nil, fmt.Errorf("ERROR: Unable to determine current user")
	}

	n := u.HomeDir + credentialsPath
	f, err := os.Open(n)
	if err != nil {
		return nil, fmt.Errorf("ERROR: Unable to locate or open Azure credentials at %s (%v)", n, err)
	}

	b, err := ioutil.ReadAll(f)
	if err != nil {
		return nil, fmt.Errorf("ERROR: Unable to read %s (%v)", n, err)
	}

	c := map[string]interface{}{}
	err = json.Unmarshal(b, &c)
	if err != nil {
		return nil, fmt.Errorf("ERROR: %s contained invalid JSON (%s)", n, err)
	}

	return ensureValueStrings(c), nil
}

// AuthenticateForARM uses LoadCredentials to load user credentials and uses them to authenticate
// and create a auth token that can be used by subsequent calls to ARM-based APIs.
//
// Note: Storing crendentials in a local file must be secured and not shared. It is used here
// simply to reduce code in the examples, but it is not suggested as a best (or even good)
// practice.
func AuthenticateForARM() (spt *azure.ServicePrincipalToken, sid string, err error) {
	
	c, err := LoadCredentials()
	if err != nil {
		return
	}
	
	sid = c["subscriptionID"]
	tid := c["tenantID"]
	cid := c["clientID"]
	secret := c["clientSecret"]

	spt,err = azure.NewServicePrincipalToken(cid, secret, tid, azure.AzureResourceManagerScope)
	if err != nil {
		return
	}
	
	return 
}

func ensureValueStrings(mapOfInterface map[string]interface{}) map[string]string {
	mapOfStrings := make(map[string]string)
	for key, value := range mapOfInterface {
		mapOfStrings[key] = ensureValueString(value)
	}
	return mapOfStrings
}

func ensureValueString(value interface{}) string {
	if value == nil {
		return ""
	}
	switch v := value.(type) {
	case string:
		return v
	default:
		return fmt.Sprintf("%v", v)
	}
}

// RequestObserver is the signature of an observer function that will be called for each
// HTTP request.
type RequestObserver func(*http.Request)

// ResponseObserver is the signature of an observer function that will be called for each
// HTTP response.
type ResponseObserver func(*http.Response)

// WithInspection provides a convenient way to hook into each HTTP request of a client. 
func WithInspection(callbacks ...RequestObserver) autorest.PrepareDecorator {
	return func(p autorest.Preparer) autorest.Preparer {
		return autorest.PreparerFunc(func(r *http.Request) (*http.Request, error) {
			fmt.Printf("Inspecting Request: %s %s\n", r.Method, r.URL)
			for _,cb := range callbacks {
				if cb != nil {
					cb(r)
				}
			}
			return p.Prepare(r)
		})
	}
}

// ByInspecting provides a convenient way to hook into each HTTP response of a client. 
func ByInspecting(callbacks ...ResponseObserver) autorest.RespondDecorator {
	return func(r autorest.Responder) autorest.Responder {
		return autorest.ResponderFunc(func(resp *http.Response) error {
			fmt.Printf("Inspecting Response: %s for %s %s\n", resp.Status, resp.Request.Method, resp.Request.URL)		   
			for _,cb := range callbacks {
				if cb != nil {
					cb(resp)
				}
			}
			return r.Respond(resp)
		})
	}
}

// RandBytes creates a byte array with length 'n' and fills it with random numbers. By making
// the data readable characters, looking at the data is easier.
func RandBytes(n int) []byte {
	if n <= 0 {
		panic("negative number")
	}
	const alphanum = "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	var bytes = make([]byte, n)
	rand.Read(bytes)
	for i, b := range bytes {
		bytes[i] = alphanum[b%byte(len(alphanum))]
	}
	return bytes	
}

// RandString generates a random string containing only lower-case letters and numbers.
// It is given a character count as its only parameter.
func RandString(n int) string {
	if n <= 0 {
		panic("negative number")
	}
	const alphanum = "0123456789abcdefghijklmnopqrstuvwxyz"
	var bytes = make([]byte, n)
	rand.Read(bytes)
	for i, b := range bytes {
		bytes[i] = alphanum[b%byte(len(alphanum))]
	}
	return string(bytes)
}

// GetStorageClient returns an Azure storage client, which can be used to retrieve service 
// clients for blobs, files, and queues.
func GetStorageClient(storageAccount string, storageAccountKey string) storage.Client {
	cli, _ := storage.NewBasicClient(storageAccount, storageAccountKey)
	return cli
}