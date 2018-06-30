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
	"fmt"
	"log"
	"logsub/packages/sub"
	"net/url"
	"os"
	"os/exec"
	"os/user"

	"github.com/Azure/azure-sdk-for-go/profiles/latest/storage/mgmt/storage"
	"github.com/Azure/azure-storage-file-go/2017-07-29/azfile"
	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

func iptables(vmname string, rgname string, storagename string) {
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
	//////////////////////
	//get ip talbes **************************************
	buf.WriteString("az vm run-command invoke --resource-group " + rgname)
	buf.WriteString(" --name " + vmname)
	buf.WriteString(" --command-id RunShellScript --scripts ")
	buf.WriteString("'iptables-save >>")
	buf.WriteString("/mnt/forlogs/" + vmname)
	buf.WriteString(".log'\n")

	mycmd := buf.String()
	fmt.Println("\nGoing to execute\n", mycmd)
	usr, _ := user.Current()

	f, err := os.Create(usr.HomeDir + "/iptables.sh")
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
	cmd := exec.Command("sh", usr.HomeDir+"/iptables.sh")

	err = cmd.Run()
	if err != nil {
		fmt.Printf("Command finished with error(ignore is 1,2): %v\n", err)
	}
	var out bytes.Buffer
	cmd.Stdout = &out

	color.Green("finished please take the logs from the storage share")

}

// getIpTablesCmd represents the getIpTables command
var getIpTablesCmd = &cobra.Command{
	Use:   "getIpTables",
	Short: "Get the list of ip tables from specific node",
	Long:  `This will get the iptables from the node,Just type getiptables and follow instructions`,

	Run: func(cmd *cobra.Command, args []string) {

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
		if vmname == "" {
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
		fmt.Println("getting the ip tables to the share")
		iptables(vmname, rgname, *n)

	},
}

func init() {
	rootCmd.AddCommand(getIpTablesCmd)
	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// getIpTablesCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// getIpTablesCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
