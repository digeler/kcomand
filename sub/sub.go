package sub

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"time"

	"github.com/Azure/go-autorest/autorest"
	"github.com/Azure/go-autorest/autorest/adal"
)

type Details struct {
	ClientID                       string `json:"clientId"`
	ClientSecret                   string `json:"clientSecret"`
	SubscriptionID                 string `json:"subscriptionId"`
	TenantID                       string `json:"tenantId"`
	ActiveDirectoryEndpointURL     string `json:"activeDirectoryEndpointUrl"`
	ResourceManagerEndpointURL     string `json:"resourceManagerEndpointUrl"`
	ActiveDirectoryGraphResourceID string `json:"activeDirectoryGraphResourceId"`
	SqlManagementEndpointURL       string `json:"sqlManagementEndpointUrl"`
	GalleryEndpointURL             string `json:"galleryEndpointUrl"`
	ManagementEndpointURL          string `json:"managementEndpointUrl"`
	Rgname                         string `json:"rgname"`
	Vnetname                       string `json:"vnetname"`
	Loglocation                    string `json:"loglocation"`
	Subnetname                     string `json:"subnetname"`
	Location                       string `json:"location"`
	SshPublicKey                   string `json:"sshPublicKey"`
}

func Readfromauth() (autorest.Authorizer, string, string, string, string, string, string, string, error) {

	var auth string
	fmt.Println("please enter the full path to the auth file \n E.g /usr/local/auth")
	_, err := fmt.Scan(&auth)
	if err != nil {
		fmt.Println(err)
	}

	fmt.Printf("your path is set to %s \n", auth)

	Jsonfile, err := os.Open(auth)

	if err != nil {
		log.Panic(err)
	}
	fmt.Println("success open file")
	defer Jsonfile.Close()
	time.Sleep(5000)

	Value, _ := ioutil.ReadAll(Jsonfile)

	var Sub Details
	json.Unmarshal(Value, &Sub)
	config, err := adal.NewOAuthConfig(Sub.ActiveDirectoryEndpointURL, Sub.TenantID)
	if err != nil {
		fmt.Println(err)
	}

	spToken, err := adal.NewServicePrincipalToken(*config, Sub.ClientID, Sub.ClientSecret, Sub.ResourceManagerEndpointURL)
	if err != nil {
		fmt.Println(err)
	}

	return autorest.NewBearerAuthorizer(spToken), Sub.SubscriptionID, Sub.Rgname, Sub.Vnetname, Sub.Loglocation, Sub.Subnetname, Sub.Location, Sub.SshPublicKey, err

}
