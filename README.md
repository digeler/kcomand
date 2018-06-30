# How to use :
clone the repo using git clone : git clone https://github.com/digeler/kcomand.git

to run currently i have only one option cshare.
./kcomand and then press enter.

ps : make sure you edit the auth file and also used az login to connect to the correct sub

short video here : https://youtu.be/c2NFa6EXl9s

Available Commands:

  cshare -- this will create storage account and mount the share to the vms and start collecting the logs for the vm you want
    
  getIpTables --  Get the list of ip tables from specific node
  
  getcnilog --    This will collect all CNI related logs from specific node
  
  getkubeconfig -- This will collect your kube configuration runtime to a file
  
  help    --     Help about any command
