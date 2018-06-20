// Copyright © 2018 NAME HERE <EMAIL ADDRESS>
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package cmd

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"log"
	"logsub/packages/nic"
	"logsub/packages/sub"
	"net/url"
	"os"
	"os/exec"
	"os/user"
	"strings"

	"github.com/Azure/azure-sdk-for-go/profiles/latest/storage/mgmt/storage"
	"github.com/Azure/azure-storage-file-go/2017-07-29/azfile"
	"github.com/Azure/go-autorest/autorest"
	"github.com/Azure/go-autorest/autorest/to"
	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

const (
	// DefaultBaseURI is the default URI used for the service Network
	DefaultBaseURI = "https://management.azure.com"
)

var (
	ctx = context.Background()

	authorizer autorest.Authorizer
	rg         string
	global     string
)

func createshare(vmname string, rgname string) error {
	fmt.Println("Going to create the mount /mnt/forlogs/ on vm ", vmname)
	fmt.Printf("using command az vm run-command  invoke --resource-group %s --name %s --command-id RunShellScript --scripts mkdir -p /mnt/forlogs", rgname, vmname)
	color.Red("\n make sure the vmname and rg are correct before pressing enter\n")
	color.Green("Press 'Enter' to continue...")
	bufio.NewReader(os.Stdin).ReadBytes('\n')
	color.Green("please wait .... processing......")
	cmd := exec.Command("az", "vm", "run-command", "invoke", "--resource-group", rgname, "--name", vmname, "--command-id", "RunShellScript", "--scripts", "mkdir -p /mnt/forlogs")

	err := cmd.Run()
	if err != nil {
		fmt.Printf("completed : %v\n", err)
		fmt.Printf("please ignore if exit 1\n")
	}
	return err
}

func createstorage(ctx context.Context, storageAccountsClient storage.AccountsClient, rgname string, location string) (s storage.Account, err error) {

	saname := nic.Randname(5)
	r := strings.Replace(saname, "", "k8log", 1)
	m := strings.ToLower(r)
	fmt.Printf("creating storage account for logging name %s ", m)
	future, err := storageAccountsClient.Create(
		ctx,
		rgname,
		m,
		storage.AccountCreateParameters{
			Sku: &storage.Sku{
				Name: storage.StandardLRS},
			Kind:     storage.Storage,
			Location: to.StringPtr(location),
			AccountPropertiesCreateParameters: &storage.AccountPropertiesCreateParameters{},
		})
	if err != nil {
		return s, fmt.Errorf("cannot create storage account: %v", err)
	}

	err = future.WaitForCompletion(ctx, storageAccountsClient.Client)
	if err != nil {
		return s, fmt.Errorf("cannot get the storage account create future response: %v", err)
	}

	return future.Result(storageAccountsClient)

}
func holdkey(key string) string {
	global := key
	return global

}

func dooper(vmname string, rgname string, storagename string) {
	buf := bytes.Buffer{}

	//install cifs
	buf.WriteString("az vm run-command invoke --resource-group " + rgname)
	buf.WriteString(" --name " + vmname)
	buf.WriteString(" --command-id RunShellScript --scripts ")
	buf.WriteString("'sudo apt-get update && sudo apt-get install cifs-utils'\n")
	///////////////////
	buf.WriteString("az vm run-command invoke --resource-group " + rgname)
	buf.WriteString(" --name " + vmname)
	buf.WriteString(" --command-id RunShellScript --scripts ")
	buf.WriteString("'sudo mount -t cifs //" + storagename)
	buf.WriteString(".file.core.windows.net/k8logs")
	buf.WriteString(" /mnt/forlogs/ -o vers=3.0,")
	buf.WriteString("username=" + storagename)
	buf.WriteString(",password=" + global)
	buf.WriteString(",dir_mode=0777,file_mode=0777,sec=ntlmssp'\n")
	//kubelog command **************************************
	buf.WriteString("az vm run-command invoke --resource-group " + rgname)
	buf.WriteString(" --name " + vmname)
	buf.WriteString(" --command-id RunShellScript --scripts ")
	buf.WriteString("'journalctl -u kube* --no-pager>>")
	buf.WriteString("/mnt/forlogs/" + vmname)
	buf.WriteString(".log'\n")
	//clusterlog**************************************
	buf.WriteString("az vm run-command invoke --resource-group " + rgname)
	buf.WriteString(" --name " + vmname)
	buf.WriteString(" --command-id RunShellScript --scripts ")
	buf.WriteString("'cp /var/log/azure/cluster-provision.log ")
	buf.WriteString("/mnt/forlogs/" + vmname)
	buf.WriteString(".cluster-provision.log'\n")
	//syslog//////////////////////////////////
	buf.WriteString("az vm run-command invoke --resource-group " + rgname)
	buf.WriteString(" --name " + vmname)
	buf.WriteString(" --command-id RunShellScript --scripts ")
	buf.WriteString("'mkdir /mnt/forlogs/syslog" + vmname)
	buf.WriteString(" && cp /var/log/syslog* /mnt/forlogs/syslog'" + vmname)
	buf.WriteString("\n")
	//journalall/////////////////////////////////////
	buf.WriteString("az vm run-command invoke --resource-group " + rgname)
	buf.WriteString(" --name " + vmname)
	buf.WriteString(" --command-id RunShellScript --scripts ")
	buf.WriteString("'journalctl --no-pager >>")
	buf.WriteString(" /mnt/forlogs/" + vmname)
	buf.WriteString(".journal'\n")
	//iptables/////////////////////////////////////////////////////
	buf.WriteString("az vm run-command invoke --resource-group " + rgname)
	buf.WriteString(" --name " + vmname)
	buf.WriteString(" --command-id RunShellScript --scripts ")
	buf.WriteString("'sudo iptables-save >>")
	buf.WriteString(" /mnt/forlogs/" + vmname)
	buf.WriteString(".iptable'\n")

	mycmd := buf.String()
	fmt.Println("\nGoing to execute\n", mycmd)

	usr, _ := user.Current()

	f, err := os.Create(usr.HomeDir + "/cmd.sh")
	if err != nil {
		fmt.Println(err)
	}
	defer f.Close()
	//var out bytes.Buffer
	//cmd.Stdout = &out
	_, err = f.WriteString(mycmd)
	if err != nil {
		fmt.Printf("Command finished with error: %v", err)
	}

	f.Sync()
	//fmt.Println(usr.HomeDir + "/cmd.sh")
	cmd := exec.Command("sh", usr.HomeDir+"/cmd.sh")

	err = cmd.Run()
	if err != nil {
		fmt.Printf("Command finished with error(ignore is 1,2): %v\n", err)
	}
	var out bytes.Buffer
	cmd.Stdout = &out

	color.Green("finished please take the logs from the storage share")

}

