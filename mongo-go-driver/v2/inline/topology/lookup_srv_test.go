package topology_test

import (
	"fmt"
	"net"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestLookupSRV(t *testing.T) {
	cname, srvs, err := net.LookupSRV("mongodb", "tcp", "cluster0.ezdmaqc.mongodb.net")
	require.NoError(t, err)

	fmt.Println("CNAME:", cname)
	fmt.Println("SRVs:", srvs)
}
