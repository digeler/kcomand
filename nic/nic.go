package nic

import (
	"context"
	"fmt"
	"math/rand"
	"time"

	"github.com/Azure/azure-sdk-for-go/profiles/latest/network/mgmt/network"
	"github.com/Azure/go-autorest/autorest/to"
)

func CreateNIC(ctx context.Context, vnetName string, subnetName string, nicName string, location string, rgname string, pubipname string, expand string, subnetsClient network.SubnetsClient, intclient network.InterfacesClient, ipclient network.PublicIPAddressesClient) (nic network.Interface, err error) {

	subnet, err := subnetsClient.Get(ctx, rgname, vnetName, subnetName, "")

	public, err := ipclient.Get(ctx, rgname, pubipname, "")

	nicParams := network.Interface{
		Name:     to.StringPtr(nicName),
		Location: to.StringPtr(location),
		InterfacePropertiesFormat: &network.InterfacePropertiesFormat{
			IPConfigurations: &[]network.InterfaceIPConfiguration{
				{
					Name: to.StringPtr("ipConfig1"),
					InterfaceIPConfigurationPropertiesFormat: &network.InterfaceIPConfigurationPropertiesFormat{
						Subnet: &subnet,
						PrivateIPAllocationMethod: network.Dynamic,
						PublicIPAddress:           &public,
					},
				},
			},
		},
	}

	future, err := intclient.CreateOrUpdate(ctx, rgname, nicName, nicParams)
	if err != nil {
		return nic, fmt.Errorf("cannot create nic: %v", err)
	}

	err = future.WaitForCompletion(ctx, intclient.Client)
	if err != nil {
		return nic, fmt.Errorf("cannot get nic create or update future response: %v", err)
	}

	return future.Result(intclient)
}

//CreatePublicIP creates a new public IP
//var ipclient PublicIPAddressesClient

func CreatePublicIP(ctx context.Context, ipName string, ipclient network.PublicIPAddressesClient, location string, rg string) (ip network.PublicIPAddress, err error) {

	future, err := ipclient.CreateOrUpdate(
		ctx,

		rg,
		ipName,
		network.PublicIPAddress{
			Name:     to.StringPtr(ipName),
			Location: to.StringPtr(location),
			PublicIPAddressPropertiesFormat: &network.PublicIPAddressPropertiesFormat{
				PublicIPAddressVersion:   network.IPv4,
				PublicIPAllocationMethod: network.Dynamic,
			},
		},
	)

	if err != nil {
		return ip, fmt.Errorf("cannot create public ip address: %v", err)
	}

	err = future.WaitForCompletion(ctx, ipclient.Client)
	if err != nil {
		return ip, fmt.Errorf("cannot get public ip address create or update future response: %v", err)
	}

	return future.Result(ipclient)
}

func Randname(n int) string {
	rand.Seed(time.Now().UnixNano())
	var letters = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")
	b := make([]rune, n)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]

	}

	return string(b)
}