// cshareCmd represents the cshare command
var cshareCmd = &cobra.Command{
	Use:   "cshare",
	Short: "this will create storage account and mount the share to the vms and start collecting the logs for the vm you want",
	Long: `This command will ask you for the vm name and will collect all the logs needed:

Example: kcommand cshare `,
	Run: func(cmd *cobra.Command, args []string) {

		//fmt.Println("please enter the vm to mount the share on ", vmname)
		//fmt.Scanf("%s", &vmname)
		//fmt.Println(vmname)
		token, subid, rgname, vnetname, loglocation, subnetName, location, sshpubkey, _ := sub.Readfromauth()
		fmt.Printf("\n\n")
		k := color.New(color.FgCyan, color.Bold)
		k.Printf("**********Please make sure you run az login and selected the correct sub,or else it fail **********\n\n\n")
		color.Green("Press 'Enter' to continue...")
		bufio.NewReader(os.Stdin).ReadBytes('\n')

		color.Yellow("make sure the correct details below : *********************** ")
		color.Yellow("RG: %s\n", rgname)
		color.Yellow("SUBID: %s\n", subid)
		color.Yellow("VNETNAME: %s\n", vnetname)
		color.Yellow("LOGLOCATION: %s\n", loglocation)
		color.Yellow("SUBNET: %s\n", subnetName)
		color.Yellow("LOCATION: %s\n", location)
		color.Yellow("PUBKEY: %s\n", sshpubkey)

		color.Green("**************************************************************************************************************************")
		color.Red("!!!This tool relies on vm guest agent make sure agent is ok!!!!")
		color.Red("Also if the shell closes sooner as expected, you have the command at current dir name cmd.sh ,so you can run manually")
		color.Green("**************************************************************************************************************************")
		color.Green("Press 'Enter' to continue...")
		bufio.NewReader(os.Stdin).ReadBytes('\n')
		var vmname string
		fmt.Println("please enter the vm name you would like to collect the logs from ")
		fmt.Scanf("%s", &vmname)
		if &vmname == nil {
			fmt.Println("vmname cannot be nill")
			os.Exit(1)
		}
		createshare(vmname, rgname)
		fmt.Printf("creating storage account \n")
		storageAccountsClient := storage.NewAccountsClient(subid)
		storageAccountsClient.Authorizer = token
		d, err1 := createstorage(ctx, storageAccountsClient, rgname, location)
		if err1 != nil {
			fmt.Println(err1)
		}
		n := d.Name
		gkeys, err := storageAccountsClient.ListKeys(ctx, rgname, *n)
		f := gkeys.Keys
		for _, l := range *f {

			c := holdkey(*l.Value)
			global = c
		}
		fmt.Println("please wait creating the share")
		credential := azfile.NewSharedKeyCredential(*n, global)
		p := azfile.NewPipeline(credential, azfile.PipelineOptions{})
		u, _ := url.Parse(fmt.Sprintf("https://%s.file.core.windows.net", *n))
		serviceURL := azfile.NewServiceURL(*u, p)
		shareURL := serviceURL.NewShareURL("k8logs")
		_, err = shareURL.Create(ctx, azfile.Metadata{}, 0)
		if err != nil && err.(azfile.StorageError) != nil && err.(azfile.StorageError).ServiceCode() != azfile.ServiceCodeShareAlreadyExists {
			log.Fatal(err)
		}
		fmt.Println("share was created", shareURL.String())
		dooper(vmname, rgname, *n)

	},
}

func init() {

	rootCmd.AddCommand(cshareCmd)

}
