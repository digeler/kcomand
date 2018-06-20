package helper
import (
	"context"
	"fmt"
	"logsub/packages/nic"
	"logsub/packages/sshc"
	"logsub/packages/sub"
	"logsub/packages/vm"

	"golang.org/x/crypto/ssh"

	"strings"

	"github.com/Azure/azure-sdk-for-go/services/network/mgmt/2018-01-01/network"

	"github.com/Azure/azure-sdk-for-go/profiles/latest/compute/mgmt/compute"
	"github.com/Azure/azure-sdk-for-go/services/resources/mgmt/2018-02-01/resources"
	"github.com/Azure/go-autorest/autorest"
)




