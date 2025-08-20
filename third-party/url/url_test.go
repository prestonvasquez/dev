package url

import (
	"fmt"
	"net/url"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestURLWithMongoDB(t *testing.T) {
	str := "mongodb://user:password@localhost:27017/dbname?authSource=admin&ssl=true"

	u, err := url.Parse(str)
	require.NoError(t, err)

	q := u.Query()
	q.Set("page", "2")

	u.RawQuery = q.Encode()

	fmt.Println("Parsed URL:", u.String())
}
